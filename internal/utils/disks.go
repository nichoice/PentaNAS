package utils

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	dk "pnas/internal/models"
	"strconv"
	"strings"
	"time"
)

type Disks struct{}

const (
	DiskTypeHDD     = "HDD"
	DiskTypeSSD     = "SSD"
	DiskTypeNVMe    = "NVMe"
	DiskTypeVirtual = "Virtual"
	DiskTypeQemu    = "Qemu"
	DiskTypeUnkown  = "Unkown"
)

const (
	HealthStatusNormal   = "GOOD"     // 良好
	HealthStatusWarning  = "Warning"  // 警告
	HealthStatusCritical = "Critical" //严重
	HealthStatusUnknown  = "Unknown"  //未知
)

const (
	LSBLK    = "/usr/bin/lsblk"
	DF       = "/usr/bin/df"
	SMARTCTL = "/usr/sbin/smartctl"
)

// 获取磁盘使用情况
func (d *Disks) getDiskUsage(mountPoint string) (*dk.DiskUsage, error) {
	var getDiskUseCmd []string
	getDiskUseCmd = []string{"-B1", mountPoint}
	result := ExecCommand(ExecOptions{Timeout: 10 * time.Second}, DF, getDiskUseCmd...)
	if result.Error != nil {
		return nil, result.Error
	}
	lines := strings.Split(result.Stdout, "\n")
	if len(lines) < 2 {
		return nil, errors.New("df 输出格式异常")
	}

	//解析第二行
	fields := strings.Fields(lines[1])
	if len(fields) < 4 {
		return nil, errors.New("df 输出字段不足")
	}
	total, _ := strconv.ParseUint(fields[1], 10, 64)
	used, _ := strconv.ParseUint(fields[2], 10, 64)
	avail, _ := strconv.ParseUint(fields[3], 10, 64)

	userPercent := 0.0
	if total > 0 {
		userPercent = float64(used) / float64(total) * 100
	}
	return &dk.DiskUsage{
		Total:      total,
		Used:       used,
		Avail:      avail,
		UsePercent: userPercent,
	}, nil

}

// formatBytes 格式化字节数为可读格式
func (d *Disks) formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// formatTemperature 格式化温度为可读格式
func (d *Disks) formatTemperature(temperature int) string {
	if temperature < 0 {
		return fmt.Sprintf("%d°C", temperature)
	}
	return fmt.Sprintf("%d°C", temperature)
}

func (d *Disks) getSystemDiskPath() []string {
	var systemPaths []string
	// 读取/proc/mounts查找系统关键挂载点
	file, err := os.Open("/proc/mounts")
	if err != nil {
		return systemPaths
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) >= 2 {
			device := fields[0]
			mountPoint := fields[1]
			if mountPoint == "/" || mountPoint == "/boot" || mountPoint == "/var" || mountPoint == "/usr" ||
				mountPoint == "/home" || mountPoint == "/tmp" || mountPoint == "/var/log" || mountPoint == "/sys" ||
				strings.HasPrefix(mountPoint, "/proc") {
				systemPaths = append(systemPaths, device)
			}
		}
	}
	return systemPaths
}

func (d *Disks) isSystemDevice(devicePath string, systemPaths []string) bool {
	for _, sysPath := range systemPaths {
		if strings.Contains(sysPath, devicePath) || strings.Contains(devicePath, sysPath) {
			return true
		}
	}
	return false
}

func (d *Disks) getDiskTemperature(devicePath string) int {
	// 尝试从hwmon系统文件获取温度
	patterns := []string{
		"/sys/class/hwmon/hwmon*/temp*_input",
		"/sys/class/thermal/thermal_zone*/temp",
	}

	for _, pattern := range patterns {
		files, _ := filepath.Glob(pattern)
		for _, file := range files {
			if strings.Contains(file, devicePath) {
				content, err := os.ReadFile(file)
				if err != nil {
					continue
				}
				temp, err := strconv.Atoi(strings.TrimSpace(string(content)))
				if err != nil {
					continue
				}
				return temp / 1000
			}
		}
	}

	// 如果smartctl 可用
	smartctlCmd := []string{"-A", devicePath}
	result := ExecCommand(ExecOptions{Timeout: 10 * time.Second}, SMARTCTL, smartctlCmd...)
	if result.Error != nil {
		return 0
	}

	var smartInfo dk.SmartInfo
	if err := json.Unmarshal([]byte(result.Stdout), &smartInfo); err != nil {
		return 0
	}
	temp := smartInfo.Temperature.Current
	if temp > 0 {
		return temp
	}

	// lines := strings.Split(result.Stdout, "\n")
	// for _, line := range lines {
	// 	if strings.Contains(strings.ToLower(line), "temperature") {
	// 		fields := strings.Fields(line)
	// 		for _, field := range fields {
	// 			if temp, err := strconv.Atoi(field); err == nil && temp < 200 {
	// 				return temp
	// 			}
	// 		}
	// 	}
	// }

	return 0
}

func CheckDiskHealth(devicePath string) (bool, error) {
	smartctlCmd := []string{"-A", devicePath}
	result := ExecCommand(ExecOptions{Timeout: 10 * time.Second}, SMARTCTL, smartctlCmd...)
	if result.Error != nil {
		return false, result.Error
	}

	var smartHealthInfo dk.SmartHealthInfo
	if err := json.Unmarshal([]byte(result.Stdout), &smartHealthInfo); err != nil {
		return false, err
	}

	health := smartHealthInfo.SmartStatus.Passed

	return health, nil
}

// 扫描磁盘
func (d *Disks) ScanDisks() ([]*dk.Disk, error) {

	var scanDisksCmd []string
	scanDisksCmd = []string{"-b", "-J", "-o", "NAME,PATH,SIZE,SERIAL,ROTA,MODEL,VENDOR,TYPE,MAJ:MIN,MOUNTPOINT,FSTYPE"}
	result := ExecCommand(ExecOptions{Timeout: 3 * time.Second}, LSBLK, scanDisksCmd...)
	if result.Error != nil {
		return nil, result.Error
	}
	var data struct {
		BlockDevices []dk.BlockDevice `json:"blockdevices"`
	}

	if err := json.Unmarshal([]byte(result.Stdout), &data); err != nil {
		return nil, fmt.Errorf("解析lsblk输出失败: %v", err)
	}

	var disks []*dk.Disk

	for _, device := range data.BlockDevices {
		if device.Type == "loop" || device.Type == "rom" {
			continue
		}

		disk := &dk.Disk{
			Name:         device.Name,
			Path:         device.Path,
			Size:         d.formatBytes(device.Size),
			Serial:       device.Serial,
			Rota:         device.Rota,
			Model:        device.Model,
			Vendor:       device.Vendor,
			Type:         device.Type,
			MajMin:       device.MajMin,
			MountPoint:   device.Mountpoint,
			FileSystem:   device.GetFstype(),
			IsSystemDisk: device.Type == "disk",
			IsOnline:     true,
			Temperature:  0,
			Health:       "OK", //默认他为健康
		}

		// // 检查是否为系统盘
		// systemPaths := d.getSystemDiskPath()
		// disk.IsSystemDisk = d.isSystemDevice(device.Path, systemPaths)

		// // 获取磁盘使用情况
		// if device.MountPoint != "" {
		// 	usage, err := d.getDiskUsage(device.MountPoint)
		// 	if err == nil {
		// 		disk.UsedSize += usage.Used
		// 		disk.AvailSize += usage.Avail
		// 		if device.Size > 0 {
		// 			disk.UsageRate = float64(disk.UsedSize) / float64(device.Size) * 100
		// 		}
		// 	}
		// }

		// // 获取磁盘温度
		// disk.Temperature = d.getDiskTemperature(device.Path)

		disks = append(disks, disk)
	}
	return disks, nil
}
