package models

import (
	"fmt"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
	"time"
	"videogame/common/global/variable"
)

type UserAdmin struct {
	Username string `gorm:"size:30;not null;default:'';comment:账号"`
	Nickname string `gorm:"size:50;not null;default:'';comment:昵称"`
	Password string `gorm:"not null;default:'';comment:密码"`

	Status uint `gorm:"type:tinyint unsigned;not null;default:1;comment:是否允许登录:1是,2否"`
	Level  uint `gorm:"type:int unsigned;not null;default:1;comment:等级"`

	//	Login_Time *time.Time `gorm:"type:datetime;default:NULL;comment:上次登录时间"`
	LoginIp string `gorm:"size:50;comment:登录IP"`
}

// TableName 重写 GORM 的默认表名
func (UserAdmin) TableName() string {
	return "or_control_account" // 指定自定义表名为
}

// 对象和方法的写法示例：
func (u *UserAdmin) Create() error {
	return variable.GormDbMysql.Clauses(dbresolver.Write).Create(u).Error // 使用指针类型 u
}

func (u *UserAdmin) Delete() error {
	return variable.GormDbMysql.Clauses(dbresolver.Write).Delete(u).Error // 使用指针类型 u
}

func (u *UserAdmin) Save() error {
	return variable.GormDbMysql.Clauses(dbresolver.Write).Save(u).Error
}

// 更新控端的IP
func (u *UserAdmin) UpdateIp(userName string, ip string, tm *time.Time) error {

	res := variable.GormDbMysql.Model(&UserAdmin{}).Where("Username = ?", userName).Update("Login_Ip", ip).Update("login_time", tm)

	if res.Error != nil {
		return fmt.Errorf("更新控端IP  %s状态失败：%v", userName, res.Error)
	}

	return nil
}

func (u *UserAdmin) GetUsers() ([]*UserAdmin, error) {

	var users []*UserAdmin

	if err := variable.GormDbMysql.Find(&users).Error; err != nil {
		// 错误处理
		return nil, err
	}

	return users, nil
}

// 读取指定的荷管的数据
func (u *UserAdmin) GetUserAdmin(uid string) (*UserAdmin, error) {

	var user UserAdmin

	result := variable.GormDbMysql.Clauses(dbresolver.Read).Where("Username = ?", uid).First(&user) // 使用 user_id

	return &user, result.Error
}

//// 直接提供全局公共调用的写法示例：
//func GetAllAdmins() ([]*UserAdmin, error) {
//
//	var users []*UserAdmin
//
//	if err := variable.GormDbMysql.Model(&UserAdmin{}).Find(&users).Error; err != nil {
//		// 错误处理
//		return nil, err
//	}
//
//	return users, nil
//}
//
//func GetAdminByID(id uint64) (*UserAdmin, error) {
//
//	var user UserAdmin
//
//	return &user, variable.GormDbMysql.Clauses(dbresolver.Read).First(&user, id).Error
//}

func GetAdminByUid(uid string) (*UserAdmin, error) {

	var user UserAdmin

	result := variable.GormDbMysql.Clauses(dbresolver.Read).Where("Username = ?", uid).First(&user) // 使用 user_id

	return &user, result.Error
}

// 高级：定义可复用的db scope示例：
func (u *UserAdmin) UidScope(uid string) func(db *gorm.DB) *gorm.DB {

	return func(db *gorm.DB) *gorm.DB {
		return db.Where("user_id = ?", uid)
	}

}

// 使用scope的使用案例，注意要传入被scope改写后的的conn：
func (u *UserAdmin) GetOneUserWithScope(conn *gorm.DB, fields string) ([]*UserAdmin, error) {

	var users []*UserAdmin

	if err := conn.Select(fields).Find(&users).Error; err != nil {
		// 错误处理
		return nil, err
	}
	return users, nil
}
