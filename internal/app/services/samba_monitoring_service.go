package services

import (
	"bufio"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"pnas/internal/app/dto"
	"pnas/internal/models"
)

type SambaMonitoringService struct {
	db *gorm.DB
}

func NewSambaMonitoringService(db *gorm.DB) *SambaMonitoringService {
	return &SambaMonitoringService{db: db}
}

// GetSambaStatus 获取Samba整体状态
func (s *SambaMonitoringService) GetSambaStatus() (*dto.SambaStatusResponse, error) {
	// 检查服务状态
	services, err := s.GetServiceStatus()
	if err != nil {
		return nil, fmt.Errorf("获取服务状态失败: %w", err)
	}

	// 获取版本信息
	version, err := s.getSambaVersion()
	if err != nil {
		version = "unknown"
	}

	// 获取运行时间
	uptime, err := s.getSambaUptime()
	if err != nil {
		uptime = "unknown"
	}

	// 获取共享统计
	totalShares, activeShares, err := s.getShareStatistics()
	if err != nil {
		totalShares, activeShares = 0, 0
	}

	// 获取账号统计
	totalAccounts, activeAccounts, err := s.getAccountStatistics()
	if err != nil {
		totalAccounts, activeAccounts = 0, 0
	}

	// 获取活跃连接
	connections, err := s.GetActiveConnections()
	if err != nil {
		connections = []*dto.SambaConnectionResponse{}
	}

	// 检查是否运行
	isRunning := false
	for _, service := range services {
		if service.ServiceName == "smbd" && service.Status == "running" {
			isRunning = true
			break
		}
	}

	// Convert slices
	serviceResponses := make([]dto.SambaServiceResponse, len(services))
	for i, service := range services {
		serviceResponses[i] = *service
	}

	connectionResponses := make([]dto.SambaConnectionResponse, len(connections))
	for i, conn := range connections {
		connectionResponses[i] = *conn
	}

	return &dto.SambaStatusResponse{
		IsRunning:         isRunning,
		Version:           version,
		Uptime:            uptime,
		TotalShares:       totalShares,
		ActiveShares:      activeShares,
		TotalAccounts:     totalAccounts,
		ActiveAccounts:    activeAccounts,
		ActiveConnections: len(connections),
		Services:          serviceResponses,
		RecentConnections: connectionResponses,
	}, nil
}

// GetServiceStatus 获取Samba服务状态
func (s *SambaMonitoringService) GetServiceStatus() ([]*dto.SambaServiceResponse, error) {
	services := []string{"smbd", "nmbd", "winbindd"}
	var responses []*dto.SambaServiceResponse

	for _, serviceName := range services {
		status, err := s.getSystemServiceStatus(serviceName)
		if err != nil {
			// 如果服务不存在，跳过
			continue
		}

		service := &dto.SambaServiceResponse{
			ID:            uuid.New().String(),
			ServiceName:   serviceName,
			ProcessID:     status.ProcessID,
			Status:        status.Status,
			StartTime:     status.StartTime,
			LastCheckTime: time.Now(),
			MemoryUsage:   status.MemoryUsage,
			CPUUsage:      status.CPUUsage,
		}

		// 获取连接数（仅对smbd）
		if serviceName == "smbd" {
			connectionCount, err := s.getConnectionCount()
			if err == nil {
				service.ConnectionCount = connectionCount
			}
		}

		responses = append(responses, service)

		// 保存或更新服务状态到数据库
		s.saveServiceStatus(service)
	}

	return responses, nil
}

// GetActiveConnections 获取活跃连接
func (s *SambaMonitoringService) GetActiveConnections() ([]*dto.SambaConnectionResponse, error) {
	// 使用smbstatus命令获取连接信息
	cmd := exec.Command("smbstatus", "-b")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("获取Samba连接状态失败: %w", err)
	}

	connections := s.parseConnections(string(output))

	// 保存连接信息到数据库
	for _, conn := range connections {
		s.saveConnection(conn)
	}

	return connections, nil
}

// GetConnectionHistory 获取连接历史
func (s *SambaMonitoringService) GetConnectionHistory(hours int, offset, limit int) ([]*dto.SambaConnectionResponse, int64, error) {
	var connections []models.SambaConnection
	var total int64

	since := time.Now().Add(-time.Duration(hours) * time.Hour)
	query := s.db.Model(&models.SambaConnection{}).Where("connected_at >= ?", since)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("统计连接历史数量失败: %w", err)
	}

	if err := query.Order("connected_at DESC").
		Offset(offset).Limit(limit).Find(&connections).Error; err != nil {
		return nil, 0, fmt.Errorf("获取连接历史失败: %w", err)
	}

	responses := make([]*dto.SambaConnectionResponse, len(connections))
	for i, conn := range connections {
		responses[i] = s.toConnectionResponse(&conn)
	}

	return responses, total, nil
}

// GetAuditLogs 获取审计日志
func (s *SambaMonitoringService) GetAuditLogs(filters map[string]interface{}, offset, limit int) ([]*dto.SambaAuditLogResponse, int64, error) {
	var logs []models.SambaAuditLog
	var total int64

	query := s.db.Model(&models.SambaAuditLog{})

	// 应用过滤条件
	if username, ok := filters["username"].(string); ok && username != "" {
		query = query.Where("username LIKE ?", "%"+username+"%")
	}
	if clientIP, ok := filters["client_ip"].(string); ok && clientIP != "" {
		query = query.Where("client_ip = ?", clientIP)
	}
	if shareName, ok := filters["share_name"].(string); ok && shareName != "" {
		query = query.Where("share_name = ?", shareName)
	}
	if operation, ok := filters["operation"].(string); ok && operation != "" {
		query = query.Where("operation = ?", operation)
	}
	if success, ok := filters["success"].(bool); ok {
		query = query.Where("success = ?", success)
	}

	// 时间范围过滤
	if startTime, ok := filters["start_time"].(time.Time); ok {
		query = query.Where("timestamp >= ?", startTime)
	}
	if endTime, ok := filters["end_time"].(time.Time); ok {
		query = query.Where("timestamp <= ?", endTime)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("统计审计日志数量失败: %w", err)
	}

	if err := query.Order("timestamp DESC").
		Offset(offset).Limit(limit).Find(&logs).Error; err != nil {
		return nil, 0, fmt.Errorf("获取审计日志失败: %w", err)
	}

	responses := make([]*dto.SambaAuditLogResponse, len(logs))
	for i, log := range logs {
		responses[i] = s.toAuditLogResponse(&log)
	}

	return responses, total, nil
}

// EnableMultiChannel 启用多通道支持
func (s *SambaMonitoringService) EnableMultiChannel(shareID string) error {
	// 获取共享
	var share models.SambaShare
	if err := s.db.First(&share, "id = ?", shareID).Error; err != nil {
		return fmt.Errorf("共享不存在: %w", err)
	}

	// 检查系统是否支持多通道
	if err := s.validateMultiChannelSupport(); err != nil {
		return fmt.Errorf("系统不支持多通道: %w", err)
	}

	// 更新共享配置
	if err := s.db.Model(&share).Update("enable_multi_channel", true).Error; err != nil {
		return fmt.Errorf("启用多通道失败: %w", err)
	}

	return nil
}

// DisableMultiChannel 禁用多通道支持
func (s *SambaMonitoringService) DisableMultiChannel(shareID string) error {
	// 获取共享
	var share models.SambaShare
	if err := s.db.First(&share, "id = ?", shareID).Error; err != nil {
		return fmt.Errorf("共享不存在: %w", err)
	}

	// 更新共享配置
	if err := s.db.Model(&share).Update("enable_multi_channel", false).Error; err != nil {
		return fmt.Errorf("禁用多通道失败: %w", err)
	}

	return nil
}

// GetMultiChannelStatus 获取多通道状态
func (s *SambaMonitoringService) GetMultiChannelStatus() (map[string]interface{}, error) {
	// 检查全局多通道配置
	var config models.SambaGlobalConfig
	if err := s.db.Where("is_active = ?", true).First(&config).Error; err != nil {
		return nil, fmt.Errorf("获取全局配置失败: %w", err)
	}

	// 获取启用多通道的共享数量
	var enabledShares int64
	if err := s.db.Model(&models.SambaShare{}).
		Where("enable_multi_channel = ? AND is_enabled = ?", true, true).
		Count(&enabledShares).Error; err != nil {
		return nil, fmt.Errorf("统计多通道共享数量失败: %w", err)
	}

	// 检查网络接口数量
	interfaces, err := s.getNetworkInterfaces()
	if err != nil {
		interfaces = []string{}
	}

	status := map[string]interface{}{
		"global_enabled":     config.EnableMultiChannel,
		"max_channels":       config.MaxChannels,
		"enabled_shares":     enabledShares,
		"network_interfaces": interfaces,
		"system_support":     s.validateMultiChannelSupport() == nil,
	}

	return status, nil
}

// 私有方法

// getSystemServiceStatus 获取系统服务状态
func (s *SambaMonitoringService) getSystemServiceStatus(serviceName string) (*ServiceStatus, error) {
	// 使用systemctl获取服务状态
	cmd := exec.Command("systemctl", "status", serviceName)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("获取服务状态失败: %w", err)
	}

	status := &ServiceStatus{
		ServiceName: serviceName,
	}

	// 解析systemctl输出
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "Active:") {
			if strings.Contains(line, "active (running)") {
				status.Status = "running"
			} else if strings.Contains(line, "inactive") {
				status.Status = "stopped"
			} else {
				status.Status = "error"
			}
		}
		if strings.Contains(line, "Main PID:") {
			re := regexp.MustCompile(`Main PID: (\d+)`)
			matches := re.FindStringSubmatch(line)
			if len(matches) > 1 {
				if pid, err := strconv.Atoi(matches[1]); err == nil {
					status.ProcessID = pid
				}
			}
		}
	}

	// 获取进程资源使用情况
	if status.ProcessID > 0 {
		status.MemoryUsage, status.CPUUsage = s.getProcessResourceUsage(status.ProcessID)
	}

	return status, nil
}

// parseConnections 解析smbstatus输出
func (s *SambaMonitoringService) parseConnections(output string) []*dto.SambaConnectionResponse {
	var connections []*dto.SambaConnectionResponse

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "Samba version") ||
		   strings.HasPrefix(line, "PID") || strings.HasPrefix(line, "----") {
			continue
		}

		// 解析连接行，格式通常为：PID Username Group Machine Protocol Version Encryption Signing
		fields := strings.Fields(line)
		if len(fields) >= 4 {
			connection := &dto.SambaConnectionResponse{
				ID:             uuid.New().String(),
				ClientIP:       s.extractClientIP(fields[3]),
				ClientHostname: fields[3],
				Username:       fields[1],
				ConnectedAt:    time.Now(), // smbstatus不提供连接时间，使用当前时间
				LastActivity:   time.Now(),
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Now(),
			}
			connections = append(connections, connection)
		}
	}

	return connections
}

// extractClientIP 从机器名中提取IP地址
func (s *SambaMonitoringService) extractClientIP(machine string) string {
	// 如果包含括号，提取括号中的IP
	re := regexp.MustCompile(`\(([0-9.]+)\)`)
	matches := re.FindStringSubmatch(machine)
	if len(matches) > 1 {
		return matches[1]
	}
	return machine
}

// getConnectionCount 获取连接数
func (s *SambaMonitoringService) getConnectionCount() (int, error) {
	cmd := exec.Command("smbstatus", "-b")
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	lines := strings.Split(string(output), "\n")
	count := 0
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "Samba version") &&
		   !strings.HasPrefix(line, "PID") && !strings.HasPrefix(line, "----") {
			count++
		}
	}

	return count, nil
}

// getSambaVersion 获取Samba版本
func (s *SambaMonitoringService) getSambaVersion() (string, error) {
	cmd := exec.Command("smbd", "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	version := strings.TrimSpace(string(output))
	parts := strings.Fields(version)
	if len(parts) >= 2 {
		return parts[1], nil
	}
	return version, nil
}

// getSambaUptime 获取Samba运行时间
func (s *SambaMonitoringService) getSambaUptime() (string, error) {
	// 通过ps命令获取smbd进程的启动时间
	cmd := exec.Command("ps", "-eo", "pid,etime,comm")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "smbd") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				return fields[1], nil
			}
		}
	}

	return "unknown", nil
}

// getShareStatistics 获取共享统计
func (s *SambaMonitoringService) getShareStatistics() (int, int, error) {
	var totalShares int64
	var activeShares int64

	if err := s.db.Model(&models.SambaShare{}).Count(&totalShares).Error; err != nil {
		return 0, 0, err
	}

	if err := s.db.Model(&models.SambaShare{}).Where("is_enabled = ?", true).Count(&activeShares).Error; err != nil {
		return 0, 0, err
	}

	return int(totalShares), int(activeShares), nil
}

// getAccountStatistics 获取账号统计
func (s *SambaMonitoringService) getAccountStatistics() (int, int, error) {
	var totalAccounts int64
	var activeAccounts int64

	if err := s.db.Model(&models.SambaAccount{}).Count(&totalAccounts).Error; err != nil {
		return 0, 0, err
	}

	if err := s.db.Model(&models.SambaAccount{}).Where("is_enabled = ?", true).Count(&activeAccounts).Error; err != nil {
		return 0, 0, err
	}

	return int(totalAccounts), int(activeAccounts), nil
}

// getProcessResourceUsage 获取进程资源使用情况
func (s *SambaMonitoringService) getProcessResourceUsage(pid int) (int64, float64) {
	// 使用ps命令获取进程资源使用情况
	cmd := exec.Command("ps", "-o", "rss,pcpu", "-p", strconv.Itoa(pid))
	output, err := cmd.Output()
	if err != nil {
		return 0, 0
	}

	lines := strings.Split(string(output), "\n")
	if len(lines) < 2 {
		return 0, 0
	}

	fields := strings.Fields(lines[1])
	if len(fields) < 2 {
		return 0, 0
	}

	// RSS内存使用量（KB）
	memory, _ := strconv.ParseInt(fields[0], 10, 64)
	// CPU使用百分比
	cpu, _ := strconv.ParseFloat(fields[1], 64)

	return memory, cpu
}

// validateMultiChannelSupport 验证多通道支持
func (s *SambaMonitoringService) validateMultiChannelSupport() error {
	// 检查Samba版本是否支持多通道（需要4.4+）
	version, err := s.getSambaVersion()
	if err != nil {
		return fmt.Errorf("无法获取Samba版本: %w", err)
	}

	// 简单的版本检查（实际应该更严格）
	if !strings.Contains(version, "4.") ||
	   (strings.Contains(version, "4.0") || strings.Contains(version, "4.1") ||
	    strings.Contains(version, "4.2") || strings.Contains(version, "4.3")) {
		return fmt.Errorf("Samba版本 %s 不支持多通道功能", version)
	}

	return nil
}

// getNetworkInterfaces 获取网络接口
func (s *SambaMonitoringService) getNetworkInterfaces() ([]string, error) {
	cmd := exec.Command("ip", "addr", "show")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var interfaces []string
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, ": ") && !strings.Contains(line, "lo:") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				interfaceName := strings.TrimSpace(parts[1])
				interfaces = append(interfaces, interfaceName)
			}
		}
	}

	return interfaces, nil
}

// saveServiceStatus 保存服务状态到数据库
func (s *SambaMonitoringService) saveServiceStatus(service *dto.SambaServiceResponse) {
	dbService := models.SambaService{
		Base:            models.Base{ID: service.ID},
		ServiceName:     service.ServiceName,
		ProcessID:       service.ProcessID,
		Status:          service.Status,
		StartTime:       service.StartTime,
		LastCheckTime:   service.LastCheckTime,
		MemoryUsage:     service.MemoryUsage,
		CPUUsage:        service.CPUUsage,
		ConnectionCount: service.ConnectionCount,
	}

	// 使用UPSERT操作
	s.db.Save(&dbService)
}

// saveConnection 保存连接信息到数据库
func (s *SambaMonitoringService) saveConnection(conn *dto.SambaConnectionResponse) {
	dbConn := models.SambaConnection{
		Base:           models.Base{ID: conn.ID},
		ClientIP:       conn.ClientIP,
		ClientHostname: conn.ClientHostname,
		Username:       conn.Username,
		ShareName:      conn.ShareName,
		ConnectedAt:    conn.ConnectedAt,
		LastActivity:   conn.LastActivity,
		FilesOpen:      conn.FilesOpen,
		BytesRead:      conn.BytesRead,
		BytesWritten:   conn.BytesWritten,
	}

	s.db.Save(&dbConn)
}

// toConnectionResponse 转换连接响应格式
func (s *SambaMonitoringService) toConnectionResponse(conn *models.SambaConnection) *dto.SambaConnectionResponse {
	return &dto.SambaConnectionResponse{
		ID:             conn.ID,
		ClientIP:       conn.ClientIP,
		ClientHostname: conn.ClientHostname,
		Username:       conn.Username,
		ShareName:      conn.ShareName,
		ConnectedAt:    conn.ConnectedAt,
		LastActivity:   conn.LastActivity,
		FilesOpen:      conn.FilesOpen,
		BytesRead:      conn.BytesRead,
		BytesWritten:   conn.BytesWritten,
		CreatedAt:      conn.CreatedAt,
		UpdatedAt:      conn.UpdatedAt,
	}
}

// toAuditLogResponse 转换审计日志响应格式
func (s *SambaMonitoringService) toAuditLogResponse(log *models.SambaAuditLog) *dto.SambaAuditLogResponse {
	return &dto.SambaAuditLogResponse{
		ID:               log.ID,
		Username:         log.Username,
		ClientIP:         log.ClientIP,
		ShareName:        log.ShareName,
		Operation:        log.Operation,
		FilePath:         log.FilePath,
		Success:          log.Success,
		ErrorMessage:     log.ErrorMessage,
		Timestamp:        log.Timestamp,
		BytesTransferred: log.BytesTransferred,
		CreatedAt:        log.CreatedAt,
		UpdatedAt:        log.UpdatedAt,
	}
}

// ServiceStatus 服务状态结构
type ServiceStatus struct {
	ServiceName string
	ProcessID   int
	Status      string
	StartTime   *time.Time
	MemoryUsage int64
	CPUUsage    float64
}