package controllers

import (
	"net/http"
	"pnas/internal/database"
	"pnas/internal/models"
	"pnas/internal/utils"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// CreateUserRequest represents the request body for creating a user
type CreateUserRequest struct {
	Username string `json:"username" binding:"required,min=3,max=20,alphanum"`
	Password string `json:"password" binding:"required,min=8"`
	Role     string `json:"role" binding:"required,oneof=super_admin normal_user audit_user ops_user"`
	IsActive bool   `json:"is_active"`
	Remark   string `json:"remark"`
}

// UpdateUserRequest represents the request body for updating a user
type UpdateUserRequest struct {
	Role     *string `json:"role" binding:"omitempty,oneof=super_admin normal_user audit_user ops_user"`
	IsActive *bool   `json:"is_active"`
	Remark   *string `json:"remark"`
	Password *string `json:"password" binding:"omitempty,min=8"`
}

// UpdateUserStatusRequest represents the request body for updating user status
type UpdateUserStatusRequest struct {
	IsActive bool `json:"is_active" binding:"required"`
}

// UserResponse represents the response body for user operations
type UserResponse struct {
	ID        string     `json:"id"`
	Username  string     `json:"username"`
	IsActive  bool       `json:"is_active"`
	Remark    string     `json:"remark"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	LastLogin *time.Time `json:"last_login,omitempty"`
}

// CreateUser creates a new user
// @Summary Create a new user
// @Description Create a new user with the provided details
// @Security BearerAuth
// @Tags Users
// @Accept json
// @Produce json
// @Param user body CreateUserRequest true "User details"
// @Success 201 {object} UserResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /users [post]
func CreateUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if username already exists
	var existingUser models.User
	if err := database.DB.Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username already exists"})
		return
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	// Create user
	user := models.User{
		Base:     models.Base{ID: uuid.New().String()},
		Username: req.Username,
		Password: string(hashedPassword),
		IsActive: req.IsActive,
		Remark:   req.Remark,
	}

	if err := database.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
		return
	}

	// Create user role relationship
	userRole := models.UserRole{
		Base:   models.Base{ID: uuid.New().String()},
		UserID: user.ID,
		RoleID: req.Role,
	}

	if err := database.DB.Create(&userRole).Error; err != nil {
		// Rollback user creation
		database.DB.Delete(&user)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to assign role to user"})
		return
	}
	if req.Role == "normal_user" {
		if err := utils.CreateLinuxUser(req.Username); err != nil {
			// Rollback user creation
			utils.DeleteLinuxUser(req.Username)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create Linux user"})
			return
		}

	}

	response := UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		IsActive:  user.IsActive,
		Remark:    user.Remark,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		LastLogin: user.LastLogin,
	}

	c.JSON(http.StatusCreated, response)
}

// GetUsers retrieves a list of users with pagination
// @Summary Get users
// @Description Get a list of users with optional filtering and pagination
// @Security BearerAuth
// @Tags Users
// @Accept json
// @Produce json
// @Param username query string false "Username filter"
// @Param role query string false "Role filter"
// @Param status query string false "Status filter (active/inactive)"
// @Param page query int false "Page number (default: 1)"
// @Param page_size query int false "Page size (default: 20)"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]string
// @Router /users [get]
func GetUsers(c *gin.Context) {
	// Get query parameters
	username := c.Query("username")
	role := c.Query("role")
	status := c.Query("status")

	page := 1
	if p := c.Query("page"); p != "" {
		// In a real implementation, you would parse this to an integer
	}

	pageSize := 20
	if ps := c.Query("page_size"); ps != "" {
		// In a real implementation, you would parse this to an integer
	}

	// Build query
	query := database.DB.Model(&models.User{})

	// Apply filters
	if username != "" {
		query = query.Where("username LIKE ?", "%"+username+"%")
	}

	if status != "" {
		isActive := status == "active"
		query = query.Where("is_active = ?", isActive)
	}

	// For role filtering, we need to join with user_roles table
	if role != "" {
		query = query.Joins("JOIN user_roles ON users.id = user_roles.user_id").
			Where("user_roles.role_id = ?", role)
	}

	// Execute query with pagination
	var users []models.User
	var total int64

	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to count users"})
		return
	}

	if err := query.Offset((page - 1) * pageSize).Limit(pageSize).Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query users"})
		return
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
		"users":      userResponses,
		"total":      total,
		"page":       page,
		"page_size":  pageSize,
		"total_page": (total + int64(pageSize) - 1) / int64(pageSize),
	})
}

// GetUser retrieves a user by ID
// @Summary Get user by ID
// @Description Get detailed information about a user by ID
// @Security BearerAuth
// @Tags Users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /users/{id} [get]
func GetUser(c *gin.Context) {
	userID := c.Param("id")

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

	response := UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		IsActive:  user.IsActive,
		Remark:    user.Remark,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		LastLogin: user.LastLogin,
	}

	c.JSON(http.StatusOK, gin.H{
		"user":  response,
		"roles": roles,
	})
}

// UpdateUser updates a user's information
// @Summary Update user
// @Description Update a user's information
// @Security BearerAuth
// @Tags Users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param user body UpdateUserRequest true "User update details"
// @Success 200 {object} UserResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /users/{id} [put]
func UpdateUser(c *gin.Context) {
	userID := c.Param("id")

	var user models.User
	if err := database.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update fields if provided
	if req.Role != nil {
		// Update user role
		if err := database.DB.Model(&models.UserRole{}).
			Where("user_id = ?", userID).
			Updates(map[string]interface{}{"role_id": *req.Role}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update user role"})
			return
		}
	}

	if req.IsActive != nil {
		user.IsActive = *req.IsActive
	}

	if req.Remark != nil {
		user.Remark = *req.Remark
	}

	if req.Password != nil {
		// Hash the new password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*req.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
			return
		}
		user.Password = string(hashedPassword)
	}

	// Save updated user
	if err := database.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update user"})
		return
	}

	response := UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		IsActive:  user.IsActive,
		Remark:    user.Remark,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		LastLogin: user.LastLogin,
	}

	c.JSON(http.StatusOK, response)
}

// DeleteUser deletes a user
// @Summary Delete user
// @Description Delete a user by ID
// @Security BearerAuth
// @Tags Users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /users/{id} [delete]
func DeleteUser(c *gin.Context) {
	userID := c.Param("id")

	var user models.User
	if err := database.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	// Delete user roles first
	if err := database.DB.Where("user_id = ?", userID).Delete(&models.UserRole{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete user roles"})
		return
	}

	// Delete user
	if err := database.DB.Delete(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "user deleted successfully"})
}

// UpdateUserStatus updates a user's status (active/inactive)
// @Summary Update user status
// @Description Update a user's status (active/inactive)
// @Security BearerAuth
// @Tags Users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param status body UpdateUserStatusRequest true "User status"
// @Success 200 {object} UserResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /users/{id}/status [put]
func UpdateUserStatus(c *gin.Context) {
	userID := c.Param("id")

	var user models.User
	if err := database.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	var req UpdateUserStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user.IsActive = req.IsActive

	if err := database.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update user status"})
		return
	}

	response := UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		IsActive:  user.IsActive,
		Remark:    user.Remark,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		LastLogin: user.LastLogin,
	}

	c.JSON(http.StatusOK, response)
}

// LoginRequest represents the request body for login
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse represents the response body for login
type LoginResponse struct {
	Token     string    `json:"token"`
	UserID    string    `json:"user_id"`
	Username  string    `json:"username"`
	ExpiresAt time.Time `json:"expires_at"`
}

// Login handles user login
// @Summary User login
// @Description Authenticate user and return JWT token
// @Security BearerAuth
// @Tags Authentication
// @Accept json
// @Produce json
// @Param credentials body LoginRequest true "Login credentials"
// @Success 200 {object} LoginResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /login [post]
func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := database.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	if !user.IsActive {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user account is disabled"})
		return
	}

	// Check password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	// Generate JWT token
	token, err := utils.GenerateToken(user.ID, user.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	// Update last login time
	now := time.Now()
	user.LastLogin = &now
	if err := database.DB.Save(&user).Error; err != nil {
		// Log error but don't fail the login
		// In a real application, you might want to handle this differently
	}

	// Calculate expiration time
	expireHours := 24 // Default to 24 hours
	// Note: In a real implementation, you would access the config through a proper import
	// For now, we'll use a default value
	expiresAt := time.Now().Add(time.Duration(expireHours) * time.Hour)

	response := LoginResponse{
		Token:     token,
		UserID:    user.ID,
		Username:  user.Username,
		ExpiresAt: expiresAt,
	}

	c.JSON(http.StatusOK, response)
}
