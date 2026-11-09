package mysql_helper

import (
	"errors"
	"gorm.io/gorm"
)

func Count(m Model, condition *gorm.DB) (int64, error) {
	count := int64(0)
	res := condition.Table(m.DatabaseTable()).Count(&count)
	if res.Error != nil {
		return 0, res.Error
	}
	return count, nil
}

func IsExists(m Model, condition *gorm.DB) (bool, error) {
	count, err := Count(m, condition)
	if err != nil {
		return false, err
	}
	if count == 0 {
		return false, nil
	} else {
		return true, nil
	}
}

func Select(m Model, result interface{}, fields string, condition *gorm.DB) error {
	res := condition.Table(m.DatabaseTable()).Select(fields).Find(result)
	return res.Error
}

func SelectOne(m Model, result interface{}, fields interface{}, condition *gorm.DB) (bool, error) {
	res := condition.Table(m.DatabaseTable()).Select(fields).First(result)
	// 判断是否是记录未找到的错误
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return false, nil
		}
	}
	return true, nil
}

func SelectWithPage(m Model, result interface{}, fields interface{}, condition *gorm.DB, page int, pageSize int) error {
	if page >= 1 {
		page -= 1
	}
	res := condition.Table(m.DatabaseTable()).Select(fields).Limit(pageSize).Offset(page * pageSize).Find(result)
	return res.Error
}

func SelectAndCount(m Model, result interface{}, fields interface{}, condition *gorm.DB, page int, pageSize int) (int64, error) {
	count, err := Count(m, condition)
	if err != nil {
		return 0, err
	}
	if count == 0 {
		return 0, nil
	}

	// 查询的数据，分页不存在
	if count < ((int64(page) - 1) * int64(pageSize)) {
		return count, nil
	}

	// 查询数据
	err = SelectWithPage(m, result, fields, condition, int(page), int(pageSize))

	return count, err
}
