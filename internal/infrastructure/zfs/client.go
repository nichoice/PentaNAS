package zfs

import (
	"encoding/json"
	"fmt"
	"pnas/internal/utils"
	"strconv"
	"strings"
	"time"
)

// Client defines the interface for ZFS operations
type Client interface {
	// Pool operations
	CreatePool(name string, vdevs []VDevSpec, opts PoolOptions) error
	DestroyPool(name string, force bool) error
	ListPools() ([]PoolInfo, error)
	GetPoolStatus(name string) (*PoolStatus, error)
	ExportPool(name string) error
	ImportPool(name string, opts ImportOptions) error
	ScrubPool(name string) error
	GetScrubStatus(name string) (*ScrubStatus, error)

	// Dataset operations
	CreateDataset(name string, opts DatasetOptions) error
	DestroyDataset(name string, recursive bool) error
	ListDatasets(pool string) ([]DatasetInfo, error)
	GetDatasetProperties(name string) (*DatasetProperties, error)
	SetDatasetProperty(name string, property string, value string) error
	MountDataset(name string) error
	UnmountDataset(name string, force bool) error

	// Volume (zvol) operations
	CreateVolume(name string, size uint64, opts VolumeOptions) error
	DestroyVolume(name string) error
	ListVolumes(pool string) ([]VolumeInfo, error)
	ResizeVolume(name string, newSize uint64) error

	// Snapshot operations
	CreateSnapshot(dataset string, snapName string, recursive bool) error
	DestroySnapshot(name string, recursive bool) error
	ListSnapshots(dataset string) ([]SnapshotInfo, error)
	RollbackSnapshot(name string) error
	CloneSnapshot(snapshot string, target string) error

	// Encryption operations
	LoadKey(dataset string, keyLocation string) error
	UnloadKey(dataset string, recursive bool) error
	ChangeKey(dataset string, keyLocation string) error
	GetKeyStatus(dataset string) (string, error)

	// Cache operations
	AddCache(pool string, device string) error
	RemoveCache(pool string, device string) error
	AddLog(pool string, device string) error
	RemoveLog(pool string, device string) error

	// Property operations
	GetProperty(dataset string, property string) (string, error)
	SetProperty(dataset string, property string, value string) error
	InheritProperty(dataset string, property string) error
}

// ZFSClient implements the Client interface
type ZFSClient struct{}

// NewZFSClient creates a new ZFS client
func NewZFSClient() Client {
	return &ZFSClient{}
}

// VDevSpec represents a vdev specification
type VDevSpec struct {
	Type    string   `json:"type"`    // mirror, raidz, raidz2, raidz3, stripe, cache, log
	Devices []string `json:"devices"` // /dev/sdb, /dev/sdc
}

// PoolOptions represents pool creation options
type PoolOptions struct {
	Ashift      int               `json:"ashift"`       // 9, 12, 13
	Features    []string          `json:"features"`
	Properties  map[string]string `json:"properties"`
	Mountpoint  string            `json:"mountpoint"`
	Force       bool              `json:"force"`
}

// ImportOptions represents pool import options
type ImportOptions struct {
	Force      bool   `json:"force"`
	Directory  string `json:"directory"`  // Search directory for pools
	AltRoot    string `json:"alt_root"`
	Readonly   bool   `json:"readonly"`
}

// DatasetOptions represents dataset creation options
type DatasetOptions struct {
	Type           string            `json:"type"`            // filesystem, volume
	Mountpoint     string            `json:"mountpoint"`
	Quota          uint64            `json:"quota"`
	Reservation    uint64            `json:"reservation"`
	Compression    string            `json:"compression"`
	Dedup          string            `json:"dedup"`
	Encryption     string            `json:"encryption"`
	KeyLocation    string            `json:"key_location"`
	KeyFormat      string            `json:"key_format"`     // passphrase, raw, hex
	Properties     map[string]string `json:"properties"`
	CreateParents  bool              `json:"create_parents"`
}

// VolumeOptions represents volume creation options
type VolumeOptions struct {
	BlockSize   int               `json:"block_size"`
	Sparse      bool              `json:"sparse"`
	Compression string            `json:"compression"`
	Dedup       string            `json:"dedup"`
	Encryption  string            `json:"encryption"`
	KeyLocation string            `json:"key_location"`
	Properties  map[string]string `json:"properties"`
}

// PoolInfo represents basic pool information
type PoolInfo struct {
	Name        string  `json:"name"`
	Size        uint64  `json:"size"`
	Allocated   uint64  `json:"allocated"`
	Free        uint64  `json:"free"`
	Capacity    float64 `json:"capacity"`
	Health      string  `json:"health"`
	Dedup       string  `json:"dedup"`
	AltRoot     string  `json:"altroot"`
	Version     string  `json:"version"`
}

// PoolStatus represents detailed pool status
type PoolStatus struct {
	Name   string      `json:"name"`
	State  string      `json:"state"`
	Status string      `json:"status"`
	Action string      `json:"action"`
	Scan   *ScanInfo   `json:"scan"`
	Config []VDevInfo  `json:"config"`
	Errors []ErrorInfo `json:"errors"`
}

// VDevInfo represents vdev information
type VDevInfo struct {
	Name    string      `json:"name"`
	Type    string      `json:"type"`
	State   string      `json:"state"`
	Read    int         `json:"read"`
	Write   int         `json:"write"`
	Cksum   int         `json:"cksum"`
	Children []VDevInfo `json:"children,omitempty"`
}

// ScanInfo represents scrub/resilver scan information
type ScanInfo struct {
	Function   string    `json:"function"`    // scrub, resilver
	State      string    `json:"state"`       // scanning, finished, canceled
	StartTime  time.Time `json:"start_time"`
	EndTime    time.Time `json:"end_time"`
	Scanned    uint64    `json:"scanned"`
	ToScan     uint64    `json:"to_scan"`
	Errors     int       `json:"errors"`
	Repaired   uint64    `json:"repaired"`
	Progress   float64   `json:"progress"`
}

// ScrubStatus represents scrub operation status
type ScrubStatus struct {
	State       string    `json:"state"`
	Progress    float64   `json:"progress"`
	Scanned     uint64    `json:"scanned"`
	ToScan      uint64    `json:"to_scan"`
	Errors      int       `json:"errors"`
	Repaired    uint64    `json:"repaired"`
	StartTime   time.Time `json:"start_time"`
	EndTime     *time.Time `json:"end_time"`
	Duration    int64     `json:"duration"`
}

// ErrorInfo represents pool errors
type ErrorInfo struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}

// DatasetInfo represents dataset information
type DatasetInfo struct {
	Name        string  `json:"name"`
	Type        string  `json:"type"`
	Used        uint64  `json:"used"`
	Available   uint64  `json:"available"`
	Refer       uint64  `json:"refer"`
	Mountpoint  string  `json:"mountpoint"`
	Compression string  `json:"compression"`
	Quota       uint64  `json:"quota"`
}

// DatasetProperties represents all dataset properties
type DatasetProperties struct {
	Name            string            `json:"name"`
	Type            string            `json:"type"`
	Used            uint64            `json:"used"`
	Available       uint64            `json:"available"`
	Referenced      uint64            `json:"referenced"`
	Mountpoint      string            `json:"mountpoint"`
	Mounted         bool              `json:"mounted"`
	Compression     string            `json:"compression"`
	CompressRatio   string            `json:"compressratio"`
	Dedup           string            `json:"dedup"`
	Encryption      string            `json:"encryption"`
	EncryptionRoot  string            `json:"encryptionroot"`
	KeyStatus       string            `json:"keystatus"`
	Quota           uint64            `json:"quota"`
	Reservation     uint64            `json:"reservation"`
	RecordSize      int               `json:"recordsize"`
	ReadOnly        bool              `json:"readonly"`
	Atime           bool              `json:"atime"`
	Sync            string            `json:"sync"`
	AllProperties   map[string]string `json:"all_properties"`
}

// VolumeInfo represents volume information
type VolumeInfo struct {
	Name       string `json:"name"`
	Size       uint64 `json:"size"`
	Used       uint64 `json:"used"`
	Available  uint64 `json:"available"`
	VolSize    uint64 `json:"volsize"`
	VolBlock   int    `json:"volblocksize"`
	DevicePath string `json:"device_path"`
}

// SnapshotInfo represents snapshot information
type SnapshotInfo struct {
	Name       string    `json:"name"`
	Used       uint64    `json:"used"`
	Referenced uint64    `json:"referenced"`
	CreateTime time.Time `json:"create_time"`
	Dataset    string    `json:"dataset"`
	SnapName   string    `json:"snap_name"`
}

// Helper function to execute zpool commands
func (c *ZFSClient) execZpool(args ...string) (*utils.CmdResult, error) {
	result := utils.ExecCommand(utils.ExecOptions{
		Timeout: 60 * time.Second,
	}, "zpool", args...)

	if result.Error != nil {
		return result, fmt.Errorf("zpool command failed: %w, stderr: %s", result.Error, result.Stderr)
	}

	return result, nil
}

// Helper function to execute zfs commands
func (c *ZFSClient) execZfs(args ...string) (*utils.CmdResult, error) {
	result := utils.ExecCommand(utils.ExecOptions{
		Timeout: 60 * time.Second,
	}, "zfs", args...)

	if result.Error != nil {
		return result, fmt.Errorf("zfs command failed: %w, stderr: %s", result.Error, result.Stderr)
	}

	return result, nil
}

// Helper to parse size strings (e.g., "1G", "500M")
func parseSize(sizeStr string) (uint64, error) {
	sizeStr = strings.TrimSpace(sizeStr)
	if sizeStr == "-" || sizeStr == "" {
		return 0, nil
	}

	// Remove units and parse
	multiplier := uint64(1)
	if strings.HasSuffix(sizeStr, "T") {
		multiplier = 1024 * 1024 * 1024 * 1024
		sizeStr = strings.TrimSuffix(sizeStr, "T")
	} else if strings.HasSuffix(sizeStr, "G") {
		multiplier = 1024 * 1024 * 1024
		sizeStr = strings.TrimSuffix(sizeStr, "G")
	} else if strings.HasSuffix(sizeStr, "M") {
		multiplier = 1024 * 1024
		sizeStr = strings.TrimSuffix(sizeStr, "M")
	} else if strings.HasSuffix(sizeStr, "K") {
		multiplier = 1024
		sizeStr = strings.TrimSuffix(sizeStr, "K")
	}

	val, err := strconv.ParseFloat(sizeStr, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse size: %w", err)
	}

	return uint64(val * float64(multiplier)), nil
}

// Helper to format size to human readable
func formatSize(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// Helper to convert property value to bool
func parseBool(value string) bool {
	return value == "on" || value == "yes" || value == "true"
}

// Helper to convert bool to ZFS property value
func boolToZFSValue(value bool) string {
	if value {
		return "on"
	}
	return "off"
}
