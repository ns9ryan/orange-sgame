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
    "errors"
)

//根据dataType得到tranId的订单类型备注
var TranIdDataTypeMap = map[int32]string{
    0: "bet",           //下注订单
    1: "cancel_bet",    //小费订单
    2: "tip",           //结算订单
    3: "win",           //取消下注订单
    4: "other",         //其他订单
}

// MerchantBalanceRequest 商户余额请求参数
type CreditDebitRequest struct {
    Username          string        `json:"username"`
    Code              string        `json:"currency"` //币种，如 USD、CNY 等

    GameTableId       int32         `json:"game_table_id"`    //4
    GameRoundId       string        `json:"game_round_id"`    //"jPNX07jf_xU7VV6IQPZrr"
    OrangeTransType   string        `json:"orange_trans_type"`//动作备注：bet或者tip或者win
    OrangeTrasnTime   string        `json:"orange_trans_time"`//"2025-03-18T14:23:54+08:00"
    OrangeTransId     string        `json:"orange_trans_id"`  //"DTX9m_edKSfsTDLo_SXCxaVz"
    Amount            float64       `json:"amount"`           //0.5000
}

type CreditDebitDataReturn struct {
    ActionType        string        `json:"action_type"`      //动作类型，debit或者credit
    Amount            float64       `json:"amount"`           //0.5000
    Balance           float64       `json:"balance"`          //130.5000
    Code              string        `json:"currency"`             //币种，如 USD、CNY 等

    MerchantTrasnTime string        `json:"merchant_trans_time"`//"2025-03-18T14:23:54+08:00"
    MerchantTransId   string        `json:"merchant_trans_id"`  //"DTX9m_edKSfsTDLo_SXCxaVz"

    OrangeTrasnTime   string        `json:"orange_trans_time"`//"2025-03-18T14:23:54+08:00"
    OrangeTransId     string        `json:"orange_trans_id"`  //"DTX9m_edKSfsTDLo_SXCxaVz"
}

func (m *MerchantApiModule) BalanceDec(queueId int64, request *rpc.WalletOptReq, player *player.Player) (error, float64) {
    var (
        balance float64         // 保存返回的余额
        data    *ApiResponse    // 保存商户返回的数据
        errRet  error           // 最终要返回的 error
    )

    fmt.Printf("MerchantApiModule BalanceDec, request=%v, player=%v\n", queueId, request, player)

    done := make(chan struct{}) // 用于通知主 goroutine 执行完成

    // 发起异步执行
    m.AsyncDoByQueue(queueId, func() bool {
        defer close(done)

        var err error
        data, err = m.DebitFromMerchant(request, player)

        fmt.Printf("BalanceDec AsyncDoByQueue =%v\n", data.Code)

        if err != nil {
            fmt.Printf("从商户扣款失败: %v\n", err)
            errRet = fmt.Errorf("商户扣款失败: %w", err)
            return false
        }

        if data.Code == 200 {
            var debitDataReturn CreditDebitDataReturn
            err := json.Unmarshal(data.Data, &debitDataReturn)
            if err == nil {
                balance = debitDataReturn.Balance
            } else {
                fmt.Printf("Unmarshal debitDataReturn 回调错误: %v\n", err)
                errRet = fmt.Errorf("数据解析失败: %w", err)
                return false
            }
        } else {
            errRet = fmt.Errorf("商户返回错误码: %d", data.Code)
            return false
        }

        return true
    }, nil)

    // 等待异步任务完成，或超时
    select {
    case <-done:
    case <-time.After(30 * time.Second):
        fmt.Printf("DebitFromMerchant() 等待商户余额响应超时\n")
        errRet = fmt.Errorf("等待商户响应超时")
    }

    return errRet, balance
}

// FetchBalanceFromMerchant 调用商户接口获取玩家实时余额
func (m *MerchantApiModule) DebitFromMerchant(request *rpc.WalletOptReq, player *player.Player) (*ApiResponse, error) {
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
    reqBody := CreditDebitRequest{
        Username:           cutName,
        Code:               player.Code,
        GameTableId:        request.DeskNo,
        GameRoundId:        request.RoundId,
        OrangeTransType:    GetTranIdAction(request.DataType),
        OrangeTrasnTime:    time.Now().UTC().Format("2006-01-02T15:04:05+00:00"),
        OrangeTransId:      request.TranId,
        Amount:             float64(request.Point),
    }

    fmt.Printf("DebitFromMerchant, reqBody=%v\n", reqBody)

    // 构建 HTTP 请求
    // 构建 HTTP 请求
    requestUrl := fmt.Sprintf("%s%s", merchant.WalletUrl, "/single-wallet/debit/")
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
    fmt.Printf("扣款return resp.Body=%v\n", resp.Body)

    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }

    if result.Code != 200 {
        fmt.Printf("result.Msg=%s\n", result.Msg)
        return &result, errors.New(result.Msg)
    }

    return &result, nil
}

func (m *MerchantApiModule) BalanceAdd(queueId int64, request *rpc.WalletOptReq, player *player.Player) (error, float64) {
    var (
        balance float64
        data    *ApiResponse
        errRet  error
    )

    fmt.Printf("MerchantApiModule BalanceAdd, request=%v, player=%v\n", queueId, request, player)

    done := make(chan struct{})

    m.AsyncDoByQueue(queueId, func() bool {
        defer close(done)

        var err error
        data, err = m.CreditFromMerchant(request, player)

        if err != nil {
            fmt.Printf("从商户加点失败: %v\n", err)
            errRet = fmt.Errorf("商户加点失败: %w", err)
            return false
        }

        if data.Code == 200 {
            var debitDataReturn CreditDebitDataReturn

            err := json.Unmarshal(data.Data, &debitDataReturn)
            if err == nil {
                balance = debitDataReturn.Balance
            } else {
                fmt.Printf("Unmarshal debitDataReturn 回调错误: %v\n", err)
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
        fmt.Printf("CreditFromMerchant() 等待商户余额响应超时\n")
        errRet = fmt.Errorf("等待商户加点响应超时")
    }

    return errRet, balance
}

// FetchBalanceFromMerchant 调用商户接口获取玩家实时余额
func (m *MerchantApiModule) CreditFromMerchant(request *rpc.WalletOptReq, player *player.Player) (*ApiResponse, error) {
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
    reqBody := CreditDebitRequest{
        Username:           cutName,
        Code:               player.Code,
        GameTableId:        request.DeskNo,
        GameRoundId:        request.RoundId,
        OrangeTransType:    GetTranIdAction(request.DataType),
        OrangeTrasnTime:    time.Now().UTC().Format("2006-01-02T15:04:05+00:00"),
        OrangeTransId:      request.TranId,
        Amount:             float64(request.Point),
    }

    fmt.Printf("CreditFromMerchant, reqBody=%v\n", reqBody)

    // 构建 HTTP 请求
    requestUrl := fmt.Sprintf("%s%s", merchant.WalletUrl, "/single-wallet/credit/")
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
    fmt.Printf("加款return resp.Body=%v\n", resp.Body)

    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }

    if result.Code != 200 {
        fmt.Printf("result.Msg=%s\n", result.Msg)
        return nil, fmt.Errorf("商户返回错误: %s", result.Msg)
    }

    return &result, nil
}
