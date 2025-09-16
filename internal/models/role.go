package models

// Predefined roles
const (
	RoleSuperAdmin = "super_admin"
	RoleNormalUser = "normal_user"
	RoleAuditUser  = "audit_user"
	RoleOpsUser    = "ops_user"
)

// Role represents a role in the system
type Role struct {
	Base
	Name        string `gorm:"type:varchar(50);uniqueIndex;not null" json:"name"`
	Description string `gorm:"type:text" json:"description"`
}

// UserRole represents the relationship between users and roles
type UserRole struct {
	Base
	UserID string `gorm:"type:varchar(36);not null;index" json:"user_id"`
	RoleID string `gorm:"type:varchar(36);not null;index" json:"role_id"`
}