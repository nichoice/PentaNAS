package models

import (
	"time"
)

// File represents a file or directory in the file system
type File struct {
	Base
	Name          string     `gorm:"index:idx_parent_name,priority:2;not null" json:"name"`                 // 文件名
	Path          string     `gorm:"uniqueIndex;not null" json:"path"`                                      // 完整路径 /wuzhou/user1/docs/file.txt
	ParentID      *string    `gorm:"index:idx_parent_name,priority:1;index:idx_parent_id" json:"parent_id"` // 父目录ID
	OwnerID       string     `gorm:"index:idx_owner;not null" json:"owner_id"`                              // 所有者用户ID
	Size          int64      `gorm:"index:idx_size" json:"size"`                                            // 文件大小(字节)
	MimeType      string     `json:"mime_type"`                                                             // MIME类型
	Extension     string     `gorm:"index:idx_extension" json:"extension"`                                  // 文件扩展名
	IsDirectory   bool       `gorm:"index:idx_is_directory;not null" json:"is_directory"`                   // 是否为目录
	IsDeleted     bool       `gorm:"index:idx_is_deleted;default:false" json:"is_deleted"`                  // 是否已删除（回收站）
	DeletedAt     *time.Time `gorm:"index:idx_deleted_at" json:"deleted_at,omitempty"`                      // 删除时间
	DeletedBy     *string    `json:"deleted_by,omitempty"`                                                  // 删除操作的用户ID
	MD5Hash       string     `gorm:"index:idx_md5" json:"md5_hash,omitempty"`                               // MD5哈希（用于去重）
	SHA256Hash    string     `json:"sha256_hash,omitempty"`                                                 // SHA256哈希
	StoragePath   string     `json:"storage_path,omitempty"`                                                // 实际存储路径（可能与逻辑路径不同）
	Version       int        `gorm:"default:1" json:"version"`                                              // 版本号
	Tags          string     `json:"tags,omitempty"`                                                        // 标签（JSON数组）
	Description   string     `json:"description,omitempty"`                                                 // 描述
	ShareLink     string     `gorm:"uniqueIndex:idx_share_link" json:"share_link,omitempty"`                // 分享链接
	ShareExpiry   *time.Time `json:"share_expiry,omitempty"`                                                // 分享过期时间
	DownloadCount int64      `gorm:"default:0" json:"download_count"`                                       // 下载次数
	ViewCount     int64      `gorm:"default:0" json:"view_count"`                                           // 查看次数
	LastAccessAt  *time.Time `json:"last_access_at,omitempty"`                                              // 最后访问时间

	// 关联
	Owner       User             `gorm:"foreignKey:OwnerID" json:"-"`
	Permissions []FilePermission `gorm:"foreignKey:FileID" json:"-"`
	Versions    []FileVersion    `gorm:"foreignKey:FileID" json:"-"`
}

// TableName specifies the table name for File
func (File) TableName() string {
	return "files"
}

// FilePermission represents file access permissions
type FilePermission struct {
	Base
	FileID     string     `gorm:"index:idx_file_user,priority:1;not null" json:"file_id"`
	UserID     *string    `gorm:"index:idx_file_user,priority:2;index:idx_user_id" json:"user_id,omitempty"` // 用户ID（用户权限）
	RoleID     *string    `gorm:"index:idx_role_id" json:"role_id,omitempty"`                                // 角色ID（角色权限）
	Permission string     `gorm:"not null" json:"permission"`                                                // read, write, delete, share
	GrantedBy  string     `json:"granted_by"`                                                                // 授权者用户ID
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`                                                      // 权限过期时间

	// 关联
	File File  `gorm:"foreignKey:FileID" json:"-"`
	User *User `gorm:"foreignKey:UserID" json:"-"`
	Role *Role `gorm:"foreignKey:RoleID" json:"-"`
}

// TableName specifies the table name for FilePermission
func (FilePermission) TableName() string {
	return "file_permissions"
}

// FileVersion represents a file version (for version control)
type FileVersion struct {
	Base
	FileID      string `gorm:"index:idx_file_version,priority:1;not null" json:"file_id"`
	Version     int    `gorm:"index:idx_file_version,priority:2;not null" json:"version"`
	Size        int64  `json:"size"`
	MD5Hash     string `json:"md5_hash"`
	StoragePath string `json:"storage_path"`
	CreatedBy   string `json:"created_by"`
	Comment     string `json:"comment,omitempty"`

	// 关联
	File File `gorm:"foreignKey:FileID" json:"-"`
}

// TableName specifies the table name for FileVersion
func (FileVersion) TableName() string {
	return "file_versions"
}

// FileActivity represents file activity log (for audit and analytics)
type FileActivity struct {
	Base
	FileID    string `gorm:"index:idx_file_activity;not null" json:"file_id"`
	UserID    string `gorm:"index:idx_user_activity;not null" json:"user_id"`
	Action    string `gorm:"index:idx_action;not null" json:"action"` // upload, download, delete, restore, share, view
	IPAddress string `json:"ip_address,omitempty"`
	UserAgent string `json:"user_agent,omitempty"`
	Details   string `json:"details,omitempty"` // JSON格式的详细信息

	// 关联
	File File `gorm:"foreignKey:FileID" json:"-"`
	User User `gorm:"foreignKey:UserID" json:"-"`
}

// TableName specifies the table name for FileActivity
func (FileActivity) TableName() string {
	return "file_activities"
}

// FileShare represents a file share link
type FileShare struct {
	Base
	FileID        string     `gorm:"index:idx_file_share;not null" json:"file_id"`
	ShareCode     string     `gorm:"uniqueIndex;not null" json:"share_code"` // 分享码
	Password      string     `json:"password,omitempty"`                     // 访问密码（加密存储）
	ExpiresAt     *time.Time `gorm:"index:idx_expires_at" json:"expires_at,omitempty"`
	MaxDownloads  int        `json:"max_downloads,omitempty"`         // 最大下载次数
	DownloadCount int        `gorm:"default:0" json:"download_count"` // 已下载次数
	CreatedBy     string     `json:"created_by"`
	IsActive      bool       `gorm:"default:true" json:"is_active"`

	// 关联
	File File `gorm:"foreignKey:FileID" json:"-"`
}

// TableName specifies the table name for FileShare
func (FileShare) TableName() string {
	return "file_shares"
}

// FileTag represents file tags for categorization
type FileTag struct {
	Base
	Name        string `gorm:"uniqueIndex;not null" json:"name"`
	Color       string `json:"color,omitempty"` // 标签颜色
	Description string `json:"description,omitempty"`
	UsageCount  int64  `gorm:"default:0" json:"usage_count"` // 使用次数
	CreatedBy   string `json:"created_by"`
}

// TableName specifies the table name for FileTag
func (FileTag) TableName() string {
	return "file_tags"
}

// FileTagMapping represents the many-to-many relationship between files and tags
type FileTagMapping struct {
	FileID string `gorm:"primaryKey;index:idx_file_tag,priority:1" json:"file_id"`
	TagID  string `gorm:"primaryKey;index:idx_file_tag,priority:2" json:"tag_id"`

	File File    `gorm:"foreignKey:FileID" json:"-"`
	Tag  FileTag `gorm:"foreignKey:TagID" json:"-"`
}

// TableName specifies the table name for FileTagMapping
func (FileTagMapping) TableName() string {
	return "file_tag_mappings"
}

// RecycleBin represents deleted files metadata for quick recovery
type RecycleBin struct {
	Base
	FileID       string    `gorm:"uniqueIndex;not null" json:"file_id"`
	OriginalPath string    `json:"original_path"`
	OriginalName string    `json:"original_name"`
	DeletedBy    string    `json:"deleted_by"`
	DeletedAt    time.Time `gorm:"index:idx_recycle_bin_deleted_at" json:"deleted_at"`
	Size         int64     `json:"size"`
	AutoDeleteAt time.Time `gorm:"index:idx_auto_delete" json:"auto_delete_at"` // 自动清理时间（30天后）

	// 关联
	File File `gorm:"foreignKey:FileID" json:"-"`
}

// TableName specifies the table name for RecycleBin
func (RecycleBin) TableName() string {
	return "recycle_bin"
}

// FileIndex represents file metadata index for fast search (denormalized)
type FileIndex struct {
	Base
	FileID      string    `gorm:"uniqueIndex;not null" json:"file_id"`
	Name        string    `gorm:"index:idx_name_search" json:"name"`      // 文件名（用于搜索）
	NameLower   string    `gorm:"index:idx_name_lower" json:"name_lower"` // 小写文件名（用于不区分大小写搜索）
	Extension   string    `gorm:"index:idx_ext" json:"extension"`
	Path        string    `gorm:"index:idx_path" json:"path"`
	OwnerID     string    `gorm:"index:idx_owner_search" json:"owner_id"`
	Size        int64     `gorm:"index:idx_size_range" json:"size"`
	IsDirectory bool      `gorm:"index:idx_type" json:"is_directory"`
	IsDeleted   bool      `gorm:"index:idx_deleted_search" json:"is_deleted"`
	Tags        string    `json:"tags"` // JSON数组
	UpdatedAt   time.Time `gorm:"index:idx_updated_search" json:"updated_at"`

	// 全文搜索字段（可选，如果使用PostgreSQL的全文搜索）
	SearchVector string `gorm:"type:tsvector" json:"-"`
}

// TableName specifies the table name for FileIndex
func (FileIndex) TableName() string {
	return "file_index"
}

// StorageStats represents storage usage statistics
type StorageStats struct {
	Base
	UserID           string    `gorm:"uniqueIndex;not null" json:"user_id"`
	TotalFiles       int64     `gorm:"default:0" json:"total_files"`
	TotalDirectories int64     `gorm:"default:0" json:"total_directories"`
	TotalSize        int64     `gorm:"default:0" json:"total_size"`
	DeletedFiles     int64     `gorm:"default:0" json:"deleted_files"`
	DeletedSize      int64     `gorm:"default:0" json:"deleted_size"`
	QuotaLimit       int64     `json:"quota_limit,omitempty"` // 配额限制（字节）
	LastUpdatedAt    time.Time `json:"last_updated_at"`

	// 关联
	User User `gorm:"foreignKey:UserID" json:"-"`
}

// TableName specifies the table name for StorageStats
func (StorageStats) TableName() string {
	return "storage_stats"
}
