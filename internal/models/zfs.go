package models

import (
	"time"
)

// ZFSPool represents a ZFS storage pool
type ZFSPool struct {
	Base
	Name        string  `gorm:"uniqueIndex;not null" json:"name"`
	Size        uint64  `json:"size"`         // Total size in bytes
	Allocated   uint64  `json:"allocated"`    // Allocated space in bytes
	Free        uint64  `json:"free"`         // Free space in bytes
	Health      string  `json:"health"`       // ONLINE, DEGRADED, FAULTED, OFFLINE, UNAVAIL
	Capacity    float64 `json:"capacity"`     // Usage percentage (0-100)
	Dedup       string  `json:"dedup"`        // on, off, verify
	Compression string  `json:"compression"`  // lz4, gzip, zstd, off
	Ashift      int     `json:"ashift"`       // Sector size (9=512B, 12=4K, 13=8K)
	AutoExpand  bool    `json:"auto_expand"`  // Automatically expand pool
	Comment     string  `json:"comment"`
	VDevs       string  `gorm:"type:text" json:"vdevs"` // JSON encoded vdev structure
	Status      string  `json:"status"`       // active, inactive, destroyed
}

// ZFSDataset represents a ZFS dataset (filesystem)
type ZFSDataset struct {
	Base
	Name            string  `gorm:"uniqueIndex;not null" json:"name"` // pool/dataset/child
	Pool            string  `gorm:"index;not null" json:"pool"`
	Type            string  `json:"type"`             // filesystem, volume, snapshot
	Mountpoint      string  `json:"mountpoint"`
	Quota           uint64  `json:"quota"`            // Quota in bytes (0 = none)
	Reservation     uint64  `json:"reservation"`      // Reservation in bytes
	Used            uint64  `json:"used"`             // Used space in bytes
	Available       uint64  `json:"available"`        // Available space in bytes
	Compression     string  `json:"compression"`      // lz4, gzip-N, zstd, off, inherit
	Dedup           string  `json:"dedup"`            // on, off, verify, inherit
	Encryption      string  `json:"encryption"`       // aes-256-gcm, aes-256-ccm, off, inherit
	EncryptionRoot  string  `json:"encryption_root"`  // Root dataset for encryption
	KeyStatus       string  `json:"key_status"`       // available, unavailable
	ReadOnly        bool    `json:"read_only"`
	Atime           bool    `json:"atime"`            // Access time updates
	RecordSize      int     `json:"record_size"`      // Record size (128K default)
	Sync            string  `json:"sync"`             // standard, always, disabled
	Snapdir         string  `json:"snapdir"`          // hidden, visible
	ShareNFS        string  `json:"share_nfs"`
	ShareSMB        string  `json:"share_smb"`
	Comment         string  `json:"comment"`
	Status          string  `json:"status"`           // active, inactive, destroyed
}

// ZFSVolume represents a ZFS volume (block device)
type ZFSVolume struct {
	Base
	Name            string  `gorm:"uniqueIndex;not null" json:"name"` // pool/volume
	Pool            string  `gorm:"index;not null" json:"pool"`
	Size            uint64  `gorm:"not null" json:"size"`            // Volume size in bytes
	BlockSize       int     `json:"block_size"`                       // Block size (8K default)
	Used            uint64  `json:"used"`
	Available       uint64  `json:"available"`
	Compression     string  `json:"compression"`
	Dedup           string  `json:"dedup"`
	Encryption      string  `json:"encryption"`
	EncryptionRoot  string  `json:"encryption_root"`
	KeyStatus       string  `json:"key_status"`
	Sync            string  `json:"sync"`
	Sparse          bool    `json:"sparse"`           // Sparse volume
	DevicePath      string  `json:"device_path"`      // /dev/zvol/pool/volume
	Comment         string  `json:"comment"`
	Status          string  `json:"status"`
}

// ZFSSnapshot represents a ZFS snapshot
type ZFSSnapshot struct {
	Base
	Name       string    `gorm:"uniqueIndex;not null" json:"name"` // pool/dataset@snapshot
	Pool       string    `gorm:"index;not null" json:"pool"`
	Dataset    string    `gorm:"index;not null" json:"dataset"`
	SnapName   string    `gorm:"index;not null" json:"snap_name"`
	Used       uint64    `json:"used"`
	Referenced uint64    `json:"referenced"`
	SnapTime   time.Time `json:"snap_time"`
	Comment    string    `json:"comment"`
	Status     string    `json:"status"`
}

// ZFSCache represents L2ARC (SSD cache) configuration
type ZFSCache struct {
	Base
	PoolName   string `gorm:"index;not null" json:"pool_name"`
	Device     string `gorm:"not null" json:"device"`        // /dev/sdb
	Size       uint64 `json:"size"`
	Type       string `json:"type"`                          // cache (L2ARC), log (SLOG)
	Health     string `json:"health"`
	Comment    string `json:"comment"`
	Status     string `json:"status"`                        // active, inactive, removed
}

// ZFSEncryptionKey represents encryption key metadata (not the actual key)
type ZFSEncryptionKey struct {
	Base
	DatasetName     string    `gorm:"uniqueIndex;not null" json:"dataset_name"`
	Format          string    `json:"format"`           // passphrase, raw, hex
	Location        string    `json:"location"`         // prompt, file://path
	KeyStatus       string    `json:"key_status"`       // available, unavailable
	EncryptionAlgo  string    `json:"encryption_algo"`  // aes-256-gcm, aes-256-ccm
	Loaded          bool      `json:"loaded"`
	LoadedAt        *time.Time `json:"loaded_at"`
	UnloadedAt      *time.Time `json:"unloaded_at"`
	Comment         string    `json:"comment"`
}

// ZFSScrub represents scrub operation status
type ZFSScrub struct {
	Base
	PoolName       string     `gorm:"index;not null" json:"pool_name"`
	State          string     `json:"state"`          // running, completed, canceled, paused
	StartTime      time.Time  `json:"start_time"`
	EndTime        *time.Time `json:"end_time"`
	Duration       int64      `json:"duration"`       // seconds
	BytesScanned   uint64     `json:"bytes_scanned"`
	BytesToScan    uint64     `json:"bytes_to_scan"`
	ErrorsFound    int        `json:"errors_found"`
	ErrorsRepaired int        `json:"errors_repaired"`
	Progress       float64    `json:"progress"`       // 0-100
}

// TableName overrides
func (ZFSPool) TableName() string {
	return "zfs_pools"
}

func (ZFSDataset) TableName() string {
	return "zfs_datasets"
}

func (ZFSVolume) TableName() string {
	return "zfs_volumes"
}

func (ZFSSnapshot) TableName() string {
	return "zfs_snapshots"
}

func (ZFSCache) TableName() string {
	return "zfs_caches"
}

func (ZFSEncryptionKey) TableName() string {
	return "zfs_encryption_keys"
}

func (ZFSScrub) TableName() string {
	return "zfs_scrubs"
}
