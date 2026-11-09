package models

import (
	"time"
)

// BaseModel 公共model
type BaseModel struct {
	CreatedTime time.Time `gorm:"autoCreateTime;comment:创建时间"`
	UpdatedTime time.Time `gorm:"autoUpdateTime;comment:更新时间"`
}
