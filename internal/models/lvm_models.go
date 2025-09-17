package models

import (
	"fmt"
	"log"
)

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
	LVName     string     `json:"lv_name"`
	VGName     string     `json:"vg_name"`
	LVUUID     string     `json:"lv_uuid"`
	LVSize     string     `json:"lv_size"`
	LVAttr     string     `json:"lv_attr"`
	LVAttrInfo LVAttrInfo `json:"lv_attr_info"`
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

// 把 10 字节属性字符串解析成人类可读字段
type LVAttrInfo struct {
	VolumeType  string `json:"volume_type"`
	Permission  string `json:"permission"`
	AllocPolicy string `json:"alloc_policy"`
	MinIO       string `json:"min_io"`
	State       string `json:"state"`
	Device      string `json:"device"`
	TargetType  string `json:"target_type"`
}

func (lv *LogicalVolume) GetLVAttrInfo() LVAttrInfo {
	var lvAttrInfo *LVAttrInfo
	lvAttrInfo, err := lv.parseLVAttributes(lv.LVAttr)
	if err != nil {
		log.Printf("failed to parse LV attributes: %v", err)
	}
	lv.LVAttrInfo = *lvAttrInfo
	return lv.LVAttrInfo
}

// ParseLVAttributes 解析 lv_attr 字符串，返回结构化信息
func (lv *LogicalVolume) parseLVAttributes(attr string) (*LVAttrInfo, error) {
	if len(attr) < 10 {
		return nil, fmt.Errorf("invalid attr length: %d < 10", len(attr))
	}

	info := &LVAttrInfo{}

	// 1. 卷类型
	switch attr[0] {
	case '-':
		info.VolumeType = "普通线性卷"
	case 's':
		info.VolumeType = "快照卷"
	case 'r':
		info.VolumeType = "精简卷"
	case 'V':
		info.VolumeType = "虚拟卷"
	case 'm':
		info.VolumeType = "镜像卷"
	case 'M':
		info.VolumeType = "镜像卷（未同步）"
	case 'o':
		info.VolumeType = "原点卷"
	case 'O':
		info.VolumeType = "原点卷（精简配置）"
	case 'i':
		info.VolumeType = "镜像或镜像日志"
	case 'I':
		info.VolumeType = "镜像或镜像日志（未同步）"
	case 'c':
		info.VolumeType = "卷组件"
	case 'C':
		info.VolumeType = "卷组件（精简配置）"
	case 'v':
		info.VolumeType = "虚拟卷"
	case 'e':
		info.VolumeType = "精简池元数据卷"
	case 'p':
		info.VolumeType = "精简池数据卷"
	default:
		info.VolumeType = fmt.Sprintf("未知(%c)", attr[0])
	}

	// 2. 权限
	switch attr[1] {
	case 'w':
		info.Permission = "可写"
	case 'r':
		info.Permission = "只读"
	case 'R':
		info.Permission = "只读激活"
	default:
		info.Permission = fmt.Sprintf("未知(%c)", attr[1])
	}

	// 3. 分配策略
	switch attr[2] {
	case 'i':
		info.AllocPolicy = "继承"
	case 'c':
		info.AllocPolicy = "连续"
	case 'n':
		info.AllocPolicy = "正常"
	case 'a':
		info.AllocPolicy = "任意"
	case 'C':
		info.AllocPolicy = "集群"
	default:
		info.AllocPolicy = fmt.Sprintf("未知(%c)", attr[2])
	}

	// 4. 最小 IO
	switch attr[3] {
	case 'm':
		info.MinIO = "最小"
	case '-':
		info.MinIO = "无特殊设置"
	default:
		info.MinIO = fmt.Sprintf("未知(%c)", attr[3])
	}

	// 5. 状态
	switch attr[4] {
	case 'a':
		info.State = "活动"
	case 's':
		info.State = "挂起"
	case 'I':
		info.State = "快照无效"
	case 'S':
		info.State = "挂起的快照"
	case 'm':
		info.State = "合并失败"
	case 'M':
		info.State = "合并目标"
	case 'd':
		info.State = "设备映射而不暂停"
	case 'i':
		info.State = "镜像不一致"
	case 'c':
		info.State = "一致性检查中"
	case 'C':
		info.State = "一致性检查暂停"
	case 'X':
		info.State = "未知"
	default:
		info.State = fmt.Sprintf("未知(%c)", attr[4])
	}

	// 6. 设备状态
	switch attr[5] {
	case 'o':
		info.Device = "打开"
	case '-':
		info.Device = "关闭"
	default:
		info.Device = fmt.Sprintf("未知(%c)", attr[5])
	}

	// 7-10. 目标类型
	target := attr[6:10]
	switch target {
	case "----":
		info.TargetType = "无特殊设置"
	case "thin":
		info.TargetType = "精简配置"
	case "raid":
		info.TargetType = "RAID"
	case "snap":
		info.TargetType = "快照"
	case "cache":
		info.TargetType = "缓存"
	case "mirr":
		info.TargetType = "镜像"
	case "orig":
		info.TargetType = "原始"
	default:
		info.TargetType = fmt.Sprintf("未知(%s)", target)
	}

	return info, nil
}
