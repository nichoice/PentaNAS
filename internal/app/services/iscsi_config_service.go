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

type iSCSIConfigService struct {
	db *gorm.DB
}

func NewiSCSIConfigService() *iSCSIConfigService {
	return &iSCSIConfigService{db: database.DB}
}

// GetGlobalConfig gets the current iSCSI global configuration
func (s *iSCSIConfigService) GetGlobalConfig() (*dto.iSCSIGlobalConfigResponse, error) {
	var config models.iSCSIGlobalConfig
	if err := s.db.Where("is_active = ?", true).First(&config).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Create default config if none exists
			return s.createDefaultConfig()
		}
		return nil, fmt.Errorf("failed to get global config: %w", err)
	}

	return s.convertConfigToResponse(&config), nil
}

// UpdateGlobalConfig updates the iSCSI global configuration
func (s *iSCSIConfigService) UpdateGlobalConfig(req *dto.UpdateiSCSIGlobalConfigRequest) (*dto.iSCSIGlobalConfigResponse, error) {
	var config models.iSCSIGlobalConfig
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

	// Update fields
	if req.TargetPort != nil {
		config.TargetPort = *req.TargetPort
	}
	if req.MaxSessions != nil {
		config.MaxSessions = *req.MaxSessions
	}
	if req.MaxConnections != nil {
		config.MaxConnections = *req.MaxConnections
	}
	if req.MaxRecvDataSegmentLength != nil {
		config.MaxRecvDataSegmentLength = *req.MaxRecvDataSegmentLength
	}
	if req.MaxXmitDataSegmentLength != nil {
		config.MaxXmitDataSegmentLength = *req.MaxXmitDataSegmentLength
	}
	if req.MaxBurstLength != nil {
		config.MaxBurstLength = *req.MaxBurstLength
	}
	if req.FirstBurstLength != nil {
		config.FirstBurstLength = *req.FirstBurstLength
	}
	if req.MaxOutstandingR2T != nil {
		config.MaxOutstandingR2T = *req.MaxOutstandingR2T
	}
	if req.DefaultTime2Wait != nil {
		config.DefaultTime2Wait = *req.DefaultTime2Wait
	}
	if req.DefaultTime2Retain != nil {
		config.DefaultTime2Retain = *req.DefaultTime2Retain
	}
	if req.LoginTimeout != nil {
		config.LoginTimeout = *req.LoginTimeout
	}
	if req.LogoutTimeout != nil {
		config.LogoutTimeout = *req.LogoutTimeout
	}
	if req.RequireAuth != nil {
		config.RequireAuth = *req.RequireAuth
	}
	if req.AllowDuplicateSessions != nil {
		config.AllowDuplicateSessions = *req.AllowDuplicateSessions
	}
	if req.LogLevel != nil {
		config.LogLevel = *req.LogLevel
	}
	if req.LogFile != "" {
		config.LogFile = req.LogFile
	}
	if req.EnableDebugLog != nil {
		config.EnableDebugLog = *req.EnableDebugLog
	}

	if err := s.db.Save(&config).Error; err != nil {
		return nil, fmt.Errorf("failed to update global config: %w", err)
	}

	return s.convertConfigToResponse(&config), nil
}

// createDefaultConfig creates a default iSCSI configuration
func (s *iSCSIConfigService) createDefaultConfig() (*dto.iSCSIGlobalConfigResponse, error) {
	config := models.iSCSIGlobalConfig{
		Base: models.Base{
			ID: uuid.New().String(),
		},
		TargetPort:                   3260,
		MaxSessions:                  256,
		MaxConnections:               1,
		MaxRecvDataSegmentLength:     8192,
		MaxXmitDataSegmentLength:     8192,
		MaxBurstLength:               262144,
		FirstBurstLength:             65536,
		MaxOutstandingR2T:            1,
		DefaultTime2Wait:             2,
		DefaultTime2Retain:           20,
		LoginTimeout:                 30,
		LogoutTimeout:                30,
		RequireAuth:                  false,
		AllowDuplicateSessions:       false,
		LogLevel:                     1,
		LogFile:                      "/var/log/iscsi/iscsi.log",
		EnableDebugLog:               false,
		IsActive:                     true,
	}

	if err := s.db.Create(&config).Error; err != nil {
		return nil, fmt.Errorf("failed to create default config: %w", err)
	}

	return s.convertConfigToResponse(&config), nil
}

// convertConfigToResponse converts model to response DTO
func (s *iSCSIConfigService) convertConfigToResponse(config *models.iSCSIGlobalConfig) *dto.iSCSIGlobalConfigResponse {
	return &dto.iSCSIGlobalConfigResponse{
		ID:                           config.ID,
		TargetPort:                   config.TargetPort,
		MaxSessions:                  config.MaxSessions,
		MaxConnections:               config.MaxConnections,
		MaxRecvDataSegmentLength:     config.MaxRecvDataSegmentLength,
		MaxXmitDataSegmentLength:     config.MaxXmitDataSegmentLength,
		MaxBurstLength:               config.MaxBurstLength,
		FirstBurstLength:             config.FirstBurstLength,
		MaxOutstandingR2T:            config.MaxOutstandingR2T,
		DefaultTime2Wait:             config.DefaultTime2Wait,
		DefaultTime2Retain:           config.DefaultTime2Retain,
		LoginTimeout:                 config.LoginTimeout,
		LogoutTimeout:                config.LogoutTimeout,
		RequireAuth:                  config.RequireAuth,
		AllowDuplicateSessions:       config.AllowDuplicateSessions,
		LogLevel:                     config.LogLevel,
		LogFile:                      config.LogFile,
		EnableDebugLog:               config.EnableDebugLog,
		IsActive:                     config.IsActive,
		CreatedAt:                    config.CreatedAt,
		UpdatedAt:                    config.UpdatedAt,
	}
}