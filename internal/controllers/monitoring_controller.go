package controllers

import (
	"fmt"
	"net/http"
	"pnas/internal/app/dto"
	"pnas/internal/services"

	"github.com/gin-gonic/gin"
)

var (
	monitoringService *services.MonitoringService
	websocketHub      *services.WebSocketHub
)

// InitMonitoringServices initializes monitoring related services
func InitMonitoringServices() {
	monitoringService = services.NewMonitoringService()
	websocketHub = services.NewWebSocketHub(monitoringService)
	websocketHub.Start()
}

// GetSystemInfo retrieves system information
// @Summary Get system information
// @Description Get detailed system information including OS, kernel, hardware details
// @Security BearerAuth
// @Tags Monitoring
// @Accept json
// @Produce json
// @Success 200 {object} dto.SystemInfo "System information"
// @Failure 500 {object} map[string]string
// @Router /monitoring/system [get]
func GetSystemInfo(c *gin.Context) {
	if monitoringService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "monitoring service not initialized"})
		return
	}

	info, err := monitoringService.GetSystemInfo()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, info)
}

// GetCPUInfo retrieves CPU information and usage
// @Summary Get CPU information
// @Description Get detailed CPU information including usage, load average, and per-CPU statistics
// @Security BearerAuth
// @Tags Monitoring
// @Accept json
// @Produce json
// @Success 200 {object} dto.CPUInfo "CPU information"
// @Failure 500 {object} map[string]string
// @Router /monitoring/cpu [get]
func GetCPUInfo(c *gin.Context) {
	if monitoringService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "monitoring service not initialized"})
		return
	}

	info, err := monitoringService.GetCPUInfo()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, info)
}

// GetMemoryInfo retrieves memory information and usage
// @Summary Get memory information
// @Description Get detailed memory information including RAM and swap usage
// @Security BearerAuth
// @Tags Monitoring
// @Accept json
// @Produce json
// @Success 200 {object} dto.MemoryInfo "Memory information"
// @Failure 500 {object} map[string]string
// @Router /monitoring/memory [get]
func GetMemoryInfo(c *gin.Context) {
	if monitoringService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "monitoring service not initialized"})
		return
	}

	info, err := monitoringService.GetMemoryInfo()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, info)
}

// GetDiskInfo retrieves disk information and usage
// @Summary Get disk information
// @Description Get detailed disk information including usage and I/O statistics
// @Security BearerAuth
// @Tags Monitoring
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "Disk information including usage and I/O"
// @Failure 500 {object} map[string]string
// @Router /monitoring/disk [get]
func GetDiskInfo(c *gin.Context) {
	if monitoringService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "monitoring service not initialized"})
		return
	}

	diskInfo, err := monitoringService.GetDiskInfo()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	diskIOInfo, err := monitoringService.GetDiskIOInfo()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := map[string]interface{}{
		"disks":   diskInfo,
		"disk_io": diskIOInfo,
	}

	c.JSON(http.StatusOK, response)
}

// GetNetworkInfo retrieves network information and statistics
// @Summary Get network information
// @Description Get detailed network information including interfaces and I/O statistics
// @Security BearerAuth
// @Tags Monitoring
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "Network information including interfaces and I/O"
// @Failure 500 {object} map[string]string
// @Router /monitoring/network [get]
func GetNetworkInfo(c *gin.Context) {
	if monitoringService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "monitoring service not initialized"})
		return
	}

	networkInfo, err := monitoringService.GetNetworkInfo()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	networkIOInfo, err := monitoringService.GetNetworkIOInfo()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := map[string]interface{}{
		"networks":   networkInfo,
		"network_io": networkIOInfo,
	}

	c.JSON(http.StatusOK, response)
}

// GetStorageProtocolInfo retrieves storage protocol monitoring information
// @Summary Get storage protocol information
// @Description Get monitoring information for storage protocols (SMB, NFS, iSCSI)
// @Security BearerAuth
// @Tags Monitoring
// @Accept json
// @Produce json
// @Success 200 {object} []dto.StorageProtocolInfo "Storage protocol information"
// @Failure 500 {object} map[string]string
// @Router /monitoring/storage-protocols [get]
func GetStorageProtocolInfo(c *gin.Context) {
	if monitoringService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "monitoring service not initialized"})
		return
	}

	info, err := monitoringService.GetStorageProtocolInfo()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, info)
}

// GetCompleteMonitoringData retrieves all monitoring data
// @Summary Get complete monitoring data
// @Description Get all available monitoring data in a single response
// @Security BearerAuth
// @Tags Monitoring
// @Accept json
// @Produce json
// @Success 200 {object} dto.MonitoringData "Complete monitoring data"
// @Failure 500 {object} map[string]string
// @Router /monitoring/complete [get]
func GetCompleteMonitoringData(c *gin.Context) {
	if monitoringService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "monitoring service not initialized"})
		return
	}

	data, err := monitoringService.GetCompleteMonitoringData()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, data)
}

// GetWebSocketInfo retrieves WebSocket connection information
// @Summary Get WebSocket information
// @Description Get information about current WebSocket connections
// @Security BearerAuth
// @Tags Monitoring
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "WebSocket connection information"
// @Failure 500 {object} map[string]string
// @Router /monitoring/websocket/info [get]
func GetWebSocketInfo(c *gin.Context) {
	if websocketHub == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "websocket hub not initialized"})
		return
	}

	info := map[string]interface{}{
		"client_count": websocketHub.GetClientCount(),
		"clients":      websocketHub.GetConnectedClients(),
	}

	c.JSON(http.StatusOK, info)
}

// HandleWebSocket handles WebSocket connections
// @Summary WebSocket endpoint for real-time monitoring
// @Description Establish WebSocket connection for real-time monitoring data
// @Tags Monitoring
// @Accept json
// @Produce json
// @Success 101 {string} string "Switching Protocols"
// @Router /monitoring/websocket [get]
func HandleWebSocket(c *gin.Context) {
	if websocketHub == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "websocket hub not initialized"})
		return
	}

	websocketHub.HandleWebSocket(c)
}

// BroadcastMessage broadcasts a message to all WebSocket clients
// @Summary Broadcast message to WebSocket clients
// @Description Send a message to all connected WebSocket clients
// @Security BearerAuth
// @Tags Monitoring
// @Accept json
// @Produce json
// @Param message body dto.WebSocketMessage true "Message to broadcast"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /monitoring/websocket/broadcast [post]
func BroadcastMessage(c *gin.Context) {
	if websocketHub == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "websocket hub not initialized"})
		return
	}

	var message dto.WebSocketMessage
	if err := c.ShouldBindJSON(&message); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	websocketHub.BroadcastMessage(message)
	c.JSON(http.StatusOK, gin.H{"message": "broadcast sent successfully"})
}

// SendToClient sends a message to a specific WebSocket client
// @Summary Send message to specific WebSocket client
// @Description Send a message to a specific connected WebSocket client
// @Security BearerAuth
// @Tags Monitoring
// @Accept json
// @Produce json
// @Param client_id path string true "Client ID"
// @Param message body dto.WebSocketMessage true "Message to send"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /monitoring/websocket/client/{client_id} [post]
func SendToClient(c *gin.Context) {
	if websocketHub == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "websocket hub not initialized"})
		return
	}

	clientID := c.Param("client_id")
	if clientID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "client_id is required"})
		return
	}

	var message dto.WebSocketMessage
	if err := c.ShouldBindJSON(&message); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := websocketHub.SendToClient(clientID, message)
	if err != nil {
		if err.Error() == "client "+clientID+" not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "message sent successfully"})
}

// GetServiceStatus retrieves status of specific services
// @Summary Get service status
// @Description Get status information for specific system services
// @Security BearerAuth
// @Tags Monitoring
// @Accept json
// @Produce json
// @Param service query string true "Service name to check"
// @Success 200 {object} map[string]interface{} "Service status information"
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /monitoring/service/status [get]
func GetServiceStatus(c *gin.Context) {
	if monitoringService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "monitoring service not initialized"})
		return
	}

	serviceName := c.Query("service")
	if serviceName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "service parameter is required"})
		return
	}

	status := monitoringService.GetServiceStatus(serviceName)
	c.JSON(http.StatusOK, status)
}

// GetMetrics retrieves system metrics in Prometheus format
// @Summary Get system metrics
// @Description Get system metrics in Prometheus format for monitoring integration
// @Security BearerAuth
// @Tags Monitoring
// @Accept json
// @Produce text/plain
// @Success 200 {string} string "Prometheus metrics format"
// @Failure 500 {object} map[string]string
// @Router /monitoring/metrics [get]
func GetMetrics(c *gin.Context) {
	if monitoringService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "monitoring service not initialized"})
		return
	}

	// This would integrate with Prometheus client library
	// For now, return a simple response
	metrics := generatePrometheusMetrics()
	c.Data(http.StatusOK, "text/plain; charset=utf-8", []byte(metrics))
}

// generatePrometheusMetrics generates Prometheus-compatible metrics
func generatePrometheusMetrics() string {
	if monitoringService == nil {
		return "# HELP pnas_up PNAS monitoring service status\n# TYPE pnas_up gauge\npnas_up 0\n"
	}

	// Get current monitoring data
	data, err := monitoringService.GetCompleteMonitoringData()
	if err != nil {
		return "# HELP pnas_up PNAS monitoring service status\n# TYPE pnas_up gauge\npnas_up 0\n"
	}

	metrics := "# HELP pnas_up PNAS monitoring service status\n# TYPE pnas_up gauge\npnas_up 1\n\n"

	// CPU metrics
	if len(data.CPU.Usage) > 0 {
		metrics += "# HELP pnas_cpu_usage_percent CPU usage percentage\n# TYPE pnas_cpu_usage_percent gauge\n"
		for i, usage := range data.CPU.Usage {
			metrics += fmt.Sprintf("pnas_cpu_usage_percent{cpu=\"%d\"} %.2f\n", i, usage)
		}
		metrics += "\n"
	}

	// Memory metrics
	metrics += "# HELP pnas_memory_total_bytes Total memory in bytes\n# TYPE pnas_memory_total_bytes gauge\n"
	metrics += fmt.Sprintf("pnas_memory_total_bytes %.0f\n", float64(data.Memory.Total))
	metrics += "# HELP pnas_memory_used_bytes Used memory in bytes\n# TYPE pnas_memory_used_bytes gauge\n"
	metrics += fmt.Sprintf("pnas_memory_used_bytes %.0f\n", float64(data.Memory.Used))
	metrics += "# HELP pnas_memory_usage_percent Memory usage percentage\n# TYPE pnas_memory_usage_percent gauge\n"
	metrics += fmt.Sprintf("pnas_memory_usage_percent %.2f\n\n", data.Memory.UsedPercent)

	// Disk metrics
	if len(data.Disks) > 0 {
		metrics += "# HELP pnas_disk_total_bytes Total disk space in bytes\n# TYPE pnas_disk_total_bytes gauge\n"
		metrics += "# HELP pnas_disk_used_bytes Used disk space in bytes\n# TYPE pnas_disk_used_bytes gauge\n"
		metrics += "# HELP pnas_disk_usage_percent Disk usage percentage\n# TYPE pnas_disk_usage_percent gauge\n"

		for _, disk := range data.Disks {
			device := disk.Device
			if device == "" {
				device = disk.Mountpoint
			}
			metrics += fmt.Sprintf("pnas_disk_total_bytes{device=\"%s\",mountpoint=\"%s\"} %.0f\n", device, disk.Mountpoint, float64(disk.Total))
			metrics += fmt.Sprintf("pnas_disk_used_bytes{device=\"%s\",mountpoint=\"%s\"} %.0f\n", device, disk.Mountpoint, float64(disk.Used))
			metrics += fmt.Sprintf("pnas_disk_usage_percent{device=\"%s\",mountpoint=\"%s\"} %.2f\n", device, disk.Mountpoint, disk.UsedPercent)
		}
		metrics += "\n"
	}

	// Network metrics
	if len(data.NetworkIO) > 0 {
		metrics += "# HELP pnas_network_bytes_sent_total Total bytes sent\n# TYPE pnas_network_bytes_sent_total counter\n"
		metrics += "# HELP pnas_network_bytes_recv_total Total bytes received\n# TYPE pnas_network_bytes_recv_total counter\n"

		for _, netIO := range data.NetworkIO {
			metrics += fmt.Sprintf("pnas_network_bytes_sent_total{interface=\"%s\"} %d\n", netIO.Name, netIO.BytesSent)
			metrics += fmt.Sprintf("pnas_network_bytes_recv_total{interface=\"%s\"} %d\n", netIO.Name, netIO.BytesRecv)
		}
		metrics += "\n"
	}

	// Storage protocol metrics
	if len(data.StorageProtocols) > 0 {
		metrics += "# HELP pnas_storage_protocol_connections Active connections for storage protocols\n# TYPE pnas_storage_protocol_connections gauge\n"

		for _, protocol := range data.StorageProtocols {
			status := 0
			if protocol.Status == "running" {
				status = 1
			}
			metrics += fmt.Sprintf("pnas_storage_protocol_connections{protocol=\"%s\"} %d\n", protocol.Protocol, protocol.Connections)
			metrics += fmt.Sprintf("pnas_storage_protocol_status{protocol=\"%s\"} %d\n", protocol.Protocol, status)
		}
		metrics += "\n"
	}

	// System uptime
	metrics += "# HELP pnas_system_uptime_seconds System uptime in seconds\n# TYPE pnas_system_uptime_seconds gauge\n"
	metrics += fmt.Sprintf("pnas_system_uptime_seconds %d\n", data.SystemInfo.Uptime)

	return metrics
}

// CleanupMonitoringServices cleans up monitoring services
func CleanupMonitoringServices() {
	if websocketHub != nil {
		websocketHub.Stop()
	}
	if monitoringService != nil {
		monitoringService.Close()
	}
}