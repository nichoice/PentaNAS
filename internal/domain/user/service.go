package user

import (
	"context"
	"fmt"
	"time"
)

// Service 用户领域服务
type Service struct {
	userRepo     Repository
	roleRepo     RoleRepository
	permRepo     PermissionRepository
	userRoleRepo UserRoleRepository
}

// NewService 创建用户服务
func NewService(
	userRepo Repository,
	roleRepo RoleRepository,
	permRepo PermissionRepository,
	userRoleRepo UserRoleRepository,
) *Service {
	return &Service{
		userRepo:     userRepo,
		roleRepo:     roleRepo,
		permRepo:     permRepo,
		userRoleRepo: userRoleRepo,
	}
}

// CreateUser 创建用户
func (s *Service) CreateUser(ctx context.Context, username, email, password, fullName string) (*User, error) {
	// 检查用户名是否已存在
	existingUser, _ := s.userRepo.GetByUsername(ctx, username)
	if existingUser != nil {
		return nil, fmt.Errorf("username %s already exists", username)
	}

	// 检查邮箱是否已存在
	existingUser, _ = s.userRepo.GetByEmail(ctx, email)
	if existingUser != nil {
		return nil, fmt.Errorf("email %s already exists", email)
	}

	// 创建新用户
	user := &User{
		Username:  username,
		Email:     email,
		FullName:  fullName,
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 设置密码
	if err := user.SetPassword(password); err != nil {
		return nil, fmt.Errorf("failed to set password: %w", err)
	}

	// 保存用户
	if err := s.userRepo.Save(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to save user: %w", err)
	}

	return user, nil
}

// Authenticate 用户认证
func (s *Service) Authenticate(ctx context.Context, username, password string) (*User, error) {
	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	if !user.IsActive {
		return nil, fmt.Errorf("user account is inactive")
	}

	if !user.CheckPassword(password) {
		return nil, fmt.Errorf("invalid password")
	}

	// 更新最后登录时间
	if err := s.userRepo.UpdateLastLogin(ctx, user.ID); err != nil {
		// 记录错误但不影响登录
		// TODO: 添加日志记录
	}

	return user, nil
}

// GetUser 获取用户信息
func (s *Service) GetUser(ctx context.Context, id uint) (*User, error) {
	return s.userRepo.GetByID(ctx, id)
}

// UpdateUser 更新用户信息
func (s *Service) UpdateUser(ctx context.Context, id uint, updates map[string]interface{}) error {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// 更新字段
	if fullName, ok := updates["full_name"].(string); ok {
		user.FullName = fullName
	}
	if email, ok := updates["email"].(string); ok {
		// 检查邮箱是否已被其他用户使用
		existingUser, _ := s.userRepo.GetByEmail(ctx, email)
		if existingUser != nil && existingUser.ID != id {
			return fmt.Errorf("email %s already exists", email)
		}
		user.Email = email
	}
	if isActive, ok := updates["is_active"].(bool); ok {
		user.IsActive = isActive
	}

	user.UpdatedAt = time.Now()

	return s.userRepo.Save(ctx, user)
}

// ChangePassword 修改密码
func (s *Service) ChangePassword(ctx context.Context, id uint, oldPassword, newPassword string) error {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	if !user.CheckPassword(oldPassword) {
		return fmt.Errorf("old password is incorrect")
	}

	if err := user.SetPassword(newPassword); err != nil {
		return fmt.Errorf("failed to set new password: %w", err)
	}

	user.UpdatedAt = time.Now()
	return s.userRepo.Save(ctx, user)
}

// AssignRole 分配角色给用户
func (s *Service) AssignRole(ctx context.Context, userID, roleID uint) error {
	// 检查用户是否存在
	_, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// 检查角色是否存在
	_, err = s.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		return fmt.Errorf("role not found: %w", err)
	}

	return s.userRoleRepo.AssignRole(ctx, userID, roleID)
}

// RevokeRole 撤销用户角色
func (s *Service) RevokeRole(ctx context.Context, userID, roleID uint) error {
	return s.userRoleRepo.RevokeRole(ctx, userID, roleID)
}

// GetUserRoles 获取用户角色
func (s *Service) GetUserRoles(ctx context.Context, userID uint) ([]Role, error) {
	return s.userRoleRepo.GetUserRoles(ctx, userID)
}