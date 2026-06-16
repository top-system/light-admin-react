package database

import (
	"gorm.io/gorm"
)

// Model base model
type Model struct {
	ID        string         `gorm:"column:id;primaryKey;type:char(32);" json:"id"`
	CreatedAt Datetime       `gorm:"column:created_at;autoCreateTime;" json:"created_at"`
	UpdatedAt Datetime       `gorm:"column:updated_at;autoUpdateTime;" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index;" json:"-"`
}
