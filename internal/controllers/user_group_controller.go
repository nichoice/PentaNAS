package controllers

import (
	"net/http"
	"strconv"
	"time"

	"pnas/api/v1"
	"pnas/internal/models"
	"pnas/internal/repositories"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type UserGroupController struct {
	logger        *zap.Logger
	userGroupRepo repositories.UserGroupRepository
}

func NewUserGroupController(logger *zap.Logger, userGroupRepo repositories.UserGroupRepository) *UserGroupController {
	return &UserGroupController{
		logger:        logger,
		userGroupRepo: userGroupRepo,
	}
}

// CreateUserGroup 创建用户组
// @Summary 创建用户组
// @Description 创建新用户组
// @Tags 用户组管理
// @Accept json
// @Produce json
// @Param group body v1.CreateUserGroupRequest true "用户组信息"
// @Success 201 {object} v1.CommonResponse{data=v1.UserGroupResponse}
// @Failure 400 {object} v1.ErrorResponse
// @Failure 500 {object} v1.ErrorResponse
// @Router /api/v1/user-groups [post]
func (c *UserGroupController) CreateUserGroup(ctx *gin.Context) {
	var req v1.CreateUserGroupRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.logger.Error("创建用户组参数绑定失败", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, v1.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "请求参数错误",
			Error:   err.Error(),
		})
		return
	}

	// 检查用户组名是否已存在
	if existingGroup, _ := c.userGroupRepo.GetByName(req.Name); existingGroup != nil {
		c.logger.Warn("用户组名已存在", zap.String("group_name", req.Name))
		ctx.JSON(http.StatusBadRequest, v1.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "用户组名已存在",
		})
		return
	}

	// 创建用户组
	group := &models.UserGroup{
		Name:        req.Name,
		Description: req.Description,
		Status:      1, // 默认启用
	}

	if err := c.userGroupRepo.Create(group); err != nil {
		c.logger.Error("创建用户组失败", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, v1.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "创建用户组失败",
			Error:   err.Error(),
		})
		return
	}

	response := c.convertToUserGroupResponse(group)
	c.logger.Info("用户组创建成功", zap.String("group_name", group.Name), zap.Uint("group_id", group.ID))

	ctx.JSON(http.StatusCreated, v1.CommonResponse{
		Code:    http.StatusCreated,
		Message: "用户组创建成功",
		Data:    response,
	})
}

// GetUserGroup 获取用户组详情
// @Summary 获取用户组详情
// @Description 根据用户组ID获取用户组详细信息
// @Tags 用户组管理
// @Accept json
// @Produce json
// @Param id path int true "用户组ID"
// @Success 200 {object} v1.CommonResponse{data=v1.UserGroupResponse}
// @Failure 404 {object} v1.ErrorResponse
// @Failure 500 {object} v1.ErrorResponse
// @Router /api/v1/user-groups/{id} [get]
func (c *UserGroupController) GetUserGroup(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, v1.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "无效的用户组ID",
		})
		return
	}

	group, err := c.userGroupRepo.GetByID(uint(id))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			ctx.JSON(http.StatusNotFound, v1.ErrorResponse{
				Code:    http.StatusNotFound,
				Message: "用户组不存在",
			})
		} else {
			c.logger.Error("获取用户组失败", zap.Error(err))
			ctx.JSON(http.StatusInternalServerError, v1.ErrorResponse{
				Code:    http.StatusInternalServerError,
				Message: "获取用户组失败",
			})
		}
		return
	}

	response := c.convertToUserGroupResponse(group)
	ctx.JSON(http.StatusOK, v1.CommonResponse{
		Code:    http.StatusOK,
		Message: "获取成功",
		Data:    response,
	})
}

// GetUserGroupWithUsers 获取用户组及其用户列表
// @Summary 获取用户组及其用户列表
// @Description 根据用户组ID获取用户组详细信息及其包含的用户列表
// @Tags 用户组管理
// @Accept json
// @Produce json
// @Param id path int true "用户组ID"
// @Success 200 {object} v1.CommonResponse{data=v1.UserGroupWithUsersResponse}
// @Failure 404 {object} v1.ErrorResponse
// @Failure 500 {object} v1.ErrorResponse
// @Router /api/v1/user-groups/{id}/users [get]
func (c *UserGroupController) GetUserGroupWithUsers(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, v1.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "无效的用户组ID",
		})
		return
	}

	group, err := c.userGroupRepo.GetWithUsers(uint(id))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			ctx.JSON(http.StatusNotFound, v1.ErrorResponse{
				Code:    http.StatusNotFound,
				Message: "用户组不存在",
			})
		} else {
			c.logger.Error("获取用户组及用户失败", zap.Error(err))
			ctx.JSON(http.StatusInternalServerError, v1.ErrorResponse{
				Code:    http.StatusInternalServerError,
				Message: "获取用户组及用户失败",
			})
		}
		return
	}

	// 转换用户列表
	users := make([]v1.UserResponse, len(group.Users))
	for i, user := range group.Users {
		users[i] = v1.UserResponse{
			ID:        user.ID,
			Username:  user.Username,
			Status:    user.Status,
			UserType:  user.UserType,
			GroupID:   user.GroupID,
			CreatedAt: user.CreatedAt.Format(time.RFC3339),
			UpdatedAt: user.UpdatedAt.Format(time.RFC3339),
		}
	}

	response := v1.UserGroupWithUsersResponse{
		UserGroupResponse: c.convertToUserGroupResponse(group),
		Users:             users,
	}

	ctx.JSON(http.StatusOK, v1.CommonResponse{
		Code:    http.StatusOK,
		Message: "获取成功",
		Data:    response,
	})
}

// UpdateUserGroup 更新用户组
// @Summary 更新用户组
// @Description 更新用户组信息
// @Tags 用户组管理
// @Accept json
// @Produce json
// @Param id path int true "用户组ID"
// @Param group body v1.UpdateUserGroupRequest true "更新的用户组信息"
// @Success 200 {object} v1.CommonResponse{data=v1.UserGroupResponse}
// @Failure 400 {object} v1.ErrorResponse
// @Failure 404 {object} v1.ErrorResponse
// @Failure 500 {object} v1.ErrorResponse
// @Router /api/v1/user-groups/{id} [put]
func (c *UserGroupController) UpdateUserGroup(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, v1.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "无效的用户组ID",
		})
		return
	}

	var req v1.UpdateUserGroupRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.logger.Error("更新用户组参数绑定失败", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, v1.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "请求参数错误",
			Error:   err.Error(),
		})
		return
	}

	// 获取现有用户组
	group, err := c.userGroupRepo.GetByID(uint(id))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			ctx.JSON(http.StatusNotFound, v1.ErrorResponse{
				Code:    http.StatusNotFound,
				Message: "用户组不存在",
			})
		} else {
			c.logger.Error("获取用户组失败", zap.Error(err))
			ctx.JSON(http.StatusInternalServerError, v1.ErrorResponse{
				Code:    http.StatusInternalServerError,
				Message: "获取用户组失败",
			})
		}
		return
	}

	// 更新字段
	if req.Name != nil {
		// 检查用户组名是否已被其他用户组使用
		if existingGroup, _ := c.userGroupRepo.GetByName(*req.Name); existingGroup != nil && existingGroup.ID != group.ID {
			ctx.JSON(http.StatusBadRequest, v1.ErrorResponse{
				Code:    http.StatusBadRequest,
				Message: "用户组名已存在",
			})
			return
		}
		group.Name = *req.Name
	}
	if req.Description != nil {
		group.Description = *req.Description
	}
	if req.Status != nil {
		group.Status = *req.Status
	}

	if err := c.userGroupRepo.Update(group); err != nil {
		c.logger.Error("更新用户组失败", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, v1.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "更新用户组失败",
			Error:   err.Error(),
		})
		return
	}

	response := c.convertToUserGroupResponse(group)
	c.logger.Info("用户组更新成功", zap.Uint("group_id", group.ID))

	ctx.JSON(http.StatusOK, v1.CommonResponse{
		Code:    http.StatusOK,
		Message: "用户组更新成功",
		Data:    response,
	})
}

// DeleteUserGroup 删除用户组
// @Summary 删除用户组
// @Description 软删除用户组
// @Tags 用户组管理
// @Accept json
// @Produce json
// @Param id path int true "用户组ID"
// @Success 200 {object} v1.CommonResponse
// @Failure 404 {object} v1.ErrorResponse
// @Failure 500 {object} v1.ErrorResponse
// @Router /api/v1/user-groups/{id} [delete]
func (c *UserGroupController) DeleteUserGroup(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, v1.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "无效的用户组ID",
		})
		return
	}

	// 检查用户组是否存在
	if _, err := c.userGroupRepo.GetByID(uint(id)); err != nil {
		if err == gorm.ErrRecordNotFound {
			ctx.JSON(http.StatusNotFound, v1.ErrorResponse{
				Code:    http.StatusNotFound,
				Message: "用户组不存在",
			})
		} else {
			c.logger.Error("获取用户组失败", zap.Error(err))
			ctx.JSON(http.StatusInternalServerError, v1.ErrorResponse{
				Code:    http.StatusInternalServerError,
				Message: "获取用户组失败",
			})
		}
		return
	}

	// TODO: 检查用户组下是否还有用户，如果有则不允许删除
	// 这里可以添加业务逻辑检查

	if err := c.userGroupRepo.Delete(uint(id)); err != nil {
		c.logger.Error("删除用户组失败", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, v1.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "删除用户组失败",
			Error:   err.Error(),
		})
		return
	}

	c.logger.Info("用户组删除成功", zap.Uint("group_id", uint(id)))
	ctx.JSON(http.StatusOK, v1.CommonResponse{
		Code:    http.StatusOK,
		Message: "用户组删除成功",
	})
}

// ListUserGroups 获取用户组列表
// @Summary 获取用户组列表
// @Description 分页获取用户组列表
// @Tags 用户组管理
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param size query int false "每页数量" default(10)
// @Success 200 {object} v1.CommonResponse{data=v1.UserGroupListResponse}
// @Failure 500 {object} v1.ErrorResponse
// @Router /api/v1/user-groups [get]
func (c *UserGroupController) ListUserGroups(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(ctx.DefaultQuery("size", "10"))

	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 10
	}

	offset := (page - 1) * size
	groups, total, err := c.userGroupRepo.List(offset, size)
	if err != nil {
		c.logger.Error("获取用户组列表失败", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, v1.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "获取用户组列表失败",
			Error:   err.Error(),
		})
		return
	}

	// 转换为响应格式
	groupResponses := make([]v1.UserGroupResponse, len(groups))
	for i, group := range groups {
		groupResponses[i] = c.convertToUserGroupResponse(group)
	}

	response := v1.UserGroupListResponse{
		Groups: groupResponses,
		Total:  total,
		Page:   page,
		Size:   size,
	}

	ctx.JSON(http.StatusOK, v1.CommonResponse{
		Code:    http.StatusOK,
		Message: "获取成功",
		Data:    response,
	})
}

// convertToUserGroupResponse 转换用户组模型为响应格式
func (c *UserGroupController) convertToUserGroupResponse(group *models.UserGroup) v1.UserGroupResponse {
	return v1.UserGroupResponse{
		ID:          group.ID,
		Name:        group.Name,
		Description: group.Description,
		Status:      group.Status,
		CreatedAt:   group.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   group.UpdatedAt.Format(time.RFC3339),
	}
}