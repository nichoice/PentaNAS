package controllers

import (
	"net/http"
	"pnas/internal/database"
	"pnas/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AssignRoleRequest represents the request body for assigning a role to a user
type AssignRoleRequest struct {
	UserID string `json:"user_id" binding:"required"`
	RoleID string `json:"role_id" binding:"required,oneof=super_admin normal_user audit_user ops_user"`
}

// RevokeRoleRequest represents the request body for revoking a role from a user
type RevokeRoleRequest struct {
	UserID string `json:"user_id" binding:"required"`
	RoleID string `json:"role_id" binding:"required,oneof=super_admin normal_user audit_user ops_user"`
}

// AssignRole assigns a role to a user
// @Summary Assign role to user
// @Description Assign a predefined role to a user
// @Tags Authorization
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param assignment body AssignRoleRequest true "Role assignment details"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /auth/assign [post]
func AssignRole(c *gin.Context) {
	var req AssignRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if user exists
	var user models.User
	if err := database.DB.Where("id = ?", req.UserID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	// Check if role exists
	var role models.Role
	if err := database.DB.Where("id = ?", req.RoleID).First(&role).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "role not found"})
		return
	}

	// Check if user already has this role
	var existingUserRole models.UserRole
	if err := database.DB.Where("user_id = ? AND role_id = ?", req.UserID, req.RoleID).First(&existingUserRole).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user already has this role"})
		return
	}

	// Create user role relationship
	userRole := models.UserRole{
		Base:   models.Base{ID: uuid.New().String()},
		UserID: req.UserID,
		RoleID: req.RoleID,
	}

	if err := database.DB.Create(&userRole).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to assign role to user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "role assigned successfully",
		"user_id": req.UserID,
		"role_id": req.RoleID,
	})
}

// RevokeRole revokes a role from a user
// @Summary Revoke role from user
// @Description Revoke a role from a user
// @Tags Authorization
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param revocation body RevokeRoleRequest true "Role revocation details"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /auth/revoke [post]
func RevokeRole(c *gin.Context) {
	var req RevokeRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if user exists
	var user models.User
	if err := database.DB.Where("id = ?", req.UserID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	// Check if role exists
	var role models.Role
	if err := database.DB.Where("id = ?", req.RoleID).First(&role).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "role not found"})
		return
	}

	// Check if user has this role
	var userRole models.UserRole
	if err := database.DB.Where("user_id = ? AND role_id = ?", req.UserID, req.RoleID).First(&userRole).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user does not have this role"})
		return
	}

	// Delete user role relationship
	if err := database.DB.Delete(&userRole).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to revoke role from user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "role revoked successfully",
		"user_id": req.UserID,
		"role_id": req.RoleID,
	})
}

// GetUserRoles retrieves roles assigned to a user
// @Summary Get user roles
// @Description Get roles assigned to a user
// @Tags Authorization
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param user_id path string true "User ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /auth/users/{user_id}/roles [get]
func GetUserRoles(c *gin.Context) {
	userID := c.Param("user_id")

	// Check if user exists
	var user models.User
	if err := database.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	// Get user roles
	var userRoles []models.UserRole
	if err := database.DB.Where("user_id = ?", userID).Find(&userRoles).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query user roles"})
		return
	}

	// Get role details
	roleIDs := make([]string, len(userRoles))
	for i, ur := range userRoles {
		roleIDs[i] = ur.RoleID
	}

	var roles []models.Role
	if len(roleIDs) > 0 {
		if err := database.DB.Where("id IN ?", roleIDs).Find(&roles).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query roles"})
			return
		}
	}

	roleResponses := make([]RoleResponse, len(roles))
	for i, role := range roles {
		roleResponses[i] = RoleResponse{
			ID:          role.ID,
			Name:        role.Name,
			Description: role.Description,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id": userID,
		"roles":   roleResponses,
	})
}

// GetRoleUsers retrieves users assigned to a role
// @Summary Get role users
// @Description Get users assigned to a role
// @Tags Authorization
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param role_id path string true "Role ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /auth/roles/{role_id}/users [get]
func GetRoleUsers(c *gin.Context) {
	roleID := c.Param("role_id")

	// Check if role exists
	var role models.Role
	if err := database.DB.Where("id = ?", roleID).First(&role).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "role not found"})
		return
	}

	// Get users with this role
	var userRoles []models.UserRole
	if err := database.DB.Where("role_id = ?", roleID).Find(&userRoles).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query user roles"})
		return
	}

	// Get user details
	userIDs := make([]string, len(userRoles))
	for i, ur := range userRoles {
		userIDs[i] = ur.UserID
	}

	var users []models.User
	if len(userIDs) > 0 {
		if err := database.DB.Where("id IN ?", userIDs).Find(&users).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query users"})
			return
		}
	}

	// Convert to response format
	userResponses := make([]UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = UserResponse{
			ID:        user.ID,
			Username:  user.Username,
			IsActive:  user.IsActive,
			Remark:    user.Remark,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			LastLogin: user.LastLogin,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"role_id": roleID,
		"users":   userResponses,
	})
}
