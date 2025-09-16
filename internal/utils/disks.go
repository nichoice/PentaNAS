package utils

import (
	"encoding/json"
	"fmt"
	"time"
)

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
	LSBLK = "/usr/bin/lsblk"
)

var lsblkOutput struct {
	BlockDevices []struct {
		Name       string `json:"name"`
		Path       string `json:"path"`
		Size       string `json:"size"`
		Serial     string `json:"serial"`
		Rota       bool   `json:"rota"`
		Model      string `json:"model"`
		Vendor     string `json:"vendor"`
		Type       string `json:"type"`
		MajMin     string `json:"maj:min"`
		MountPoint string `json:"mountpoint"`
		FsType     string `json:"fstype"`
		Children   []struct {
			Name       string `json:"name"`
			Path       string `json:"path"`
			Size       string `json:"size"`
			MountPoint string `json:"mountpoint"`
			FsType     string `json:"fstype"`
		} `json:"children"`
	} `json:"blockdevices"`
}

// Disk 磁盘模型
type Disk struct {
	Name         string  `json:"name"`           //磁盘名称
	Path         string  `json:"path"`           //磁盘路径
	Size         uint64  `json:"size"`           //磁盘大小(字节)
	UsedSize     uint64  `json:"used_size"`      //已使用大小(字节)
	AvailSize    uint64  `json:"avail_size"`     //可用大小(字节)
	UsageRate    float64 `json:"usage_rate"`     //使用率百分比
	Serial       string  `json:"serial"`         //序列号
	Rota         bool    `json:"rota"`           //是否为机械硬盘
	Model        string  `json:"model"`          //型号
	Vendor       string  `json:"vendor"`         //厂商
	Type         string  `json:"type"`           //磁盘类型
	MajMin       string  `json:"maj_min"`        //主次设备号
	MountPoint   string  `json:"mount_point"`    //挂载点
	FileSystem   string  `json:"file_system"`    //文件系统
	IsSystemDisk bool    `json:"is_system_disk"` //是否为系统盘
	IsOnline     bool    `json:"is_online"`      //是否在线
	Temperature  int     `json:"temperature"`    //温度
	Health       string  `json:"health"`         //健康状态
}

// formatBytes 格式化字节数为人类可读格式
func formatBytes(bytes uint64) string {
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

// formatTemperature 格式化温度为人类可读格式
func formatTemperature(temperature int) string {
	if temperature < 0 {
		return fmt.Sprintf("%d°C", temperature)
	}
	return fmt.Sprintf("%d°C", temperature)
}

// 扫描磁盘
func ScanDisks() ([]*Disk, error) {
	var disks []*Disk
	var scanDisksCmd []string
	scanDisksCmd = []string{"-J", "-o", "NAME,PATH,SIZE,SERIAL,ROTA,MODEL,VENDOR,TYPE,MAJ:MIN,MOUNTPOINT,FSTYPE"}

	result := ExecCommand(ExecOptions{Timeout: 10 * time.Second}, LSBLK, scanDisksCmd...)
	if result.Error != nil {
		return nil, result.Error
	}
	if err := json.Unmarshal([]byte(result.Stdout), &lsblkOutput); err != nil {
		return nil, fmt.Errorf("解析lsblk输出失败: %v", err)
	}

	return disks, nil
}
