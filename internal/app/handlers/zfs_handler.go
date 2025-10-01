package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"pnas/internal/app/dto"
	"pnas/internal/app/services"
)

// ZFSHandler handles ZFS HTTP requests
type ZFSHandler struct {
	service *services.ZFSService
}

// NewZFSHandler creates a new ZFS handler
func NewZFSHandler() *ZFSHandler {
	return &ZFSHandler{
		service: services.NewZFSService(),
	}
}

// Pool Handlers

// CreatePool creates a new ZFS pool
// @Summary Create ZFS pool
// @Tags ZFS Pools
// @Accept json
// @Produce json
// @Param pool body dto.CreatePoolRequest true "Pool creation request"
// @Success 201 {object} dto.PoolResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /zfs/pools [post]
func (h *ZFSHandler) CreatePool(c *gin.Context) {
	var req dto.CreatePoolRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid request",
			Details: err.Error(),
		})
		return
	}

	pool, err := h.service.CreatePool(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Failed to create pool",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, pool)
}

// ListPools lists all ZFS pools
// @Summary List ZFS pools
// @Tags ZFS Pools
// @Produce json
// @Param pool query string false "Filter by pool name"
// @Param status query string false "Filter by status"
// @Param page query int false "Page number"
// @Param page_size query int false "Page size"
// @Success 200 {object} dto.ListResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /zfs/pools [get]
func (h *ZFSHandler) ListPools(c *gin.Context) {
	var req dto.ListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid query parameters",
			Details: err.Error(),
		})
		return
	}

	resp, err := h.service.ListPools(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Failed to list pools",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetPoolStatus gets detailed pool status
// @Summary Get pool status
// @Tags ZFS Pools
// @Produce json
// @Param name path string true "Pool name"
// @Success 200 {object} dto.PoolStatusResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /zfs/pools/{name}/status [get]
func (h *ZFSHandler) GetPoolStatus(c *gin.Context) {
	name := c.Param("name")

	status, err := h.service.GetPoolStatus(name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Failed to get pool status",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, status)
}

// DestroyPool destroys a ZFS pool
// @Summary Destroy ZFS pool
// @Tags ZFS Pools
// @Produce json
// @Param name path string true "Pool name"
// @Param force query bool false "Force destroy"
// @Success 200 {object} dto.MessageResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /zfs/pools/{name} [delete]
func (h *ZFSHandler) DestroyPool(c *gin.Context) {
	name := c.Param("name")
	force := c.Query("force") == "true"

	if err := h.service.DestroyPool(name, force); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Failed to destroy pool",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.MessageResponse{
		Message: "Pool destroyed successfully",
	})
}

// ScrubPool starts data scrubbing on a pool
// @Summary Start pool scrub
// @Tags ZFS Pools
// @Produce json
// @Param name path string true "Pool name"
// @Success 200 {object} dto.MessageResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /zfs/pools/{name}/scrub [post]
func (h *ZFSHandler) ScrubPool(c *gin.Context) {
	name := c.Param("name")

	if err := h.service.ScrubPool(name); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Failed to start scrub",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.MessageResponse{
		Message: "Scrub started successfully",
	})
}

// Dataset Handlers

// CreateDataset creates a new ZFS dataset
// @Summary Create ZFS dataset
// @Tags ZFS Datasets
// @Accept json
// @Produce json
// @Param dataset body dto.CreateDatasetRequest true "Dataset creation request"
// @Success 201 {object} dto.DatasetResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /zfs/datasets [post]
func (h *ZFSHandler) CreateDataset(c *gin.Context) {
	var req dto.CreateDatasetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid request",
			Details: err.Error(),
		})
		return
	}

	dataset, err := h.service.CreateDataset(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Failed to create dataset",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, dataset)
}

// ListDatasets lists all ZFS datasets
// @Summary List ZFS datasets
// @Tags ZFS Datasets
// @Produce json
// @Param pool query string false "Filter by pool"
// @Param status query string false "Filter by status"
// @Param page query int false "Page number"
// @Param page_size query int false "Page size"
// @Success 200 {object} dto.ListResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /zfs/datasets [get]
func (h *ZFSHandler) ListDatasets(c *gin.Context) {
	var req dto.ListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid query parameters",
			Details: err.Error(),
		})
		return
	}

	resp, err := h.service.ListDatasets(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Failed to list datasets",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// UpdateDataset updates a ZFS dataset
// @Summary Update ZFS dataset
// @Tags ZFS Datasets
// @Accept json
// @Produce json
// @Param name path string true "Dataset name"
// @Param dataset body dto.UpdateDatasetRequest true "Dataset update request"
// @Success 200 {object} dto.DatasetResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /zfs/datasets/{name} [put]
func (h *ZFSHandler) UpdateDataset(c *gin.Context) {
	name := c.Param("name")

	var req dto.UpdateDatasetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid request",
			Details: err.Error(),
		})
		return
	}

	dataset, err := h.service.UpdateDataset(name, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Failed to update dataset",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dataset)
}

// DestroyDataset destroys a ZFS dataset
// @Summary Destroy ZFS dataset
// @Tags ZFS Datasets
// @Produce json
// @Param name path string true "Dataset name"
// @Param recursive query bool false "Recursive destroy"
// @Success 200 {object} dto.MessageResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /zfs/datasets/{name} [delete]
func (h *ZFSHandler) DestroyDataset(c *gin.Context) {
	name := c.Param("name")
	recursive := c.Query("recursive") == "true"

	if err := h.service.DestroyDataset(name, recursive); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Failed to destroy dataset",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.MessageResponse{
		Message: "Dataset destroyed successfully",
	})
}

// Snapshot Handlers

// CreateSnapshot creates a new ZFS snapshot
// @Summary Create ZFS snapshot
// @Tags ZFS Snapshots
// @Accept json
// @Produce json
// @Param snapshot body dto.CreateSnapshotRequest true "Snapshot creation request"
// @Success 201 {object} dto.SnapshotResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /zfs/snapshots [post]
func (h *ZFSHandler) CreateSnapshot(c *gin.Context) {
	var req dto.CreateSnapshotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid request",
			Details: err.Error(),
		})
		return
	}

	snapshot, err := h.service.CreateSnapshot(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Failed to create snapshot",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, snapshot)
}

// ListSnapshots lists all ZFS snapshots
// @Summary List ZFS snapshots
// @Tags ZFS Snapshots
// @Produce json
// @Param pool query string false "Filter by pool"
// @Param status query string false "Filter by status"
// @Param page query int false "Page number"
// @Param page_size query int false "Page size"
// @Success 200 {object} dto.ListResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /zfs/snapshots [get]
func (h *ZFSHandler) ListSnapshots(c *gin.Context) {
	var req dto.ListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid query parameters",
			Details: err.Error(),
		})
		return
	}

	resp, err := h.service.ListSnapshots(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Failed to list snapshots",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// RollbackSnapshot rolls back to a snapshot
// @Summary Rollback to snapshot
// @Tags ZFS Snapshots
// @Accept json
// @Produce json
// @Param rollback body dto.RollbackSnapshotRequest true "Rollback request"
// @Success 200 {object} dto.MessageResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /zfs/snapshots/rollback [post]
func (h *ZFSHandler) RollbackSnapshot(c *gin.Context) {
	var req dto.RollbackSnapshotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid request",
			Details: err.Error(),
		})
		return
	}

	if err := h.service.RollbackSnapshot(req); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Failed to rollback snapshot",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.MessageResponse{
		Message: "Snapshot rolled back successfully",
	})
}

// CloneSnapshot clones a snapshot
// @Summary Clone snapshot
// @Tags ZFS Snapshots
// @Accept json
// @Produce json
// @Param clone body dto.CloneSnapshotRequest true "Clone request"
// @Success 200 {object} dto.MessageResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /zfs/snapshots/clone [post]
func (h *ZFSHandler) CloneSnapshot(c *gin.Context) {
	var req dto.CloneSnapshotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid request",
			Details: err.Error(),
		})
		return
	}

	if err := h.service.CloneSnapshot(req); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Failed to clone snapshot",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.MessageResponse{
		Message: "Snapshot cloned successfully",
	})
}

// DestroySnapshot destroys a ZFS snapshot
// @Summary Destroy ZFS snapshot
// @Tags ZFS Snapshots
// @Produce json
// @Param name path string true "Snapshot name"
// @Success 200 {object} dto.MessageResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /zfs/snapshots/{name} [delete]
func (h *ZFSHandler) DestroySnapshot(c *gin.Context) {
	name := c.Param("name")

	if err := h.service.DestroySnapshot(name); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Failed to destroy snapshot",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.MessageResponse{
		Message: "Snapshot destroyed successfully",
	})
}

// Volume Handlers

// CreateVolume creates a new ZFS volume
// @Summary Create ZFS volume
// @Tags ZFS Volumes
// @Accept json
// @Produce json
// @Param volume body dto.CreateVolumeRequest true "Volume creation request"
// @Success 201 {object} dto.VolumeResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /zfs/volumes [post]
func (h *ZFSHandler) CreateVolume(c *gin.Context) {
	var req dto.CreateVolumeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid request",
			Details: err.Error(),
		})
		return
	}

	volume, err := h.service.CreateVolume(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Failed to create volume",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, volume)
}

// ListVolumes lists all ZFS volumes
// @Summary List ZFS volumes
// @Tags ZFS Volumes
// @Produce json
// @Param pool query string false "Filter by pool"
// @Param status query string false "Filter by status"
// @Param page query int false "Page number"
// @Param page_size query int false "Page size"
// @Success 200 {object} dto.ListResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /zfs/volumes [get]
func (h *ZFSHandler) ListVolumes(c *gin.Context) {
	var req dto.ListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid query parameters",
			Details: err.Error(),
		})
		return
	}

	resp, err := h.service.ListVolumes(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Failed to list volumes",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ResizeVolume resizes a ZFS volume
// @Summary Resize ZFS volume
// @Tags ZFS Volumes
// @Accept json
// @Produce json
// @Param name path string true "Volume name"
// @Param resize body dto.ResizeVolumeRequest true "Resize request"
// @Success 200 {object} dto.VolumeResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /zfs/volumes/{name}/resize [put]
func (h *ZFSHandler) ResizeVolume(c *gin.Context) {
	name := c.Param("name")

	var req dto.ResizeVolumeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid request",
			Details: err.Error(),
		})
		return
	}

	volume, err := h.service.ResizeVolume(name, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Failed to resize volume",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, volume)
}

// DestroyVolume destroys a ZFS volume
// @Summary Destroy ZFS volume
// @Tags ZFS Volumes
// @Produce json
// @Param name path string true "Volume name"
// @Success 200 {object} dto.MessageResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /zfs/volumes/{name} [delete]
func (h *ZFSHandler) DestroyVolume(c *gin.Context) {
	name := c.Param("name")

	if err := h.service.DestroyVolume(name); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Failed to destroy volume",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.MessageResponse{
		Message: "Volume destroyed successfully",
	})
}
