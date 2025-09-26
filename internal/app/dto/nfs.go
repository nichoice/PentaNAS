package dto

import (
	"time"
	"pnas/internal/models"
)

// CreateNFSExportRequest 创建NFS导出请求
type CreateNFSExportRequest struct {
	Path            string                      `json:"path" binding:"required"`
	Name            string                      `json:"name" binding:"required"`
	Comment         string                      `json:"comment"`
	Version         models.NFSVersion           `json:"version"`
	SecurityFlavor  models.NFSSecurityFlavor    `json:"security_flavor"`
	Permission      models.NFSPermission        `json:"permission"`
	OperationMode   models.NFSOperationMode     `json:"operation_mode"`
	RootSquash      bool                        `json:"root_squash"`
	AllSquash       bool                        `json:"all_squash"`
	AnonUID         int                         `json:"anon_uid"`
	AnonGID         int                         `json:"anon_gid"`
	AllowedHosts    string                      `json:"allowed_hosts"`
	DeniedHosts     string                      `json:"denied_hosts"`
	ReadSize        int                         `json:"read_size"`
	WriteSize       int                         `json:"write_size"`
	SubtreeCheck    bool                        `json:"subtree_check"`
	EnableVersioning bool                       `json:"enable_versioning"`
	MaxVersions     int                         `json:"max_versions"`
	VersionRetention int                        `json:"version_retention"`
	EnableMultipath bool                        `json:"enable_multipath"`
	MultipathPolicy string                      `json:"multipath_policy"`
}

// UpdateNFSExportRequest 更新NFS导出请求
type UpdateNFSExportRequest struct {
	Comment         *string                     `json:"comment"`
	IsEnabled       *bool                       `json:"is_enabled"`
	Version         *models.NFSVersion          `json:"version"`
	SecurityFlavor  *models.NFSSecurityFlavor   `json:"security_flavor"`
	Permission      *models.NFSPermission       `json:"permission"`
	OperationMode   *models.NFSOperationMode    `json:"operation_mode"`
	RootSquash      *bool                       `json:"root_squash"`
	AllSquash       *bool                       `json:"all_squash"`
	AnonUID         *int                        `json:"anon_uid"`
	AnonGID         *int                        `json:"anon_gid"`
	AllowedHosts    *string                     `json:"allowed_hosts"`
	DeniedHosts     *string                     `json:"denied_hosts"`
	ReadSize        *int                        `json:"read_size"`
	WriteSize       *int                        `json:"write_size"`
	SubtreeCheck    *bool                       `json:"subtree_check"`
	EnableVersioning *bool                      `json:"enable_versioning"`
	MaxVersions     *int                        `json:"max_versions"`
	VersionRetention *int                       `json:"version_retention"`
	EnableMultipath *bool                       `json:"enable_multipath"`
	MultipathPolicy *string                     `json:"multipath_policy"`
}

// NFSExportResponse NFS导出响应
type NFSExportResponse struct {
	ID               string                      `json:"id"`
	Path             string                      `json:"path"`
	Name             string                      `json:"name"`
	Comment          string                      `json:"comment"`
	IsEnabled        bool                        `json:"is_enabled"`
	Version          models.NFSVersion           `json:"version"`
	SecurityFlavor   models.NFSSecurityFlavor    `json:"security_flavor"`
	Permission       models.NFSPermission        `json:"permission"`
	OperationMode    models.NFSOperationMode     `json:"operation_mode"`
	RootSquash       bool                        `json:"root_squash"`
	AllSquash        bool                        `json:"all_squash"`
	AnonUID          int                         `json:"anon_uid"`
	AnonGID          int                         `json:"anon_gid"`
	AllowedHosts     string                      `json:"allowed_hosts"`
	DeniedHosts      string                      `json:"denied_hosts"`
	ReadSize         int                         `json:"read_size"`
	WriteSize        int                         `json:"write_size"`
	SubtreeCheck     bool                        `json:"subtree_check"`
	EnableVersioning bool                        `json:"enable_versioning"`
	MaxVersions      int                         `json:"max_versions"`
	VersionRetention int                         `json:"version_retention"`
	EnableMultipath  bool                        `json:"enable_multipath"`
	MultipathPolicy  string                      `json:"multipath_policy"`
	CreatedAt        time.Time                   `json:"created_at"`
	UpdatedAt        time.Time                   `json:"updated_at"`
	ClientAccess     []NFSClientAccessResponse   `json:"client_access,omitempty"`
	Snapshots        []NFSSnapshotResponse       `json:"snapshots,omitempty"`
	MultipathConf    []NFSMultipathConfResponse  `json:"multipath_conf,omitempty"`
}

// CreateNFSClientAccessRequest 创建客户端访问控制请求
type CreateNFSClientAccessRequest struct {
	ExportID       string                      `json:"export_id" binding:"required"`
	ClientHost     string                      `json:"client_host" binding:"required"`
	Permission     models.NFSPermission        `json:"permission" binding:"required"`
	SecurityFlavor models.NFSSecurityFlavor    `json:"security_flavor"`
	RootSquash     bool                        `json:"root_squash"`
	AllSquash      bool                        `json:"all_squash"`
	AnonUID        int                         `json:"anon_uid"`
	AnonGID        int                         `json:"anon_gid"`
}

// NFSClientAccessResponse 客户端访问控制响应
type NFSClientAccessResponse struct {
	ID             string                      `json:"id"`
	ExportID       string                      `json:"export_id"`
	ClientHost     string                      `json:"client_host"`
	Permission     models.NFSPermission        `json:"permission"`
	SecurityFlavor models.NFSSecurityFlavor    `json:"security_flavor"`
	RootSquash     bool                        `json:"root_squash"`
	AllSquash      bool                        `json:"all_squash"`
	AnonUID        int                         `json:"anon_uid"`
	AnonGID        int                         `json:"anon_gid"`
	CreatedAt      time.Time                   `json:"created_at"`
	UpdatedAt      time.Time                   `json:"updated_at"`
}

// CreateNFSSnapshotRequest 创建快照请求
type CreateNFSSnapshotRequest struct {
	ExportID     string `json:"export_id" binding:"required"`
	SnapshotName string `json:"snapshot_name" binding:"required"`
	Comment      string `json:"comment"`
}

// NFSSnapshotResponse 快照响应
type NFSSnapshotResponse struct {
	ID             string     `json:"id"`
	ExportID       string     `json:"export_id"`
	SnapshotName   string     `json:"snapshot_name"`
	SnapshotPath   string     `json:"snapshot_path"`
	FileSize       int64      `json:"file_size"`
	CreatedBy      string     `json:"created_by"`
	Comment        string     `json:"comment"`
	IsAutoSnapshot bool       `json:"is_auto_snapshot"`
	ExpiresAt      *time.Time `json:"expires_at"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// CreateNFSMultipathConfRequest 创建多路径配置请求
type CreateNFSMultipathConfRequest struct {
	ExportID    string `json:"export_id" binding:"required"`
	PathName    string `json:"path_name" binding:"required"`
	NetworkPath string `json:"network_path" binding:"required"`
	Priority    int    `json:"priority"`
	Weight      int    `json:"weight"`
}

// NFSMultipathConfResponse 多路径配置响应
type NFSMultipathConfResponse struct {
	ID            string    `json:"id"`
	ExportID      string    `json:"export_id"`
	PathName      string    `json:"path_name"`
	NetworkPath   string    `json:"network_path"`
	Priority      int       `json:"priority"`
	IsActive      bool      `json:"is_active"`
	Weight        int       `json:"weight"`
	HealthStatus  string    `json:"health_status"`
	LastCheckTime time.Time `json:"last_check_time"`
	ResponseTime  float64   `json:"response_time"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// UpdateNFSGlobalConfigRequest 更新全局配置请求
type UpdateNFSGlobalConfigRequest struct {
	Version               *models.NFSVersion `json:"version"`
	PortmapperPort        *int               `json:"portmapper_port"`
	NFSPort               *int               `json:"nfs_port"`
	MountdPort            *int               `json:"mountd_port"`
	StatdPort             *int               `json:"statd_port"`
	LockdPort             *int               `json:"lockd_port"`
	ThreadCount           *int               `json:"thread_count"`
	MaxConnections        *int               `json:"max_connections"`
	ReadAhead             *int               `json:"read_ahead"`
	WriteBuffer           *int               `json:"write_buffer"`
	AttributeTimeout      *int               `json:"attribute_timeout"`
	DirectoryTimeout      *int               `json:"directory_timeout"`
	RequireSecurePort     *bool              `json:"require_secure_port"`
	EnableTCP             *bool              `json:"enable_tcp"`
	EnableUDP             *bool              `json:"enable_udp"`
	LogLevel              *int               `json:"log_level"`
	LogFile               *string            `json:"log_file"`
	EnableDebugLog        *bool              `json:"enable_debug_log"`
	EnableMultipath       *bool              `json:"enable_multipath"`
	MultipathPolicy       *string            `json:"multipath_policy"`
	HealthCheckInterval   *int               `json:"health_check_interval"`
}

// NFSGlobalConfigResponse 全局配置响应
type NFSGlobalConfigResponse struct {
	ID                  string            `json:"id"`
	Version             models.NFSVersion `json:"version"`
	PortmapperPort      int               `json:"portmapper_port"`
	NFSPort             int               `json:"nfs_port"`
	MountdPort          int               `json:"mountd_port"`
	StatdPort           int               `json:"statd_port"`
	LockdPort           int               `json:"lockd_port"`
	ThreadCount         int               `json:"thread_count"`
	MaxConnections      int               `json:"max_connections"`
	ReadAhead           int               `json:"read_ahead"`
	WriteBuffer         int               `json:"write_buffer"`
	AttributeTimeout    int               `json:"attribute_timeout"`
	DirectoryTimeout    int               `json:"directory_timeout"`
	RequireSecurePort   bool              `json:"require_secure_port"`
	EnableTCP           bool              `json:"enable_tcp"`
	EnableUDP           bool              `json:"enable_udp"`
	LogLevel            int               `json:"log_level"`
	LogFile             string            `json:"log_file"`
	EnableDebugLog      bool              `json:"enable_debug_log"`
	EnableMultipath     bool              `json:"enable_multipath"`
	MultipathPolicy     string            `json:"multipath_policy"`
	HealthCheckInterval int               `json:"health_check_interval"`
	IsActive            bool              `json:"is_active"`
	CreatedAt           time.Time         `json:"created_at"`
	UpdatedAt           time.Time         `json:"updated_at"`
}

// NFSServiceResponse NFS服务状态响应
type NFSServiceResponse struct {
	ID              string     `json:"id"`
	ServiceName     string     `json:"service_name"`
	ProcessID       int        `json:"process_id"`
	Status          string     `json:"status"`
	StartTime       *time.Time `json:"start_time"`
	LastCheckTime   time.Time  `json:"last_check_time"`
	MemoryUsage     int64      `json:"memory_usage"`
	CPUUsage        float64    `json:"cpu_usage"`
	ConnectionCount int        `json:"connection_count"`
	ExportCount     int        `json:"export_count"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// NFSConnectionResponse NFS连接响应
type NFSConnectionResponse struct {
	ID               string            `json:"id"`
	ClientIP         string            `json:"client_ip"`
	ClientHostname   string            `json:"client_hostname"`
	ExportPath       string            `json:"export_path"`
	MountPoint       string            `json:"mount_point"`
	Version          models.NFSVersion `json:"version"`
	ConnectedAt      time.Time         `json:"connected_at"`
	LastActivity     time.Time         `json:"last_activity"`
	BytesRead        int64             `json:"bytes_read"`
	BytesWritten     int64             `json:"bytes_written"`
	OperationsRead   int64             `json:"operations_read"`
	OperationsWrite  int64             `json:"operations_write"`
	CreatedAt        time.Time         `json:"created_at"`
	UpdatedAt        time.Time         `json:"updated_at"`
}

// CreateNFSQuotaRequest 创建配额请求
type CreateNFSQuotaRequest struct {
	ExportID       string `json:"export_id" binding:"required"`
	QuotaType      string `json:"quota_type" binding:"required"`
	TargetID       string `json:"target_id" binding:"required"`
	HardLimitSize  int64  `json:"hard_limit_size"`
	SoftLimitSize  int64  `json:"soft_limit_size"`
	HardLimitFiles int64  `json:"hard_limit_files"`
	SoftLimitFiles int64  `json:"soft_limit_files"`
	GracePeriod    int    `json:"grace_period"`
}

// NFSQuotaResponse 配额响应
type NFSQuotaResponse struct {
	ID             string    `json:"id"`
	ExportID       string    `json:"export_id"`
	QuotaType      string    `json:"quota_type"`
	TargetID       string    `json:"target_id"`
	HardLimitSize  int64     `json:"hard_limit_size"`
	SoftLimitSize  int64     `json:"soft_limit_size"`
	HardLimitFiles int64     `json:"hard_limit_files"`
	SoftLimitFiles int64     `json:"soft_limit_files"`
	CurrentSize    int64     `json:"current_size"`
	CurrentFiles   int64     `json:"current_files"`
	GracePeriod    int       `json:"grace_period"`
	IsEnabled      bool      `json:"is_enabled"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// NFSAuditLogResponse 审计日志响应
type NFSAuditLogResponse struct {
	ID               string    `json:"id"`
	ClientIP         string    `json:"client_ip"`
	ClientHostname   string    `json:"client_hostname"`
	ExportPath       string    `json:"export_path"`
	Operation        string    `json:"operation"`
	FilePath         string    `json:"file_path"`
	Success          bool      `json:"success"`
	ErrorMessage     string    `json:"error_message"`
	Timestamp        time.Time `json:"timestamp"`
	BytesTransferred int64     `json:"bytes_transferred"`
	ResponseTime     float64   `json:"response_time"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// NFSOperationRequest 异步操作请求
type NFSOperationRequest struct {
	OperationType string      `json:"operation_type" binding:"required"` // export_reload, service_restart, snapshot_create
	ExportID      string      `json:"export_id,omitempty"`
	Parameters    interface{} `json:"parameters,omitempty"`
	Priority      int         `json:"priority"` // 1-10, 1 = highest
}

// NFSOperationResponse 异步操作响应
type NFSOperationResponse struct {
	OperationID   string      `json:"operation_id"`
	Status        string      `json:"status"` // pending, running, completed, failed
	Progress      int         `json:"progress"` // 0-100
	Message       string      `json:"message"`
	Result        interface{} `json:"result,omitempty"`
	StartTime     time.Time   `json:"start_time"`
	EndTime       *time.Time  `json:"end_time,omitempty"`
}

// NFSHealthCheckResponse 健康检查响应
type NFSHealthCheckResponse struct {
	Overall         string                      `json:"overall"` // healthy, degraded, critical
	Services        []NFSServiceResponse        `json:"services"`
	Exports         []NFSExportHealthResponse   `json:"exports"`
	MultipathHealth []NFSMultipathHealthResponse `json:"multipath_health"`
	CheckTime       time.Time                   `json:"check_time"`
}

// NFSExportHealthResponse 导出健康状态
type NFSExportHealthResponse struct {
	ExportID      string  `json:"export_id"`
	ExportName    string  `json:"export_name"`
	Status        string  `json:"status"` // healthy, warning, critical
	Message       string  `json:"message"`
	Accessibility bool    `json:"accessibility"`
	ResponseTime  float64 `json:"response_time"`
	ActivePaths   int     `json:"active_paths"`
	TotalPaths    int     `json:"total_paths"`
}

// NFSMultipathHealthResponse 多路径健康状态
type NFSMultipathHealthResponse struct {
	ExportID      string  `json:"export_id"`
	PathName      string  `json:"path_name"`
	Status        string  `json:"status"` // healthy, degraded, failed
	ResponseTime  float64 `json:"response_time"`
	ActivePaths   int     `json:"active_paths"`
	TotalPaths    int     `json:"total_paths"`
}