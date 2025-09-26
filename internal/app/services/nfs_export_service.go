package services

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"pnas/internal/app/dto"
	"pnas/internal/database"
	"pnas/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type NFSExportService struct {
	db *gorm.DB
}

func NewNFSExportService() *NFSExportService {
	return &NFSExportService{db: database.DB}
}

// CreateExport creates a new NFS export with permission validation
func (s *NFSExportService) CreateExport(req *dto.CreateNFSExportRequest) (*dto.NFSExportResponse, error) {
	// Validate path exists and is accessible
	if err := s.validateExportPath(req.Path); err != nil {
		return nil, fmt.Errorf("invalid export path: %w", err)
	}

	// Check for duplicate export name
	var existingExport models.NFSExport
	if err := s.db.Where("name = ?", req.Name).First(&existingExport).Error; err == nil {
		return nil, fmt.Errorf("export with name '%s' already exists", req.Name)
	}

	// Create export
	export := models.NFSExport{
		Base:             models.Base{ID: uuid.New().String()},
		Path:             req.Path,
		Name:             req.Name,
		Comment:          req.Comment,
		IsEnabled:        true,
		Version:          s.getVersionOrDefault(req.Version),
		SecurityFlavor:   s.getSecurityFlavorOrDefault(req.SecurityFlavor),
		Permission:       s.getPermissionOrDefault(req.Permission),
		OperationMode:    s.getOperationModeOrDefault(req.OperationMode),
		RootSquash:       req.RootSquash,
		AllSquash:        req.AllSquash,
		AnonUID:          s.getAnonUIDOrDefault(req.AnonUID),
		AnonGID:          s.getAnonGIDOrDefault(req.AnonGID),
		AllowedHosts:     req.AllowedHosts,
		DeniedHosts:      req.DeniedHosts,
		ReadSize:         s.getReadSizeOrDefault(req.ReadSize),
		WriteSize:        s.getWriteSizeOrDefault(req.WriteSize),
		SubtreeCheck:     req.SubtreeCheck,
		EnableVersioning: req.EnableVersioning,
		MaxVersions:      s.getMaxVersionsOrDefault(req.MaxVersions),
		VersionRetention: s.getVersionRetentionOrDefault(req.VersionRetention),
		EnableMultipath:  req.EnableMultipath,
		MultipathPolicy:  s.getMultipathPolicyOrDefault(req.MultipathPolicy),
	}

	if err := s.db.Create(&export).Error; err != nil {
		return nil, fmt.Errorf("failed to create export: %w", err)
	}

	// Generate exports file
	if err := s.generateExportsFile(); err != nil {
		return nil, fmt.Errorf("failed to generate exports file: %w", err)
	}

	return s.convertToResponse(&export), nil
}

// GetExports gets all NFS exports with optional filtering
func (s *NFSExportService) GetExports(enabled *bool) ([]dto.NFSExportResponse, error) {
	var exports []models.NFSExport
	query := s.db.Preload("ClientAccess").Preload("Snapshots").Preload("MultipathConf")

	if enabled != nil {
		query = query.Where("is_enabled = ?", *enabled)
	}

	if err := query.Find(&exports).Error; err != nil {
		return nil, fmt.Errorf("failed to get exports: %w", err)
	}

	responses := make([]dto.NFSExportResponse, len(exports))
	for i, export := range exports {
		responses[i] = *s.convertToResponse(&export)
	}

	return responses, nil
}

// GetExport gets a specific NFS export by ID
func (s *NFSExportService) GetExport(id string) (*dto.NFSExportResponse, error) {
	var export models.NFSExport
	if err := s.db.Preload("ClientAccess").Preload("Snapshots").Preload("MultipathConf").
		Where("id = ?", id).First(&export).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("export not found")
		}
		return nil, fmt.Errorf("failed to get export: %w", err)
	}

	return s.convertToResponse(&export), nil
}

// UpdateExport updates an NFS export
func (s *NFSExportService) UpdateExport(id string, req *dto.UpdateNFSExportRequest) (*dto.NFSExportResponse, error) {
	var export models.NFSExport
	if err := s.db.Where("id = ?", id).First(&export).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("export not found")
		}
		return nil, fmt.Errorf("failed to get export: %w", err)
	}

	// Update fields if provided
	if req.Comment != nil {
		export.Comment = *req.Comment
	}
	if req.IsEnabled != nil {
		export.IsEnabled = *req.IsEnabled
	}
	if req.Version != nil {
		export.Version = *req.Version
	}
	if req.SecurityFlavor != nil {
		export.SecurityFlavor = *req.SecurityFlavor
	}
	if req.Permission != nil {
		export.Permission = *req.Permission
	}
	if req.OperationMode != nil {
		export.OperationMode = *req.OperationMode
	}
	if req.RootSquash != nil {
		export.RootSquash = *req.RootSquash
	}
	if req.AllSquash != nil {
		export.AllSquash = *req.AllSquash
	}
	if req.AnonUID != nil {
		export.AnonUID = *req.AnonUID
	}
	if req.AnonGID != nil {
		export.AnonGID = *req.AnonGID
	}
	if req.AllowedHosts != nil {
		export.AllowedHosts = *req.AllowedHosts
	}
	if req.DeniedHosts != nil {
		export.DeniedHosts = *req.DeniedHosts
	}
	if req.ReadSize != nil {
		export.ReadSize = *req.ReadSize
	}
	if req.WriteSize != nil {
		export.WriteSize = *req.WriteSize
	}
	if req.SubtreeCheck != nil {
		export.SubtreeCheck = *req.SubtreeCheck
	}
	if req.EnableVersioning != nil {
		export.EnableVersioning = *req.EnableVersioning
	}
	if req.MaxVersions != nil {
		export.MaxVersions = *req.MaxVersions
	}
	if req.VersionRetention != nil {
		export.VersionRetention = *req.VersionRetention
	}
	if req.EnableMultipath != nil {
		export.EnableMultipath = *req.EnableMultipath
	}
	if req.MultipathPolicy != nil {
		export.MultipathPolicy = *req.MultipathPolicy
	}

	if err := s.db.Save(&export).Error; err != nil {
		return nil, fmt.Errorf("failed to update export: %w", err)
	}

	// Regenerate exports file
	if err := s.generateExportsFile(); err != nil {
		return nil, fmt.Errorf("failed to regenerate exports file: %w", err)
	}

	return s.convertToResponse(&export), nil
}

// DeleteExport deletes an NFS export
func (s *NFSExportService) DeleteExport(id string) error {
	var export models.NFSExport
	if err := s.db.Where("id = ?", id).First(&export).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("export not found")
		}
		return fmt.Errorf("failed to get export: %w", err)
	}

	// Delete related records first
	if err := s.db.Where("export_id = ?", id).Delete(&models.NFSClientAccess{}).Error; err != nil {
		return fmt.Errorf("failed to delete client access records: %w", err)
	}
	if err := s.db.Where("export_id = ?", id).Delete(&models.NFSSnapshot{}).Error; err != nil {
		return fmt.Errorf("failed to delete snapshots: %w", err)
	}
	if err := s.db.Where("export_id = ?", id).Delete(&models.NFSMultipathConf{}).Error; err != nil {
		return fmt.Errorf("failed to delete multipath configs: %w", err)
	}
	if err := s.db.Where("export_id = ?", id).Delete(&models.NFSQuota{}).Error; err != nil {
		return fmt.Errorf("failed to delete quotas: %w", err)
	}

	// Delete export
	if err := s.db.Delete(&export).Error; err != nil {
		return fmt.Errorf("failed to delete export: %w", err)
	}

	// Regenerate exports file
	if err := s.generateExportsFile(); err != nil {
		return fmt.Errorf("failed to regenerate exports file: %w", err)
	}

	return nil
}

// validateExportPath validates that the export path exists and is accessible
func (s *NFSExportService) validateExportPath(path string) error {
	// Check if path exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("path does not exist: %s", path)
	} else if err != nil {
		return fmt.Errorf("cannot access path: %w", err)
	}

	// Check if path is absolute
	if !filepath.IsAbs(path) {
		return fmt.Errorf("path must be absolute: %s", path)
	}

	// Check if path is readable
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("path is not readable: %w", err)
	}
	file.Close()

	return nil
}

// generateExportsFile generates the /etc/exports file from database
func (s *NFSExportService) generateExportsFile() error {
	var exports []models.NFSExport
	if err := s.db.Preload("ClientAccess").Where("is_enabled = ?", true).Find(&exports).Error; err != nil {
		return fmt.Errorf("failed to get enabled exports: %w", err)
	}

	var lines []string
	lines = append(lines, "# Generated by PNAS - Do not edit manually")
	lines = append(lines, fmt.Sprintf("# Generated at: %s", time.Now().Format(time.RFC3339)))
	lines = append(lines, "")

	for _, export := range exports {
		line := s.generateExportLine(&export)
		if line != "" {
			lines = append(lines, line)
		}
	}

	content := strings.Join(lines, "\n") + "\n"

	// Write to exports file (in production, this would be /etc/exports)
	// For now, write to a test location
	exportsPath := "/tmp/nfs_exports"
	if err := os.WriteFile(exportsPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write exports file: %w", err)
	}

	return nil
}

// generateExportLine generates a single export line for /etc/exports
func (s *NFSExportService) generateExportLine(export *models.NFSExport) string {
	if len(export.ClientAccess) == 0 {
		// If no specific client access, use global settings
		options := s.buildExportOptions(export, "", export.Permission, export.SecurityFlavor,
			export.RootSquash, export.AllSquash, export.AnonUID, export.AnonGID)

		hosts := export.AllowedHosts
		if hosts == "" {
			hosts = "*"
		}

		return fmt.Sprintf("%s %s(%s)", export.Path, hosts, options)
	}

	var parts []string
	for _, access := range export.ClientAccess {
		options := s.buildExportOptions(export, access.ClientHost, access.Permission,
			access.SecurityFlavor, access.RootSquash, access.AllSquash, access.AnonUID, access.AnonGID)
		parts = append(parts, fmt.Sprintf("%s(%s)", access.ClientHost, options))
	}

	if len(parts) > 0 {
		return fmt.Sprintf("%s %s", export.Path, strings.Join(parts, " "))
	}

	return ""
}

// buildExportOptions builds the options string for an export
func (s *NFSExportService) buildExportOptions(export *models.NFSExport, clientHost string,
	permission models.NFSPermission, securityFlavor models.NFSSecurityFlavor,
	rootSquash, allSquash bool, anonUID, anonGID int) string {

	var options []string

	// Permission
	options = append(options, string(permission))

	// Operation mode
	options = append(options, string(export.OperationMode))

	// Squashing options
	if rootSquash {
		options = append(options, "root_squash")
	} else {
		options = append(options, "no_root_squash")
	}

	if allSquash {
		options = append(options, "all_squash")
	}

	// Anonymous user/group mapping
	if anonUID != 65534 {
		options = append(options, fmt.Sprintf("anonuid=%d", anonUID))
	}
	if anonGID != 65534 {
		options = append(options, fmt.Sprintf("anongid=%d", anonGID))
	}

	// Security flavor
	if securityFlavor != models.SecSys {
		options = append(options, fmt.Sprintf("sec=%s", string(securityFlavor)))
	}

	// Subtree check
	if export.SubtreeCheck {
		options = append(options, "subtree_check")
	} else {
		options = append(options, "no_subtree_check")
	}

	// Buffer sizes
	if export.ReadSize != 131072 {
		options = append(options, fmt.Sprintf("rsize=%d", export.ReadSize))
	}
	if export.WriteSize != 131072 {
		options = append(options, fmt.Sprintf("wsize=%d", export.WriteSize))
	}

	return strings.Join(options, ",")
}

// convertToResponse converts model to DTO response
func (s *NFSExportService) convertToResponse(export *models.NFSExport) *dto.NFSExportResponse {
	response := &dto.NFSExportResponse{
		ID:               export.ID,
		Path:             export.Path,
		Name:             export.Name,
		Comment:          export.Comment,
		IsEnabled:        export.IsEnabled,
		Version:          export.Version,
		SecurityFlavor:   export.SecurityFlavor,
		Permission:       export.Permission,
		OperationMode:    export.OperationMode,
		RootSquash:       export.RootSquash,
		AllSquash:        export.AllSquash,
		AnonUID:          export.AnonUID,
		AnonGID:          export.AnonGID,
		AllowedHosts:     export.AllowedHosts,
		DeniedHosts:      export.DeniedHosts,
		ReadSize:         export.ReadSize,
		WriteSize:        export.WriteSize,
		SubtreeCheck:     export.SubtreeCheck,
		EnableVersioning: export.EnableVersioning,
		MaxVersions:      export.MaxVersions,
		VersionRetention: export.VersionRetention,
		EnableMultipath:  export.EnableMultipath,
		MultipathPolicy:  export.MultipathPolicy,
		CreatedAt:        export.CreatedAt,
		UpdatedAt:        export.UpdatedAt,
	}

	// Convert client access
	for _, access := range export.ClientAccess {
		response.ClientAccess = append(response.ClientAccess, dto.NFSClientAccessResponse{
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
		})
	}

	// Convert snapshots
	for _, snapshot := range export.Snapshots {
		response.Snapshots = append(response.Snapshots, dto.NFSSnapshotResponse{
			ID:             snapshot.ID,
			ExportID:       snapshot.ExportID,
			SnapshotName:   snapshot.SnapshotName,
			SnapshotPath:   snapshot.SnapshotPath,
			FileSize:       snapshot.FileSize,
			CreatedBy:      snapshot.CreatedBy,
			Comment:        snapshot.Comment,
			IsAutoSnapshot: snapshot.IsAutoSnapshot,
			ExpiresAt:      snapshot.ExpiresAt,
			CreatedAt:      snapshot.CreatedAt,
			UpdatedAt:      snapshot.UpdatedAt,
		})
	}

	// Convert multipath configurations
	for _, mp := range export.MultipathConf {
		response.MultipathConf = append(response.MultipathConf, dto.NFSMultipathConfResponse{
			ID:            mp.ID,
			ExportID:      mp.ExportID,
			PathName:      mp.PathName,
			NetworkPath:   mp.NetworkPath,
			Priority:      mp.Priority,
			IsActive:      mp.IsActive,
			Weight:        mp.Weight,
			HealthStatus:  mp.HealthStatus,
			LastCheckTime: mp.LastCheckTime,
			ResponseTime:  mp.ResponseTime,
			CreatedAt:     mp.CreatedAt,
			UpdatedAt:     mp.UpdatedAt,
		})
	}

	return response
}

// Helper functions to provide defaults
func (s *NFSExportService) getVersionOrDefault(version models.NFSVersion) models.NFSVersion {
	if version == "" {
		return models.NFSVersion4
	}
	return version
}

func (s *NFSExportService) getSecurityFlavorOrDefault(flavor models.NFSSecurityFlavor) models.NFSSecurityFlavor {
	if flavor == "" {
		return models.SecSys
	}
	return flavor
}

func (s *NFSExportService) getPermissionOrDefault(permission models.NFSPermission) models.NFSPermission {
	if permission == "" {
		return models.PermissionReadWrite
	}
	return permission
}

func (s *NFSExportService) getOperationModeOrDefault(mode models.NFSOperationMode) models.NFSOperationMode {
	if mode == "" {
		return models.ModeSync
	}
	return mode
}

func (s *NFSExportService) getAnonUIDOrDefault(uid int) int {
	if uid == 0 {
		return 65534
	}
	return uid
}

func (s *NFSExportService) getAnonGIDOrDefault(gid int) int {
	if gid == 0 {
		return 65534
	}
	return gid
}

func (s *NFSExportService) getReadSizeOrDefault(size int) int {
	if size == 0 {
		return 131072
	}
	return size
}

func (s *NFSExportService) getWriteSizeOrDefault(size int) int {
	if size == 0 {
		return 131072
	}
	return size
}

func (s *NFSExportService) getMaxVersionsOrDefault(versions int) int {
	if versions == 0 {
		return 10
	}
	return versions
}

func (s *NFSExportService) getVersionRetentionOrDefault(retention int) int {
	if retention == 0 {
		return 30
	}
	return retention
}

func (s *NFSExportService) getMultipathPolicyOrDefault(policy string) string {
	if policy == "" {
		return "round_robin"
	}
	return policy
}