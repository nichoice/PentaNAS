package controllers

import (
	"pnas/internal/utils"

	"github.com/gin-gonic/gin"
)

type CreateVGRequest struct {
	VgName  string   `json:"vg_name" binding:"required"`
	Devices []string `json:"devices" binding:"required"`
}

func GetDisks(c *gin.Context) {
	disk := &utils.Disks{}
	result, err := disk.ScanDisks()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, result)
}

func GetVG(c *gin.Context) {
	vg := &utils.LVM{}
	result, err := vg.GetVGS()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, result)
}

func GetLV(c *gin.Context) {
	lv := &utils.LVM{}
	result, err := lv.GetLVS()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, result)
}

func CreateVG(c *gin.Context) {

}
