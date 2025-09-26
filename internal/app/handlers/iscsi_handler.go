package handlers

import (
	"net/http"
	"strconv"

	"pnas/internal/app/dto"
	"pnas/internal/app/services"

	"github.com/gin-gonic/gin"
)

type ISCSIHandler struct {
	targetService      *services.ISCSITargetService
	lunService         *services.ISCSILUNService
	aclService         *services.ISCSIACLService
	configService      *services.ISCSIConfigService
	serviceService     *services.ISCSIServiceService
	sessionService     *services.ISCSISessionService
	connectionService  *services.ISCSIConnectionService
	poolService        *services.ISCSIStoragePoolService
	performanceService *services.ISCSIPerformanceService
	auditService       *services.ISCSIAuditService
}

func NewISCSIHandler() *ISCSIHandler {
	return &ISCSIHandler{
		targetService:      services.NewISCSITargetService(),
		lunService:         services.NewISCSILUNService(),
		aclService:         services.NewISCSIACLService(),
		configService:      services.NewISCSIConfigService(),
		serviceService:     services.NewISCSIServiceService(),
		sessionService:     services.NewISCSISessionService(),
		connectionService:  services.NewISCSIConnectionService(),
		poolService:        services.NewISCSIStoragePoolService(),
		performanceService: services.NewISCSIPerformanceService(),
		auditService:       services.NewISCSIAuditService(),
	}
}

// Target management endpoints

// @Summary Create iSCSI target
// @Description Create a new iSCSI target with specified IQN
// @Tags iSCSI
// @Accept json
// @Produce json
// @Param target body dto.CreateISCSITargetRequest true "Target configuration"
// @Success 201 {object} dto.ISCSITargetResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /iscsi/targets [post]
func (h *ISCSIHandler) CreateTarget(c *gin.Context) {
	var req dto.CreateISCSITargetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	target, err := h.targetService.CreateTarget(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, target)
}

// @Summary Get all iSCSI targets
// @Description Get a list of all iSCSI targets with optional filtering
// @Tags iSCSI
// @Produce json
// @Param status query string false "Filter by target status"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} dto.ISCSITargetListResponse
// @Failure 500 {object} map[string]string
// @Router /iscsi/targets [get]
func (h *ISCSIHandler) GetTargets(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	status := c.Query("status")

	targets, total, err := h.targetService.GetTargets(page, limit, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := dto.ISCSITargetListResponse{
		Targets: targets,
		Total:   total,
		Page:    page,
		Limit:   limit,
	}

	c.JSON(http.StatusOK, response)
}

// @Summary Get iSCSI target by ID
// @Description Get detailed information about a specific iSCSI target
// @Tags iSCSI
// @Produce json
// @Param id path string true "Target ID"
// @Success 200 {object} dto.ISCSITargetResponse
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /iscsi/targets/{id} [get]
func (h *ISCSIHandler) GetTarget(c *gin.Context) {
	id := c.Param("id")

	target, err := h.targetService.GetTargetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Target not found"})
		return
	}

	c.JSON(http.StatusOK, target)
}

// @Summary Update iSCSI target
// @Description Update an existing iSCSI target configuration
// @Tags iSCSI
// @Accept json
// @Produce json
// @Param id path string true "Target ID"
// @Param target body dto.UpdateISCSITargetRequest true "Updated target configuration"
// @Success 200 {object} dto.ISCSITargetResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /iscsi/targets/{id} [put]
func (h *ISCSIHandler) UpdateTarget(c *gin.Context) {
	id := c.Param("id")

	var req dto.UpdateISCSITargetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	target, err := h.targetService.UpdateTarget(id, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, target)
}

// @Summary Delete iSCSI target
// @Description Delete an iSCSI target and all its associated resources
// @Tags iSCSI
// @Produce json
// @Param id path string true "Target ID"
// @Success 204
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /iscsi/targets/{id} [delete]
func (h *ISCSIHandler) DeleteTarget(c *gin.Context) {
	id := c.Param("id")

	err := h.targetService.DeleteTarget(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// @Summary Start iSCSI target
// @Description Start an inactive iSCSI target
// @Tags iSCSI
// @Produce json
// @Param id path string true "Target ID"
// @Success 200 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /iscsi/targets/{id}/start [post]
func (h *ISCSIHandler) StartTarget(c *gin.Context) {
	id := c.Param("id")

	err := h.targetService.StartTarget(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Target started successfully"})
}

// @Summary Stop iSCSI target
// @Description Stop an active iSCSI target
// @Tags iSCSI
// @Produce json
// @Param id path string true "Target ID"
// @Success 200 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /iscsi/targets/{id}/stop [post]
func (h *ISCSIHandler) StopTarget(c *gin.Context) {
	id := c.Param("id")

	err := h.targetService.StopTarget(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Target stopped successfully"})
}

// @Summary Get target status
// @Description Get the current operational status of an iSCSI target
// @Tags iSCSI
// @Produce json
// @Param id path string true "Target ID"
// @Success 200 {object} dto.ISCSITargetStatusResponse
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /iscsi/targets/{id}/status [get]
func (h *ISCSIHandler) GetTargetStatus(c *gin.Context) {
	id := c.Param("id")

	status, err := h.targetService.GetTargetStatus(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, status)
}

// @Summary Get target LUNs
// @Description Get all LUNs associated with a specific target
// @Tags iSCSI
// @Produce json
// @Param id path string true "Target ID"
// @Success 200 {object} dto.ISCSILUNListResponse
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /iscsi/targets/{id}/luns [get]
func (h *ISCSIHandler) GetTargetLUNs(c *gin.Context) {
	targetID := c.Param("id")

	luns, err := h.lunService.GetLUNsByTarget(targetID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := dto.ISCSILUNListResponse{
		LUNs: luns,
	}

	c.JSON(http.StatusOK, response)
}

// LUN management endpoints

// @Summary Create iSCSI LUN
// @Description Create a new LUN with specified storage configuration
// @Tags iSCSI
// @Accept json
// @Produce json
// @Param lun body dto.CreateISCSILUNRequest true "LUN configuration"
// @Success 201 {object} dto.ISCSILUNResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /iscsi/luns [post]
func (h *ISCSIHandler) CreateLUN(c *gin.Context) {
	var req dto.CreateISCSILUNRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	lun, err := h.lunService.CreateLUN(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, lun)
}

// @Summary Get all LUNs
// @Description Get a list of all LUNs with optional filtering
// @Tags iSCSI
// @Produce json
// @Param device_type query string false "Filter by device type"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} dto.ISCSILUNListResponse
// @Failure 500 {object} map[string]string
// @Router /iscsi/luns [get]
func (h *ISCSIHandler) GetLUNs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	deviceType := c.Query("device_type")

	luns, total, err := h.lunService.GetLUNs(page, limit, deviceType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := dto.ISCSILUNListResponse{
		LUNs:  luns,
		Total: total,
		Page:  page,
		Limit: limit,
	}

	c.JSON(http.StatusOK, response)
}

// @Summary Get LUN by ID
// @Description Get detailed information about a specific LUN
// @Tags iSCSI
// @Produce json
// @Param id path string true "LUN ID"
// @Success 200 {object} dto.ISCSILUNResponse
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /iscsi/luns/{id} [get]
func (h *ISCSIHandler) GetLUN(c *gin.Context) {
	id := c.Param("id")

	lun, err := h.lunService.GetLUNByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "LUN not found"})
		return
	}

	c.JSON(http.StatusOK, lun)
}

// @Summary Update LUN
// @Description Update an existing LUN configuration
// @Tags iSCSI
// @Accept json
// @Produce json
// @Param id path string true "LUN ID"
// @Param lun body dto.UpdateISCSILUNRequest true "Updated LUN configuration"
// @Success 200 {object} dto.ISCSILUNResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /iscsi/luns/{id} [put]
func (h *ISCSIHandler) UpdateLUN(c *gin.Context) {
	id := c.Param("id")

	var req dto.UpdateISCSILUNRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	lun, err := h.lunService.UpdateLUN(id, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, lun)
}

// @Summary Delete LUN
// @Description Delete a LUN and unmap it from all targets
// @Tags iSCSI
// @Produce json
// @Param id path string true "LUN ID"
// @Success 204
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /iscsi/luns/{id} [delete]
func (h *ISCSIHandler) DeleteLUN(c *gin.Context) {
	id := c.Param("id")

	err := h.lunService.DeleteLUN(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// @Summary Map LUN to target
// @Description Map a LUN to a specific target
// @Tags iSCSI
// @Accept json
// @Produce json
// @Param id path string true "LUN ID"
// @Param mapping body dto.MapLUNRequest true "Mapping configuration"
// @Success 200 {object} dto.ISCSILUNMappingResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /iscsi/luns/{id}/map [post]
func (h *ISCSIHandler) MapLUNToTarget(c *gin.Context) {
	lunID := c.Param("id")

	var req dto.MapLUNRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	mapping, err := h.lunService.MapLUNToTarget(lunID, req.TargetID, req.LUN)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, mapping)
}

// @Summary Unmap LUN from target
// @Description Remove a LUN mapping from a specific target
// @Tags iSCSI
// @Produce json
// @Param id path string true "LUN ID"
// @Param target_id path string true "Target ID"
// @Success 204
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /iscsi/luns/{id}/map/{target_id} [delete]
func (h *ISCSIHandler) UnmapLUNFromTarget(c *gin.Context) {
	lunID := c.Param("id")
	targetID := c.Param("target_id")

	err := h.lunService.UnmapLUNFromTarget(lunID, targetID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// ACL management endpoints

// @Summary Create ACL
// @Description Create a new Access Control List entry for iSCSI target access
// @Tags iSCSI
// @Accept json
// @Produce json
// @Param acl body dto.CreateISCSIACLRequest true "ACL configuration"
// @Success 201 {object} dto.ISCSIACLResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /iscsi/acls [post]
func (h *ISCSIHandler) CreateACL(c *gin.Context) {
	var req dto.CreateISCSIACLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	acl, err := h.aclService.CreateACL(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, acl)
}

// @Summary Get all ACLs
// @Description Get a list of all ACLs with optional filtering
// @Tags iSCSI
// @Produce json
// @Param target_id query string false "Filter by target ID"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} dto.ISCSIACLListResponse
// @Failure 500 {object} map[string]string
// @Router /iscsi/acls [get]
func (h *ISCSIHandler) GetACLs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	targetID := c.Query("target_id")

	acls, total, err := h.aclService.GetACLs(page, limit, targetID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := dto.ISCSIACLListResponse{
		ACLs:  acls,
		Total: total,
		Page:  page,
		Limit: limit,
	}

	c.JSON(http.StatusOK, response)
}

// @Summary Get ACL by ID
// @Description Get detailed information about a specific ACL
// @Tags iSCSI
// @Produce json
// @Param id path string true "ACL ID"
// @Success 200 {object} dto.ISCSIACLResponse
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /iscsi/acls/{id} [get]
func (h *ISCSIHandler) GetACL(c *gin.Context) {
	id := c.Param("id")

	acl, err := h.aclService.GetACLByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "ACL not found"})
		return
	}

	c.JSON(http.StatusOK, acl)
}

// @Summary Update ACL
// @Description Update an existing ACL configuration
// @Tags iSCSI
// @Accept json
// @Produce json
// @Param id path string true "ACL ID"
// @Param acl body dto.UpdateISCSIACLRequest true "Updated ACL configuration"
// @Success 200 {object} dto.ISCSIACLResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /iscsi/acls/{id} [put]
func (h *ISCSIHandler) UpdateACL(c *gin.Context) {
	id := c.Param("id")

	var req dto.UpdateISCSIACLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	acl, err := h.aclService.UpdateACL(id, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, acl)
}

// @Summary Delete ACL
// @Description Delete an ACL and revoke target access
// @Tags iSCSI
// @Produce json
// @Param id path string true "ACL ID"
// @Success 204
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /iscsi/acls/{id} [delete]
func (h *ISCSIHandler) DeleteACL(c *gin.Context) {
	id := c.Param("id")

	err := h.aclService.DeleteACL(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// Global configuration endpoints

// @Summary Get global configuration
// @Description Get the current global iSCSI configuration
// @Tags iSCSI
// @Produce json
// @Success 200 {object} dto.ISCSIGlobalConfigResponse
// @Failure 500 {object} map[string]string
// @Router /iscsi/config [get]
func (h *ISCSIHandler) GetGlobalConfig(c *gin.Context) {
	config, err := h.configService.GetGlobalConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, config)
}

// @Summary Update global configuration
// @Description Update the global iSCSI configuration
// @Tags iSCSI
// @Accept json
// @Produce json
// @Param config body dto.UpdateISCSIGlobalConfigRequest true "Updated configuration"
// @Success 200 {object} dto.ISCSIGlobalConfigResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /iscsi/config [put]
func (h *ISCSIHandler) UpdateGlobalConfig(c *gin.Context) {
	var req dto.UpdateISCSIGlobalConfigRequest
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

// @Summary Reset global configuration
// @Description Reset the global iSCSI configuration to defaults
// @Tags iSCSI
// @Produce json
// @Success 200 {object} dto.ISCSIGlobalConfigResponse
// @Failure 500 {object} map[string]string
// @Router /iscsi/config/reset [post]
func (h *ISCSIHandler) ResetGlobalConfig(c *gin.Context) {
	config, err := h.configService.ResetGlobalConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, config)
}

// Service management endpoints

// @Summary Get service status
// @Description Get the current status of the iSCSI service
// @Tags iSCSI
// @Produce json
// @Success 200 {object} dto.ISCSIServiceStatusResponse
// @Failure 500 {object} map[string]string
// @Router /iscsi/service/status [get]
func (h *ISCSIHandler) GetServiceStatus(c *gin.Context) {
	status, err := h.serviceService.GetServiceStatus()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, status)
}

// @Summary Start service
// @Description Start the iSCSI service
// @Tags iSCSI
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /iscsi/service/start [post]
func (h *ISCSIHandler) StartService(c *gin.Context) {
	err := h.serviceService.StartService()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "iSCSI service started successfully"})
}

// @Summary Stop service
// @Description Stop the iSCSI service
// @Tags iSCSI
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /iscsi/service/stop [post]
func (h *ISCSIHandler) StopService(c *gin.Context) {
	err := h.serviceService.StopService()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "iSCSI service stopped successfully"})
}

// @Summary Restart service
// @Description Restart the iSCSI service
// @Tags iSCSI
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /iscsi/service/restart [post]
func (h *ISCSIHandler) RestartService(c *gin.Context) {
	err := h.serviceService.RestartService()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "iSCSI service restarted successfully"})
}

// Session management endpoints

// @Summary Get all sessions
// @Description Get a list of all active iSCSI sessions
// @Tags iSCSI
// @Produce json
// @Param target_id query string false "Filter by target ID"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} dto.ISCSISessionListResponse
// @Failure 500 {object} map[string]string
// @Router /iscsi/sessions [get]
func (h *ISCSIHandler) GetSessions(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	targetID := c.Query("target_id")

	sessions, total, err := h.sessionService.GetSessions(page, limit, targetID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := dto.ISCSISessionListResponse{
		Sessions: sessions,
		Total:    total,
		Page:     page,
		Limit:    limit,
	}

	c.JSON(http.StatusOK, response)
}

// @Summary Get session by ID
// @Description Get detailed information about a specific session
// @Tags iSCSI
// @Produce json
// @Param id path string true "Session ID"
// @Success 200 {object} dto.ISCSISessionResponse
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /iscsi/sessions/{id} [get]
func (h *ISCSIHandler) GetSession(c *gin.Context) {
	id := c.Param("id")

	session, err := h.sessionService.GetSessionByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Session not found"})
		return
	}

	c.JSON(http.StatusOK, session)
}

// @Summary Terminate session
// @Description Forcefully terminate an active iSCSI session
// @Tags iSCSI
// @Produce json
// @Param id path string true "Session ID"
// @Success 204
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /iscsi/sessions/{id} [delete]
func (h *ISCSIHandler) TerminateSession(c *gin.Context) {
	id := c.Param("id")

	err := h.sessionService.TerminateSession(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// Connection monitoring endpoints

// @Summary Get all connections
// @Description Get a list of all active iSCSI connections
// @Tags iSCSI
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} dto.ISCSIConnectionListResponse
// @Failure 500 {object} map[string]string
// @Router /iscsi/connections [get]
func (h *ISCSIHandler) GetConnections(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	connections, total, err := h.connectionService.GetConnections(page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := dto.ISCSIConnectionListResponse{
		Connections: connections,
		Total:       total,
		Page:        page,
		Limit:       limit,
	}

	c.JSON(http.StatusOK, response)
}

// @Summary Get connection history
// @Description Get historical connection data for analysis
// @Tags iSCSI
// @Produce json
// @Param from query string false "Start date (YYYY-MM-DD)"
// @Param to query string false "End date (YYYY-MM-DD)"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} dto.ISCSIConnectionHistoryResponse
// @Failure 500 {object} map[string]string
// @Router /iscsi/connections/history [get]
func (h *ISCSIHandler) GetConnectionHistory(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	from := c.Query("from")
	to := c.Query("to")

	history, total, err := h.connectionService.GetConnectionHistory(page, limit, from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := dto.ISCSIConnectionHistoryResponse{
		History: history,
		Total:   total,
		Page:    page,
		Limit:   limit,
	}

	c.JSON(http.StatusOK, response)
}

// Storage pool management endpoints

// @Summary Create storage pool
// @Description Create a new storage pool for LUN allocation
// @Tags iSCSI
// @Accept json
// @Produce json
// @Param pool body dto.CreateISCSIStoragePoolRequest true "Pool configuration"
// @Success 201 {object} dto.ISCSIStoragePoolResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /iscsi/pools [post]
func (h *ISCSIHandler) CreateStoragePool(c *gin.Context) {
	var req dto.CreateISCSIStoragePoolRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	pool, err := h.poolService.CreateStoragePool(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, pool)
}

// @Summary Get all storage pools
// @Description Get a list of all storage pools
// @Tags iSCSI
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} dto.ISCSIStoragePoolListResponse
// @Failure 500 {object} map[string]string
// @Router /iscsi/pools [get]
func (h *ISCSIHandler) GetStoragePools(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	pools, total, err := h.poolService.GetStoragePools(page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := dto.ISCSIStoragePoolListResponse{
		Pools: pools,
		Total: total,
		Page:  page,
		Limit: limit,
	}

	c.JSON(http.StatusOK, response)
}

// @Summary Get storage pool by ID
// @Description Get detailed information about a specific storage pool
// @Tags iSCSI
// @Produce json
// @Param id path string true "Pool ID"
// @Success 200 {object} dto.ISCSIStoragePoolResponse
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /iscsi/pools/{id} [get]
func (h *ISCSIHandler) GetStoragePool(c *gin.Context) {
	id := c.Param("id")

	pool, err := h.poolService.GetStoragePoolByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Storage pool not found"})
		return
	}

	c.JSON(http.StatusOK, pool)
}

// @Summary Update storage pool
// @Description Update an existing storage pool configuration
// @Tags iSCSI
// @Accept json
// @Produce json
// @Param id path string true "Pool ID"
// @Param pool body dto.UpdateISCSIStoragePoolRequest true "Updated pool configuration"
// @Success 200 {object} dto.ISCSIStoragePoolResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /iscsi/pools/{id} [put]
func (h *ISCSIHandler) UpdateStoragePool(c *gin.Context) {
	id := c.Param("id")

	var req dto.UpdateISCSIStoragePoolRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	pool, err := h.poolService.UpdateStoragePool(id, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, pool)
}

// @Summary Delete storage pool
// @Description Delete a storage pool and all its LUNs
// @Tags iSCSI
// @Produce json
// @Param id path string true "Pool ID"
// @Success 204
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /iscsi/pools/{id} [delete]
func (h *ISCSIHandler) DeleteStoragePool(c *gin.Context) {
	id := c.Param("id")

	err := h.poolService.DeleteStoragePool(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// Performance monitoring endpoints

// @Summary Get performance statistics
// @Description Get overall performance statistics for the iSCSI service
// @Tags iSCSI
// @Produce json
// @Param from query string false "Start date (YYYY-MM-DD)"
// @Param to query string false "End date (YYYY-MM-DD)"
// @Success 200 {object} dto.ISCSIPerformanceStatsResponse
// @Failure 500 {object} map[string]string
// @Router /iscsi/monitoring/performance [get]
func (h *ISCSIHandler) GetPerformanceStats(c *gin.Context) {
	from := c.Query("from")
	to := c.Query("to")

	stats, err := h.performanceService.GetPerformanceStats(from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// @Summary Get target performance statistics
// @Description Get performance statistics for a specific target
// @Tags iSCSI
// @Produce json
// @Param target_id path string true "Target ID"
// @Param from query string false "Start date (YYYY-MM-DD)"
// @Param to query string false "End date (YYYY-MM-DD)"
// @Success 200 {object} dto.ISCSITargetPerformanceResponse
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /iscsi/monitoring/targets/{target_id}/stats [get]
func (h *ISCSIHandler) GetTargetStats(c *gin.Context) {
	targetID := c.Param("target_id")
	from := c.Query("from")
	to := c.Query("to")

	stats, err := h.performanceService.GetTargetStats(targetID, from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// @Summary Get LUN performance statistics
// @Description Get performance statistics for a specific LUN
// @Tags iSCSI
// @Produce json
// @Param lun_id path string true "LUN ID"
// @Param from query string false "Start date (YYYY-MM-DD)"
// @Param to query string false "End date (YYYY-MM-DD)"
// @Success 200 {object} dto.ISCSILUNPerformanceResponse
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /iscsi/monitoring/luns/{lun_id}/stats [get]
func (h *ISCSIHandler) GetLUNStats(c *gin.Context) {
	lunID := c.Param("lun_id")
	from := c.Query("from")
	to := c.Query("to")

	stats, err := h.performanceService.GetLUNStats(lunID, from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// Audit logging endpoints

// @Summary Get audit logs
// @Description Get audit logs for iSCSI operations
// @Tags iSCSI
// @Produce json
// @Param action query string false "Filter by action type"
// @Param user_id query string false "Filter by user ID"
// @Param from query string false "Start date (YYYY-MM-DD)"
// @Param to query string false "End date (YYYY-MM-DD)"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} dto.ISCSIAuditLogListResponse
// @Failure 500 {object} map[string]string
// @Router /iscsi/audit/logs [get]
func (h *ISCSIHandler) GetAuditLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	action := c.Query("action")
	userID := c.Query("user_id")
	from := c.Query("from")
	to := c.Query("to")

	logs, total, err := h.auditService.GetAuditLogs(page, limit, action, userID, from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := dto.ISCSIAuditLogListResponse{
		Logs:  logs,
		Total: total,
		Page:  page,
		Limit: limit,
	}

	c.JSON(http.StatusOK, response)
}

// @Summary Get audit statistics
// @Description Get statistical information about audit activities
// @Tags iSCSI
// @Produce json
// @Param from query string false "Start date (YYYY-MM-DD)"
// @Param to query string false "End date (YYYY-MM-DD)"
// @Success 200 {object} dto.ISCSIAuditStatsResponse
// @Failure 500 {object} map[string]string
// @Router /iscsi/audit/stats [get]
func (h *ISCSIHandler) GetAuditStats(c *gin.Context) {
	from := c.Query("from")
	to := c.Query("to")

	stats, err := h.auditService.GetAuditStats(from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}