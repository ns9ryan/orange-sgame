package mysql_helper

import "gorm.io/gorm"

type Model interface {
	DatabaseTable() string //database + table
	Create(conn *gorm.DB) error
}

// LoadBalance 库&表的负载均衡器
type LoadBalance interface {
	Database() string
	Table() string
}
