package merchantapicall

import (
    "encoding/json"
    "time"
    "net/http"
    "bytes"
    "fmt"
    models "videogame/common/model"
    "strings"
)

// MerchantBalanceRequest 商户余额请求参数
type GetBalanceRequest struct {
    Username string        `json:"username"`
    Code     string        `json:"currency"` // 币种，如 USD、CNY 等
}

type ApiResponse struct {
    Code    int             `json:"code"`
    Msg     string          `json:"msg"`
    Data    json.RawMessage `json:"data"`
}

type GetBalanceDataReturn struct {
    Balance float64        `json:"balance"`
    Code    string         `json:"currency"`
}

var (
    codeDigitsMap map[string]uint
)

func init() {
    if codeDigitsMap == nil {
        codeDigitsMap = make(map[string]uint)
        currs, err := models.GetSystemCurrency()
        if err != nil {
            fmt.Printf("获取系统币种列表失败")
            return
        }else{
            for _, curr := range currs {
                codeDigitsMap[curr.Code] = curr.Digits
            }       
        }

        fmt.Printf("GetSystemCurrency, codeDigitsMap=%v\n", codeDigitsMap)
    }
}

func (m *MerchantApiModule) GetBalance(queueId int64, userId, code, merchantKey string) (error, float64) {
    var (
        balance float64
        data    *ApiResponse
        errRet  error
    )

    fmt.Printf("MerchantApiModule GetBalance, queueId=%d, userId=%s, code=%s, merchantKey=%s---------------------\n", queueId, userId, code, merchantKey)

    done := make(chan struct{})

    m.AsyncDoByQueue(queueId, func() bool {
        defer close(done)

        var err error

        data, err = m.FetchBalanceFromMerchant(userId, code, merchantKey)

        if err != nil {
            fmt.Printf("从商户获取余额失败: %v\n", err)
            errRet = fmt.Errorf("获取余额失败: %w", err)
            return false
        }

        if data.Code == 200 {
            var getBalanceData GetBalanceDataReturn

            err := json.Unmarshal(data.Data, &getBalanceData)
            if err == nil {
                balance = getBalanceData.Balance
            } else {
                fmt.Printf("Unmarshal getBalanceData 回调错误: %v\n", err)
                errRet = fmt.Errorf("余额数据解析失败: %w", err)
                return false
            }
        } else {
            errRet = fmt.Errorf("商户返回错误码: %d", data.Code)
            return false
        }

        return true
    }, nil)

    select {
    case <-done:
    case <-time.After(30 * time.Second):
        fmt.Printf("GetBalance() 等待商户余额响应超时\n")
        errRet = fmt.Errorf("等待商户余额响应超时")
    }

    return errRet, balance
}

type Merchant struct {
    MerchantCode       string
    MerchantName       string

    WalletUrl          string
    OrToken            string
}

// FetchBalanceFromMerchant 调用商户接口获取玩家实时余额
func (m *MerchantApiModule) FetchBalanceFromMerchant(username, code, merchantKey string) (*ApiResponse, error) {
    const (
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

        fmt.Println("===========数据库商户信息:%v", merchantInfo)

        merchant = &Merchant{
            MerchantCode: merchantInfo.MerchantCode,
            MerchantName: merchantInfo.MerchantName,
            OrToken:      merchantInfo.OrToken,
            WalletUrl:    merchantInfo.WalletUrl,
        }

        m.mapMerchant[merchantKey] = merchant

    }

    fmt.Printf("username=%s, merchantKey=%s\n", username, merchant.MerchantCode)
    cutName := getUsernamePrefix(username, merchant.MerchantCode)

    // 请求数据
    reqBody := GetBalanceRequest{
        Username: cutName,
        Code:     code,
    }

    // 构建 HTTP 请求
    fmt.Printf("FetchBalanceFromMerchant, reqBody=%v\n", reqBody)

    requestUrl := fmt.Sprintf("%s%s", merchant.WalletUrl, "/single-wallet/get-balance/")

    fmt.Printf("requestUrl=%s\n", requestUrl)

    bodyBytes, _ := json.Marshal(reqBody)
    req, err := http.NewRequest("POST", requestUrl, bytes.NewBuffer(bodyBytes))

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

func getUsernamePrefix(username, merchantKey string) string {
    username = strings.ToLower(username)
    prefix := strings.ToLower(merchantKey) + "_"
    if strings.HasPrefix(username, prefix) {
        return username[len(prefix):]
    }
    return username
}


