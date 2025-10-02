package models

import (
	"time"

	"gorm.io/gorm"
)

// User represents a user in the system
type User struct {
	Base
	Username  string     `gorm:"type:varchar(50);uniqueIndex;not null" json:"username" validate:"required,min=3,max=20,alphanum"`
	Password  string     `gorm:"type:varchar(255);not null" json:"-" validate:"required,min=8"`
	IsActive  bool       `gorm:"type:boolean;default:true;not null" json:"is_active"`
	Remark    string     `gorm:"type:text" json:"remark"`
	LastLogin *time.Time `json:"last_login,omitempty"`
}

// BeforeCreate hook to set ID
func (u *User) BeforeCreate(tx *gorm.DB) error {
	// In a real application, you would generate a UUID here
	// For simplicity, we'll use a placeholder
	// You might want to use github.com/google/uuid for real UUID generation
	return nil
}
