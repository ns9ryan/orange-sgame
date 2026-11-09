package models

import (
	"gorm.io/plugin/dbresolver"
	"time"
	"videogame/common/global/variable"
)

// Player 玩家表结构体
type SUserPlayer struct {
	ID uint `gorm:"primarykey;type:int unsigned;comment:ID"`

	RootKey            uint   `gorm:"type:int unsigned;not null;default:0;comment:顶级Key"`
	MerchantId         string `gorm:"size:32;not null;comment:对应商户表的username"`
	Currency           string `gorm:"type:varchar(10);not null;comment:币种"`
	MerchantCurrencyId string `gorm:"size:50;not null;comment:商户币种唯一标识"`

	ApiUser  string `gorm:"not null;comment:存储api会员原始账号"`
	Username string `gorm:"uniqueindex;size:32;not null;comment:账号"`
	Nickname string `gorm:"size:32;not null;default:'';comment:昵称"`
	ParentId string `gorm:"size:32;not null;default:'';comment:上级玩家的的username"`

	LoginTimeFirst *time.Time `gorm:"default:NULL;comment:首次次登录时间"`
	LoginTime      *time.Time `gorm:"default:NULL;comment:上次登录时间"`

	LoginIpFirst string `gorm:"size:45;comment:首次次登录IP"`

	LoginIp string `gorm:"size:45;comment:上次登录IP"`

	Status    uint `gorm:"type:tinyint unsigned;not null;default:1;comment:是否允许登录:0冻结,1正常"`
	BetPaused uint `gorm:"type:tinyint unsigned;not null;default:0;comment:是否暂停投注:0正常,1暂停"`
	MaxWin    uint `gorm:"type:int unsigned;not null;default:0;comment:最高赢额,0表示不限制"`

	IsOnline   uint `gorm:"type:tinyint unsigned;not null;default:0;comment:是否在线:0否,1是"`
	PlayerType uint `gorm:"type:tinyint unsigned;not null;default:1;comment:玩家类型:1现金代理,2佣金代理"`
	MoneyType  uint `gorm:"type:tinyint unsigned;not null;default:1;comment:玩家模式类型:1真钱,2试玩"`
	Level      uint `gorm:"type:tinyint unsigned;not null;default:1;comment:玩家级别"`

	//Email      string `gorm:"size:100;not null;default:'';comment:邮箱"`
	//Phone      string `gorm:"size:20;not null;default:'';comment:手机号"`
	Password   string `gorm:"not null;default:'';comment:密码"`
	Chips      string `gorm:"not null;comment:筹码"`
	TemplateId string `gorm:"size:36;not null;comment:商户限红模板ID"`
}

// TableName 重写 GORM 的默认表名
func (SUserPlayer) TableName() string {
	return "or_player" // 指定自定义表名为
}

func (u *SUserPlayer) Save() error {
	return variable.GormDbMysql.Clauses(dbresolver.Write).Save(u).Error
}

// 首次登陆的IP写入
func (u *SUserPlayer) WriteLoginIpFirst(userId string, ipaddr string) error {

	nowTime := time.Now()

	//res := variable.GormDbMysql.Model(&SUserPlayer{}).Where("username=?", userId).Update("login_ip_first", ipaddr).Update("login_ip", ipaddr)

	res := variable.GormDbMysql.Model(&SUserPlayer{}).Where("username=?", userId).Updates(map[string]interface{}{
		"login_ip_first":   ipaddr,
		"login_time_first": &nowTime,
		"login_ip":         ipaddr,
		"login_time":       &nowTime,
		"is_online":        1})

	if res.Error != nil {
		return res.Error
	}

	return nil
}

// 首次登陆的IP写入
func (u *SUserPlayer) WriteLoginIp(userId string, ipaddr string) error {

	nowTime := time.Now()

	res := variable.GormDbMysql.Model(&SUserPlayer{}).Where("username=?", userId).Updates(map[string]interface{}{
		"login_ip":   ipaddr,
		"login_time": &nowTime,
		"is_online":  1})

	if res.Error != nil {
		return res.Error
	}

	return nil
}

// 存入玩家在线标志
func (u *SUserPlayer) WriteOnline(userId string, online int) error {

	res := variable.GormDbMysql.Model(&SUserPlayer{}).Where("username=?", userId).Updates(map[string]interface{}{
		"is_online": online})

	if res.Error != nil {
		return res.Error
	}

	return nil

}

func GetPlayerByUid(uid string) (*SUserPlayer, error) {

	var user SUserPlayer

	result := variable.GormDbMysql.Clauses(dbresolver.Read).Where("username = ?", uid).First(&user) // 使用 user_id

	return &user, result.Error
}
