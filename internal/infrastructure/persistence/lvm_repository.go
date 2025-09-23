package persistence

import (
	"context"
	"encoding/json"
	"fmt"
	"pnas/internal/domain/storage"
	"pnas/internal/shared/utils"
)

// LVMRepository LVM仓储实现
type LVMRepository struct {
	cmdExecutor *utils.CommandExecutor
}

// NewLVMRepository 创建LVM仓储
func NewLVMRepository(cmdExecutor *utils.CommandExecutor) *LVMRepository {
	return &LVMRepository{
		cmdExecutor: cmdExecutor,
	}
}

// LVM命令路径常量
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

// LVM报告结构
type LVMReport struct {
	Report []struct {
		PV []storage.PhysicalVolume `json:"pv,omitempty"`
		VG []storage.VolumeGroup    `json:"vg,omitempty"`
		LV []storage.LogicalVolume  `json:"lv,omitempty"`
	} `json:"report"`
}

// GetPhysicalVolumes 获取物理卷
func (r *LVMRepository) GetPhysicalVolumes(ctx context.Context) ([]storage.PhysicalVolume, error) {
	args := []string{"-o", "pv_name,vg_name,lv_name,pv_size,lv_size,vg_size,pv_uuid,vg_uuid,lv_uuid,pv_attr",
		"--reportformat", "json", "--units", "B"}

	result, err := r.cmdExecutor.Execute(ctx, PVS, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute pvs command: %w", err)
	}

	var report LVMReport
	if err := json.Unmarshal([]byte(result.Stdout), &report); err != nil {
		return nil, fmt.Errorf("failed to parse pvs output: %w", err)
	}

	if len(report.Report) == 0 {
		return []storage.PhysicalVolume{}, nil
	}

	return report.Report[0].PV, nil
}

// GetVolumeGroups 获取卷组
func (r *LVMRepository) GetVolumeGroups(ctx context.Context) ([]storage.VolumeGroup, error) {
	args := []string{"-o", "vg_name,vg_uuid,vg_attr,vg_size,vg_free,vg_extent_count,vg_extent_size,vg_free_count",
		"--reportformat", "json", "--units", "B"}

	result, err := r.cmdExecutor.Execute(ctx, VGS, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute vgs command: %w", err)
	}

	var report LVMReport
	if err := json.Unmarshal([]byte(result.Stdout), &report); err != nil {
		return nil, fmt.Errorf("failed to parse vgs output: %w", err)
	}

	if len(report.Report) == 0 {
		return []storage.VolumeGroup{}, nil
	}

	return report.Report[0].VG, nil
}

// GetLogicalVolumes 获取逻辑卷
func (r *LVMRepository) GetLogicalVolumes(ctx context.Context) ([]storage.LogicalVolume, error) {
	args := []string{"-o", "lv_name,vg_name,lv_uuid,lv_size,lv_attr",
		"--reportformat", "json", "--units", "B"}

	result, err := r.cmdExecutor.Execute(ctx, LVS, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute lvs command: %w", err)
	}

	var report LVMReport
	if err := json.Unmarshal([]byte(result.Stdout), &report); err != nil {
		return nil, fmt.Errorf("failed to parse lvs output: %w", err)
	}

	if len(report.Report) == 0 {
		return []storage.LogicalVolume{}, nil
	}

	return report.Report[0].LV, nil
}

// CreatePhysicalVolume 创建物理卷
func (r *LVMRepository) CreatePhysicalVolume(ctx context.Context, devices []string) error {
	args := append(devices, "-y", "-f")

	_, err := r.cmdExecutor.Execute(ctx, PVCREATE, args...)
	if err != nil {
		return fmt.Errorf("failed to create physical volume: %w", err)
	}

	return nil
}

// CreateVolumeGroup 创建卷组
func (r *LVMRepository) CreateVolumeGroup(ctx context.Context, vgName string, pvDevices []string) error {
	// 先检查并创建物理卷
	for _, device := range pvDevices {
		if !r.isPVExists(ctx, device) {
			if err := r.CreatePhysicalVolume(ctx, []string{device}); err != nil {
				return fmt.Errorf("failed to create PV for device %s: %w", device, err)
			}
		}
	}

	// 创建卷组
	args := append([]string{vgName}, pvDevices...)
	args = append(args, "-y", "-f")

	_, err := r.cmdExecutor.Execute(ctx, VGCREATE, args...)
	if err != nil {
		return fmt.Errorf("failed to create volume group: %w", err)
	}

	return nil
}

// CreateLogicalVolume 创建逻辑卷
func (r *LVMRepository) CreateLogicalVolume(ctx context.Context, vgName, lvName, size string) error {
	var args []string
	if size == "all" {
		args = []string{"-l", "100%VG", "-n", lvName, vgName}
	} else {
		args = []string{"-L", size, "-n", lvName, vgName}
	}

	_, err := r.cmdExecutor.Execute(ctx, LVCREATE, args...)
	if err != nil {
		return fmt.Errorf("failed to create logical volume: %w", err)
	}

	return nil
}

// RemovePhysicalVolume 删除物理卷
func (r *LVMRepository) RemovePhysicalVolume(ctx context.Context, devices []string) error {
	args := append(devices, "-y", "-f")

	_, err := r.cmdExecutor.Execute(ctx, PVREMOVE, args...)
	if err != nil {
		return fmt.Errorf("failed to remove physical volume: %w", err)
	}

	return nil
}

// RemoveVolumeGroup 删除卷组
func (r *LVMRepository) RemoveVolumeGroup(ctx context.Context, vgName string) error {
	args := []string{vgName, "-y"}

	_, err := r.cmdExecutor.Execute(ctx, VGREMOVE, args...)
	if err != nil {
		return fmt.Errorf("failed to remove volume group: %w", err)
	}

	return nil
}

// RemoveLogicalVolume 删除逻辑卷
func (r *LVMRepository) RemoveLogicalVolume(ctx context.Context, vgName, lvName string) error {
	args := []string{vgName + "/" + lvName, "-y"}

	_, err := r.cmdExecutor.Execute(ctx, LVREMOVE, args...)
	if err != nil {
		return fmt.Errorf("failed to remove logical volume: %w", err)
	}

	return nil
}

// isPVExists 检查物理卷是否存在
func (r *LVMRepository) isPVExists(ctx context.Context, device string) bool {
	pvs, err := r.GetPhysicalVolumes(ctx)
	if err != nil {
		return false
	}

	for _, pv := range pvs {
		if pv.PVName == device {
			return true
		}
	}
	return false
}