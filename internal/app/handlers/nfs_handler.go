package handlers

import (
	"net/http"
	"strconv"

	"pnas/internal/app/dto"
	"pnas/internal/app/services"
	"pnas/internal/models"

	"github.com/gin-gonic/gin"
)

type NFSHandler struct {
	exportService    *services.NFSExportService
	configService    *services.NFSConfigService
	operationService *services.NFSOperationService
	multipathService *services.NFSMultipathService
}

func NewNFSHandler() *NFSHandler {
	return &NFSHandler{
		exportService:    services.NewNFSExportService(),
		configService:    services.NewNFSConfigService(),
		operationService: services.NewNFSOperationService(),
		multipathService: services.NewNFSMultipathService(),
	}
}

// Export management endpoints

// @Summary Create NFS export
// @Description Create a new NFS export with specified configuration
// @Tags NFS
// @Accept json
// @Produce json
// @Param export body dto.CreateNFSExportRequest true "Export configuration"
// @Success 201 {object} dto.NFSExportResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /nfs/exports [post]
func (h *NFSHandler) CreateExport(c *gin.Context) {
	var req dto.CreateNFSExportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	export, err := h.exportService.CreateExport(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, export)
}

// @Summary Get NFS exports
// @Description Get all NFS exports, optionally filtered by enabled status
// @Tags NFS
// @Produce json
// @Param enabled query bool false "Filter by enabled status"
// @Success 200 {array} dto.NFSExportResponse
// @Failure 500 {object} map[string]string
// @Router /nfs/exports [get]
func (h *NFSHandler) GetExports(c *gin.Context) {
	var enabled *bool
	if enabledStr := c.Query("enabled"); enabledStr != "" {
		if enabledBool, err := strconv.ParseBool(enabledStr); err == nil {
			enabled = &enabledBool
		}
	}

	exports, err := h.exportService.GetExports(enabled)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, exports)
}

// @Summary Get NFS export
// @Description Get a specific NFS export by ID
// @Tags NFS
// @Produce json
// @Param id path string true "Export ID"
// @Success 200 {object} dto.NFSExportResponse
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /nfs/exports/{id} [get]
func (h *NFSHandler) GetExport(c *gin.Context) {
	id := c.Param("id")

	export, err := h.exportService.GetExport(id)
	if err != nil {
		if err.Error() == "export not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, export)
}

// @Summary Update NFS export
// @Description Update an existing NFS export
// @Tags NFS
// @Accept json
// @Produce json
// @Param id path string true "Export ID"
// @Param export body dto.UpdateNFSExportRequest true "Updated export configuration"
// @Success 200 {object} dto.NFSExportResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /nfs/exports/{id} [put]
func (h *NFSHandler) UpdateExport(c *gin.Context) {
	id := c.Param("id")

	var req dto.UpdateNFSExportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	export, err := h.exportService.UpdateExport(id, &req)
	if err != nil {
		if err.Error() == "export not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, export)
}

// @Summary Delete NFS export
// @Description Delete an NFS export
// @Tags NFS
// @Param id path string true "Export ID"
// @Success 204
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /nfs/exports/{id} [delete]
func (h *NFSHandler) DeleteExport(c *gin.Context) {
	id := c.Param("id")

	if err := h.exportService.DeleteExport(id); err != nil {
		if err.Error() == "export not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.Status(http.StatusNoContent)
}

// Configuration management endpoints

// @Summary Get NFS global configuration
// @Description Get the current NFS global configuration
// @Tags NFS Configuration
// @Produce json
// @Success 200 {object} dto.NFSGlobalConfigResponse
// @Failure 500 {object} map[string]string
// @Router /nfs/config [get]
func (h *NFSHandler) GetGlobalConfig(c *gin.Context) {
	config, err := h.configService.GetGlobalConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, config)
}

// @Summary Update NFS global configuration
// @Description Update the NFS global configuration
// @Tags NFS Configuration
// @Accept json
// @Produce json
// @Param config body dto.UpdateNFSGlobalConfigRequest true "Updated configuration"
// @Success 200 {object} dto.NFSGlobalConfigResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /nfs/config [put]
func (h *NFSHandler) UpdateGlobalConfig(c *gin.Context) {
	var req dto.UpdateNFSGlobalConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	config, err := h.configService.UpdateGlobalConfig(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, config)
}

// @Summary Reset NFS global configuration
// @Description Reset the NFS global configuration to defaults
// @Tags NFS Configuration
// @Produce json
// @Success 200 {object} dto.NFSGlobalConfigResponse
// @Failure 500 {object} map[string]string
// @Router /nfs/config/reset [post]
func (h *NFSHandler) ResetGlobalConfig(c *gin.Context) {
	config, err := h.configService.ResetGlobalConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, config)
}

// Client access control endpoints

// @Summary Create client access control
// @Description Create a new client access control for an NFS export
// @Tags NFS Access Control
// @Accept json
// @Produce json
// @Param access body dto.CreateNFSClientAccessRequest true "Client access configuration"
// @Success 201 {object} dto.NFSClientAccessResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /nfs/client-access [post]
func (h *NFSHandler) CreateClientAccess(c *gin.Context) {
	var req dto.CreateNFSClientAccessRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	access, err := h.configService.CreateClientAccess(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, access)
}

// @Summary Get client access controls
// @Description Get client access controls for an NFS export
// @Tags NFS Access Control
// @Produce json
// @Param export_id path string true "Export ID"
// @Success 200 {array} dto.NFSClientAccessResponse
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /nfs/exports/{export_id}/client-access [get]
func (h *NFSHandler) GetClientAccess(c *gin.Context) {
	exportID := c.Param("export_id")

	access, err := h.configService.GetClientAccess(exportID)
	if err != nil {
		if err.Error() == "export not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, access)
}

// @Summary Update client access control
// @Description Update a client access control
// @Tags NFS Access Control
// @Accept json
// @Produce json
// @Param id path string true "Access control ID"
// @Param access body map[string]interface{} true "Updated access configuration"
// @Success 200 {object} dto.NFSClientAccessResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /nfs/client-access/{id} [put]
func (h *NFSHandler) UpdateClientAccess(c *gin.Context) {
	id := c.Param("id")

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var permission *models.NFSPermission
	var securityFlavor *models.NFSSecurityFlavor
	var rootSquash *bool
	var allSquash *bool

	if p, ok := req["permission"].(string); ok {
		perm := models.NFSPermission(p)
		permission = &perm
	}
	if sf, ok := req["security_flavor"].(string); ok {
		flavor := models.NFSSecurityFlavor(sf)
		securityFlavor = &flavor
	}
	if rs, ok := req["root_squash"].(bool); ok {
		rootSquash = &rs
	}
	if as, ok := req["all_squash"].(bool); ok {
		allSquash = &as
	}

	access, err := h.configService.UpdateClientAccess(id, permission, securityFlavor, rootSquash, allSquash)
	if err != nil {
		if err.Error() == "client access not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, access)
}

// @Summary Delete client access control
// @Description Delete a client access control
// @Tags NFS Access Control
// @Param id path string true "Access control ID"
// @Success 204
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /nfs/client-access/{id} [delete]
func (h *NFSHandler) DeleteClientAccess(c *gin.Context) {
	id := c.Param("id")

	if err := h.configService.DeleteClientAccess(id); err != nil {
		if err.Error() == "client access not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.Status(http.StatusNoContent)
}

// Operation management endpoints

// @Summary Execute async NFS operation
// @Description Execute an asynchronous NFS operation
// @Tags NFS Operations
// @Accept json
// @Produce json
// @Param operation body dto.NFSOperationRequest true "Operation configuration"
// @Success 202 {object} dto.NFSOperationResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /nfs/operations/async [post]
func (h *NFSHandler) ExecuteAsyncOperation(c *gin.Context) {
	var req dto.NFSOperationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	operation, err := h.operationService.ExecuteAsyncOperation(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusAccepted, operation)
}

// @Summary Execute sync NFS operation
// @Description Execute a synchronous NFS operation
// @Tags NFS Operations
// @Accept json
// @Produce json
// @Param operation body dto.NFSOperationRequest true "Operation configuration"
// @Success 200 {object} dto.NFSOperationResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /nfs/operations/sync [post]
func (h *NFSHandler) ExecuteSyncOperation(c *gin.Context) {
	var req dto.NFSOperationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	operation, err := h.operationService.ExecuteSyncOperation(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, operation)
}

// @Summary Get operation status
// @Description Get the status of an NFS operation
// @Tags NFS Operations
// @Produce json
// @Param operation_id path string true "Operation ID"
// @Success 200 {object} dto.NFSOperationResponse
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /nfs/operations/{operation_id} [get]
func (h *NFSHandler) GetOperationStatus(c *gin.Context) {
	operationID := c.Param("operation_id")

	operation, err := h.operationService.GetOperationStatus(operationID)
	if err != nil {
		if err.Error() == "operation not found: "+operationID {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, operation)
}

// @Summary Get all operations
// @Description Get all NFS operations with optional status filter
// @Tags NFS Operations
// @Produce json
// @Param status query string false "Filter by operation status"
// @Success 200 {array} dto.NFSOperationResponse
// @Failure 500 {object} map[string]string
// @Router /nfs/operations [get]
func (h *NFSHandler) GetAllOperations(c *gin.Context) {
	status := c.Query("status")

	operations, err := h.operationService.GetAllOperations(status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, operations)
}

// @Summary Cancel operation
// @Description Cancel a pending or running NFS operation
// @Tags NFS Operations
// @Param operation_id path string true "Operation ID"
// @Success 200 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /nfs/operations/{operation_id}/cancel [post]
func (h *NFSHandler) CancelOperation(c *gin.Context) {
	operationID := c.Param("operation_id")

	if err := h.operationService.CancelOperation(operationID); err != nil {
		if err.Error() == "operation not found: "+operationID {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else if err.Error() == "cannot cancel completed operation" || err.Error() == "cannot cancel failed operation" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Operation cancelled successfully"})
}

// Multipath management endpoints

// @Summary Create multipath configuration
// @Description Create a new multipath configuration for an NFS export
// @Tags NFS Multipath
// @Accept json
// @Produce json
// @Param multipath body dto.CreateNFSMultipathConfRequest true "Multipath configuration"
// @Success 201 {object} dto.NFSMultipathConfResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /nfs/multipath [post]
func (h *NFSHandler) CreateMultipathConf(c *gin.Context) {
	var req dto.CreateNFSMultipathConfRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	multipath, err := h.multipathService.CreateMultipathConf(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, multipath)
}

// @Summary Get multipath configurations
// @Description Get multipath configurations for an NFS export
// @Tags NFS Multipath
// @Produce json
// @Param export_id path string true "Export ID"
// @Success 200 {array} dto.NFSMultipathConfResponse
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /nfs/exports/{export_id}/multipath [get]
func (h *NFSHandler) GetMultipathConfs(c *gin.Context) {
	exportID := c.Param("export_id")

	confs, err := h.multipathService.GetMultipathConfs(exportID)
	if err != nil {
		if err.Error() == "export not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, confs)
}

// @Summary Update multipath configuration
// @Description Update a multipath configuration
// @Tags NFS Multipath
// @Accept json
// @Produce json
// @Param id path string true "Multipath configuration ID"
// @Param multipath body map[string]interface{} true "Updated multipath configuration"
// @Success 200 {object} dto.NFSMultipathConfResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /nfs/multipath/{id} [put]
func (h *NFSHandler) UpdateMultipathConf(c *gin.Context) {
	id := c.Param("id")

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var priority *int
	var weight *int
	var isActive *bool

	if p, ok := req["priority"].(float64); ok {
		pInt := int(p)
		priority = &pInt
	}
	if w, ok := req["weight"].(float64); ok {
		wInt := int(w)
		weight = &wInt
	}
	if a, ok := req["is_active"].(bool); ok {
		isActive = &a
	}

	multipath, err := h.multipathService.UpdateMultipathConf(id, priority, weight, isActive)
	if err != nil {
		if err.Error() == "multipath configuration not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, multipath)
}

// @Summary Delete multipath configuration
// @Description Delete a multipath configuration
// @Tags NFS Multipath
// @Param id path string true "Multipath configuration ID"
// @Success 204
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /nfs/multipath/{id} [delete]
func (h *NFSHandler) DeleteMultipathConf(c *gin.Context) {
	id := c.Param("id")

	if err := h.multipathService.DeleteMultipathConf(id); err != nil {
		if err.Error() == "multipath configuration not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.Status(http.StatusNoContent)
}

// @Summary Check multipath health
// @Description Check the health status of all multipath configurations
// @Tags NFS Multipath
// @Produce json
// @Success 200 {object} dto.NFSHealthCheckResponse
// @Failure 500 {object} map[string]string
// @Router /nfs/multipath/health [get]
func (h *NFSHandler) CheckMultipathHealth(c *gin.Context) {
	health, err := h.multipathService.CheckAllPathsHealth()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, health)
}

// @Summary Get path load balance
// @Description Get load balancing information for multipath configurations
// @Tags NFS Multipath
// @Produce json
// @Param export_id path string true "Export ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /nfs/exports/{export_id}/multipath/load-balance [get]
func (h *NFSHandler) GetPathLoadBalance(c *gin.Context) {
	exportID := c.Param("export_id")

	balance, err := h.multipathService.GetPathLoadBalance(exportID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, balance)
}

// @Summary Auto optimize paths
// @Description Automatically optimize multipath configuration
// @Tags NFS Multipath
// @Produce json
// @Param export_id path string true "Export ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /nfs/exports/{export_id}/multipath/optimize [post]
func (h *NFSHandler) AutoOptimizePaths(c *gin.Context) {
	exportID := c.Param("export_id")

	result, err := h.multipathService.AutoOptimizePaths(exportID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// Quota management endpoints

// @Summary Create NFS quota
// @Description Create a new NFS quota
// @Tags NFS Quotas
// @Accept json
// @Produce json
// @Param quota body dto.CreateNFSQuotaRequest true "Quota configuration"
// @Success 201 {object} dto.NFSQuotaResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /nfs/quotas [post]
func (h *NFSHandler) CreateQuota(c *gin.Context) {
	var req dto.CreateNFSQuotaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	quota, err := h.configService.CreateQuota(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, quota)
}

// @Summary Get NFS quotas
// @Description Get NFS quotas for an export
// @Tags NFS Quotas
// @Produce json
// @Param export_id path string true "Export ID"
// @Success 200 {array} dto.NFSQuotaResponse
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /nfs/exports/{export_id}/quotas [get]
func (h *NFSHandler) GetQuotas(c *gin.Context) {
	exportID := c.Param("export_id")

	quotas, err := h.configService.GetQuotas(exportID)
	if err != nil {
		if err.Error() == "export not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, quotas)
}