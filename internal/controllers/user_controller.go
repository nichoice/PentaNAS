package controllers

import (
	"net/http"
	"strconv"
	"time"

	"pnas/api/v1"
	"pnas/internal/models"
	"pnas/internal/repositories"
	"pnas/internal/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type UserController struct {
	logger        *zap.Logger
	userRepo      repositories.UserRepository
	userGroupRepo repositories.UserGroupRepository
}

func NewUserController(logger *zap.Logger, userRepo repositories.UserRepository, userGroupRepo repositories.UserGroupRepository) *UserController {
	return &UserController{
		logger:        logger,
		userRepo:      userRepo,
		userGroupRepo: userGroupRepo,
	}
}

// CreateUser 创建用户
// @Summary 创建用户
// @Description 创建新用户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param user body v1.CreateUserRequest true "用户信息"
// @Success 201 {object} response.BaseResponse{data=v1.UserResponse}
// @Failure 400 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/users [post]
func (c *UserController) CreateUser(ctx *gin.Context) {
	var req v1.CreateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.logger.Error("创建用户参数绑定失败", zap.Error(err))
		response.ValidationError(ctx, err)
		return
	}

	// 检查用户名是否已存在
	if existingUser, _ := c.userRepo.GetByUsername(req.Username); existingUser != nil {
		c.logger.Warn("用户名已存在", zap.String("username", req.Username))
		response.BadRequest(ctx, "user.username.exists", nil)
		return
	}

	// 检查用户组是否存在
	if _, err := c.userGroupRepo.GetByID(req.GroupID); err != nil {
		c.logger.Error("用户组不存在", zap.Uint("group_id", req.GroupID), zap.Error(err))
		response.BadRequest(ctx, "user_group.not_found", err)
		return
	}

	// 创建用户
	user := &models.User{
		Username: req.Username,
		Password: req.Password, // 实际应用中需要加密
		Status:   models.UserStatusActive,
		UserType: req.UserType,
		GroupID:  req.GroupID,
	}

	if err := c.userRepo.Create(user); err != nil {
		c.logger.Error("创建用户失败", zap.Error(err))
		response.InternalServerError(ctx, "user.create.failed", err)
		return
	}

	// 获取完整的用户信息（包含用户组）
	createdUser, err := c.userRepo.GetByID(user.ID)
	if err != nil {
		c.logger.Error("获取创建的用户信息失败", zap.Error(err))
		response.InternalServerError(ctx, "server.database_error", err)
		return
	}

	responseData := c.convertToUserResponse(createdUser)
	c.logger.Info("用户创建成功", zap.String("username", user.Username), zap.Uint("user_id", user.ID))

	response.SuccessWithMessage(ctx, "user.create.success", responseData)
}

// GetUser 获取用户详情
// @Summary 获取用户详情
// @Description 根据用户ID获取用户详细信息
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param id path int true "用户ID"
// @Success 200 {object} v1.CommonResponse{data=v1.UserResponse}
// @Failure 404 {object} v1.ErrorResponse
// @Failure 500 {object} v1.ErrorResponse
// @Router /api/v1/users/{id} [get]
func (c *UserController) GetUser(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, v1.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "无效的用户ID",
		})
		return
	}

	user, err := c.userRepo.GetByID(uint(id))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			ctx.JSON(http.StatusNotFound, v1.ErrorResponse{
				Code:    http.StatusNotFound,
				Message: "用户不存在",
			})
		} else {
			c.logger.Error("获取用户失败", zap.Error(err))
			ctx.JSON(http.StatusInternalServerError, v1.ErrorResponse{
				Code:    http.StatusInternalServerError,
				Message: "获取用户失败",
			})
		}
		return
	}

	response := c.convertToUserResponse(user)
	ctx.JSON(http.StatusOK, v1.CommonResponse{
		Code:    http.StatusOK,
		Message: "获取成功",
		Data:    response,
	})
}

// UpdateUser 更新用户
// @Summary 更新用户
// @Description 更新用户信息
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param id path int true "用户ID"
// @Param user body v1.UpdateUserRequest true "更新的用户信息"
// @Success 200 {object} v1.CommonResponse{data=v1.UserResponse}
// @Failure 400 {object} v1.ErrorResponse
// @Failure 404 {object} v1.ErrorResponse
// @Failure 500 {object} v1.ErrorResponse
// @Router /api/v1/users/{id} [put]
func (c *UserController) UpdateUser(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, v1.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "无效的用户ID",
		})
		return
	}

	var req v1.UpdateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.logger.Error("更新用户参数绑定失败", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, v1.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "请求参数错误",
			Error:   err.Error(),
		})
		return
	}

	// 获取现有用户
	user, err := c.userRepo.GetByID(uint(id))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			ctx.JSON(http.StatusNotFound, v1.ErrorResponse{
				Code:    http.StatusNotFound,
				Message: "用户不存在",
			})
		} else {
			c.logger.Error("获取用户失败", zap.Error(err))
			ctx.JSON(http.StatusInternalServerError, v1.ErrorResponse{
				Code:    http.StatusInternalServerError,
				Message: "获取用户失败",
			})
		}
		return
	}

	// 更新字段
	if req.Username != nil {
		// 检查用户名是否已被其他用户使用
		if existingUser, _ := c.userRepo.GetByUsername(*req.Username); existingUser != nil && existingUser.ID != user.ID {
			ctx.JSON(http.StatusBadRequest, v1.ErrorResponse{
				Code:    http.StatusBadRequest,
				Message: "用户名已存在",
			})
			return
		}
		user.Username = *req.Username
	}
	if req.Password != nil {
		user.Password = *req.Password // 实际应用中需要加密
	}
	if req.Status != nil {
		user.Status = *req.Status
	}
	if req.UserType != nil {
		user.UserType = *req.UserType
	}
	if req.GroupID != nil {
		// 检查用户组是否存在
		if _, err := c.userGroupRepo.GetByID(*req.GroupID); err != nil {
			ctx.JSON(http.StatusBadRequest, v1.ErrorResponse{
				Code:    http.StatusBadRequest,
				Message: "指定的用户组不存在",
			})
			return
		}
		user.GroupID = *req.GroupID
	}

	if err := c.userRepo.Update(user); err != nil {
		c.logger.Error("更新用户失败", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, v1.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "更新用户失败",
			Error:   err.Error(),
		})
		return
	}

	// 重新获取更新后的用户信息
	updatedUser, err := c.userRepo.GetByID(user.ID)
	if err != nil {
		c.logger.Error("获取更新后的用户信息失败", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, v1.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "获取用户信息失败",
		})
		return
	}

	response := c.convertToUserResponse(updatedUser)
	c.logger.Info("用户更新成功", zap.Uint("user_id", user.ID))

	ctx.JSON(http.StatusOK, v1.CommonResponse{
		Code:    http.StatusOK,
		Message: "用户更新成功",
		Data:    response,
	})
}

// DeleteUser 删除用户
// @Summary 删除用户
// @Description 软删除用户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param id path int true "用户ID"
// @Success 200 {object} v1.CommonResponse
// @Failure 404 {object} v1.ErrorResponse
// @Failure 500 {object} v1.ErrorResponse
// @Router /api/v1/users/{id} [delete]
func (c *UserController) DeleteUser(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, v1.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "无效的用户ID",
		})
		return
	}

	// 检查用户是否存在
	if _, err := c.userRepo.GetByID(uint(id)); err != nil {
		if err == gorm.ErrRecordNotFound {
			ctx.JSON(http.StatusNotFound, v1.ErrorResponse{
				Code:    http.StatusNotFound,
				Message: "用户不存在",
			})
		} else {
			c.logger.Error("获取用户失败", zap.Error(err))
			ctx.JSON(http.StatusInternalServerError, v1.ErrorResponse{
				Code:    http.StatusInternalServerError,
				Message: "获取用户失败",
			})
		}
		return
	}

	if err := c.userRepo.Delete(uint(id)); err != nil {
		c.logger.Error("删除用户失败", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, v1.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "删除用户失败",
			Error:   err.Error(),
		})
		return
	}

	c.logger.Info("用户删除成功", zap.Uint("user_id", uint(id)))
	ctx.JSON(http.StatusOK, v1.CommonResponse{
		Code:    http.StatusOK,
		Message: "用户删除成功",
	})
}

// ListUsers 获取用户列表
// @Summary 获取用户列表
// @Description 分页获取用户列表
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param size query int false "每页数量" default(10)
// @Param user_type query int false "用户类型筛选"
// @Param group_id query int false "用户组ID筛选"
// @Success 200 {object} v1.CommonResponse{data=v1.UserListResponse}
// @Failure 500 {object} v1.ErrorResponse
// @Router /api/v1/users [get]
func (c *UserController) ListUsers(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(ctx.DefaultQuery("size", "10"))
	userTypeStr := ctx.Query("user_type")
	groupIDStr := ctx.Query("group_id")

	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 10
	}

	offset := (page - 1) * size
	var users []*models.User
	var total int64
	var err error

	// 根据筛选条件获取用户列表
	if userTypeStr != "" {
		userType, err := strconv.Atoi(userTypeStr)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, v1.ErrorResponse{
				Code:    http.StatusBadRequest,
				Message: "无效的用户类型",
			})
			return
		}
		users, total, err = c.userRepo.GetByUserType(models.UserType(userType), offset, size)
	} else if groupIDStr != "" {
		groupID, err := strconv.ParseUint(groupIDStr, 10, 32)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, v1.ErrorResponse{
				Code:    http.StatusBadRequest,
				Message: "无效的用户组ID",
			})
			return
		}
		users, total, err = c.userRepo.GetByGroupID(uint(groupID), offset, size)
	} else {
		users, total, err = c.userRepo.List(offset, size)
	}

	if err != nil {
		c.logger.Error("获取用户列表失败", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, v1.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "获取用户列表失败",
			Error:   err.Error(),
		})
		return
	}

	// 转换为响应格式
	userResponses := make([]v1.UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = c.convertToUserResponse(user)
	}

	response := v1.UserListResponse{
		Users: userResponses,
		Total: total,
		Page:  page,
		Size:  size,
	}

	ctx.JSON(http.StatusOK, v1.CommonResponse{
		Code:    http.StatusOK,
		Message: "获取成功",
		Data:    response,
	})
}

// convertToUserResponse 转换用户模型为响应格式
func (c *UserController) convertToUserResponse(user *models.User) v1.UserResponse {
	response := v1.UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Status:    user.Status,
		UserType:  user.UserType,
		GroupID:   user.GroupID,
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
		UpdatedAt: user.UpdatedAt.Format(time.RFC3339),
	}

	// 如果包含用户组信息
	if user.Group.ID != 0 {
		response.Group = &v1.UserGroupResponse{
			ID:          user.Group.ID,
			Name:        user.Group.Name,
			Description: user.Group.Description,
			Status:      user.Group.Status,
			CreatedAt:   user.Group.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   user.Group.UpdatedAt.Format(time.RFC3339),
		}
	}

	return response
}