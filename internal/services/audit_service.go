package services

import (
	"context"
	"crypto/md5"
	"fmt"
	"os"
	"path/filepath"
	"pnas/internal/database"
	"pnas/internal/logging"
	"pnas/internal/models"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// AuditEvent 审计事件结构
type AuditEvent struct {
	UserID      string
	Username    string
	Operation   models.AuditOperation
	FilePath    string
	FileSize    int64
	ClientIP    string
	UserAgent   string
	Status      models.AuditStatus
	ErrorMsg    string
	Duration    time.Duration
	Source      models.AuditSource
	Metadata    map[string]interface{}
}

// AuditService 审计服务
type AuditService struct {
	eventChan     chan *AuditEvent
	batchSize     int
	flushInterval time.Duration
	ctx           context.Context
	cancel        context.CancelFunc
	wg            sync.WaitGroup
	db            *gorm.DB
	logger        *zap.Logger
}

var (
	auditServiceInstance *AuditService
	auditServiceOnce     sync.Once
)

// GetAuditService 获取审计服务单例
func GetAuditService() *AuditService {
	auditServiceOnce.Do(func() {
		auditServiceInstance = NewAuditService()
	})
	return auditServiceInstance
}

// NewAuditService 创建新的审计服务
func NewAuditService() *AuditService {
	ctx, cancel := context.WithCancel(context.Background())

	service := &AuditService{
		eventChan:     make(chan *AuditEvent, 10000), // 缓冲10000个事件
		batchSize:     100,                           // 批量处理100条
		flushInterval: time.Second * 5,               // 5秒刷新一次
		ctx:           ctx,
		cancel:        cancel,
		db:            database.DB,
		logger:        logging.Logger,
	}

	return service
}

// Start 启动审计服务
func (s *AuditService) Start() {
	s.wg.Add(2)
	go s.batchProcessor()
	go s.statisticsUpdater()
	s.logger.Info("Audit service started")
}

// Stop 停止审计服务
func (s *AuditService) Stop() {
	s.cancel()
	close(s.eventChan)
	s.wg.Wait()
	s.logger.Info("Audit service stopped")
}

// LogEvent 记录审计事件（异步）
func (s *AuditService) LogEvent(event *AuditEvent) {
	select {
	case s.eventChan <- event:
		// 事件成功加入队列
	default:
		// 队列满了，记录警告但不阻塞
		s.logger.Warn("Audit event queue is full, dropping event",
			zap.String("operation", string(event.Operation)),
			zap.String("file_path", event.FilePath))
	}
}

// batchProcessor 批量处理器
func (s *AuditService) batchProcessor() {
	defer s.wg.Done()

	batch := make([]*AuditEvent, 0, s.batchSize)
	timer := time.NewTimer(s.flushInterval)
	defer timer.Stop()

	for {
		select {
		case <-s.ctx.Done():
			// 处理剩余的事件
			if len(batch) > 0 {
				s.processBatch(batch)
			}
			return

		case event, ok := <-s.eventChan:
			if !ok {
				// 通道关闭，处理剩余事件
				if len(batch) > 0 {
					s.processBatch(batch)
				}
				return
			}

			batch = append(batch, event)

			// 批次满了，立即处理
			if len(batch) >= s.batchSize {
				s.processBatch(batch)
				batch = batch[:0] // 重置切片
				if !timer.Stop() {
					<-timer.C
				}
				timer.Reset(s.flushInterval)
			}

		case <-timer.C:
			// 定时刷新
			if len(batch) > 0 {
				s.processBatch(batch)
				batch = batch[:0]
			}
			timer.Reset(s.flushInterval)
		}
	}
}

// processBatch 处理批量事件
func (s *AuditService) processBatch(events []*AuditEvent) {
	if len(events) == 0 {
		return
	}

	auditLogs := make([]models.FileAuditLog, len(events))
	for i, event := range events {
		auditLogs[i] = models.FileAuditLog{
			Base: models.Base{
				ID:        uuid.New().String(),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			UserID:    event.UserID,
			Username:  event.Username,
			Operation: string(event.Operation),
			FilePath:  event.FilePath,
			FileName:  filepath.Base(event.FilePath),
			FileSize:  event.FileSize,
			ClientIP:  event.ClientIP,
			UserAgent: event.UserAgent,
			Status:    string(event.Status),
			ErrorMsg:  event.ErrorMsg,
			Duration:  event.Duration.Milliseconds(),
			Hash:      s.calculateFileHash(event.FilePath),
			Source:    string(event.Source),
			Metadata:  s.serializeMetadata(event.Metadata),
		}
	}

	// 批量插入
	if err := s.db.Create(&auditLogs).Error; err != nil {
		s.logger.Error("Failed to batch insert audit logs",
			zap.Error(err),
			zap.Int("batch_size", len(events)))

		// 尝试逐个插入以找出问题记录
		s.fallbackInsert(auditLogs)
	} else {
		s.logger.Debug("Successfully inserted audit logs",
			zap.Int("batch_size", len(events)))
	}

	// 更新文件访问统计
	s.updateFileAccessStats(events)
}

// fallbackInsert 备用插入方法
func (s *AuditService) fallbackInsert(auditLogs []models.FileAuditLog) {
	for _, log := range auditLogs {
		if err := s.db.Create(&log).Error; err != nil {
			s.logger.Error("Failed to insert individual audit log",
				zap.Error(err),
				zap.String("file_path", log.FilePath),
				zap.String("operation", log.Operation))
		}
	}
}

// updateFileAccessStats 更新文件访问统计
func (s *AuditService) updateFileAccessStats(events []*AuditEvent) {
	statsMap := make(map[string]*models.FileAccessStats)
	userMap := make(map[string]map[string]bool) // filePath -> userID -> bool

	for _, event := range events {
		if event.Status != models.AuditStatusSuccess {
			continue
		}

		if stats, exists := statsMap[event.FilePath]; exists {
			stats.AccessCount++
			stats.LastAccess = time.Now()
		} else {
			statsMap[event.FilePath] = &models.FileAccessStats{
				Base: models.Base{
					ID:        uuid.New().String(),
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				FilePath:    event.FilePath,
				FileName:    filepath.Base(event.FilePath),
				AccessCount: 1,
				LastAccess:  time.Now(),
				FirstAccess: time.Now(),
				UserCount:   1,
			}
		}

		// 统计不同用户数量
		if userMap[event.FilePath] == nil {
			userMap[event.FilePath] = make(map[string]bool)
		}
		userMap[event.FilePath][event.UserID] = true
	}

	// 更新用户数量
	for filePath, users := range userMap {
		if stats, exists := statsMap[filePath]; exists {
			stats.UserCount = int64(len(users))
		}
	}

	// 批量更新或插入统计信息
	for _, stats := range statsMap {
		if err := s.db.Where("file_path = ?", stats.FilePath).
			Assign(*stats).
			FirstOrCreate(stats).Error; err != nil {
			s.logger.Error("Failed to update file access stats",
				zap.Error(err),
				zap.String("file_path", stats.FilePath))
		}
	}
}

// statisticsUpdater 定期更新统计信息
func (s *AuditService) statisticsUpdater() {
	defer s.wg.Done()

	ticker := time.NewTicker(time.Hour) // 每小时更新一次
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.updateDailySummary()
		}
	}
}

// updateDailySummary 更新每日汇总
func (s *AuditService) updateDailySummary() {
	today := time.Now().Format("2006-01-02")
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")

	// 汇总昨天的数据
	s.summarizeDay(yesterday)

	// 汇总今天的数据
	s.summarizeDay(today)
}

// summarizeDay 汇总指定日期的数据
func (s *AuditService) summarizeDay(date string) {
	startTime, _ := time.Parse("2006-01-02", date)
	endTime := startTime.Add(24 * time.Hour)

	var summaries []models.AuditSummary

	// 按用户和操作分组统计
	result := s.db.Raw(`
		SELECT
			? as date,
			user_id,
			username,
			operation,
			COUNT(*) as total_count,
			SUM(CASE WHEN status = 'success' THEN 1 ELSE 0 END) as success_count,
			SUM(CASE WHEN status = 'failed' THEN 1 ELSE 0 END) as failed_count,
			SUM(CASE WHEN status = 'success' THEN file_size ELSE 0 END) as total_size
		FROM file_audit_logs
		WHERE created_at >= ? AND created_at < ?
		GROUP BY user_id, username, operation
	`, date, startTime, endTime).Scan(&summaries)

	if result.Error != nil {
		s.logger.Error("Failed to generate daily summary",
			zap.Error(result.Error),
			zap.String("date", date))
		return
	}

	// 更新或插入汇总数据
	for _, summary := range summaries {
		summary.ID = uuid.New().String()
		summary.CreatedAt = time.Now()
		summary.UpdatedAt = time.Now()

		if err := s.db.Where("date = ? AND user_id = ? AND operation = ?",
			summary.Date, summary.UserID, summary.Operation).
			Assign(summary).
			FirstOrCreate(&summary).Error; err != nil {
			s.logger.Error("Failed to upsert audit summary",
				zap.Error(err),
				zap.String("date", date))
		}
	}
}

// calculateFileHash 计算文件哈希
func (s *AuditService) calculateFileHash(filePath string) string {
	if info, err := os.Stat(filePath); err != nil || info.IsDir() {
		return ""
	}

	// 对于大文件，只计算前1MB的哈希以提高性能
	file, err := os.Open(filePath)
	if err != nil {
		return ""
	}
	defer file.Close()

	hash := md5.New()
	buffer := make([]byte, 1024*1024) // 1MB buffer

	if _, err := file.Read(buffer); err != nil {
		return ""
	}

	hash.Write(buffer)
	return fmt.Sprintf("%x", hash.Sum(nil))
}

// serializeMetadata 序列化元数据
func (s *AuditService) serializeMetadata(metadata map[string]interface{}) string {
	if len(metadata) == 0 {
		return ""
	}

	// 简单的键值对序列化
	result := ""
	for k, v := range metadata {
		if result != "" {
			result += ","
		}
		result += fmt.Sprintf("%s=%v", k, v)
	}

	return result
}

// LogFileOperation 便捷方法：记录文件操作
func LogFileOperation(userID, username string, operation models.AuditOperation,
	filePath string, status models.AuditStatus, source models.AuditSource,
	opts ...func(*AuditEvent)) {

	event := &AuditEvent{
		UserID:    userID,
		Username:  username,
		Operation: operation,
		FilePath:  filePath,
		Status:    status,
		Source:    source,
	}

	// 应用可选参数
	for _, opt := range opts {
		opt(event)
	}

	// 获取文件大小
	if info, err := os.Stat(filePath); err == nil && !info.IsDir() {
		event.FileSize = info.Size()
	}

	GetAuditService().LogEvent(event)
}

// 便捷设置选项
func WithClientInfo(ip, userAgent string) func(*AuditEvent) {
	return func(e *AuditEvent) {
		e.ClientIP = ip
		e.UserAgent = userAgent
	}
}

func WithError(err error) func(*AuditEvent) {
	return func(e *AuditEvent) {
		e.Status = models.AuditStatusFailed
		e.ErrorMsg = err.Error()
	}
}

func WithDuration(duration time.Duration) func(*AuditEvent) {
	return func(e *AuditEvent) {
		e.Duration = duration
	}
}

func WithMetadata(metadata map[string]interface{}) func(*AuditEvent) {
	return func(e *AuditEvent) {
		e.Metadata = metadata
	}
}

func WithFileSize(size int64) func(*AuditEvent) {
	return func(e *AuditEvent) {
		e.FileSize = size
	}
}