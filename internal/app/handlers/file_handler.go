package handlers

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"pnas/internal/app/dto"
	"pnas/internal/app/services"
	"pnas/internal/models"
)

// FileHandler handles file management HTTP requests
type FileHandler struct {
	fileSvc       *services.FileService
	permissionSvc *services.FilePermissionService
	recycleBinSvc *services.RecycleBinService
}

// NewFileHandler creates a new file handler
func NewFileHandler() *FileHandler {
	return &FileHandler{
		fileSvc:       services.NewFileService(),
		permissionSvc: services.NewFilePermissionService(),
		recycleBinSvc: services.NewRecycleBinService(),
	}
}

// Directory Operations

// CreateDirectory creates a new directory
// @Summary Create directory
// @Tags File Management
// @Accept json
// @Produce json
// @Param directory body dto.CreateDirectoryRequest true "Directory creation request"
// @Success 201 {object} dto.FileResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /files/directories [post]
func (h *FileHandler) CreateDirectory(c *gin.Context) {
	userID := c.GetString("user_id")

	var req dto.CreateDirectoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid request",
			Details: err.Error(),
		})
		return
	}

	file, err := h.fileSvc.CreateDirectory(userID, req.ParentPath, req.Name, req.Description)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Failed to create directory",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, h.fileToResponse(file, userID))
}

// File Upload/Download

// UploadFile handles file upload with streaming
// @Summary Upload file
// @Tags File Management
// @Accept multipart/form-data
// @Produce json
// @Param target_path formData string true "Target directory path"
// @Param description formData string false "File description"
// @Param file formData file true "File to upload"
// @Success 201 {object} dto.FileResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /files/upload [post]
func (h *FileHandler) UploadFile(c *gin.Context) {
	userID := c.GetString("user_id")

	// Get form values
	targetPath := c.PostForm("target_path")
	description := c.PostForm("description")

	if targetPath == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "target_path is required",
		})
		return
	}

	// Get uploaded file
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "No file uploaded",
			Details: err.Error(),
		})
		return
	}

	// Upload file
	file, err := h.fileSvc.UploadFile(userID, targetPath, description, fileHeader)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Failed to upload file",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, h.fileToResponse(file, userID))
}

// DownloadFile handles file download with streaming
// @Summary Download file
// @Tags File Management
// @Produce octet-stream
// @Param id path string true "File ID"
// @Success 200 {file} binary
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /files/{id}/download [get]
func (h *FileHandler) DownloadFile(c *gin.Context) {
	userID := c.GetString("user_id")
	fileID := c.Param("id")

	file, f, err := h.fileSvc.DownloadFile(userID, fileID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Failed to download file",
			Details: err.Error(),
		})
		return
	}
	defer f.Close()

	// Set headers for download
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", file.Name))
	c.Header("Content-Type", file.MimeType)
	c.Header("Content-Length", strconv.FormatInt(file.Size, 10))

	// Stream file to response
	if _, err := io.Copy(c.Writer, f); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Failed to stream file",
			Details: err.Error(),
		})
		return
	}
}

// File Listing and Search

// ListFiles lists files in a directory
// @Summary List files
// @Tags File Management
// @Produce json
// @Param parent_path query string false "Parent directory path"
// @Param page query int false "Page number"
// @Param page_size query int false "Page size"
// @Param extension query string false "File extension filter"
// @Param is_directory query bool false "Filter by directory"
// @Param min_size query int false "Minimum file size"
// @Param max_size query int false "Maximum file size"
// @Success 200 {object} dto.ListResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /files [get]
func (h *FileHandler) ListFiles(c *gin.Context) {
	userID := c.GetString("user_id")

	var req dto.FileListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid query parameters",
			Details: err.Error(),
		})
		return
	}

	// Build filters
	filters := make(map[string]interface{})
	if req.Extension != "" {
		filters["extension"] = req.Extension
	}
	if req.IsDirectory != nil {
		filters["is_directory"] = *req.IsDirectory
	}
	if req.MinSize > 0 {
		filters["min_size"] = req.MinSize
	}
	if req.MaxSize > 0 {
		filters["max_size"] = req.MaxSize
	}

	files, total, err := h.fileSvc.ListFiles(userID, req.ParentPath, req.Page, req.PageSize, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Failed to list files",
			Details: err.Error(),
		})
		return
	}

	// Convert to responses
	items := make([]dto.FileResponse, len(files))
	for i, file := range files {
		items[i] = h.fileToResponse(&file, userID)
	}

	// Calculate pagination
	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 {
		pageSize = 50
	}
	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	c.JSON(http.StatusOK, dto.ListResponse{
		Items:      items,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	})
}

// SearchFiles searches files
// @Summary Search files
// @Tags File Management
// @Produce json
// @Param keyword query string true "Search keyword"
// @Param page query int false "Page number"
// @Param page_size query int false "Page size"
// @Param extension query string false "File extension filter"
// @Success 200 {object} dto.ListResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /files/search [get]
func (h *FileHandler) SearchFiles(c *gin.Context) {
	userID := c.GetString("user_id")

	var req dto.FileSearchRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid query parameters",
			Details: err.Error(),
		})
		return
	}

	// Build filters
	filters := make(map[string]interface{})
	if req.Extension != "" {
		filters["extension"] = req.Extension
	}
	if req.IsDirectory != nil {
		filters["is_directory"] = *req.IsDirectory
	}

	files, total, err := h.fileSvc.SearchFiles(userID, req.Keyword, filters, req.Page, req.PageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Failed to search files",
			Details: err.Error(),
		})
		return
	}

	// Convert to responses
	items := make([]dto.FileResponse, len(files))
	for i, file := range files {
		items[i] = h.fileToResponse(&file, userID)
	}

	// Calculate pagination
	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 {
		pageSize = 50
	}
	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	c.JSON(http.StatusOK, dto.ListResponse{
		Items:      items,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	})
}

// File Operations

// MoveFile moves a file or directory
// @Summary Move file
// @Tags File Management
// @Accept json
// @Produce json
// @Param move body dto.MoveFileRequest true "Move request"
// @Success 200 {object} dto.MessageResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /files/move [put]
func (h *FileHandler) MoveFile(c *gin.Context) {
	userID := c.GetString("user_id")

	var req dto.MoveFileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid request",
			Details: err.Error(),
		})
		return
	}

	if err := h.fileSvc.MoveFile(userID, req.FileID, req.NewParentPath); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Failed to move file",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.MessageResponse{
		Message: "File moved successfully",
	})
}

// RenameFile renames a file or directory
// @Summary Rename file
// @Tags File Management
// @Accept json
// @Produce json
// @Param id path string true "File ID"
// @Param rename body dto.RenameFileRequest true "Rename request"
// @Success 200 {object} dto.MessageResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /files/{id}/rename [put]
func (h *FileHandler) RenameFile(c *gin.Context) {
	userID := c.GetString("user_id")
	fileID := c.Param("id")

	var req dto.RenameFileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid request",
			Details: err.Error(),
		})
		return
	}

	if err := h.fileSvc.RenameFile(userID, fileID, req.NewName); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Failed to rename file",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.MessageResponse{
		Message: "File renamed successfully",
	})
}

// DeleteFile deletes a file (moves to recycle bin)
// @Summary Delete file
// @Tags File Management
// @Produce json
// @Param id path string true "File ID"
// @Success 200 {object} dto.MessageResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /files/{id} [delete]
func (h *FileHandler) DeleteFile(c *gin.Context) {
	userID := c.GetString("user_id")
	fileID := c.Param("id")

	if err := h.fileSvc.DeleteFile(userID, fileID); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Failed to delete file",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.MessageResponse{
		Message: "File moved to recycle bin",
	})
}

// Permission Management

// GrantPermission grants permission on a file
// @Summary Grant file permission
// @Tags File Permissions
// @Accept json
// @Produce json
// @Param permission body dto.GrantPermissionRequest true "Grant permission request"
// @Success 200 {object} dto.MessageResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /files/permissions/grant [post]
func (h *FileHandler) GrantPermission(c *gin.Context) {
	userID := c.GetString("user_id")

	var req dto.GrantPermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid request",
			Details: err.Error(),
		})
		return
	}

	// Either UserID or RoleID must be provided
	if req.UserID == nil && req.RoleID == nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Either user_id or role_id must be provided",
		})
		return
	}

	var err error
	if req.UserID != nil {
		err = h.permissionSvc.GrantUserPermission(req.FileID, *req.UserID, userID, req.Permission, req.ExpiresAt)
	} else {
		err = h.permissionSvc.GrantRolePermission(req.FileID, *req.RoleID, userID, req.Permission, req.ExpiresAt)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Failed to grant permission",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.MessageResponse{
		Message: "Permission granted successfully",
	})
}

// RevokePermission revokes permission on a file
// @Summary Revoke file permission
// @Tags File Permissions
// @Accept json
// @Produce json
// @Param permission body dto.RevokePermissionRequest true "Revoke permission request"
// @Success 200 {object} dto.MessageResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /files/permissions/revoke [post]
func (h *FileHandler) RevokePermission(c *gin.Context) {
	userID := c.GetString("user_id")

	var req dto.RevokePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid request",
			Details: err.Error(),
		})
		return
	}

	var err error
	if req.UserID != nil {
		err = h.permissionSvc.RevokeUserPermission(req.FileID, *req.UserID, userID, req.Permission)
	} else if req.RoleID != nil {
		err = h.permissionSvc.RevokeRolePermission(req.FileID, *req.RoleID, userID, req.Permission)
	} else {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Either user_id or role_id must be provided",
		})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Failed to revoke permission",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.MessageResponse{
		Message: "Permission revoked successfully",
	})
}

// GetFilePermissions gets all permissions for a file
// @Summary Get file permissions
// @Tags File Permissions
// @Produce json
// @Param id path string true "File ID"
// @Success 200 {array} dto.FilePermissionResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /files/{id}/permissions [get]
func (h *FileHandler) GetFilePermissions(c *gin.Context) {
	fileID := c.Param("id")

	permissions, err := h.permissionSvc.GetFilePermissions(fileID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Failed to get file permissions",
			Details: err.Error(),
		})
		return
	}

	// Convert to responses
	items := make([]dto.FilePermissionResponse, len(permissions))
	for i, perm := range permissions {
		items[i] = h.permissionToResponse(&perm)
	}

	c.JSON(http.StatusOK, items)
}

// Recycle Bin

// ListRecycleBin lists files in recycle bin
// @Summary List recycle bin
// @Tags Recycle Bin
// @Produce json
// @Param page query int false "Page number"
// @Param page_size query int false "Page size"
// @Success 200 {object} dto.ListResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /files/recycle-bin [get]
func (h *FileHandler) ListRecycleBin(c *gin.Context) {
	userID := c.GetString("user_id")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))

	items, total, err := h.recycleBinSvc.ListRecycleBin(userID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Failed to list recycle bin",
			Details: err.Error(),
		})
		return
	}

	// Convert to responses
	responses := make([]dto.RecycleBinItemResponse, len(items))
	for i, item := range items {
		responses[i] = h.recycleBinItemToResponse(&item)
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	c.JSON(http.StatusOK, dto.ListResponse{
		Items:      responses,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	})
}

// RestoreFile restores a file from recycle bin
// @Summary Restore file from recycle bin
// @Tags Recycle Bin
// @Accept json
// @Produce json
// @Param restore body dto.RestoreFileRequest true "Restore request"
// @Success 200 {object} dto.MessageResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /files/recycle-bin/restore [post]
func (h *FileHandler) RestoreFile(c *gin.Context) {
	userID := c.GetString("user_id")

	var req dto.RestoreFileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid request",
			Details: err.Error(),
		})
		return
	}

	if err := h.recycleBinSvc.RestoreFromRecycleBin(userID, req.FileID); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Failed to restore file",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.MessageResponse{
		Message: "File restored successfully",
	})
}

// PermanentlyDeleteFile permanently deletes a file from recycle bin
// @Summary Permanently delete file
// @Tags Recycle Bin
// @Produce json
// @Param id path string true "File ID"
// @Success 200 {object} dto.MessageResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /files/recycle-bin/{id} [delete]
func (h *FileHandler) PermanentlyDeleteFile(c *gin.Context) {
	userID := c.GetString("user_id")
	fileID := c.Param("id")

	if err := h.recycleBinSvc.PermanentlyDelete(userID, fileID); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Failed to permanently delete file",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.MessageResponse{
		Message: "File permanently deleted",
	})
}

// EmptyRecycleBin empties the recycle bin
// @Summary Empty recycle bin
// @Tags Recycle Bin
// @Produce json
// @Success 200 {object} dto.MessageResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /files/recycle-bin/empty [post]
func (h *FileHandler) EmptyRecycleBin(c *gin.Context) {
	userID := c.GetString("user_id")

	deletedCount, err := h.recycleBinSvc.EmptyRecycleBin(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Failed to empty recycle bin",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.MessageResponse{
		Message: fmt.Sprintf("Permanently deleted %d files", deletedCount),
	})
}

// Helper functions

// fileToResponse converts a file model to response DTO
func (h *FileHandler) fileToResponse(file *models.File, userID string) dto.FileResponse {
	resp := dto.FileResponse{
		ID:            file.ID,
		Name:          file.Name,
		Path:          file.Path,
		ParentID:      file.ParentID,
		OwnerID:       file.OwnerID,
		Size:          file.Size,
		SizeHuman:     formatSize(file.Size),
		MimeType:      file.MimeType,
		Extension:     file.Extension,
		IsDirectory:   file.IsDirectory,
		Version:       file.Version,
		Description:   file.Description,
		DownloadCount: file.DownloadCount,
		ViewCount:     file.ViewCount,
		ShareLink:     file.ShareLink,
		CreatedAt:     file.CreatedAt,
		UpdatedAt:     file.UpdatedAt,
		LastAccessAt:  file.LastAccessAt,
	}

	// Check permissions
	resp.CanRead, _ = h.permissionSvc.CheckPermission(userID, file.ID, services.PermissionRead)
	resp.CanWrite, _ = h.permissionSvc.CheckPermission(userID, file.ID, services.PermissionWrite)
	resp.CanDelete, _ = h.permissionSvc.CheckPermission(userID, file.ID, services.PermissionDelete)
	resp.CanShare, _ = h.permissionSvc.CheckPermission(userID, file.ID, services.PermissionShare)

	return resp
}

// permissionToResponse converts permission model to response DTO
func (h *FileHandler) permissionToResponse(perm *models.FilePermission) dto.FilePermissionResponse {
	resp := dto.FilePermissionResponse{
		ID:         perm.ID,
		FileID:     perm.FileID,
		UserID:     perm.UserID,
		RoleID:     perm.RoleID,
		Permission: perm.Permission,
		GrantedBy:  perm.GrantedBy,
		ExpiresAt:  perm.ExpiresAt,
		CreatedAt:  perm.CreatedAt,
	}

	return resp
}

// recycleBinItemToResponse converts recycle bin item to response DTO
func (h *FileHandler) recycleBinItemToResponse(item *models.RecycleBin) dto.RecycleBinItemResponse {
	now := time.Now()
	daysLeft := int(item.AutoDeleteAt.Sub(now).Hours() / 24)
	if daysLeft < 0 {
		daysLeft = 0
	}

	return dto.RecycleBinItemResponse{
		ID:           item.ID,
		FileID:       item.FileID,
		OriginalPath: item.OriginalPath,
		OriginalName: item.OriginalName,
		DeletedBy:    item.DeletedBy,
		DeletedAt:    item.DeletedAt,
		Size:         item.Size,
		SizeHuman:    formatSize(item.Size),
		AutoDeleteAt: item.AutoDeleteAt,
		DaysLeft:     daysLeft,
	}
}

// formatSize formats file size to human-readable format
func formatSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}
