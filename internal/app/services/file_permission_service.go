package services

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"pnas/internal/database"
	"pnas/internal/models"
)

// FilePermissionService handles file permission logic
type FilePermissionService struct {
	db *gorm.DB
}

// NewFilePermissionService creates a new file permission service
func NewFilePermissionService() *FilePermissionService {
	return &FilePermissionService{
		db: database.DB,
	}
}

// Permission constants
const (
	PermissionRead   = "read"
	PermissionWrite  = "write"
	PermissionDelete = "delete"
	PermissionShare  = "share"
	PermissionAdmin  = "admin" // Full control
)

// CheckPermission checks if a user has specific permission on a file
// Returns true if user is admin or has the required permission
func (s *FilePermissionService) CheckPermission(userID, fileID, permission string) (bool, error) {
	// Check if user is admin
	isAdmin, err := s.isAdmin(userID)
	if err != nil {
		return false, err
	}
	if isAdmin {
		return true, nil
	}

	// Check if user is the owner
	var file models.File
	if err := s.db.Select("owner_id").Where("id = ? AND is_deleted = ?", fileID, false).First(&file).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, fmt.Errorf("file not found")
		}
		return false, err
	}

	if file.OwnerID == userID {
		return true, nil
	}

	// Check explicit user permission
	hasPermission, err := s.hasUserPermission(userID, fileID, permission)
	if err != nil {
		return false, err
	}
	if hasPermission {
		return true, nil
	}

	// Check role-based permission
	hasRolePermission, err := s.hasRolePermission(userID, fileID, permission)
	if err != nil {
		return false, err
	}

	return hasRolePermission, nil
}

// CheckPathPermission checks permission on a file path (including parent directories)
// This checks if user has permission to access the entire path
func (s *FilePermissionService) CheckPathPermission(userID, filePath, permission string) (bool, error) {
	// Get file by path
	var file models.File
	if err := s.db.Where("path = ? AND is_deleted = ?", filePath, false).First(&file).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, fmt.Errorf("file not found")
		}
		return false, err
	}

	return s.CheckPermission(userID, file.ID, permission)
}

// GrantUserPermission grants a permission to a user for a file
func (s *FilePermissionService) GrantUserPermission(fileID, userID, grantedBy, permission string, expiresAt *time.Time) error {
	// Check if granter has permission to grant
	canGrant, err := s.CheckPermission(grantedBy, fileID, PermissionAdmin)
	if err != nil {
		return err
	}
	if !canGrant {
		return fmt.Errorf("user does not have permission to grant access")
	}

	// Check if permission already exists
	var existing models.FilePermission
	result := s.db.Where("file_id = ? AND user_id = ? AND permission = ?", fileID, userID, permission).First(&existing)

	if result.Error == nil {
		// Update existing permission
		updates := map[string]interface{}{
			"granted_by": grantedBy,
			"expires_at": expiresAt,
		}
		return s.db.Model(&existing).Updates(updates).Error
	}

	// Create new permission
	perm := models.FilePermission{
		Base: models.Base{
			ID: generateID(),
		},
		FileID:     fileID,
		UserID:     &userID,
		Permission: permission,
		GrantedBy:  grantedBy,
		ExpiresAt:  expiresAt,
	}

	return s.db.Create(&perm).Error
}

// GrantRolePermission grants a permission to a role for a file
func (s *FilePermissionService) GrantRolePermission(fileID, roleID, grantedBy, permission string, expiresAt *time.Time) error {
	// Check if granter has permission to grant
	canGrant, err := s.CheckPermission(grantedBy, fileID, PermissionAdmin)
	if err != nil {
		return err
	}
	if !canGrant {
		return fmt.Errorf("user does not have permission to grant access")
	}

	// Check if permission already exists
	var existing models.FilePermission
	result := s.db.Where("file_id = ? AND role_id = ? AND permission = ?", fileID, roleID, permission).First(&existing)

	if result.Error == nil {
		// Update existing permission
		updates := map[string]interface{}{
			"granted_by": grantedBy,
			"expires_at": expiresAt,
		}
		return s.db.Model(&existing).Updates(updates).Error
	}

	// Create new permission
	perm := models.FilePermission{
		Base: models.Base{
			ID: generateID(),
		},
		FileID:     fileID,
		RoleID:     &roleID,
		Permission: permission,
		GrantedBy:  grantedBy,
		ExpiresAt:  expiresAt,
	}

	return s.db.Create(&perm).Error
}

// RevokeUserPermission revokes a user's permission on a file
func (s *FilePermissionService) RevokeUserPermission(fileID, userID, revokedBy, permission string) error {
	// Check if revoker has permission
	canRevoke, err := s.CheckPermission(revokedBy, fileID, PermissionAdmin)
	if err != nil {
		return err
	}
	if !canRevoke {
		return fmt.Errorf("user does not have permission to revoke access")
	}

	return s.db.Where("file_id = ? AND user_id = ? AND permission = ?", fileID, userID, permission).
		Delete(&models.FilePermission{}).Error
}

// RevokeRolePermission revokes a role's permission on a file
func (s *FilePermissionService) RevokeRolePermission(fileID, roleID, revokedBy, permission string) error {
	// Check if revoker has permission
	canRevoke, err := s.CheckPermission(revokedBy, fileID, PermissionAdmin)
	if err != nil {
		return err
	}
	if !canRevoke {
		return fmt.Errorf("user does not have permission to revoke access")
	}

	return s.db.Where("file_id = ? AND role_id = ? AND permission = ?", fileID, roleID, permission).
		Delete(&models.FilePermission{}).Error
}

// GetFilePermissions gets all permissions for a file
func (s *FilePermissionService) GetFilePermissions(fileID string) ([]models.FilePermission, error) {
	var permissions []models.FilePermission
	if err := s.db.Where("file_id = ?", fileID).
		Preload("User").
		Preload("Role").
		Find(&permissions).Error; err != nil {
		return nil, err
	}
	return permissions, nil
}

// GetUserPermissions gets all files a user has explicit permissions on
func (s *FilePermissionService) GetUserPermissions(userID string) ([]models.FilePermission, error) {
	var permissions []models.FilePermission
	if err := s.db.Where("user_id = ?", userID).
		Preload("File").
		Find(&permissions).Error; err != nil {
		return nil, err
	}
	return permissions, nil
}

// CleanupExpiredPermissions removes expired permissions (should be run periodically)
func (s *FilePermissionService) CleanupExpiredPermissions() (int64, error) {
	now := time.Now()
	result := s.db.Where("expires_at IS NOT NULL AND expires_at < ?", now).
		Delete(&models.FilePermission{})
	return result.RowsAffected, result.Error
}

// Helper functions

// isAdmin checks if user has admin role
func (s *FilePermissionService) isAdmin(userID string) (bool, error) {
	var count int64
	err := s.db.Table("user_roles").
		Joins("JOIN roles ON roles.id = user_roles.role_id").
		Where("user_roles.user_id = ? AND roles.name = ?", userID, "admin").
		Count(&count).Error
	return count > 0, err
}

// hasUserPermission checks if user has explicit permission
func (s *FilePermissionService) hasUserPermission(userID, fileID, permission string) (bool, error) {
	var count int64
	now := time.Now()

	// Check for specific permission or admin permission
	err := s.db.Model(&models.FilePermission{}).
		Where("file_id = ? AND user_id = ? AND (permission = ? OR permission = ?)",
			fileID, userID, permission, PermissionAdmin).
		Where("(expires_at IS NULL OR expires_at > ?)", now).
		Count(&count).Error

	return count > 0, err
}

// hasRolePermission checks if user has permission through role
func (s *FilePermissionService) hasRolePermission(userID, fileID, permission string) (bool, error) {
	var count int64
	now := time.Now()

	// Get user's role IDs
	var roleIDs []string
	if err := s.db.Table("user_roles").
		Select("role_id").
		Where("user_id = ?", userID).
		Pluck("role_id", &roleIDs).Error; err != nil {
		return false, err
	}

	if len(roleIDs) == 0 {
		return false, nil
	}

	// Check if any of user's roles have permission
	err := s.db.Model(&models.FilePermission{}).
		Where("file_id = ? AND role_id IN ? AND (permission = ? OR permission = ?)",
			fileID, roleIDs, permission, PermissionAdmin).
		Where("(expires_at IS NULL OR expires_at > ?)", now).
		Count(&count).Error

	return count > 0, err
}

// InheritParentPermissions copies permissions from parent directory to child
func (s *FilePermissionService) InheritParentPermissions(parentID, childID string) error {
	// Get parent permissions
	var parentPerms []models.FilePermission
	if err := s.db.Where("file_id = ?", parentID).Find(&parentPerms).Error; err != nil {
		return err
	}

	// Copy permissions to child
	for _, perm := range parentPerms {
		childPerm := models.FilePermission{
			Base: models.Base{
				ID: generateID(),
			},
			FileID:     childID,
			UserID:     perm.UserID,
			RoleID:     perm.RoleID,
			Permission: perm.Permission,
			GrantedBy:  perm.GrantedBy,
			ExpiresAt:  perm.ExpiresAt,
		}
		if err := s.db.Create(&childPerm).Error; err != nil {
			return err
		}
	}

	return nil
}

// Helper to generate ID (using existing UUID generation)
func generateID() string {
	return uuid.New().String()
}
