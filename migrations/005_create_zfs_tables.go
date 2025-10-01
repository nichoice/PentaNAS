package migrations

import (
	"pnas/internal/database"
	"pnas/internal/models"

	"gorm.io/gorm"
)

func init() {
	database.RegisterMigration("005", "create_zfs_tables", up005, down005)
}

func up005(db *gorm.DB) error {
	// Create ZFS tables
	return db.AutoMigrate(
		&models.ZFSPool{},
		&models.ZFSDataset{},
		&models.ZFSVolume{},
		&models.ZFSSnapshot{},
		&models.ZFSCache{},
		&models.ZFSEncryptionKey{},
		&models.ZFSScrub{},
	)
}

func down005(db *gorm.DB) error {
	// Drop ZFS tables
	return db.Migrator().DropTable(
		&models.ZFSPool{},
		&models.ZFSDataset{},
		&models.ZFSVolume{},
		&models.ZFSSnapshot{},
		&models.ZFSCache{},
		&models.ZFSEncryptionKey{},
		&models.ZFSScrub{},
	)
}
