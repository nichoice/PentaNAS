package services

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os/exec"
	"strings"
	"time"

	"pnas/internal/app/dto"
	"pnas/internal/database"
	"pnas/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type NFSMultipathService struct {
	db *gorm.DB
}

func NewNFSMultipathService() *NFSMultipathService {
	return &NFSMultipathService{db: database.DB}
}

// CreateMultipathConf creates a new multipath configuration
func (s *NFSMultipathService) CreateMultipathConf(req *dto.CreateNFSMultipathConfRequest) (*dto.NFSMultipathConfResponse, error) {
	// Validate export exists
	var export models.NFSExport
	if err := s.db.Where("id = ?", req.ExportID).First(&export).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("export not found")
		}
		return nil, fmt.Errorf("failed to get export: %w", err)
	}

	// Check if export has multipath enabled
	if !export.EnableMultipath {
		return nil, fmt.Errorf("multipath is not enabled for this export")
	}

	// Validate network path
	if err := s.validateNetworkPath(req.NetworkPath); err != nil {
		return nil, fmt.Errorf("invalid network path: %w", err)
	}

	// Check for duplicate path name within the same export
	var existingConf models.NFSMultipathConf
	if err := s.db.Where("export_id = ? AND path_name = ?", req.ExportID, req.PathName).First(&existingConf).Error; err == nil {
		return nil, fmt.Errorf("path name '%s' already exists for this export", req.PathName)
	}

	// Create multipath configuration
	multipathConf := models.NFSMultipathConf{
		Base:          models.Base{ID: uuid.New().String()},
		ExportID:      req.ExportID,
		PathName:      req.PathName,
		NetworkPath:   req.NetworkPath,
		Priority:      s.getPriorityOrDefault(req.Priority),
		IsActive:      true,
		Weight:        s.getWeightOrDefault(req.Weight),
		HealthStatus:  "unknown",
		LastCheckTime: time.Now(),
	}

	if err := s.db.Create(&multipathConf).Error; err != nil {
		return nil, fmt.Errorf("failed to create multipath configuration: %w", err)
	}

	// Test initial connectivity
	go s.performHealthCheck(multipathConf.ID)

	return s.convertToResponse(&multipathConf), nil
}

// GetMultipathConfs gets all multipath configurations for an export
func (s *NFSMultipathService) GetMultipathConfs(exportID string) ([]dto.NFSMultipathConfResponse, error) {
	// Validate export exists
	var export models.NFSExport
	if err := s.db.Where("id = ?", exportID).First(&export).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("export not found")
		}
		return nil, fmt.Errorf("failed to get export: %w", err)
	}

	var confs []models.NFSMultipathConf
	if err := s.db.Where("export_id = ?", exportID).Order("priority ASC, path_name ASC").Find(&confs).Error; err != nil {
		return nil, fmt.Errorf("failed to get multipath configurations: %w", err)
	}

	responses := make([]dto.NFSMultipathConfResponse, len(confs))
	for i, conf := range confs {
		responses[i] = *s.convertToResponse(&conf)
	}

	return responses, nil
}

// GetMultipathConf gets a specific multipath configuration
func (s *NFSMultipathService) GetMultipathConf(id string) (*dto.NFSMultipathConfResponse, error) {
	var conf models.NFSMultipathConf
	if err := s.db.Where("id = ?", id).First(&conf).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("multipath configuration not found")
		}
		return nil, fmt.Errorf("failed to get multipath configuration: %w", err)
	}

	return s.convertToResponse(&conf), nil
}

// UpdateMultipathConf updates a multipath configuration
func (s *NFSMultipathService) UpdateMultipathConf(id string, priority *int, weight *int, isActive *bool) (*dto.NFSMultipathConfResponse, error) {
	var conf models.NFSMultipathConf
	if err := s.db.Where("id = ?", id).First(&conf).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("multipath configuration not found")
		}
		return nil, fmt.Errorf("failed to get multipath configuration: %w", err)
	}

	// Update fields if provided
	if priority != nil {
		conf.Priority = *priority
	}
	if weight != nil {
		conf.Weight = *weight
	}
	if isActive != nil {
		conf.IsActive = *isActive
	}

	if err := s.db.Save(&conf).Error; err != nil {
		return nil, fmt.Errorf("failed to update multipath configuration: %w", err)
	}

	// Trigger health check if reactivated
	if isActive != nil && *isActive {
		go s.performHealthCheck(conf.ID)
	}

	return s.convertToResponse(&conf), nil
}

// DeleteMultipathConf deletes a multipath configuration
func (s *NFSMultipathService) DeleteMultipathConf(id string) error {
	var conf models.NFSMultipathConf
	if err := s.db.Where("id = ?", id).First(&conf).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("multipath configuration not found")
		}
		return fmt.Errorf("failed to get multipath configuration: %w", err)
	}

	if err := s.db.Delete(&conf).Error; err != nil {
		return fmt.Errorf("failed to delete multipath configuration: %w", err)
	}

	return nil
}

// CheckAllPathsHealth checks the health of all multipath configurations
func (s *NFSMultipathService) CheckAllPathsHealth() (*dto.NFSHealthCheckResponse, error) {
	var confs []models.NFSMultipathConf
	if err := s.db.Preload("Export").Where("is_active = ?", true).Find(&confs).Error; err != nil {
		return nil, fmt.Errorf("failed to get multipath configurations: %w", err)
	}

	multipathHealth := make([]dto.NFSMultipathHealthResponse, 0)
	exportHealth := make(map[string]*dto.NFSExportHealthResponse)

	for _, conf := range confs {
		// Test path connectivity
		start := time.Now()
		healthy := s.testPathConnectivity(conf.NetworkPath)
		responseTime := time.Since(start).Seconds() * 1000 // Convert to milliseconds

		// Update database
		status := "healthy"
		if !healthy {
			status = "failed"
		}

		conf.HealthStatus = status
		conf.LastCheckTime = time.Now()
		conf.ResponseTime = responseTime
		s.db.Save(&conf)

		// Add to multipath health results
		multipathHealth = append(multipathHealth, dto.NFSMultipathHealthResponse{
			ExportID:     conf.ExportID,
			PathName:     conf.PathName,
			Status:       status,
			ResponseTime: responseTime,
			ActivePaths:  1,
			TotalPaths:   1,
		})

		// Aggregate export health
		if exportHealthItem, exists := exportHealth[conf.ExportID]; exists {
			exportHealthItem.TotalPaths++
			if healthy {
				exportHealthItem.ActivePaths++
			}
			if responseTime > exportHealthItem.ResponseTime {
				exportHealthItem.ResponseTime = responseTime
			}
		} else {
			activePaths := 0
			if healthy {
				activePaths = 1
			}
			exportHealth[conf.ExportID] = &dto.NFSExportHealthResponse{
				ExportID:      conf.ExportID,
				ExportName:    conf.Export.Name,
				Status:        "healthy",
				Accessibility: healthy,
				ResponseTime:  responseTime,
				ActivePaths:   activePaths,
				TotalPaths:    1,
			}
		}
	}

	// Convert export health map to slice and determine overall status
	exports := make([]dto.NFSExportHealthResponse, 0)
	overallHealthy := true
	criticalExports := 0

	for _, health := range exportHealth {
		// Determine export status based on path health
		if health.ActivePaths == 0 {
			health.Status = "critical"
			health.Message = "No active paths available"
			criticalExports++
			overallHealthy = false
		} else if health.ActivePaths < health.TotalPaths {
			health.Status = "warning"
			health.Message = fmt.Sprintf("%d of %d paths active", health.ActivePaths, health.TotalPaths)
		} else {
			health.Status = "healthy"
			health.Message = "All paths active"
		}

		exports = append(exports, *health)
	}

	// Determine overall system health
	overall := "healthy"
	if criticalExports > 0 {
		overall = "critical"
	} else if criticalExports == 0 && len(exports) > 0 && !overallHealthy {
		overall = "degraded"
	}

	// Update multipath health aggregation
	for i := range multipathHealth {
		for _, exportHealthItem := range exportHealth {
			if multipathHealth[i].ExportID == exportHealthItem.ExportID {
				multipathHealth[i].ActivePaths = exportHealthItem.ActivePaths
				multipathHealth[i].TotalPaths = exportHealthItem.TotalPaths
				break
			}
		}
	}

	return &dto.NFSHealthCheckResponse{
		Overall:         overall,
		Exports:         exports,
		MultipathHealth: multipathHealth,
		CheckTime:       time.Now(),
	}, nil
}

// GetPathLoadBalance gets load balancing information for multipath configurations
func (s *NFSMultipathService) GetPathLoadBalance(exportID string) (map[string]interface{}, error) {
	var confs []models.NFSMultipathConf
	if err := s.db.Where("export_id = ? AND is_active = ?", exportID, true).Find(&confs).Error; err != nil {
		return nil, fmt.Errorf("failed to get multipath configurations: %w", err)
	}

	if len(confs) == 0 {
		return nil, fmt.Errorf("no active multipath configurations found for export")
	}

	// Calculate load distribution based on weights
	totalWeight := 0
	healthyPaths := 0
	pathDetails := make([]map[string]interface{}, 0)

	for _, conf := range confs {
		if conf.HealthStatus == "healthy" {
			totalWeight += conf.Weight
			healthyPaths++
		}

		pathDetails = append(pathDetails, map[string]interface{}{
			"path_id":       conf.ID,
			"path_name":     conf.PathName,
			"network_path":  conf.NetworkPath,
			"priority":      conf.Priority,
			"weight":        conf.Weight,
			"is_active":     conf.IsActive,
			"health_status": conf.HealthStatus,
			"response_time": conf.ResponseTime,
		})
	}

	// Calculate load percentages
	for i, conf := range confs {
		if conf.HealthStatus == "healthy" && totalWeight > 0 {
			loadPercentage := float64(conf.Weight) / float64(totalWeight) * 100
			pathDetails[i]["load_percentage"] = loadPercentage
		} else {
			pathDetails[i]["load_percentage"] = 0.0
		}
	}

	return map[string]interface{}{
		"export_id":     exportID,
		"total_paths":   len(confs),
		"active_paths":  healthyPaths,
		"total_weight":  totalWeight,
		"path_details":  pathDetails,
	}, nil
}

// AutoOptimizePaths automatically optimizes multipath configuration
func (s *NFSMultipathService) AutoOptimizePaths(exportID string) (map[string]interface{}, error) {
	var confs []models.NFSMultipathConf
	if err := s.db.Where("export_id = ?", exportID).Find(&confs).Error; err != nil {
		return nil, fmt.Errorf("failed to get multipath configurations: %w", err)
	}

	if len(confs) == 0 {
		return nil, fmt.Errorf("no multipath configurations found for export")
	}

	optimizationResults := make([]map[string]interface{}, 0)
	totalOptimized := 0

	for _, conf := range confs {
		original := map[string]interface{}{
			"priority": conf.Priority,
			"weight":   conf.Weight,
			"active":   conf.IsActive,
		}

		optimized := false

		// Disable paths that have been unhealthy for too long
		if conf.HealthStatus == "failed" {
			timeSinceLastCheck := time.Since(conf.LastCheckTime)
			if timeSinceLastCheck > 5*time.Minute && conf.IsActive {
				conf.IsActive = false
				optimized = true
			}
		}

		// Reactivate paths that have recovered
		if conf.HealthStatus == "healthy" && !conf.IsActive {
			conf.IsActive = true
			optimized = true
		}

		// Optimize weights based on response time
		if conf.HealthStatus == "healthy" && conf.ResponseTime > 0 {
			// Lower response time = higher weight
			baseWeight := 10
			if conf.ResponseTime < 50 { // < 50ms
				conf.Weight = baseWeight + 5
			} else if conf.ResponseTime < 100 { // 50-100ms
				conf.Weight = baseWeight + 2
			} else if conf.ResponseTime < 200 { // 100-200ms
				conf.Weight = baseWeight
			} else { // > 200ms
				conf.Weight = baseWeight - 2
				if conf.Weight < 1 {
					conf.Weight = 1
				}
			}

			if conf.Weight != original["weight"] {
				optimized = true
			}
		}

		if optimized {
			s.db.Save(&conf)
			totalOptimized++
		}

		optimizationResults = append(optimizationResults, map[string]interface{}{
			"path_id":    conf.ID,
			"path_name":  conf.PathName,
			"original":   original,
			"optimized": map[string]interface{}{
				"priority": conf.Priority,
				"weight":   conf.Weight,
				"active":   conf.IsActive,
			},
			"changed": optimized,
		})
	}

	return map[string]interface{}{
		"export_id":         exportID,
		"total_paths":       len(confs),
		"paths_optimized":   totalOptimized,
		"optimization_time": time.Now(),
		"results":          optimizationResults,
	}, nil
}

// Helper functions

func (s *NFSMultipathService) validateNetworkPath(networkPath string) error {
	// Basic validation - check if it looks like a valid network path
	if networkPath == "" {
		return fmt.Errorf("network path cannot be empty")
	}

	// Check if it's a valid IP address or hostname
	parts := strings.Split(networkPath, ":")
	if len(parts) > 0 {
		if net.ParseIP(parts[0]) == nil {
			// Not an IP, check if it's a valid hostname format
			if len(parts[0]) == 0 {
				return fmt.Errorf("invalid hostname format")
			}
		}
	}

	return nil
}

func (s *NFSMultipathService) testPathConnectivity(networkPath string) bool {
	// Extract hostname/IP from network path
	host := networkPath
	if strings.Contains(networkPath, ":") {
		parts := strings.Split(networkPath, ":")
		host = parts[0]
	}

	// Test connectivity with ping
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "ping", "-c", "1", "-W", "2", host)
	return cmd.Run() == nil
}

func (s *NFSMultipathService) performHealthCheck(confID string) {
	var conf models.NFSMultipathConf
	if err := s.db.Where("id = ?", confID).First(&conf).Error; err != nil {
		return
	}

	start := time.Now()
	healthy := s.testPathConnectivity(conf.NetworkPath)
	responseTime := time.Since(start).Seconds() * 1000 // Convert to milliseconds

	status := "healthy"
	if !healthy {
		status = "failed"
	}

	conf.HealthStatus = status
	conf.LastCheckTime = time.Now()
	conf.ResponseTime = responseTime
	s.db.Save(&conf)
}

func (s *NFSMultipathService) convertToResponse(conf *models.NFSMultipathConf) *dto.NFSMultipathConfResponse {
	return &dto.NFSMultipathConfResponse{
		ID:            conf.ID,
		ExportID:      conf.ExportID,
		PathName:      conf.PathName,
		NetworkPath:   conf.NetworkPath,
		Priority:      conf.Priority,
		IsActive:      conf.IsActive,
		Weight:        conf.Weight,
		HealthStatus:  conf.HealthStatus,
		LastCheckTime: conf.LastCheckTime,
		ResponseTime:  conf.ResponseTime,
		CreatedAt:     conf.CreatedAt,
		UpdatedAt:     conf.UpdatedAt,
	}
}

func (s *NFSMultipathService) getPriorityOrDefault(priority int) int {
	if priority == 0 {
		return 1
	}
	return priority
}

func (s *NFSMultipathService) getWeightOrDefault(weight int) int {
	if weight == 0 {
		return 1
	}
	return weight
}