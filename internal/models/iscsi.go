package models

import "time"

// ISCSITargetStatus Target状态枚举
type ISCSITargetStatus string

const (
	ISCSIStatusActive   ISCSITargetStatus = "active"
	ISCSIStatusInactive ISCSITargetStatus = "inactive"
	ISCSIStatusError    ISCSITargetStatus = "error"
)

// ISCSIPermission 权限级别
type ISCSIPermission string

const (
	ISCSIPermissionReadOnly  ISCSIPermission = "ro"
	ISCSIPermissionReadWrite ISCSIPermission = "rw"
	ISCSIPermissionDeny      ISCSIPermission = "deny"
)

// ISCSIAuthType 认证类型
type ISCSIAuthType string

const (
	ISCSIAuthNone ISCSIAuthType = "none"
	ISCSIAuthCHAP ISCSIAuthType = "chap"
)

// ISCSIDeviceType 设备类型
type ISCSIDeviceType string

const (
	ISCSIDeviceBlock ISCSIDeviceType = "block"
	ISCSIDeviceFile  ISCSIDeviceType = "file"
	ISCSIDeviceTCMU  ISCSIDeviceType = "tcmu"
)

// ISCSITarget 描述一个 iSCSI Target 的持久化配置，包含运行状态、别名及与 LUN/ACL 的关联。
type ISCSITarget struct {
	Base
	Name           string            `gorm:"type:varchar(223);uniqueIndex;not null" json:"name"` // IQN format
	Alias          string            `gorm:"type:varchar(100)" json:"alias"`
	Status         ISCSITargetStatus `gorm:"type:varchar(20);default:'inactive'" json:"status"`
	Comment        string            `gorm:"type:text" json:"comment"`
	IsEnabled      bool              `gorm:"default:true" json:"is_enabled"`

	// 关联关系
	LUNs           []ISCSILUN        `gorm:"foreignKey:TargetID" json:"luns,omitempty"`
	ACLs           []ISCSIACL        `gorm:"foreignKey:TargetID" json:"acls,omitempty"`
	Sessions       []ISCSISession    `gorm:"foreignKey:TargetID" json:"sessions,omitempty"`
}

// ISCSILUN 表示 LUN(Logical Unit Number) 的元数据与存储路径信息，对应单个可映射卷。
type ISCSILUN struct {
	Base
	TargetID       string          `gorm:"type:varchar(36);index;not null" json:"target_id"`
	LUN            int             `gorm:"not null" json:"lun"` // LUN编号，0-255
	Name           string          `gorm:"type:varchar(100);not null" json:"name"`
	DeviceType     ISCSIDeviceType `gorm:"type:varchar(20);not null" json:"device_type"`
	DevicePath     string          `gorm:"type:varchar(500);not null" json:"device_path"`
	Size           int64           `json:"size"` // bytes
	BlockSize      int             `gorm:"default:512" json:"block_size"`
	IsEnabled      bool            `gorm:"default:true" json:"is_enabled"`

	// 性能选项
	ReadOnly       bool            `gorm:"default:false" json:"read_only"`
	WriteThrough   bool            `gorm:"default:false" json:"write_through"`
	WriteBack      bool            `gorm:"default:true" json:"write_back"`

	// TCMU特定选项
	TCMUHandler    string          `gorm:"type:varchar(100)" json:"tcmu_handler"`
	TCMUOptions    string          `gorm:"type:text" json:"tcmu_options"`

	Comment        string          `gorm:"type:text" json:"comment"`

	// 关联关系
	Target         ISCSITarget     `gorm:"foreignKey:TargetID" json:"target,omitempty"`
}

// ISCSIACL 维护 Initiator 与 Target 间的访问控制策略，可定义权限及认证方式。
type ISCSIACL struct {
	Base
	TargetID       string          `gorm:"type:varchar(36);index;not null" json:"target_id"`
	InitiatorName  string          `gorm:"type:varchar(223);not null" json:"initiator_name"` // IQN format
	Permission     ISCSIPermission `gorm:"type:varchar(10);default:'rw'" json:"permission"`
	IsEnabled      bool            `gorm:"default:true" json:"is_enabled"`

	// 认证配置
	AuthType       ISCSIAuthType   `gorm:"type:varchar(20);default:'none'" json:"auth_type"`
	Username       string          `gorm:"type:varchar(100)" json:"username"`
	Password       string          `gorm:"type:varchar(255)" json:"-"` // 不在JSON中显示
	MutualAuth     bool            `gorm:"default:false" json:"mutual_auth"`
	MutualUsername string          `gorm:"type:varchar(100)" json:"mutual_username"`
	MutualPassword string          `gorm:"type:varchar(255)" json:"-"` // 不在JSON中显示

	Comment        string          `gorm:"type:text" json:"comment"`

	// LUN映射权限
	LUNMappings    []ISCSILUNMapping `gorm:"foreignKey:ACLID" json:"lun_mappings,omitempty"`

	// 关联关系
	Target         ISCSITarget     `gorm:"foreignKey:TargetID" json:"target,omitempty"`
	Sessions       []ISCSISession  `gorm:"foreignKey:ACLID" json:"sessions,omitempty"`
}

// ISCSILUNMapping 记录 ACL 与 LUN 的绑定关系及针对性的权限覆盖。
type ISCSILUNMapping struct {
	Base
	ACLID          string          `gorm:"type:varchar(36);index;not null" json:"acl_id"`
	LUNID          string          `gorm:"type:varchar(36);index;not null" json:"lun_id"`
	Permission     ISCSIPermission `gorm:"type:varchar(10);default:'rw'" json:"permission"`
	IsEnabled      bool            `gorm:"default:true" json:"is_enabled"`

	// 关联关系
	ACL            ISCSIACL        `gorm:"foreignKey:ACLID" json:"acl,omitempty"`
	LUN            ISCSILUN        `gorm:"foreignKey:LUNID" json:"lun,omitempty"`
}

// ISCSIGlobalConfig 保存 iSCSI 服务级别的全局参数，例如端口、会话阈值与日志策略。
type ISCSIGlobalConfig struct {
	Base

	// 服务配置
	TargetPort     int             `gorm:"default:3260" json:"target_port"`
	MaxSessions    int             `gorm:"default:256" json:"max_sessions"`
	MaxConnections int             `gorm:"default:1" json:"max_connections"`

	// 性能配置
	MaxRecvDataSegmentLength int `gorm:"default:8192" json:"max_recv_data_segment_length"`
	MaxXmitDataSegmentLength int `gorm:"default:8192" json:"max_xmit_data_segment_length"`
	MaxBurstLength           int `gorm:"default:262144" json:"max_burst_length"`
	FirstBurstLength         int `gorm:"default:65536" json:"first_burst_length"`
	MaxOutstandingR2T        int `gorm:"default:1" json:"max_outstanding_r2t"`

	// 超时配置
	DefaultTime2Wait    int         `gorm:"default:2" json:"default_time2_wait"`
	DefaultTime2Retain  int         `gorm:"default:20" json:"default_time2_retain"`
	LoginTimeout        int         `gorm:"default:30" json:"login_timeout"`
	LogoutTimeout       int         `gorm:"default:30" json:"logout_timeout"`

	// 认证配置
	RequireAuth         bool        `gorm:"default:false" json:"require_auth"`
	AllowDuplicateSessions bool     `gorm:"default:false" json:"allow_duplicate_sessions"`

	// 日志配置
	LogLevel            int         `gorm:"default:1" json:"log_level"`
	LogFile             string      `gorm:"type:varchar(255);default:'/var/log/iscsi/iscsi.log'" json:"log_file"`
	EnableDebugLog      bool        `gorm:"default:false" json:"enable_debug_log"`

	IsActive            bool        `gorm:"default:true" json:"is_active"`
}

// ISCSIService 追踪 iSCSI 相关后台服务（如 target、tcmu-runner）的实时运行状态。
type ISCSIService struct {
	Base
	ServiceName       string     `gorm:"type:varchar(100);not null" json:"service_name"` // target, tcmu-runner
	ProcessID         int        `json:"process_id"`
	Status            string     `gorm:"type:varchar(20);not null" json:"status"` // running, stopped, error
	StartTime         *time.Time `json:"start_time"`
	LastCheckTime     time.Time  `json:"last_check_time"`
	MemoryUsage       int64      `json:"memory_usage"`    // KB
	CPUUsage          float64    `json:"cpu_usage"`       // percentage
	ConnectionCount   int        `json:"connection_count"`
	SessionCount      int        `json:"session_count"`
}

// ISCSISession 表示 Initiator 与 Target 间建立的长连接会话及其统计信息。
type ISCSISession struct {
	Base
	TargetID          string    `gorm:"type:varchar(36);index;not null" json:"target_id"`
	ACLID             string    `gorm:"type:varchar(36);index" json:"acl_id"`
	InitiatorName     string    `gorm:"type:varchar(223);not null" json:"initiator_name"`
	InitiatorIP       string    `gorm:"type:varchar(50);not null" json:"initiator_ip"`
	TargetSessionID   string    `gorm:"type:varchar(100);not null" json:"target_session_id"`
	InitiatorSessionID string   `gorm:"type:varchar(100);not null" json:"initiator_session_id"`

	// 会话状态
	Status            string    `gorm:"type:varchar(20);not null" json:"status"` // active, closing, closed
	ConnectedAt       time.Time `json:"connected_at"`
	LastActivity      time.Time `json:"last_activity"`

	// 统计信息
	BytesRead         int64     `json:"bytes_read"`
	BytesWritten      int64     `json:"bytes_written"`
	CommandsCompleted int64     `json:"commands_completed"`
	IOErrors          int64     `json:"io_errors"`

	// 关联关系
	Target            ISCSITarget `gorm:"foreignKey:TargetID" json:"target,omitempty"`
	ACL               ISCSIACL    `gorm:"foreignKey:ACLID" json:"acl,omitempty"`
	Connections       []ISCSIConnection `gorm:"foreignKey:SessionID" json:"connections,omitempty"`
}

// ISCSIConnection 描述会话下具体 TCP 连接的状态与读写指标。
type ISCSIConnection struct {
	Base
	SessionID         string    `gorm:"type:varchar(36);index;not null" json:"session_id"`
	ConnectionID      int       `gorm:"not null" json:"connection_id"`
	InitiatorIP       string    `gorm:"type:varchar(50);not null" json:"initiator_ip"`
	InitiatorPort     int       `json:"initiator_port"`
	TargetIP          string    `gorm:"type:varchar(50);not null" json:"target_ip"`
	TargetPort        int       `json:"target_port"`

	// 连接状态
	Status            string    `gorm:"type:varchar(20);not null" json:"status"` // active, closing, closed
	ConnectedAt       time.Time `json:"connected_at"`
	LastActivity      time.Time `json:"last_activity"`

	// 统计信息
	BytesRead         int64     `json:"bytes_read"`
	BytesWritten      int64     `json:"bytes_written"`
	IOPs              int64     `json:"iops"`

	// 关联关系
	Session           ISCSISession `gorm:"foreignKey:SessionID" json:"session,omitempty"`
}

// ISCSIAuditLog 记录 iSCSI 操作行为，用于审计与排障。
type ISCSIAuditLog struct {
	Base
	TargetName        string    `gorm:"type:varchar(223);index" json:"target_name"`
	InitiatorName     string    `gorm:"type:varchar(223);index" json:"initiator_name"`
	InitiatorIP       string    `gorm:"type:varchar(50);index" json:"initiator_ip"`
	Operation         string    `gorm:"type:varchar(50);index" json:"operation"` // login, logout, read, write, create_target, delete_target
	LUN               int       `json:"lun"`
	Success           bool      `gorm:"index" json:"success"`
	ErrorMessage      string    `gorm:"type:text" json:"error_message"`
	Timestamp         time.Time `gorm:"index" json:"timestamp"`
	BytesTransferred  int64     `json:"bytes_transferred"`
	ResponseTime      float64   `json:"response_time"` // ms
}

// ISCSIStoragePool 表示可供 LUN 使用的后端存储池及容量状况。
type ISCSIStoragePool struct {
	Base
	Name              string    `gorm:"type:varchar(100);uniqueIndex;not null" json:"name"`
	Type              string    `gorm:"type:varchar(50);not null" json:"type"` // file, block, lvm, zfs
	Path              string    `gorm:"type:varchar(500);not null" json:"path"`
	Size              int64     `json:"size"`       // bytes
	Used              int64     `json:"used"`       // bytes
	Available         int64     `json:"available"`  // bytes
	IsEnabled         bool      `gorm:"default:true" json:"is_enabled"`
	Comment           string    `gorm:"type:text" json:"comment"`

	// 性能配置
	BlockSize         int       `gorm:"default:512" json:"block_size"`
	AllocationUnit    int64     `gorm:"default:1048576" json:"allocation_unit"` // bytes, 1MB default

	// 关联关系
	LUNs              []ISCSILUN `gorm:"foreignKey:DevicePath;references:Path" json:"luns,omitempty"`
}

// ISCSIPerformanceStats 聚合 iSCSI 系统的历史性能指标，便于分析趋势。
type ISCSIPerformanceStats struct {
	Base
	TargetID          string    `gorm:"type:varchar(36);index;not null" json:"target_id"`
	LUNID             string    `gorm:"type:varchar(36);index" json:"lun_id"`

	// 时间戳
	Timestamp         time.Time `gorm:"index" json:"timestamp"`

	// IO统计
	ReadIOPs          int64     `json:"read_iops"`
	WriteIOPs         int64     `json:"write_iops"`
	ReadBandwidth     int64     `json:"read_bandwidth"`    // bytes/sec
	WriteBandwidth    int64     `json:"write_bandwidth"`   // bytes/sec
	ReadLatency       float64   `json:"read_latency"`      // ms
	WriteLatency      float64   `json:"write_latency"`     // ms

	// 队列统计
	QueueDepth        int       `json:"queue_depth"`
	QueueTime         float64   `json:"queue_time"`        // ms

	// 错误统计
	ReadErrors        int64     `json:"read_errors"`
	WriteErrors       int64     `json:"write_errors"`
	TimeoutErrors     int64     `json:"timeout_errors"`

	// 关联关系
	Target            ISCSITarget `gorm:"foreignKey:TargetID" json:"target,omitempty"`
	LUN               ISCSILUN    `gorm:"foreignKey:LUNID" json:"lun,omitempty"`
}
