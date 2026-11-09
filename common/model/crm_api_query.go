package models

import (
    "gorm.io/plugin/dbresolver"
    "videogame/common/global/variable"
    "time"
)

// TransactionType 交易类型
type TransactionType uint

// PlayTransaction 玩家交易流水表结构体
type SPlayTransaction struct {
    ID              uint            `gorm:"primarykey;type:int unsigned;comment:ID"`
    Username        string          `gorm:"not null;comment:玩家账号,每个玩家的唯一值"`
    Code            string          `gorm:"type:varchar(10);not null;comment:币种"`
    Digits          uint            `gorm:"type:int unsigned;not null;comment:币种的最小单位倍数"`
    TransactionType TransactionType `gorm:"type:tinyint unsigned;not null;comment:交易类型:1充值,2提现,3下注,4中奖,5返水,6调整"`
    Before          uint            `gorm:"type:int unsigned;not null;default:0;comment:变动前余额,存储为最小单位"`
    Amount          int             `gorm:"type:int;not null;default:0;comment:变动金额,存储为最小单位"`
    After           uint            `gorm:"type:int unsigned;not null;default:0;comment:变动后余额,存储为最小单位"`
    OrangeTransId   string          `gorm:"size:32;not null;comment:系统产生的订单号,全局唯一"`
    MerchantTransId string          `gorm:"size:32;not null;default:'';comment:商户产生的订单号,api商户传过来的订单号"`
    ReferenceId     string          `gorm:"size:32;not null;default:'';comment:关联的业务ID(如订单号/投注ID),用来以后扩展关联具体的详细信息表"`
    Remark          string          `gorm:"not null;default:'';comment:备注"`
}

// TableName 重写 GORM 的默认表名
func (SPlayTransaction) TableName() string {
    return "or_play_transaction"
}

// GameCategory  游戏分类表结构体
type SGameCategory struct {
    ID               uint   `gorm:"primarykey;type:int unsigned;comment:ID"`
    GameCategoryCode string `gorm:"size:32;not null;comment:分类"`
    SubCode          string `gorm:"uniqueindex;size:32;not null;comment:子分类"`
    Description      string `gorm:"size:100;not null;default:'';comment:描述"`
    Status           uint   `gorm:"type:tinyint unsigned;not null;default:1;comment:是否启用:0否,1是"`
    IsHide           uint           `gorm:"type:tinyint unsigned;not null;default:0;comment:是否隐藏:0否,1是"`
}

// TableName 重写 GORM 的默认表名
func (SGameCategory) TableName() string {
    return "or_game_category"
}

// WalletType 钱包类型
type WalletType uint

// Merchant 商户主表结构体
type SMerchant struct {
    ID           uint   `gorm:"primarykey;type:int unsigned;comment:ID"`
    MerchantCode string `gorm:"uniqueindex;size:10;not null;comment:商户号"`
    MerchantName string `gorm:"size:32;not null;comment:商户名称"`
    ParentId     string `gorm:"size:32;not null;default:'';comment:上级商户的username"`
    Username     string `gorm:"uniqueindex;size:32;not null;comment:账号"`
    Nickname     string `gorm:"size:32;not null;comment:昵称"`
    Password     string `gorm:"not null;comment:密码"`

    CommissionRate uint `gorm:"type:tinyint unsigned;not null;default:0;comment:成数"`
    RollingRate    uint `gorm:"type:tinyint unsigned;not null;default:0;comment:洗码比"`

    Status     uint       `gorm:"type:tinyint unsigned;not null;default:1;comment:是否允许登录:0否,1是"`
    IsActive   uint       `gorm:"type:tinyint unsigned;not null;default:0;comment:是否激活(第一次登录成功后改成激活):0否,1是"`
    UseGoogle  uint       `gorm:"type:tinyint unsigned;not null;default:0;comment:是否启用谷歌登录验证:0否,1是"`
    BetPaused  uint       `gorm:"type:tinyint unsigned;not null;default:0;comment:是否暂停投注:0正常,1暂停"`
    IsMain     uint       `gorm:"type:tinyint unsigned;not null;default:1;comment:是否主账号:0否,1是"`
    WalletType WalletType `gorm:"type:tinyint unsigned;not null;default:2;comment:钱包类型:1单钱包,2多钱包"`

    LoginTime *time.Time `gorm:"default:NULL;comment:上次登录时间"`
    LoginIp   string     `gorm:"size:50;comment:登录IP"`

    Chips         string `gorm:"not null;comment:筹码"`
    BettingLimits string `gorm:"not null;comment:限红"`
    ChipsTemplate string `gorm:"not null;comment:筹码模板"`
    RoleId        uint   `gorm:"type:int unsigned;not null;default:3;comment:角色ID"`

    // api相关
    AesKeySalt string `gorm:"not null;comment:aesKey盐"`
    Md5Salt    string `gorm:"not null;comment:md5盐"`
    OrToken    string `gorm:"not null;comment:商户api调用验证or_token"`
    OrTokenKey string `gorm:"not null;comment:生成token用的商户唯一key，永久有效"`
    WalletUrl  string `gorm:"not null;default:'';comment:单钱包URL"`
    WhiteList  string `gorm:"not null;default:'';comment:白名单IP,逗号分隔"`
}

func (SMerchant) TableName() string {
    return "or_merchant"
}

// SystemCurrency 系统币种表结构体
type SSystemCurrency struct {
    ID          uint   `gorm:"primarykey;type:int unsigned;comment:ID"`
    Code        string `gorm:"type:varchar(10);not null;unique;comment:字母编码"`
    Name        string `gorm:"type:varchar(32);not null;comment:名称"`
    Symbol      string `gorm:"type:varchar(10);not null;comment:符号"`
    Digits      uint   `gorm:"type:int unsigned;not null;comment:币种的最小单位倍数"`
    Number      string `gorm:"type:varchar(10);not null;comment:数字编码"`
    Region      string `gorm:"size:20;not null;comment:国家/地区/组织"`
    Description string `gorm:"size:20;not null;comment:备注"`
}

func (SSystemCurrency) TableName() string {
    return "or_system_currency"
}

// SystemCurrency 系统币种表结构体
type SBetArea struct {
    ID               uint   `gorm:"primarykey;type:int unsigned;comment:ID"`
    GameCategoryCode string `gorm:"size:32;not null;comment:游戏分类"`
    AreaId           string `gorm:"uniqueindex;size:10;not null;not null;comment:区域ID"`
    AreaKey          string `gorm:"size:50;not null;default:'';comment:区域标识"`
    AreaName         string `gorm:"size:50;not null;default:'';comment:区域名称"`
    Description      string `gorm:"size:100;not null;default:'';comment:下注区域说明"`
    Odds             string `gorm:"size:100;not null;default:'';comment:赔率"`
    MinBetAmount     uint   `gorm:"not null;default:0;comment:最小投注金额"`
    MaxBetAmount     uint   `gorm:"not null;default:0;comment:最大投注金额"`
    Digits           uint   `gorm:"type:int unsigned;not null;default:1;comment:投注的最小单位倍数"`
}

func (SBetArea) TableName() string {
    return "or_bet_area"
}

// PlayerType 玩家类型
type PlayerType uint

// PlayerLevel 玩家级别
type PlayerLevel uint

// PlayerCreateType 玩家来源
type PlayerCreateType uint

// PlayerMoneyType 玩家模式类型（真钱/试玩）
type PlayerMoneyType uint

// PlayerDeviceType 设备类型
type PlayerDeviceType uint

type SPlayer struct {
    ID                 uint   `gorm:"primarykey;type:int unsigned;comment:ID"`
    MerchantId         string `gorm:"size:32;not null;comment:对应商户表的username"`
    Currency           string `gorm:"type:varchar(10);not null;comment:币种"`
    MerchantCurrencyId string `gorm:"size:50;not null;comment:商户币种唯一标识"`
    ApiUser            string `gorm:"not null;comment:存储api会员原始账号"`
    Username           string `gorm:"uniqueindex;size:32;not null;comment:账号"`
    Nickname           string `gorm:"size:32;not null;default:'';comment:昵称"`
    ParentId           string `gorm:"size:32;not null;default:'';comment:上级会员的的username"`

    LoginTime *time.Time `gorm:"default:NULL;comment:上次登录时间"`
    LoginIp   string     `gorm:"size:45;comment:登录IP"`

    Status     uint             `gorm:"type:tinyint unsigned;not null;default:1;comment:是否冻结:0冻结,1正常"`
    BetPaused  uint             `gorm:"type:tinyint unsigned;not null;default:0;comment:是否暂停投注:0正常,1暂停"`
    IsOnline   uint             `gorm:"type:tinyint unsigned;not null;default:0;comment:是否在线:0否,1是"`
    PlayerType PlayerType       `gorm:"type:tinyint unsigned;not null;default:1;comment:玩家类型:1现金代理,2佣金代理"`
    MoneyType  PlayerMoneyType  `gorm:"type:tinyint unsigned;not null;default:0;comment:玩家模式类型:0真钱,1试玩"`
    CreateType PlayerCreateType `gorm:"type:tinyint unsigned;not null;default:1;comment:玩家来源:1后台创建,2API创建"`
    DeviceType PlayerDeviceType `gorm:"type:tinyint unsigned;not null;default:0;comment:设备类型:0桌面H5横屏,1手机H5竖屏"`
    Level      PlayerLevel      `gorm:"type:tinyint unsigned;not null;default:1;comment:玩家级别"`
    MaxWin     uint             `gorm:"type:int unsigned;not null;default:0;comment:最高赢额,0表示不限制"`

    Email         string `gorm:"size:100;not null;default:'';comment:邮箱"`
    Phone         string `gorm:"size:20;not null;default:'';comment:手机号"`
    Password      string `gorm:"size:80;not null;default:'';comment:密码"`
    Chips         string `gorm:"not null;comment:筹码"`
    BettingLimits string `gorm:"not null;comment:限红"`
}

func (SPlayer) TableName() string {
    return "or_player"
}

type SServiceNode struct {
    ID              uint                  `gorm:"primarykey;type:int unsigned;comment:ID"`
    Ip              string                `gorm:"size:50;comment:IP地址"`
    Port            uint                  `gorm:"type:int unsigned;not null;comment:端口"`
    OnlineCount     uint                  `gorm:"type:int unsigned;not null;default:0;comment:当前在线玩家数（针对连接数）"`
    Disabled        uint                  `gorm:"type:tinyint unsigned;not null;default:0;comment:是否禁用:0否,1是"`
    Note            string                `gorm:"not null;default:'';comment:备注"`
    ServiceKey      NodeServiceKey        `gorm:"size:64;not null;default:'login_ws';comment:服务标识，例如 login_ws 或 wallet_tcp 或 chat_ws等"`
    ServiceConnType ServiceConnectionType `gorm:"size:10;not null;default:'ws';comment:服务类型，ws=WebSocket，tcp=TCP"`
    BaseModel
}

type NodeServiceKey string

type ServiceConnectionType string

func (SServiceNode) TableName() string {
    return "or_service_node"
}

func GetServiceNode() (*SServiceNode, error) {
    var serviceNode SServiceNode

    result := variable.GormDbMysql.Clauses(dbresolver.Read).
        Select("ip, port").
        First(&serviceNode)

    return &serviceNode, result.Error
}

func GetTransByMid(merchatId string) (*SPlayTransaction, error) {

    var trans SPlayTransaction
    result := variable.GormDbMysql.Clauses(dbresolver.Read).Where("merchant_trans_id = ? AND wallet_field = ?", merchatId, "balance").First(&trans) // 使用 user_id

    return &trans, result.Error
}

func GetBetAccountRecords(username string, code string, startTime time.Time, endTime time.Time, offset int) ([]*SBetAccountRecord, error) {
    var bets []*SBetAccountRecord

    result := variable.GormDbMysql.Clauses(dbresolver.Read).
        Where("user_id = ? AND currency = ? AND bet_time BETWEEN ? AND ?", username, code, startTime, endTime).
        Order("id ASC").
        Limit(500).
        Offset(offset * 500).
        Find(&bets)

    return bets, result.Error
}

func GetBetAccountStats(username string, code string, startTime time.Time, endTime time.Time) (int, int, int, error) {
    var resultWin struct {
        BetAmountSum int64
        WinLostSum   int64
        BetsCount    int64
    }

    //先统计winlost > 0的有赢的
    err := variable.GormDbMysql.Clauses(dbresolver.Read).
        Model(&SBetAccountRecord{}).
        Select("COALESCE(SUM(bet_amount), 0) AS bet_amount_sum, COALESCE(SUM(win_lost), 0) AS win_lost_sum, COALESCE(SUM(bet_count), 0) AS bets_count").
        Where("user_id = ? AND currency = ? AND bet_time BETWEEN ? AND ? AND win_lost > 0", username, code, startTime, endTime).
        Scan(&resultWin).Error

    var resultLose struct {
        BetAmountSum int64
        WinLostSum   int64
        BetsCount    int64
    }

    //再统计winlost = 0输掉的局
    err = variable.GormDbMysql.Clauses(dbresolver.Read).
        Model(&SBetAccountRecord{}).
        Select("COALESCE(SUM(bet_amount), 0) AS bet_amount_sum, COALESCE(SUM(win_lost), 0) AS win_lost_sum, COALESCE(SUM(bet_count), 0) AS bets_count").
        Where("user_id = ? AND currency = ? AND bet_time BETWEEN ? AND ? AND win_lost = 0", username, code, startTime, endTime).
        Scan(&resultLose).Error

    return int(resultWin.BetAmountSum + resultLose.BetAmountSum), int(resultWin.WinLostSum - resultLose.BetAmountSum), int(resultWin.BetsCount + resultLose.BetsCount), err
}

func GetGameCatalog() ([]*SGameCategory, error) {
    var games []*SGameCategory

    result := variable.GormDbMysql.Clauses(dbresolver.Read).
        Where("is_hide", 0).
        Find(&games)

    return games, result.Error
}

func GetGameCatalogById(gameSubCode string) (*SGameCategory, error) {
    var game SGameCategory
    result := variable.GormDbMysql.Clauses(dbresolver.Read).Where("sub_code = ?", gameSubCode).First(&game)

    return &game, result.Error
}

func GetGameRules(gameTopCode string) ([]*SBetArea, error) {
    var gameRules []*SBetArea

    result := variable.GormDbMysql.Clauses(dbresolver.Read).
        Where("game_category_code = ?", gameTopCode).
        Find(&gameRules)

    return gameRules, result.Error
}

func GetGameTables(gameSubCode string) ([]*SGameTable, error) {
    var gameTables []*SGameTable

    result := variable.GormDbMysql.Clauses(dbresolver.Read).
        Where("category_sub = ?", gameSubCode).
        Find(&gameTables)

    return gameTables, result.Error
}

func GetGameTable(gameSubCode string, deskNo int32) (*SGameTable, error) {
    var gameTable *SGameTable

    result := variable.GormDbMysql.Clauses(dbresolver.Read).
        Where("category_sub = ? AND desk_no = ?", gameSubCode, uint(deskNo)).
        First(&gameTable)

    return gameTable, result.Error
}

func GetMerchantByCode(merchantCode string) (*SMerchant, error) {

    var merchant SMerchant
    result := variable.GormDbMysql.Clauses(dbresolver.Read).Where("merchant_code = ?", merchantCode).First(&merchant) // 使用 user_id

    return &merchant, result.Error
}

func GetMerchantByName(merchantName string) (*SMerchant, error) {

    var merchant SMerchant
    result := variable.GormDbMysql.Clauses(dbresolver.Read).Where("username = ?", merchantName).First(&merchant) // 使用 user_id

    return &merchant, result.Error
}

func GetOnlinePlayers(merchantName string, code string, offset int) ([]*SPlayer, error) {
    var players []*SPlayer

    query := variable.GormDbMysql.Clauses(dbresolver.Read).
        Select("api_user, username, currency, login_time, login_ip").
        Where("merchant_id = ?", merchantName).
        //Where("api_user <> username").
        Order("id ASC").
        Limit(500).
        Offset(offset * 500)

    if code != "" {
        query = query.Where("currency = ?", code)
    }

    result := query.Find(&players)
    return players, result.Error
}

func GetBetRecords(merchantCode string, username string, gameTypeId string, gameTableId int, gameRoundId string, startTime time.Time, endTime time.Time, offset int) ([]*SBetRecord, error) {
    var bets []*SBetRecord

    query := variable.GormDbMysql.Clauses(dbresolver.Read).
        Where("bet_time BETWEEN ? AND ?", startTime, endTime).
        Order("id ASC").
        Limit(500).
        Offset(offset * 500)

    if merchantCode != "" {
        query = query.Where("merchant_code = ?", merchantCode)
    }

    if username != "" {
        query = query.Where("user_id = ?", username)
    }

    if gameTypeId != "" {
        query = query.Where("type = ?", gameTypeId)
    }

    if gameTableId > 0 {
        query = query.Where("desk_no = ?", gameTableId)
    }

    if gameRoundId != "" {
        query = query.Where("round_id = ?", gameRoundId)
    }

    result := query.Find(&bets)
    return bets, result.Error
}

func CountBetRecords(merchantCode string, username string, gameTypeId string, gameTableId int, gameRoundId string, startTime time.Time, endTime time.Time) (int64, error) {
    var count int64

    query := variable.GormDbMysql.Clauses(dbresolver.Read).
        Model(&SBetRecord{}).
        Where("bet_time BETWEEN ? AND ?", startTime, endTime)

    if merchantCode != "" {
        query = query.Where("merchant_code = ?", merchantCode)
    }

    if username != "" {
        query = query.Where("user_id = ?", username)
    }

    if gameTypeId != "" {
        query = query.Where("type = ?", gameTypeId)
    }

    if gameTableId > 0 {
        query = query.Where("desk_no = ?", gameTableId)
    }

    if gameRoundId != "" {
        query = query.Where("round_id = ?", gameRoundId)
    }

    result := query.Count(&count)
    return count, result.Error
}

func GetTipRecords(merchantCode string, username string, dealername string, gameTypeId string, gameTableId int, gameRoundId string, startTime time.Time, endTime time.Time, offset int) ([]*STipRecord, error) {
    var tips []*STipRecord

    query := variable.GormDbMysql.Clauses(dbresolver.Read).
        Where("tip_time BETWEEN ? AND ?", startTime, endTime).
        Order("id ASC").
        Limit(500).
        Offset(offset * 500)

    if merchantCode != "" {
        query = query.Where("merchant_code = ?", merchantCode)
    }

    if username != "" {
        query = query.Where("user_name = ?", username)
    }

    if dealername != "" {
        query = query.Where("dealer_id = ?", dealername)
    }

    if gameTypeId != "" {
        query = query.Where("sub_type = ?", gameTypeId)
    }

    if gameTableId > 0 {
        query = query.Where("desk_no = ?", gameTableId)
    }

    if gameRoundId != "" {
        query = query.Where("round_id = ?", gameRoundId)
    }

    result := query.Find(&tips)
    return tips, result.Error
}

func CountTipRecords(merchantCode string, username string, dealername string, gameTypeId string, gameTableId int, gameRoundId string, startTime time.Time, endTime time.Time) (int64, error) {
    var count int64

    query := variable.GormDbMysql.Clauses(dbresolver.Read).
        Model(&STipRecord{}).
        Where("tip_time BETWEEN ? AND ?", startTime, endTime)

    if merchantCode != "" {
        query = query.Where("merchant_code = ?", merchantCode)
    }

    if username != "" {
        query = query.Where("user_name = ?", username)
    }

    if dealername != "" {
        query = query.Where("dealer_id = ?", dealername)
    }

    if gameTypeId != "" {
        query = query.Where("sub_type = ?", gameTypeId)
    }

    if gameTableId > 0 {
        query = query.Where("desk_no = ?", gameTableId)
    }

    if gameRoundId != "" {
        query = query.Where("round_id = ?", gameRoundId)
    }

    result := query.Count(&count)
    return count, result.Error
}

func GetSystemCurrency() ([]*SSystemCurrency, error) {
    var currs []*SSystemCurrency

    result := variable.GormDbMysql.Clauses(dbresolver.Read).
        Select("code, digits").
        Find(&currs)

    return currs, result.Error
}

func GetBetAreas() ([]*SBetArea, error) {
    var areas []*SBetArea

    result := variable.GormDbMysql.Clauses(dbresolver.Read).
        Select("area_id, area_key").
        Find(&areas)

    return areas, result.Error
}