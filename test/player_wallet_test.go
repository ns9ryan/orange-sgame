package test

import (
	"fmt"
	"testing"
	"time"
	"videogame/common/global/variable"
	models "videogame/common/model"
	"videogame/common/util/yml_config"
)

type PlayerWallet struct {
	ID                  uint      `gorm:"primarykey;type:int unsigned;comment:ID"`
	MerchantKey         string    `gorm:"size:30;not null;default:'';comment:商户ID"`
	CurrencyCode        string    `gorm:"size:10;not null;default:'';comment:币种"`
	MerchantCurrencyKey string    `gorm:"size:50;not null;default:'';comment:商户币种"`
	Username            string    `gorm:"size:30;not null;default:'';comment:账号"`
	Digits              int       `gorm:"size:10;not null;default:0;comment:数字"`
	InitialBalance      float64   `gorm:"type:decimal(10,2);not null;default:0.00;comment:初始余额"`
	Balance             float64   `gorm:"type:decimal(10,2);not null;default:0.00;comment:余额"`
	Tip                 float64   `gorm:"type:decimal(10,2);not null;default:0.00;comment:小费"`
	TotalBet            float64   `gorm:"type:decimal(10,2);not null;default:0.00;comment:总投注"`
	TotalBetEffective   float64   `gorm:"type:decimal(10,2);not null;default:0.00;comment:有效总投注"`
	TotalWater          float64   `gorm:"type:decimal(10,2);not null;default:0.00;comment:总水"`
	TotalWinLost        float64   `gorm:"type:decimal(10,2);not null;default:0.00;comment:总输赢"`
	LastBet             float64   `gorm:"type:decimal(10,2);not null;default:0.00;comment:上次投注"`
	LastBetEffective    float64   `gorm:"type:decimal(10,2);not null;default:0.00;comment:上次有效投注"`
	LastWater           float64   `gorm:"type:decimal(10,2);not null;default:0.00;comment:上次水"`
	LastWinLost         float64   `gorm:"type:decimal(10,2);not null;default:0.00;comment:上次输赢"`
	CreatedTime         time.Time `gorm:"default:CURRENT_TIMESTAMP;comment:创建时间"`
	UpdatedTime         time.Time `gorm:"default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;comment:更新时间"`
}

// TableName returns the table name for this struct
func (PlayerWallet) TableName() string {
	return "or_player_wallet" // 你可以根据实际需要修改为你的表名
}

func TestInsertWallet(t *testing.T) {
	// 2.启动针对配置文件(confgi.yml、gorm_v2.yml)变化的监听， 配置文件操作指针，初始化为全局变量
	variable.ConfigYml = yml_config.CreateYamlFactory()
	variable.ConfigYml.ConfigFileChangeListen()

	// config>gorm_v2.yml 启动文件变化监听事件
	variable.ConfigGormv2Yml = variable.ConfigYml.Clone("gorm_v2")
	variable.ConfigGormv2Yml.ConfigFileChangeListen()

	//初始化数据库mysql
	_ = models.InitMysql()

	// 插入数据
	var players []PlayerWallet
	for i := 1; i <= 2000; i++ {
		username := fmt.Sprintf("root%04d", i)
		players = append(players, PlayerWallet{
			MerchantKey:         "merchant001",
			CurrencyCode:        "CNY",
			MerchantCurrencyKey: "merchant001_CNY",
			Username:            username,
			Digits:              2,
			InitialBalance:      10000000.00,
			Balance:             10000000.00,
			Tip:                 0.00,
			TotalBet:            0.00,
			TotalBetEffective:   0.00,
			TotalWater:          0.00,
			TotalWinLost:        0.00,
			LastBet:             0.00,
			LastBetEffective:    0.00,
			LastWater:           0.00,
			LastWinLost:         0.00,
			CreatedTime:         time.Now(),
			UpdatedTime:         time.Now(),
		})
	}

	// 批量插入
	if err := variable.GormDbMysql.Create(&players).Error; err != nil {
		fmt.Println("插入数据失败:", err)
	} else {
		fmt.Println("成功插入数据")
	}

}

func TestUpdateWallet(t *testing.T) {

	//2.启动针对配置文件(confgi.yml、gorm_v2.yml)变化的监听， 配置文件操作指针，初始化为全局变量
	variable.ConfigYml = yml_config.CreateYamlFactory()
	variable.ConfigYml.ConfigFileChangeListen()

	// config>gorm_v2.yml 启动文件变化监听事件
	variable.ConfigGormv2Yml = variable.ConfigYml.Clone("gorm_v2")
	variable.ConfigGormv2Yml.ConfigFileChangeListen()

	//初始化数据库mysql
	_ = models.InitMysql()

	var wallet PlayerWallet

	result := variable.GormDbMysql.Model(&wallet).
		Where("username LIKE ?", "%root%").
		Updates(map[string]interface{}{
			"balance":             10000000,
			"total_bet":           0.00,
			"total_bet_effective": 0.00,
			"total_water":         0.00,
			"total_win_lost":      0.00,
			"last_bet":            0.00,
			"last_bet_effective":  0.00,
			"last_water":          0.00,
			"last_win_lost":       0.00,
			"updated_time":        time.Now(),
		})
	if result.Error != nil {
		fmt.Println("更新失败:", result.Error)
	} else {
		fmt.Printf("成功更新 %d 条记录\n", result.RowsAffected)
	}

}

func TestDelWallet(t *testing.T) {

	//2.启动针对配置文件(confgi.yml、gorm_v2.yml)变化的监听， 配置文件操作指针，初始化为全局变量
	variable.ConfigYml = yml_config.CreateYamlFactory()
	variable.ConfigYml.ConfigFileChangeListen()

	// config>gorm_v2.yml 启动文件变化监听事件
	variable.ConfigGormv2Yml = variable.ConfigYml.Clone("gorm_v2")
	variable.ConfigGormv2Yml.ConfigFileChangeListen()

	//初始化数据库mysql
	_ = models.InitMysql()

	var wallet PlayerWallet

	result := variable.GormDbMysql.
		Where("username LIKE ?", "%root%").
		Delete(&wallet)

	if result.Error != nil {
		fmt.Println("删除失败:", result.Error)
	} else {
		fmt.Printf("成功删除 %d 条记录\n", result.RowsAffected)
	}
}
