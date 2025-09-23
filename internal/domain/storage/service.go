package storage

import (
	"context"
	"fmt"
)

// Service 存储领域服务
type Service struct {
	diskRepo        DiskRepository
	volumeRepo      VolumeRepository
	storagePoolRepo StoragePoolRepository
}

// NewService 创建存储服务
func NewService(diskRepo DiskRepository, volumeRepo VolumeRepository, storagePoolRepo StoragePoolRepository) *Service {
	return &Service{
		diskRepo:        diskRepo,
		volumeRepo:      volumeRepo,
		storagePoolRepo: storagePoolRepo,
	}
}

// ScanDisks 扫描系统磁盘
func (s *Service) ScanDisks(ctx context.Context) ([]Disk, error) {
	return s.diskRepo.GetAll(ctx)
}

// GetVolumeGroups 获取卷组信息
func (s *Service) GetVolumeGroups(ctx context.Context) ([]VolumeGroup, error) {
	return s.volumeRepo.GetVolumeGroups(ctx)
}

// GetLogicalVolumes 获取逻辑卷信息
func (s *Service) GetLogicalVolumes(ctx context.Context) ([]LogicalVolume, error) {
	return s.volumeRepo.GetLogicalVolumes(ctx)
}

// CreateVolumeGroup 创建卷组
func (s *Service) CreateVolumeGroup(ctx context.Context, vgName string, devices []string) error {
	// 检查设备是否可用
	for _, device := range devices {
		disk, err := s.diskRepo.GetByPath(ctx, device)
		if err != nil {
			return fmt.Errorf("device %s not found: %w", device, err)
		}
		if disk.IsSystemDisk {
			return fmt.Errorf("cannot use system disk %s for volume group", device)
		}
	}

	// 创建卷组
	return s.volumeRepo.CreateVolumeGroup(ctx, vgName, devices)
}

// CreateLogicalVolume 创建逻辑卷
func (s *Service) CreateLogicalVolume(ctx context.Context, vgName, lvName, size string) error {
	// 检查卷组是否存在
	vgs, err := s.volumeRepo.GetVolumeGroups(ctx)
	if err != nil {
		return fmt.Errorf("failed to get volume groups: %w", err)
	}

	found := false
	for _, vg := range vgs {
		if vg.VGName == vgName {
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("volume group %s not found", vgName)
	}

	return s.volumeRepo.CreateLogicalVolume(ctx, vgName, lvName, size)
}

// GetStorageOverview 获取存储概览
func (s *Service) GetStorageOverview(ctx context.Context) (*StorageOverview, error) {
	disks, err := s.diskRepo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get disks: %w", err)
	}

	vgs, err := s.volumeRepo.GetVolumeGroups(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get volume groups: %w", err)
	}

	lvs, err := s.volumeRepo.GetLogicalVolumes(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get logical volumes: %w", err)
	}

	return &StorageOverview{
		TotalDisks:        len(disks),
		TotalVolumeGroups: len(vgs),
		TotalLogicalVolumes: len(lvs),
		Disks:             disks,
		VolumeGroups:      vgs,
		LogicalVolumes:    lvs,
	}, nil
}

// StorageOverview 存储概览
type StorageOverview struct {
	TotalDisks          int             `json:"total_disks"`
	TotalVolumeGroups   int             `json:"total_volume_groups"`
	TotalLogicalVolumes int             `json:"total_logical_volumes"`
	Disks               []Disk          `json:"disks"`
	VolumeGroups        []VolumeGroup   `json:"volume_groups"`
	LogicalVolumes      []LogicalVolume `json:"logical_volumes"`
}