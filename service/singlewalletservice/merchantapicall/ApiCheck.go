package merchantapicall

import (
    "encoding/json"
    "time"
    "net/http"
    "errors"
    "fmt"
    models "videogame/common/model"
)

func (m *MerchantApiModule) CheckMerchantServiceOK(queueId int64, merchantKey string) (error, bool) {
    var (
        data    *ApiResponse
    )

    fmt.Printf("MerchantApiModule CheckMerchantServiceOK, queueId=%d, merchantKey=%s\n", queueId, merchantKey)

    done := make(chan struct{})

    m.AsyncDoByQueue(queueId, func() bool {

        var err error
        data, err = m.CheckMerchantOK(merchantKey)

        if err != nil {
            fmt.Printf("检查商户单钱包服失败: %v", err)
            return false
        }

        return true

    }, func(err error) {

        defer close(done)

        if err != nil {
            fmt.Printf("CheckMerchantOK请求错误: %v", err)
            return
        }

        if data.Code != 200 {
            fmt.Printf("CheckMerchantOK失败，msg=%s\n", data.Msg)
            return 
        }

        return
    })

    select {
    case <-done:
    case <-time.After(30 * time.Second):
        fmt.Printf("CheckFromMerchant() 等待商户余额响应超时")
        return errors.New("请求超时"), false
    }

    return nil, true
}

// CheckMerchantOK 检查商户的API接口是否可用
func (m *MerchantApiModule) CheckMerchantOK(merchantKey string) (*ApiResponse, error) {
    const (
        merchantURL     = "http://127.0.0.1:9003/orange/api/v1/single_wallet/ok/"
        orToken     = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHQiOjE3NDY3Nzc4NTAsImlhdCI6MTc0NDE4NTg1MCwibWVyY2hhbnRfY29kZSI6IlZFWlVYNzdTIn0.aHyl32p5pvhhM_fnfzx7NQxWwq1QsuC9DaqAO7Mi0eM" //将替换为db里商户给我们的jwt token
        requestTimeout  = 30 * time.Second
    )

    fmt.Println("MerchantApiModule FetchBalanceFromMerchant---------------------")

    merchant, ok := m.mapMerchant[merchantKey]
    if !ok {
        fmt.Printf("商户不在map里，用 merchantKey查询...\n")

        merchantInfo, err := models.GetMerchantByName(merchantKey)
        if err != nil {
            return nil, err
        }

        merchant = &Merchant{
            MerchantCode: merchantInfo.MerchantCode,
            MerchantName: merchantInfo.MerchantName,
            OrToken:      orToken,
            WalletUrl:    merchantURL,
        }

        m.mapMerchant[merchantKey] = merchant

    }

    // 构建 HTTP 请求
    requestUrl := fmt.Sprintf("%s%s", merchant.WalletUrl, "/healthcheck/")
    req, err := http.NewRequest("GET", requestUrl, nil)
    if err != nil {
        return nil, err
    }
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("OrToken", merchant.OrToken)

    client := &http.Client{Timeout: requestTimeout}
    resp, err := client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    // 解析响应
    var result ApiResponse
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }
    if result.Code != 200 {
        return nil, fmt.Errorf("商户返回错误: %s", result.Msg)
    }

    return &result, nil
}
