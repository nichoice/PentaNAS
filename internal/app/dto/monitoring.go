package dto

import "time"

// SystemInfo represents system information
type SystemInfo struct {
	Hostname        string    `json:"hostname"`
	Platform        string    `json:"platform"`
	PlatformFamily  string    `json:"platform_family"`
	PlatformVersion string    `json:"platform_version"`
	KernelVersion   string    `json:"kernel_version"`
	KernelArch      string    `json:"kernel_arch"`
	Uptime          uint64    `json:"uptime"`
	BootTime        time.Time `json:"boot_time"`
	Procs           uint64    `json:"procs"`
	OS              string    `json:"os"`
	HostID          string    `json:"host_id"`
}

// CPUInfo represents CPU information and usage
type CPUInfo struct {
	ModelName   string           `json:"model_name"`
	Cores       int32            `json:"cores"`
	Threads     int32            `json:"threads"`
	Family      string           `json:"family"`
	Speed       float64          `json:"speed_mhz"`
	CacheSize   int32            `json:"cache_size"`
	Usage       []float64        `json:"usage_percent"`
	Temperature float64          `json:"temperature"`
	LoadAvg     *LoadAvgInfo     `json:"load_avg"`
	PerCPU      []PerCPUInfo     `json:"per_cpu"`
}

// LoadAvgInfo represents system load average
type LoadAvgInfo struct {
	Load1  float64 `json:"load1"`
	Load5  float64 `json:"load5"`
	Load15 float64 `json:"load15"`
}

// PerCPUInfo represents per-CPU statistics
type PerCPUInfo struct {
	CPU    string  `json:"cpu"`
	User   float64 `json:"user"`
	System float64 `json:"system"`
	Idle   float64 `json:"idle"`
	Nice   float64 `json:"nice"`
	Iowait float64 `json:"iowait"`
	Irq    float64 `json:"irq"`
	Softirq float64 `json:"softirq"`
	Steal  float64 `json:"steal"`
	Guest  float64 `json:"guest"`
}

// MemoryInfo represents memory information and usage
type MemoryInfo struct {
	Total       uint64  `json:"total"`
	Available   uint64  `json:"available"`
	Used        uint64  `json:"used"`
	UsedPercent float64 `json:"used_percent"`
	Free        uint64  `json:"free"`
	Active      uint64  `json:"active"`
	Inactive    uint64  `json:"inactive"`
	Wired       uint64  `json:"wired"`
	Laundry     uint64  `json:"laundry"`
	Buffers     uint64  `json:"buffers"`
	Cached      uint64  `json:"cached"`
	Shared      uint64  `json:"shared"`
	Slab        uint64  `json:"slab"`
	SwapTotal   uint64  `json:"swap_total"`
	SwapUsed    uint64  `json:"swap_used"`
	SwapFree    uint64  `json:"swap_free"`
	SwapPercent float64 `json:"swap_percent"`
}

// DiskInfo represents disk information and usage
type DiskInfo struct {
	Device      string  `json:"device"`
	Mountpoint  string  `json:"mountpoint"`
	Fstype      string  `json:"fstype"`
	Total       uint64  `json:"total"`
	Free        uint64  `json:"free"`
	Used        uint64  `json:"used"`
	UsedPercent float64 `json:"used_percent"`
	InodesTotal uint64  `json:"inodes_total"`
	InodesUsed  uint64  `json:"inodes_used"`
	InodesFree  uint64  `json:"inodes_free"`
	InodesUsedPercent float64 `json:"inodes_used_percent"`
}

// DiskIOInfo represents disk I/O statistics
type DiskIOInfo struct {
	Name            string `json:"name"`
	ReadCount       uint64 `json:"read_count"`
	MergedReadCount uint64 `json:"merged_read_count"`
	WriteCount      uint64 `json:"write_count"`
	MergedWriteCount uint64 `json:"merged_write_count"`
	ReadBytes       uint64 `json:"read_bytes"`
	WriteBytes      uint64 `json:"write_bytes"`
	ReadTime        uint64 `json:"read_time"`
	WriteTime       uint64 `json:"write_time"`
	IopsInProgress  uint64 `json:"iops_in_progress"`
	IoTime          uint64 `json:"io_time"`
	WeightedIO      uint64 `json:"weighted_io"`
}

// NetworkInfo represents network interface information
type NetworkInfo struct {
	Name         string `json:"name"`
	MTU          int    `json:"mtu"`
	HardwareAddr string `json:"hardware_addr"`
	Flags        []string `json:"flags"`
	Addrs        []NetworkAddr `json:"addrs"`
}

// NetworkAddr represents network address
type NetworkAddr struct {
	Addr string `json:"addr"`
	Net  string `json:"net"`
}

// NetworkIOInfo represents network I/O statistics
type NetworkIOInfo struct {
	Name        string `json:"name"`
	BytesSent   uint64 `json:"bytes_sent"`
	BytesRecv   uint64 `json:"bytes_recv"`
	PacketsSent uint64 `json:"packets_sent"`
	PacketsRecv uint64 `json:"packets_recv"`
	Errin       uint64 `json:"errin"`
	Errout      uint64 `json:"errout"`
	Dropin      uint64 `json:"dropin"`
	Dropout     uint64 `json:"dropout"`
	Fifoin      uint64 `json:"fifoin"`
	Fifoout     uint64 `json:"fifoout"`
}

// StorageProtocolInfo represents storage protocol monitoring
type StorageProtocolInfo struct {
	Protocol    string                 `json:"protocol"`
	Status      string                 `json:"status"`
	Connections int                    `json:"connections"`
	Shares      []ShareInfo            `json:"shares,omitempty"`
	Targets     []ISCSITargetInfo      `json:"targets,omitempty"`
	Statistics  map[string]interface{} `json:"statistics"`
}

// ShareInfo represents SMB/NFS share information
type ShareInfo struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	Protocol    string `json:"protocol"`
	Status      string `json:"status"`
	Connections int    `json:"connections"`
	ReadBytes   uint64 `json:"read_bytes"`
	WriteBytes  uint64 `json:"write_bytes"`
	ReadOps     uint64 `json:"read_ops"`
	WriteOps    uint64 `json:"write_ops"`
}

// ISCSITargetInfo represents iSCSI target information
type ISCSITargetInfo struct {
	IQN         string `json:"iqn"`
	Status      string `json:"status"`
	Sessions    int    `json:"sessions"`
	ReadBytes   uint64 `json:"read_bytes"`
	WriteBytes  uint64 `json:"write_bytes"`
	ReadOps     uint64 `json:"read_ops"`
	WriteOps    uint64 `json:"write_ops"`
}

// MonitoringData represents complete monitoring data
type MonitoringData struct {
	Timestamp       time.Time                 `json:"timestamp"`
	SystemInfo      SystemInfo                `json:"system_info"`
	CPU             CPUInfo                   `json:"cpu"`
	Memory          MemoryInfo                `json:"memory"`
	Disks           []DiskInfo                `json:"disks"`
	DiskIO          []DiskIOInfo              `json:"disk_io"`
	Networks        []NetworkInfo             `json:"networks"`
	NetworkIO       []NetworkIOInfo           `json:"network_io"`
	StorageProtocols []StorageProtocolInfo    `json:"storage_protocols"`
}

// PromQLQueryRequest represents a PromQL query request
type PromQLQueryRequest struct {
	Query     string    `json:"query" binding:"required"`
	Time      time.Time `json:"time,omitempty"`
	StartTime time.Time `json:"start_time,omitempty"`
	EndTime   time.Time `json:"end_time,omitempty"`
	Step      string    `json:"step,omitempty"`
	Timeout   string    `json:"timeout,omitempty"`
}

// PromQLQueryResponse represents a PromQL query response
type PromQLQueryResponse struct {
	Status string      `json:"status"`
	Data   interface{} `json:"data"`
	Error  string      `json:"error,omitempty"`
}

// WebSocketMessage represents WebSocket message structure
type WebSocketMessage struct {
	Type      string      `json:"type"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
	ClientID  string      `json:"client_id,omitempty"`
}

// MonitoringAlert represents monitoring alert
type MonitoringAlert struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Level       string                 `json:"level"`
	Message     string                 `json:"message"`
	Metric      string                 `json:"metric"`
	Value       float64                `json:"value"`
	Threshold   float64                `json:"threshold"`
	Timestamp   time.Time              `json:"timestamp"`
	Tags        map[string]string      `json:"tags"`
	Annotations map[string]interface{} `json:"annotations"`
}