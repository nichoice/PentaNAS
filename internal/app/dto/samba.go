package dto

import "time"

// CreateSambaAccountRequest 创建Samba账号请求
type CreateSambaAccountRequest struct {
	UserID      string `json:"user_id" binding:"required"`
	SambaUser   string `json:"samba_user" binding:"required,min=3,max=50,alphanum"`
	Password    string `json:"password" binding:"required,min=6"`
	Role        string `json:"role" binding:"required,oneof=samba_admin samba_user samba_guest samba_backup samba_timemachine"`
	IsEnabled   *bool  `json:"is_enabled,omitempty"`
	Description string `json:"description,omitempty"`
}

// UpdateSambaAccountRequest 更新Samba账号请求
type UpdateSambaAccountRequest struct {
	Password    string `json:"password,omitempty" binding:"omitempty,min=6"`
	Role        string `json:"role,omitempty" binding:"omitempty,oneof=samba_admin samba_user samba_guest samba_backup samba_timemachine"`
	IsEnabled   *bool  `json:"is_enabled,omitempty"`
	Description string `json:"description,omitempty"`
}

// SambaAccountResponse Samba账号响应
type SambaAccountResponse struct {
	ID          string                  `json:"id"`
	UserID      string                  `json:"user_id"`
	SambaUser   string                  `json:"samba_user"`
	Role        string                  `json:"role"`
	IsEnabled   bool                    `json:"is_enabled"`
	Description string                  `json:"description"`
	LastLogin   *time.Time              `json:"last_login"`
	CreatedAt   time.Time               `json:"created_at"`
	UpdatedAt   time.Time               `json:"updated_at"`
	User        *UserResponse           `json:"user,omitempty"`
	ShareAccess []SambaShareAccessResponse `json:"share_access,omitempty"`
}

// CreateSambaShareRequest 创建Samba共享请求
type CreateSambaShareRequest struct {
	Name               string `json:"name" binding:"required,min=1,max=100"`
	Path               string `json:"path" binding:"required"`
	Comment            string `json:"comment,omitempty"`
	IsEnabled          *bool  `json:"is_enabled,omitempty"`
	AllowGuest         *bool  `json:"allow_guest,omitempty"`
	GuestOnly          *bool  `json:"guest_only,omitempty"`
	Browseable         *bool  `json:"browseable,omitempty"`
	Writable           *bool  `json:"writable,omitempty"`
	CreateMask         string `json:"create_mask,omitempty"`
	DirectoryMask      string `json:"directory_mask,omitempty"`
	ForceCreateMode    string `json:"force_create_mode,omitempty"`
	ForceDirectoryMode string `json:"force_directory_mode,omitempty"`

	// 高级功能配置
	EnableTimeMachine  *bool  `json:"enable_time_machine,omitempty"`
	TimeMachineQuota   *int64 `json:"time_machine_quota,omitempty"`
	EnableRecycleBin   *bool  `json:"enable_recycle_bin,omitempty"`
	RecycleBinPath     string `json:"recycle_bin_path,omitempty"`
	EnableMultiChannel *bool  `json:"enable_multi_channel,omitempty"`
}

// UpdateSambaShareRequest 更新Samba共享请求
type UpdateSambaShareRequest struct {
	Comment            string `json:"comment,omitempty"`
	IsEnabled          *bool  `json:"is_enabled,omitempty"`
	AllowGuest         *bool  `json:"allow_guest,omitempty"`
	GuestOnly          *bool  `json:"guest_only,omitempty"`
	Browseable         *bool  `json:"browseable,omitempty"`
	Writable           *bool  `json:"writable,omitempty"`
	CreateMask         string `json:"create_mask,omitempty"`
	DirectoryMask      string `json:"directory_mask,omitempty"`
	ForceCreateMode    string `json:"force_create_mode,omitempty"`
	ForceDirectoryMode string `json:"force_directory_mode,omitempty"`

	// 高级功能配置
	EnableTimeMachine  *bool  `json:"enable_time_machine,omitempty"`
	TimeMachineQuota   *int64 `json:"time_machine_quota,omitempty"`
	EnableRecycleBin   *bool  `json:"enable_recycle_bin,omitempty"`
	RecycleBinPath     string `json:"recycle_bin_path,omitempty"`
	EnableMultiChannel *bool  `json:"enable_multi_channel,omitempty"`
}

// SambaShareResponse Samba共享响应
type SambaShareResponse struct {
	ID                 string                     `json:"id"`
	Name               string                     `json:"name"`
	Path               string                     `json:"path"`
	Comment            string                     `json:"comment"`
	IsEnabled          bool                       `json:"is_enabled"`
	AllowGuest         bool                       `json:"allow_guest"`
	GuestOnly          bool                       `json:"guest_only"`
	Browseable         bool                       `json:"browseable"`
	Writable           bool                       `json:"writable"`
	CreateMask         string                     `json:"create_mask"`
	DirectoryMask      string                     `json:"directory_mask"`
	ForceCreateMode    string                     `json:"force_create_mode"`
	ForceDirectoryMode string                     `json:"force_directory_mode"`

	// 高级功能配置
	EnableTimeMachine  bool                       `json:"enable_time_machine"`
	TimeMachineQuota   int64                      `json:"time_machine_quota"`
	EnableRecycleBin   bool                       `json:"enable_recycle_bin"`
	RecycleBinPath     string                     `json:"recycle_bin_path"`
	EnableMultiChannel bool                       `json:"enable_multi_channel"`

	CreatedAt          time.Time                  `json:"created_at"`
	UpdatedAt          time.Time                  `json:"updated_at"`
	ShareAccess        []SambaShareAccessResponse `json:"share_access,omitempty"`
}

// SetSambaShareAccessRequest 设置共享访问权限请求
type SetSambaShareAccessRequest struct {
	ShareID    string `json:"share_id" binding:"required"`
	AccountID  string `json:"account_id" binding:"required"`
	Permission string `json:"permission" binding:"required,oneof=read write admin"`
}

// SambaShareAccessResponse Samba共享访问权限响应
type SambaShareAccessResponse struct {
	ID         string                `json:"id"`
	ShareID    string                `json:"share_id"`
	AccountID  string                `json:"account_id"`
	Permission string                `json:"permission"`
	CreatedAt  time.Time             `json:"created_at"`
	UpdatedAt  time.Time             `json:"updated_at"`
	Share      *SambaShareResponse   `json:"share,omitempty"`
	Account    *SambaAccountResponse `json:"account,omitempty"`
}

// UpdateSambaGlobalConfigRequest 更新Samba全局配置请求
type UpdateSambaGlobalConfigRequest struct {
	Version              string `json:"version,omitempty"`
	ServerString         string `json:"server_string,omitempty"`
	Workgroup            string `json:"workgroup,omitempty"`
	NetbiosName          string `json:"netbios_name,omitempty"`
	SecurityLevel        string `json:"security_level,omitempty" binding:"omitempty,oneof=share user server domain ads"`
	EncryptPasswords     *bool  `json:"encrypt_passwords,omitempty"`
	PassdbBackend        string `json:"passdb_backend,omitempty"`

	// 网络配置
	Interfaces           string `json:"interfaces,omitempty"`
	BindInterfacesOnly   *bool  `json:"bind_interfaces_only,omitempty"`
	SocketOptions        string `json:"socket_options,omitempty"`

	// 日志配置
	LogLevel             *int   `json:"log_level,omitempty"`
	LogFile              string `json:"log_file,omitempty"`
	MaxLogSize           *int   `json:"max_log_size,omitempty"`

	// 性能配置
	DeadTime             *int   `json:"dead_time,omitempty"`
	GetWDCacheTime       *int   `json:"getwd_cache_time,omitempty"`
	LPQCacheTime         *int   `json:"lpq_cache_time,omitempty"`
	MaxConnections       *int   `json:"max_connections,omitempty"`

	// 多通道配置
	EnableMultiChannel   *bool  `json:"enable_multi_channel,omitempty"`
	MaxChannels          *int   `json:"max_channels,omitempty"`

	// 其他配置
	MapToGuest           string `json:"map_to_guest,omitempty" binding:"omitempty,oneof=Never 'Bad User' 'Bad Password'"`
	GuestAccount         string `json:"guest_account,omitempty"`
	HostsAllow           string `json:"hosts_allow,omitempty"`
	HostsDeny            string `json:"hosts_deny,omitempty"`

	// 审计配置
	EnableAuditing       *bool  `json:"enable_auditing,omitempty"`
	AuditPrefix          string `json:"audit_prefix,omitempty"`
	FullAuditPrefix      string `json:"full_audit_prefix,omitempty"`
}

// SambaGlobalConfigResponse Samba全局配置响应
type SambaGlobalConfigResponse struct {
	ID                   string    `json:"id"`
	Version              string    `json:"version"`
	ServerString         string    `json:"server_string"`
	Workgroup            string    `json:"workgroup"`
	NetbiosName          string    `json:"netbios_name"`
	SecurityLevel        string    `json:"security_level"`
	EncryptPasswords     bool      `json:"encrypt_passwords"`
	PassdbBackend        string    `json:"passdb_backend"`

	// 网络配置
	Interfaces           string    `json:"interfaces"`
	BindInterfacesOnly   bool      `json:"bind_interfaces_only"`
	SocketOptions        string    `json:"socket_options"`

	// 日志配置
	LogLevel             int       `json:"log_level"`
	LogFile              string    `json:"log_file"`
	MaxLogSize           int       `json:"max_log_size"`

	// 性能配置
	DeadTime             int       `json:"dead_time"`
	GetWDCacheTime       int       `json:"getwd_cache_time"`
	LPQCacheTime         int       `json:"lpq_cache_time"`
	MaxConnections       int       `json:"max_connections"`

	// 多通道配置
	EnableMultiChannel   bool      `json:"enable_multi_channel"`
	MaxChannels          int       `json:"max_channels"`

	// 其他配置
	MapToGuest           string    `json:"map_to_guest"`
	GuestAccount         string    `json:"guest_account"`
	HostsAllow           string    `json:"hosts_allow"`
	HostsDeny            string    `json:"hosts_deny"`

	// 审计配置
	EnableAuditing       bool      `json:"enable_auditing"`
	AuditPrefix          string    `json:"audit_prefix"`
	FullAuditPrefix      string    `json:"full_audit_prefix"`

	IsActive             bool      `json:"is_active"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

// SambaServiceResponse Samba服务状态响应
type SambaServiceResponse struct {
	ID              string     `json:"id"`
	ServiceName     string     `json:"service_name"`
	ProcessID       int        `json:"process_id"`
	Status          string     `json:"status"`
	StartTime       *time.Time `json:"start_time"`
	LastCheckTime   time.Time  `json:"last_check_time"`
	MemoryUsage     int64      `json:"memory_usage"`
	CPUUsage        float64    `json:"cpu_usage"`
	ConnectionCount int        `json:"connection_count"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// SambaConnectionResponse 活跃连接响应
type SambaConnectionResponse struct {
	ID             string    `json:"id"`
	ClientIP       string    `json:"client_ip"`
	ClientHostname string    `json:"client_hostname"`
	Username       string    `json:"username"`
	ShareName      string    `json:"share_name"`
	ConnectedAt    time.Time `json:"connected_at"`
	LastActivity   time.Time `json:"last_activity"`
	FilesOpen      int       `json:"files_open"`
	BytesRead      int64     `json:"bytes_read"`
	BytesWritten   int64     `json:"bytes_written"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// SambaAuditLogResponse Samba审计日志响应
type SambaAuditLogResponse struct {
	ID               string    `json:"id"`
	Username         string    `json:"username"`
	ClientIP         string    `json:"client_ip"`
	ShareName        string    `json:"share_name"`
	Operation        string    `json:"operation"`
	FilePath         string    `json:"file_path"`
	Success          bool      `json:"success"`
	ErrorMessage     string    `json:"error_message"`
	Timestamp        time.Time `json:"timestamp"`
	BytesTransferred int64     `json:"bytes_transferred"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// RestoreRecycleBinItemRequest 恢复回收站条目请求
type RestoreRecycleBinItemRequest struct {
	ItemID      string `json:"item_id" binding:"required"`
	RestorePath string `json:"restore_path,omitempty"` // 可选，不提供则恢复到原路径
}

// RecycleBinItemResponse 回收站条目响应
type RecycleBinItemResponse struct {
	ID           string               `json:"id"`
	ShareID      string               `json:"share_id"`
	OriginalPath string               `json:"original_path"`
	RecyclePath  string               `json:"recycle_path"`
	DeletedBy    string               `json:"deleted_by"`
	DeletedAt    time.Time            `json:"deleted_at"`
	FileSize     int64                `json:"file_size"`
	FileType     string               `json:"file_type"`
	IsDirectory  bool                 `json:"is_directory"`
	ExpiresAt    *time.Time           `json:"expires_at"`
	CreatedAt    time.Time            `json:"created_at"`
	UpdatedAt    time.Time            `json:"updated_at"`
	Share        *SambaShareResponse  `json:"share,omitempty"`
}

// SambaStatusResponse Samba整体状态响应
type SambaStatusResponse struct {
	IsRunning         bool                    `json:"is_running"`
	Version           string                  `json:"version"`
	Uptime            string                  `json:"uptime"`
	TotalShares       int                     `json:"total_shares"`
	ActiveShares      int                     `json:"active_shares"`
	TotalAccounts     int                     `json:"total_accounts"`
	ActiveAccounts    int                     `json:"active_accounts"`
	ActiveConnections int                     `json:"active_connections"`
	Services          []SambaServiceResponse  `json:"services"`
	RecentConnections []SambaConnectionResponse `json:"recent_connections"`
}

// SambaConfigGenerateResponse 生成配置文件响应
type SambaConfigGenerateResponse struct {
	ConfigContent string    `json:"config_content"`
	FilePath      string    `json:"file_path"`
	GeneratedAt   time.Time `json:"generated_at"`
	ChecksumMD5   string    `json:"checksum_md5"`
}