package services

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"pnas/internal/database"
	"pnas/internal/models"
)

// RecycleBinService handles recycle bin operations
type RecycleBinService struct {
	db            *gorm.DB
	permissionSvc *FilePermissionService
}

// NewRecycleBinService creates a new recycle bin service
func NewRecycleBinService() *RecycleBinService {
	return &RecycleBinService{
		db:            database.DB,
		permissionSvc: NewFilePermissionService(),
	}
}

// MoveToRecycleBin moves a file to recycle bin (soft delete)
func (s *RecycleBinService) MoveToRecycleBin(userID, fileID string) error {
	// Get file
	var file models.File
	if err := s.db.Where("id = ? AND is_deleted = ?", fileID, false).First(&file).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("file not found")
		}
		return err
	}

	// Check permissions
	hasPermission, err := s.permissionSvc.CheckPermission(userID, fileID, PermissionDelete)
	if err != nil {
		return fmt.Errorf("permission check failed: %w", err)
	}
	if !hasPermission {
		return fmt.Errorf("user does not have delete permission")
	}

	// Start transaction
	return s.db.Transaction(func(tx *gorm.DB) error {
		now := time.Now()
		autoDeleteAt := now.AddDate(0, 0, RecycleBinDays)

		// Mark file as deleted
		updates := map[string]interface{}{
			"is_deleted": true,
			"deleted_at": &now,
			"deleted_by": &userID,
		}

		if err := tx.Model(&file).Updates(updates).Error; err != nil {
			return fmt.Errorf("failed to mark file as deleted: %w", err)
		}

		// Create recycle bin entry
		recycleBinEntry := models.RecycleBin{
			Base: models.Base{
				ID: uuid.New().String(),
			},
			FileID:       fileID,
			OriginalPath: file.Path,
			OriginalName: file.Name,
			DeletedBy:    userID,
			DeletedAt:    now,
			Size:         file.Size,
			AutoDeleteAt: autoDeleteAt,
		}

		if err := tx.Create(&recycleBinEntry).Error; err != nil {
			return fmt.Errorf("failed to create recycle bin entry: %w", err)
		}

		// Update file index
		tx.Model(&models.FileIndex{}).
			Where("file_id = ?", fileID).
			Update("is_deleted", true)

		// Update user stats
		s.updateUserStatsAfterDelete(tx, file.OwnerID, file.Size)

		// If directory, recursively delete children
		if file.IsDirectory {
			if err := s.moveChildrenToRecycleBin(tx, userID, fileID, now, autoDeleteAt); err != nil {
				return fmt.Errorf("failed to delete children: %w", err)
			}
		}

		return nil
	})
}

// ListRecycleBin lists files in recycle bin for a user
func (s *RecycleBinService) ListRecycleBin(userID string, page, pageSize int) ([]models.RecycleBin, int64, error) {
	// Validate pagination
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 50
	}
	if pageSize > MaxListPageSize {
		pageSize = MaxListPageSize
	}

	// Check if user is admin
	isAdmin, _ := s.permissionSvc.isAdmin(userID)

	query := s.db.Model(&models.RecycleBin{}).Preload("File")

	// Non-admin users can only see their own deleted files
	if !isAdmin {
		query = query.Where("deleted_by = ?", userID)
	}

	// Count total
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get results
	var items []models.RecycleBin
	offset := (page - 1) * pageSize
	if err := query.
		Order("deleted_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&items).Error; err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

// RestoreFromRecycleBin restores a file from recycle bin
func (s *RecycleBinService) RestoreFromRecycleBin(userID, fileID string) error {
	// Get recycle bin entry
	var recycleBinEntry models.RecycleBin
	if err := s.db.Where("file_id = ?", fileID).First(&recycleBinEntry).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("file not found in recycle bin")
		}
		return err
	}

	// Check if user is admin or the one who deleted the file
	isAdmin, _ := s.permissionSvc.isAdmin(userID)
	if !isAdmin && recycleBinEntry.DeletedBy != userID {
		return fmt.Errorf("user does not have permission to restore this file")
	}

	// Get file
	var file models.File
	if err := s.db.Where("id = ?", fileID).First(&file).Error; err != nil {
		return fmt.Errorf("file record not found: %w", err)
	}

	// Check if original path is available
	var existingFile models.File
	if err := s.db.Where("path = ? AND is_deleted = ?", recycleBinEntry.OriginalPath, false).
		First(&existingFile).Error; err == nil {
		return fmt.Errorf("a file already exists at the original path")
	}

	// Start transaction
	return s.db.Transaction(func(tx *gorm.DB) error {
		// Restore file
		updates := map[string]interface{}{
			"is_deleted": false,
			"deleted_at": nil,
			"deleted_by": nil,
		}

		if err := tx.Model(&file).Updates(updates).Error; err != nil {
			return fmt.Errorf("failed to restore file: %w", err)
		}

		// Remove from recycle bin
		if err := tx.Delete(&recycleBinEntry).Error; err != nil {
			return fmt.Errorf("failed to remove from recycle bin: %w", err)
		}

		// Update file index
		tx.Model(&models.FileIndex{}).
			Where("file_id = ?", fileID).
			Update("is_deleted", false)

		// Update user stats
		s.updateUserStatsAfterRestore(tx, file.OwnerID, file.Size)

		// If directory, recursively restore children
		if file.IsDirectory {
			if err := s.restoreChildren(tx, fileID); err != nil {
				return fmt.Errorf("failed to restore children: %w", err)
			}
		}

		return nil
	})
}

// PermanentlyDelete permanently deletes a file from recycle bin
func (s *RecycleBinService) PermanentlyDelete(userID, fileID string) error {
	// Get recycle bin entry
	var recycleBinEntry models.RecycleBin
	if err := s.db.Where("file_id = ?", fileID).First(&recycleBinEntry).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("file not found in recycle bin")
		}
		return err
	}

	// Check if user is admin or the one who deleted the file
	isAdmin, _ := s.permissionSvc.isAdmin(userID)
	if !isAdmin && recycleBinEntry.DeletedBy != userID {
		return fmt.Errorf("user does not have permission to permanently delete this file")
	}

	// Get file
	var file models.File
	if err := s.db.Where("id = ?", fileID).First(&file).Error; err != nil {
		return fmt.Errorf("file record not found: %w", err)
	}

	// Start transaction
	return s.db.Transaction(func(tx *gorm.DB) error {
		// Delete physical file
		if err := s.deletePhysicalFile(file.StoragePath, file.IsDirectory); err != nil {
			return fmt.Errorf("failed to delete physical file: %w", err)
		}

		// If directory, recursively delete children
		if file.IsDirectory {
			if err := s.permanentlyDeleteChildren(tx, fileID); err != nil {
				return fmt.Errorf("failed to delete children: %w", err)
			}
		}

		// Delete file permissions
		if err := tx.Where("file_id = ?", fileID).Delete(&models.FilePermission{}).Error; err != nil {
			return fmt.Errorf("failed to delete file permissions: %w", err)
		}

		// Delete file versions
		if err := tx.Where("file_id = ?", fileID).Delete(&models.FileVersion{}).Error; err != nil {
			return fmt.Errorf("failed to delete file versions: %w", err)
		}

		// Delete file shares
		if err := tx.Where("file_id = ?", fileID).Delete(&models.FileShare{}).Error; err != nil {
			return fmt.Errorf("failed to delete file shares: %w", err)
		}

		// Delete file activities
		if err := tx.Where("file_id = ?", fileID).Delete(&models.FileActivity{}).Error; err != nil {
			return fmt.Errorf("failed to delete file activities: %w", err)
		}

		// Delete file index
		if err := tx.Where("file_id = ?", fileID).Delete(&models.FileIndex{}).Error; err != nil {
			return fmt.Errorf("failed to delete file index: %w", err)
		}

		// Delete recycle bin entry
		if err := tx.Delete(&recycleBinEntry).Error; err != nil {
			return fmt.Errorf("failed to remove from recycle bin: %w", err)
		}

		// Delete file record
		if err := tx.Delete(&file).Error; err != nil {
			return fmt.Errorf("failed to delete file record: %w", err)
		}

		return nil
	})
}

// EmptyRecycleBin permanently deletes all files in recycle bin for a user
func (s *RecycleBinService) EmptyRecycleBin(userID string) (int64, error) {
	// Get all files in recycle bin for user
	var entries []models.RecycleBin
	query := s.db.Where("deleted_by = ?", userID)

	// Check if user is admin
	isAdmin, _ := s.permissionSvc.isAdmin(userID)
	if isAdmin {
		// Admin can empty entire recycle bin
		query = s.db.Model(&models.RecycleBin{})
	}

	if err := query.Find(&entries).Error; err != nil {
		return 0, err
	}

	var deletedCount int64
	for _, entry := range entries {
		if err := s.PermanentlyDelete(userID, entry.FileID); err != nil {
			// Log error but continue
			continue
		}
		deletedCount++
	}

	return deletedCount, nil
}

// CleanupExpiredFiles permanently deletes files that have been in recycle bin for too long
// This should be run as a scheduled job
func (s *RecycleBinService) CleanupExpiredFiles() (int64, error) {
	now := time.Now()

	// Get expired files
	var entries []models.RecycleBin
	if err := s.db.Where("auto_delete_at < ?", now).Find(&entries).Error; err != nil {
		return 0, err
	}

	var deletedCount int64
	for _, entry := range entries {
		// Use admin context for cleanup
		if err := s.permanentlyDeleteInternal(entry.FileID); err != nil {
			// Log error but continue
			continue
		}
		deletedCount++
	}

	return deletedCount, nil
}

// Helper functions

// moveChildrenToRecycleBin recursively moves children to recycle bin
func (s *RecycleBinService) moveChildrenToRecycleBin(tx *gorm.DB, userID, parentID string, deletedAt time.Time, autoDeleteAt time.Time) error {
	var children []models.File
	if err := tx.Where("parent_id = ? AND is_deleted = ?", parentID, false).Find(&children).Error; err != nil {
		return err
	}

	for _, child := range children {
		// Mark as deleted
		updates := map[string]interface{}{
			"is_deleted": true,
			"deleted_at": &deletedAt,
			"deleted_by": &userID,
		}

		if err := tx.Model(&child).Updates(updates).Error; err != nil {
			return err
		}

		// Create recycle bin entry
		recycleBinEntry := models.RecycleBin{
			Base: models.Base{
				ID: uuid.New().String(),
			},
			FileID:       child.ID,
			OriginalPath: child.Path,
			OriginalName: child.Name,
			DeletedBy:    userID,
			DeletedAt:    deletedAt,
			Size:         child.Size,
			AutoDeleteAt: autoDeleteAt,
		}

		if err := tx.Create(&recycleBinEntry).Error; err != nil {
			return err
		}

		// Update index
		tx.Model(&models.FileIndex{}).
			Where("file_id = ?", child.ID).
			Update("is_deleted", true)

		// Recurse for directories
		if child.IsDirectory {
			if err := s.moveChildrenToRecycleBin(tx, userID, child.ID, deletedAt, autoDeleteAt); err != nil {
				return err
			}
		}
	}

	return nil
}

// restoreChildren recursively restores children
func (s *RecycleBinService) restoreChildren(tx *gorm.DB, parentID string) error {
	var children []models.File
	if err := tx.Where("parent_id = ? AND is_deleted = ?", parentID, true).Find(&children).Error; err != nil {
		return err
	}

	for _, child := range children {
		// Restore file
		updates := map[string]interface{}{
			"is_deleted": false,
			"deleted_at": nil,
			"deleted_by": nil,
		}

		if err := tx.Model(&child).Updates(updates).Error; err != nil {
			return err
		}

		// Remove from recycle bin
		if err := tx.Where("file_id = ?", child.ID).Delete(&models.RecycleBin{}).Error; err != nil {
			return err
		}

		// Update index
		tx.Model(&models.FileIndex{}).
			Where("file_id = ?", child.ID).
			Update("is_deleted", false)

		// Recurse for directories
		if child.IsDirectory {
			if err := s.restoreChildren(tx, child.ID); err != nil {
				return err
			}
		}
	}

	return nil
}

// permanentlyDeleteChildren recursively deletes children
func (s *RecycleBinService) permanentlyDeleteChildren(tx *gorm.DB, parentID string) error {
	var children []models.File
	if err := tx.Where("parent_id = ? AND is_deleted = ?", parentID, true).Find(&children).Error; err != nil {
		return err
	}

	for _, child := range children {
		// Delete physical file
		if err := s.deletePhysicalFile(child.StoragePath, child.IsDirectory); err != nil {
			return err
		}

		// Recurse for directories
		if child.IsDirectory {
			if err := s.permanentlyDeleteChildren(tx, child.ID); err != nil {
				return err
			}
		}

		// Delete related records
		tx.Where("file_id = ?", child.ID).Delete(&models.FilePermission{})
		tx.Where("file_id = ?", child.ID).Delete(&models.FileVersion{})
		tx.Where("file_id = ?", child.ID).Delete(&models.FileShare{})
		tx.Where("file_id = ?", child.ID).Delete(&models.FileActivity{})
		tx.Where("file_id = ?", child.ID).Delete(&models.FileIndex{})
		tx.Where("file_id = ?", child.ID).Delete(&models.RecycleBin{})

		// Delete file record
		if err := tx.Delete(&child).Error; err != nil {
			return err
		}
	}

	return nil
}

// deletePhysicalFile deletes the physical file or directory
func (s *RecycleBinService) deletePhysicalFile(path string, isDirectory bool) error {
	if isDirectory {
		return os.RemoveAll(path)
	}
	return os.Remove(path)
}

// permanentlyDeleteInternal permanently deletes a file (internal, no permission check)
func (s *RecycleBinService) permanentlyDeleteInternal(fileID string) error {
	var file models.File
	if err := s.db.Where("id = ?", fileID).First(&file).Error; err != nil {
		return err
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		// Delete physical file
		if err := s.deletePhysicalFile(file.StoragePath, file.IsDirectory); err != nil {
			return err
		}

		// Delete related records
		tx.Where("file_id = ?", fileID).Delete(&models.FilePermission{})
		tx.Where("file_id = ?", fileID).Delete(&models.FileVersion{})
		tx.Where("file_id = ?", fileID).Delete(&models.FileShare{})
		tx.Where("file_id = ?", fileID).Delete(&models.FileActivity{})
		tx.Where("file_id = ?", fileID).Delete(&models.FileIndex{})
		tx.Where("file_id = ?", fileID).Delete(&models.RecycleBin{})

		// Delete file record
		return tx.Delete(&file).Error
	})
}

// updateUserStatsAfterDelete updates user stats after deletion
func (s *RecycleBinService) updateUserStatsAfterDelete(tx *gorm.DB, userID string, size int64) {
	tx.Model(&models.StorageStats{}).
		Where("user_id = ?", userID).
		Updates(map[string]interface{}{
			"deleted_files": gorm.Expr("deleted_files + 1"),
			"deleted_size":  gorm.Expr("deleted_size + ?", size),
			"total_files":   gorm.Expr("total_files - 1"),
			"total_size":    gorm.Expr("total_size - ?", size),
		})
}

// updateUserStatsAfterRestore updates user stats after restoration
func (s *RecycleBinService) updateUserStatsAfterRestore(tx *gorm.DB, userID string, size int64) {
	tx.Model(&models.StorageStats{}).
		Where("user_id = ?", userID).
		Updates(map[string]interface{}{
			"deleted_files": gorm.Expr("deleted_files - 1"),
			"deleted_size":  gorm.Expr("deleted_size - ?", size),
			"total_files":   gorm.Expr("total_files + 1"),
			"total_size":    gorm.Expr("total_size + ?", size),
		})
}
