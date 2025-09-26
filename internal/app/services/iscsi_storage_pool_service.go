package services

import (
	"errors"
	"fmt"
	"os"

	"pnas/internal/app/dto"
	"pnas/internal/database"
	"pnas/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type iSCSIStoragePoolService struct {
	db *gorm.DB
}

func NewiSCSIStoragePoolService() *iSCSIStoragePoolService {
	return &iSCSIStoragePoolService{db: database.DB}
}

// CreateStoragePool creates a new storage pool
func (s *iSCSIStoragePoolService) CreateStoragePool(req *dto.CreateiSCSIStoragePoolRequest) (*dto.iSCSIStoragePoolResponse, error) {
	// Check if pool name already exists
	var existingPool models.iSCSIStoragePool
	if err := s.db.Where("name = ?", req.Name).First(&existingPool).Error; err == nil {
		return nil, fmt.Errorf("storage pool with name %s already exists", req.Name)
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to check pool existence: %w", err)
	}

	// Validate path exists for certain types
	if req.Type == "file" || req.Type == "lvm" {
		if _, err := os.Stat(req.Path); os.IsNotExist(err) {
			return nil, fmt.Errorf("path %s does not exist", req.Path)
		}
	}

	pool := models.iSCSIStoragePool{
		Base: models.Base{
			ID: uuid.New().String(),
		},
		Name:           req.Name,
		Type:           req.Type,
		Path:           req.Path,
		Size:           req.Size,
		Used:           0,
		Available:      req.Size,
		IsEnabled:      true,
		Comment:        req.Comment,
		BlockSize:      req.BlockSize,
		AllocationUnit: req.AllocationUnit,
	}

	if pool.BlockSize == 0 {
		pool.BlockSize = 4096 // Default block size
	}
	if pool.AllocationUnit == 0 {
		pool.AllocationUnit = 1048576 // 1MB default
	}

	if err := s.db.Create(&pool).Error; err != nil {
		return nil, fmt.Errorf("failed to create storage pool: %w", err)
	}

	return s.convertPoolToResponse(&pool), nil
}

// GetStoragePool gets a storage pool by ID
func (s *iSCSIStoragePoolService) GetStoragePool(id string) (*dto.iSCSIStoragePoolResponse, error) {
	var pool models.iSCSIStoragePool
	if err := s.db.First(&pool, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("storage pool not found")
		}
		return nil, fmt.Errorf("failed to get storage pool: %w", err)
	}

	return s.convertPoolToResponse(&pool), nil
}

// ListStoragePools lists all storage pools
func (s *iSCSIStoragePoolService) ListStoragePools() ([]*dto.iSCSIStoragePoolResponse, error) {
	var pools []models.iSCSIStoragePool
	if err := s.db.Find(&pools).Error; err != nil {
		return nil, fmt.Errorf("failed to list storage pools: %w", err)
	}

	responses := make([]*dto.iSCSIStoragePoolResponse, len(pools))
	for i, pool := range pools {
		responses[i] = s.convertPoolToResponse(&pool)
	}

	return responses, nil
}

// UpdateStoragePool updates an existing storage pool
func (s *iSCSIStoragePoolService) UpdateStoragePool(id string, req *dto.UpdateiSCSIStoragePoolRequest) (*dto.iSCSIStoragePoolResponse, error) {
	var pool models.iSCSIStoragePool
	if err := s.db.First(&pool, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("storage pool not found")
		}
		return nil, fmt.Errorf("failed to get storage pool: %w", err)
	}

	// Update fields
	if req.Size != nil && *req.Size > 0 {
		if *req.Size < pool.Used {
			return nil, fmt.Errorf("cannot reduce pool size below used space (%d bytes)", pool.Used)
		}
		pool.Size = *req.Size
		pool.Available = *req.Size - pool.Used
	}
	if req.BlockSize != nil && *req.BlockSize > 0 {
		pool.BlockSize = *req.BlockSize
	}
	if req.AllocationUnit != nil && *req.AllocationUnit > 0 {
		pool.AllocationUnit = *req.AllocationUnit
	}
	if req.IsEnabled != nil {
		pool.IsEnabled = *req.IsEnabled
	}
	if req.Comment != "" {
		pool.Comment = req.Comment
	}

	if err := s.db.Save(&pool).Error; err != nil {
		return nil, fmt.Errorf("failed to update storage pool: %w", err)
	}

	return s.convertPoolToResponse(&pool), nil
}

// DeleteStoragePool deletes a storage pool
func (s *iSCSIStoragePoolService) DeleteStoragePool(id string) error {
	var pool models.iSCSIStoragePool
	if err := s.db.First(&pool, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("storage pool not found")
		}
		return fmt.Errorf("failed to get storage pool: %w", err)
	}

	// Check if pool has LUNs
	var lunCount int64
	s.db.Model(&models.iSCSILUN{}).Where("device_path LIKE ?", pool.Path+"%").Count(&lunCount)
	if lunCount > 0 {
		return fmt.Errorf("cannot delete storage pool with existing LUNs")
	}

	if err := s.db.Delete(&pool).Error; err != nil {
		return fmt.Errorf("failed to delete storage pool: %w", err)
	}

	return nil
}

// UpdatePoolUsage updates the usage statistics of a storage pool
func (s *iSCSIStoragePoolService) UpdatePoolUsage(id string, usedBytes int64) error {
	var pool models.iSCSIStoragePool
	if err := s.db.First(&pool, "id = ?", id).Error; err != nil {
		return fmt.Errorf("failed to get storage pool: %w", err)
	}

	if usedBytes > pool.Size {
		return fmt.Errorf("used bytes cannot exceed pool size")
	}

	pool.Used = usedBytes
	pool.Available = pool.Size - usedBytes

	if err := s.db.Save(&pool).Error; err != nil {
		return fmt.Errorf("failed to update pool usage: %w", err)
	}

	return nil
}

// convertPoolToResponse converts model to response DTO
func (s *iSCSIStoragePoolService) convertPoolToResponse(pool *models.iSCSIStoragePool) *dto.iSCSIStoragePoolResponse {
	var usagePercent float64
	if pool.Size > 0 {
		usagePercent = float64(pool.Used) / float64(pool.Size) * 100
	}

	response := &dto.iSCSIStoragePoolResponse{
		ID:             pool.ID,
		Name:           pool.Name,
		Type:           pool.Type,
		Path:           pool.Path,
		Size:           pool.Size,
		Used:           pool.Used,
		Available:      pool.Available,
		UsagePercent:   usagePercent,
		BlockSize:      pool.BlockSize,
		AllocationUnit: pool.AllocationUnit,
		IsEnabled:      pool.IsEnabled,
		Comment:        pool.Comment,
		CreatedAt:      pool.CreatedAt,
		UpdatedAt:      pool.UpdatedAt,
	}

	// Count associated LUNs
	var lunCount int64
	s.db.Model(&models.iSCSILUN{}).Where("device_path LIKE ?", pool.Path+"%").Count(&lunCount)
	response.LUNCount = int(lunCount)

	return response
}