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

// InitSambaModels initializes the Samba-related models in the database
func InitSambaModels(db *gorm.DB) {
	// Auto-migrate all Samba models
	err := db.AutoMigrate(
		&SambaAccount{},
		&SambaShare{},
		&SambaShareAccess{},
		&SambaGlobalConfig{},
		&SambaService{},
		&SambaConnection{},
		&SambaAuditLog{},
		&RecycleBinItem{},
	)

	if err != nil {
		log.Printf("Failed to migrate Samba models: %v", err)
	} else {
		log.Println("Samba models migrated successfully")
	}

	// Initialize default Samba global configuration
	var config SambaGlobalConfig
	result := db.Where("is_active = ?", true).First(&config)

	if result.Error != nil {
		// Create default configuration
		defaultConfig := SambaGlobalConfig{
			Base:                 Base{ID: "default-samba-config"},
			Version:              SambaVersion4_18,
			ServerString:         "PNAS Samba Server",
			Workgroup:            "WORKGROUP",
			NetbiosName:          "PNAS",
			SecurityLevel:        "user",
			EncryptPasswords:     true,
			PassdbBackend:        "tdbsam",
			Interfaces:           "",
			BindInterfacesOnly:   false,
			SocketOptions:        "TCP_NODELAY IPTOS_LOWDELAY SO_RCVBUF=524288 SO_SNDBUF=524288",
			LogLevel:             1,
			LogFile:              "/var/log/samba/samba.log",
			MaxLogSize:           5000,
			DeadTime:             15,
			GetWDCacheTime:       10,
			LPQCacheTime:         10,
			MaxConnections:       0,
			EnableMultiChannel:   false,
			MaxChannels:          4,
			MapToGuest:           "Never",
			GuestAccount:         "nobody",
			HostsAllow:           "",
			HostsDeny:            "",
			EnableAuditing:       false,
			AuditPrefix:          "",
			FullAuditPrefix:      "",
			IsActive:             true,
		}

		if err := db.Create(&defaultConfig).Error; err != nil {
			log.Printf("Failed to create default Samba configuration: %v", err)
		} else {
			log.Println("Created default Samba configuration")
		}
	} else {
		log.Println("Samba configuration already exists")
	}
}