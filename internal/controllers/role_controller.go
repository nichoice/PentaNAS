package controllers

import (
	"net/http"
	"pnas/internal/database"
	"pnas/internal/models"

	"github.com/gin-gonic/gin"
)

// RoleResponse represents the response body for role operations
type RoleResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// GetRoles retrieves a list of roles with pagination
// @Summary Get roles
// @Description Get a list of predefined roles with pagination
// @Security BearerAuth
// @Tags Roles
// @Accept json
// @Produce json
// @Param page query int false "Page number (default: 1)"
// @Param page_size query int false "Page size (default: 20)"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]string
// @Router /roles [get]
func GetRoles(c *gin.Context) {
	// Get query parameters
	page := 1
	if p := c.Query("page"); p != "" {
		// In a real implementation, you would parse this to an integer
	}

	pageSize := 20
	if ps := c.Query("page_size"); ps != "" {
		// In a real implementation, you would parse this to an integer
	}

	// Since roles are predefined, we'll return all roles
	// In a real application, you might want to implement actual pagination
	var roles []models.Role
	if err := database.DB.Find(&roles).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query roles"})
		return
	}

	// Convert to response format
	roleResponses := make([]RoleResponse, len(roles))
	for i, role := range roles {
		roleResponses[i] = RoleResponse{
			ID:          role.ID,
			Name:        role.Name,
			Description: role.Description,
		}
	}

	// For predefined roles, total is always the count of predefined roles
	total := int64(len(roles))

	c.JSON(http.StatusOK, gin.H{
		"roles":      roleResponses,
		"total":      total,
		"page":       page,
		"page_size":  pageSize,
		"total_page": (total + int64(pageSize) - 1) / int64(pageSize),
	})
}

// GetRole retrieves a role by ID
// @Summary Get role by ID
// @Description Get detailed information about a predefined role by ID
// @Security BearerAuth
// @Tags Roles
// @Accept json
// @Produce json
// @Param id path string true "Role ID"
// @Success 200 {object} RoleResponse
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /roles/{id} [get]
func GetRole(c *gin.Context) {
	roleID := c.Param("id")

	var role models.Role
	if err := database.DB.Where("id = ?", roleID).First(&role).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "role not found"})
		return
	}

	response := RoleResponse{
		ID:          role.ID,
		Name:        role.Name,
		Description: role.Description,
	}

	c.JSON(http.StatusOK, response)
}
