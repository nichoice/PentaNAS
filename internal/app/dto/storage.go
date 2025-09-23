package dto

// CreateVGRequest 创建卷组请求
type CreateVGRequest struct {
	VgName  string   `json:"vg_name" binding:"required"`
	Devices []string `json:"devices" binding:"required"`
}

// CreateLVRequest 创建逻辑卷请求
type CreateLVRequest struct {
	VgName string `json:"vg_name" binding:"required"`
	LvName string `json:"lv_name" binding:"required"`
	Size   string `json:"size" binding:"required"`
}

// DiskResponse 磁盘响应
type DiskResponse struct {
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
}

// VolumeGroupResponse 卷组响应
type VolumeGroupResponse struct {
	VGName         string `json:"vg_name"`
	VGUUID         string `json:"vg_uuid"`
	VGAttr         string `json:"vg_attr"`
	VGSize         string `json:"vg_size"`
	VGFree         string `json:"vg_free"`
	VGExtentCount  string `json:"vg_extent_count"`
	VGExtentSize   string `json:"vg_extent_size"`
	VGFreeCount    string `json:"vg_free_count"`
}

// LogicalVolumeResponse 逻辑卷响应
type LogicalVolumeResponse struct {
	LVName string `json:"lv_name"`
	VGName string `json:"vg_name"`
	LVUUID string `json:"lv_uuid"`
	LVSize string `json:"lv_size"`
	LVAttr string `json:"lv_attr"`
}

// StorageOverviewResponse 存储概览响应
type StorageOverviewResponse struct {
	TotalDisks          int                     `json:"total_disks"`
	TotalVolumeGroups   int                     `json:"total_volume_groups"`
	TotalLogicalVolumes int                     `json:"total_logical_volumes"`
	Disks               []DiskResponse          `json:"disks"`
	VolumeGroups        []VolumeGroupResponse   `json:"volume_groups"`
	LogicalVolumes      []LogicalVolumeResponse `json:"logical_volumes"`
}