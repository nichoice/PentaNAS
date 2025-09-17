package models

// StoragePool 存储池模型
type PhysicalVolume struct {
	PVName string `json:"pv_name"`
	VGName string `json:"vg_name"`
	LVName string `json:"lv_name"`
	PVSize string `json:"pv_size"`
	LVSize string `json:"lv_size"`
	VGSize string `json:"vg_size"`
	PVUUID string `json:"pv_uuid"`
	VGUUID string `json:"vg_uuid"`
	LVUUID string `json:"lv_uuid"`
}

// 定义卷组(VG)结构体
type VolumeGroup struct {
	VGName        string `json:"vg_name"`
	VGUUID        string `json:"vg_uuid"`
	VGAttr        string `json:"vg_attr"`
	VGSize        string `json:"vg_size"`
	VGFree        string `json:"vg_free"`
	VGExtentCount string `json:"vg_extent_count"`
	VGExtentSize  string `json:"vg_extent_size"`
	VGFreeCount   string `json:"vg_free_count"`
}

// 定义逻辑卷(LV)结构体
type LogicalVolume struct {
	LVName string `json:"lv_name"`
	VGName string `json:"vg_name"`
	LVUUID string `json:"lv_uuid"`
	LVSize string `json:"lv_size"`
	LVAttr string `json:"lv_attr"`
}

type Report struct {
	PV []PhysicalVolume `json:"pv"`
	VG []VolumeGroup    `json:"vg"`
	LV []LogicalVolume  `json:"lv"`
}

type ReportLog struct {
	Report []Report `json:"report"`
	Log    []string `json:"log"`
}
