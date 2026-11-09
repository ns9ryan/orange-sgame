package models

import (
	"fmt"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
	"time"
	"videogame/common/global/variable"
)

type UserDealer struct {
	//Username   string     `gorm:"size:30;not null;default:'';comment:账号"`
	Employee_number string     `gorm:"size:20;not null;default:'';comment:工号"`
	Nickname        string     `gorm:"size:50;not null;default:'';comment:昵称"`
	Balance_amount  int        `gorm:"type:int unsigned;not null;default:0;comment:荷管打赏金额"`
	Status          uint       `gorm:"type:tinyint unsigned;not null;default:1;comment:是否允许登录:1是,2否"`
	LoginTime       *time.Time `gorm:"default:NULL;comment:上次登录时间"`
	LoginIp         string     `gorm:"size:50;comment:登录IP"`
	Password        string     `gorm:"not null;default:'';comment:密码"`
	Avatar          string     `gorm:"not null;default:'';comment:头像"`

	Level int `gorm:"type:tinyint unsigned;not null;default:1;comment:荷管级别"`
}

// TableName 重写 GORM 的默认表名
func (UserDealer) TableName() string {
	return "or_dealer" // 指定自定义表名为
}

// 对象和方法的写法示例：
func (u *UserDealer) Create() error {
	return variable.GormDbMysql.Clauses(dbresolver.Write).Create(u).Error // 使用指针类型 u
}

func (u *UserDealer) Delete() error {
	return variable.GormDbMysql.Clauses(dbresolver.Write).Delete(u).Error // 使用指针类型 u
}

func (u *UserDealer) Save() error {
	return variable.GormDbMysql.Clauses(dbresolver.Write).Save(u).Error
}

// 更新控端的IP
func (u *UserDealer) UpdateIp(userName string, ip string) error {

	res := variable.GormDbMysql.Model(&UserDealer{}).Where("Username = ?", userName).Update("LoginIp", ip)

	if res.Error != nil {
		return fmt.Errorf("更新控端IP  %s状态失败：%v", userName, res.Error)
	}

	return nil
}

func (u *UserDealer) GetUsers() ([]*UserDealer, error) {

	var users []*UserDealer

	if err := variable.GormDbMysql.Find(&users).Error; err != nil {
		// 错误处理
		return nil, err
	}

	return users, nil
}

// 读取指定的荷管的数据
func (u *UserDealer) GetUserDealer(uid string) (*UserDealer, error) {

	var user UserDealer

	result := variable.GormDbMysql.Clauses(dbresolver.Read).Where("Username = ?", uid).First(&user) // 使用 user_id

	return &user, result.Error
}

// 直接提供全局公共调用的写法示例：
func GetAllUsers() ([]*UserDealer, error) {

	var users []*UserDealer

	if err := variable.GormDbMysql.Find(&users).Error; err != nil {
		// 错误处理
		return nil, err
	}

	return users, nil
}

func GetUserByID(id uint64) (*UserDealer, error) {

	var user UserDealer

	return &user, variable.GormDbMysql.Clauses(dbresolver.Read).First(&user, id).Error
}

func GetDealerByUid(uid string) (*UserDealer, error) {

	var user UserDealer

	result := variable.GormDbMysql.Clauses(dbresolver.Read).Where("Employee_number = ?", uid).First(&user) // 使用 user_id

	return &user, result.Error
}

// CreateUserWithTransaction 在函数内部管理事务的创建、提交与回滚
func CreateUserWithTransaction(user *UserDealer) error {
	// 1. 开始事务
	tx := variable.GormDbMysql.Clauses(dbresolver.Write).Session(&gorm.Session{}).Begin()

	// 2. 在函数结束时确保事务正确提交或回滚
	defer func() {
		if r := recover(); r != nil {
			// 如果发生panic，进行回滚
			tx.Rollback()
			// 可以选择处理panic，比如记录日志或返回错误
		} else if tx.Error == nil {
			// 如果没有错误，提交事务
			tx.Commit()
		} else {
			// 如果有错误，回滚事务
			tx.Rollback()
		}
	}()

	// 3. 插入用户信息
	if err := tx.Create(user).Error; err != nil {
		tx.Error = fmt.Errorf("插入用户失败：%v", err)
		return tx.Error
	}

	// 4. 模拟业务操作（更新用户邮箱等）
	//user.Email = "updated_" + user.Email // 更新邮箱为新的值
	//if err := tx.Model(&UserDealer{}).Where("id = ?", user.Id).Update("email", user.Email).Error; err != nil {
	//	tx.Error = fmt.Errorf("更新用户邮箱失败：%v", err)
	//	return tx.Error
	//}

	// 如果没有错误，defer 会自动提交事务
	return nil
}

// UpdateUserWithTransaction 接受 tx 作为参数并在事务内执行更新操作
func UpdateUserWithTransaction(tx *gorm.DB, userId string, newEmail string, newPhone string) error {

	// 1. 查询用户
	var user UserDealer

	if err := tx.Where("user_id = ?", userId).First(&user).Error; err != nil {
		tx.Rollback() // 如果查询失败，回滚事务
		return fmt.Errorf("查询用户失败：%v", err)
	}

	// 2. 更新邮箱和手机号
	//user.Email = newEmail
	//user.Phone = newPhone
	//if err := tx.Save(&user).Error; err != nil {
	//	tx.Rollback() // 更新失败，回滚事务
	//	return fmt.Errorf("更新用户信息失败：%v", err)
	//}

	// 3. 提交事务
	return tx.Commit().Error // 提交事务
}

// 高级：定义可复用的db scope示例：
func (u *UserDealer) UidScope(uid string) func(db *gorm.DB) *gorm.DB {

	return func(db *gorm.DB) *gorm.DB {
		return db.Where("user_id = ?", uid)
	}

}

// 使用scope的使用案例，注意要传入被scope改写后的的conn：
func (u *UserDealer) GetOneUserWithScope(conn *gorm.DB, fields string) ([]*UserDealer, error) {

	var users []*UserDealer

	if err := conn.Select(fields).Find(&users).Error; err != nil {
		// 错误处理
		return nil, err
	}
	return users, nil
}
