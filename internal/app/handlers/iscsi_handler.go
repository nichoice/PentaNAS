package handlers

import (
	"net/http"
	"strconv"

	"pnas/internal/app/dto"
	"pnas/internal/app/services"

	"github.com/gin-gonic/gin"
)

type iSCSIHandler struct {
	configService  *services.iSCSIConfigService
	targetService  *services.iSCSITargetService
	lunService     *services.iSCSILUNService
	aclService     *services.iSCSIACLService
}

func NewiSCSIHandler() *iSCSIHandler {
	return &iSCSIHandler{
		configService:  services.NewiSCSIConfigService(),
		targetService:  services.NewiSCSITargetService(),
		lunService:     services.NewiSCSILUNService(),
		aclService:     services.NewiSCSIACLService(),
	}
}

// Global Configuration

// @Summary Get iSCSI global configuration
// @Description Get current iSCSI global configuration
// @Tags iSCSI
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} dto.iSCSIGlobalConfigResponse
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/iscsi/config [get]
func (h *iSCSIHandler) GetGlobalConfig(c *gin.Context) {
	config, err := h.configService.GetGlobalConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, config)
}

// @Summary Update iSCSI global configuration
// @Description Update iSCSI global configuration
// @Tags iSCSI
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param config body dto.UpdateiSCSIGlobalConfigRequest true "Configuration update request"
// @Success 200 {object} dto.iSCSIGlobalConfigResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/iscsi/config [put]
func (h *iSCSIHandler) UpdateGlobalConfig(c *gin.Context) {
	var req dto.UpdateiSCSIGlobalConfigRequest
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

// Targets

// @Summary Create iSCSI target
// @Description Create a new iSCSI target
// @Tags iSCSI
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param target body dto.CreateiSCSITargetRequest true "Target creation request"
// @Success 201 {object} dto.iSCSITargetResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/iscsi/targets [post]
func (h *iSCSIHandler) CreateTarget(c *gin.Context) {
	var req dto.CreateiSCSITargetRequest
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

// @Summary Get iSCSI target
// @Description Get iSCSI target by ID
// @Tags iSCSI
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Target ID"
// @Success 200 {object} dto.iSCSITargetResponse
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/iscsi/targets/{id} [get]
func (h *iSCSIHandler) GetTarget(c *gin.Context) {
	id := c.Param("id")

	target, err := h.targetService.GetTarget(id)
	if err != nil {
		if err.Error() == "target not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, target)
}

// @Summary List iSCSI targets
// @Description List iSCSI targets with pagination and filters
// @Tags iSCSI
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(10)
// @Param status query string false "Target status"
// @Param is_enabled query bool false "Target enabled status"
// @Param search query string false "Search in name, alias, or comment"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/iscsi/targets [get]
func (h *iSCSIHandler) ListTargets(c *gin.Context) {
	var query dto.iSCSITargetListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	targets, total, err := h.targetService.ListTargets(&query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  targets,
		"total": total,
		"page":  query.Page,
		"page_size": query.PageSize,
	})
}

// @Summary Update iSCSI target
// @Description Update iSCSI target
// @Tags iSCSI
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Target ID"
// @Param target body dto.UpdateiSCSITargetRequest true "Target update request"
// @Success 200 {object} dto.iSCSITargetResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/iscsi/targets/{id} [put]
func (h *iSCSIHandler) UpdateTarget(c *gin.Context) {
	id := c.Param("id")

	var req dto.UpdateiSCSITargetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	target, err := h.targetService.UpdateTarget(id, &req)
	if err != nil {
		if err.Error() == "target not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, target)
}

// @Summary Delete iSCSI target
// @Description Delete iSCSI target
// @Tags iSCSI
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Target ID"
// @Success 204
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/iscsi/targets/{id} [delete]
func (h *iSCSIHandler) DeleteTarget(c *gin.Context) {
	id := c.Param("id")

	if err := h.targetService.DeleteTarget(id); err != nil {
		if err.Error() == "target not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// @Summary Batch operations on targets
// @Description Perform batch operations (enable/disable/delete) on targets
// @Tags iSCSI
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body dto.BatchiSCSITargetRequest true "Batch operation request"
// @Success 200 {object} dto.BatchOperationResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/iscsi/targets/batch [post]
func (h *iSCSIHandler) BatchOperationTargets(c *gin.Context) {
	var req dto.BatchiSCSITargetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.targetService.BatchOperationTargets(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// @Summary Get target statistics
// @Description Get target statistics
// @Tags iSCSI
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/iscsi/targets/stats [get]
func (h *iSCSIHandler) GetTargetStats(c *gin.Context) {
	stats, err := h.targetService.GetTargetStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// LUNs

// @Summary Create iSCSI LUN
// @Description Create a new iSCSI LUN
// @Tags iSCSI
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param lun body dto.CreateiSCSILUNRequest true "LUN creation request"
// @Success 201 {object} dto.iSCSILUNResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/iscsi/luns [post]
func (h *iSCSIHandler) CreateLUN(c *gin.Context) {
	var req dto.CreateiSCSILUNRequest
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

// @Summary Get iSCSI LUN
// @Description Get iSCSI LUN by ID
// @Tags iSCSI
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "LUN ID"
// @Success 200 {object} dto.iSCSILUNResponse
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/iscsi/luns/{id} [get]
func (h *iSCSIHandler) GetLUN(c *gin.Context) {
	id := c.Param("id")

	lun, err := h.lunService.GetLUN(id)
	if err != nil {
		if err.Error() == "LUN not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, lun)
}

// @Summary List iSCSI LUNs
// @Description List iSCSI LUNs with pagination and filters
// @Tags iSCSI
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(10)
// @Param target_id query string false "Target ID"
// @Param device_type query string false "Device type"
// @Param is_enabled query bool false "LUN enabled status"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/iscsi/luns [get]
func (h *iSCSIHandler) ListLUNs(c *gin.Context) {
	var query dto.iSCSILUNListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	luns, total, err := h.lunService.ListLUNs(&query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  luns,
		"total": total,
		"page":  query.Page,
		"page_size": query.PageSize,
	})
}

// @Summary Update iSCSI LUN
// @Description Update iSCSI LUN
// @Tags iSCSI
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "LUN ID"
// @Param lun body dto.UpdateiSCSILUNRequest true "LUN update request"
// @Success 200 {object} dto.iSCSILUNResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/iscsi/luns/{id} [put]
func (h *iSCSIHandler) UpdateLUN(c *gin.Context) {
	id := c.Param("id")

	var req dto.UpdateiSCSILUNRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	lun, err := h.lunService.UpdateLUN(id, &req)
	if err != nil {
		if err.Error() == "LUN not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, lun)
}

// @Summary Delete iSCSI LUN
// @Description Delete iSCSI LUN
// @Tags iSCSI
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "LUN ID"
// @Success 204
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/iscsi/luns/{id} [delete]
func (h *iSCSIHandler) DeleteLUN(c *gin.Context) {
	id := c.Param("id")

	if err := h.lunService.DeleteLUN(id); err != nil {
		if err.Error() == "LUN not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// @Summary Batch operations on LUNs
// @Description Perform batch operations (enable/disable/delete) on LUNs
// @Tags iSCSI
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body dto.BatchiSCSILUNRequest true "Batch operation request"
// @Success 200 {object} dto.BatchOperationResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/iscsi/luns/batch [post]
func (h *iSCSIHandler) BatchOperationLUNs(c *gin.Context) {
	var req dto.BatchiSCSILUNRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.lunService.BatchOperationLUNs(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// ACLs

// @Summary Create iSCSI ACL
// @Description Create a new iSCSI ACL
// @Tags iSCSI
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param acl body dto.CreateiSCSIACLRequest true "ACL creation request"
// @Success 201 {object} dto.iSCSIACLResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/iscsi/acls [post]
func (h *iSCSIHandler) CreateACL(c *gin.Context) {
	var req dto.CreateiSCSIACLRequest
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

// @Summary Get iSCSI ACL
// @Description Get iSCSI ACL by ID
// @Tags iSCSI
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "ACL ID"
// @Success 200 {object} dto.iSCSIACLResponse
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/iscsi/acls/{id} [get]
func (h *iSCSIHandler) GetACL(c *gin.Context) {
	id := c.Param("id")

	acl, err := h.aclService.GetACL(id)
	if err != nil {
		if err.Error() == "ACL not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, acl)
}

// @Summary List iSCSI ACLs
// @Description List iSCSI ACLs with pagination and filters
// @Tags iSCSI
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(10)
// @Param target_id query string false "Target ID"
// @Param permission query string false "Permission"
// @Param auth_type query string false "Authentication type"
// @Param is_enabled query bool false "ACL enabled status"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/iscsi/acls [get]
func (h *iSCSIHandler) ListACLs(c *gin.Context) {
	var query dto.iSCSIACLListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	acls, total, err := h.aclService.ListACLs(&query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  acls,
		"total": total,
		"page":  query.Page,
		"page_size": query.PageSize,
	})
}

// @Summary Update iSCSI ACL
// @Description Update iSCSI ACL
// @Tags iSCSI
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "ACL ID"
// @Param acl body dto.UpdateiSCSIACLRequest true "ACL update request"
// @Success 200 {object} dto.iSCSIACLResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/iscsi/acls/{id} [put]
func (h *iSCSIHandler) UpdateACL(c *gin.Context) {
	id := c.Param("id")

	var req dto.UpdateiSCSIACLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	acl, err := h.aclService.UpdateACL(id, &req)
	if err != nil {
		if err.Error() == "ACL not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, acl)
}

// @Summary Delete iSCSI ACL
// @Description Delete iSCSI ACL
// @Tags iSCSI
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "ACL ID"
// @Success 204
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/iscsi/acls/{id} [delete]
func (h *iSCSIHandler) DeleteACL(c *gin.Context) {
	id := c.Param("id")

	if err := h.aclService.DeleteACL(id); err != nil {
		if err.Error() == "ACL not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// @Summary Batch operations on ACLs
// @Description Perform batch operations (enable/disable/delete) on ACLs
// @Tags iSCSI
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body dto.BatchiSCSIACLRequest true "Batch operation request"
// @Success 200 {object} dto.BatchOperationResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/iscsi/acls/batch [post]
func (h *iSCSIHandler) BatchOperationACLs(c *gin.Context) {
	var req dto.BatchiSCSIACLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.aclService.BatchOperationACLs(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// LUN Mappings

// @Summary Create LUN mapping
// @Description Create a LUN mapping for an ACL
// @Tags iSCSI
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param mapping body dto.CreateiSCSILUNMappingRequest true "LUN mapping creation request"
// @Success 201 {object} dto.iSCSILUNMappingResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/iscsi/lun-mappings [post]
func (h *iSCSIHandler) CreateLUNMapping(c *gin.Context) {
	var req dto.CreateiSCSILUNMappingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	mapping, err := h.aclService.CreateLUNMapping(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, mapping)
}

// @Summary Update LUN mapping
// @Description Update a LUN mapping
// @Tags iSCSI
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Mapping ID"
// @Param mapping body dto.UpdateiSCSILUNMappingRequest true "LUN mapping update request"
// @Success 200 {object} dto.iSCSILUNMappingResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/iscsi/lun-mappings/{id} [put]
func (h *iSCSIHandler) UpdateLUNMapping(c *gin.Context) {
	id := c.Param("id")

	var req dto.UpdateiSCSILUNMappingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	mapping, err := h.aclService.UpdateLUNMapping(id, &req)
	if err != nil {
		if err.Error() == "LUN mapping not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, mapping)
}

// @Summary Delete LUN mapping
// @Description Delete a LUN mapping
// @Tags iSCSI
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Mapping ID"
// @Success 204
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/iscsi/lun-mappings/{id} [delete]
func (h *iSCSIHandler) DeleteLUNMapping(c *gin.Context) {
	id := c.Param("id")

	if err := h.aclService.DeleteLUNMapping(id); err != nil {
		if err.Error() == "LUN mapping not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}