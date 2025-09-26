package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"sync"
	"time"

	"pnas/internal/app/dto"
	"pnas/internal/database"
	"pnas/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// NFSOperationService handles asynchronous and synchronous NFS operations
type NFSOperationService struct {
	db         *gorm.DB
	operations map[string]*OperationStatus
	mutex      sync.RWMutex
}

// OperationStatus tracks the status of an async operation
type OperationStatus struct {
	ID          string      `json:"id"`
	Type        string      `json:"type"`
	Status      string      `json:"status"`
	Progress    int         `json:"progress"`
	Message     string      `json:"message"`
	Result      interface{} `json:"result,omitempty"`
	StartTime   time.Time   `json:"start_time"`
	EndTime     *time.Time  `json:"end_time,omitempty"`
	Priority    int         `json:"priority"`
	Parameters  interface{} `json:"parameters,omitempty"`
}

func NewNFSOperationService() *NFSOperationService {
	return &NFSOperationService{
		db:         database.DB,
		operations: make(map[string]*OperationStatus),
	}
}

// ExecuteAsyncOperation starts an asynchronous operation
func (s *NFSOperationService) ExecuteAsyncOperation(req *dto.NFSOperationRequest) (*dto.NFSOperationResponse, error) {
	operationID := uuid.New().String()

	operation := &OperationStatus{
		ID:         operationID,
		Type:       req.OperationType,
		Status:     "pending",
		Progress:   0,
		Message:    "Operation queued",
		StartTime:  time.Now(),
		Priority:   req.Priority,
		Parameters: req.Parameters,
	}

	s.mutex.Lock()
	s.operations[operationID] = operation
	s.mutex.Unlock()

	// Start operation in goroutine
	go s.executeOperation(operationID, req)

	return s.convertOperationToResponse(operation), nil
}

// ExecuteSyncOperation executes a synchronous operation
func (s *NFSOperationService) ExecuteSyncOperation(req *dto.NFSOperationRequest) (*dto.NFSOperationResponse, error) {
	operationID := uuid.New().String()

	operation := &OperationStatus{
		ID:         operationID,
		Type:       req.OperationType,
		Status:     "running",
		Progress:   0,
		Message:    "Operation starting",
		StartTime:  time.Now(),
		Priority:   req.Priority,
		Parameters: req.Parameters,
	}

	s.mutex.Lock()
	s.operations[operationID] = operation
	s.mutex.Unlock()

	// Execute operation synchronously
	result, err := s.performOperation(operation)
	now := time.Now()
	operation.EndTime = &now

	if err != nil {
		operation.Status = "failed"
		operation.Message = err.Error()
		operation.Progress = 0
	} else {
		operation.Status = "completed"
		operation.Message = "Operation completed successfully"
		operation.Progress = 100
		operation.Result = result
	}

	return s.convertOperationToResponse(operation), err
}

// GetOperationStatus gets the status of an operation
func (s *NFSOperationService) GetOperationStatus(operationID string) (*dto.NFSOperationResponse, error) {
	s.mutex.RLock()
	operation, exists := s.operations[operationID]
	s.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("operation not found: %s", operationID)
	}

	return s.convertOperationToResponse(operation), nil
}

// GetAllOperations gets all operations with optional status filter
func (s *NFSOperationService) GetAllOperations(status string) ([]dto.NFSOperationResponse, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	var responses []dto.NFSOperationResponse
	for _, operation := range s.operations {
		if status == "" || operation.Status == status {
			responses = append(responses, *s.convertOperationToResponse(operation))
		}
	}

	return responses, nil
}

// CancelOperation cancels a pending or running operation
func (s *NFSOperationService) CancelOperation(operationID string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	operation, exists := s.operations[operationID]
	if !exists {
		return fmt.Errorf("operation not found: %s", operationID)
	}

	if operation.Status == "completed" || operation.Status == "failed" {
		return fmt.Errorf("cannot cancel %s operation", operation.Status)
	}

	now := time.Now()
	operation.Status = "cancelled"
	operation.Message = "Operation cancelled by user"
	operation.EndTime = &now

	return nil
}

// executeOperation executes an operation asynchronously
func (s *NFSOperationService) executeOperation(operationID string, req *dto.NFSOperationRequest) {
	s.mutex.Lock()
	operation := s.operations[operationID]
	s.mutex.Unlock()

	operation.Status = "running"
	operation.Message = "Operation in progress"
	operation.Progress = 10

	result, err := s.performOperation(operation)
	now := time.Now()
	operation.EndTime = &now

	if err != nil {
		operation.Status = "failed"
		operation.Message = err.Error()
		operation.Progress = 0
	} else {
		operation.Status = "completed"
		operation.Message = "Operation completed successfully"
		operation.Progress = 100
		operation.Result = result
	}
}

// performOperation performs the actual operation based on type
func (s *NFSOperationService) performOperation(operation *OperationStatus) (interface{}, error) {
	switch operation.Type {
	case "export_reload":
		return s.reloadExports(operation)
	case "service_restart":
		return s.restartNFSService(operation)
	case "snapshot_create":
		return s.createSnapshot(operation)
	case "export_test":
		return s.testExport(operation)
	case "multipath_check":
		return s.checkMultipathHealth(operation)
	case "quota_sync":
		return s.syncQuotas(operation)
	default:
		return nil, fmt.Errorf("unknown operation type: %s", operation.Type)
	}
}

// reloadExports reloads NFS exports
func (s *NFSOperationService) reloadExports(operation *OperationStatus) (interface{}, error) {
	operation.Progress = 20
	operation.Message = "Validating export configuration"

	// Generate exports file first
	exportService := NewNFSExportService()
	if err := exportService.generateExportsFile(); err != nil {
		return nil, fmt.Errorf("failed to generate exports file: %w", err)
	}

	operation.Progress = 50
	operation.Message = "Reloading NFS exports"

	// Reload exports (in production, this would be: exportfs -ra)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "exportfs", "-ra")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to reload exports: %w, output: %s", err, string(output))
	}

	operation.Progress = 80
	operation.Message = "Verifying export reload"

	// Verify exports are loaded
	cmd = exec.CommandContext(ctx, "exportfs", "-v")
	output, err = cmd.CombinedOutput()
	if err != nil {
		log.Printf("Warning: failed to verify exports: %v", err)
	}

	return map[string]interface{}{
		"exports_reloaded": true,
		"exports_output":   string(output),
	}, nil
}

// restartNFSService restarts NFS services
func (s *NFSOperationService) restartNFSService(operation *OperationStatus) (interface{}, error) {
	services := []string{"nfs-server", "rpc-statd", "rpc-idmapd"}
	results := make(map[string]interface{})

	totalServices := len(services)
	for i, service := range services {
		operation.Progress = 20 + (i*60)/totalServices
		operation.Message = fmt.Sprintf("Restarting %s", service)

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

		// Stop service
		cmd := exec.CommandContext(ctx, "systemctl", "stop", service)
		if output, err := cmd.CombinedOutput(); err != nil {
			cancel()
			return nil, fmt.Errorf("failed to stop %s: %w, output: %s", service, err, string(output))
		}

		// Start service
		cmd = exec.CommandContext(ctx, "systemctl", "start", service)
		if output, err := cmd.CombinedOutput(); err != nil {
			cancel()
			return nil, fmt.Errorf("failed to start %s: %w, output: %s", service, err, string(output))
		}

		// Check status
		cmd = exec.CommandContext(ctx, "systemctl", "is-active", service)
		output, err := cmd.CombinedOutput()
		status := "unknown"
		if err == nil && string(output) == "active\n" {
			status = "running"
		}

		results[service] = map[string]interface{}{
			"status":    status,
			"restarted": true,
		}

		cancel()
	}

	operation.Progress = 90
	operation.Message = "Verifying service status"

	return results, nil
}

// createSnapshot creates a snapshot of an export
func (s *NFSOperationService) createSnapshot(operation *OperationStatus) (interface{}, error) {
	// Parse parameters
	params, ok := operation.Parameters.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid parameters for snapshot creation")
	}

	exportID, ok := params["export_id"].(string)
	if !ok {
		return nil, fmt.Errorf("export_id is required")
	}

	snapshotName, ok := params["snapshot_name"].(string)
	if !ok {
		return nil, fmt.Errorf("snapshot_name is required")
	}

	operation.Progress = 20
	operation.Message = "Finding export"

	// Get export
	var export models.NFSExport
	if err := s.db.Where("id = ?", exportID).First(&export).Error; err != nil {
		return nil, fmt.Errorf("export not found: %w", err)
	}

	operation.Progress = 40
	operation.Message = "Creating snapshot"

	// Create snapshot record
	snapshot := models.NFSSnapshot{
		Base:           models.Base{ID: uuid.New().String()},
		ExportID:       exportID,
		SnapshotName:   snapshotName,
		SnapshotPath:   fmt.Sprintf("%s/.snapshots/%s_%d", export.Path, snapshotName, time.Now().Unix()),
		CreatedBy:      "system", // Would be from authentication context
		Comment:        fmt.Sprintf("Automatic snapshot via operation %s", operation.ID),
		IsAutoSnapshot: true,
	}

	// Calculate expiration if versioning is enabled
	if export.EnableVersioning && export.VersionRetention > 0 {
		expiresAt := time.Now().AddDate(0, 0, export.VersionRetention)
		snapshot.ExpiresAt = &expiresAt
	}

	operation.Progress = 60
	operation.Message = "Saving snapshot metadata"

	if err := s.db.Create(&snapshot).Error; err != nil {
		return nil, fmt.Errorf("failed to create snapshot record: %w", err)
	}

	operation.Progress = 80
	operation.Message = "Cleaning up old snapshots"

	// Clean up old snapshots if versioning is enabled
	if export.EnableVersioning {
		s.cleanupOldSnapshots(export.ID, export.MaxVersions)
	}

	return map[string]interface{}{
		"snapshot_id":   snapshot.ID,
		"snapshot_name": snapshot.SnapshotName,
		"snapshot_path": snapshot.SnapshotPath,
		"export_id":     exportID,
	}, nil
}

// testExport tests connectivity to an export
func (s *NFSOperationService) testExport(operation *OperationStatus) (interface{}, error) {
	params, ok := operation.Parameters.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid parameters for export test")
	}

	exportID, ok := params["export_id"].(string)
	if !ok {
		return nil, fmt.Errorf("export_id is required")
	}

	operation.Progress = 20
	operation.Message = "Finding export"

	var export models.NFSExport
	if err := s.db.Where("id = ?", exportID).First(&export).Error; err != nil {
		return nil, fmt.Errorf("export not found: %w", err)
	}

	operation.Progress = 50
	operation.Message = "Testing export accessibility"

	// Test if export is accessible
	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "showmount", "-e", "localhost")
	output, err := cmd.CombinedOutput()
	responseTime := time.Since(start).Milliseconds()

	accessible := err == nil
	message := "Export is accessible"
	if !accessible {
		message = fmt.Sprintf("Export is not accessible: %v", err)
	}

	operation.Progress = 100

	return map[string]interface{}{
		"export_id":     exportID,
		"export_name":   export.Name,
		"export_path":   export.Path,
		"accessible":    accessible,
		"response_time": responseTime,
		"message":       message,
		"output":        string(output),
	}, nil
}

// checkMultipathHealth checks multipath configuration health
func (s *NFSOperationService) checkMultipathHealth(operation *OperationStatus) (interface{}, error) {
	operation.Progress = 20
	operation.Message = "Checking multipath configurations"

	var multipathConfs []models.NFSMultipathConf
	if err := s.db.Preload("Export").Find(&multipathConfs).Error; err != nil {
		return nil, fmt.Errorf("failed to get multipath configurations: %w", err)
	}

	results := make([]map[string]interface{}, 0, len(multipathConfs))
	totalConfs := len(multipathConfs)

	for i, conf := range multipathConfs {
		operation.Progress = 20 + (i*70)/totalConfs
		operation.Message = fmt.Sprintf("Checking path %s", conf.PathName)

		start := time.Now()
		healthy := s.testNetworkPath(conf.NetworkPath)
		responseTime := time.Since(start).Seconds() * 1000 // Convert to milliseconds

		status := "healthy"
		if !healthy {
			status = "failed"
		}

		// Update database
		conf.HealthStatus = status
		conf.LastCheckTime = time.Now()
		conf.ResponseTime = responseTime
		s.db.Save(&conf)

		results = append(results, map[string]interface{}{
			"path_id":       conf.ID,
			"path_name":     conf.PathName,
			"network_path":  conf.NetworkPath,
			"export_name":   conf.Export.Name,
			"status":        status,
			"response_time": responseTime,
			"is_active":     conf.IsActive,
		})
	}

	operation.Progress = 100

	return map[string]interface{}{
		"total_paths":    totalConfs,
		"healthy_paths":  s.countHealthyPaths(results),
		"failed_paths":   s.countFailedPaths(results),
		"path_details":   results,
	}, nil
}

// syncQuotas synchronizes NFS quotas
func (s *NFSOperationService) syncQuotas(operation *OperationStatus) (interface{}, error) {
	operation.Progress = 20
	operation.Message = "Fetching quota configurations"

	var quotas []models.NFSQuota
	if err := s.db.Preload("Export").Where("is_enabled = ?", true).Find(&quotas).Error; err != nil {
		return nil, fmt.Errorf("failed to get quotas: %w", err)
	}

	results := make([]map[string]interface{}, 0, len(quotas))
	totalQuotas := len(quotas)

	for i, quota := range quotas {
		operation.Progress = 20 + (i*70)/totalQuotas
		operation.Message = fmt.Sprintf("Syncing quota for %s", quota.TargetID)

		// Simulate quota sync (in production, this would interact with actual quota system)
		syncResult := map[string]interface{}{
			"quota_id":       quota.ID,
			"export_name":    quota.Export.Name,
			"quota_type":     quota.QuotaType,
			"target_id":      quota.TargetID,
			"current_size":   quota.CurrentSize,
			"current_files":  quota.CurrentFiles,
			"hard_limit":     quota.HardLimitSize,
			"soft_limit":     quota.SoftLimitSize,
			"synced":         true,
		}

		results = append(results, syncResult)
	}

	operation.Progress = 100

	return map[string]interface{}{
		"total_quotas": totalQuotas,
		"synced":       len(results),
		"quota_details": results,
	}, nil
}

// Helper functions

func (s *NFSOperationService) cleanupOldSnapshots(exportID string, maxVersions int) {
	var snapshots []models.NFSSnapshot
	s.db.Where("export_id = ? AND is_auto_snapshot = ?", exportID, true).
		Order("created_at DESC").Find(&snapshots)

	if len(snapshots) > maxVersions {
		for _, snapshot := range snapshots[maxVersions:] {
			s.db.Delete(&snapshot)
		}
	}
}

func (s *NFSOperationService) testNetworkPath(networkPath string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "ping", "-c", "1", "-W", "3", networkPath)
	return cmd.Run() == nil
}

func (s *NFSOperationService) countHealthyPaths(results []map[string]interface{}) int {
	count := 0
	for _, result := range results {
		if status, ok := result["status"].(string); ok && status == "healthy" {
			count++
		}
	}
	return count
}

func (s *NFSOperationService) countFailedPaths(results []map[string]interface{}) int {
	count := 0
	for _, result := range results {
		if status, ok := result["status"].(string); ok && status == "failed" {
			count++
		}
	}
	return count
}

func (s *NFSOperationService) convertOperationToResponse(operation *OperationStatus) *dto.NFSOperationResponse {
	response := &dto.NFSOperationResponse{
		OperationID: operation.ID,
		Status:      operation.Status,
		Progress:    operation.Progress,
		Message:     operation.Message,
		StartTime:   operation.StartTime,
		EndTime:     operation.EndTime,
	}

	if operation.Result != nil {
		// Convert result to JSON for consistent API response
		if resultBytes, err := json.Marshal(operation.Result); err == nil {
			var resultMap interface{}
			if json.Unmarshal(resultBytes, &resultMap) == nil {
				response.Result = resultMap
			}
		}
	}

	return response
}