package models

import "time"

// SambaVersion Samba版本类型
type SambaVersion string

const (
	SambaVersion3_6 SambaVersion = "3.6"
	SambaVersion4_0 SambaVersion = "4.0"
	SambaVersion4_2 SambaVersion = "4.2"
	SambaVersion4_4 SambaVersion = "4.4"
	SambaVersion4_6 SambaVersion = "4.6"
	SambaVersion4_8 SambaVersion = "4.8"
	SambaVersion4_9 SambaVersion = "4.9"
	SambaVersion4_10 SambaVersion = "4.10"
	SambaVersion4_11 SambaVersion = "4.11"
	SambaVersion4_12 SambaVersion = "4.12"
	SambaVersion4_13 SambaVersion = "4.13"
	SambaVersion4_14 SambaVersion = "4.14"
	SambaVersion4_15 SambaVersion = "4.15"
	SambaVersion4_16 SambaVersion = "4.16"
	SambaVersion4_17 SambaVersion = "4.17"
	SambaVersion4_18 SambaVersion = "4.18"
	SambaVersion4_19 SambaVersion = "4.19"
	SambaVersion4_20 SambaVersion = "4.20"
)

// SambaAccountRole Samba账号角色
type SambaAccountRole string

const (
	RoleSambaAdmin   SambaAccountRole = "samba_admin"   // Samba管理员
	RoleSambaUser    SambaAccountRole = "samba_user"    // 普通用户
	RoleSambaGuest   SambaAccountRole = "samba_guest"   // 访客用户
	RoleSambaBackup  SambaAccountRole = "samba_backup"  // 备份用户
	RoleSambaTimeMachine SambaAccountRole = "samba_timemachine" // 时间机器用户
)

// SambaAccount Samba账号管理
type SambaAccount struct {
	Base
	UserID      string           `gorm:"type:varchar(36);index;not null" json:"user_id"`
	SambaUser   string           `gorm:"type:varchar(100);uniqueIndex;not null" json:"samba_user"`
	Password    string           `gorm:"type:varchar(255);not null" json:"-"`
	Role        SambaAccountRole `gorm:"type:varchar(50);not null" json:"role"`
	IsEnabled   bool             `gorm:"default:true" json:"is_enabled"`
	Description string           `gorm:"type:text" json:"description"`
	LastLogin   *time.Time       `json:"last_login"`

	// 关联关系
	User        User             `gorm:"foreignKey:UserID" json:"user,omitempty"`
	ShareAccess []SambaShareAccess `gorm:"foreignKey:AccountID" json:"share_access,omitempty"`
}

// SambaShare Samba共享配置
type SambaShare struct {
	Base
	Name               string          `gorm:"type:varchar(100);uniqueIndex;not null" json:"name"`
	Path               string          `gorm:"type:varchar(500);not null" json:"path"`
	Comment            string          `gorm:"type:text" json:"comment"`
	IsEnabled          bool            `gorm:"default:true" json:"is_enabled"`
	AllowGuest         bool            `gorm:"default:false" json:"allow_guest"`
	GuestOnly          bool            `gorm:"default:false" json:"guest_only"`
	Browseable         bool            `gorm:"default:true" json:"browseable"`
	Writable           bool            `gorm:"default:false" json:"writable"`
	CreateMask         string          `gorm:"type:varchar(10);default:'0664'" json:"create_mask"`
	DirectoryMask      string          `gorm:"type:varchar(10);default:'0775'" json:"directory_mask"`
	ForceCreateMode    string          `gorm:"type:varchar(10)" json:"force_create_mode"`
	ForceDirectoryMode string          `gorm:"type:varchar(10)" json:"force_directory_mode"`

	// 高级功能配置
	EnableTimeMachine  bool            `gorm:"default:false" json:"enable_time_machine"`
	TimeMachineQuota   int64           `gorm:"default:0" json:"time_machine_quota"` // MB
	EnableRecycleBin   bool            `gorm:"default:false" json:"enable_recycle_bin"`
	RecycleBinPath     string          `gorm:"type:varchar(500)" json:"recycle_bin_path"`
	EnableMultiChannel bool            `gorm:"default:false" json:"enable_multi_channel"`

	// 关联关系
	ShareAccess []SambaShareAccess `gorm:"foreignKey:ShareID" json:"share_access,omitempty"`
}

// SambaShareAccess Samba共享访问权限
type SambaShareAccess struct {
	Base
	ShareID     string           `gorm:"type:varchar(36);index;not null" json:"share_id"`
	AccountID   string           `gorm:"type:varchar(36);index;not null" json:"account_id"`
	Permission  string           `gorm:"type:varchar(20);not null" json:"permission"` // read, write, admin

	// 关联关系
	Share       SambaShare       `gorm:"foreignKey:ShareID" json:"share,omitempty"`
	Account     SambaAccount     `gorm:"foreignKey:AccountID" json:"account,omitempty"`
}

// SambaGlobalConfig Samba全局配置
type SambaGlobalConfig struct {
	Base
	Version              SambaVersion `gorm:"type:varchar(10);not null" json:"version"`
	ServerString         string       `gorm:"type:varchar(255);default:'Samba Server'" json:"server_string"`
	Workgroup            string       `gorm:"type:varchar(50);default:'WORKGROUP'" json:"workgroup"`
	NetbiosName          string       `gorm:"type:varchar(50)" json:"netbios_name"`
	SecurityLevel        string       `gorm:"type:varchar(20);default:'user'" json:"security_level"` // share, user, server, domain, ads
	EncryptPasswords     bool         `gorm:"default:true" json:"encrypt_passwords"`
	PassdbBackend        string       `gorm:"type:varchar(100);default:'tdbsam'" json:"passdb_backend"`

	// 网络配置
	Interfaces           string       `gorm:"type:text" json:"interfaces"`
	BindInterfacesOnly   bool         `gorm:"default:false" json:"bind_interfaces_only"`
	SocketOptions        string       `gorm:"type:text" json:"socket_options"`

	// 日志配置
	LogLevel             int          `gorm:"default:1" json:"log_level"`
	LogFile              string       `gorm:"type:varchar(255);default:'/var/log/samba/samba.log'" json:"log_file"`
	MaxLogSize           int          `gorm:"default:5000" json:"max_log_size"` // KB

	// 性能配置
	DeadTime             int          `gorm:"default:15" json:"dead_time"` // minutes
	GetWDCacheTime       int          `gorm:"default:10" json:"getwd_cache_time"` // seconds
	LPQCacheTime         int          `gorm:"default:10" json:"lpq_cache_time"` // seconds
	MaxConnections       int          `gorm:"default:0" json:"max_connections"` // 0 = unlimited

	// 多通道配置
	EnableMultiChannel   bool         `gorm:"default:false" json:"enable_multi_channel"`
	MaxChannels          int          `gorm:"default:4" json:"max_channels"`

	// 其他配置
	MapToGuest           string       `gorm:"type:varchar(50);default:'Never'" json:"map_to_guest"` // Never, Bad User, Bad Password
	GuestAccount         string       `gorm:"type:varchar(50);default:'nobody'" json:"guest_account"`
	HostsAllow           string       `gorm:"type:text" json:"hosts_allow"`
	HostsDeny            string       `gorm:"type:text" json:"hosts_deny"`

	// 审计配置
	EnableAuditing       bool         `gorm:"default:false" json:"enable_auditing"`
	AuditPrefix          string       `gorm:"type:varchar(100)" json:"audit_prefix"`
	FullAuditPrefix      string       `gorm:"type:varchar(100)" json:"full_audit_prefix"`

	IsActive             bool         `gorm:"default:true" json:"is_active"`
}

// SambaService Samba服务状态
type SambaService struct {
	Base
	ServiceName    string     `gorm:"type:varchar(100);not null" json:"service_name"` // smbd, nmbd, winbindd
	ProcessID      int        `json:"process_id"`
	Status         string     `gorm:"type:varchar(20);not null" json:"status"` // running, stopped, error
	StartTime      *time.Time `json:"start_time"`
	LastCheckTime  time.Time  `json:"last_check_time"`
	MemoryUsage    int64      `json:"memory_usage"` // KB
	CPUUsage       float64    `json:"cpu_usage"`    // percentage
	ConnectionCount int       `json:"connection_count"`
}

// SambaConnection 活跃连接
type SambaConnection struct {
	Base
	ClientIP       string    `gorm:"type:varchar(50);not null" json:"client_ip"`
	ClientHostname string    `gorm:"type:varchar(255)" json:"client_hostname"`
	Username       string    `gorm:"type:varchar(100)" json:"username"`
	ShareName      string    `gorm:"type:varchar(100)" json:"share_name"`
	ConnectedAt    time.Time `json:"connected_at"`
	LastActivity   time.Time `json:"last_activity"`
	FilesOpen      int       `json:"files_open"`
	BytesRead      int64     `json:"bytes_read"`
	BytesWritten   int64     `json:"bytes_written"`
}

// SambaAuditLog Samba审计日志
type SambaAuditLog struct {
	Base
	Username      string    `gorm:"type:varchar(100);index" json:"username"`
	ClientIP      string    `gorm:"type:varchar(50);index" json:"client_ip"`
	ShareName     string    `gorm:"type:varchar(100);index" json:"share_name"`
	Operation     string    `gorm:"type:varchar(50);index" json:"operation"` // connect, disconnect, open, close, read, write, delete, rename
	FilePath      string    `gorm:"type:varchar(1000)" json:"file_path"`
	Success       bool      `gorm:"index" json:"success"`
	ErrorMessage  string    `gorm:"type:text" json:"error_message"`
	Timestamp     time.Time `gorm:"index" json:"timestamp"`
	BytesTransferred int64  `json:"bytes_transferred"`
}

// RecycleBinItem 回收站条目
type RecycleBinItem struct {
	Base
	ShareID        string    `gorm:"type:varchar(36);index;not null" json:"share_id"`
	OriginalPath   string    `gorm:"type:varchar(1000);not null" json:"original_path"`
	RecyclePath    string    `gorm:"type:varchar(1000);not null" json:"recycle_path"`
	DeletedBy      string    `gorm:"type:varchar(100)" json:"deleted_by"`
	DeletedAt      time.Time `json:"deleted_at"`
	FileSize       int64     `json:"file_size"`
	FileType       string    `gorm:"type:varchar(50)" json:"file_type"`
	IsDirectory    bool      `json:"is_directory"`
	ExpiresAt      *time.Time `json:"expires_at"`

	// 关联关系
	Share          SambaShare `gorm:"foreignKey:ShareID" json:"share,omitempty"`
}