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
    //"errors"
)

// OpenAnswerRequest 通知商户结算请求参数
type OpenAnswerRequest struct {
    GameTableId       int32         `json:"game_table_id"`    //4
    GameRoundId       string        `json:"game_round_id"`    //对局ID
}

func (m *MerchantApiModule) OpenAnswer(queueId int64, request *rpc.WalletOptReq, player *player.Player) (error) {
    var (
        data    *ApiResponse
        errRet  error
    )

    fmt.Printf("MerchantApiModule OpenAnswer, request=%v, player=%v\n", queueId, request, player)

    done := make(chan struct{})

    m.AsyncDoByQueue(queueId, func() bool {
        defer close(done)

        var err error
        data, err = m.AnswerForMerchant(request, player)

        if err != nil {
            fmt.Printf("通知商户结算失败: %v\n", err)
            errRet = fmt.Errorf("通知商户结算失败: %w", err)
            return false
        }

        if data.Code != 200 {
            errRet = fmt.Errorf("商户返回错误码: %d", data.Code)
            return false
        }

        return true
    }, nil)

    select {
    case <-done:
    case <-time.After(30 * time.Second):
        fmt.Printf("AnswerForMerchant() 等待商户响应超时\n")
        errRet = fmt.Errorf("等待商户加点响应超时")
    }

    return errRet
}

// FetchBalanceFromMerchant 调用商户接口获取玩家实时余额
func (m *MerchantApiModule) AnswerForMerchant(request *rpc.WalletOptReq, player *player.Player) (*ApiResponse, error) {
    const (
        requestTimeout  = 30 * time.Second
    )

    fmt.Println("MerchantApiModule AnswerForMerchant---------------------")

    _, ok := m.RoundAnswer[request.RoundId]
    if !ok {
        m.RoundAnswer[request.RoundId] = request.DeskNo
    }else{
        fmt.Printf("商户的其他玩家触发过Answer，跳过:%s\n", request.RoundId)

        if len(m.RoundAnswer) >= 2048 {
        // 已满，清空
            m.RoundAnswer = make(map[string]int32, 2048)
        }

        var result ApiResponse
        result.Code = 200
        result.Msg = ""

        return &result, nil
    }

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

    // 请求数据
    reqBody := OpenAnswerRequest{
        GameTableId:        request.DeskNo,
        GameRoundId:        request.RoundId,
    }

    // 构建 HTTP 请求
    requestUrl := fmt.Sprintf("%s%s", merchant.WalletUrl, "/game-round/settle/")

    fmt.Printf("AnswerForMerchant, reqBody=%v, requestUrl=%s\n", reqBody, requestUrl)

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

    fmt.Printf("AnswerForMerchant return result=%v\n", result)

    if result.Code != 200 {
        fmt.Printf("result.Msg=%s\n", result.Msg)
        return nil, fmt.Errorf("商户返回错误: %s", result.Msg)
    }

    return &result, nil
}


// 根据 dataType 获取动作类型，如果不存在则返回 "other"
func GetTranIdAction(dataType int32) string {
    if action, ok := TranIdDataTypeMap[dataType]; ok {
        return action
    }
    return "other"
}
