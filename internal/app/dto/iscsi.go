package dto

import "time"

// iSCSI Target创建请求
type CreateiSCSITargetRequest struct {
	Name      string `json:"name" binding:"required" validate:"iqn" example:"iqn.2024-01.com.example:target01"`
	Alias     string `json:"alias" example:"Target 01"`
	Comment   string `json:"comment" example:"iSCSI target for backup storage"`
	IsEnabled bool   `json:"is_enabled" example:"true"`
}

// iSCSI Target更新请求
type UpdateiSCSITargetRequest struct {
	Alias     string `json:"alias" example:"Target 01 Updated"`
	Comment   string `json:"comment" example:"Updated comment"`
	IsEnabled *bool  `json:"is_enabled" example:"false"`
}

// iSCSI Target响应
type iSCSITargetResponse struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Alias     string            `json:"alias"`
	Status    string            `json:"status"`
	Comment   string            `json:"comment"`
	IsEnabled bool              `json:"is_enabled"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
	LUNCount  int               `json:"lun_count"`
	ACLCount  int               `json:"acl_count"`
}

// iSCSI LUN创建请求
type CreateiSCSILUNRequest struct {
	TargetID      string `json:"target_id" binding:"required"`
	LUN           int    `json:"lun" binding:"required,min=0,max=255" example:"0"`
	Name          string `json:"name" binding:"required" example:"lun0"`
	DeviceType    string `json:"device_type" binding:"required,oneof=block file tcmu" example:"file"`
	DevicePath    string `json:"device_path" binding:"required" example:"/data/iscsi/lun0.img"`
	Size          int64  `json:"size" binding:"required,min=1" example:"10737418240"` // 10GB
	BlockSize     int    `json:"block_size" example:"512"`
	ReadOnly      bool   `json:"read_only" example:"false"`
	WriteThrough  bool   `json:"write_through" example:"false"`
	WriteBack     bool   `json:"write_back" example:"true"`
	TCMUHandler   string `json:"tcmu_handler" example:"file"`
	TCMUOptions   string `json:"tcmu_options" example:"file=/data/iscsi/tcmu_lun0.img"`
	Comment       string `json:"comment" example:"LUN for database storage"`
}

// iSCSI LUN更新请求
type UpdateiSCSILUNRequest struct {
	Name         string `json:"name" example:"lun0_updated"`
	Size         *int64 `json:"size" example:"21474836480"` // 20GB
	ReadOnly     *bool  `json:"read_only" example:"false"`
	WriteThrough *bool  `json:"write_through" example:"false"`
	WriteBack    *bool  `json:"write_back" example:"true"`
	IsEnabled    *bool  `json:"is_enabled" example:"true"`
	TCMUOptions  string `json:"tcmu_options"`
	Comment      string `json:"comment"`
}

// iSCSI LUN响应
type iSCSILUNResponse struct {
	ID           string    `json:"id"`
	TargetID     string    `json:"target_id"`
	LUN          int       `json:"lun"`
	Name         string    `json:"name"`
	DeviceType   string    `json:"device_type"`
	DevicePath   string    `json:"device_path"`
	Size         int64     `json:"size"`
	BlockSize    int       `json:"block_size"`
	ReadOnly     bool      `json:"read_only"`
	WriteThrough bool      `json:"write_through"`
	WriteBack    bool      `json:"write_back"`
	IsEnabled    bool      `json:"is_enabled"`
	TCMUHandler  string    `json:"tcmu_handler,omitempty"`
	TCMUOptions  string    `json:"tcmu_options,omitempty"`
	Comment      string    `json:"comment"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// iSCSI ACL创建请求
type CreateiSCSIACLRequest struct {
	TargetID       string `json:"target_id" binding:"required"`
	InitiatorName  string `json:"initiator_name" binding:"required" validate:"iqn" example:"iqn.2024-01.com.client:initiator01"`
	Permission     string `json:"permission" binding:"required,oneof=ro rw deny" example:"rw"`
	AuthType       string `json:"auth_type" binding:"oneof=none chap" example:"chap"`
	Username       string `json:"username" example:"iscsi_user"`
	Password       string `json:"password" example:"secure_password"`
	MutualAuth     bool   `json:"mutual_auth" example:"false"`
	MutualUsername string `json:"mutual_username" example:"mutual_user"`
	MutualPassword string `json:"mutual_password" example:"mutual_password"`
	Comment        string `json:"comment" example:"ACL for client server"`
}

// iSCSI ACL更新请求
type UpdateiSCSIACLRequest struct {
	Permission     string `json:"permission" binding:"omitempty,oneof=ro rw deny" example:"ro"`
	AuthType       string `json:"auth_type" binding:"omitempty,oneof=none chap" example:"none"`
	Username       string `json:"username" example:"updated_user"`
	Password       string `json:"password" example:"updated_password"`
	MutualAuth     *bool  `json:"mutual_auth" example:"true"`
	MutualUsername string `json:"mutual_username" example:"updated_mutual_user"`
	MutualPassword string `json:"mutual_password" example:"updated_mutual_password"`
	IsEnabled      *bool  `json:"is_enabled" example:"true"`
	Comment        string `json:"comment" example:"Updated ACL comment"`
}

// iSCSI ACL响应
type iSCSIACLResponse struct {
	ID             string                    `json:"id"`
	TargetID       string                    `json:"target_id"`
	InitiatorName  string                    `json:"initiator_name"`
	Permission     string                    `json:"permission"`
	AuthType       string                    `json:"auth_type"`
	Username       string                    `json:"username,omitempty"`
	MutualAuth     bool                      `json:"mutual_auth"`
	MutualUsername string                    `json:"mutual_username,omitempty"`
	IsEnabled      bool                      `json:"is_enabled"`
	Comment        string                    `json:"comment"`
	CreatedAt      time.Time                 `json:"created_at"`
	UpdatedAt      time.Time                 `json:"updated_at"`
	LUNMappings    []iSCSILUNMappingResponse `json:"lun_mappings,omitempty"`
}

// iSCSI LUN映射创建请求
type CreateiSCSILUNMappingRequest struct {
	ACLID      string `json:"acl_id" binding:"required"`
	LUNID      string `json:"lun_id" binding:"required"`
	Permission string `json:"permission" binding:"required,oneof=ro rw deny" example:"rw"`
}

// iSCSI LUN映射更新请求
type UpdateiSCSILUNMappingRequest struct {
	Permission string `json:"permission" binding:"omitempty,oneof=ro rw deny" example:"ro"`
	IsEnabled  *bool  `json:"is_enabled" example:"true"`
}

// iSCSI LUN映射响应
type iSCSILUNMappingResponse struct {
	ID         string    `json:"id"`
	ACLID      string    `json:"acl_id"`
	LUNID      string    `json:"lun_id"`
	Permission string    `json:"permission"`
	IsEnabled  bool      `json:"is_enabled"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// iSCSI全局配置更新请求
type UpdateiSCSIGlobalConfigRequest struct {
	TargetPort                   *int  `json:"target_port" example:"3260"`
	MaxSessions                  *int  `json:"max_sessions" example:"512"`
	MaxConnections               *int  `json:"max_connections" example:"2"`
	MaxRecvDataSegmentLength     *int  `json:"max_recv_data_segment_length" example:"16384"`
	MaxXmitDataSegmentLength     *int  `json:"max_xmit_data_segment_length" example:"16384"`
	MaxBurstLength               *int  `json:"max_burst_length" example:"524288"`
	FirstBurstLength             *int  `json:"first_burst_length" example:"131072"`
	MaxOutstandingR2T            *int  `json:"max_outstanding_r2t" example:"2"`
	DefaultTime2Wait             *int  `json:"default_time2_wait" example:"2"`
	DefaultTime2Retain           *int  `json:"default_time2_retain" example:"20"`
	LoginTimeout                 *int  `json:"login_timeout" example:"60"`
	LogoutTimeout                *int  `json:"logout_timeout" example:"60"`
	RequireAuth                  *bool `json:"require_auth" example:"true"`
	AllowDuplicateSessions       *bool `json:"allow_duplicate_sessions" example:"false"`
	LogLevel                     *int  `json:"log_level" example:"2"`
	LogFile                      string `json:"log_file" example:"/var/log/iscsi/target.log"`
	EnableDebugLog               *bool `json:"enable_debug_log" example:"true"`
}

// iSCSI全局配置响应
type iSCSIGlobalConfigResponse struct {
	ID                           string    `json:"id"`
	TargetPort                   int       `json:"target_port"`
	MaxSessions                  int       `json:"max_sessions"`
	MaxConnections               int       `json:"max_connections"`
	MaxRecvDataSegmentLength     int       `json:"max_recv_data_segment_length"`
	MaxXmitDataSegmentLength     int       `json:"max_xmit_data_segment_length"`
	MaxBurstLength               int       `json:"max_burst_length"`
	FirstBurstLength             int       `json:"first_burst_length"`
	MaxOutstandingR2T            int       `json:"max_outstanding_r2t"`
	DefaultTime2Wait             int       `json:"default_time2_wait"`
	DefaultTime2Retain           int       `json:"default_time2_retain"`
	LoginTimeout                 int       `json:"login_timeout"`
	LogoutTimeout                int       `json:"logout_timeout"`
	RequireAuth                  bool      `json:"require_auth"`
	AllowDuplicateSessions       bool      `json:"allow_duplicate_sessions"`
	LogLevel                     int       `json:"log_level"`
	LogFile                      string    `json:"log_file"`
	EnableDebugLog               bool      `json:"enable_debug_log"`
	IsActive                     bool      `json:"is_active"`
	CreatedAt                    time.Time `json:"created_at"`
	UpdatedAt                    time.Time `json:"updated_at"`
}

// iSCSI存储池创建请求
type CreateiSCSIStoragePoolRequest struct {
	Name           string `json:"name" binding:"required" example:"pool01"`
	Type           string `json:"type" binding:"required,oneof=file block lvm zfs" example:"file"`
	Path           string `json:"path" binding:"required" example:"/data/iscsi/pool01"`
	Size           int64  `json:"size" binding:"required,min=1" example:"107374182400"` // 100GB
	BlockSize      int    `json:"block_size" example:"4096"`
	AllocationUnit int64  `json:"allocation_unit" example:"1048576"` // 1MB
	Comment        string `json:"comment" example:"Storage pool for iSCSI LUNs"`
}

// iSCSI存储池更新请求
type UpdateiSCSIStoragePoolRequest struct {
	Size           *int64 `json:"size" example:"214748364800"` // 200GB
	BlockSize      *int   `json:"block_size" example:"4096"`
	AllocationUnit *int64 `json:"allocation_unit" example:"2097152"` // 2MB
	IsEnabled      *bool  `json:"is_enabled" example:"true"`
	Comment        string `json:"comment" example:"Updated storage pool"`
}

// iSCSI存储池响应
type iSCSIStoragePoolResponse struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Type           string    `json:"type"`
	Path           string    `json:"path"`
	Size           int64     `json:"size"`
	Used           int64     `json:"used"`
	Available      int64     `json:"available"`
	UsagePercent   float64   `json:"usage_percent"`
	BlockSize      int       `json:"block_size"`
	AllocationUnit int64     `json:"allocation_unit"`
	IsEnabled      bool      `json:"is_enabled"`
	Comment        string    `json:"comment"`
	LUNCount       int       `json:"lun_count"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// iSCSI会话响应
type iSCSISessionResponse struct {
	ID                 string    `json:"id"`
	TargetID           string    `json:"target_id"`
	TargetName         string    `json:"target_name"`
	ACLID              string    `json:"acl_id,omitempty"`
	InitiatorName      string    `json:"initiator_name"`
	InitiatorIP        string    `json:"initiator_ip"`
	TargetSessionID    string    `json:"target_session_id"`
	InitiatorSessionID string    `json:"initiator_session_id"`
	Status             string    `json:"status"`
	ConnectedAt        time.Time `json:"connected_at"`
	LastActivity       time.Time `json:"last_activity"`
	Duration           string    `json:"duration"`
	BytesRead          int64     `json:"bytes_read"`
	BytesWritten       int64     `json:"bytes_written"`
	CommandsCompleted  int64     `json:"commands_completed"`
	IOErrors           int64     `json:"io_errors"`
	ConnectionCount    int       `json:"connection_count"`
}

// iSCSI连接响应
type iSCSIConnectionResponse struct {
	ID            string    `json:"id"`
	SessionID     string    `json:"session_id"`
	ConnectionID  int       `json:"connection_id"`
	InitiatorIP   string    `json:"initiator_ip"`
	InitiatorPort int       `json:"initiator_port"`
	TargetIP      string    `json:"target_ip"`
	TargetPort    int       `json:"target_port"`
	Status        string    `json:"status"`
	ConnectedAt   time.Time `json:"connected_at"`
	LastActivity  time.Time `json:"last_activity"`
	BytesRead     int64     `json:"bytes_read"`
	BytesWritten  int64     `json:"bytes_written"`
	IOPs          int64     `json:"iops"`
}

// iSCSI服务状态响应
type iSCSIServiceResponse struct {
	ID              string     `json:"id"`
	ServiceName     string     `json:"service_name"`
	ProcessID       int        `json:"process_id"`
	Status          string     `json:"status"`
	StartTime       *time.Time `json:"start_time"`
	LastCheckTime   time.Time  `json:"last_check_time"`
	MemoryUsage     int64      `json:"memory_usage"`
	CPUUsage        float64    `json:"cpu_usage"`
	ConnectionCount int        `json:"connection_count"`
	SessionCount    int        `json:"session_count"`
	Uptime          string     `json:"uptime"`
}

// iSCSI性能统计响应
type iSCSIPerformanceStatsResponse struct {
	ID             string    `json:"id"`
	TargetID       string    `json:"target_id"`
	TargetName     string    `json:"target_name"`
	LUNID          string    `json:"lun_id,omitempty"`
	LUNName        string    `json:"lun_name,omitempty"`
	Timestamp      time.Time `json:"timestamp"`
	ReadIOPs       int64     `json:"read_iops"`
	WriteIOPs      int64     `json:"write_iops"`
	TotalIOPs      int64     `json:"total_iops"`
	ReadBandwidth  int64     `json:"read_bandwidth"`
	WriteBandwidth int64     `json:"write_bandwidth"`
	TotalBandwidth int64     `json:"total_bandwidth"`
	ReadLatency    float64   `json:"read_latency"`
	WriteLatency   float64   `json:"write_latency"`
	AverageLatency float64   `json:"average_latency"`
	QueueDepth     int       `json:"queue_depth"`
	QueueTime      float64   `json:"queue_time"`
	ReadErrors     int64     `json:"read_errors"`
	WriteErrors    int64     `json:"write_errors"`
	TimeoutErrors  int64     `json:"timeout_errors"`
	TotalErrors    int64     `json:"total_errors"`
}

// iSCSI审计日志响应
type iSCSIAuditLogResponse struct {
	ID               string    `json:"id"`
	TargetName       string    `json:"target_name"`
	InitiatorName    string    `json:"initiator_name"`
	InitiatorIP      string    `json:"initiator_ip"`
	Operation        string    `json:"operation"`
	LUN              int       `json:"lun"`
	Success          bool      `json:"success"`
	ErrorMessage     string    `json:"error_message,omitempty"`
	Timestamp        time.Time `json:"timestamp"`
	BytesTransferred int64     `json:"bytes_transferred"`
	ResponseTime     float64   `json:"response_time"`
}

// 批量操作请求
type BatchiSCSITargetRequest struct {
	Action    string   `json:"action" binding:"required,oneof=enable disable delete" example:"enable"`
	TargetIDs []string `json:"target_ids" binding:"required"`
}

type BatchiSCSILUNRequest struct {
	Action string   `json:"action" binding:"required,oneof=enable disable delete" example:"enable"`
	LUNIDs []string `json:"lun_ids" binding:"required"`
}

type BatchiSCSIACLRequest struct {
	Action string   `json:"action" binding:"required,oneof=enable disable delete" example:"enable"`
	ACLIDs []string `json:"acl_ids" binding:"required"`
}

// 批量操作响应
type BatchOperationResponse struct {
	Success []string `json:"success"`
	Failed  []string `json:"failed"`
	Errors  []string `json:"errors,omitempty"`
}

// 查询参数
type iSCSITargetListQuery struct {
	Page      int    `form:"page" example:"1"`
	PageSize  int    `form:"page_size" example:"10"`
	Status    string `form:"status" example:"active"`
	IsEnabled *bool  `form:"is_enabled" example:"true"`
	Search    string `form:"search" example:"target01"`
}

type iSCSILUNListQuery struct {
	Page       int    `form:"page" example:"1"`
	PageSize   int    `form:"page_size" example:"10"`
	TargetID   string `form:"target_id"`
	DeviceType string `form:"device_type" example:"file"`
	IsEnabled  *bool  `form:"is_enabled" example:"true"`
}

type iSCSIACLListQuery struct {
	Page       int    `form:"page" example:"1"`
	PageSize   int    `form:"page_size" example:"10"`
	TargetID   string `form:"target_id"`
	Permission string `form:"permission" example:"rw"`
	AuthType   string `form:"auth_type" example:"chap"`
	IsEnabled  *bool  `form:"is_enabled" example:"true"`
}

type iSCSISessionListQuery struct {
	Page      int    `form:"page" example:"1"`
	PageSize  int    `form:"page_size" example:"10"`
	TargetID  string `form:"target_id"`
	Status    string `form:"status" example:"active"`
	ClientIP  string `form:"client_ip" example:"192.168.1.100"`
}

type iSCSIPerformanceStatsQuery struct {
	Page     int       `form:"page" example:"1"`
	PageSize int       `form:"page_size" example:"10"`
	TargetID string    `form:"target_id"`
	LUNID    string    `form:"lun_id"`
	From     time.Time `form:"from" example:"2024-01-01T00:00:00Z"`
	To       time.Time `form:"to" example:"2024-12-31T23:59:59Z"`
}