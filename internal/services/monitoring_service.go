package services

import (
	"context"
	"fmt"
	"net"
	"os/exec"
	"pnas/internal/app/dto"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
	psnet "github.com/shirou/gopsutil/v3/net"
)

// MonitoringService handles system monitoring operations
type MonitoringService struct {
	ctx    context.Context
	cancel context.CancelFunc
}

// NewMonitoringService creates a new monitoring service
func NewMonitoringService() *MonitoringService {
	ctx, cancel := context.WithCancel(context.Background())
	return &MonitoringService{
		ctx:    ctx,
		cancel: cancel,
	}
}

// Close stops the monitoring service
func (s *MonitoringService) Close() {
	s.cancel()
}

// GetSystemInfo retrieves system information
func (s *MonitoringService) GetSystemInfo() (*dto.SystemInfo, error) {
	hostInfo, err := host.Info()
	if err != nil {
		return nil, fmt.Errorf("failed to get host info: %w", err)
	}

	bootTime := time.Unix(int64(hostInfo.BootTime), 0)

	return &dto.SystemInfo{
		Hostname:        hostInfo.Hostname,
		Platform:        hostInfo.Platform,
		PlatformFamily:  hostInfo.PlatformFamily,
		PlatformVersion: hostInfo.PlatformVersion,
		KernelVersion:   hostInfo.KernelVersion,
		KernelArch:      hostInfo.KernelArch,
		Uptime:          hostInfo.Uptime,
		BootTime:        bootTime,
		Procs:           hostInfo.Procs,
		OS:              hostInfo.OS,
		HostID:          hostInfo.HostID,
	}, nil
}

// GetCPUInfo retrieves CPU information and usage
func (s *MonitoringService) GetCPUInfo() (*dto.CPUInfo, error) {
	// Get CPU info
	cpuInfos, err := cpu.Info()
	if err != nil {
		return nil, fmt.Errorf("failed to get CPU info: %w", err)
	}

	if len(cpuInfos) == 0 {
		return nil, fmt.Errorf("no CPU information available")
	}

	cpuInfo := cpuInfos[0]

	// Get CPU usage
	usage, err := cpu.Percent(time.Second, true)
	if err != nil {
		return nil, fmt.Errorf("failed to get CPU usage: %w", err)
	}

	// Get CPU times for detailed stats
	cpuTimes, err := cpu.Times(true)
	if err != nil {
		return nil, fmt.Errorf("failed to get CPU times: %w", err)
	}

	// Get load average
	loadAvg, err := load.Avg()
	var loadInfo *dto.LoadAvgInfo
	if err == nil {
		loadInfo = &dto.LoadAvgInfo{
			Load1:  loadAvg.Load1,
			Load5:  loadAvg.Load5,
			Load15: loadAvg.Load15,
		}
	}

	// Convert CPU times to per-CPU info
	perCPU := make([]dto.PerCPUInfo, len(cpuTimes))
	for i, t := range cpuTimes {
		perCPU[i] = dto.PerCPUInfo{
			CPU:     t.CPU,
			User:    t.User,
			System:  t.System,
			Idle:    t.Idle,
			Nice:    t.Nice,
			Iowait:  t.Iowait,
			Irq:     t.Irq,
			Softirq: t.Softirq,
			Steal:   t.Steal,
			Guest:   t.Guest,
		}
	}

	// Get CPU temperature (Linux only)
	temperature := s.getCPUTemperature()

	return &dto.CPUInfo{
		ModelName:   cpuInfo.ModelName,
		Cores:       cpuInfo.Cores,
		Threads:     cpuInfo.Cores, // Assume threads = cores for simplicity
		Family:      cpuInfo.Family,
		Speed:       cpuInfo.Mhz,
		CacheSize:   cpuInfo.CacheSize,
		Usage:       usage,
		Temperature: temperature,
		LoadAvg:     loadInfo,
		PerCPU:      perCPU,
	}, nil
}

// GetMemoryInfo retrieves memory information and usage
func (s *MonitoringService) GetMemoryInfo() (*dto.MemoryInfo, error) {
	vmStat, err := mem.VirtualMemory()
	if err != nil {
		return nil, fmt.Errorf("failed to get virtual memory info: %w", err)
	}

	swapStat, err := mem.SwapMemory()
	if err != nil {
		return nil, fmt.Errorf("failed to get swap memory info: %w", err)
	}

	return &dto.MemoryInfo{
		Total:       vmStat.Total,
		Available:   vmStat.Available,
		Used:        vmStat.Used,
		UsedPercent: vmStat.UsedPercent,
		Free:        vmStat.Free,
		Active:      vmStat.Active,
		Inactive:    vmStat.Inactive,
		Wired:       vmStat.Wired,
		Laundry:     vmStat.Laundry,
		Buffers:     vmStat.Buffers,
		Cached:      vmStat.Cached,
		Shared:      vmStat.Shared,
		Slab:        vmStat.Slab,
		SwapTotal:   swapStat.Total,
		SwapUsed:    swapStat.Used,
		SwapFree:    swapStat.Free,
		SwapPercent: swapStat.UsedPercent,
	}, nil
}

// GetDiskInfo retrieves disk information and usage
func (s *MonitoringService) GetDiskInfo() ([]dto.DiskInfo, error) {
	partitions, err := disk.Partitions(false)
	if err != nil {
		return nil, fmt.Errorf("failed to get disk partitions: %w", err)
	}

	var disks []dto.DiskInfo
	for _, partition := range partitions {
		usage, err := disk.Usage(partition.Mountpoint)
		if err != nil {
			continue // Skip partitions that can't be accessed
		}

		diskInfo := dto.DiskInfo{
			Device:      partition.Device,
			Mountpoint:  partition.Mountpoint,
			Fstype:      partition.Fstype,
			Total:       usage.Total,
			Free:        usage.Free,
			Used:        usage.Used,
			UsedPercent: usage.UsedPercent,
			InodesTotal: usage.InodesTotal,
			InodesUsed:  usage.InodesUsed,
			InodesFree:  usage.InodesFree,
			InodesUsedPercent: usage.InodesUsedPercent,
		}
		disks = append(disks, diskInfo)
	}

	return disks, nil
}

// GetDiskIOInfo retrieves disk I/O statistics
func (s *MonitoringService) GetDiskIOInfo() ([]dto.DiskIOInfo, error) {
	ioCounters, err := disk.IOCounters()
	if err != nil {
		return nil, fmt.Errorf("failed to get disk I/O counters: %w", err)
	}

	var diskIOs []dto.DiskIOInfo
	for name, counter := range ioCounters {
		diskIO := dto.DiskIOInfo{
			Name:             name,
			ReadCount:        counter.ReadCount,
			MergedReadCount:  counter.MergedReadCount,
			WriteCount:       counter.WriteCount,
			MergedWriteCount: counter.MergedWriteCount,
			ReadBytes:        counter.ReadBytes,
			WriteBytes:       counter.WriteBytes,
			ReadTime:         counter.ReadTime,
			WriteTime:        counter.WriteTime,
			IopsInProgress:   counter.IopsInProgress,
			IoTime:           counter.IoTime,
			WeightedIO:       counter.WeightedIO,
		}
		diskIOs = append(diskIOs, diskIO)
	}

	return diskIOs, nil
}

// GetNetworkInfo retrieves network interface information
func (s *MonitoringService) GetNetworkInfo() ([]dto.NetworkInfo, error) {
	interfaces, err := psnet.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("failed to get network interfaces: %w", err)
	}

	var networks []dto.NetworkInfo
	for _, iface := range interfaces {
		addrs := make([]dto.NetworkAddr, len(iface.Addrs))
		for i, addr := range iface.Addrs {
			addrs[i] = dto.NetworkAddr{
				Addr: addr.Addr,
				Net:  "tcp", // Default network type
			}
		}

		network := dto.NetworkInfo{
			Name:         iface.Name,
			MTU:          iface.MTU,
			HardwareAddr: iface.HardwareAddr,
			Flags:        iface.Flags,
			Addrs:        addrs,
		}
		networks = append(networks, network)
	}

	return networks, nil
}

// GetNetworkIOInfo retrieves network I/O statistics
func (s *MonitoringService) GetNetworkIOInfo() ([]dto.NetworkIOInfo, error) {
	ioCounters, err := psnet.IOCounters(true)
	if err != nil {
		return nil, fmt.Errorf("failed to get network I/O counters: %w", err)
	}

	var networkIOs []dto.NetworkIOInfo
	for _, counter := range ioCounters {
		networkIO := dto.NetworkIOInfo{
			Name:        counter.Name,
			BytesSent:   counter.BytesSent,
			BytesRecv:   counter.BytesRecv,
			PacketsSent: counter.PacketsSent,
			PacketsRecv: counter.PacketsRecv,
			Errin:       counter.Errin,
			Errout:      counter.Errout,
			Dropin:      counter.Dropin,
			Dropout:     counter.Dropout,
			Fifoin:      counter.Fifoin,
			Fifoout:     counter.Fifoout,
		}
		networkIOs = append(networkIOs, networkIO)
	}

	return networkIOs, nil
}

// GetStorageProtocolInfo retrieves storage protocol monitoring information
func (s *MonitoringService) GetStorageProtocolInfo() ([]dto.StorageProtocolInfo, error) {
	var protocols []dto.StorageProtocolInfo

	// Check SMB/CIFS
	smbInfo := s.getSMBInfo()
	if smbInfo != nil {
		protocols = append(protocols, *smbInfo)
	}

	// Check NFS
	nfsInfo := s.getNFSInfo()
	if nfsInfo != nil {
		protocols = append(protocols, *nfsInfo)
	}

	// Check iSCSI
	iscsiInfo := s.getISCSIInfo()
	if iscsiInfo != nil {
		protocols = append(protocols, *iscsiInfo)
	}

	return protocols, nil
}

// GetCompleteMonitoringData retrieves all monitoring data
func (s *MonitoringService) GetCompleteMonitoringData() (*dto.MonitoringData, error) {
	systemInfo, err := s.GetSystemInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to get system info: %w", err)
	}

	cpuInfo, err := s.GetCPUInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to get CPU info: %w", err)
	}

	memoryInfo, err := s.GetMemoryInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to get memory info: %w", err)
	}

	diskInfo, err := s.GetDiskInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to get disk info: %w", err)
	}

	diskIOInfo, err := s.GetDiskIOInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to get disk I/O info: %w", err)
	}

	networkInfo, err := s.GetNetworkInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to get network info: %w", err)
	}

	networkIOInfo, err := s.GetNetworkIOInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to get network I/O info: %w", err)
	}

	storageProtocolInfo, err := s.GetStorageProtocolInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to get storage protocol info: %w", err)
	}

	return &dto.MonitoringData{
		Timestamp:        time.Now(),
		SystemInfo:       *systemInfo,
		CPU:              *cpuInfo,
		Memory:           *memoryInfo,
		Disks:            diskInfo,
		DiskIO:           diskIOInfo,
		Networks:         networkInfo,
		NetworkIO:        networkIOInfo,
		StorageProtocols: storageProtocolInfo,
	}, nil
}

// Helper functions

// getCPUTemperature gets CPU temperature (Linux only)
func (s *MonitoringService) getCPUTemperature() float64 {
	if runtime.GOOS != "linux" {
		return 0
	}

	// Try to read from thermal zones
	cmd := exec.Command("bash", "-c", "cat /sys/class/thermal/thermal_zone*/temp 2>/dev/null | head -1")
	output, err := cmd.Output()
	if err != nil {
		return 0
	}

	temp, err := strconv.ParseFloat(strings.TrimSpace(string(output)), 64)
	if err != nil {
		return 0
	}

	// Convert from millidegrees to degrees
	return temp / 1000
}

// getSMBInfo gets SMB/CIFS protocol information
func (s *MonitoringService) getSMBInfo() *dto.StorageProtocolInfo {
	if runtime.GOOS != "linux" {
		return nil
	}

	// Check if SMB service is running
	cmd := exec.Command("systemctl", "is-active", "smbd")
	output, err := cmd.Output()
	status := "stopped"
	if err == nil && strings.TrimSpace(string(output)) == "active" {
		status = "running"
	}

	connections := s.getSMBConnections()
	shares := s.getSMBShares()

	return &dto.StorageProtocolInfo{
		Protocol:    "SMB",
		Status:      status,
		Connections: connections,
		Shares:      shares,
		Statistics:  make(map[string]interface{}),
	}
}

// getNFSInfo gets NFS protocol information
func (s *MonitoringService) getNFSInfo() *dto.StorageProtocolInfo {
	if runtime.GOOS != "linux" {
		return nil
	}

	// Check if NFS service is running
	cmd := exec.Command("systemctl", "is-active", "nfs-server")
	output, err := cmd.Output()
	status := "stopped"
	if err == nil && strings.TrimSpace(string(output)) == "active" {
		status = "running"
	}

	connections := s.getNFSConnections()
	shares := s.getNFSShares()

	return &dto.StorageProtocolInfo{
		Protocol:    "NFS",
		Status:      status,
		Connections: connections,
		Shares:      shares,
		Statistics:  make(map[string]interface{}),
	}
}

// getISCSIInfo gets iSCSI protocol information
func (s *MonitoringService) getISCSIInfo() *dto.StorageProtocolInfo {
	if runtime.GOOS != "linux" {
		return nil
	}

	// Check if iSCSI target service is running
	cmd := exec.Command("systemctl", "is-active", "target")
	output, err := cmd.Output()
	status := "stopped"
	if err == nil && strings.TrimSpace(string(output)) == "active" {
		status = "running"
	}

	connections := s.getISCSIConnections()
	targets := s.getISCSITargets()

	return &dto.StorageProtocolInfo{
		Protocol:    "iSCSI",
		Status:      status,
		Connections: connections,
		Targets:     targets,
		Statistics:  make(map[string]interface{}),
	}
}

// getSMBConnections gets SMB connection count
func (s *MonitoringService) getSMBConnections() int {
	cmd := exec.Command("smbstatus", "-b")
	output, err := cmd.Output()
	if err != nil {
		return 0
	}

	lines := strings.Split(string(output), "\n")
	count := 0
	for _, line := range lines {
		if strings.Contains(line, "pid") && strings.Contains(line, "machine") {
			count++
		}
	}
	return count
}

// getSMBShares gets SMB shares information
func (s *MonitoringService) getSMBShares() []dto.ShareInfo {
	var shares []dto.ShareInfo

	cmd := exec.Command("smbstatus", "-S")
	output, err := cmd.Output()
	if err != nil {
		return shares
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "Service") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) >= 3 {
			share := dto.ShareInfo{
				Name:        fields[0],
				Path:        fields[1],
				Protocol:    "SMB",
				Status:      "active",
				Connections: 0, // Would need more parsing to get accurate count
			}
			shares = append(shares, share)
		}
	}

	return shares
}

// getNFSConnections gets NFS connection count
func (s *MonitoringService) getNFSConnections() int {
	cmd := exec.Command("ss", "-tn", "sport", "= :2049")
	output, err := cmd.Output()
	if err != nil {
		return 0
	}

	lines := strings.Split(string(output), "\n")
	return len(lines) - 1 // Subtract header line
}

// getNFSShares gets NFS shares information
func (s *MonitoringService) getNFSShares() []dto.ShareInfo {
	var shares []dto.ShareInfo

	cmd := exec.Command("exportfs", "-v")
	output, err := cmd.Output()
	if err != nil {
		return shares
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) >= 1 {
			share := dto.ShareInfo{
				Name:     parts[0],
				Path:     parts[0],
				Protocol: "NFS",
				Status:   "active",
			}
			shares = append(shares, share)
		}
	}

	return shares
}

// getISCSIConnections gets iSCSI connection count
func (s *MonitoringService) getISCSIConnections() int {
	cmd := exec.Command("ss", "-tn", "sport", "= :3260")
	output, err := cmd.Output()
	if err != nil {
		return 0
	}

	lines := strings.Split(string(output), "\n")
	return len(lines) - 1 // Subtract header line
}

// getISCSITargets gets iSCSI targets information
func (s *MonitoringService) getISCSITargets() []dto.ISCSITargetInfo {
	var targets []dto.ISCSITargetInfo

	cmd := exec.Command("targetcli", "ls")
	output, err := cmd.Output()
	if err != nil {
		return targets
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "iqn.") {
			fields := strings.Fields(line)
			if len(fields) >= 1 {
				target := dto.ISCSITargetInfo{
					IQN:    fields[0],
					Status: "active",
				}
				targets = append(targets, target)
			}
		}
	}

	return targets
}

// IsServiceRunning checks if a systemd service is running
func (s *MonitoringService) IsServiceRunning(serviceName string) bool {
	if runtime.GOOS != "linux" {
		return false
	}

	cmd := exec.Command("systemctl", "is-active", serviceName)
	output, err := cmd.Output()
	return err == nil && strings.TrimSpace(string(output)) == "active"
}

// GetServiceStatus gets detailed service status
func (s *MonitoringService) GetServiceStatus(serviceName string) map[string]interface{} {
	status := make(map[string]interface{})

	if runtime.GOOS != "linux" {
		status["supported"] = false
		return status
	}

	cmd := exec.Command("systemctl", "status", serviceName)
	output, err := cmd.Output()

	status["supported"] = true
	status["running"] = s.IsServiceRunning(serviceName)
	status["output"] = string(output)
	status["error"] = ""

	if err != nil {
		status["error"] = err.Error()
	}

	return status
}

// GetNetworkConnections gets network connections for a specific port
func (s *MonitoringService) GetNetworkConnections(port int) ([]net.Conn, error) {
	// This is a simplified implementation
	// In a real application, you might want to use netstat or ss commands

	connections := make([]net.Conn, 0)

	// Check if anything is listening on the port
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		// Port is likely in use
		return connections, nil
	}
	listener.Close()

	return connections, nil
}