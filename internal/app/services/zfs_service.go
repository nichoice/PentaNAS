package services

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"pnas/internal/app/dto"
	"pnas/internal/database"
	"pnas/internal/infrastructure/zfs"
	"pnas/internal/models"
)

// ZFSService handles ZFS business logic
type ZFSService struct {
	client zfs.Client
	db     *gorm.DB
}

// NewZFSService creates a new ZFS service instance
func NewZFSService() *ZFSService {
	return &ZFSService{
		client: zfs.NewZFSClient(),
		db:     database.DB,
	}
}

// Pool Management

func (s *ZFSService) CreatePool(req dto.CreatePoolRequest) (*dto.PoolResponse, error) {
	// Convert VDevs from DTO to infrastructure type
	vdevs := make([]zfs.VDevSpec, len(req.VDevs))
	for i, v := range req.VDevs {
		vdevs[i] = zfs.VDevSpec{
			Type:    v.Type,
			Devices: v.Devices,
		}
	}

	// Create pool options
	opts := zfs.PoolOptions{
		Ashift:     req.Ashift,
		Mountpoint: req.Mountpoint,
		Properties: req.Properties,
		Force:      req.Force,
	}
	if req.Compression != "" {
		if opts.Properties == nil {
			opts.Properties = make(map[string]string)
		}
		opts.Properties["compression"] = req.Compression
	}

	// Create pool using infrastructure layer
	if err := s.client.CreatePool(req.Name, vdevs, opts); err != nil {
		return nil, fmt.Errorf("failed to create pool: %w", err)
	}

	// Get pool info to sync with database
	pools, err := s.client.ListPools()
	if err != nil {
		return nil, fmt.Errorf("pool created but failed to get info: %w", err)
	}

	var poolInfo *zfs.PoolInfo
	for i := range pools {
		if pools[i].Name == req.Name {
			poolInfo = &pools[i]
			break
		}
	}
	if poolInfo == nil {
		return nil, fmt.Errorf("pool created but not found in list")
	}

	// Save to database
	vdevsJSON := s.vdevsToString(req.VDevs)
	pool := models.ZFSPool{
		Base:        models.Base{ID: uuid.New().String()},
		Name:        poolInfo.Name,
		Size:        poolInfo.Size,
		Allocated:   poolInfo.Allocated,
		Free:        poolInfo.Free,
		Capacity:    poolInfo.Capacity,
		Health:      poolInfo.Health,
		Dedup:       poolInfo.Dedup,
		Compression: poolInfo.Compression,
		Status:      "ONLINE",
		VDevs:       vdevsJSON,
		Properties:  req.Properties,
	}

	if err := s.db.Create(&pool).Error; err != nil {
		return nil, fmt.Errorf("failed to save pool to database: %w", err)
	}

	return s.poolToResponse(&pool), nil
}

func (s *ZFSService) ListPools(req dto.ListRequest) (*dto.ListResponse, error) {
	var pools []models.ZFSPool
	query := s.db.Model(&models.ZFSPool{})

	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}

	// Count total
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count pools: %w", err)
	}

	// Apply pagination
	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	if err := query.Offset(offset).Limit(pageSize).Find(&pools).Error; err != nil {
		return nil, fmt.Errorf("failed to list pools: %w", err)
	}

	// Convert to responses
	items := make([]dto.PoolResponse, len(pools))
	for i, pool := range pools {
		items[i] = *s.poolToResponse(&pool)
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	return &dto.ListResponse{
		Items:      items,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

func (s *ZFSService) GetPoolStatus(name string) (*dto.PoolStatusResponse, error) {
	status, err := s.client.GetPoolStatus(name)
	if err != nil {
		return nil, fmt.Errorf("failed to get pool status: %w", err)
	}

	// Convert to DTO
	resp := &dto.PoolStatusResponse{
		Name:   status.Name,
		State:  status.State,
		Status: status.Status,
		Action: status.Action,
		Config: s.convertVDevInfos(status.VDevs),
		Errors: s.convertErrors(status.Errors),
	}

	if status.Scan != nil {
		resp.Scan = &dto.ScanInfo{
			Function:  status.Scan.Function,
			State:     status.Scan.State,
			StartTime: status.Scan.StartTime,
			EndTime:   status.Scan.EndTime,
			Scanned:   status.Scan.Scanned,
			ToScan:    status.Scan.ToScan,
			Errors:    status.Scan.Errors,
			Repaired:  status.Scan.Repaired,
			Progress:  status.Scan.Progress,
		}
	}

	return resp, nil
}

func (s *ZFSService) DestroyPool(name string, force bool) error {
	// Delete from database
	if err := s.db.Where("name = ?", name).Delete(&models.ZFSPool{}).Error; err != nil {
		return fmt.Errorf("failed to delete pool from database: %w", err)
	}

	// Destroy pool
	if err := s.client.DestroyPool(name, force); err != nil {
		return fmt.Errorf("failed to destroy pool: %w", err)
	}

	return nil
}

func (s *ZFSService) ScrubPool(name string) error {
	if err := s.client.ScrubPool(name); err != nil {
		return fmt.Errorf("failed to start scrub: %w", err)
	}

	// Update database
	if err := s.db.Model(&models.ZFSPool{}).Where("name = ?", name).Update("status", "SCRUBBING").Error; err != nil {
		return fmt.Errorf("failed to update pool status: %w", err)
	}

	return nil
}

// Dataset Management

func (s *ZFSService) CreateDataset(req dto.CreateDatasetRequest) (*dto.DatasetResponse, error) {
	opts := zfs.DatasetOptions{
		Type:          req.Type,
		Mountpoint:    req.Mountpoint,
		Quota:         req.Quota,
		Reservation:   req.Reservation,
		Compression:   req.Compression,
		Dedup:         req.Dedup,
		Encryption:    req.Encryption,
		KeyLocation:   req.KeyLocation,
		KeyFormat:     req.KeyFormat,
		Properties:    req.Properties,
		CreateParents: req.CreateParents,
	}

	if err := s.client.CreateDataset(req.Name, opts); err != nil {
		return nil, fmt.Errorf("failed to create dataset: %w", err)
	}

	// Get dataset info
	datasets, err := s.client.ListDatasets("")
	if err != nil {
		return nil, fmt.Errorf("dataset created but failed to get info: %w", err)
	}

	var datasetInfo *zfs.DatasetInfo
	for i := range datasets {
		if datasets[i].Name == req.Name {
			datasetInfo = &datasets[i]
			break
		}
	}
	if datasetInfo == nil {
		return nil, fmt.Errorf("dataset created but not found in list")
	}

	// Extract pool name
	poolName := s.extractPoolName(req.Name)

	// Save to database
	dataset := models.ZFSDataset{
		Base:          models.Base{ID: uuid.New().String()},
		Name:          datasetInfo.Name,
		Pool:          poolName,
		Type:          datasetInfo.Type,
		Mountpoint:    datasetInfo.Mountpoint,
		Quota:         datasetInfo.Quota,
		Reservation:   datasetInfo.Reservation,
		Used:          datasetInfo.Used,
		Available:     datasetInfo.Available,
		Compression:   datasetInfo.Compression,
		CompressRatio: datasetInfo.CompressRatio,
		Dedup:         datasetInfo.Dedup,
		Encryption:    datasetInfo.Encryption,
		KeyStatus:     datasetInfo.KeyStatus,
		ReadOnly:      datasetInfo.ReadOnly,
		Atime:         datasetInfo.Atime,
		RecordSize:    datasetInfo.RecordSize,
		Status:        "ONLINE",
		AllProperties: req.Properties,
	}

	if err := s.db.Create(&dataset).Error; err != nil {
		return nil, fmt.Errorf("failed to save dataset to database: %w", err)
	}

	return s.datasetToResponse(&dataset), nil
}

func (s *ZFSService) ListDatasets(req dto.ListRequest) (*dto.ListResponse, error) {
	var datasets []models.ZFSDataset
	query := s.db.Model(&models.ZFSDataset{})

	if req.Pool != "" {
		query = query.Where("pool = ?", req.Pool)
	}
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}

	// Count total
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count datasets: %w", err)
	}

	// Apply pagination
	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	if err := query.Offset(offset).Limit(pageSize).Find(&datasets).Error; err != nil {
		return nil, fmt.Errorf("failed to list datasets: %w", err)
	}

	// Convert to responses
	items := make([]dto.DatasetResponse, len(datasets))
	for i, ds := range datasets {
		items[i] = *s.datasetToResponse(&ds)
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	return &dto.ListResponse{
		Items:      items,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

func (s *ZFSService) UpdateDataset(name string, req dto.UpdateDatasetRequest) (*dto.DatasetResponse, error) {
	if req.Quota != nil {
		if err := s.client.SetDatasetProperty(name, "quota", fmt.Sprintf("%d", *req.Quota)); err != nil {
			return nil, fmt.Errorf("failed to set quota: %w", err)
		}
	}
	if req.Reservation != nil {
		if err := s.client.SetDatasetProperty(name, "reservation", fmt.Sprintf("%d", *req.Reservation)); err != nil {
			return nil, fmt.Errorf("failed to set reservation: %w", err)
		}
	}
	if req.Compression != nil {
		if err := s.client.SetDatasetProperty(name, "compression", *req.Compression); err != nil {
			return nil, fmt.Errorf("failed to set compression: %w", err)
		}
	}
	if req.ReadOnly != nil {
		value := "off"
		if *req.ReadOnly {
			value = "on"
		}
		if err := s.client.SetDatasetProperty(name, "readonly", value); err != nil {
			return nil, fmt.Errorf("failed to set readonly: %w", err)
		}
	}
	if req.Atime != nil {
		value := "off"
		if *req.Atime {
			value = "on"
		}
		if err := s.client.SetDatasetProperty(name, "atime", value); err != nil {
			return nil, fmt.Errorf("failed to set atime: %w", err)
		}
	}

	// Update custom properties
	for key, value := range req.Properties {
		if err := s.client.SetDatasetProperty(name, key, value); err != nil {
			return nil, fmt.Errorf("failed to set property %s: %w", key, err)
		}
	}

	// Refresh from ZFS
	datasets, err := s.client.ListDatasets("")
	if err != nil {
		return nil, fmt.Errorf("failed to get updated dataset info: %w", err)
	}

	var datasetInfo *zfs.DatasetInfo
	for i := range datasets {
		if datasets[i].Name == name {
			datasetInfo = &datasets[i]
			break
		}
	}
	if datasetInfo == nil {
		return nil, fmt.Errorf("dataset not found after update")
	}

	// Update database
	updates := map[string]interface{}{
		"quota":          datasetInfo.Quota,
		"reservation":    datasetInfo.Reservation,
		"used":           datasetInfo.Used,
		"available":      datasetInfo.Available,
		"compression":    datasetInfo.Compression,
		"compress_ratio": datasetInfo.CompressRatio,
		"read_only":      datasetInfo.ReadOnly,
		"atime":          datasetInfo.Atime,
	}

	if err := s.db.Model(&models.ZFSDataset{}).Where("name = ?", name).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to update dataset in database: %w", err)
	}

	// Get updated dataset from database
	var dataset models.ZFSDataset
	if err := s.db.Where("name = ?", name).First(&dataset).Error; err != nil {
		return nil, fmt.Errorf("failed to get updated dataset: %w", err)
	}

	return s.datasetToResponse(&dataset), nil
}

func (s *ZFSService) DestroyDataset(name string, recursive bool) error {
	// Delete from database
	if err := s.db.Where("name = ?", name).Delete(&models.ZFSDataset{}).Error; err != nil {
		return fmt.Errorf("failed to delete dataset from database: %w", err)
	}

	// Destroy dataset
	if err := s.client.DestroyDataset(name, recursive); err != nil {
		return fmt.Errorf("failed to destroy dataset: %w", err)
	}

	return nil
}

// Snapshot Management

func (s *ZFSService) CreateSnapshot(req dto.CreateSnapshotRequest) (*dto.SnapshotResponse, error) {
	if err := s.client.CreateSnapshot(req.Dataset, req.SnapName, req.Recursive); err != nil {
		return nil, fmt.Errorf("failed to create snapshot: %w", err)
	}

	fullName := fmt.Sprintf("%s@%s", req.Dataset, req.SnapName)
	poolName := s.extractPoolName(req.Dataset)

	// Save to database
	snapshot := models.ZFSSnapshot{
		Base:     models.Base{ID: uuid.New().String()},
		Name:     fullName,
		Pool:     poolName,
		Dataset:  req.Dataset,
		SnapName: req.SnapName,
		SnapTime: time.Now(),
		Comment:  req.Comment,
		Status:   "ACTIVE",
	}

	if err := s.db.Create(&snapshot).Error; err != nil {
		return nil, fmt.Errorf("failed to save snapshot to database: %w", err)
	}

	return s.snapshotToResponse(&snapshot), nil
}

func (s *ZFSService) ListSnapshots(req dto.ListRequest) (*dto.ListResponse, error) {
	var snapshots []models.ZFSSnapshot
	query := s.db.Model(&models.ZFSSnapshot{})

	if req.Pool != "" {
		query = query.Where("pool = ?", req.Pool)
	}
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}

	// Count total
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count snapshots: %w", err)
	}

	// Apply pagination
	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	if err := query.Offset(offset).Limit(pageSize).Order("snap_time DESC").Find(&snapshots).Error; err != nil {
		return nil, fmt.Errorf("failed to list snapshots: %w", err)
	}

	// Convert to responses
	items := make([]dto.SnapshotResponse, len(snapshots))
	for i, snap := range snapshots {
		items[i] = *s.snapshotToResponse(&snap)
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	return &dto.ListResponse{
		Items:      items,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

func (s *ZFSService) RollbackSnapshot(req dto.RollbackSnapshotRequest) error {
	if err := s.client.RollbackSnapshot(req.Snapshot); err != nil {
		return fmt.Errorf("failed to rollback snapshot: %w", err)
	}
	return nil
}

func (s *ZFSService) CloneSnapshot(req dto.CloneSnapshotRequest) error {
	if err := s.client.CloneSnapshot(req.Snapshot, req.Target); err != nil {
		return fmt.Errorf("failed to clone snapshot: %w", err)
	}
	return nil
}

func (s *ZFSService) DestroySnapshot(name string) error {
	// Delete from database
	if err := s.db.Where("name = ?", name).Delete(&models.ZFSSnapshot{}).Error; err != nil {
		return fmt.Errorf("failed to delete snapshot from database: %w", err)
	}

	// Destroy snapshot
	if err := s.client.DestroySnapshot(name); err != nil {
		return fmt.Errorf("failed to destroy snapshot: %w", err)
	}

	return nil
}

// Volume Management

func (s *ZFSService) CreateVolume(req dto.CreateVolumeRequest) (*dto.VolumeResponse, error) {
	opts := zfs.VolumeOptions{
		BlockSize:   req.BlockSize,
		Sparse:      req.Sparse,
		Compression: req.Compression,
		Dedup:       req.Dedup,
		Encryption:  req.Encryption,
		KeyLocation: req.KeyLocation,
		Properties:  req.Properties,
	}

	if err := s.client.CreateVolume(req.Name, req.Size, opts); err != nil {
		return nil, fmt.Errorf("failed to create volume: %w", err)
	}

	poolName := s.extractPoolName(req.Name)
	devicePath := fmt.Sprintf("/dev/zvol/%s", req.Name)

	// Save to database
	volume := models.ZFSVolume{
		Base:        models.Base{ID: uuid.New().String()},
		Name:        req.Name,
		Pool:        poolName,
		Size:        req.Size,
		BlockSize:   req.BlockSize,
		Compression: req.Compression,
		Dedup:       req.Dedup,
		Encryption:  req.Encryption,
		Sparse:      req.Sparse,
		DevicePath:  devicePath,
		Status:      "ONLINE",
	}

	if err := s.db.Create(&volume).Error; err != nil {
		return nil, fmt.Errorf("failed to save volume to database: %w", err)
	}

	return s.volumeToResponse(&volume), nil
}

func (s *ZFSService) ListVolumes(req dto.ListRequest) (*dto.ListResponse, error) {
	var volumes []models.ZFSVolume
	query := s.db.Model(&models.ZFSVolume{})

	if req.Pool != "" {
		query = query.Where("pool = ?", req.Pool)
	}
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}

	// Count total
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count volumes: %w", err)
	}

	// Apply pagination
	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	if err := query.Offset(offset).Limit(pageSize).Find(&volumes).Error; err != nil {
		return nil, fmt.Errorf("failed to list volumes: %w", err)
	}

	// Convert to responses
	items := make([]dto.VolumeResponse, len(volumes))
	for i, vol := range volumes {
		items[i] = *s.volumeToResponse(&vol)
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	return &dto.ListResponse{
		Items:      items,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

func (s *ZFSService) ResizeVolume(name string, req dto.ResizeVolumeRequest) (*dto.VolumeResponse, error) {
	if err := s.client.ResizeVolume(name, req.NewSize); err != nil {
		return nil, fmt.Errorf("failed to resize volume: %w", err)
	}

	// Update database
	if err := s.db.Model(&models.ZFSVolume{}).Where("name = ?", name).Update("size", req.NewSize).Error; err != nil {
		return nil, fmt.Errorf("failed to update volume in database: %w", err)
	}

	// Get updated volume
	var volume models.ZFSVolume
	if err := s.db.Where("name = ?", name).First(&volume).Error; err != nil {
		return nil, fmt.Errorf("failed to get updated volume: %w", err)
	}

	return s.volumeToResponse(&volume), nil
}

func (s *ZFSService) DestroyVolume(name string) error {
	// Delete from database
	if err := s.db.Where("name = ?", name).Delete(&models.ZFSVolume{}).Error; err != nil {
		return fmt.Errorf("failed to delete volume from database: %w", err)
	}

	// Destroy volume
	if err := s.client.DestroyVolume(name); err != nil {
		return fmt.Errorf("failed to destroy volume: %w", err)
	}

	return nil
}

// Helper functions

func (s *ZFSService) poolToResponse(pool *models.ZFSPool) *dto.PoolResponse {
	return &dto.PoolResponse{
		ID:          pool.ID,
		Name:        pool.Name,
		Size:        pool.Size,
		Allocated:   pool.Allocated,
		Free:        pool.Free,
		Capacity:    pool.Capacity,
		Health:      pool.Health,
		Dedup:       pool.Dedup,
		Compression: pool.Compression,
		Status:      pool.Status,
		VDevs:       pool.VDevs,
		Properties:  pool.Properties,
		CreatedAt:   pool.CreatedAt,
	}
}

func (s *ZFSService) datasetToResponse(dataset *models.ZFSDataset) *dto.DatasetResponse {
	return &dto.DatasetResponse{
		ID:             dataset.ID,
		Name:           dataset.Name,
		Pool:           dataset.Pool,
		Type:           dataset.Type,
		Mountpoint:     dataset.Mountpoint,
		Quota:          dataset.Quota,
		Reservation:    dataset.Reservation,
		Used:           dataset.Used,
		Available:      dataset.Available,
		Compression:    dataset.Compression,
		CompressRatio:  dataset.CompressRatio,
		Dedup:          dataset.Dedup,
		Encryption:     dataset.Encryption,
		KeyStatus:      dataset.KeyStatus,
		ReadOnly:       dataset.ReadOnly,
		Atime:          dataset.Atime,
		RecordSize:     dataset.RecordSize,
		Status:         dataset.Status,
		AllProperties:  dataset.AllProperties,
		CreatedAt:      dataset.CreatedAt,
	}
}

func (s *ZFSService) snapshotToResponse(snapshot *models.ZFSSnapshot) *dto.SnapshotResponse {
	return &dto.SnapshotResponse{
		ID:         snapshot.ID,
		Name:       snapshot.Name,
		Pool:       snapshot.Pool,
		Dataset:    snapshot.Dataset,
		SnapName:   snapshot.SnapName,
		Used:       snapshot.Used,
		Referenced: snapshot.Referenced,
		SnapTime:   snapshot.SnapTime,
		Comment:    snapshot.Comment,
		Status:     snapshot.Status,
		CreatedAt:  snapshot.CreatedAt,
	}
}

func (s *ZFSService) volumeToResponse(volume *models.ZFSVolume) *dto.VolumeResponse {
	return &dto.VolumeResponse{
		ID:          volume.ID,
		Name:        volume.Name,
		Pool:        volume.Pool,
		Size:        volume.Size,
		BlockSize:   volume.BlockSize,
		Used:        volume.Used,
		Available:   volume.Available,
		Compression: volume.Compression,
		Dedup:       volume.Dedup,
		Encryption:  volume.Encryption,
		KeyStatus:   volume.KeyStatus,
		Sparse:      volume.Sparse,
		DevicePath:  volume.DevicePath,
		Status:      volume.Status,
		CreatedAt:   volume.CreatedAt,
	}
}

func (s *ZFSService) vdevsToString(vdevs []dto.VDevSpec) string {
	result := ""
	for i, vdev := range vdevs {
		if i > 0 {
			result += " "
		}
		result += vdev.Type
		for _, dev := range vdev.Devices {
			result += " " + dev
		}
	}
	return result
}

func (s *ZFSService) extractPoolName(name string) string {
	// Extract pool name from dataset/volume name (e.g., "mypool/dataset" -> "mypool")
	for i := 0; i < len(name); i++ {
		if name[i] == '/' {
			return name[:i]
		}
	}
	return name
}

func (s *ZFSService) convertVDevInfos(infos []zfs.VDevInfo) []dto.VDevInfo {
	result := make([]dto.VDevInfo, len(infos))
	for i, info := range infos {
		result[i] = dto.VDevInfo{
			Name:     info.Name,
			Type:     info.Type,
			State:    info.State,
			Read:     info.Read,
			Write:    info.Write,
			Cksum:    info.Cksum,
			Children: s.convertVDevInfos(info.Children),
		}
	}
	return result
}

func (s *ZFSService) convertErrors(errors []string) []dto.ErrorInfo {
	result := make([]dto.ErrorInfo, len(errors))
	for i, err := range errors {
		result[i] = dto.ErrorInfo{
			Type:        "ERROR",
			Description: err,
		}
	}
	return result
}
