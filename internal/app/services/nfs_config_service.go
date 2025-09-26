package services

import (
	"errors"
	"fmt"

	"pnas/internal/app/dto"
	"pnas/internal/database"
	"pnas/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type NFSConfigService struct {
	db *gorm.DB
}

func NewNFSConfigService() *NFSConfigService {
	return &NFSConfigService{db: database.DB}
}

// GetGlobalConfig gets the current NFS global configuration
func (s *NFSConfigService) GetGlobalConfig() (*dto.NFSGlobalConfigResponse, error) {
	var config models.NFSGlobalConfig
	if err := s.db.Where("is_active = ?", true).First(&config).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Create default config if none exists
			return s.createDefaultConfig()
		}
		return nil, fmt.Errorf("failed to get global config: %w", err)
	}

	return s.convertConfigToResponse(&config), nil
}

// UpdateGlobalConfig updates the NFS global configuration
func (s *NFSConfigService) UpdateGlobalConfig(req *dto.UpdateNFSGlobalConfigRequest) (*dto.NFSGlobalConfigResponse, error) {
	var config models.NFSGlobalConfig
	if err := s.db.Where("is_active = ?", true).First(&config).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Create default config if none exists
			_, err := s.createDefaultConfig()
			if err != nil {
				return nil, err
			}
			// Get the created config for updating
			if err := s.db.Where("is_active = ?", true).First(&config).Error; err != nil {
				return nil, fmt.Errorf("failed to get created config: %w", err)
			}
		} else {
			return nil, fmt.Errorf("failed to get global config: %w", err)
		}
	}

	// Update fields if provided
	if req.Version != nil {
		config.Version = *req.Version
	}
	if req.PortmapperPort != nil {
		config.PortmapperPort = *req.PortmapperPort
	}
	if req.NFSPort != nil {
		config.NFSPort = *req.NFSPort
	}
	if req.MountdPort != nil {
		config.MountdPort = *req.MountdPort
	}
	if req.StatdPort != nil {
		config.StatdPort = *req.StatdPort
	}
	if req.LockdPort != nil {
		config.LockdPort = *req.LockdPort
	}
	if req.ThreadCount != nil {
		config.ThreadCount = *req.ThreadCount
	}
	if req.MaxConnections != nil {
		config.MaxConnections = *req.MaxConnections
	}
	if req.ReadAhead != nil {
		config.ReadAhead = *req.ReadAhead
	}
	if req.WriteBuffer != nil {
		config.WriteBuffer = *req.WriteBuffer
	}
	if req.AttributeTimeout != nil {
		config.AttributeTimeout = *req.AttributeTimeout
	}
	if req.DirectoryTimeout != nil {
		config.DirectoryTimeout = *req.DirectoryTimeout
	}
	if req.RequireSecurePort != nil {
		config.RequireSecurePort = *req.RequireSecurePort
	}
	if req.EnableTCP != nil {
		config.EnableTCP = *req.EnableTCP
	}
	if req.EnableUDP != nil {
		config.EnableUDP = *req.EnableUDP
	}
	if req.LogLevel != nil {
		config.LogLevel = *req.LogLevel
	}
	if req.LogFile != nil {
		config.LogFile = *req.LogFile
	}
	if req.EnableDebugLog != nil {
		config.EnableDebugLog = *req.EnableDebugLog
	}
	if req.EnableMultipath != nil {
		config.EnableMultipath = *req.EnableMultipath
	}
	if req.MultipathPolicy != nil {
		config.MultipathPolicy = *req.MultipathPolicy
	}
	if req.HealthCheckInterval != nil {
		config.HealthCheckInterval = *req.HealthCheckInterval
	}

	if err := s.db.Save(&config).Error; err != nil {
		return nil, fmt.Errorf("failed to update global config: %w", err)
	}

	return s.convertConfigToResponse(&config), nil
}

// ResetGlobalConfig resets the global configuration to defaults
func (s *NFSConfigService) ResetGlobalConfig() (*dto.NFSGlobalConfigResponse, error) {
	// Deactivate current config
	if err := s.db.Model(&models.NFSGlobalConfig{}).Where("is_active = ?", true).Update("is_active", false).Error; err != nil {
		return nil, fmt.Errorf("failed to deactivate current config: %w", err)
	}

	// Create new default config
	return s.createDefaultConfig()
}

// CreateClientAccess creates a new client access control
func (s *NFSConfigService) CreateClientAccess(req *dto.CreateNFSClientAccessRequest) (*dto.NFSClientAccessResponse, error) {
	// Validate export exists
	var export models.NFSExport
	if err := s.db.Where("id = ?", req.ExportID).First(&export).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("export not found")
		}
		return nil, fmt.Errorf("failed to get export: %w", err)
	}

	// Check for duplicate client access
	var existingAccess models.NFSClientAccess
	if err := s.db.Where("export_id = ? AND client_host = ?", req.ExportID, req.ClientHost).First(&existingAccess).Error; err == nil {
		return nil, fmt.Errorf("client access already exists for host '%s'", req.ClientHost)
	}

	// Create client access
	clientAccess := models.NFSClientAccess{
		Base:           models.Base{ID: uuid.New().String()},
		ExportID:       req.ExportID,
		ClientHost:     req.ClientHost,
		Permission:     req.Permission,
		SecurityFlavor: s.getSecurityFlavorOrDefault(req.SecurityFlavor),
		RootSquash:     req.RootSquash,
		AllSquash:      req.AllSquash,
		AnonUID:        s.getAnonUIDOrDefault(req.AnonUID),
		AnonGID:        s.getAnonGIDOrDefault(req.AnonGID),
	}

	if err := s.db.Create(&clientAccess).Error; err != nil {
		return nil, fmt.Errorf("failed to create client access: %w", err)
	}

	// Regenerate exports file
	exportService := NewNFSExportService()
	if err := exportService.generateExportsFile(); err != nil {
		return nil, fmt.Errorf("failed to regenerate exports file: %w", err)
	}

	return s.convertClientAccessToResponse(&clientAccess), nil
}

// GetClientAccess gets client access controls for an export
func (s *NFSConfigService) GetClientAccess(exportID string) ([]dto.NFSClientAccessResponse, error) {
	// Validate export exists
	var export models.NFSExport
	if err := s.db.Where("id = ?", exportID).First(&export).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("export not found")
		}
		return nil, fmt.Errorf("failed to get export: %w", err)
	}

	var clientAccess []models.NFSClientAccess
	if err := s.db.Where("export_id = ?", exportID).Find(&clientAccess).Error; err != nil {
		return nil, fmt.Errorf("failed to get client access: %w", err)
	}

	responses := make([]dto.NFSClientAccessResponse, len(clientAccess))
	for i, access := range clientAccess {
		responses[i] = *s.convertClientAccessToResponse(&access)
	}

	return responses, nil
}

// UpdateClientAccess updates a client access control
func (s *NFSConfigService) UpdateClientAccess(id string, permission *models.NFSPermission, securityFlavor *models.NFSSecurityFlavor, rootSquash *bool, allSquash *bool) (*dto.NFSClientAccessResponse, error) {
	var clientAccess models.NFSClientAccess
	if err := s.db.Where("id = ?", id).First(&clientAccess).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("client access not found")
		}
		return nil, fmt.Errorf("failed to get client access: %w", err)
	}

	// Update fields if provided
	if permission != nil {
		clientAccess.Permission = *permission
	}
	if securityFlavor != nil {
		clientAccess.SecurityFlavor = *securityFlavor
	}
	if rootSquash != nil {
		clientAccess.RootSquash = *rootSquash
	}
	if allSquash != nil {
		clientAccess.AllSquash = *allSquash
	}

	if err := s.db.Save(&clientAccess).Error; err != nil {
		return nil, fmt.Errorf("failed to update client access: %w", err)
	}

	// Regenerate exports file
	exportService := NewNFSExportService()
	if err := exportService.generateExportsFile(); err != nil {
		return nil, fmt.Errorf("failed to regenerate exports file: %w", err)
	}

	return s.convertClientAccessToResponse(&clientAccess), nil
}

// DeleteClientAccess deletes a client access control
func (s *NFSConfigService) DeleteClientAccess(id string) error {
	var clientAccess models.NFSClientAccess
	if err := s.db.Where("id = ?", id).First(&clientAccess).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("client access not found")
		}
		return fmt.Errorf("failed to get client access: %w", err)
	}

	if err := s.db.Delete(&clientAccess).Error; err != nil {
		return fmt.Errorf("failed to delete client access: %w", err)
	}

	// Regenerate exports file
	exportService := NewNFSExportService()
	if err := exportService.generateExportsFile(); err != nil {
		return fmt.Errorf("failed to regenerate exports file: %w", err)
	}

	return nil
}

// CreateQuota creates a new NFS quota
func (s *NFSConfigService) CreateQuota(req *dto.CreateNFSQuotaRequest) (*dto.NFSQuotaResponse, error) {
	// Validate export exists
	var export models.NFSExport
	if err := s.db.Where("id = ?", req.ExportID).First(&export).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("export not found")
		}
		return nil, fmt.Errorf("failed to get export: %w", err)
	}

	// Check for duplicate quota
	var existingQuota models.NFSQuota
	if err := s.db.Where("export_id = ? AND quota_type = ? AND target_id = ?", req.ExportID, req.QuotaType, req.TargetID).First(&existingQuota).Error; err == nil {
		return nil, fmt.Errorf("quota already exists for target '%s'", req.TargetID)
	}

	// Create quota
	quota := models.NFSQuota{
		Base:           models.Base{ID: uuid.New().String()},
		ExportID:       req.ExportID,
		QuotaType:      req.QuotaType,
		TargetID:       req.TargetID,
		HardLimitSize:  req.HardLimitSize,
		SoftLimitSize:  req.SoftLimitSize,
		HardLimitFiles: req.HardLimitFiles,
		SoftLimitFiles: req.SoftLimitFiles,
		GracePeriod:    s.getGracePeriodOrDefault(req.GracePeriod),
		IsEnabled:      true,
	}

	if err := s.db.Create(&quota).Error; err != nil {
		return nil, fmt.Errorf("failed to create quota: %w", err)
	}

	return s.convertQuotaToResponse(&quota), nil
}

// GetQuotas gets quotas for an export
func (s *NFSConfigService) GetQuotas(exportID string) ([]dto.NFSQuotaResponse, error) {
	// Validate export exists
	var export models.NFSExport
	if err := s.db.Where("id = ?", exportID).First(&export).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("export not found")
		}
		return nil, fmt.Errorf("failed to get export: %w", err)
	}

	var quotas []models.NFSQuota
	if err := s.db.Where("export_id = ?", exportID).Find(&quotas).Error; err != nil {
		return nil, fmt.Errorf("failed to get quotas: %w", err)
	}

	responses := make([]dto.NFSQuotaResponse, len(quotas))
	for i, quota := range quotas {
		responses[i] = *s.convertQuotaToResponse(&quota)
	}

	return responses, nil
}

// Helper functions

func (s *NFSConfigService) createDefaultConfig() (*dto.NFSGlobalConfigResponse, error) {
	config := models.NFSGlobalConfig{
		Base:                models.Base{ID: uuid.New().String()},
		Version:             models.NFSVersion4,
		PortmapperPort:      111,
		NFSPort:             2049,
		MountdPort:          20048,
		StatdPort:           662,
		LockdPort:           32803,
		ThreadCount:         8,
		MaxConnections:      1024,
		ReadAhead:           128,
		WriteBuffer:         128,
		AttributeTimeout:    60,
		DirectoryTimeout:    60,
		RequireSecurePort:   false,
		EnableTCP:           true,
		EnableUDP:           false,
		LogLevel:            1,
		LogFile:             "/var/log/nfs.log",
		EnableDebugLog:      false,
		EnableMultipath:     false,
		MultipathPolicy:     "round_robin",
		HealthCheckInterval: 30,
		IsActive:            true,
	}

	if err := s.db.Create(&config).Error; err != nil {
		return nil, fmt.Errorf("failed to create default config: %w", err)
	}

	return s.convertConfigToResponse(&config), nil
}

func (s *NFSConfigService) convertConfigToResponse(config *models.NFSGlobalConfig) *dto.NFSGlobalConfigResponse {
	return &dto.NFSGlobalConfigResponse{
		ID:                  config.ID,
		Version:             config.Version,
		PortmapperPort:      config.PortmapperPort,
		NFSPort:             config.NFSPort,
		MountdPort:          config.MountdPort,
		StatdPort:           config.StatdPort,
		LockdPort:           config.LockdPort,
		ThreadCount:         config.ThreadCount,
		MaxConnections:      config.MaxConnections,
		ReadAhead:           config.ReadAhead,
		WriteBuffer:         config.WriteBuffer,
		AttributeTimeout:    config.AttributeTimeout,
		DirectoryTimeout:    config.DirectoryTimeout,
		RequireSecurePort:   config.RequireSecurePort,
		EnableTCP:           config.EnableTCP,
		EnableUDP:           config.EnableUDP,
		LogLevel:            config.LogLevel,
		LogFile:             config.LogFile,
		EnableDebugLog:      config.EnableDebugLog,
		EnableMultipath:     config.EnableMultipath,
		MultipathPolicy:     config.MultipathPolicy,
		HealthCheckInterval: config.HealthCheckInterval,
		IsActive:            config.IsActive,
		CreatedAt:           config.CreatedAt,
		UpdatedAt:           config.UpdatedAt,
	}
}

func (s *NFSConfigService) convertClientAccessToResponse(access *models.NFSClientAccess) *dto.NFSClientAccessResponse {
	return &dto.NFSClientAccessResponse{
		ID:             access.ID,
		ExportID:       access.ExportID,
		ClientHost:     access.ClientHost,
		Permission:     access.Permission,
		SecurityFlavor: access.SecurityFlavor,
		RootSquash:     access.RootSquash,
		AllSquash:      access.AllSquash,
		AnonUID:        access.AnonUID,
		AnonGID:        access.AnonGID,
		CreatedAt:      access.CreatedAt,
		UpdatedAt:      access.UpdatedAt,
	}
}

func (s *NFSConfigService) convertQuotaToResponse(quota *models.NFSQuota) *dto.NFSQuotaResponse {
	return &dto.NFSQuotaResponse{
		ID:             quota.ID,
		ExportID:       quota.ExportID,
		QuotaType:      quota.QuotaType,
		TargetID:       quota.TargetID,
		HardLimitSize:  quota.HardLimitSize,
		SoftLimitSize:  quota.SoftLimitSize,
		HardLimitFiles: quota.HardLimitFiles,
		SoftLimitFiles: quota.SoftLimitFiles,
		CurrentSize:    quota.CurrentSize,
		CurrentFiles:   quota.CurrentFiles,
		GracePeriod:    quota.GracePeriod,
		IsEnabled:      quota.IsEnabled,
		CreatedAt:      quota.CreatedAt,
		UpdatedAt:      quota.UpdatedAt,
	}
}

func (s *NFSConfigService) getSecurityFlavorOrDefault(flavor models.NFSSecurityFlavor) models.NFSSecurityFlavor {
	if flavor == "" {
		return models.SecSys
	}
	return flavor
}

func (s *NFSConfigService) getAnonUIDOrDefault(uid int) int {
	if uid == 0 {
		return 65534
	}
	return uid
}

func (s *NFSConfigService) getAnonGIDOrDefault(gid int) int {
	if gid == 0 {
		return 65534
	}
	return gid
}

func (s *NFSConfigService) getGracePeriodOrDefault(period int) int {
	if period == 0 {
		return 7
	}
	return period
}