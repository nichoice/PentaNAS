package services

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"pnas/internal/app/dto"
	"pnas/internal/database"
	"pnas/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type iSCSILUNService struct {
	db     *gorm.DB
	cmdSvc *iSCSICommandService
}

func NewiSCSILUNService() *iSCSILUNService {
	return &iSCSILUNService{
		db:     database.DB,
		cmdSvc: NewiSCSICommandService(),
	}
}

// CreateLUN creates a new LUN for an iSCSI target
func (s *iSCSILUNService) CreateLUN(req *dto.CreateiSCSILUNRequest) (*dto.iSCSILUNResponse, error) {
	// Verify target exists
	var target models.iSCSITarget
	if err := s.db.First(&target, "id = ?", req.TargetID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("target not found")
		}
		return nil, fmt.Errorf("failed to get target: %w", err)
	}

	// Check if LUN number already exists for this target
	var existingLUN models.iSCSILUN
	if err := s.db.Where("target_id = ? AND lun = ?", req.TargetID, req.LUN).First(&existingLUN).Error; err == nil {
		return nil, fmt.Errorf("LUN %d already exists for target %s", req.LUN, target.Name)
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to check LUN existence: %w", err)
	}

	// Create LUN record
	lun := models.iSCSILUN{
		Base: models.Base{
			ID: uuid.New().String(),
		},
		TargetID:     req.TargetID,
		LUN:          req.LUN,
		Name:         req.Name,
		DeviceType:   models.iSCSIDeviceType(req.DeviceType),
		DevicePath:   req.DevicePath,
		Size:         req.Size,
		BlockSize:    req.BlockSize,
		ReadOnly:     req.ReadOnly,
		WriteThrough: req.WriteThrough,
		WriteBack:    req.WriteBack,
		TCMUHandler:  req.TCMUHandler,
		TCMUOptions:  req.TCMUOptions,
		Comment:      req.Comment,
		IsEnabled:    true,
	}

	if lun.BlockSize == 0 {
		lun.BlockSize = 512 // Default block size
	}

	// Create device file if type is file and file doesn't exist
	if req.DeviceType == "file" {
		if err := s.createFileDevice(req.DevicePath, req.Size); err != nil {
			return nil, fmt.Errorf("failed to create device file: %w", err)
		}
	}

	// Validate device path exists for block devices
	if req.DeviceType == "block" {
		if _, err := os.Stat(req.DevicePath); os.IsNotExist(err) {
			return nil, fmt.Errorf("block device %s does not exist", req.DevicePath)
		}
	}

	// Save to database
	if err := s.db.Create(&lun).Error; err != nil {
		return nil, fmt.Errorf("failed to create LUN: %w", err)
	}

	// Apply to system if target is active
	if target.Status == models.iSCSIStatusActive {
		if err := s.cmdSvc.CreateLUN(target.Name, req.LUN, req.DevicePath, req.DeviceType); err != nil {
			// Log error but don't rollback database operation
			fmt.Printf("Warning: failed to create LUN in system: %v\n", err)
		}
	}

	return s.convertLUNToResponse(&lun), nil
}

// GetLUN gets a LUN by ID
func (s *iSCSILUNService) GetLUN(id string) (*dto.iSCSILUNResponse, error) {
	var lun models.iSCSILUN
	if err := s.db.Preload("Target").First(&lun, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("LUN not found")
		}
		return nil, fmt.Errorf("failed to get LUN: %w", err)
	}

	return s.convertLUNToResponse(&lun), nil
}

// ListLUNs lists LUNs with pagination and filters
func (s *iSCSILUNService) ListLUNs(query *dto.iSCSILUNListQuery) ([]*dto.iSCSILUNResponse, int64, error) {
	var luns []models.iSCSILUN
	var total int64

	db := s.db.Model(&models.iSCSILUN{}).Preload("Target")

	// Apply filters
	if query.TargetID != "" {
		db = db.Where("target_id = ?", query.TargetID)
	}
	if query.DeviceType != "" {
		db = db.Where("device_type = ?", query.DeviceType)
	}
	if query.IsEnabled != nil {
		db = db.Where("is_enabled = ?", *query.IsEnabled)
	}

	// Get total count
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count LUNs: %w", err)
	}

	// Apply pagination
	page := query.Page
	if page <= 0 {
		page = 1
	}
	pageSize := query.PageSize
	if pageSize <= 0 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize
	if err := db.Offset(offset).Limit(pageSize).Find(&luns).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list LUNs: %w", err)
	}

	responses := make([]*dto.iSCSILUNResponse, len(luns))
	for i, lun := range luns {
		responses[i] = s.convertLUNToResponse(&lun)
	}

	return responses, total, nil
}

// UpdateLUN updates an existing LUN
func (s *iSCSILUNService) UpdateLUN(id string, req *dto.UpdateiSCSILUNRequest) (*dto.iSCSILUNResponse, error) {
	var lun models.iSCSILUN
	if err := s.db.Preload("Target").First(&lun, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("LUN not found")
		}
		return nil, fmt.Errorf("failed to get LUN: %w", err)
	}

	// Update fields
	if req.Name != "" {
		lun.Name = req.Name
	}
	if req.Size != nil && *req.Size > 0 {
		// Handle size change
		oldSize := lun.Size
		lun.Size = *req.Size

		// For file-backed LUNs, resize the file
		if lun.DeviceType == models.iSCSIDeviceFile {
			if err := s.cmdSvc.ResizeFileBackedLUN(lun.DevicePath, *req.Size); err != nil {
				return nil, fmt.Errorf("failed to resize device file: %w", err)
			}
		}

		// Log size change
		fmt.Printf("LUN %s size changed from %d to %d bytes\n", lun.Name, oldSize, *req.Size)
	}
	if req.ReadOnly != nil {
		lun.ReadOnly = *req.ReadOnly
	}
	if req.WriteThrough != nil {
		lun.WriteThrough = *req.WriteThrough
	}
	if req.WriteBack != nil {
		lun.WriteBack = *req.WriteBack
	}
	if req.IsEnabled != nil {
		lun.IsEnabled = *req.IsEnabled
	}
	if req.TCMUOptions != "" {
		lun.TCMUOptions = req.TCMUOptions
	}
	if req.Comment != "" {
		lun.Comment = req.Comment
	}

	if err := s.db.Save(&lun).Error; err != nil {
		return nil, fmt.Errorf("failed to update LUN: %w", err)
	}

	return s.convertLUNToResponse(&lun), nil
}

// DeleteLUN deletes a LUN
func (s *iSCSILUNService) DeleteLUN(id string) error {
	var lun models.iSCSILUN
	if err := s.db.Preload("Target").First(&lun, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("LUN not found")
		}
		return fmt.Errorf("failed to get LUN: %w", err)
	}

	// Check for LUN mappings
	var mappingCount int64
	s.db.Model(&models.iSCSILUNMapping{}).Where("lun_id = ?", id).Count(&mappingCount)
	if mappingCount > 0 {
		return fmt.Errorf("cannot delete LUN with existing ACL mappings")
	}

	// Remove from system if target is active
	if lun.Target.Status == models.iSCSIStatusActive {
		backstoreName := s.generateBackstoreName(lun.Target.Name, lun.LUN, string(lun.DeviceType))
		if err := s.cmdSvc.DeleteLUN(lun.Target.Name, lun.LUN, backstoreName); err != nil {
			fmt.Printf("Warning: failed to remove LUN from system: %v\n", err)
		}
	}

	// Delete device file if it's a file-backed LUN and was created by us
	if lun.DeviceType == models.iSCSIDeviceFile {
		if err := s.deleteFileDevice(lun.DevicePath); err != nil {
			fmt.Printf("Warning: failed to delete device file %s: %v\n", lun.DevicePath, err)
		}
	}

	if err := s.db.Delete(&lun).Error; err != nil {
		return fmt.Errorf("failed to delete LUN: %w", err)
	}

	return nil
}

// BatchOperationLUNs performs batch operations on LUNs
func (s *iSCSILUNService) BatchOperationLUNs(req *dto.BatchiSCSILUNRequest) (*dto.BatchOperationResponse, error) {
	response := &dto.BatchOperationResponse{
		Success: []string{},
		Failed:  []string{},
		Errors:  []string{},
	}

	for _, lunID := range req.LUNIDs {
		var lun models.iSCSILUN
		if err := s.db.Preload("Target").First(&lun, "id = ?", lunID).Error; err != nil {
			response.Failed = append(response.Failed, lunID)
			response.Errors = append(response.Errors, fmt.Sprintf("LUN %s: not found", lunID))
			continue
		}

		switch req.Action {
		case "enable":
			if !lun.IsEnabled {
				lun.IsEnabled = true
				if err := s.db.Save(&lun).Error; err != nil {
					response.Failed = append(response.Failed, lunID)
					response.Errors = append(response.Errors, fmt.Sprintf("LUN %s: failed to enable", lunID))
				} else {
					response.Success = append(response.Success, lunID)
				}
			} else {
				response.Success = append(response.Success, lunID)
			}

		case "disable":
			if lun.IsEnabled {
				lun.IsEnabled = false
				if err := s.db.Save(&lun).Error; err != nil {
					response.Failed = append(response.Failed, lunID)
					response.Errors = append(response.Errors, fmt.Sprintf("LUN %s: failed to disable", lunID))
				} else {
					response.Success = append(response.Success, lunID)
				}
			} else {
				response.Success = append(response.Success, lunID)
			}

		case "delete":
			// Check for mappings
			var mappingCount int64
			s.db.Model(&models.iSCSILUNMapping{}).Where("lun_id = ?", lunID).Count(&mappingCount)
			if mappingCount > 0 {
				response.Failed = append(response.Failed, lunID)
				response.Errors = append(response.Errors, fmt.Sprintf("LUN %s: has ACL mappings", lunID))
				continue
			}

			// Remove from system
			if lun.Target.Status == models.iSCSIStatusActive {
				backstoreName := s.generateBackstoreName(lun.Target.Name, lun.LUN, string(lun.DeviceType))
				s.cmdSvc.DeleteLUN(lun.Target.Name, lun.LUN, backstoreName)
			}

			// Delete device file if file-backed
			if lun.DeviceType == models.iSCSIDeviceFile {
				s.deleteFileDevice(lun.DevicePath)
			}

			if err := s.db.Delete(&lun).Error; err != nil {
				response.Failed = append(response.Failed, lunID)
				response.Errors = append(response.Errors, fmt.Sprintf("LUN %s: failed to delete", lunID))
			} else {
				response.Success = append(response.Success, lunID)
			}

		default:
			response.Failed = append(response.Failed, lunID)
			response.Errors = append(response.Errors, fmt.Sprintf("LUN %s: invalid action %s", lunID, req.Action))
		}
	}

	return response, nil
}

// GetLUNsByTarget gets all LUNs for a specific target
func (s *iSCSILUNService) GetLUNsByTarget(targetID string) ([]*dto.iSCSILUNResponse, error) {
	var luns []models.iSCSILUN
	if err := s.db.Where("target_id = ?", targetID).Find(&luns).Error; err != nil {
		return nil, fmt.Errorf("failed to get LUNs for target: %w", err)
	}

	responses := make([]*dto.iSCSILUNResponse, len(luns))
	for i, lun := range luns {
		responses[i] = s.convertLUNToResponse(&lun)
	}

	return responses, nil
}

// createFileDevice creates a file-backed device
func (s *iSCSILUNService) createFileDevice(filePath string, sizeBytes int64) error {
	// Check if file already exists
	if _, err := os.Stat(filePath); err == nil {
		return fmt.Errorf("file %s already exists", filePath)
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Create the file-backed LUN
	if err := s.cmdSvc.CreateFileBackedLUN(filePath, sizeBytes); err != nil {
		return fmt.Errorf("failed to create file-backed LUN: %w", err)
	}

	return nil
}

// deleteFileDevice deletes a file-backed device
func (s *iSCSILUNService) deleteFileDevice(filePath string) error {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil // File doesn't exist, nothing to delete
	}

	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("failed to remove file %s: %w", filePath, err)
	}

	return nil
}

// generateBackstoreName generates a backstore name based on target and LUN
func (s *iSCSILUNService) generateBackstoreName(targetName string, lunID int, deviceType string) string {
	sanitizedName := strings.ReplaceAll(targetName, ":", "_")
	sanitizedName = strings.ReplaceAll(sanitizedName, ".", "_")
	return fmt.Sprintf("%s_%s_lun%d", deviceType, sanitizedName, lunID)
}

// convertLUNToResponse converts model to response DTO
func (s *iSCSILUNService) convertLUNToResponse(lun *models.iSCSILUN) *dto.iSCSILUNResponse {
	return &dto.iSCSILUNResponse{
		ID:           lun.ID,
		TargetID:     lun.TargetID,
		LUN:          lun.LUN,
		Name:         lun.Name,
		DeviceType:   string(lun.DeviceType),
		DevicePath:   lun.DevicePath,
		Size:         lun.Size,
		BlockSize:    lun.BlockSize,
		ReadOnly:     lun.ReadOnly,
		WriteThrough: lun.WriteThrough,
		WriteBack:    lun.WriteBack,
		IsEnabled:    lun.IsEnabled,
		TCMUHandler:  lun.TCMUHandler,
		TCMUOptions:  lun.TCMUOptions,
		Comment:      lun.Comment,
		CreatedAt:    lun.CreatedAt,
		UpdatedAt:    lun.UpdatedAt,
	}
}