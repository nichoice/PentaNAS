package services

import (
	"errors"
	"fmt"
	"math"
	"strings"

	"pnas/internal/app/dto"
	"pnas/internal/database"
	"pnas/internal/models"
	"pnas/internal/shared/utils"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type iSCSITargetService struct {
	db     *gorm.DB
	cmdSvc *iSCSICommandService
}

func NewiSCSITargetService() *iSCSITargetService {
	return &iSCSITargetService{
		db:     database.DB,
		cmdSvc: NewiSCSICommandService(),
	}
}

// CreateTarget creates a new iSCSI target
func (s *iSCSITargetService) CreateTarget(req *dto.CreateiSCSITargetRequest) (*dto.iSCSITargetResponse, error) {
	// Validate IQN format
	if !s.isValidIQN(req.Name) {
		return nil, fmt.Errorf("invalid IQN format: %s", req.Name)
	}

	// Check if target name already exists
	var existing models.iSCSITarget
	if err := s.db.Where("name = ?", req.Name).First(&existing).Error; err == nil {
		return nil, fmt.Errorf("target with name %s already exists", req.Name)
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to check target existence: %w", err)
	}

	target := models.iSCSITarget{
		Base: models.Base{
			ID: uuid.New().String(),
		},
		Name:      req.Name,
		Alias:     req.Alias,
		Status:    models.iSCSIStatusInactive,
		Comment:   req.Comment,
		IsEnabled: req.IsEnabled,
	}

	if err := s.db.Create(&target).Error; err != nil {
		return nil, fmt.Errorf("failed to create target: %w", err)
	}

	// Apply target configuration
	if target.IsEnabled {
		if err := s.cmdSvc.CreateTarget(target.Name); err != nil {
			// Log error but don't fail the database operation
			fmt.Printf("Warning: failed to create target in system: %v\n", err)
			target.Status = models.iSCSIStatusError
			s.db.Save(&target)
		} else {
			target.Status = models.iSCSIStatusActive
			s.db.Save(&target)
		}
	}

	return s.convertTargetToResponse(&target), nil
}

// GetTarget gets a target by ID
func (s *iSCSITargetService) GetTarget(id string) (*dto.iSCSITargetResponse, error) {
	var target models.iSCSITarget
	if err := s.db.Preload("LUNs").Preload("ACLs").First(&target, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("target not found")
		}
		return nil, fmt.Errorf("failed to get target: %w", err)
	}

	return s.convertTargetToResponse(&target), nil
}

// ListTargets lists all targets with pagination
func (s *iSCSITargetService) ListTargets(query *dto.iSCSITargetListQuery) ([]*dto.iSCSITargetResponse, int64, error) {
	var targets []models.iSCSITarget
	var total int64

	db := s.db.Model(&models.iSCSITarget{}).Preload("LUNs").Preload("ACLs")

	// Apply filters
	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}
	if query.IsEnabled != nil {
		db = db.Where("is_enabled = ?", *query.IsEnabled)
	}
	if query.Search != "" {
		searchPattern := "%" + query.Search + "%"
		db = db.Where("name LIKE ? OR alias LIKE ? OR comment LIKE ?", searchPattern, searchPattern, searchPattern)
	}

	// Get total count
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count targets: %w", err)
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
	if err := db.Offset(offset).Limit(pageSize).Find(&targets).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list targets: %w", err)
	}

	responses := make([]*dto.iSCSITargetResponse, len(targets))
	for i, target := range targets {
		responses[i] = s.convertTargetToResponse(&target)
	}

	return responses, total, nil
}

// UpdateTarget updates an existing target
func (s *iSCSITargetService) UpdateTarget(id string, req *dto.UpdateiSCSITargetRequest) (*dto.iSCSITargetResponse, error) {
	var target models.iSCSITarget
	if err := s.db.First(&target, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("target not found")
		}
		return nil, fmt.Errorf("failed to get target: %w", err)
	}

	oldEnabled := target.IsEnabled

	// Update fields
	if req.Alias != "" {
		target.Alias = req.Alias
	}
	if req.Comment != "" {
		target.Comment = req.Comment
	}
	if req.IsEnabled != nil {
		target.IsEnabled = *req.IsEnabled
	}

	if err := s.db.Save(&target).Error; err != nil {
		return nil, fmt.Errorf("failed to update target: %w", err)
	}

	// Handle enable/disable state changes
	if oldEnabled != target.IsEnabled {
		if target.IsEnabled {
			// Enable target
			if err := s.cmdSvc.CreateTarget(target.Name); err != nil {
				fmt.Printf("Warning: failed to enable target in system: %v\n", err)
				target.Status = models.iSCSIStatusError
			} else {
				target.Status = models.iSCSIStatusActive
			}
		} else {
			// Disable target
			if err := s.cmdSvc.DeleteTarget(target.Name); err != nil {
				fmt.Printf("Warning: failed to disable target in system: %v\n", err)
				target.Status = models.iSCSIStatusError
			} else {
				target.Status = models.iSCSIStatusInactive
			}
		}
		s.db.Save(&target)
	}

	return s.convertTargetToResponse(&target), nil
}

// DeleteTarget deletes a target
func (s *iSCSITargetService) DeleteTarget(id string) error {
	var target models.iSCSITarget
	if err := s.db.Preload("LUNs").Preload("ACLs").First(&target, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("target not found")
		}
		return fmt.Errorf("failed to get target: %w", err)
	}

	// Check if target has LUNs or ACLs
	if len(target.LUNs) > 0 {
		return fmt.Errorf("cannot delete target with existing LUNs")
	}
	if len(target.ACLs) > 0 {
		return fmt.Errorf("cannot delete target with existing ACLs")
	}

	// Remove from system if active
	if target.Status == models.iSCSIStatusActive {
		if err := s.cmdSvc.DeleteTarget(target.Name); err != nil {
			fmt.Printf("Warning: failed to remove target from system: %v\n", err)
		}
	}

	if err := s.db.Delete(&target).Error; err != nil {
		return fmt.Errorf("failed to delete target: %w", err)
	}

	return nil
}

// BatchOperationTargets performs batch operations on targets
func (s *iSCSITargetService) BatchOperationTargets(req *dto.BatchiSCSITargetRequest) (*dto.BatchOperationResponse, error) {
	response := &dto.BatchOperationResponse{
		Success: []string{},
		Failed:  []string{},
		Errors:  []string{},
	}

	for _, targetID := range req.TargetIDs {
		var target models.iSCSITarget
		if err := s.db.First(&target, "id = ?", targetID).Error; err != nil {
			response.Failed = append(response.Failed, targetID)
			response.Errors = append(response.Errors, fmt.Sprintf("Target %s: not found", targetID))
			continue
		}

		switch req.Action {
		case "enable":
			if !target.IsEnabled {
				target.IsEnabled = true
				if err := s.db.Save(&target).Error; err != nil {
					response.Failed = append(response.Failed, targetID)
					response.Errors = append(response.Errors, fmt.Sprintf("Target %s: failed to enable", targetID))
				} else {
					// Apply to system
					if err := s.cmdSvc.CreateTarget(target.Name); err != nil {
						fmt.Printf("Warning: failed to enable target %s in system: %v\n", target.Name, err)
						target.Status = models.iSCSIStatusError
					} else {
						target.Status = models.iSCSIStatusActive
					}
					s.db.Save(&target)
					response.Success = append(response.Success, targetID)
				}
			} else {
				response.Success = append(response.Success, targetID)
			}

		case "disable":
			if target.IsEnabled {
				target.IsEnabled = false
				if err := s.db.Save(&target).Error; err != nil {
					response.Failed = append(response.Failed, targetID)
					response.Errors = append(response.Errors, fmt.Sprintf("Target %s: failed to disable", targetID))
				} else {
					// Remove from system
					if err := s.cmdSvc.DeleteTarget(target.Name); err != nil {
						fmt.Printf("Warning: failed to disable target %s in system: %v\n", target.Name, err)
						target.Status = models.iSCSIStatusError
					} else {
						target.Status = models.iSCSIStatusInactive
					}
					s.db.Save(&target)
					response.Success = append(response.Success, targetID)
				}
			} else {
				response.Success = append(response.Success, targetID)
			}

		case "delete":
			// Check dependencies
			var lunCount, aclCount int64
			s.db.Model(&models.iSCSILUN{}).Where("target_id = ?", targetID).Count(&lunCount)
			s.db.Model(&models.iSCSIACL{}).Where("target_id = ?", targetID).Count(&aclCount)

			if lunCount > 0 || aclCount > 0 {
				response.Failed = append(response.Failed, targetID)
				response.Errors = append(response.Errors, fmt.Sprintf("Target %s: has dependencies (LUNs or ACLs)", targetID))
				continue
			}

			// Remove from system
			if target.Status == models.iSCSIStatusActive {
				s.cmdSvc.DeleteTarget(target.Name)
			}

			if err := s.db.Delete(&target).Error; err != nil {
				response.Failed = append(response.Failed, targetID)
				response.Errors = append(response.Errors, fmt.Sprintf("Target %s: failed to delete", targetID))
			} else {
				response.Success = append(response.Success, targetID)
			}

		default:
			response.Failed = append(response.Failed, targetID)
			response.Errors = append(response.Errors, fmt.Sprintf("Target %s: invalid action %s", targetID, req.Action))
		}
	}

	return response, nil
}

// GetTargetStats gets target statistics
func (s *iSCSITargetService) GetTargetStats() (map[string]interface{}, error) {
	var stats struct {
		Total    int64 `json:"total"`
		Active   int64 `json:"active"`
		Inactive int64 `json:"inactive"`
		Error    int64 `json:"error"`
		Enabled  int64 `json:"enabled"`
		Disabled int64 `json:"disabled"`
	}

	// Total targets
	s.db.Model(&models.iSCSITarget{}).Count(&stats.Total)

	// By status
	s.db.Model(&models.iSCSITarget{}).Where("status = ?", models.iSCSIStatusActive).Count(&stats.Active)
	s.db.Model(&models.iSCSITarget{}).Where("status = ?", models.iSCSIStatusInactive).Count(&stats.Inactive)
	s.db.Model(&models.iSCSITarget{}).Where("status = ?", models.iSCSIStatusError).Count(&stats.Error)

	// By enabled state
	s.db.Model(&models.iSCSITarget{}).Where("is_enabled = ?", true).Count(&stats.Enabled)
	s.db.Model(&models.iSCSITarget{}).Where("is_enabled = ?", false).Count(&stats.Disabled)

	result := map[string]interface{}{
		"total":    stats.Total,
		"active":   stats.Active,
		"inactive": stats.Inactive,
		"error":    stats.Error,
		"enabled":  stats.Enabled,
		"disabled": stats.Disabled,
	}

	return result, nil
}

// isValidIQN validates IQN format
func (s *iSCSITargetService) isValidIQN(iqn string) bool {
	// Basic IQN validation: should start with "iqn." followed by date and domain
	if !strings.HasPrefix(iqn, "iqn.") {
		return false
	}

	parts := strings.Split(iqn, ":")
	if len(parts) < 2 {
		return false
	}

	// More detailed validation can be added here
	return true
}

// convertTargetToResponse converts model to response DTO
func (s *iSCSITargetService) convertTargetToResponse(target *models.iSCSITarget) *dto.iSCSITargetResponse {
	response := &dto.iSCSITargetResponse{
		ID:        target.ID,
		Name:      target.Name,
		Alias:     target.Alias,
		Status:    string(target.Status),
		Comment:   target.Comment,
		IsEnabled: target.IsEnabled,
		CreatedAt: target.CreatedAt,
		UpdatedAt: target.UpdatedAt,
	}

	// Count associated resources
	if target.LUNs != nil {
		response.LUNCount = len(target.LUNs)
	}
	if target.ACLs != nil {
		response.ACLCount = len(target.ACLs)
	}

	return response
}