package models

import (
	"time"

	"gorm.io/gorm"
)

// Base model with common fields
type Base struct {
	ID        string          `gorm:"type:varchar(36);primaryKey" json:"id"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
	DeletedAt *gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}
