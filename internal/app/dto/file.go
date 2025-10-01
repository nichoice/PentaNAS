package dto

import (
	"time"
)

// File DTOs

// CreateDirectoryRequest represents a request to create a directory
type CreateDirectoryRequest struct {
	ParentPath  string `json:"parent_path" binding:"required"` // 父目录路径
	Name        string `json:"name" binding:"required"`        // 目录名
	Description string `json:"description"`                    // 描述
}

// UploadFileRequest represents file upload metadata (multipart form data)
type UploadFileRequest struct {
	TargetPath  string `form:"target_path" binding:"required"` // 目标路径
	Description string `form:"description"`                    // 描述
	// File is handled separately through multipart.FileHeader
}

// FileResponse represents a file or directory
type FileResponse struct {
	ID            string     `json:"id"`
	Name          string     `json:"name"`
	Path          string     `json:"path"`
	ParentID      *string    `json:"parent_id,omitempty"`
	OwnerID       string     `json:"owner_id"`
	OwnerName     string     `json:"owner_name,omitempty"`
	Size          int64      `json:"size"`
	SizeHuman     string     `json:"size_human"`           // 人类可读的大小
	MimeType      string     `json:"mime_type,omitempty"`
	Extension     string     `json:"extension,omitempty"`
	IsDirectory   bool       `json:"is_directory"`
	Version       int        `json:"version"`
	Tags          []string   `json:"tags,omitempty"`
	Description   string     `json:"description,omitempty"`
	DownloadCount int64      `json:"download_count"`
	ViewCount     int64      `json:"view_count"`
	ShareLink     string     `json:"share_link,omitempty"`
	CanRead       bool       `json:"can_read"`
	CanWrite      bool       `json:"can_write"`
	CanDelete     bool       `json:"can_delete"`
	CanShare      bool       `json:"can_share"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	LastAccessAt  *time.Time `json:"last_access_at,omitempty"`
}

// FileListRequest represents a request to list files
type FileListRequest struct {
	ParentPath  string `form:"parent_path"`              // 父目录路径
	Page        int    `form:"page"`                     // 页码
	PageSize    int    `form:"page_size"`                // 每页大小
	Extension   string `form:"extension"`                // 文件扩展名过滤
	IsDirectory *bool  `form:"is_directory"`             // 是否为目录
	MinSize     int64  `form:"min_size"`                 // 最小文件大小
	MaxSize     int64  `form:"max_size"`                 // 最大文件大小
	SortBy      string `form:"sort_by"`                  // 排序字段: name, size, created_at, updated_at
	SortOrder   string `form:"sort_order"`               // 排序顺序: asc, desc
}

// FileSearchRequest represents a file search request
type FileSearchRequest struct {
	Keyword     string `form:"keyword" binding:"required"` // 搜索关键词
	Page        int    `form:"page"`
	PageSize    int    `form:"page_size"`
	Extension   string `form:"extension"`
	IsDirectory *bool  `form:"is_directory"`
	OwnerID     string `form:"owner_id"` // 管理员可以搜索指定用户的文件
}

// MoveFileRequest represents a request to move a file
type MoveFileRequest struct {
	FileID        string `json:"file_id" binding:"required"`
	NewParentPath string `json:"new_parent_path" binding:"required"`
}

// RenameFileRequest represents a request to rename a file
type RenameFileRequest struct {
	NewName string `json:"new_name" binding:"required"`
}

// BatchOperationRequest represents a batch file operation
type BatchOperationRequest struct {
	FileIDs   []string `json:"file_ids" binding:"required,min=1"`
	Operation string   `json:"operation" binding:"required"` // delete, move, copy
	TargetPath string  `json:"target_path,omitempty"`        // For move/copy operations
}

// BatchOperationResponse represents batch operation result
type BatchOperationResponse struct {
	SuccessCount int      `json:"success_count"`
	FailureCount int      `json:"failure_count"`
	Errors       []string `json:"errors,omitempty"`
}

// Permission DTOs

// GrantPermissionRequest represents a permission grant request
type GrantPermissionRequest struct {
	FileID     string     `json:"file_id" binding:"required"`
	UserID     *string    `json:"user_id"`                                      // 授予用户（二选一）
	RoleID     *string    `json:"role_id"`                                      // 授予角色（二选一）
	Permission string     `json:"permission" binding:"required,oneof=read write delete share admin"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
}

// RevokePermissionRequest represents a permission revoke request
type RevokePermissionRequest struct {
	FileID     string  `json:"file_id" binding:"required"`
	UserID     *string `json:"user_id"`
	RoleID     *string `json:"role_id"`
	Permission string  `json:"permission" binding:"required"`
}

// FilePermissionResponse represents file permission info
type FilePermissionResponse struct {
	ID         string     `json:"id"`
	FileID     string     `json:"file_id"`
	FileName   string     `json:"file_name,omitempty"`
	UserID     *string    `json:"user_id,omitempty"`
	UserName   *string    `json:"user_name,omitempty"`
	RoleID     *string    `json:"role_id,omitempty"`
	RoleName   *string    `json:"role_name,omitempty"`
	Permission string     `json:"permission"`
	GrantedBy  string     `json:"granted_by"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

// Recycle Bin DTOs

// RecycleBinItemResponse represents an item in recycle bin
type RecycleBinItemResponse struct {
	ID           string    `json:"id"`
	FileID       string    `json:"file_id"`
	OriginalPath string    `json:"original_path"`
	OriginalName string    `json:"original_name"`
	DeletedBy    string    `json:"deleted_by"`
	DeletedByName string   `json:"deleted_by_name,omitempty"`
	DeletedAt    time.Time `json:"deleted_at"`
	Size         int64     `json:"size"`
	SizeHuman    string    `json:"size_human"`
	AutoDeleteAt time.Time `json:"auto_delete_at"`
	DaysLeft     int       `json:"days_left"` // 距离自动删除的天数
}

// RestoreFileRequest represents a file restore request
type RestoreFileRequest struct {
	FileID string `json:"file_id" binding:"required"`
}

// Share DTOs

// CreateShareRequest represents a file share creation request
type CreateShareRequest struct {
	FileID       string     `json:"file_id" binding:"required"`
	Password     string     `json:"password,omitempty"`       // 访问密码
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`     // 过期时间
	MaxDownloads int        `json:"max_downloads,omitempty"`  // 最大下载次数
}

// ShareResponse represents a file share link
type ShareResponse struct {
	ID            string     `json:"id"`
	FileID        string     `json:"file_id"`
	FileName      string     `json:"file_name"`
	ShareCode     string     `json:"share_code"`
	ShareURL      string     `json:"share_url"` // 完整的分享链接
	Password      string     `json:"password,omitempty"`
	ExpiresAt     *time.Time `json:"expires_at,omitempty"`
	MaxDownloads  int        `json:"max_downloads,omitempty"`
	DownloadCount int        `json:"download_count"`
	IsActive      bool       `json:"is_active"`
	CreatedBy     string     `json:"created_by"`
	CreatedAt     time.Time  `json:"created_at"`
}

// AccessShareRequest represents access to shared file
type AccessShareRequest struct {
	ShareCode string `json:"share_code" binding:"required"`
	Password  string `json:"password,omitempty"`
}

// Storage Stats DTOs

// StorageStatsResponse represents user storage statistics
type StorageStatsResponse struct {
	UserID           string  `json:"user_id"`
	TotalFiles       int64   `json:"total_files"`
	TotalDirectories int64   `json:"total_directories"`
	TotalSize        int64   `json:"total_size"`
	TotalSizeHuman   string  `json:"total_size_human"`
	DeletedFiles     int64   `json:"deleted_files"`
	DeletedSize      int64   `json:"deleted_size"`
	DeletedSizeHuman string  `json:"deleted_size_human"`
	QuotaLimit       int64   `json:"quota_limit,omitempty"`
	QuotaLimitHuman  string  `json:"quota_limit_human,omitempty"`
	QuotaUsage       float64 `json:"quota_usage"` // 使用百分比
	LastUpdatedAt    time.Time `json:"last_updated_at"`
}

// SetQuotaRequest represents a quota update request
type SetQuotaRequest struct {
	UserID     string `json:"user_id" binding:"required"`
	QuotaLimit int64  `json:"quota_limit" binding:"required,min=0"` // 字节数，0表示无限制
}

// File Activity DTOs

// FileActivityResponse represents a file activity log entry
type FileActivityResponse struct {
	ID        string    `json:"id"`
	FileID    string    `json:"file_id"`
	FileName  string    `json:"file_name,omitempty"`
	UserID    string    `json:"user_id"`
	UserName  string    `json:"user_name,omitempty"`
	Action    string    `json:"action"`
	IPAddress string    `json:"ip_address,omitempty"`
	UserAgent string    `json:"user_agent,omitempty"`
	Details   string    `json:"details,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// FileActivityRequest represents a request to get file activities
type FileActivityRequest struct {
	FileID   string `form:"file_id"`   // 文件ID过滤
	UserID   string `form:"user_id"`   // 用户ID过滤
	Action   string `form:"action"`    // 操作类型过滤
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
}

// Tag DTOs

// CreateTagRequest represents a tag creation request
type CreateTagRequest struct {
	Name        string `json:"name" binding:"required"`
	Color       string `json:"color,omitempty"`
	Description string `json:"description,omitempty"`
}

// TagResponse represents a tag
type TagResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Color       string `json:"color,omitempty"`
	Description string `json:"description,omitempty"`
	UsageCount  int64  `json:"usage_count"`
	CreatedBy   string `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
}

// AddTagsRequest represents a request to add tags to a file
type AddTagsRequest struct {
	FileID string   `json:"file_id" binding:"required"`
	TagIDs []string `json:"tag_ids" binding:"required,min=1"`
}

// Breadcrumb DTOs

// BreadcrumbItem represents a path component for navigation
type BreadcrumbItem struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

// BreadcrumbResponse represents the full path breadcrumb
type BreadcrumbResponse struct {
	Items []BreadcrumbItem `json:"items"`
}

// File Version DTOs

// FileVersionResponse represents a file version
type FileVersionResponse struct {
	ID          string    `json:"id"`
	FileID      string    `json:"file_id"`
	Version     int       `json:"version"`
	Size        int64     `json:"size"`
	SizeHuman   string    `json:"size_human"`
	MD5Hash     string    `json:"md5_hash"`
	CreatedBy   string    `json:"created_by"`
	CreatedByName string  `json:"created_by_name,omitempty"`
	Comment     string    `json:"comment,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// Advanced Search DTOs

// AdvancedSearchRequest represents an advanced search request
type AdvancedSearchRequest struct {
	Keyword       string     `form:"keyword"`
	Extension     string     `form:"extension"`
	MinSize       int64      `form:"min_size"`
	MaxSize       int64      `form:"max_size"`
	CreatedAfter  *time.Time `form:"created_after"`
	CreatedBefore *time.Time `form:"created_before"`
	Tags          []string   `form:"tags"`
	OwnerID       string     `form:"owner_id"`
	IsDirectory   *bool      `form:"is_directory"`
	Page          int        `form:"page"`
	PageSize      int        `form:"page_size"`
}

// Helper response types

// FileTreeNode represents a node in file tree
type FileTreeNode struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Path        string         `json:"path"`
	IsDirectory bool           `json:"is_directory"`
	Size        int64          `json:"size"`
	Children    []FileTreeNode `json:"children,omitempty"`
}

// DiskUsageResponse represents disk usage by file type
type DiskUsageResponse struct {
	Extension string `json:"extension"`
	Count     int64  `json:"count"`
	TotalSize int64  `json:"total_size"`
	SizeHuman string `json:"size_human"`
}

// RecentFilesRequest represents recent files request
type RecentFilesRequest struct {
	Days     int    `form:"days"`      // 最近几天，默认7天
	Action   string `form:"action"`    // 操作类型: upload, download, view
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
}
