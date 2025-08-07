package v1

import "pnas/internal/models"

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
	Username string             `json:"username" binding:"required,min=3,max=50" example:"testuser"`
	Password string             `json:"password" binding:"required,min=6" example:"password123"`
	UserType models.UserType    `json:"user_type" binding:"required,min=1,max=4" example:"4"`
	GroupID  uint               `json:"group_id" binding:"required" example:"1"`
}

// UpdateUserRequest 更新用户请求
type UpdateUserRequest struct {
	Username *string             `json:"username,omitempty" binding:"omitempty,min=3,max=50" example:"testuser"`
	Password *string             `json:"password,omitempty" binding:"omitempty,min=6" example:"newpassword123"`
	Status   *models.UserStatus  `json:"status,omitempty" binding:"omitempty,min=0,max=2" example:"1"`
	UserType *models.UserType    `json:"user_type,omitempty" binding:"omitempty,min=1,max=4" example:"4"`
	GroupID  *uint               `json:"group_id,omitempty" example:"1"`
}

// UserResponse 用户响应
type UserResponse struct {
	ID       uint                `json:"id" example:"1"`
	Username string              `json:"username" example:"testuser"`
	Status   models.UserStatus   `json:"status" example:"1"`
	UserType models.UserType     `json:"user_type" example:"4"`
	GroupID  uint                `json:"group_id" example:"1"`
	Group    *UserGroupResponse  `json:"group,omitempty"`
	CreatedAt string             `json:"created_at" example:"2025-08-07T15:30:00Z"`
	UpdatedAt string             `json:"updated_at" example:"2025-08-07T15:30:00Z"`
}

// UserListResponse 用户列表响应
type UserListResponse struct {
	Users []UserResponse `json:"users"`
	Total int64          `json:"total" example:"100"`
	Page  int            `json:"page" example:"1"`
	Size  int            `json:"size" example:"10"`
}

// UserGroupResponse 用户组响应
type UserGroupResponse struct {
	ID          uint   `json:"id" example:"1"`
	Name        string `json:"name" example:"系统管理员组"`
	Description string `json:"description" example:"负责系统配置、用户管理、系统维护等工作"`
	Status      int    `json:"status" example:"1"`
	CreatedAt   string `json:"created_at" example:"2025-08-07T15:30:00Z"`
	UpdatedAt   string `json:"updated_at" example:"2025-08-07T15:30:00Z"`
}

// CreateUserGroupRequest 创建用户组请求
type CreateUserGroupRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=50" example:"开发组"`
	Description string `json:"description" binding:"max=200" example:"开发人员用户组"`
}

// UpdateUserGroupRequest 更新用户组请求
type UpdateUserGroupRequest struct {
	Name        *string `json:"name,omitempty" binding:"omitempty,min=2,max=50" example:"开发组"`
	Description *string `json:"description,omitempty" binding:"omitempty,max=200" example:"开发人员用户组"`
	Status      *int    `json:"status,omitempty" binding:"omitempty,min=0,max=1" example:"1"`
}

// UserGroupListResponse 用户组列表响应
type UserGroupListResponse struct {
	Groups []UserGroupResponse `json:"groups"`
	Total  int64               `json:"total" example:"10"`
	Page   int                 `json:"page" example:"1"`
	Size   int                 `json:"size" example:"10"`
}

// UserGroupWithUsersResponse 用户组及其用户响应
type UserGroupWithUsersResponse struct {
	UserGroupResponse
	Users []UserResponse `json:"users"`
}

// CommonResponse 通用响应
type CommonResponse struct {
	Code    int         `json:"code" example:"200"`
	Message string      `json:"message" example:"操作成功"`
	Data    interface{} `json:"data,omitempty"`
}

// ErrorResponse 错误响应
type ErrorResponse struct {
	Code    int    `json:"code" example:"400"`
	Message string `json:"message" example:"请求参数错误"`
	Error   string `json:"error,omitempty" example:"validation failed"`
}