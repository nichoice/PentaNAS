package services

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"pnas/internal/database"
	"pnas/internal/models"
)

const (
	BaseStoragePath  = "/wuzhou"
	MaxFileSize      = 10 * 1024 * 1024 * 1024 // 10GB
	ChunkSize        = 8 * 1024 * 1024          // 8MB chunks for streaming
	RecycleBinDays   = 30                       // Days before permanent deletion
	MaxListPageSize  = 1000
	CacheExpiration  = 5 * time.Minute
)

// FileService handles file operations with performance optimizations
type FileService struct {
	db              *gorm.DB
	permissionSvc   *FilePermissionService
	recycleBinSvc   *RecycleBinService
}

// NewFileService creates a new file service
func NewFileService() *FileService {
	return &FileService{
		db:              database.DB,
		permissionSvc:   NewFilePermissionService(),
		recycleBinSvc:   NewRecycleBinService(),
	}
}

// CreateDirectory creates a new directory
func (s *FileService) CreateDirectory(userID, parentPath, name, description string) (*models.File, error) {
	// Build full path
	fullPath := filepath.Join(parentPath, name)
	if !strings.HasPrefix(fullPath, BaseStoragePath) {
		fullPath = filepath.Join(BaseStoragePath, fullPath)
	}

	// Check if directory already exists
	var existing models.File
	if err := s.db.Where("path = ? AND is_deleted = ?", fullPath, false).First(&existing).Error; err == nil {
		return nil, fmt.Errorf("directory already exists")
	}

	// Get or create parent directory
	var parentID *string
	if parentPath != "" && parentPath != BaseStoragePath {
		parent, err := s.ensureParentDirectory(userID, parentPath)
		if err != nil {
			return nil, fmt.Errorf("failed to ensure parent directory: %w", err)
		}
		parentID = &parent.ID
	}

	// Create physical directory
	if err := os.MkdirAll(fullPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create physical directory: %w", err)
	}

	// Create database record
	dir := &models.File{
		Base: models.Base{
			ID: uuid.New().String(),
		},
		Name:        name,
		Path:        fullPath,
		ParentID:    parentID,
		OwnerID:     userID,
		Size:        0,
		IsDirectory: true,
		Description: description,
		StoragePath: fullPath,
	}

	if err := s.db.Create(dir).Error; err != nil {
		os.RemoveAll(fullPath) // Cleanup on error
		return nil, fmt.Errorf("failed to create directory record: %w", err)
	}

	// Update index
	s.updateFileIndex(dir)

	// Update user stats
	s.updateUserStats(userID)

	return dir, nil
}

// UploadFile handles file upload with streaming and deduplication
func (s *FileService) UploadFile(userID, targetPath, description string, fileHeader *multipart.FileHeader) (*models.File, error) {
	// Check permissions on target directory
	hasPermission, err := s.permissionSvc.CheckPathPermission(userID, targetPath, PermissionWrite)
	if err != nil {
		return nil, fmt.Errorf("permission check failed: %w", err)
	}
	if !hasPermission {
		return nil, fmt.Errorf("user does not have write permission on target directory")
	}

	// Check file size
	if fileHeader.Size > MaxFileSize {
		return nil, fmt.Errorf("file size exceeds maximum allowed size")
	}

	// Check user quota
	if err := s.checkUserQuota(userID, fileHeader.Size); err != nil {
		return nil, err
	}

	// Open uploaded file
	src, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer src.Close()

	// Calculate hashes while uploading
	md5Hash, sha256Hash, err := calculateHashes(src)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate file hashes: %w", err)
	}

	// Reset file pointer
	src.Seek(0, 0)

	// Check for deduplication
	existing, err := s.findDuplicateFile(md5Hash, userID)
	if err == nil && existing != nil {
		// File already exists, create a reference instead of uploading again
		return s.createFileReference(userID, targetPath, fileHeader.Filename, existing)
	}

	// Build full path
	fullPath := filepath.Join(targetPath, fileHeader.Filename)
	if !strings.HasPrefix(fullPath, BaseStoragePath) {
		fullPath = filepath.Join(BaseStoragePath, fullPath)
	}

	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	// Create destination file
	dst, err := os.Create(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dst.Close()

	// Stream copy with progress tracking
	written, err := io.Copy(dst, src)
	if err != nil {
		os.Remove(fullPath) // Cleanup on error
		return nil, fmt.Errorf("failed to save file: %w", err)
	}

	// Get file info
	ext := filepath.Ext(fileHeader.Filename)
	mimeType := fileHeader.Header.Get("Content-Type")

	// Get parent ID
	var parentID *string
	parentDir := filepath.Dir(fullPath)
	if parentDir != BaseStoragePath {
		parent, err := s.getOrCreateDirectory(userID, parentDir)
		if err != nil {
			return nil, fmt.Errorf("failed to get parent directory: %w", err)
		}
		parentID = &parent.ID
	}

	// Create database record
	file := &models.File{
		Base: models.Base{
			ID: uuid.New().String(),
		},
		Name:        fileHeader.Filename,
		Path:        fullPath,
		ParentID:    parentID,
		OwnerID:     userID,
		Size:        written,
		MimeType:    mimeType,
		Extension:   ext,
		IsDirectory: false,
		MD5Hash:     md5Hash,
		SHA256Hash:  sha256Hash,
		StoragePath: fullPath,
		Description: description,
		Version:     1,
	}

	if err := s.db.Create(file).Error; err != nil {
		os.Remove(fullPath) // Cleanup on error
		return nil, fmt.Errorf("failed to create file record: %w", err)
	}

	// Update index
	s.updateFileIndex(file)

	// Update user stats
	s.updateUserStats(userID)

	// Log activity
	s.logActivity(file.ID, userID, "upload", nil)

	return file, nil
}

// DownloadFile prepares a file for download
func (s *FileService) DownloadFile(userID, fileID string) (*models.File, *os.File, error) {
	// Get file metadata
	var file models.File
	if err := s.db.Where("id = ? AND is_deleted = ?", fileID, false).First(&file).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil, fmt.Errorf("file not found")
		}
		return nil, nil, err
	}

	// Check if it's a directory
	if file.IsDirectory {
		return nil, nil, fmt.Errorf("cannot download a directory")
	}

	// Check permissions
	hasPermission, err := s.permissionSvc.CheckPermission(userID, fileID, PermissionRead)
	if err != nil {
		return nil, nil, err
	}
	if !hasPermission {
		return nil, nil, fmt.Errorf("user does not have read permission")
	}

	// Open file
	f, err := os.Open(file.StoragePath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open file: %w", err)
	}

	// Update statistics
	now := time.Now()
	s.db.Model(&file).Updates(map[string]interface{}{
		"download_count": gorm.Expr("download_count + 1"),
		"last_access_at": &now,
	})

	// Log activity
	s.logActivity(fileID, userID, "download", nil)

	return &file, f, nil
}

// ListFiles lists files in a directory with pagination and filtering
func (s *FileService) ListFiles(userID, parentPath string, page, pageSize int, filters map[string]interface{}) ([]models.File, int64, error) {
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

	// Build query
	query := s.db.Model(&models.File{}).Where("is_deleted = ?", false)

	// Check if user is admin
	isAdmin, _ := s.permissionSvc.isAdmin(userID)

	if !isAdmin {
		// Non-admin users: filter by owner or explicit permissions
		query = query.Where("owner_id = ? OR id IN (?)",
			userID,
			s.db.Table("file_permissions").
				Select("file_id").
				Where("user_id = ? OR role_id IN (?)",
					userID,
					s.db.Table("user_roles").Select("role_id").Where("user_id = ?", userID),
				),
		)
	}

	// Filter by parent path
	if parentPath != "" {
		if parentPath == "/" || parentPath == BaseStoragePath {
			query = query.Where("parent_id IS NULL OR parent_id = ''")
		} else {
			var parent models.File
			if err := s.db.Where("path = ? AND is_deleted = ?", parentPath, false).First(&parent).Error; err != nil {
				return nil, 0, fmt.Errorf("parent directory not found")
			}
			query = query.Where("parent_id = ?", parent.ID)
		}
	}

	// Apply filters
	if extension, ok := filters["extension"].(string); ok && extension != "" {
		query = query.Where("extension = ?", extension)
	}
	if isDir, ok := filters["is_directory"].(bool); ok {
		query = query.Where("is_directory = ?", isDir)
	}
	if minSize, ok := filters["min_size"].(int64); ok {
		query = query.Where("size >= ?", minSize)
	}
	if maxSize, ok := filters["max_size"].(int64); ok {
		query = query.Where("size <= ?", maxSize)
	}

	// Count total
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get files with pagination
	var files []models.File
	offset := (page - 1) * pageSize
	if err := query.
		Order("is_directory DESC, name ASC").
		Offset(offset).
		Limit(pageSize).
		Find(&files).Error; err != nil {
		return nil, 0, err
	}

	return files, total, nil
}

// SearchFiles searches files using optimized index
func (s *FileService) SearchFiles(userID, keyword string, filters map[string]interface{}, page, pageSize int) ([]models.File, int64, error) {
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

	// Use file index for fast search
	query := s.db.Table("file_index").
		Select("files.*").
		Joins("JOIN files ON files.id = file_index.file_id").
		Where("file_index.is_deleted = ?", false)

	// Check if user is admin
	isAdmin, _ := s.permissionSvc.isAdmin(userID)

	if !isAdmin {
		query = query.Where("file_index.owner_id = ? OR files.id IN (?)",
			userID,
			s.db.Table("file_permissions").
				Select("file_id").
				Where("user_id = ? OR role_id IN (?)",
					userID,
					s.db.Table("user_roles").Select("role_id").Where("user_id = ?", userID),
				),
		)
	}

	// Keyword search (case-insensitive)
	if keyword != "" {
		keywordLower := strings.ToLower(keyword)
		query = query.Where("file_index.name_lower LIKE ?", "%"+keywordLower+"%")
	}

	// Apply filters
	if extension, ok := filters["extension"].(string); ok && extension != "" {
		query = query.Where("file_index.extension = ?", extension)
	}
	if isDir, ok := filters["is_directory"].(bool); ok {
		query = query.Where("file_index.is_directory = ?", isDir)
	}

	// Count total
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get results with pagination
	var files []models.File
	offset := (page - 1) * pageSize
	if err := query.
		Order("file_index.updated_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&files).Error; err != nil {
		return nil, 0, err
	}

	return files, total, nil
}

// MoveFile moves a file or directory
func (s *FileService) MoveFile(userID, fileID, newParentPath string) error {
	// Get file
	var file models.File
	if err := s.db.Where("id = ? AND is_deleted = ?", fileID, false).First(&file).Error; err != nil {
		return fmt.Errorf("file not found")
	}

	// Check permissions on source
	hasPermission, err := s.permissionSvc.CheckPermission(userID, fileID, PermissionWrite)
	if err != nil {
		return err
	}
	if !hasPermission {
		return fmt.Errorf("user does not have write permission on source")
	}

	// Check permissions on destination
	hasDestPermission, err := s.permissionSvc.CheckPathPermission(userID, newParentPath, PermissionWrite)
	if err != nil {
		return err
	}
	if !hasDestPermission {
		return fmt.Errorf("user does not have write permission on destination")
	}

	// Build new path
	newPath := filepath.Join(newParentPath, file.Name)

	// Move physical file/directory
	if err := os.Rename(file.StoragePath, newPath); err != nil {
		return fmt.Errorf("failed to move file: %w", err)
	}

	// Get new parent ID
	var newParentID *string
	if newParentPath != BaseStoragePath {
		parent, err := s.getOrCreateDirectory(userID, newParentPath)
		if err != nil {
			os.Rename(newPath, file.StoragePath) // Rollback
			return fmt.Errorf("failed to get parent directory: %w", err)
		}
		newParentID = &parent.ID
	}

	// Update database
	updates := map[string]interface{}{
		"path":         newPath,
		"parent_id":    newParentID,
		"storage_path": newPath,
	}

	if err := s.db.Model(&file).Updates(updates).Error; err != nil {
		os.Rename(newPath, file.StoragePath) // Rollback
		return fmt.Errorf("failed to update file record: %w", err)
	}

	// Update index
	file.Path = newPath
	file.ParentID = newParentID
	file.StoragePath = newPath
	s.updateFileIndex(&file)

	// Log activity
	s.logActivity(fileID, userID, "move", map[string]string{
		"from": file.Path,
		"to":   newPath,
	})

	return nil
}

// RenameFile renames a file or directory
func (s *FileService) RenameFile(userID, fileID, newName string) error {
	// Get file
	var file models.File
	if err := s.db.Where("id = ? AND is_deleted = ?", fileID, false).First(&file).Error; err != nil {
		return fmt.Errorf("file not found")
	}

	// Check permissions
	hasPermission, err := s.permissionSvc.CheckPermission(userID, fileID, PermissionWrite)
	if err != nil {
		return err
	}
	if !hasPermission {
		return fmt.Errorf("user does not have write permission")
	}

	// Build new path
	newPath := filepath.Join(filepath.Dir(file.Path), newName)

	// Rename physical file/directory
	if err := os.Rename(file.StoragePath, newPath); err != nil {
		return fmt.Errorf("failed to rename file: %w", err)
	}

	// Update database
	ext := filepath.Ext(newName)
	updates := map[string]interface{}{
		"name":         newName,
		"path":         newPath,
		"storage_path": newPath,
		"extension":    ext,
	}

	if err := s.db.Model(&file).Updates(updates).Error; err != nil {
		os.Rename(newPath, file.StoragePath) // Rollback
		return fmt.Errorf("failed to update file record: %w", err)
	}

	// Update index
	file.Name = newName
	file.Path = newPath
	file.Extension = ext
	s.updateFileIndex(&file)

	// Log activity
	s.logActivity(fileID, userID, "rename", map[string]string{
		"old_name": file.Name,
		"new_name": newName,
	})

	return nil
}

// DeleteFile moves file to recycle bin
func (s *FileService) DeleteFile(userID, fileID string) error {
	// Delegate to recycle bin service
	return s.recycleBinSvc.MoveToRecycleBin(userID, fileID)
}

// Helper functions

// ensureParentDirectory ensures parent directory exists
func (s *FileService) ensureParentDirectory(userID, path string) (*models.File, error) {
	var parent models.File
	if err := s.db.Where("path = ? AND is_deleted = ?", path, false).First(&parent).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// Create parent directory
			return s.CreateDirectory(userID, filepath.Dir(path), filepath.Base(path), "")
		}
		return nil, err
	}
	return &parent, nil
}

// getOrCreateDirectory gets or creates a directory
func (s *FileService) getOrCreateDirectory(userID, path string) (*models.File, error) {
	var dir models.File
	if err := s.db.Where("path = ? AND is_deleted = ?", path, false).First(&dir).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return s.CreateDirectory(userID, filepath.Dir(path), filepath.Base(path), "")
		}
		return nil, err
	}
	return &dir, nil
}

// calculateHashes calculates MD5 and SHA256 hashes
func calculateHashes(r io.Reader) (string, string, error) {
	md5Hash := md5.New()
	sha256Hash := sha256.New()

	writer := io.MultiWriter(md5Hash, sha256Hash)
	if _, err := io.Copy(writer, r); err != nil {
		return "", "", err
	}

	return hex.EncodeToString(md5Hash.Sum(nil)),
		hex.EncodeToString(sha256Hash.Sum(nil)),
		nil
}

// findDuplicateFile finds a file with the same hash (for deduplication)
func (s *FileService) findDuplicateFile(md5Hash, userID string) (*models.File, error) {
	var file models.File
	if err := s.db.Where("md5_hash = ? AND owner_id = ? AND is_deleted = ?", md5Hash, userID, false).
		First(&file).Error; err != nil {
		return nil, err
	}
	return &file, nil
}

// createFileReference creates a file reference (for deduplication)
func (s *FileService) createFileReference(userID, targetPath, filename string, original *models.File) (*models.File, error) {
	fullPath := filepath.Join(targetPath, filename)

	// Create symlink
	if err := os.Symlink(original.StoragePath, fullPath); err != nil {
		return nil, err
	}

	// Create database record
	file := &models.File{
		Base: models.Base{
			ID: uuid.New().String(),
		},
		Name:        filename,
		Path:        fullPath,
		OwnerID:     userID,
		Size:        original.Size,
		MimeType:    original.MimeType,
		Extension:   original.Extension,
		IsDirectory: false,
		MD5Hash:     original.MD5Hash,
		SHA256Hash:  original.SHA256Hash,
		StoragePath: fullPath,
	}

	if err := s.db.Create(file).Error; err != nil {
		os.Remove(fullPath)
		return nil, err
	}

	s.updateFileIndex(file)
	s.updateUserStats(userID)

	return file, nil
}

// updateFileIndex updates file index for fast search
func (s *FileService) updateFileIndex(file *models.File) {
	index := models.FileIndex{
		Base: models.Base{
			ID: uuid.New().String(),
		},
		FileID:      file.ID,
		Name:        file.Name,
		NameLower:   strings.ToLower(file.Name),
		Extension:   file.Extension,
		Path:        file.Path,
		OwnerID:     file.OwnerID,
		Size:        file.Size,
		IsDirectory: file.IsDirectory,
		IsDeleted:   file.IsDeleted,
		UpdatedAt:   time.Now(),
	}

	// Upsert
	s.db.Where("file_id = ?", file.ID).
		Assign(index).
		FirstOrCreate(&index)
}

// updateUserStats updates user storage statistics
func (s *FileService) updateUserStats(userID string) {
	go func() {
		var stats models.StorageStats
		result := s.db.Where("user_id = ?", userID).First(&stats)

		var totalFiles, totalDirs int64
		var totalSize int64

		s.db.Model(&models.File{}).
			Where("owner_id = ? AND is_deleted = ? AND is_directory = ?", userID, false, false).
			Count(&totalFiles)

		s.db.Model(&models.File{}).
			Where("owner_id = ? AND is_deleted = ? AND is_directory = ?", userID, false, true).
			Count(&totalDirs)

		s.db.Model(&models.File{}).
			Where("owner_id = ? AND is_deleted = ?", userID, false).
			Select("COALESCE(SUM(size), 0)").
			Row().Scan(&totalSize)

		if result.Error == gorm.ErrRecordNotFound {
			stats = models.StorageStats{
				Base: models.Base{
					ID: uuid.New().String(),
				},
				UserID:           userID,
				TotalFiles:       totalFiles,
				TotalDirectories: totalDirs,
				TotalSize:        totalSize,
				LastUpdatedAt:    time.Now(),
			}
			s.db.Create(&stats)
		} else {
			s.db.Model(&stats).Updates(map[string]interface{}{
				"total_files":       totalFiles,
				"total_directories": totalDirs,
				"total_size":        totalSize,
				"last_updated_at":   time.Now(),
			})
		}
	}()
}

// checkUserQuota checks if user has enough quota
func (s *FileService) checkUserQuota(userID string, additionalSize int64) error {
	var stats models.StorageStats
	if err := s.db.Where("user_id = ?", userID).First(&stats).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil // No quota set
		}
		return err
	}

	if stats.QuotaLimit > 0 && (stats.TotalSize+additionalSize) > stats.QuotaLimit {
		return fmt.Errorf("storage quota exceeded")
	}

	return nil
}

// logActivity logs file activity
func (s *FileService) logActivity(fileID, userID, action string, details map[string]string) {
	go func() {
		activity := models.FileActivity{
			Base: models.Base{
				ID: uuid.New().String(),
			},
			FileID: fileID,
			UserID: userID,
			Action: action,
		}
		s.db.Create(&activity)
	}()
}
