package test

import (
	"fmt"
	"gorm.io/plugin/dbresolver"
	"testing"
	"time"
	"videogame/common/global/variable"
	models "videogame/common/model"
	"videogame/common/util/yml_config"
)

type OrPlayer struct {
	ID                  uint       `gorm:"primarykey;type:int unsigned;comment:ID"`
	MerchantKey         string     `gorm:"size:30;not null;default:'';comment:商户ID"`
	CurrencyCode        string     `gorm:"size:10;not null;default:'';comment:币种"`
	MerchantCurrencyKey string     `gorm:"size:50;not null;default:'';comment:商户币种"`
	ApiUser             string     `gorm:"size:30;not null;default:'';comment:API 用户"`
	Username            string     `gorm:"size:30;not null;default:'';comment:账号"`
	Nickname            string     `gorm:"size:50;not null;default:'';comment:昵称"`
	Level               int        `gorm:"type:tinyint unsigned;not null;default:1;comment:玩家级别"`
	LoginTime           *time.Time `gorm:"default:NULL;comment:上次登录时间"`
	LoginIp             string     `gorm:"size:50;comment:登录IP"`
	Status              uint       `gorm:"type:tinyint unsigned;not null;default:1;comment:是否冻结:0冻结,1正常"`
	BetPaused           uint       `gorm:"type:tinyint unsigned;not null;default:0;comment:是否暂停投注:0否,1是"`
	IsOnline            uint       `gorm:"type:tinyint unsigned;not null;default:0;comment:是否在线:0否,1是"`
	PlayerType          int        `gorm:"type:tinyint unsigned;not null;default:1;comment:玩家类型:1现金代理,2佣金代理"`
	Email               string     `gorm:"size:100;not null;default:'';comment:邮箱"`
	Phone               string     `gorm:"size:20;not null;default:'';comment:手机号"`
	Password            string     `gorm:"size:80;not null;default:'';comment:密码"`
	IsEditable          uint       `gorm:"type:tinyint unsigned;not null;default:1;comment:是否可编辑"`
	IsDeletable         uint       `gorm:"type:tinyint unsigned;not null;default:1;comment:是否可删除"`
	Chips               string     `gorm:"comment:筹码"`
	BettingLimits       string     `gorm:"comment:投注限额"`
	CreatedTime         *time.Time `gorm:"default:CURRENT_TIMESTAMP;comment:创建时间"`
	UpdatedTime         *time.Time `gorm:"default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;comment:更新时间"`
	MerchantId          uint       `gorm:"type:int unsigned;not null;comment:商户ID"`
	Avatar              string     `gorm:"not null;default:'';comment:头像"`
	CurrencyId          uint       `gorm:"type:int unsigned;not null;comment:币种ID"`
}

// TableName returns the table name for this struct
func (OrPlayer) TableName() string {
	return "or_player" // 这里指定你实际使用的表名
}

func TestSelectUser(t *testing.T) {
	// 2.启动针对配置文件(confgi.yml、gorm_v2.yml)变化的监听， 配置文件操作指针，初始化为全局变量
	variable.ConfigYml = yml_config.CreateYamlFactory()
	variable.ConfigYml.ConfigFileChangeListen()

	// config>gorm_v2.yml 启动文件变化监听事件
	variable.ConfigGormv2Yml = variable.ConfigYml.Clone("gorm_v2")
	variable.ConfigGormv2Yml.ConfigFileChangeListen()

	//初始化数据库mysql
	_ = models.InitMysql()
	var tmp OrPlayer
	result := variable.GormDbMysql.
		Clauses(dbresolver.Read).
		Select("login_time").
		Where("UserName = ?", "player001").
		First(&tmp)

	if result.Error != nil {
		fmt.Printf("Error:%v \n", result.Error)
		return
	}

	fmt.Println("LoginTime:", tmp.LoginTime)

}

func TestInsertUser(t *testing.T) {
	// 2.启动针对配置文件(confgi.yml、gorm_v2.yml)变化的监听， 配置文件操作指针，初始化为全局变量
	variable.ConfigYml = yml_config.CreateYamlFactory()
	variable.ConfigYml.ConfigFileChangeListen()

	// config>gorm_v2.yml 启动文件变化监听事件
	variable.ConfigGormv2Yml = variable.ConfigYml.Clone("gorm_v2")
	variable.ConfigGormv2Yml.ConfigFileChangeListen()

	//初始化数据库mysql
	_ = models.InitMysql()
	utcTime := time.Now().UTC()
	//插入2000条数据
	var players []OrPlayer
	for i := 1; i <= 2; i++ {
		username := fmt.Sprintf("root%04d", i)
		players = append(players, OrPlayer{
			MerchantKey:         "merchant001",
			CurrencyCode:        "CNY",
			MerchantCurrencyKey: "merchant001_CNY",
			ApiUser:             username,
			Username:            username,
			Nickname:            fmt.Sprintf("玩家root%04d", i),
			Level:               1,
			LoginTime:           &utcTime,
			LoginIp:             "",
			Status:              1,
			BetPaused:           0,
			IsOnline:            0,
			PlayerType:          1,
			Email:               "",
			Phone:               "",
			Password:            "$2a$04$UTM5QvfEUGbdOp.h5OdTv.SW9Y572IX8J0CXi3uqJj1ACg77az3wO", // 预设密码
			IsEditable:          1,
			IsDeletable:         1,
			Chips:               "1,2,3,4,5,9",
			BettingLimits:       "1,2,3,4,5,6,7,8,9,10",
			CreatedTime:         &utcTime,
			UpdatedTime:         &utcTime,
			MerchantId:          1,
			Avatar:              "",
			CurrencyId:          1,
		})
	}

	// 批量插入
	if err := variable.GormDbMysql.Create(&players).Error; err != nil {
		fmt.Println("插入数据失败:", err)
	} else {
		fmt.Println("成功插入 100 条数据")
	}

}

func TestDeleteUser(t *testing.T) {
	// 2.启动针对配置文件(confgi.yml、gorm_v2.yml)变化的监听， 配置文件操作指针，初始化为全局变量
	variable.ConfigYml = yml_config.CreateYamlFactory()
	variable.ConfigYml.ConfigFileChangeListen()

	// config>gorm_v2.yml 启动文件变化监听事件
	variable.ConfigGormv2Yml = variable.ConfigYml.Clone("gorm_v2")
	variable.ConfigGormv2Yml.ConfigFileChangeListen()

	//初始化数据库mysql
	_ = models.InitMysql()

	result := variable.GormDbMysql.
		Where("username LIKE ?", "%root%").
		Delete(&OrPlayer{})

	if result.Error != nil {
		fmt.Println("删除失败:", result.Error)
	} else {
		fmt.Printf("成功删除 %d 条记录\n", result.RowsAffected)
	}

}
