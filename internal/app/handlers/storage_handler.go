package handlers

import (
	"net/http"
	"pnas/internal/app/dto"
	"pnas/internal/app/services"
	"pnas/internal/shared/errors"

	"github.com/gin-gonic/gin"
)

// StorageHandler 存储处理器
type StorageHandler struct {
	storageService *services.StorageService
}

// NewStorageHandler 创建存储处理器
func NewStorageHandler(storageService *services.StorageService) *StorageHandler {
	return &StorageHandler{
		storageService: storageService,
	}
}

// GetDisks 获取磁盘列表
// @Summary 获取磁盘列表
// @Description 获取系统中所有磁盘的信息
// @Tags 存储管理
// @Accept json
// @Produce json
// @Success 200 {object} []dto.DiskResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /api/v1/storage/disks [get]
func (h *StorageHandler) GetDisks(c *gin.Context) {
	disks, err := h.storageService.GetDisks(c.Request.Context())
	if err != nil {
		errors.HandleError(c, http.StatusInternalServerError, "Failed to get disks", err)
		return
	}

	c.JSON(http.StatusOK, disks)
}

// GetVolumeGroups 获取卷组列表
// @Summary 获取卷组列表
// @Description 获取LVM卷组信息
// @Tags 存储管理
// @Accept json
// @Produce json
// @Success 200 {object} []dto.VolumeGroupResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /api/v1/storage/vgs [get]
func (h *StorageHandler) GetVolumeGroups(c *gin.Context) {
	vgs, err := h.storageService.GetVolumeGroups(c.Request.Context())
	if err != nil {
		errors.HandleError(c, http.StatusInternalServerError, "Failed to get volume groups", err)
		return
	}

	c.JSON(http.StatusOK, vgs)
}

// GetLogicalVolumes 获取逻辑卷列表
// @Summary 获取逻辑卷列表
// @Description 获取LVM逻辑卷信息
// @Tags 存储管理
// @Accept json
// @Produce json
// @Success 200 {object} []dto.LogicalVolumeResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /api/v1/storage/lvs [get]
func (h *StorageHandler) GetLogicalVolumes(c *gin.Context) {
	lvs, err := h.storageService.GetLogicalVolumes(c.Request.Context())
	if err != nil {
		errors.HandleError(c, http.StatusInternalServerError, "Failed to get logical volumes", err)
		return
	}

	c.JSON(http.StatusOK, lvs)
}

// CreateVolumeGroup 创建卷组
// @Summary 创建卷组
// @Description 创建LVM卷组
// @Tags 存储管理
// @Accept json
// @Produce json
// @Param request body dto.CreateVGRequest true "创建卷组请求"
// @Success 200 {object} map[string]string
// @Failure 400 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /api/v1/storage/vgs [post]
func (h *StorageHandler) CreateVolumeGroup(c *gin.Context) {
	var req dto.CreateVGRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errors.HandleError(c, http.StatusBadRequest, "Invalid request", err)
		return
	}

	if err := h.storageService.CreateVolumeGroup(c.Request.Context(), req); err != nil {
		errors.HandleError(c, http.StatusInternalServerError, "Failed to create volume group", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Volume group created successfully"})
}

// CreateLogicalVolume 创建逻辑卷
// @Summary 创建逻辑卷
// @Description 创建LVM逻辑卷
// @Tags 存储管理
// @Accept json
// @Produce json
// @Param request body dto.CreateLVRequest true "创建逻辑卷请求"
// @Success 200 {object} map[string]string
// @Failure 400 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /api/v1/storage/lvs [post]
func (h *StorageHandler) CreateLogicalVolume(c *gin.Context) {
	var req dto.CreateLVRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errors.HandleError(c, http.StatusBadRequest, "Invalid request", err)
		return
	}

	if err := h.storageService.CreateLogicalVolume(c.Request.Context(), req); err != nil {
		errors.HandleError(c, http.StatusInternalServerError, "Failed to create logical volume", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Logical volume created successfully"})
}

// GetStorageOverview 获取存储概览
// @Summary 获取存储概览
// @Description 获取存储系统整体概览信息
// @Tags 存储管理
// @Accept json
// @Produce json
// @Success 200 {object} dto.StorageOverviewResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /api/v1/storage/overview [get]
func (h *StorageHandler) GetStorageOverview(c *gin.Context) {
	overview, err := h.storageService.GetStorageOverview(c.Request.Context())
	if err != nil {
		errors.HandleError(c, http.StatusInternalServerError, "Failed to get storage overview", err)
		return
	}

	c.JSON(http.StatusOK, overview)
}