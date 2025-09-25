package migrations

import (
	"pnas/internal/database"
	"pnas/internal/models"

	"gorm.io/gorm"
)

func init() {
	// Register all migrations here
	database.RegisterMigration("20240101000001", "create_users_table", createUsersTable, dropUsersTable)
	database.RegisterMigration("20240101000002", "create_roles_table", createRolesTable, dropRolesTable)
	database.RegisterMigration("20240101000003", "create_audit_tables", createAuditTables, dropAuditTables)
	database.RegisterMigration("20240101000004", "create_samba_tables", createSambaTables, dropSambaTables)
}

// Migration functions
func createUsersTable(db *gorm.DB) error {
	return db.AutoMigrate(&models.User{})
}

func dropUsersTable(db *gorm.DB) error {
	return db.Migrator().DropTable(&models.User{})
}

func createRolesTable(db *gorm.DB) error {
	if err := db.AutoMigrate(&models.Role{}); err != nil {
		return err
	}
	if err := db.AutoMigrate(&models.UserRole{}); err != nil {
		return err
	}
	models.InitRoles(db)
	return nil
}

func dropRolesTable(db *gorm.DB) error {
	if err := db.Migrator().DropTable(&models.UserRole{}); err != nil {
		return err
	}
	return db.Migrator().DropTable(&models.Role{})
}

func createAuditTables(db *gorm.DB) error {
	if err := db.AutoMigrate(&models.FileAuditLog{}); err != nil {
		return err
	}
	if err := db.AutoMigrate(&models.FileAccessStats{}); err != nil {
		return err
	}
	return db.AutoMigrate(&models.AuditSummary{})
}

func dropAuditTables(db *gorm.DB) error {
	if err := db.Migrator().DropTable(&models.AuditSummary{}); err != nil {
		return err
	}
	if err := db.Migrator().DropTable(&models.FileAccessStats{}); err != nil {
		return err
	}
	return db.Migrator().DropTable(&models.FileAuditLog{})
}

func createSambaTables(db *gorm.DB) error {
	tables := []interface{}{
		&models.SambaAccount{},
		&models.SambaShare{},
		&models.SambaShareAccess{},
		&models.SambaGlobalConfig{},
		&models.SambaService{},
		&models.SambaConnection{},
		&models.SambaAuditLog{},
		&models.RecycleBinItem{},
	}

	for _, table := range tables {
		if err := db.AutoMigrate(table); err != nil {
			return err
		}
	}

	// Initialize default Samba configuration
	return initSambaConfig(db)
}

func dropSambaTables(db *gorm.DB) error {
	tables := []interface{}{
		&models.RecycleBinItem{},
		&models.SambaAuditLog{},
		&models.SambaConnection{},
		&models.SambaService{},
		&models.SambaShareAccess{},
		&models.SambaShare{},
		&models.SambaGlobalConfig{},
		&models.SambaAccount{},
	}

	for _, table := range tables {
		if err := db.Migrator().DropTable(table); err != nil {
			return err
		}
	}
	return nil
}

func initSambaConfig(db *gorm.DB) error {
	var config models.SambaGlobalConfig
	result := db.Where("is_active = ?", true).First(&config)

	if result.Error != nil {
		defaultConfig := models.SambaGlobalConfig{
			Base:                 models.Base{ID: "default-samba-config"},
			Version:              models.SambaVersion4_18,
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
		return db.Create(&defaultConfig).Error
	}
	return nil
}