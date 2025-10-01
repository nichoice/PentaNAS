package dto

import "time"

// Pool DTOs
type CreatePoolRequest struct {
	Name        string            `json:"name" binding:"required"`
	VDevs       []VDevSpec        `json:"vdevs" binding:"required"`
	Ashift      int               `json:"ashift"`                      // 9, 12, 13
	Compression string            `json:"compression"`                 // lz4, gzip, zstd
	Mountpoint  string            `json:"mountpoint"`
	Properties  map[string]string `json:"properties"`
	Force       bool              `json:"force"`
}

type VDevSpec struct {
	Type    string   `json:"type" binding:"required"`    // mirror, raidz, raidz2, raidz3, cache, log
	Devices []string `json:"devices" binding:"required"` // /dev/sdb, /dev/sdc
}

type PoolResponse struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Size        uint64            `json:"size"`
	Allocated   uint64            `json:"allocated"`
	Free        uint64            `json:"free"`
	Capacity    float64           `json:"capacity"`
	Health      string            `json:"health"`
	Dedup       string            `json:"dedup"`
	Compression string            `json:"compression"`
	Status      string            `json:"status"`
	VDevs       string            `json:"vdevs"`
	Properties  map[string]string `json:"properties,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
}

type PoolStatusResponse struct {
	Name   string          `json:"name"`
	State  string          `json:"state"`
	Status string          `json:"status"`
	Action string          `json:"action"`
	Scan   *ScanInfo       `json:"scan,omitempty"`
	Config []VDevInfo      `json:"config"`
	Errors []ErrorInfo     `json:"errors"`
}

type ScanInfo struct {
	Function  string    `json:"function"`
	State     string    `json:"state"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Scanned   uint64    `json:"scanned"`
	ToScan    uint64    `json:"to_scan"`
	Errors    int       `json:"errors"`
	Repaired  uint64    `json:"repaired"`
	Progress  float64   `json:"progress"`
}

type VDevInfo struct {
	Name     string      `json:"name"`
	Type     string      `json:"type"`
	State    string      `json:"state"`
	Read     int         `json:"read"`
	Write    int         `json:"write"`
	Cksum    int         `json:"cksum"`
	Children []VDevInfo  `json:"children,omitempty"`
}

type ErrorInfo struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}

// Dataset DTOs
type CreateDatasetRequest struct {
	Name          string            `json:"name" binding:"required"`
	Type          string            `json:"type"`                      // filesystem (default), volume
	Mountpoint    string            `json:"mountpoint"`
	Quota         uint64            `json:"quota"`
	Reservation   uint64            `json:"reservation"`
	Compression   string            `json:"compression"`               // lz4, gzip, zstd, off
	Dedup         string            `json:"dedup"`                     // on, off
	Encryption    string            `json:"encryption"`                // aes-256-gcm, aes-256-ccm
	KeyLocation   string            `json:"key_location"`
	KeyFormat     string            `json:"key_format"`                // raw, hex, passphrase
	Properties    map[string]string `json:"properties"`
	CreateParents bool              `json:"create_parents"`
}

type UpdateDatasetRequest struct {
	Quota       *uint64           `json:"quota"`
	Reservation *uint64           `json:"reservation"`
	Compression *string           `json:"compression"`
	ReadOnly    *bool             `json:"read_only"`
	Atime       *bool             `json:"atime"`
	Properties  map[string]string `json:"properties"`
}

type DatasetResponse struct {
	ID             string            `json:"id"`
	Name           string            `json:"name"`
	Pool           string            `json:"pool"`
	Type           string            `json:"type"`
	Mountpoint     string            `json:"mountpoint"`
	Quota          uint64            `json:"quota"`
	Reservation    uint64            `json:"reservation"`
	Used           uint64            `json:"used"`
	Available      uint64            `json:"available"`
	Compression    string            `json:"compression"`
	CompressRatio  string            `json:"compress_ratio"`
	Dedup          string            `json:"dedup"`
	Encryption     string            `json:"encryption"`
	KeyStatus      string            `json:"key_status"`
	ReadOnly       bool              `json:"read_only"`
	Atime          bool              `json:"atime"`
	RecordSize     int               `json:"record_size"`
	Status         string            `json:"status"`
	AllProperties  map[string]string `json:"all_properties,omitempty"`
	CreatedAt      time.Time         `json:"created_at"`
}

// Volume DTOs
type CreateVolumeRequest struct {
	Name        string            `json:"name" binding:"required"`
	Size        uint64            `json:"size" binding:"required"`
	BlockSize   int               `json:"block_size"`
	Sparse      bool              `json:"sparse"`
	Compression string            `json:"compression"`
	Dedup       string            `json:"dedup"`
	Encryption  string            `json:"encryption"`
	KeyLocation string            `json:"key_location"`
	Properties  map[string]string `json:"properties"`
}

type ResizeVolumeRequest struct {
	NewSize uint64 `json:"new_size" binding:"required"`
}

type VolumeResponse struct {
	ID             string            `json:"id"`
	Name           string            `json:"name"`
	Pool           string            `json:"pool"`
	Size           uint64            `json:"size"`
	BlockSize      int               `json:"block_size"`
	Used           uint64            `json:"used"`
	Available      uint64            `json:"available"`
	Compression    string            `json:"compression"`
	Dedup          string            `json:"dedup"`
	Encryption     string            `json:"encryption"`
	KeyStatus      string            `json:"key_status"`
	Sparse         bool              `json:"sparse"`
	DevicePath     string            `json:"device_path"`
	Status         string            `json:"status"`
	CreatedAt      time.Time         `json:"created_at"`
}

// Snapshot DTOs
type CreateSnapshotRequest struct {
	Dataset   string `json:"dataset" binding:"required"`
	SnapName  string `json:"snap_name" binding:"required"`
	Recursive bool   `json:"recursive"`
	Comment   string `json:"comment"`
}

type RollbackSnapshotRequest struct {
	Snapshot string `json:"snapshot" binding:"required"`
}

type CloneSnapshotRequest struct {
	Snapshot string `json:"snapshot" binding:"required"`
	Target   string `json:"target" binding:"required"`
}

type SnapshotResponse struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Pool       string    `json:"pool"`
	Dataset    string    `json:"dataset"`
	SnapName   string    `json:"snap_name"`
	Used       uint64    `json:"used"`
	Referenced uint64    `json:"referenced"`
	SnapTime   time.Time `json:"snap_time"`
	Comment    string    `json:"comment"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
}

// Encryption DTOs
type LoadKeyRequest struct {
	Dataset     string `json:"dataset" binding:"required"`
	KeyLocation string `json:"key_location"` // File path or "prompt"
	Passphrase  string `json:"passphrase"`   // If using passphrase
}

type UnloadKeyRequest struct {
	Dataset   string `json:"dataset" binding:"required"`
	Recursive bool   `json:"recursive"`
}

type ChangeKeyRequest struct {
	Dataset        string `json:"dataset" binding:"required"`
	NewKeyLocation string `json:"new_key_location"`
	NewPassphrase  string `json:"new_passphrase"`
}

type KeyStatusResponse struct {
	Dataset          string `json:"dataset"`
	Encryption       string `json:"encryption"`
	EncryptionRoot   string `json:"encryption_root"`
	KeyStatus        string `json:"key_status"`
	KeyLocation      string `json:"key_location"`
	KeyFormat        string `json:"key_format"`
	Loaded           bool   `json:"loaded"`
}

// Scrub DTOs
type ScrubStatusResponse struct {
	PoolName     string     `json:"pool_name"`
	State        string     `json:"state"`
	Progress     float64    `json:"progress"`
	Scanned      uint64     `json:"scanned"`
	ToScan       uint64     `json:"to_scan"`
	Errors       int        `json:"errors"`
	Repaired     uint64     `json:"repaired"`
	StartTime    time.Time  `json:"start_time"`
	EndTime      *time.Time `json:"end_time,omitempty"`
	Duration     int64      `json:"duration"`
}

// Cache DTOs
type AddCacheRequest struct {
	Pool   string `json:"pool" binding:"required"`
	Device string `json:"device" binding:"required"`
	Type   string `json:"type" binding:"required"` // cache (L2ARC) or log (SLOG)
}

type RemoveCacheRequest struct {
	Pool   string `json:"pool" binding:"required"`
	Device string `json:"device" binding:"required"`
}

// Common response
type MessageResponse struct {
	Message string `json:"message"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
}

// List request with pagination
type ListRequest struct {
	Pool     string `form:"pool"`
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
	Status   string `form:"status"`
}

// List response with pagination
type ListResponse struct {
	Items      interface{} `json:"items"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalPages int         `json:"total_pages"`
}
