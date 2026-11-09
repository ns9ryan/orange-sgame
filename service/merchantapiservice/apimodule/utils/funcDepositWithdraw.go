//*****************************************************************
// funcDepositWithdraw.go里，放的是对玩家余额上分/下分操作的基础处理函数。
//*****************************************************************

package apifunc

import (
    "sync"
    "fmt"
    //"videogame/common/global/variable"
    "github.com/shopspring/decimal"
    "encoding/json"
    "net/http"
    "io/ioutil"
    //"time"
)

// 使用全局锁来防止并发访问相同用户的余额
var userLocks sync.Map

// 管理系统配置的基本静态汇率表：以人民币 CNY 为基准货币
var Rate = map[string]float64{
    "CNY": 1.0,         // 人民币本身
    "KRW": 180.0,       // 韩元（1 CNY = 180 KRW）
    "MYR": 0.66,        // 马来西亚币（1 CNY = 0.66 MYR）
    "USD": 0.15,        // 美元（1 CNY = 0.15 USD）
    "JPY": 17.5,        // 日元（1 CNY = 17.5 JPY）
    "THB": 5.1,         // 泰铢（1 CNY = 5.1 THB）
    "BTC": 0.000022,    // 比特币（1 CNY = 0.000022 BTC）
    "VND": 3800.0,      // 越南盾（1 CNY = 3800 VND）
    "EUR": 0.13,        // 欧元（1 CNY = 0.13 EUR）
    "HKD": 1.2,         // 港币（1 CNY = 1.2 HKD）
    "AUD": 0.21,        // 澳元（1 CNY = 0.21 AUD）
    "TWD": 4.5,         // 台币（1 CNY = 4.5 TWD）
    "IDR": 2250.0,      // 印尼盾（1 CNY = 2250 IDR）
    "INR": 11.5,        // 印度卢比（1 CNY = 11.5 INR）
    "BND": 0.21,        // 文莱币（1 CNY = 0.21 BND）
}

// 请求实时汇率API响应结构：
type ExchangeRateResponse struct {
    Result          string            `json:"result"`
    ConversionRates map[string]float64 `json:"conversion_rates"`
}

// 用于从外部API获取实时汇率，并更新Rate map;
// 平台记录的所有币种汇率，如有更改会通知相关商户，且极少变更.
func GetExchangeRates() error {
    // 以访问 ExchangeRate-API 为例
    // 假设已经注册并获得了一个 API 密钥，替换API_KEY
    url := "https://v6.exchangerate-api.com/v6/API_KEY/latest/CNY" // 替换真实API_KEY

    // 发送 HTTP 请求
    resp, err := http.Get(url)
    if err != nil {
        return fmt.Errorf("failed to get exchange rates: %v", err)
    }
    defer resp.Body.Close()

    // 读取响应体
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return fmt.Errorf("failed to read response body: %v", err)
    }

    // 解析 JSON 数据
    var result ExchangeRateResponse
    err = json.Unmarshal(body, &result)
    if err != nil {
        return fmt.Errorf("failed to parse JSON response: %v", err)
    }

    // 如果响应结果是 "error"，则说明请求失败
    if result.Result != "success" {
        return fmt.Errorf("failed to get valid exchange rates")
    }

    // 更新汇率表，只更新我们关心的货币
    for _, currency := range []string{"KRW", "MYR", "USD", "JPY", "THB", "BTC", "VND", "EUR", "HKD", "AUD", "TWD", "IDR", "INR", "BND"} {
        if rate, exists := result.ConversionRates[currency]; exists {
            Rate[currency] = rate
        }
    }

    return nil
}

// 存款上分
func DepositFunc(transferNo, userName string, amount decimal.Decimal) (decimal.Decimal, error) {
    // 获取用户锁，防止并发冲突
    lock, ok := userLocks.LoadOrStore(userName, &sync.Mutex{})
    if !ok {
        fmt.Println("A new lock was created for user:", userName)
    }

    userLock := lock.(*sync.Mutex)
    userLock.Lock()
    defer userLock.Unlock()

    // 检查存款金额是否合理
    newBalance := decimal.Zero
    if amount.LessThanOrEqual(decimal.Zero) {
        return newBalance, fmt.Errorf("存款金额必须大于0")
    }

    // 定义数据库事务
    // tx := variable.GormDbMysql.Begin()
    // defer func() {
    //     if r := recover(); r != nil {
    //         // 如果发生 panic，进行回滚
    //         tx.Rollback()
    //     }
    // }()

    // var user models.User
    // getUser, err := user.GetBalanceByUserName(tx, userName)
    // if err != nil {
    //     tx.Rollback()
    //     return newBalance, fmt.Errorf("获取用户余额失败: %w", err)
    // }

    // if getUser == nil {
    //     tx.Rollback()
    //     return newBalance, fmt.Errorf("用户 %s 不存在", userName)
    // }

    // // 更新用户余额
    // newBalance = getUser.Wallet.Add(amount)
    // err = user.UpdateBalanceTx(tx, userName, newBalance) // 修复错误变量的重名问题
    // if err != nil {
    //     tx.Rollback()
    //     return newBalance, fmt.Errorf("更新余额失败: %w", err)
    // }

    // 保存新交易单号到数据库
    // if transferNo != "" {
    //     newMerchantTranfer := &models.MerchantTransfer{
    //         TransferNo:   transferNo,
    //         Amount:       &amount,
    //         CreatedAt:    time.Now(),
    //         UpdatedAt:    time.Now(),
    //     }

    //     if err := newMerchantTranfer.CreateTx(tx); err != nil {
    //         tx.Rollback()
    //         return newBalance, fmt.Errorf("保存交易单号失败: %v", err)
    //     }
    // }

    // 提交事务，确保数据库状态更新
    // if tx.Error == nil {
    //     tx.Commit()
    // } else {
    //     tx.Rollback()
    //     return newBalance, fmt.Errorf("failed to update user balance")
    // }

    // 提交事务后，处理其他异步操作，比如清除Redis里的余额缓存等
    // ...

    return newBalance, nil
}


// 取款下分
func WithdrawFunc(transferNo, userName string, amount decimal.Decimal) (decimal.Decimal, error) {
    // 获取用户锁，防止并发冲突
    lock, ok := userLocks.LoadOrStore(userName, &sync.Mutex{})
    if !ok {
        fmt.Println("A new lock was created for user:", userName)
    }

    userLock := lock.(*sync.Mutex)
    userLock.Lock()
    defer userLock.Unlock()

    // 检查取款金额是否合理
    newBalance := decimal.Zero
    if amount.LessThanOrEqual(decimal.Zero) {
        return newBalance, fmt.Errorf("取款金额必须大于0")
    }

    // 定义数据库事务
    // tx := variable.GormDbMysql.Begin()
    // defer func() {
    //     if r := recover(); r != nil {
    //         // 如果发生 panic，进行回滚
    //         tx.Rollback()
    //     }
    // }()

    // var user models.User
    // // 查询当前用户余额，使用行级锁来确保余额一致性
    // getUser, err := user.GetBalanceByUserName(tx, userName)
    // if err != nil {
    //     tx.Rollback()
    //     return newBalance, fmt.Errorf("获取用户余额失败: %w", err)
    // }

    // if getUser == nil {
    //     tx.Rollback()
    //     return newBalance, fmt.Errorf("用户 %s 不存在", userName)
    // }

    // // 检查余额是否足够
    // if getUser.Wallet.LessThan(amount) {
    //     tx.Rollback()
    //     return newBalance, fmt.Errorf("玩家余额不足")
    // }

    // // 更新用户余额
    // newBalance = getUser.Wallet.Sub(amount)
    // err = user.UpdateBalanceTx(tx, userName, newBalance)
    // if err != nil {
    //     tx.Rollback()
    //     return newBalance, fmt.Errorf("更新余额失败: %w", err)
    // }

    // 保存新交易单号到数据库
    // if transferNo != "" {
    //     newMerchantTranfer := &models.MerchantTransfer{
    //         TransferNo:   transferNo,
    //         Amount:       &amount,
    //         CreatedAt:    time.Now(),
    //         UpdatedAt:    time.Now(),
    //     }

    //     if err := newMerchantTranfer.CreateTx(tx); err != nil {
    //         tx.Rollback()
    //         return newBalance, fmt.Errorf("保存交易单号失败: %v", err)
    //     }
    // }

    // 提交事务，确保数据库状态更新
    // if tx.Error == nil {
    //     tx.Commit()
    // } else {
    //     tx.Rollback()
    //     return newBalance, fmt.Errorf("failed to update user balance")
    // }

    // 提交事务后，处理其他异步操作，比如清除Redis里的余额缓存等
    // ...

    return newBalance, nil
}

func Float64ToDecimal(value float64) decimal.Decimal {
    // 将 float64 转换为 decimal.Decimal
    dec := decimal.NewFromFloat(value)
    
    // 设置为四位小数
    dec = dec.Round(4)
    
    return dec
}

func IntToDecimal(value int) decimal.Decimal {
    // 将 int 转换为 decimal.Decimal
    dec := decimal.NewFromInt(int64(value))
    
    // 如果需要设置为四位小数，可以手动调整
    dec = dec.Round(4)
    
    return dec
}