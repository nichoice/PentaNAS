package user

import "context"

// Repository 用户仓储接口
type Repository interface {
	GetByID(ctx context.Context, id uint) (*User, error)
	GetByUsername(ctx context.Context, username string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetAll(ctx context.Context, limit, offset int) ([]User, error)
	Save(ctx context.Context, user *User) error
	Delete(ctx context.Context, id uint) error
	UpdateLastLogin(ctx context.Context, id uint) error
}

// RoleRepository 角色仓储接口
type RoleRepository interface {
	GetByID(ctx context.Context, id uint) (*Role, error)
	GetByName(ctx context.Context, name string) (*Role, error)
	GetAll(ctx context.Context) ([]Role, error)
	Save(ctx context.Context, role *Role) error
	Delete(ctx context.Context, id uint) error
}

// PermissionRepository 权限仓储接口
type PermissionRepository interface {
	GetByID(ctx context.Context, id uint) (*Permission, error)
	GetByName(ctx context.Context, name string) (*Permission, error)
	GetAll(ctx context.Context) ([]Permission, error)
	Save(ctx context.Context, permission *Permission) error
	Delete(ctx context.Context, id uint) error
}

// UserRoleRepository 用户角色关联仓储接口
type UserRoleRepository interface {
	AssignRole(ctx context.Context, userID, roleID uint) error
	RevokeRole(ctx context.Context, userID, roleID uint) error
	GetUserRoles(ctx context.Context, userID uint) ([]Role, error)
	GetRoleUsers(ctx context.Context, roleID uint) ([]User, error)
}