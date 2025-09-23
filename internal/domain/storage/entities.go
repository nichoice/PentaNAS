package storage

import "time"

// Disk 磁盘领域实体
type Disk struct {
	Name         string  `json:"name"`
	Path         string  `json:"path"`
	Size         uint64  `json:"size"`
	UsedSize     uint64  `json:"used_size"`
	AvailSize    uint64  `json:"avail_size"`
	UsageRate    float64 `json:"usage_rate"`
	Serial       string  `json:"serial"`
	Rota         bool    `json:"rota"`
	Model        string  `json:"model"`
	Vendor       string  `json:"vendor"`
	Type         string  `json:"type"`
	MajMin       string  `json:"maj_min"`
	MountPoint   string  `json:"mount_point"`
	FileSystem   string  `json:"file_system"`
	IsSystemDisk bool    `json:"is_system_disk"`
	IsOnline     bool    `json:"is_online"`
	Temperature  int     `json:"temperature"`
	Health       string  `json:"health"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// PhysicalVolume LVM物理卷
type PhysicalVolume struct {
	PVName   string `json:"pv_name"`
	VGName   string `json:"vg_name"`
	PVSize   string `json:"pv_size"`
	PVUUID   string `json:"pv_uuid"`
	PVAttr   string `json:"pv_attr"`
	VGUuid   string `json:"vg_uuid"`
	LVName   string `json:"lv_name"`
	LVSize   string `json:"lv_size"`
	LVUuid   string `json:"lv_uuid"`
}

// VolumeGroup LVM卷组
type VolumeGroup struct {
	VGName         string `json:"vg_name"`
	VGUUID         string `json:"vg_uuid"`
	VGAttr         string `json:"vg_attr"`
	VGSize         string `json:"vg_size"`
	VGFree         string `json:"vg_free"`
	VGExtentCount  string `json:"vg_extent_count"`
	VGExtentSize   string `json:"vg_extent_size"`
	VGFreeCount    string `json:"vg_free_count"`
}

// LogicalVolume LVM逻辑卷
type LogicalVolume struct {
	LVName string `json:"lv_name"`
	VGName string `json:"vg_name"`
	LVUUID string `json:"lv_uuid"`
	LVSize string `json:"lv_size"`
	LVAttr string `json:"lv_attr"`
}

// StoragePool 存储池聚合根
type StoragePool struct {
	ID               string
	Name             string
	Type             string // LVM, ZFS, BTRFS
	TotalCapacity    uint64
	UsedCapacity     uint64
	AvailableCapacity uint64
	Status           string
	PhysicalVolumes  []PhysicalVolume
	VolumeGroups     []VolumeGroup
	LogicalVolumes   []LogicalVolume
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// IsHealthy 检查存储池是否健康
func (sp *StoragePool) IsHealthy() bool {
	return sp.Status == "healthy"
}

// GetUsagePercentage 获取使用率百分比
func (sp *StoragePool) GetUsagePercentage() float64 {
	if sp.TotalCapacity == 0 {
		return 0
	}
	return float64(sp.UsedCapacity) / float64(sp.TotalCapacity) * 100
}