package merchantapicall

import (
    "encoding/json"
    "time"
    "net/http"
    "bytes"
    "fmt"
    models "videogame/common/model"
    "videogame/common/proto/rpc"
    "videogame/service/singlewalletservice/player"
)

// MerchantBalanceRequest 商户余额请求参数
type CancelDebitRequest struct {
    Username          string        `json:"username"`
    Code              string        `json:"currency"` //币种，如 USD、CNY 等

    GameTableId       int32         `json:"game_table_id"`    //4
    GameRoundId       string        `json:"game_round_id"`    //"jPNX07jf_xU7VV6IQPZrr"
    OrangeTransType   string        `json:"orange_trans_type"`//动作备注：bet或者tip或者win
    OrangeTrasnTime   string        `json:"orange_trans_time"`//"2025-03-18T14:23:54+08:00"
    OrangeTransId     string        `json:"orange_trans_id"`  //"DTX9m_edKSfsTDLo_SXCxaVz"
    Amount            float64       `json:"amount"`           //0.5000
    Reason            string        `json:"reason"`           //取消扣款的原因，是玩家主动撤回，还是下注网络没响应取消，或者游戏服异常取消等
}

type CancelDebitDataReturn struct {
    Balance           float64       `json:"balance"`          //130.5000
    Code              string        `json:"currency"`             //币种，如 USD、CNY 等
}

func (m *MerchantApiModule) BalanceCancelBet(queueId int64, request *rpc.WalletOptReq, player *player.Player) (error, float64) {
    var (
        balance float64
        data    *ApiResponse
        errRet  error
    )

    fmt.Printf("MerchantApiModule BalanceCancelBet, request=%v\n", queueId, request)

    done := make(chan struct{})

    m.AsyncDoByQueue(queueId, func() bool {
        defer close(done)

        var err error
        data, err = m.CancelDebitFromMerchant(request, player)

        if err != nil {
            fmt.Printf("从商户加点失败: %v\n", err)
            errRet = fmt.Errorf("商户加点失败: %w", err)
            return false
        }

        if data.Code == 200 {
            var cancelDataReturn CancelDebitDataReturn

            err := json.Unmarshal(data.Data, &cancelDataReturn)
            if err == nil {
                balance = cancelDataReturn.Balance
            } else {
                fmt.Printf("Unmarshal cancelDataReturn 回调错误: %v\n", err)
                errRet = fmt.Errorf("加点数据解析失败: %w", err)
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
        fmt.Printf("CancelDebitFromMerchant() 等待商户余额响应超时\n")
        errRet = fmt.Errorf("等待商户加点响应超时")
    }

    return errRet, balance
}

// CancelDebitFromMerchant
func (m *MerchantApiModule) CancelDebitFromMerchant(request *rpc.WalletOptReq, player *player.Player) (*ApiResponse, error) {
    const (
        requestTimeout  = 30 * time.Second
    )

    fmt.Println("MerchantApiModule FetchBalanceFromMerchant---------------------")

    merchant, ok := m.mapMerchant[player.MerchantKey]
    if !ok {
        fmt.Printf("商户不在map里，用 merchantKey查询...\n")

        merchantInfo, err := models.GetMerchantByName(player.MerchantKey)
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

        m.mapMerchant[player.MerchantKey] = merchant

    }

    fmt.Printf("username=%s, merchantKey=%s\n", request.UserId, merchant.MerchantCode)
    cutName := getUsernamePrefix(request.UserId, merchant.MerchantCode)

    // 请求数据
    reqBody := CancelDebitRequest{
        Username:           cutName,
        Code:               player.Code,
        GameTableId:        request.DeskNo,
        GameRoundId:        request.RoundId,
        OrangeTransType:    "bet",
        OrangeTrasnTime:    time.Now().UTC().Format("2006-01-02T15:04:05+00:00"),
        OrangeTransId:      request.TranId,
        Amount:             float64(request.Point),
        Reason:             "Player canceled the bet.",
    }

    fmt.Printf("CancelDebitFromMerchant, reqBody=%v\n", reqBody)

    // 构建 HTTP 请求
    requestUrl := fmt.Sprintf("%s%s", merchant.WalletUrl, "/single-wallet/cancel-debit/")
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
        fmt.Printf("result.Msg=%s\n", result.Msg)
        return nil, fmt.Errorf("商户返回错误: %s", result.Msg)
    }

    return &result, nil
}

//取消下注3秒重试三次的请求，测试时备用
func (m *MerchantApiModule) CancelDebitFromMerchantTriple(request *rpc.WalletOptReq, player *player.Player) (*ApiResponse, error) {
    const (
        merchantURL    = "http://127.0.0.1:9004/merchant/api/v1/single_wallet/cancel_debit/"
        orToken        = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHQiOjE3NDY3Nzc4NTAsImlhdCI6MTc0NDE4NTg1MCwibWVyY2hhbnRfY29kZSI6IlZFWlVYNzdTIn0.aHyl32p5pvhhM_fnfzx7NQxWwq1QsuC9DaqAO7Mi0eM"
        requestTimeout = 30 * time.Second
        maxRetries     = 3
    )

    fmt.Println("MerchantApiModule CancelDebitFromMerchant---------------------")

    merchant, ok := m.mapMerchant[player.MerchantKey]
    if !ok {
        fmt.Printf("商户不在map里，用 merchantKey查询...\n")
        merchantInfo, err := models.GetMerchantByName(player.MerchantKey)
        if err != nil {
            return nil, err
        }

        fmt.Printf("===========数据库商户信息: %+v\n", merchantInfo)

        merchant = &Merchant{
            MerchantCode: merchantInfo.MerchantCode,
            MerchantName: merchantInfo.MerchantName,
            OrToken:      orToken,
            WalletUrl:    merchantURL,
        }

        m.mapMerchant[player.MerchantKey] = merchant
    }

    // 请求数据
    reqBody := CreditDebitRequest{
        Username:         request.UserId,
        Code:             player.Code,
        GameTableId:      request.DeskNo,
        GameRoundId:      "-",
        OrangeTransType:  GetTranIdAction(request.DataType),
        OrangeTrasnTime:  time.Now().UTC().Format("2006-01-02T15:04:05+00:00"),
        OrangeTransId:    request.TranId,
        Amount:           float64(request.Point),
    }

    bodyBytes, _ := json.Marshal(reqBody)

    var lastErr error
    for attempt := 1; attempt <= maxRetries; attempt++ {
        fmt.Printf("第 %d 次请求: %v\n", attempt)

        req, err := http.NewRequest("POST", merchant.WalletUrl, bytes.NewBuffer(bodyBytes))
        if err != nil {
            return nil, err
        }
        req.Header.Set("Content-Type", "application/json")
        req.Header.Set("OrToken", merchant.OrToken)

        client := &http.Client{Timeout: requestTimeout}
        resp, err := client.Do(req)
        if err != nil {
            fmt.Printf("第 %d 次请求失败: %v\n", attempt, err)
            lastErr = err
            time.Sleep(1 * time.Second)
            continue
        }
        defer resp.Body.Close()

        var result ApiResponse
        if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
            fmt.Printf("第 %d 次响应解析失败: %v\n", attempt, err)
            lastErr = err
            time.Sleep(1 * time.Second)
            continue
        }

        if result.Code != 200 {
            fmt.Printf("第 %d 次返回错误，code: %d, msg: %s\n", attempt, result.Code, result.Msg)
            lastErr = fmt.Errorf("商户返回错误: %s", result.Msg)
            time.Sleep(1 * time.Second)
            continue
        }

        // 成功
        return &result, nil
    }

    // 全部尝试失败
    return nil, fmt.Errorf("CancelDebitFromMerchant: 尝试 %d 次请求后仍失败, 最后错误: %v", maxRetries, lastErr)
}
