package controllers

import (
	"net/http"
	"pnas/internal/utils"

	"github.com/gin-gonic/gin"
)

// CreateVGRequest represents the request body for creating a volume group
type CreateVGRequest struct {
	VgName  string   `json:"vg_name" binding:"required" example:"my_volume_group"`
	Devices []string `json:"devices" binding:"required" example:["/dev/sdb","/dev/sdc"]`
}

// GetDisks retrieves system disk information
// @Summary Get system disks
// @Description Get a list of all available disks in the system
// @Security BearerAuth
// @Tags Storage
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "Disk information"
// @Failure 500 {object} map[string]string
// @Router /storage/disks [get]
func GetDisks(c *gin.Context) {
	disk := &utils.Disks{}
	result, err := disk.ScanDisks()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

// GetVG retrieves volume group information
// @Summary Get volume groups
// @Description Get a list of all LVM volume groups
// @Security BearerAuth
// @Tags Storage
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "Volume group information"
// @Failure 500 {object} map[string]string
// @Router /storage/vgs [get]
func GetVG(c *gin.Context) {
	vg := &utils.LVM{}
	result, err := vg.GetVGS()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

// GetLV retrieves logical volume information
// @Summary Get logical volumes
// @Description Get a list of all LVM logical volumes
// @Security BearerAuth
// @Tags Storage
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "Logical volume information"
// @Failure 500 {object} map[string]string
// @Router /storage/lvs [get]
func GetLV(c *gin.Context) {
	lv := &utils.LVM{}
	result, err := lv.GetLVS()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

// CreateVG creates a new volume group
// @Summary Create volume group
// @Description Create a new LVM volume group with specified devices
// @Security BearerAuth
// @Tags Storage
// @Accept json
// @Produce json
// @Param vg body CreateVGRequest true "Volume group creation details"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /storage/create_vg [post]
func CreateVG(c *gin.Context) {
	var req CreateVGRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	lvm := &utils.LVM{}
	err := lvm.CreateVG(req.VgName, req.Devices)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Volume group created successfully"})
}
