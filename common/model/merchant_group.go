package models

import (
	"gorm.io/plugin/dbresolver"
	"videogame/common/global/variable"
)

// Merchant 商户主表结构体
type MerchantGroup struct {
	ID           uint   `gorm:"primarykey;type:int unsigned;comment:ID"`
	MerchantCode string `gorm:"uniqueindex;size:10;not null;comment:商户号"`
	MerchantName string `gorm:"size:32;not null;comment:商户名称"`
	ParentId     string `gorm:"size:32;not null;default:'';comment:上级商户的username"`
	RootKey      uint   `gorm:"type:int unsigned;not null;default:0;comment:顶级Key"`
	Username     string `gorm:"uniqueindex;size:32;not null;comment:账号"`
	Nickname     string `gorm:"size:32;not null;comment:昵称"`

	CommissionRate uint `gorm:"type:int unsigned;not null;default:0;comment:成数"`
	RollingRate    uint `gorm:"type:int unsigned;not null;default:0;comment:洗码比"`

	Status    uint `gorm:"type:tinyint unsigned;not null;default:1;comment:是否允许登录:0否,1是"`
	IsActive  uint `gorm:"type:tinyint unsigned;not null;default:0;comment:是否激活(第一次登录成功后改成激活):0否,1是"`
	BetPaused uint `gorm:"type:tinyint unsigned;not null;default:0;comment:是否暂停投注:0正常,1暂停"`

	IsMain uint `gorm:"type:tinyint unsigned;not null;default:1;comment:是否主账号:0否,1是"`

	WalletType uint `gorm:"type:tinyint unsigned;not null;default:2;comment:钱包类型:1单钱包,2多钱包"`

	TemplateId string `gorm:"size:36;not null;default:'';comment:商户限红模板ID,创建会员时默认使用"`

	// api相关
	//AesKeySalt string `gorm:"not null;comment:aesKey盐"`
	//Md5Salt    string `gorm:"not null;comment:md5盐"`
	//OrToken    string `gorm:"not null;comment:商户api调用验证or_token"`
	//OrTokenKey string `gorm:"not null;comment:生成token用的商户唯一key，永久有效"`
	//WalletURL  string `gorm:"not null;default:'';comment:单钱包URL"`
	//Whitelist  string `gorm:"not null;default:'';comment:白名单IP,逗号分隔"`
}

// TableName 重写 GORM 的默认表名
func (MerchantGroup) TableName() string {
	return "or_merchant" // 指定自定义表名为
}

// 对象和方法的写法示例：
func (u *MerchantGroup) Create() error {
	return variable.GormDbMysql.Clauses(dbresolver.Write).Create(u).Error // 使用指针类型 u
}

// 读取指定的商户的数据
func (u *MerchantGroup) GetMerchantGroup(uid string) (*MerchantGroup, error) {

	var user MerchantGroup

	result := variable.GormDbMysql.Clauses(dbresolver.Read).Where("Username = ?", uid).First(&user) // 使用 user_id

	return &user, result.Error
}

// 全局的读取商户的
func GetMerchants(parentId string) ([]*MerchantGroup, error) {

	var groups []*MerchantGroup

	if err := variable.GormDbMysql.Clauses(dbresolver.Read).Where("parent_id = ?", parentId).Find(&groups).Error; err != nil {
		// 错误处理
		return nil, err
	}

	return groups, nil
}

// 读取指定的商户
func GetMerchantGroup(username string) (*MerchantGroup, error) {

	var user MerchantGroup

	result := variable.GormDbMysql.Clauses(dbresolver.Read).Where("Username = ?", username).First(&user) // 使用 user_id

	return &user, result.Error
}
