package mysql_helper

import "gorm.io/gorm"

func Create(conn *gorm.DB, m Model) error {
	res := conn.Table(m.DatabaseTable()).Create(m)
	if res.Error != nil {
		return res.Error
	}
	return nil
}
