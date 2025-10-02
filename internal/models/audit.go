package models

import (
	"time"
)

// FileAuditLog 文件操作审计日志
type FileAuditLog struct {
	Base
	UserID    string `gorm:"type:varchar(36);index" json:"user_id"`
	Username  string `gorm:"type:varchar(50);index" json:"username"`
	Operation string `gorm:"type:varchar(20);index;not null" json:"operation"`
	FilePath  string `gorm:"type:text;not null" json:"file_path"`
	FileName  string `gorm:"type:varchar(255);index" json:"file_name"`
	FileSize  int64  `gorm:"type:bigint" json:"file_size"`
	ClientIP  string `gorm:"type:varchar(45)" json:"client_ip"`
	UserAgent string `gorm:"type:text" json:"user_agent"`
	Status    string `gorm:"type:varchar(10);index;not null" json:"status"`
	ErrorMsg  string `gorm:"type:text" json:"error_msg,omitempty"`
	Duration  int64  `gorm:"type:bigint" json:"duration"`
	Hash      string `gorm:"type:varchar(64)" json:"hash,omitempty"`
	Source    string `gorm:"type:varchar(10);index;not null" json:"source"`
	Metadata  string `gorm:"type:text" json:"metadata,omitempty"`
}

// AuditOperation 审计操作类型
type AuditOperation string

const (
	AuditOperationCreate   AuditOperation = "create"
	AuditOperationRead     AuditOperation = "read"
	AuditOperationUpdate   AuditOperation = "update"
	AuditOperationDelete   AuditOperation = "delete"
	AuditOperationRename   AuditOperation = "rename"
	AuditOperationMove     AuditOperation = "move"
	AuditOperationCopy     AuditOperation = "copy"
	AuditOperationUpload   AuditOperation = "upload"
	AuditOperationDownload AuditOperation = "download"
	AuditOperationCompress AuditOperation = "compress"
	AuditOperationExtract  AuditOperation = "extract"
)

// AuditStatus 审计状态
type AuditStatus string

const (
	AuditStatusSuccess AuditStatus = "success"
	AuditStatusFailed  AuditStatus = "failed"
)

// AuditSource 审计来源
type AuditSource string

const (
	AuditSourceAPI        AuditSource = "api"
	AuditSourceFilesystem AuditSource = "filesystem"
)

// FileAccessStats 文件访问统计
type FileAccessStats struct {
	Base
	FilePath    string    `gorm:"type:text;uniqueIndex;not null" json:"file_path"`
	FileName    string    `gorm:"type:varchar(255);index" json:"file_name"`
	AccessCount int64     `gorm:"type:bigint;default:0" json:"access_count"`
	LastAccess  time.Time `json:"last_access"`
	FirstAccess time.Time `json:"first_access"`
	UserCount   int64     `gorm:"type:bigint;default:0" json:"user_count"`
}

// AuditSummary 审计汇总信息
type AuditSummary struct {
	Base
	Date         string `gorm:"type:date;uniqueIndex;not null" json:"date"`
	UserID       string `gorm:"type:varchar(36);index" json:"user_id"`
	Username     string `gorm:"type:varchar(50);index" json:"username"`
	Operation    string `gorm:"type:varchar(20);index" json:"operation"`
	TotalCount   int64  `gorm:"type:bigint;default:0" json:"total_count"`
	SuccessCount int64  `gorm:"type:bigint;default:0" json:"success_count"`
	FailedCount  int64  `gorm:"type:bigint;default:0" json:"failed_count"`
	TotalSize    int64  `gorm:"type:bigint;default:0" json:"total_size"`
}

// TableName 指定表名
func (FileAuditLog) TableName() string {
	return "file_audit_logs"
}

func (FileAccessStats) TableName() string {
	return "file_access_stats"
}

func (AuditSummary) TableName() string {
	return "audit_summaries"
}
