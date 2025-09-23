package services

import (
	"context"
	"pnas/internal/app/dto"
	"pnas/internal/domain/storage"
)

// StorageService 存储应用服务
type StorageService struct {
	storageService *storage.Service
}

// NewStorageService 创建存储应用服务
func NewStorageService(storageService *storage.Service) *StorageService {
	return &StorageService{
		storageService: storageService,
	}
}

// GetDisks 获取磁盘列表
func (s *StorageService) GetDisks(ctx context.Context) ([]dto.DiskResponse, error) {
	disks, err := s.storageService.ScanDisks(ctx)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.DiskResponse, len(disks))
	for i, disk := range disks {
		responses[i] = dto.DiskResponse{
			Name:         disk.Name,
			Path:         disk.Path,
			Size:         disk.Size,
			UsedSize:     disk.UsedSize,
			AvailSize:    disk.AvailSize,
			UsageRate:    disk.UsageRate,
			Serial:       disk.Serial,
			Rota:         disk.Rota,
			Model:        disk.Model,
			Vendor:       disk.Vendor,
			Type:         disk.Type,
			MajMin:       disk.MajMin,
			MountPoint:   disk.MountPoint,
			FileSystem:   disk.FileSystem,
			IsSystemDisk: disk.IsSystemDisk,
			IsOnline:     disk.IsOnline,
			Temperature:  disk.Temperature,
			Health:       disk.Health,
		}
	}

	return responses, nil
}

// GetVolumeGroups 获取卷组列表
func (s *StorageService) GetVolumeGroups(ctx context.Context) ([]dto.VolumeGroupResponse, error) {
	vgs, err := s.storageService.GetVolumeGroups(ctx)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.VolumeGroupResponse, len(vgs))
	for i, vg := range vgs {
		responses[i] = dto.VolumeGroupResponse{
			VGName:         vg.VGName,
			VGUUID:         vg.VGUUID,
			VGAttr:         vg.VGAttr,
			VGSize:         vg.VGSize,
			VGFree:         vg.VGFree,
			VGExtentCount:  vg.VGExtentCount,
			VGExtentSize:   vg.VGExtentSize,
			VGFreeCount:    vg.VGFreeCount,
		}
	}

	return responses, nil
}

// GetLogicalVolumes 获取逻辑卷列表
func (s *StorageService) GetLogicalVolumes(ctx context.Context) ([]dto.LogicalVolumeResponse, error) {
	lvs, err := s.storageService.GetLogicalVolumes(ctx)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.LogicalVolumeResponse, len(lvs))
	for i, lv := range lvs {
		responses[i] = dto.LogicalVolumeResponse{
			LVName: lv.LVName,
			VGName: lv.VGName,
			LVUUID: lv.LVUUID,
			LVSize: lv.LVSize,
			LVAttr: lv.LVAttr,
		}
	}

	return responses, nil
}

// CreateVolumeGroup 创建卷组
func (s *StorageService) CreateVolumeGroup(ctx context.Context, req dto.CreateVGRequest) error {
	return s.storageService.CreateVolumeGroup(ctx, req.VgName, req.Devices)
}

// CreateLogicalVolume 创建逻辑卷
func (s *StorageService) CreateLogicalVolume(ctx context.Context, req dto.CreateLVRequest) error {
	return s.storageService.CreateLogicalVolume(ctx, req.VgName, req.LvName, req.Size)
}

// GetStorageOverview 获取存储概览
func (s *StorageService) GetStorageOverview(ctx context.Context) (*dto.StorageOverviewResponse, error) {
	overview, err := s.storageService.GetStorageOverview(ctx)
	if err != nil {
		return nil, err
	}

	// 转换磁盘数据
	diskResponses := make([]dto.DiskResponse, len(overview.Disks))
	for i, disk := range overview.Disks {
		diskResponses[i] = dto.DiskResponse{
			Name:         disk.Name,
			Path:         disk.Path,
			Size:         disk.Size,
			UsedSize:     disk.UsedSize,
			AvailSize:    disk.AvailSize,
			UsageRate:    disk.UsageRate,
			Serial:       disk.Serial,
			Rota:         disk.Rota,
			Model:        disk.Model,
			Vendor:       disk.Vendor,
			Type:         disk.Type,
			MajMin:       disk.MajMin,
			MountPoint:   disk.MountPoint,
			FileSystem:   disk.FileSystem,
			IsSystemDisk: disk.IsSystemDisk,
			IsOnline:     disk.IsOnline,
			Temperature:  disk.Temperature,
			Health:       disk.Health,
		}
	}

	// 转换卷组数据
	vgResponses := make([]dto.VolumeGroupResponse, len(overview.VolumeGroups))
	for i, vg := range overview.VolumeGroups {
		vgResponses[i] = dto.VolumeGroupResponse{
			VGName:         vg.VGName,
			VGUUID:         vg.VGUUID,
			VGAttr:         vg.VGAttr,
			VGSize:         vg.VGSize,
			VGFree:         vg.VGFree,
			VGExtentCount:  vg.VGExtentCount,
			VGExtentSize:   vg.VGExtentSize,
			VGFreeCount:    vg.VGFreeCount,
		}
	}

	// 转换逻辑卷数据
	lvResponses := make([]dto.LogicalVolumeResponse, len(overview.LogicalVolumes))
	for i, lv := range overview.LogicalVolumes {
		lvResponses[i] = dto.LogicalVolumeResponse{
			LVName: lv.LVName,
			VGName: lv.VGName,
			LVUUID: lv.LVUUID,
			LVSize: lv.LVSize,
			LVAttr: lv.LVAttr,
		}
	}

	return &dto.StorageOverviewResponse{
		TotalDisks:          overview.TotalDisks,
		TotalVolumeGroups:   overview.TotalVolumeGroups,
		TotalLogicalVolumes: overview.TotalLogicalVolumes,
		Disks:               diskResponses,
		VolumeGroups:        vgResponses,
		LogicalVolumes:      lvResponses,
	}, nil
}