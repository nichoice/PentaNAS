package controllers

import (
	"net/http"
	"pnas/internal/database"
	"pnas/internal/models"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// AuditQueryParams 审计查询参数
type AuditQueryParams struct {
	UserID      string `form:"user_id" json:"user_id"`
	Username    string `form:"username" json:"username"`
	Operation   string `form:"operation" json:"operation"`
	FilePath    string `form:"file_path" json:"file_path"`
	Status      string `form:"status" json:"status"`
	Source      string `form:"source" json:"source"`
	StartTime   string `form:"start_time" json:"start_time"`
	EndTime     string `form:"end_time" json:"end_time"`
	Page        int    `form:"page" json:"page"`
	PageSize    int    `form:"page_size" json:"page_size"`
	OrderBy     string `form:"order_by" json:"order_by"`
	OrderDir    string `form:"order_dir" json:"order_dir"`
}

// FileAuditLogResponse 文件审计日志响应结构（用于Swagger）
type FileAuditLogResponse struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Username  string    `json:"username"`
	Operation string    `json:"operation"`
	FilePath  string    `json:"file_path"`
	Success   bool      `json:"success"`
	Message   string    `json:"message"`
	Source    string    `json:"source"`
	ClientIP  string    `json:"client_ip"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// AuditStatsResponse 审计统计响应
type AuditStatsResponse struct {
	TotalCount       int64                    `json:"total_count"`
	SuccessCount     int64                    `json:"success_count"`
	FailedCount      int64                    `json:"failed_count"`
	OperationStats   []OperationStat          `json:"operation_stats"`
	UserStats        []UserStat               `json:"user_stats"`
	HourlyStats      []HourlyStats            `json:"hourly_stats"`
	DailyStats       []DailyStats             `json:"daily_stats"`
	TopFiles         []FileAccessStat         `json:"top_files"`
	RecentActivities []FileAuditLogResponse   `json:"recent_activities"`
}

type OperationStat struct {
	Operation string `json:"operation"`
	Count     int64  `json:"count"`
	Percentage float64 `json:"percentage"`
}

type UserStat struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Count    int64  `json:"count"`
	LastActivity time.Time `json:"last_activity"`
}

type HourlyStats struct {
	Hour  int   `json:"hour"`
	Count int64 `json:"count"`
}

type DailyStats struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

type FileAccessStat struct {
	FilePath    string    `json:"file_path"`
	FileName    string    `json:"file_name"`
	AccessCount int64     `json:"access_count"`
	UserCount   int64     `json:"user_count"`
	LastAccess  time.Time `json:"last_access"`
	FileSize    int64     `json:"file_size"`
}

type AnomalyDetectionResponse struct {
	SuspiciousActivities []SuspiciousActivity `json:"suspicious_activities"`
	Alerts              []SecurityAlert      `json:"alerts"`
	RiskScore           float64              `json:"risk_score"`
}

type SuspiciousActivity struct {
	UserID      string    `json:"user_id"`
	Username    string    `json:"username"`
	Operation   string    `json:"operation"`
	Count       int64     `json:"count"`
	TimeRange   string    `json:"time_range"`
	RiskLevel   string    `json:"risk_level"`
	Description string    `json:"description"`
	FirstTime   time.Time `json:"first_time"`
	LastTime    time.Time `json:"last_time"`
}

type SecurityAlert struct {
	Type        string    `json:"type"`
	Description string    `json:"description"`
	Severity    string    `json:"severity"`
	Count       int64     `json:"count"`
	FirstTime   time.Time `json:"first_time"`
	LastTime    time.Time `json:"last_time"`
}

// GetAuditLogs 获取审计日志
// @Summary Get audit logs
// @Description Get audit logs with filtering and pagination
// @Security BearerAuth
// @Tags Audit
// @Accept json
// @Produce json
// @Param user_id query string false "User ID"
// @Param username query string false "Username"
// @Param operation query string false "Operation type"
// @Param file_path query string false "File path"
// @Param status query string false "Status"
// @Param source query string false "Source"
// @Param start_time query string false "Start time (RFC3339)"
// @Param end_time query string false "End time (RFC3339)"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Param order_by query string false "Order by field" default(created_at)
// @Param order_dir query string false "Order direction" default(desc)
// @Success 200 {object} map[string]interface{} "Audit logs"
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /audit/logs [get]
func GetAuditLogs(c *gin.Context) {
	var params AuditQueryParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 设置默认值
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 || params.PageSize > 1000 {
		params.PageSize = 20
	}
	if params.OrderBy == "" {
		params.OrderBy = "created_at"
	}
	if params.OrderDir == "" {
		params.OrderDir = "desc"
	}

	// 构建查询
	query := database.DB.Model(&models.FileAuditLog{})

	// 应用过滤条件
	if params.UserID != "" {
		query = query.Where("user_id = ?", params.UserID)
	}
	if params.Username != "" {
		query = query.Where("username LIKE ?", "%"+params.Username+"%")
	}
	if params.Operation != "" {
		query = query.Where("operation = ?", params.Operation)
	}
	if params.FilePath != "" {
		query = query.Where("file_path LIKE ?", "%"+params.FilePath+"%")
	}
	if params.Status != "" {
		query = query.Where("status = ?", params.Status)
	}
	if params.Source != "" {
		query = query.Where("source = ?", params.Source)
	}
	if params.StartTime != "" {
		if startTime, err := time.Parse(time.RFC3339, params.StartTime); err == nil {
			query = query.Where("created_at >= ?", startTime)
		}
	}
	if params.EndTime != "" {
		if endTime, err := time.Parse(time.RFC3339, params.EndTime); err == nil {
			query = query.Where("created_at <= ?", endTime)
		}
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 应用排序和分页
	offset := (params.Page - 1) * params.PageSize
	orderClause := params.OrderBy + " " + strings.ToUpper(params.OrderDir)

	var logs []models.FileAuditLog
	if err := query.Order(orderClause).Offset(offset).Limit(params.PageSize).Find(&logs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"logs":       logs,
		"total":      total,
		"page":       params.Page,
		"page_size":  params.PageSize,
		"total_pages": (total + int64(params.PageSize) - 1) / int64(params.PageSize),
	})
}

// GetAuditStats 获取审计统计
// @Summary Get audit statistics
// @Description Get comprehensive audit statistics
// @Security BearerAuth
// @Tags Audit
// @Accept json
// @Produce json
// @Param start_time query string false "Start time (RFC3339)"
// @Param end_time query string false "End time (RFC3339)"
// @Success 200 {object} AuditStatsResponse "Audit statistics"
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /audit/stats [get]
func GetAuditStats(c *gin.Context) {
	startTime := c.Query("start_time")
	endTime := c.Query("end_time")

	// 默认时间范围：最近30天
	var start, end time.Time
	if startTime != "" {
		if t, err := time.Parse(time.RFC3339, startTime); err == nil {
			start = t
		} else {
			start = time.Now().AddDate(0, 0, -30)
		}
	} else {
		start = time.Now().AddDate(0, 0, -30)
	}

	if endTime != "" {
		if t, err := time.Parse(time.RFC3339, endTime); err == nil {
			end = t
		} else {
			end = time.Now()
		}
	} else {
		end = time.Now()
	}

	response := AuditStatsResponse{}

	// 基本统计
	var totalCount, successCount, failedCount int64
	database.DB.Model(&models.FileAuditLog{}).
		Where("created_at BETWEEN ? AND ?", start, end).
		Count(&totalCount)

	database.DB.Model(&models.FileAuditLog{}).
		Where("created_at BETWEEN ? AND ? AND status = ?", start, end, "success").
		Count(&successCount)

	database.DB.Model(&models.FileAuditLog{}).
		Where("created_at BETWEEN ? AND ? AND status = ?", start, end, "failed").
		Count(&failedCount)

	response.TotalCount = totalCount
	response.SuccessCount = successCount
	response.FailedCount = failedCount

	// 操作统计
	var operationStats []OperationStat
	database.DB.Model(&models.FileAuditLog{}).
		Select("operation, COUNT(*) as count").
		Where("created_at BETWEEN ? AND ?", start, end).
		Group("operation").
		Order("count DESC").
		Scan(&operationStats)

	for i := range operationStats {
		if totalCount > 0 {
			operationStats[i].Percentage = float64(operationStats[i].Count) / float64(totalCount) * 100
		}
	}
	response.OperationStats = operationStats

	// 用户统计
	var userStats []UserStat
	database.DB.Model(&models.FileAuditLog{}).
		Select("user_id, username, COUNT(*) as count, MAX(created_at) as last_activity").
		Where("created_at BETWEEN ? AND ?", start, end).
		Group("user_id, username").
		Order("count DESC").
		Limit(10).
		Scan(&userStats)
	response.UserStats = userStats

	// 小时统计
	var hourlyStats []HourlyStats
	database.DB.Raw(`
		SELECT
			CAST(strftime('%H', created_at) AS INTEGER) as hour,
			COUNT(*) as count
		FROM file_audit_logs
		WHERE created_at BETWEEN ? AND ?
		GROUP BY hour
		ORDER BY hour
	`, start, end).Scan(&hourlyStats)
	response.HourlyStats = hourlyStats

	// 日统计
	var dailyStats []DailyStats
	database.DB.Raw(`
		SELECT
			DATE(created_at) as date,
			COUNT(*) as count
		FROM file_audit_logs
		WHERE created_at BETWEEN ? AND ?
		GROUP BY DATE(created_at)
		ORDER BY date DESC
		LIMIT 30
	`, start, end).Scan(&dailyStats)
	response.DailyStats = dailyStats

	// 热门文件
	var topFiles []FileAccessStat
	database.DB.Model(&models.FileAccessStats{}).
		Select("file_path, file_name, access_count, user_count, last_access").
		Order("access_count DESC").
		Limit(10).
		Scan(&topFiles)

	// 补充文件大小信息
	for i := range topFiles {
		var log models.FileAuditLog
		database.DB.Model(&models.FileAuditLog{}).
			Select("file_size").
			Where("file_path = ?", topFiles[i].FilePath).
			Order("created_at DESC").
			First(&log)
		topFiles[i].FileSize = log.FileSize
	}
	response.TopFiles = topFiles

	// 最近活动
	var recentActivities []models.FileAuditLog
	database.DB.Model(&models.FileAuditLog{}).
		Where("created_at BETWEEN ? AND ?", start, end).
		Order("created_at DESC").
		Limit(20).
		Find(&recentActivities)

	// 转换为响应结构体
	recentActivitiesResponse := make([]FileAuditLogResponse, len(recentActivities))
	for i, activity := range recentActivities {
		recentActivitiesResponse[i] = FileAuditLogResponse{
			ID:        activity.ID,
			UserID:    activity.UserID,
			Username:  activity.Username,
			Operation: activity.Operation,
			FilePath:  activity.FilePath,
			Success:   activity.Status == "success",
			Message:   activity.ErrorMsg,
			Source:    activity.Source,
			ClientIP:  activity.ClientIP,
			CreatedAt: activity.CreatedAt,
			UpdatedAt: activity.UpdatedAt,
		}
	}
	response.RecentActivities = recentActivitiesResponse

	c.JSON(http.StatusOK, response)
}

// GetFileAccessHeatmap 获取文件访问热度图
// @Summary Get file access heatmap
// @Description Get file access heatmap data
// @Security BearerAuth
// @Tags Audit
// @Accept json
// @Produce json
// @Param limit query int false "Limit" default(50)
// @Success 200 {object} map[string]interface{} "File access heatmap"
// @Failure 500 {object} map[string]string
// @Router /audit/heatmap [get]
func GetFileAccessHeatmap(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "50")
	limit, _ := strconv.Atoi(limitStr)
	if limit <= 0 || limit > 1000 {
		limit = 50
	}

	var heatmapData []map[string]interface{}

	result := database.DB.Raw(`
		SELECT
			f.file_path,
			f.file_name,
			f.access_count,
			f.user_count,
			f.last_access,
			COALESCE(l.file_size, 0) as file_size,
			CASE
				WHEN f.access_count > 100 THEN 'very_hot'
				WHEN f.access_count > 50 THEN 'hot'
				WHEN f.access_count > 20 THEN 'warm'
				WHEN f.access_count > 5 THEN 'cool'
				ELSE 'cold'
			END as heat_level
		FROM file_access_stats f
		LEFT JOIN (
			SELECT file_path, file_size,
				ROW_NUMBER() OVER (PARTITION BY file_path ORDER BY created_at DESC) as rn
			FROM file_audit_logs
		) l ON f.file_path = l.file_path AND l.rn = 1
		ORDER BY f.access_count DESC
		LIMIT ?
	`, limit).Scan(&heatmapData)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"heatmap": heatmapData,
		"total":   len(heatmapData),
	})
}

// DetectAnomalies 检测异常行为
// @Summary Detect anomalies
// @Description Detect suspicious activities and security anomalies
// @Security BearerAuth
// @Tags Audit
// @Accept json
// @Produce json
// @Param hours query int false "Time range in hours" default(24)
// @Success 200 {object} AnomalyDetectionResponse "Anomaly detection results"
// @Failure 500 {object} map[string]string
// @Router /audit/anomalies [get]
func DetectAnomalies(c *gin.Context) {
	hoursStr := c.DefaultQuery("hours", "24")
	hours, _ := strconv.Atoi(hoursStr)
	if hours <= 0 {
		hours = 24
	}

	timeThreshold := time.Now().Add(-time.Duration(hours) * time.Hour)

	response := AnomalyDetectionResponse{
		SuspiciousActivities: []SuspiciousActivity{},
		Alerts:              []SecurityAlert{},
		RiskScore:           0,
	}

	// 检测频繁失败的操作
	var failedActivities []SuspiciousActivity
	database.DB.Raw(`
		SELECT
			user_id,
			username,
			operation,
			COUNT(*) as count,
			MIN(created_at) as first_time,
			MAX(created_at) as last_time
		FROM file_audit_logs
		WHERE status = 'failed' AND created_at > ?
		GROUP BY user_id, username, operation
		HAVING count > 10
		ORDER BY count DESC
	`, timeThreshold).Scan(&failedActivities)

	for i := range failedActivities {
		failedActivities[i].RiskLevel = "high"
		failedActivities[i].Description = "频繁失败的操作，可能存在恶意攻击"
		failedActivities[i].TimeRange = "最近" + strconv.Itoa(hours) + "小时"
	}
	response.SuspiciousActivities = append(response.SuspiciousActivities, failedActivities...)

	// 检测异常大量的操作
	var bulkActivities []SuspiciousActivity
	database.DB.Raw(`
		SELECT
			user_id,
			username,
			operation,
			COUNT(*) as count,
			MIN(created_at) as first_time,
			MAX(created_at) as last_time
		FROM file_audit_logs
		WHERE created_at > ?
		GROUP BY user_id, username, operation
		HAVING count > 100
		ORDER BY count DESC
	`, timeThreshold).Scan(&bulkActivities)

	for i := range bulkActivities {
		bulkActivities[i].RiskLevel = "medium"
		bulkActivities[i].Description = "异常大量的操作，可能存在批量操作"
		bulkActivities[i].TimeRange = "最近" + strconv.Itoa(hours) + "小时"
	}
	response.SuspiciousActivities = append(response.SuspiciousActivities, bulkActivities...)

	// 生成安全警报
	var alerts []SecurityAlert

	// 检测暴力破解
	var bruteForceCount int64
	database.DB.Model(&models.FileAuditLog{}).
		Where("status = 'failed' AND operation = 'read' AND created_at > ?", timeThreshold).
		Count(&bruteForceCount)

	if bruteForceCount > 50 {
		alerts = append(alerts, SecurityAlert{
			Type:        "brute_force",
			Description: "检测到可能的暴力破解攻击",
			Severity:    "high",
			Count:       bruteForceCount,
			FirstTime:   timeThreshold,
			LastTime:    time.Now(),
		})
	}

	// 检测数据泄露
	var downloadCount int64
	database.DB.Model(&models.FileAuditLog{}).
		Where("operation = 'download' AND created_at > ?", timeThreshold).
		Count(&downloadCount)

	if downloadCount > 200 {
		alerts = append(alerts, SecurityAlert{
			Type:        "data_exfiltration",
			Description: "检测到异常大量的下载活动",
			Severity:    "medium",
			Count:       downloadCount,
			FirstTime:   timeThreshold,
			LastTime:    time.Now(),
		})
	}

	// 检测删除活动
	var deleteCount int64
	database.DB.Model(&models.FileAuditLog{}).
		Where("operation = 'delete' AND created_at > ?", timeThreshold).
		Count(&deleteCount)

	if deleteCount > 50 {
		alerts = append(alerts, SecurityAlert{
			Type:        "mass_deletion",
			Description: "检测到异常大量的删除活动",
			Severity:    "high",
			Count:       deleteCount,
			FirstTime:   timeThreshold,
			LastTime:    time.Now(),
		})
	}

	response.Alerts = alerts

	// 计算风险评分
	riskScore := 0.0
	for _, activity := range response.SuspiciousActivities {
		switch activity.RiskLevel {
		case "high":
			riskScore += 30
		case "medium":
			riskScore += 15
		case "low":
			riskScore += 5
		}
	}
	for _, alert := range response.Alerts {
		switch alert.Severity {
		case "high":
			riskScore += 40
		case "medium":
			riskScore += 20
		case "low":
			riskScore += 10
		}
	}

	if riskScore > 100 {
		riskScore = 100
	}
	response.RiskScore = riskScore

	c.JSON(http.StatusOK, response)
}

// GetUserActivityTimeline 获取用户活动时间线
// @Summary Get user activity timeline
// @Description Get detailed activity timeline for a specific user
// @Security BearerAuth
// @Tags Audit
// @Accept json
// @Produce json
// @Param user_id path string true "User ID"
// @Param start_time query string false "Start time (RFC3339)"
// @Param end_time query string false "End time (RFC3339)"
// @Success 200 {object} map[string]interface{} "User activity timeline"
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /audit/users/{user_id}/timeline [get]
func GetUserActivityTimeline(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户ID不能为空"})
		return
	}

	startTime := c.Query("start_time")
	endTime := c.Query("end_time")

	// 默认时间范围：最近7天
	var start, end time.Time
	if startTime != "" {
		if t, err := time.Parse(time.RFC3339, startTime); err == nil {
			start = t
		} else {
			start = time.Now().AddDate(0, 0, -7)
		}
	} else {
		start = time.Now().AddDate(0, 0, -7)
	}

	if endTime != "" {
		if t, err := time.Parse(time.RFC3339, endTime); err == nil {
			end = t
		} else {
			end = time.Now()
		}
	} else {
		end = time.Now()
	}

	var activities []models.FileAuditLog
	if err := database.DB.Model(&models.FileAuditLog{}).
		Where("user_id = ? AND created_at BETWEEN ? AND ?", userID, start, end).
		Order("created_at DESC").
		Find(&activities).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 统计信息
	var stats map[string]interface{}
	database.DB.Raw(`
		SELECT
			COUNT(*) as total_operations,
			COUNT(CASE WHEN status = 'success' THEN 1 END) as success_count,
			COUNT(CASE WHEN status = 'failed' THEN 1 END) as failed_count,
			COUNT(DISTINCT operation) as operation_types,
			COUNT(DISTINCT DATE(created_at)) as active_days
		FROM file_audit_logs
		WHERE user_id = ? AND created_at BETWEEN ? AND ?
	`, userID, start, end).Scan(&stats)

	c.JSON(http.StatusOK, gin.H{
		"activities": activities,
		"stats":      stats,
		"time_range": gin.H{
			"start": start,
			"end":   end,
		},
	})
}