package utils

import (
	"encoding/json"
	"fmt"
	lvm_models "pnas/internal/models"
	"time"
)

const (
	PVS      = "/usr/sbin/pvs"
	VGS      = "/usr/sbin/vgs"
	LVS      = "/usr/sbin/lvs"
	PVCREATE = "/usr/sbin/pvcreate"
	VGCREATE = "/usr/sbin/vgcreate"
	LVCREATE = "/usr/sbin/lvcreate"
	PVREMOVE = "/usr/sbin/pvremove"
	VGREMOVE = "/usr/sbin/vgremove"
	LVREMOVE = "/usr/sbin/lvremove"
)

type LVM struct{}

func (l *LVM) GetPVS() ([]lvm_models.PhysicalVolume, error) {
	var pvsCmd = []string{"-o", "pv_name,vg_name,lv_name,pv_size,lv_size,vg_size,pv_uuid,vg_uuid,lv_uuid,pv_attr",
		"--reportformat", "json", "--units", "B"}
	result := ExecCommand(ExecOptions{Timeout: 5 * time.Second}, PVS, pvsCmd...)
	if result.Error != nil {
		return nil, result.Error
	}

	var reportLog lvm_models.ReportLog
	if err := json.Unmarshal([]byte(result.Stdout), &reportLog); err != nil {
		return nil, fmt.Errorf("解析pvs输出失败: %v", err)
	}
	return reportLog.Report[0].PV, nil
}

func (l *LVM) GetVGS() ([]lvm_models.VolumeGroup, error) {
	var vgsCmd = []string{"-o", "vg_name,vg_uuid,vg_attr,vg_size,vg_free,vg_extent_count,vg_extent_size,vg_free_count",
		"--reportformat", "json", "--units", "B"}
	result := ExecCommand(ExecOptions{Timeout: 5 * time.Second}, VGS, vgsCmd...)
	if result.Error != nil {
		return nil, result.Error
	}
	fmt.Println("1")
	var reportLog lvm_models.ReportLog
	if err := json.Unmarshal([]byte(result.Stdout), &reportLog); err != nil {
		return nil, fmt.Errorf("解析vgs输出失败: %v", err)
	}
	return reportLog.Report[0].VG, nil
}

func (l *LVM) GetLVS() ([]lvm_models.LogicalVolume, error) {
	var lvsCmd = []string{"-o", "lv_name,vg_name,lv_uuid,lv_size,lv_attr",
		"--reportformat", "json", "--units", "B"}
	result := ExecCommand(ExecOptions{Timeout: 5 * time.Second}, LVS, lvsCmd...)
	if result.Error != nil {
		return nil, result.Error
	}

	var reportLog lvm_models.ReportLog
	if err := json.Unmarshal([]byte(result.Stdout), &reportLog); err != nil {
		return nil, fmt.Errorf("解析lvs输出失败: %v", err)
	}
	return reportLog.Report[0].LV, nil
}

func (l *LVM) CreatePV(devices []string) error {
	devices = append(devices, "-y")
	devices = append(devices, "-f")

	result := ExecCommand(ExecOptions{Timeout: 5 * time.Second}, PVCREATE, devices...)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (l *LVM) RemovePV(devices []string) error {
	devices = append(devices, "-y")
	devices = append(devices, "-f")

	result := ExecCommand(ExecOptions{Timeout: 5 * time.Second}, PVREMOVE, devices...)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (l *LVM) CreateVG(vgName string, pvDevices []string) error {
	pvDevices = append(pvDevices, "-y")
	pvDevices = append(pvDevices, "-f")
	pvDevices = append([]string{vgName}, pvDevices...)

	result := ExecCommand(ExecOptions{Timeout: 5 * time.Second}, VGCREATE, pvDevices...)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (l *LVM) RemoveVG(vgName string) error {
	var removeVG = []string{vgName, "-y"}

	result := ExecCommand(ExecOptions{Timeout: 5 * time.Second}, VGREMOVE, removeVG...)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (l *LVM) CreateLV(vgName string, lvName string, size string) error {
	var createLV = []string{}
	if size == "all" {
		createLV = append(createLV, "-l", "100%VG", "-n", lvName, vgName)
	} else {
		createLV = append(createLV, "-L", size, "-n", lvName, vgName)
	}
	result := ExecCommand(ExecOptions{Timeout: 5 * time.Second}, LVCREATE, createLV...)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// 删除逻辑卷
func (l *LVM) RemoveLV(vgName string, lvName string) error {
	var removeLV = []string{vgName + "/" + lvName, "-y"}
	result := ExecCommand(ExecOptions{Timeout: 5 * time.Second}, LVREMOVE, removeLV...)
	if result.Error != nil {
		return result.Error
	}
	return nil
}
