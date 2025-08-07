package models

import (
	"time"
	"gorm.io/gorm"
)

// BaseModel 基础模型，包含通用字段
type BaseModel struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// UserType 用户类型枚举
type UserType int

const (
	UserTypeSystem   UserType = 1 // 系统管理员
	UserTypeSecurity UserType = 2 // 安全管理员
	UserTypeAudit    UserType = 3 // 审计管理员
	UserTypeNormal   UserType = 4 // 普通用户
)

// UserStatus 用户状态枚举
type UserStatus int

const (
	UserStatusDisabled UserStatus = 0 // 禁用
	UserStatusActive   UserStatus = 1 // 正常
	UserStatusLocked   UserStatus = 2 // 锁定
)

// User 用户模型
type User struct {
	BaseModel
	Username string     `gorm:"uniqueIndex;not null;size:50" json:"username"`
	Password string     `gorm:"not null;size:255" json:"-"` // 密码不返回给前端
	Status   UserStatus `gorm:"default:1" json:"status"`    // 用户状态
	UserType UserType   `gorm:"not null;default:4" json:"user_type"`  // 用户类型，默认为普通用户
	GroupID  uint       `gorm:"not null;default:4" json:"group_id"`   // 用户组ID，默认为普通用户组
	Group    UserGroup  `gorm:"foreignKey:GroupID" json:"group,omitempty"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

// UserGroup 用户组模型
type UserGroup struct {
	BaseModel
	Name        string `gorm:"uniqueIndex;not null;size:50" json:"name"`
	Description string `gorm:"size:200" json:"description"`
	Status      int    `gorm:"default:1" json:"status"` // 1:正常 0:禁用
	Users       []User `gorm:"foreignKey:GroupID" json:"users,omitempty"`
}

// TableName 指定表名
func (UserGroup) TableName() string {
	return "user_groups"
}

// HealthCheck 健康检查记录模型
type HealthCheck struct {
	BaseModel
	ClientIP  string `gorm:"size:45" json:"client_ip"`
	UserAgent string `gorm:"size:500" json:"user_agent"`
	Status    string `gorm:"size:20;default:'success'" json:"status"`
}

// TableName 指定表名
func (HealthCheck) TableName() string {
	return "health_checks"
}