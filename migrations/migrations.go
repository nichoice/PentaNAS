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
	database.RegisterMigration("20240101000005", "create_nfs_tables", createNFSTables, dropNFSTables)
	database.RegisterMigration("20240101000006", "create_iscsi_tables", createiSCSITables, dropiSCSITables)
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

func createNFSTables(db *gorm.DB) error {
	tables := []interface{}{
		&models.NFSExport{},
		&models.NFSClientAccess{},
		&models.NFSSnapshot{},
		&models.NFSMultipathConf{},
		&models.NFSGlobalConfig{},
		&models.NFSService{},
		&models.NFSConnection{},
		&models.NFSAuditLog{},
		&models.NFSQuota{},
	}

	for _, table := range tables {
		if err := db.AutoMigrate(table); err != nil {
			return err
		}
	}

	// Initialize default NFS configuration
	return initNFSConfig(db)
}

func dropNFSTables(db *gorm.DB) error {
	tables := []interface{}{
		&models.NFSQuota{},
		&models.NFSAuditLog{},
		&models.NFSConnection{},
		&models.NFSService{},
		&models.NFSMultipathConf{},
		&models.NFSSnapshot{},
		&models.NFSClientAccess{},
		&models.NFSExport{},
		&models.NFSGlobalConfig{},
	}

	for _, table := range tables {
		if err := db.Migrator().DropTable(table); err != nil {
			return err
		}
	}
	return nil
}

func initNFSConfig(db *gorm.DB) error {
	var config models.NFSGlobalConfig
	result := db.Where("is_active = ?", true).First(&config)

	if result.Error != nil {
		defaultConfig := models.NFSGlobalConfig{
			Base:                models.Base{ID: "default-nfs-config"},
			Version:             models.NFSVersion4,
			PortmapperPort:      111,
			NFSPort:             2049,
			MountdPort:          20048,
			StatdPort:           662,
			LockdPort:           32803,
			ThreadCount:         8,
			MaxConnections:      1024,
			ReadAhead:           128,
			WriteBuffer:         128,
			AttributeTimeout:    60,
			DirectoryTimeout:    60,
			RequireSecurePort:   false,
			EnableTCP:           true,
			EnableUDP:           false,
			LogLevel:            1,
			LogFile:             "/var/log/nfs.log",
			EnableDebugLog:      false,
			EnableMultipath:     false,
			MultipathPolicy:     "round_robin",
			HealthCheckInterval: 30,
			IsActive:            true,
		}
		return db.Create(&defaultConfig).Error
	}
	return nil
}

func createiSCSITables(db *gorm.DB) error {
	tables := []interface{}{
		&models.ISCSITarget{},
		&models.ISCSILUN{},
		&models.ISCSIACL{},
		&models.ISCSILUNMapping{},
		&models.ISCSIGlobalConfig{},
		&models.ISCSIService{},
		&models.ISCSISession{},
		&models.ISCSIConnection{},
		&models.ISCSIAuditLog{},
		&models.ISCSIStoragePool{},
		&models.ISCSIPerformanceStats{},
	}

	for _, table := range tables {
		if err := db.AutoMigrate(table); err != nil {
			return err
		}
	}

	// Initialize default iSCSI configuration
	return initiSCSIConfig(db)
}

func dropiSCSITables(db *gorm.DB) error {
	tables := []interface{}{
		&models.ISCSIPerformanceStats{},
		&models.ISCSIStoragePool{},
		&models.ISCSIAuditLog{},
		&models.ISCSIConnection{},
		&models.ISCSISession{},
		&models.ISCSIService{},
		&models.ISCSILUNMapping{},
		&models.ISCSIACL{},
		&models.ISCSILUN{},
		&models.ISCSITarget{},
		&models.ISCSIGlobalConfig{},
	}

	for _, table := range tables {
		if err := db.Migrator().DropTable(table); err != nil {
			return err
		}
	}
	return nil
}

func initiSCSIConfig(db *gorm.DB) error {
	var config models.ISCSIGlobalConfig
	result := db.Where("is_active = ?", true).First(&config)

	if result.Error != nil {
		defaultConfig := models.ISCSIGlobalConfig{
			Base:                         models.Base{ID: "default-iscsi-config"},
			TargetPort:                   3260,
			MaxSessions:                  256,
			MaxConnections:               1,
			MaxRecvDataSegmentLength:     8192,
			MaxXmitDataSegmentLength:     8192,
			MaxBurstLength:               262144,
			FirstBurstLength:             65536,
			MaxOutstandingR2T:            1,
			DefaultTime2Wait:             2,
			DefaultTime2Retain:           20,
			LoginTimeout:                 30,
			LogoutTimeout:                30,
			RequireAuth:                  false,
			AllowDuplicateSessions:       false,
			LogLevel:                     1,
			LogFile:                      "/var/log/iscsi/iscsi.log",
			EnableDebugLog:               false,
			IsActive:                     true,
		}
		return db.Create(&defaultConfig).Error
	}
	return nil
}