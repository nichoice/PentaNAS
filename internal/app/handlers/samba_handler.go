package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"pnas/internal/app/dto"
	"pnas/internal/app/services"
)

type SambaHandler struct {
	accountService     *services.SambaAccountService
	shareService       *services.SambaShareService
	configService      *services.SambaConfigService
	timeMachineService *services.SambaTimeMachineService
	recycleService     *services.SambaRecycleService
	monitoringService  *services.SambaMonitoringService
}

func NewSambaHandler(db *gorm.DB) *SambaHandler {
	shareService := services.NewSambaShareService(db)
	return &SambaHandler{
		accountService:     services.NewSambaAccountService(db),
		shareService:       shareService,
		configService:      services.NewSambaConfigService(db),
		timeMachineService: services.NewSambaTimeMachineService(db, shareService),
		recycleService:     services.NewSambaRecycleService(db),
		monitoringService:  services.NewSambaMonitoringService(db),
	}
}

// Samba账号管理

// CreateAccount 创建Samba账号
func (h *SambaHandler) CreateAccount(c *gin.Context) {
	var req dto.CreateSambaAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	account, err := h.accountService.CreateAccount(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": account})
}

// UpdateAccount 更新Samba账号
func (h *SambaHandler) UpdateAccount(c *gin.Context) {
	accountID := c.Param("id")
	var req dto.UpdateSambaAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	account, err := h.accountService.UpdateAccount(accountID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": account})
}

// GetAccount 获取Samba账号
func (h *SambaHandler) GetAccount(c *gin.Context) {
	accountID := c.Param("id")

	account, err := h.accountService.GetAccount(accountID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": account})
}

// ListAccounts 获取Samba账号列表
func (h *SambaHandler) ListAccounts(c *gin.Context) {
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	role := c.Query("role")

	accounts, total, err := h.accountService.ListAccounts(offset, limit, role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  accounts,
		"total": total,
	})
}

// DeleteAccount 删除Samba账号
func (h *SambaHandler) DeleteAccount(c *gin.Context) {
	accountID := c.Param("id")

	if err := h.accountService.DeleteAccount(accountID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "账号删除成功"})
}

// Samba共享管理

// CreateShare 创建Samba共享
func (h *SambaHandler) CreateShare(c *gin.Context) {
	var req dto.CreateSambaShareRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	share, err := h.shareService.CreateShare(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": share})
}

// UpdateShare 更新Samba共享
func (h *SambaHandler) UpdateShare(c *gin.Context) {
	shareID := c.Param("id")
	var req dto.UpdateSambaShareRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	share, err := h.shareService.UpdateShare(shareID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": share})
}

// GetShare 获取Samba共享
func (h *SambaHandler) GetShare(c *gin.Context) {
	shareID := c.Param("id")

	share, err := h.shareService.GetShare(shareID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": share})
}

// ListShares 获取Samba共享列表
func (h *SambaHandler) ListShares(c *gin.Context) {
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	var enabled *bool
	if enabledStr := c.Query("enabled"); enabledStr != "" {
		if e, err := strconv.ParseBool(enabledStr); err == nil {
			enabled = &e
		}
	}

	shares, total, err := h.shareService.ListShares(offset, limit, enabled)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  shares,
		"total": total,
	})
}

// DeleteShare 删除Samba共享
func (h *SambaHandler) DeleteShare(c *gin.Context) {
	shareID := c.Param("id")

	if err := h.shareService.DeleteShare(shareID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "共享删除成功"})
}

// SetShareAccess 设置共享访问权限
func (h *SambaHandler) SetShareAccess(c *gin.Context) {
	var req dto.SetSambaShareAccessRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	access, err := h.shareService.SetShareAccess(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": access})
}

// RemoveShareAccess 移除共享访问权限
func (h *SambaHandler) RemoveShareAccess(c *gin.Context) {
	shareID := c.Param("shareId")
	accountID := c.Param("accountId")

	if err := h.shareService.RemoveShareAccess(shareID, accountID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "访问权限移除成功"})
}

// GetShareAccess 获取共享访问权限列表
func (h *SambaHandler) GetShareAccess(c *gin.Context) {
	shareID := c.Param("id")

	accessList, err := h.shareService.GetShareAccess(shareID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": accessList})
}

// Samba配置管理

// UpdateGlobalConfig 更新全局配置
func (h *SambaHandler) UpdateGlobalConfig(c *gin.Context) {
	var req dto.UpdateSambaGlobalConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	config, err := h.configService.UpdateGlobalConfig(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": config})
}

// GetGlobalConfig 获取全局配置
func (h *SambaHandler) GetGlobalConfig(c *gin.Context) {
	config, err := h.configService.GetGlobalConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": config})
}

// GenerateConfig 生成配置文件
func (h *SambaHandler) GenerateConfig(c *gin.Context) {
	configResp, err := h.configService.GenerateConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": configResp})
}

// WriteConfig 写入配置文件
func (h *SambaHandler) WriteConfig(c *gin.Context) {
	configPath := c.DefaultQuery("path", "/etc/samba/smb.conf")

	if err := h.configService.WriteConfig(configPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "配置文件写入成功"})
}

// ReloadConfig 重载配置
func (h *SambaHandler) ReloadConfig(c *gin.Context) {
	if err := h.configService.ReloadConfig(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "配置重载成功"})
}

// 时间机器功能

// EnableTimeMachine 启用时间机器
func (h *SambaHandler) EnableTimeMachine(c *gin.Context) {
	shareID := c.Param("id")
	quotaStr := c.DefaultQuery("quota", "0")
	quota, _ := strconv.ParseInt(quotaStr, 10, 64)

	if err := h.timeMachineService.EnableTimeMachine(shareID, quota); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "时间机器功能启用成功"})
}

// DisableTimeMachine 禁用时间机器
func (h *SambaHandler) DisableTimeMachine(c *gin.Context) {
	shareID := c.Param("id")

	if err := h.timeMachineService.DisableTimeMachine(shareID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "时间机器功能禁用成功"})
}

// GetTimeMachineShares 获取时间机器共享列表
func (h *SambaHandler) GetTimeMachineShares(c *gin.Context) {
	shares, err := h.timeMachineService.GetTimeMachineShares()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": shares})
}

// GetTimeMachineStatus 获取时间机器状态
func (h *SambaHandler) GetTimeMachineStatus(c *gin.Context) {
	shareID := c.Param("id")

	status, err := h.timeMachineService.GetTimeMachineStatus(shareID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": status})
}

// 回收站功能

// ListRecycleBinItems 获取回收站条目列表
func (h *SambaHandler) ListRecycleBinItems(c *gin.Context) {
	shareID := c.Query("share_id")
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	items, total, err := h.recycleService.ListRecycleBinItems(shareID, offset, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  items,
		"total": total,
	})
}

// RestoreRecycleBinItem 恢复回收站条目
func (h *SambaHandler) RestoreRecycleBinItem(c *gin.Context) {
	var req dto.RestoreRecycleBinItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.recycleService.RestoreRecycleBinItem(&req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "文件恢复成功"})
}

// PermanentDeleteRecycleBinItem 永久删除回收站条目
func (h *SambaHandler) PermanentDeleteRecycleBinItem(c *gin.Context) {
	itemID := c.Param("id")

	if err := h.recycleService.PermanentDeleteRecycleBinItem(itemID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "文件永久删除成功"})
}

// EmptyRecycleBin 清空回收站
func (h *SambaHandler) EmptyRecycleBin(c *gin.Context) {
	shareID := c.Param("shareId")

	if err := h.recycleService.EmptyRecycleBin(shareID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "回收站清空成功"})
}

// GetRecycleBinStatistics 获取回收站统计信息
func (h *SambaHandler) GetRecycleBinStatistics(c *gin.Context) {
	shareID := c.Query("share_id")

	statistics, err := h.recycleService.GetRecycleBinStatistics(shareID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": statistics})
}

// 监控功能

// GetSambaStatus 获取Samba整体状态
func (h *SambaHandler) GetSambaStatus(c *gin.Context) {
	status, err := h.monitoringService.GetSambaStatus()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": status})
}

// GetServiceStatus 获取服务状态
func (h *SambaHandler) GetServiceStatus(c *gin.Context) {
	services, err := h.monitoringService.GetServiceStatus()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": services})
}

// GetActiveConnections 获取活跃连接
func (h *SambaHandler) GetActiveConnections(c *gin.Context) {
	connections, err := h.monitoringService.GetActiveConnections()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": connections})
}

// GetConnectionHistory 获取连接历史
func (h *SambaHandler) GetConnectionHistory(c *gin.Context) {
	hours, _ := strconv.Atoi(c.DefaultQuery("hours", "24"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	connections, total, err := h.monitoringService.GetConnectionHistory(hours, offset, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  connections,
		"total": total,
	})
}

// GetAuditLogs 获取审计日志
func (h *SambaHandler) GetAuditLogs(c *gin.Context) {
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	// 构建过滤条件
	filters := make(map[string]interface{})
	if username := c.Query("username"); username != "" {
		filters["username"] = username
	}
	if clientIP := c.Query("client_ip"); clientIP != "" {
		filters["client_ip"] = clientIP
	}
	if shareName := c.Query("share_name"); shareName != "" {
		filters["share_name"] = shareName
	}
	if operation := c.Query("operation"); operation != "" {
		filters["operation"] = operation
	}
	if successStr := c.Query("success"); successStr != "" {
		if success, err := strconv.ParseBool(successStr); err == nil {
			filters["success"] = success
		}
	}

	// 时间范围过滤
	if startTimeStr := c.Query("start_time"); startTimeStr != "" {
		if startTime, err := time.Parse(time.RFC3339, startTimeStr); err == nil {
			filters["start_time"] = startTime
		}
	}
	if endTimeStr := c.Query("end_time"); endTimeStr != "" {
		if endTime, err := time.Parse(time.RFC3339, endTimeStr); err == nil {
			filters["end_time"] = endTime
		}
	}

	logs, total, err := h.monitoringService.GetAuditLogs(filters, offset, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  logs,
		"total": total,
	})
}

// EnableMultiChannel 启用多通道
func (h *SambaHandler) EnableMultiChannel(c *gin.Context) {
	shareID := c.Param("id")

	if err := h.monitoringService.EnableMultiChannel(shareID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "多通道功能启用成功"})
}

// DisableMultiChannel 禁用多通道
func (h *SambaHandler) DisableMultiChannel(c *gin.Context) {
	shareID := c.Param("id")

	if err := h.monitoringService.DisableMultiChannel(shareID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "多通道功能禁用成功"})
}

// GetMultiChannelStatus 获取多通道状态
func (h *SambaHandler) GetMultiChannelStatus(c *gin.Context) {
	status, err := h.monitoringService.GetMultiChannelStatus()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": status})
}