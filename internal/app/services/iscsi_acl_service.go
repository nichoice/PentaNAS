package services

import (
	"errors"
	"fmt"
	"strings"

	"pnas/internal/app/dto"
	"pnas/internal/database"
	"pnas/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type iSCSIACLService struct {
	db     *gorm.DB
	cmdSvc *iSCSICommandService
}

func NewiSCSIACLService() *iSCSIACLService {
	return &iSCSIACLService{
		db:     database.DB,
		cmdSvc: NewiSCSICommandService(),
	}
}

// CreateACL creates a new ACL for an iSCSI target
func (s *iSCSIACLService) CreateACL(req *dto.CreateiSCSIACLRequest) (*dto.iSCSIACLResponse, error) {
	// Verify target exists
	var target models.iSCSITarget
	if err := s.db.First(&target, "id = ?", req.TargetID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("target not found")
		}
		return nil, fmt.Errorf("failed to get target: %w", err)
	}

	// Validate IQN format
	if !s.isValidIQN(req.InitiatorName) {
		return nil, fmt.Errorf("invalid initiator IQN format: %s", req.InitiatorName)
	}

	// Check if ACL already exists for this target and initiator
	var existingACL models.iSCSIACL
	if err := s.db.Where("target_id = ? AND initiator_name = ?", req.TargetID, req.InitiatorName).First(&existingACL).Error; err == nil {
		return nil, fmt.Errorf("ACL for initiator %s already exists on target %s", req.InitiatorName, target.Name)
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to check ACL existence: %w", err)
	}

	// Validate authentication parameters
	if req.AuthType == "chap" {
		if req.Username == "" || req.Password == "" {
			return nil, fmt.Errorf("username and password are required for CHAP authentication")
		}
		if len(req.Password) < 8 || len(req.Password) > 16 {
			return nil, fmt.Errorf("CHAP password must be between 8 and 16 characters")
		}
		if req.MutualAuth && (req.MutualUsername == "" || req.MutualPassword == "") {
			return nil, fmt.Errorf("mutual username and password are required for mutual CHAP authentication")
		}
	}

	// Create ACL record
	acl := models.iSCSIACL{
		Base: models.Base{
			ID: uuid.New().String(),
		},
		TargetID:       req.TargetID,
		InitiatorName:  req.InitiatorName,
		Permission:     models.iSCSIPermission(req.Permission),
		IsEnabled:      true,
		AuthType:       models.iSCSIAuthType(req.AuthType),
		Username:       req.Username,
		Password:       req.Password,
		MutualAuth:     req.MutualAuth,
		MutualUsername: req.MutualUsername,
		MutualPassword: req.MutualPassword,
		Comment:        req.Comment,
	}

	if err := s.db.Create(&acl).Error; err != nil {
		return nil, fmt.Errorf("failed to create ACL: %w", err)
	}

	// Apply to system if target is active
	if target.Status == models.iSCSIStatusActive {
		if err := s.cmdSvc.CreateACL(target.Name, req.InitiatorName); err != nil {
			fmt.Printf("Warning: failed to create ACL in system: %v\n", err)
		} else {
			// Set authentication if required
			if req.AuthType == "chap" {
				if err := s.cmdSvc.SetACLAuth(target.Name, req.InitiatorName, req.Username, req.Password, req.MutualAuth, req.MutualUsername, req.MutualPassword); err != nil {
					fmt.Printf("Warning: failed to set ACL authentication: %v\n", err)
				}
			}
		}
	}

	return s.convertACLToResponse(&acl), nil
}

// GetACL gets an ACL by ID
func (s *iSCSIACLService) GetACL(id string) (*dto.iSCSIACLResponse, error) {
	var acl models.iSCSIACL
	if err := s.db.Preload("Target").Preload("LUNMappings").First(&acl, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("ACL not found")
		}
		return nil, fmt.Errorf("failed to get ACL: %w", err)
	}

	return s.convertACLToResponse(&acl), nil
}

// ListACLs lists ACLs with pagination and filters
func (s *iSCSIACLService) ListACLs(query *dto.iSCSIACLListQuery) ([]*dto.iSCSIACLResponse, int64, error) {
	var acls []models.iSCSIACL
	var total int64

	db := s.db.Model(&models.iSCSIACL{}).Preload("Target").Preload("LUNMappings")

	// Apply filters
	if query.TargetID != "" {
		db = db.Where("target_id = ?", query.TargetID)
	}
	if query.Permission != "" {
		db = db.Where("permission = ?", query.Permission)
	}
	if query.AuthType != "" {
		db = db.Where("auth_type = ?", query.AuthType)
	}
	if query.IsEnabled != nil {
		db = db.Where("is_enabled = ?", *query.IsEnabled)
	}

	// Get total count
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count ACLs: %w", err)
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
	if err := db.Offset(offset).Limit(pageSize).Find(&acls).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list ACLs: %w", err)
	}

	responses := make([]*dto.iSCSIACLResponse, len(acls))
	for i, acl := range acls {
		responses[i] = s.convertACLToResponse(&acl)
	}

	return responses, total, nil
}

// UpdateACL updates an existing ACL
func (s *iSCSIACLService) UpdateACL(id string, req *dto.UpdateiSCSIACLRequest) (*dto.iSCSIACLResponse, error) {
	var acl models.iSCSIACL
	if err := s.db.Preload("Target").First(&acl, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("ACL not found")
		}
		return nil, fmt.Errorf("failed to get ACL: %w", err)
	}

	oldAuthType := acl.AuthType

	// Update fields
	if req.Permission != "" {
		acl.Permission = models.iSCSIPermission(req.Permission)
	}
	if req.AuthType != "" {
		acl.AuthType = models.iSCSIAuthType(req.AuthType)
	}
	if req.Username != "" {
		acl.Username = req.Username
	}
	if req.Password != "" {
		// Validate password for CHAP
		if acl.AuthType == models.iSCSIAuthCHAP {
			if len(req.Password) < 8 || len(req.Password) > 16 {
				return nil, fmt.Errorf("CHAP password must be between 8 and 16 characters")
			}
		}
		acl.Password = req.Password
	}
	if req.MutualAuth != nil {
		acl.MutualAuth = *req.MutualAuth
	}
	if req.MutualUsername != "" {
		acl.MutualUsername = req.MutualUsername
	}
	if req.MutualPassword != "" {
		acl.MutualPassword = req.MutualPassword
	}
	if req.IsEnabled != nil {
		acl.IsEnabled = *req.IsEnabled
	}
	if req.Comment != "" {
		acl.Comment = req.Comment
	}

	// Validate authentication parameters after update
	if acl.AuthType == models.iSCSIAuthCHAP {
		if acl.Username == "" || acl.Password == "" {
			return nil, fmt.Errorf("username and password are required for CHAP authentication")
		}
		if acl.MutualAuth && (acl.MutualUsername == "" || acl.MutualPassword == "") {
			return nil, fmt.Errorf("mutual username and password are required for mutual CHAP authentication")
		}
	}

	if err := s.db.Save(&acl).Error; err != nil {
		return nil, fmt.Errorf("failed to update ACL: %w", err)
	}

	// Update system configuration if target is active
	if acl.Target.Status == models.iSCSIStatusActive {
		// Handle authentication changes
		if oldAuthType != acl.AuthType || req.Password != "" || req.Username != "" {
			if acl.AuthType == models.iSCSIAuthCHAP {
				if err := s.cmdSvc.SetACLAuth(acl.Target.Name, acl.InitiatorName, acl.Username, acl.Password, acl.MutualAuth, acl.MutualUsername, acl.MutualPassword); err != nil {
					fmt.Printf("Warning: failed to update ACL authentication: %v\n", err)
				}
			} else if oldAuthType == models.iSCSIAuthCHAP {
				// Clear authentication when switching from CHAP to none
				if err := s.cmdSvc.ClearACLAuth(acl.Target.Name, acl.InitiatorName); err != nil {
					fmt.Printf("Warning: failed to clear ACL authentication: %v\n", err)
				}
			}
		}
	}

	return s.convertACLToResponse(&acl), nil
}

// DeleteACL deletes an ACL
func (s *iSCSIACLService) DeleteACL(id string) error {
	var acl models.iSCSIACL
	if err := s.db.Preload("Target").Preload("LUNMappings").First(&acl, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("ACL not found")
		}
		return fmt.Errorf("failed to get ACL: %w", err)
	}

	// Delete associated LUN mappings first
	if len(acl.LUNMappings) > 0 {
		if err := s.db.Where("acl_id = ?", id).Delete(&models.iSCSILUNMapping{}).Error; err != nil {
			return fmt.Errorf("failed to delete LUN mappings: %w", err)
		}

		// Remove mappings from system
		if acl.Target.Status == models.iSCSIStatusActive {
			for _, mapping := range acl.LUNMappings {
				var lun models.iSCSILUN
				if s.db.First(&lun, "id = ?", mapping.LUNID).Error == nil {
					s.cmdSvc.UnmapLUNFromACL(acl.Target.Name, acl.InitiatorName, lun.LUN)
				}
			}
		}
	}

	// Remove ACL from system if target is active
	if acl.Target.Status == models.iSCSIStatusActive {
		if err := s.cmdSvc.DeleteACL(acl.Target.Name, acl.InitiatorName); err != nil {
			fmt.Printf("Warning: failed to remove ACL from system: %v\n", err)
		}
	}

	if err := s.db.Delete(&acl).Error; err != nil {
		return fmt.Errorf("failed to delete ACL: %w", err)
	}

	return nil
}

// BatchOperationACLs performs batch operations on ACLs
func (s *iSCSIACLService) BatchOperationACLs(req *dto.BatchiSCSIACLRequest) (*dto.BatchOperationResponse, error) {
	response := &dto.BatchOperationResponse{
		Success: []string{},
		Failed:  []string{},
		Errors:  []string{},
	}

	for _, aclID := range req.ACLIDs {
		var acl models.iSCSIACL
		if err := s.db.Preload("Target").Preload("LUNMappings").First(&acl, "id = ?", aclID).Error; err != nil {
			response.Failed = append(response.Failed, aclID)
			response.Errors = append(response.Errors, fmt.Sprintf("ACL %s: not found", aclID))
			continue
		}

		switch req.Action {
		case "enable":
			if !acl.IsEnabled {
				acl.IsEnabled = true
				if err := s.db.Save(&acl).Error; err != nil {
					response.Failed = append(response.Failed, aclID)
					response.Errors = append(response.Errors, fmt.Sprintf("ACL %s: failed to enable", aclID))
				} else {
					response.Success = append(response.Success, aclID)
				}
			} else {
				response.Success = append(response.Success, aclID)
			}

		case "disable":
			if acl.IsEnabled {
				acl.IsEnabled = false
				if err := s.db.Save(&acl).Error; err != nil {
					response.Failed = append(response.Failed, aclID)
					response.Errors = append(response.Errors, fmt.Sprintf("ACL %s: failed to disable", aclID))
				} else {
					response.Success = append(response.Success, aclID)
				}
			} else {
				response.Success = append(response.Success, aclID)
			}

		case "delete":
			// Delete LUN mappings first
			if len(acl.LUNMappings) > 0 {
				s.db.Where("acl_id = ?", aclID).Delete(&models.iSCSILUNMapping{})

				// Remove mappings from system
				if acl.Target.Status == models.iSCSIStatusActive {
					for _, mapping := range acl.LUNMappings {
						var lun models.iSCSILUN
						if s.db.First(&lun, "id = ?", mapping.LUNID).Error == nil {
							s.cmdSvc.UnmapLUNFromACL(acl.Target.Name, acl.InitiatorName, lun.LUN)
						}
					}
				}
			}

			// Remove from system
			if acl.Target.Status == models.iSCSIStatusActive {
				s.cmdSvc.DeleteACL(acl.Target.Name, acl.InitiatorName)
			}

			if err := s.db.Delete(&acl).Error; err != nil {
				response.Failed = append(response.Failed, aclID)
				response.Errors = append(response.Errors, fmt.Sprintf("ACL %s: failed to delete", aclID))
			} else {
				response.Success = append(response.Success, aclID)
			}

		default:
			response.Failed = append(response.Failed, aclID)
			response.Errors = append(response.Errors, fmt.Sprintf("ACL %s: invalid action %s", aclID, req.Action))
		}
	}

	return response, nil
}

// GetACLsByTarget gets all ACLs for a specific target
func (s *iSCSIACLService) GetACLsByTarget(targetID string) ([]*dto.iSCSIACLResponse, error) {
	var acls []models.iSCSIACL
	if err := s.db.Preload("LUNMappings").Where("target_id = ?", targetID).Find(&acls).Error; err != nil {
		return nil, fmt.Errorf("failed to get ACLs for target: %w", err)
	}

	responses := make([]*dto.iSCSIACLResponse, len(acls))
	for i, acl := range acls {
		responses[i] = s.convertACLToResponse(&acl)
	}

	return responses, nil
}

// CreateLUNMapping creates a LUN mapping for an ACL
func (s *iSCSIACLService) CreateLUNMapping(req *dto.CreateiSCSILUNMappingRequest) (*dto.iSCSILUNMappingResponse, error) {
	// Verify ACL exists
	var acl models.iSCSIACL
	if err := s.db.Preload("Target").First(&acl, "id = ?", req.ACLID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("ACL not found")
		}
		return nil, fmt.Errorf("failed to get ACL: %w", err)
	}

	// Verify LUN exists
	var lun models.iSCSILUN
	if err := s.db.First(&lun, "id = ?", req.LUNID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("LUN not found")
		}
		return nil, fmt.Errorf("failed to get LUN: %w", err)
	}

	// Verify LUN belongs to the same target as ACL
	if lun.TargetID != acl.TargetID {
		return nil, fmt.Errorf("LUN and ACL must belong to the same target")
	}

	// Check if mapping already exists
	var existingMapping models.iSCSILUNMapping
	if err := s.db.Where("acl_id = ? AND lun_id = ?", req.ACLID, req.LUNID).First(&existingMapping).Error; err == nil {
		return nil, fmt.Errorf("LUN mapping already exists")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to check mapping existence: %w", err)
	}

	// Create mapping
	mapping := models.iSCSILUNMapping{
		Base: models.Base{
			ID: uuid.New().String(),
		},
		ACLID:      req.ACLID,
		LUNID:      req.LUNID,
		Permission: models.iSCSIPermission(req.Permission),
		IsEnabled:  true,
	}

	if err := s.db.Create(&mapping).Error; err != nil {
		return nil, fmt.Errorf("failed to create LUN mapping: %w", err)
	}

	// Apply to system if target is active
	if acl.Target.Status == models.iSCSIStatusActive {
		if err := s.cmdSvc.MapLUNToACL(acl.Target.Name, acl.InitiatorName, lun.LUN, req.Permission); err != nil {
			fmt.Printf("Warning: failed to map LUN in system: %v\n", err)
		}
	}

	return s.convertLUNMappingToResponse(&mapping), nil
}

// UpdateLUNMapping updates a LUN mapping
func (s *iSCSIACLService) UpdateLUNMapping(id string, req *dto.UpdateiSCSILUNMappingRequest) (*dto.iSCSILUNMappingResponse, error) {
	var mapping models.iSCSILUNMapping
	if err := s.db.Preload("ACL.Target").Preload("LUN").First(&mapping, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("LUN mapping not found")
		}
		return nil, fmt.Errorf("failed to get LUN mapping: %w", err)
	}

	// Update fields
	if req.Permission != "" {
		mapping.Permission = models.iSCSIPermission(req.Permission)
	}
	if req.IsEnabled != nil {
		mapping.IsEnabled = *req.IsEnabled
	}

	if err := s.db.Save(&mapping).Error; err != nil {
		return nil, fmt.Errorf("failed to update LUN mapping: %w", err)
	}

	// Update system configuration if target is active
	if mapping.ACL.Target.Status == models.iSCSIStatusActive {
		// Remove and recreate mapping to update permissions
		s.cmdSvc.UnmapLUNFromACL(mapping.ACL.Target.Name, mapping.ACL.InitiatorName, mapping.LUN.LUN)
		if mapping.IsEnabled {
			s.cmdSvc.MapLUNToACL(mapping.ACL.Target.Name, mapping.ACL.InitiatorName, mapping.LUN.LUN, string(mapping.Permission))
		}
	}

	return s.convertLUNMappingToResponse(&mapping), nil
}

// DeleteLUNMapping deletes a LUN mapping
func (s *iSCSIACLService) DeleteLUNMapping(id string) error {
	var mapping models.iSCSILUNMapping
	if err := s.db.Preload("ACL.Target").Preload("LUN").First(&mapping, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("LUN mapping not found")
		}
		return fmt.Errorf("failed to get LUN mapping: %w", err)
	}

	// Remove from system if target is active
	if mapping.ACL.Target.Status == models.iSCSIStatusActive {
		if err := s.cmdSvc.UnmapLUNFromACL(mapping.ACL.Target.Name, mapping.ACL.InitiatorName, mapping.LUN.LUN); err != nil {
			fmt.Printf("Warning: failed to unmap LUN from system: %v\n", err)
		}
	}

	if err := s.db.Delete(&mapping).Error; err != nil {
		return fmt.Errorf("failed to delete LUN mapping: %w", err)
	}

	return nil
}

// isValidIQN validates IQN format
func (s *iSCSIACLService) isValidIQN(iqn string) bool {
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

// convertACLToResponse converts model to response DTO
func (s *iSCSIACLService) convertACLToResponse(acl *models.iSCSIACL) *dto.iSCSIACLResponse {
	response := &dto.iSCSIACLResponse{
		ID:             acl.ID,
		TargetID:       acl.TargetID,
		InitiatorName:  acl.InitiatorName,
		Permission:     string(acl.Permission),
		AuthType:       string(acl.AuthType),
		Username:       acl.Username,
		MutualAuth:     acl.MutualAuth,
		MutualUsername: acl.MutualUsername,
		IsEnabled:      acl.IsEnabled,
		Comment:        acl.Comment,
		CreatedAt:      acl.CreatedAt,
		UpdatedAt:      acl.UpdatedAt,
	}

	// Convert LUN mappings
	if acl.LUNMappings != nil {
		response.LUNMappings = make([]dto.iSCSILUNMappingResponse, len(acl.LUNMappings))
		for i, mapping := range acl.LUNMappings {
			response.LUNMappings[i] = *s.convertLUNMappingToResponse(&mapping)
		}
	}

	return response
}

// convertLUNMappingToResponse converts model to response DTO
func (s *iSCSIACLService) convertLUNMappingToResponse(mapping *models.iSCSILUNMapping) *dto.iSCSILUNMappingResponse {
	return &dto.iSCSILUNMappingResponse{
		ID:         mapping.ID,
		ACLID:      mapping.ACLID,
		LUNID:      mapping.LUNID,
		Permission: string(mapping.Permission),
		IsEnabled:  mapping.IsEnabled,
		CreatedAt:  mapping.CreatedAt,
		UpdatedAt:  mapping.UpdatedAt,
	}
}