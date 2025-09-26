package models

import "time"

// NFSVersion NFS版本类型
type NFSVersion string

const (
	NFSVersion3   NFSVersion = "3"
	NFSVersion4   NFSVersion = "4"
	NFSVersion4_1 NFSVersion = "4.1"
	NFSVersion4_2 NFSVersion = "4.2"
)

// NFSSecurityFlavor NFS安全类型
type NFSSecurityFlavor string

const (
	SecNone  NFSSecurityFlavor = "none"
	SecSys   NFSSecurityFlavor = "sys"
	SecKrb5  NFSSecurityFlavor = "krb5"
	SecKrb5i NFSSecurityFlavor = "krb5i"
	SecKrb5p NFSSecurityFlavor = "krb5p"
)

// NFSPermission NFS权限类型
type NFSPermission string

const (
	PermissionReadOnly  NFSPermission = "ro"
	PermissionReadWrite NFSPermission = "rw"
	PermissionNoAccess  NFSPermission = "none"
)

// NFSOperationMode NFS操作模式
type NFSOperationMode string

const (
	ModeSync  NFSOperationMode = "sync"
	ModeAsync NFSOperationMode = "async"
)

// NFSExport NFS导出配置
type NFSExport struct {
	Base
	Path           string            `gorm:"type:varchar(500);not null" json:"path"`
	Name           string            `gorm:"type:varchar(100);uniqueIndex;not null" json:"name"`
	Comment        string            `gorm:"type:text" json:"comment"`
	IsEnabled      bool              `gorm:"default:true" json:"is_enabled"`
	Version        NFSVersion        `gorm:"type:varchar(10);default:'4'" json:"version"`
	SecurityFlavor NFSSecurityFlavor `gorm:"type:varchar(20);default:'sys'" json:"security_flavor"`

	// 基础选项
	Permission     NFSPermission    `gorm:"type:varchar(10);default:'rw'" json:"permission"`
	OperationMode  NFSOperationMode `gorm:"type:varchar(10);default:'sync'" json:"operation_mode"`
	RootSquash     bool             `gorm:"default:true" json:"root_squash"`
	AllSquash      bool             `gorm:"default:false" json:"all_squash"`
	AnonUID        int              `gorm:"default:65534" json:"anon_uid"`
	AnonGID        int              `gorm:"default:65534" json:"anon_gid"`

	// 网络访问控制
	AllowedHosts   string           `gorm:"type:text" json:"allowed_hosts"` // 逗号分隔的主机列表
	DeniedHosts    string           `gorm:"type:text" json:"denied_hosts"`  // 逗号分隔的主机列表

	// 性能选项
	ReadSize       int              `gorm:"default:131072" json:"read_size"`     // 读取块大小
	WriteSize      int              `gorm:"default:131072" json:"write_size"`    // 写入块大小
	SubtreeCheck   bool             `gorm:"default:true" json:"subtree_check"`

	// 版本控制
	EnableVersioning    bool         `gorm:"default:false" json:"enable_versioning"`
	MaxVersions         int          `gorm:"default:10" json:"max_versions"`
	VersionRetention    int          `gorm:"default:30" json:"version_retention"` // 天数

	// 多路径支持
	EnableMultipath     bool         `gorm:"default:false" json:"enable_multipath"`
	MultipathPolicy     string       `gorm:"type:varchar(50);default:'round_robin'" json:"multipath_policy"`

	// 关联关系
	ClientAccess  []NFSClientAccess `gorm:"foreignKey:ExportID" json:"client_access,omitempty"`
	Snapshots     []NFSSnapshot     `gorm:"foreignKey:ExportID" json:"snapshots,omitempty"`
	MultipathConf []NFSMultipathConf `gorm:"foreignKey:ExportID" json:"multipath_conf,omitempty"`
}

// NFSClientAccess NFS客户端访问控制
type NFSClientAccess struct {
	Base
	ExportID       string            `gorm:"type:varchar(36);index;not null" json:"export_id"`
	ClientHost     string            `gorm:"type:varchar(255);not null" json:"client_host"` // IP或主机名或网络段
	Permission     NFSPermission     `gorm:"type:varchar(10);not null" json:"permission"`
	SecurityFlavor NFSSecurityFlavor `gorm:"type:varchar(20);default:'sys'" json:"security_flavor"`
	RootSquash     bool              `gorm:"default:true" json:"root_squash"`
	AllSquash      bool              `gorm:"default:false" json:"all_squash"`
	AnonUID        int               `gorm:"default:65534" json:"anon_uid"`
	AnonGID        int               `gorm:"default:65534" json:"anon_gid"`

	// 关联关系
	Export         NFSExport         `gorm:"foreignKey:ExportID" json:"export,omitempty"`
}

// NFSSnapshot NFS快照（版本控制）
type NFSSnapshot struct {
	Base
	ExportID       string    `gorm:"type:varchar(36);index;not null" json:"export_id"`
	SnapshotName   string    `gorm:"type:varchar(255);not null" json:"snapshot_name"`
	SnapshotPath   string    `gorm:"type:varchar(500);not null" json:"snapshot_path"`
	FileSize       int64     `json:"file_size"`
	CreatedBy      string    `gorm:"type:varchar(100)" json:"created_by"`
	Comment        string    `gorm:"type:text" json:"comment"`
	IsAutoSnapshot bool      `gorm:"default:false" json:"is_auto_snapshot"`
	ExpiresAt      *time.Time `json:"expires_at"`

	// 关联关系
	Export         NFSExport `gorm:"foreignKey:ExportID" json:"export,omitempty"`
}

// NFSMultipathConf NFS多路径配置
type NFSMultipathConf struct {
	Base
	ExportID       string `gorm:"type:varchar(36);index;not null" json:"export_id"`
	PathName       string `gorm:"type:varchar(100);not null" json:"path_name"`
	NetworkPath    string `gorm:"type:varchar(500);not null" json:"network_path"`
	Priority       int    `gorm:"default:1" json:"priority"`
	IsActive       bool   `gorm:"default:true" json:"is_active"`
	Weight         int    `gorm:"default:1" json:"weight"`

	// 路径健康状态
	HealthStatus   string `gorm:"type:varchar(20);default:'unknown'" json:"health_status"` // healthy, degraded, failed, unknown
	LastCheckTime  time.Time `json:"last_check_time"`
	ResponseTime   float64   `json:"response_time"` // ms

	// 关联关系
	Export         NFSExport `gorm:"foreignKey:ExportID" json:"export,omitempty"`
}

// NFSGlobalConfig NFS全局配置
type NFSGlobalConfig struct {
	Base
	Version            NFSVersion `gorm:"type:varchar(10);not null" json:"version"`

	// 服务配置
	PortmapperPort     int        `gorm:"default:111" json:"portmapper_port"`
	NFSPort            int        `gorm:"default:2049" json:"nfs_port"`
	MountdPort         int        `gorm:"default:20048" json:"mountd_port"`
	StatdPort          int        `gorm:"default:662" json:"statd_port"`
	LockdPort          int        `gorm:"default:32803" json:"lockd_port"`

	// 性能配置
	ThreadCount        int        `gorm:"default:8" json:"thread_count"`
	MaxConnections     int        `gorm:"default:1024" json:"max_connections"`
	ReadAhead          int        `gorm:"default:128" json:"read_ahead"` // KB
	WriteBuffer        int        `gorm:"default:128" json:"write_buffer"` // KB

	// 缓存配置
	AttributeTimeout   int        `gorm:"default:60" json:"attribute_timeout"` // seconds
	DirectoryTimeout   int        `gorm:"default:60" json:"directory_timeout"` // seconds

	// 安全配置
	RequireSecurePort  bool       `gorm:"default:false" json:"require_secure_port"`
	EnableTCP          bool       `gorm:"default:true" json:"enable_tcp"`
	EnableUDP          bool       `gorm:"default:false" json:"enable_udp"`

	// 日志配置
	LogLevel           int        `gorm:"default:1" json:"log_level"`
	LogFile            string     `gorm:"type:varchar(255);default:'/var/log/nfs.log'" json:"log_file"`
	EnableDebugLog     bool       `gorm:"default:false" json:"enable_debug_log"`

	// 多路径全局配置
	EnableMultipath    bool       `gorm:"default:false" json:"enable_multipath"`
	MultipathPolicy    string     `gorm:"type:varchar(50);default:'round_robin'" json:"multipath_policy"`
	HealthCheckInterval int       `gorm:"default:30" json:"health_check_interval"` // seconds

	IsActive           bool       `gorm:"default:true" json:"is_active"`
}

// NFSService NFS服务状态
type NFSService struct {
	Base
	ServiceName       string     `gorm:"type:varchar(100);not null" json:"service_name"` // nfs-server, rpc-statd, rpc-idmapd
	ProcessID         int        `json:"process_id"`
	Status            string     `gorm:"type:varchar(20);not null" json:"status"` // running, stopped, error
	StartTime         *time.Time `json:"start_time"`
	LastCheckTime     time.Time  `json:"last_check_time"`
	MemoryUsage       int64      `json:"memory_usage"`    // KB
	CPUUsage          float64    `json:"cpu_usage"`       // percentage
	ConnectionCount   int        `json:"connection_count"`
	ExportCount       int        `json:"export_count"`
}

// NFSConnection 活跃连接
type NFSConnection struct {
	Base
	ClientIP          string    `gorm:"type:varchar(50);not null" json:"client_ip"`
	ClientHostname    string    `gorm:"type:varchar(255)" json:"client_hostname"`
	ExportPath        string    `gorm:"type:varchar(500)" json:"export_path"`
	MountPoint        string    `gorm:"type:varchar(500)" json:"mount_point"`
	Version           NFSVersion `gorm:"type:varchar(10)" json:"version"`
	ConnectedAt       time.Time `json:"connected_at"`
	LastActivity      time.Time `json:"last_activity"`
	BytesRead         int64     `json:"bytes_read"`
	BytesWritten      int64     `json:"bytes_written"`
	OperationsRead    int64     `json:"operations_read"`
	OperationsWrite   int64     `json:"operations_write"`
}

// NFSAuditLog NFS审计日志
type NFSAuditLog struct {
	Base
	ClientIP          string    `gorm:"type:varchar(50);index" json:"client_ip"`
	ClientHostname    string    `gorm:"type:varchar(255)" json:"client_hostname"`
	ExportPath        string    `gorm:"type:varchar(500);index" json:"export_path"`
	Operation         string    `gorm:"type:varchar(50);index" json:"operation"` // mount, umount, read, write, create, delete, rename
	FilePath          string    `gorm:"type:varchar(1000)" json:"file_path"`
	Success           bool      `gorm:"index" json:"success"`
	ErrorMessage      string    `gorm:"type:text" json:"error_message"`
	Timestamp         time.Time `gorm:"index" json:"timestamp"`
	BytesTransferred  int64     `json:"bytes_transferred"`
	ResponseTime      float64   `json:"response_time"` // ms
}

// NFSQuota NFS配额管理
type NFSQuota struct {
	Base
	ExportID          string `gorm:"type:varchar(36);index;not null" json:"export_id"`
	QuotaType         string `gorm:"type:varchar(20);not null" json:"quota_type"` // user, group, export
	TargetID          string `gorm:"type:varchar(100);not null" json:"target_id"` // user_id, group_id, or export_id

	// 配额限制
	HardLimitSize     int64  `gorm:"default:0" json:"hard_limit_size"`  // bytes, 0 = unlimited
	SoftLimitSize     int64  `gorm:"default:0" json:"soft_limit_size"`  // bytes, 0 = unlimited
	HardLimitFiles    int64  `gorm:"default:0" json:"hard_limit_files"` // file count, 0 = unlimited
	SoftLimitFiles    int64  `gorm:"default:0" json:"soft_limit_files"` // file count, 0 = unlimited

	// 当前使用量
	CurrentSize       int64  `json:"current_size"`
	CurrentFiles      int64  `json:"current_files"`

	// 宽限期
	GracePeriod       int    `gorm:"default:7" json:"grace_period"` // days

	IsEnabled         bool   `gorm:"default:true" json:"is_enabled"`

	// 关联关系
	Export            NFSExport `gorm:"foreignKey:ExportID" json:"export,omitempty"`
}