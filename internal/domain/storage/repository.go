package storage

import "context"

// DiskRepository 磁盘仓储接口
type DiskRepository interface {
	GetAll(ctx context.Context) ([]Disk, error)
	GetByID(ctx context.Context, id string) (*Disk, error)
	GetByPath(ctx context.Context, path string) (*Disk, error)
	Save(ctx context.Context, disk *Disk) error
	Delete(ctx context.Context, id string) error
}

// VolumeRepository LVM卷管理仓储接口
type VolumeRepository interface {
	GetPhysicalVolumes(ctx context.Context) ([]PhysicalVolume, error)
	GetVolumeGroups(ctx context.Context) ([]VolumeGroup, error)
	GetLogicalVolumes(ctx context.Context) ([]LogicalVolume, error)
	CreatePhysicalVolume(ctx context.Context, devices []string) error
	CreateVolumeGroup(ctx context.Context, vgName string, pvDevices []string) error
	CreateLogicalVolume(ctx context.Context, vgName, lvName, size string) error
	RemovePhysicalVolume(ctx context.Context, devices []string) error
	RemoveVolumeGroup(ctx context.Context, vgName string) error
	RemoveLogicalVolume(ctx context.Context, vgName, lvName string) error
}

// StoragePoolRepository 存储池仓储接口
type StoragePoolRepository interface {
	GetAll(ctx context.Context) ([]StoragePool, error)
	GetByID(ctx context.Context, id string) (*StoragePool, error)
	Save(ctx context.Context, pool *StoragePool) error
	Delete(ctx context.Context, id string) error
}