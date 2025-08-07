package v1

import "pnas/internal/models"

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required" example:"sysadmin"`
	Password string `json:"password" binding:"required" example:"password123"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Status  string    `json:"status" example:"success"`
	Message string    `json:"message" example:"登录成功"`
	Data    LoginData `json:"data"`
}

// LoginData 登录数据
type LoginData struct {
	Token    string           `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	User     UserInfo         `json:"user"`
	ExpiresAt string          `json:"expires_at" example:"2025-08-08T23:59:59Z"`
}

// UserInfo 用户信息
type UserInfo struct {
	ID       uint                `json:"id" example:"1"`
	Username string              `json:"username" example:"sysadmin"`
	UserType models.UserType     `json:"user_type" example:"1"`
	Status   models.UserStatus   `json:"status" example:"1"`
	Group    UserGroupInfo       `json:"group"`
}

// UserGroupInfo 用户组信息
type UserGroupInfo struct {
	ID          uint   `json:"id" example:"1"`
	Name        string `json:"name" example:"系统管理员组"`
	Description string `json:"description" example:"负责系统配置、用户管理、系统维护"`
}

// RefreshTokenRequest 刷新Token请求
type RefreshTokenRequest struct {
	Token string `json:"token" binding:"required" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

// RefreshTokenResponse 刷新Token响应
type RefreshTokenResponse struct {
	Status  string           `json:"status" example:"success"`
	Message string           `json:"message" example:"Token刷新成功"`
	Data    RefreshTokenData `json:"data"`
}

// RefreshTokenData 刷新Token数据
type RefreshTokenData struct {
	Token     string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	ExpiresAt string `json:"expires_at" example:"2025-08-08T23:59:59Z"`
}

// LogoutResponse 登出响应
type LogoutResponse struct {
	Status  string `json:"status" example:"success"`
	Message string `json:"message" example:"登出成功"`
}