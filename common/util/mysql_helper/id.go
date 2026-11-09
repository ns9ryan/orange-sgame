package mysql_helper

import (
	"errors"
	"fmt"
	"gorm.io/gorm"
)

func SelectById(conn *gorm.DB, m Model, fields interface{}, id interface{}) (bool, error) {
	if err := VerifyId(id); err != nil {
		return false, err
	}

	res := conn.Table(m.DatabaseTable()).Select(fields).Where("id=?", id).First(m)
	// 判断是否是记录未找到的错误
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return false, nil
		}
	}
	return true, nil
}

func SelectInId(conn *gorm.DB, m Model, fields interface{}, idSlice interface{}, result interface{}) error {
	if err := VerifyIdSlice(idSlice); err != nil {
		return err
	}
	res := conn.Table(m.DatabaseTable()).Select(fields).Where("id in (?)", idSlice).Find(result)
	if res.Error != nil {
		return res.Error
	}
	return nil
}

func DeleteById(conn *gorm.DB, m Model, id interface{}) (bool, error) {
	if err := VerifyId(id); err != nil {
		return false, err
	}
	res := conn.Table(m.DatabaseTable()).Where("id=?", id).Delete(nil)
	if res.Error != nil {
		return false, res.Error
	}
	if res.RowsAffected == 0 {
		return false, nil
	}
	if res.RowsAffected == 1 {
		return true, nil
	}

	return false, fmt.Errorf("删除数据数量大于1")
}

func DeleteInId(conn *gorm.DB, m Model, idSlice interface{}) (int64, error) {
	if err := VerifyIdSlice(idSlice); err != nil {
		return 0, err
	}
	res := conn.Table(m.DatabaseTable()).Where("id in (?)", idSlice).Delete(nil)
	if res.Error != nil {
		return 0, res.Error
	}
	return res.RowsAffected, nil
}

func IsExistsById(conn *gorm.DB, m Model, id interface{}) (bool, error) {
	if err := VerifyId(id); err != nil {
		return false, err
	}
	count := int64(0)
	res := conn.Table(m.DatabaseTable()).Where("id=?", id).Count(&count)
	if res.Error != nil {
		return false, res.Error
	}
	if count == 0 {
		return false, nil
	}
	if count == 1 {
		return true, nil
	}
	return false, fmt.Errorf("数据数量不是唯一")
}

func IsExistsInId(conn *gorm.DB, m Model, id interface{}) (bool, error) {
	switch id.(type) {
	case []uint64:
		for _, row := range id.([]uint64) {
			exists, err := IsExistsById(conn, m, row)
			if err != nil {
				return false, err
			}
			if exists {
				return true, nil
			}
		}
		break
	case []int64:
		for _, row := range id.([]int64) {
			exists, err := IsExistsById(conn, m, row)
			if err != nil {
				return false, err
			}
			if exists {
				return true, nil
			}
		}
		break
	case []uint32:
		for _, row := range id.([]uint32) {
			exists, err := IsExistsById(conn, m, row)
			if err != nil {
				return false, err
			}
			if exists {
				return true, nil
			}
		}
		break
	case []int32:
		for _, row := range id.([]int32) {
			exists, err := IsExistsById(conn, m, row)
			if err != nil {
				return false, err
			}
			if exists {
				return true, nil
			}
		}
		break
	default:
		return false, fmt.Errorf("不支持的数据类型 id=%v", id)
	}

	return false, nil
}

func UpdateById(conn *gorm.DB, m Model, id interface{}) (bool, error) {
	if err := VerifyId(id); err != nil {
		return false, err
	}
	res := conn.Table(m.DatabaseTable()).Where("id=?", id).Updates(m)
	if res.Error != nil {
		return false, res.Error
	}
	if res.RowsAffected == 0 {
		return false, nil
	}
	if res.RowsAffected == 1 {
		return true, nil
	}

	return true, fmt.Errorf("更新的数据数量大于1")
}
