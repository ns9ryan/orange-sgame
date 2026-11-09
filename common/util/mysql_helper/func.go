package mysql_helper

import (
	"fmt"
	"gorm.io/gorm"
)

func VerifyId(id interface{}) error {
	switch id.(type) {
	case uint64:
	case uint:
	case uint32:
	case uint16:
	case uint8:
	case int64:
	case int:
	case int32:
	case int16:
	case int8:
	default:
		return fmt.Errorf("不支持的数量ID类型 id=%v", id)
	}

	return nil
}

func VerifyIdSlice(idSlice interface{}) error {
	switch idSlice.(type) {
	case []uint64:
	case []uint:
	case []uint32:
	case []uint16:
	case []uint8:
	case []int64:
	case []int:
	case []int32:
	case []int16:
	case []int8:
	default:
		return fmt.Errorf("不支持的数量ID类型 idSlice=%v", idSlice)
	}

	return nil
}

func VerifyString(str interface{}) error {
	switch str.(type) {
	case string:
	default:
		return fmt.Errorf("不支持的String类型 str=%v", str)
	}

	return nil
}

func IsDatabaseExists(conn *gorm.DB, databaseName string) (bool, error) {
	count := 0
	res := conn.Raw("select count(1) from information_schema.SCHEMATA where SCHEMA_NAME=?", databaseName)
	if res.Error != nil {
		return false, res.Error
	}
	rows, err := res.Rows()
	if err != nil {
		return false, err
	}
	for rows.Next() {
		if err := rows.Scan(&count); err != nil {
			return false, err
		}
	}

	if count <= 0 {
		return false, nil
	} else {
		return true, nil
	}
}

// GetDatabaseName 获取当前数据库名称
func GetDatabaseName(conn *gorm.DB) string {
	var dbName string
	_ = conn.Raw("SELECT DATABASE()").Scan(&dbName).Error
	return dbName
}
