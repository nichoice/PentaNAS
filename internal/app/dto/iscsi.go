package dto

import (
	"time"
	"pnas/internal/models"
)

// CreateISCSITargetRequest 创建iSCSI目标请求
type CreateISCSITargetRequest struct {
	Name      string `json:"name" binding:"required"`    // IQN format
	Alias     string `json:"alias"`
	Comment   string `json:"comment"`
	IsEnabled bool   `json:"is_enabled"`
}

// UpdateISCSITargetRequest 更新iSCSI目标请求
type UpdateISCSITargetRequest struct {
	Alias     *string `json:"alias"`
	Comment   *string `json:"comment"`
	IsEnabled *bool   `json:"is_enabled"`
}

// ISCSITargetResponse iSCSI目标响应
type ISCSITargetResponse struct {
	ID        string                         `json:"id"`
	Name      string                         `json:"name"`
	Alias     string                         `json:"alias"`
	Status    models.ISCSITargetStatus       `json:"status"`
	Comment   string                         `json:"comment"`
	IsEnabled bool                           `json:"is_enabled"`
	CreatedAt time.Time                      `json:"created_at"`
	UpdatedAt time.Time                      `json:"updated_at"`
}

// ISCSITargetListResponse 用于分页返回 Target 列表及总数信息。
type ISCSITargetListResponse struct {
	Targets []ISCSITargetResponse `json:"targets"`
	Total   int64                 `json:"total"`
	Page    int                   `json:"page"`
	Limit   int                   `json:"limit"`
}

// ISCSITargetStatusResponse 概述 Target 的运行状态与活跃会话数量。
type ISCSITargetStatusResponse struct {
	ID            string                   `json:"id"`
	Name          string                   `json:"name"`
	Status        models.ISCSITargetStatus `json:"status"`
	ActiveSessions int                     `json:"active_sessions"`
	TotalLUNs     int                      `json:"total_luns"`
	LastActivity  *time.Time               `json:"last_activity,omitempty"`
}

// CreateISCSILUNRequest 创建iSCSI LUN请求
type CreateISCSILUNRequest struct {
	Name         string                    `json:"name" binding:"required"`
	DeviceType   models.ISCSIDeviceType    `json:"device_type"`
	Size         int64                     `json:"size" binding:"required"`
	DevicePath   string                    `json:"device_path"`
	StoragePoolID *string                  `json:"storage_pool_id"`
	Comment      string                    `json:"comment"`
	IsEnabled    bool                      `json:"is_enabled"`
	ReadOnly     bool                      `json:"read_only"`
	BlockSize    int                       `json:"block_size"`
}

// UpdateISCSILUNRequest 更新iSCSI LUN请求
type UpdateISCSILUNRequest struct {
	Comment     *string `json:"comment"`
	IsEnabled   *bool   `json:"is_enabled"`
	ReadOnly    *bool   `json:"read_only"`
	BlockSize   *int    `json:"block_size"`
}

// ISCSILUNResponse iSCSI LUN响应
type ISCSILUNResponse struct {
	ID            string                    `json:"id"`
	Name          string                    `json:"name"`
	DeviceType    models.ISCSIDeviceType    `json:"device_type"`
	Size          int64                     `json:"size"`
	DevicePath    string                    `json:"device_path"`
	StoragePoolID *string                   `json:"storage_pool_id"`
	Comment       string                    `json:"comment"`
	IsEnabled     bool                      `json:"is_enabled"`
	ReadOnly      bool                      `json:"read_only"`
	BlockSize     int                       `json:"block_size"`
	CreatedAt     time.Time                 `json:"created_at"`
	UpdatedAt     time.Time                 `json:"updated_at"`
}

// ISCSILUNListResponse 携带 LUN 列表与分页统计，方便前端渲染表格。
type ISCSILUNListResponse struct {
	LUNs  []ISCSILUNResponse `json:"luns"`
	Total int64              `json:"total"`
	Page  int                `json:"page"`
	Limit int                `json:"limit"`
}

// MapLUNRequest 映射LUN请求
type MapLUNRequest struct {
	TargetID string `json:"target_id" binding:"required"`
	LUN      int    `json:"lun" binding:"required"`
}

// ISCSILUNMappingResponse 返回 LUN 与 Target 的映射结果。
type ISCSILUNMappingResponse struct {
	ID       string    `json:"id"`
	TargetID string    `json:"target_id"`
	LUNID    string    `json:"lun_id"`
	LUN      int       `json:"lun"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateISCSIACLRequest 创建iSCSI ACL请求
type CreateISCSIACLRequest struct {
	TargetID       string                     `json:"target_id" binding:"required"`
	InitiatorIQN   string                     `json:"initiator_iqn" binding:"required"`
	AuthType       models.ISCSIAuthType       `json:"auth_type"`
	Username       string                     `json:"username"`
	Password       string                     `json:"password"`
	MutualUsername string                     `json:"mutual_username"`
	MutualPassword string                     `json:"mutual_password"`
	Permission     models.ISCSIPermission     `json:"permission"`
	Comment        string                     `json:"comment"`
	IsEnabled      bool                       `json:"is_enabled"`
}

// UpdateISCSIACLRequest 更新iSCSI ACL请求
type UpdateISCSIACLRequest struct {
	AuthType       *models.ISCSIAuthType      `json:"auth_type"`
	Username       *string                    `json:"username"`
	Password       *string                    `json:"password"`
	MutualUsername *string                    `json:"mutual_username"`
	MutualPassword *string                    `json:"mutual_password"`
	Permission     *models.ISCSIPermission    `json:"permission"`
	Comment        *string                    `json:"comment"`
	IsEnabled      *bool                      `json:"is_enabled"`
}

// ISCSIACLResponse iSCSI ACL响应
type ISCSIACLResponse struct {
	ID             string                     `json:"id"`
	TargetID       string                     `json:"target_id"`
	InitiatorIQN   string                     `json:"initiator_iqn"`
	AuthType       models.ISCSIAuthType       `json:"auth_type"`
	Username       string                     `json:"username"`
	MutualUsername string                     `json:"mutual_username"`
	Permission     models.ISCSIPermission     `json:"permission"`
	Comment        string                     `json:"comment"`
	IsEnabled      bool                       `json:"is_enabled"`
	CreatedAt      time.Time                  `json:"created_at"`
	UpdatedAt      time.Time                  `json:"updated_at"`
}

// ISCSIACLListResponse 封装 ACL 集合的分页数据。
type ISCSIACLListResponse struct {
	ACLs  []ISCSIACLResponse `json:"acls"`
	Total int64              `json:"total"`
	Page  int                `json:"page"`
	Limit int                `json:"limit"`
}

// UpdateISCSIGlobalConfigRequest 更新iSCSI全局配置请求
type UpdateISCSIGlobalConfigRequest struct {
	TargetPort                   *int    `json:"target_port"`
	MaxSessions                  *int    `json:"max_sessions"`
	MaxConnections               *int    `json:"max_connections"`
	MaxRecvDataSegmentLength     *int    `json:"max_recv_data_segment_length"`
	MaxXmitDataSegmentLength     *int    `json:"max_xmit_data_segment_length"`
	MaxBurstLength               *int    `json:"max_burst_length"`
	FirstBurstLength             *int    `json:"first_burst_length"`
	MaxOutstandingR2T            *int    `json:"max_outstanding_r2t"`
	DefaultTime2Wait             *int    `json:"default_time2_wait"`
	DefaultTime2Retain           *int    `json:"default_time2_retain"`
	LoginTimeout                 *int    `json:"login_timeout"`
	LogoutTimeout                *int    `json:"logout_timeout"`
	RequireAuth                  *bool   `json:"require_auth"`
	AllowDuplicateSessions       *bool   `json:"allow_duplicate_sessions"`
	LogLevel                     *int    `json:"log_level"`
	LogFile                      *string `json:"log_file"`
	EnableDebugLog               *bool   `json:"enable_debug_log"`
}

// ISCSIGlobalConfigResponse 返回当前生效的 iSCSI 全局配置。
type ISCSIGlobalConfigResponse struct {
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

// ISCSIServiceStatusResponse 汇总底层 iSCSI 服务的运行概况。
type ISCSIServiceStatusResponse struct {
	ServiceName     string                    `json:"service_name"`
	Status          string                    `json:"status"`
	ProcessID       int                       `json:"process_id"`
	ActiveTargets   int                       `json:"active_targets"`
	ActiveSessions  int                       `json:"active_sessions"`
	TotalConnections int                      `json:"total_connections"`
	Uptime          string                    `json:"uptime"`
	LastStarted     *time.Time                `json:"last_started,omitempty"`
}

// ISCSISessionResponse 描述单个 iSCSI 会话的基础信息与统计。
type ISCSISessionResponse struct {
	ID               string    `json:"id"`
	TargetID         string    `json:"target_id"`
	InitiatorIQN     string    `json:"initiator_iqn"`
	InitiatorAddress string    `json:"initiator_address"`
	SessionType      string    `json:"session_type"`
	Status           string    `json:"status"`
	ConnectionCount  int       `json:"connection_count"`
	CreatedAt        time.Time `json:"created_at"`
	LastActivity     time.Time `json:"last_activity"`
}

// ISCSISessionListResponse 用于分页返回当前会话列表。
type ISCSISessionListResponse struct {
	Sessions []ISCSISessionResponse `json:"sessions"`
	Total    int64                  `json:"total"`
	Page     int                    `json:"page"`
	Limit    int                    `json:"limit"`
}

// ISCSIConnectionResponse 描述单条 iSCSI 网络连接。
type ISCSIConnectionResponse struct {
	ID               string    `json:"id"`
	SessionID        string    `json:"session_id"`
	TargetID         string    `json:"target_id"`
	InitiatorAddress string    `json:"initiator_address"`
	TargetAddress    string    `json:"target_address"`
	Status           string    `json:"status"`
	BytesTransferred int64     `json:"bytes_transferred"`
	CreatedAt        time.Time `json:"created_at"`
	LastActivity     time.Time `json:"last_activity"`
}

// ISCSIConnectionListResponse 返回连接集合及分页元信息。
type ISCSIConnectionListResponse struct {
	Connections []ISCSIConnectionResponse `json:"connections"`
	Total       int64                     `json:"total"`
	Page        int                       `json:"page"`
	Limit       int                       `json:"limit"`
}

// ISCSIConnectionHistoryResponse 用于呈现历史连接记录。
type ISCSIConnectionHistoryResponse struct {
	History []ISCSIConnectionResponse `json:"history"`
	Total   int64                     `json:"total"`
	Page    int                       `json:"page"`
	Limit   int                       `json:"limit"`
}

// CreateISCSIStoragePoolRequest 创建iSCSI存储池请求
type CreateISCSIStoragePoolRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Path        string `json:"path" binding:"required"`
	MaxSize     int64  `json:"max_size"`
	IsEnabled   bool   `json:"is_enabled"`
}

// UpdateISCSIStoragePoolRequest 更新iSCSI存储池请求
type UpdateISCSIStoragePoolRequest struct {
	Description *string `json:"description"`
	MaxSize     *int64  `json:"max_size"`
	IsEnabled   *bool   `json:"is_enabled"`
}

// ISCSIStoragePoolResponse iSCSI存储池响应
type ISCSIStoragePoolResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Path        string    `json:"path"`
	MaxSize     int64     `json:"max_size"`
	UsedSize    int64     `json:"used_size"`
	FreeSize    int64     `json:"free_size"`
	IsEnabled   bool      `json:"is_enabled"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ISCSIStoragePoolListResponse 封装可用存储池集合及分页信息。
type ISCSIStoragePoolListResponse struct {
	Pools []ISCSIStoragePoolResponse `json:"pools"`
	Total int64                      `json:"total"`
	Page  int                        `json:"page"`
	Limit int                        `json:"limit"`
}

// ISCSIPerformanceStatsResponse 汇总 iSCSI 服务整体性能指标。
type ISCSIPerformanceStatsResponse struct {
	TotalIOPS        int64   `json:"total_iops"`
	ReadIOPS         int64   `json:"read_iops"`
	WriteIOPS        int64   `json:"write_iops"`
	TotalThroughput  int64   `json:"total_throughput"`
	ReadThroughput   int64   `json:"read_throughput"`
	WriteThroughput  int64   `json:"write_throughput"`
	AverageLatency   float64 `json:"average_latency"`
	ReadLatency      float64 `json:"read_latency"`
	WriteLatency     float64 `json:"write_latency"`
	ActiveSessions   int     `json:"active_sessions"`
	ActiveTargets    int     `json:"active_targets"`
	TotalDataTransfer int64  `json:"total_data_transfer"`
	ErrorCount       int64   `json:"error_count"`
}

// ISCSITargetPerformanceResponse 展示单个 Target 的性能概要。
type ISCSITargetPerformanceResponse struct {
	TargetID        string  `json:"target_id"`
	TargetName      string  `json:"target_name"`
	IOPS            int64   `json:"iops"`
	Throughput      int64   `json:"throughput"`
	AverageLatency  float64 `json:"average_latency"`
	ActiveSessions  int     `json:"active_sessions"`
	DataTransfer    int64   `json:"data_transfer"`
	ErrorCount      int64   `json:"error_count"`
}

// ISCSILUNPerformanceResponse 用于分析 LUN 的 IOPS/带宽趋势。
type ISCSILUNPerformanceResponse struct {
	LUNID          string  `json:"lun_id"`
	LUNName        string  `json:"lun_name"`
	IOPS           int64   `json:"iops"`
	Throughput     int64   `json:"throughput"`
	AverageLatency float64 `json:"average_latency"`
	DataTransfer   int64   `json:"data_transfer"`
	ErrorCount     int64   `json:"error_count"`
}

// ISCSIAuditLogResponse 表示单条审计日志及附加上下文。
type ISCSIAuditLogResponse struct {
	ID          string                 `json:"id"`
	Action      string                  `json:"action"`
	ResourceID  string                 `json:"resource_id"`
	ResourceType string                `json:"resource_type"`
	UserID      string                 `json:"user_id"`
	ClientIP    string                 `json:"client_ip"`
	Details     map[string]interface{} `json:"details"`
	Status      string                 `json:"status"`
	ErrorMessage string                `json:"error_message,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
}

// ISCSIAuditLogListResponse 返回审计日志分页数据。
type ISCSIAuditLogListResponse struct {
	Logs  []ISCSIAuditLogResponse `json:"logs"`
	Total int64                   `json:"total"`
	Page  int                     `json:"page"`
	Limit int                     `json:"limit"`
}

// ISCSIAuditStatsResponse 聚合审计行为的统计数据。
type ISCSIAuditStatsResponse struct {
	TotalActions       int64            `json:"total_actions"`
	SuccessfulActions  int64            `json:"successful_actions"`
	FailedActions      int64            `json:"failed_actions"`
	ActionsByType      map[string]int64 `json:"actions_by_type"`
	ActionsByUser      map[string]int64 `json:"actions_by_user"`
	ActionsByHour      map[string]int64 `json:"actions_by_hour"`
	MostActiveUsers    []string         `json:"most_active_users"`
	MostCommonActions  []string         `json:"most_common_actions"`
}
