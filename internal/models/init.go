package models

import (
	"log"

	"gorm.io/gorm"
)

// InitRoles initializes the predefined roles in the database
func InitRoles(db *gorm.DB) {
	roles := []Role{
		{
			Base:        Base{ID: RoleSuperAdmin},
			Name:        "超级管理员",
			Description: "拥有系统最高权限，可以管理所有功能模块",
		},
		{
			Base:        Base{ID: RoleNormalUser},
			Name:        "普通用户",
			Description: "只能访问其有权限的文件管理中的文件",
		},
		{
			Base:        Base{ID: RoleAuditUser},
			Name:        "审计用户",
			Description: "具有查看系统日志和审计记录的权限",
		},
		{
			Base:        Base{ID: RoleOpsUser},
			Name:        "运维用户",
			Description: "具有系统运维相关权限，如监控、备份等",
		},
	}

	for _, role := range roles {
		// Check if role already exists
		var existingRole Role
		result := db.Where("id = ?", role.ID).First(&existingRole)
		
		if result.Error != nil {
			// Role doesn't exist, create it
			if err := db.Create(&role).Error; err != nil {
				log.Printf("Failed to create role %s: %v", role.Name, err)
			} else {
				log.Printf("Created role: %s", role.Name)
			}
		} else {
			log.Printf("Role %s already exists", role.Name)
		}
	}
}