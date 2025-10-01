package migrations

import (
	"gorm.io/gorm"
	"pnas/internal/models"
)

func init() {
	Migrations = append(Migrations, &Migration{
		ID:   "006_create_file_tables",
		Name: "Create file management tables",
		Up:   up006,
		Down: down006,
	})
}

func up006(db *gorm.DB) error {
	// Create all file management tables
	if err := db.AutoMigrate(
		&models.File{},
		&models.FilePermission{},
		&models.FileVersion{},
		&models.FileActivity{},
		&models.FileShare{},
		&models.FileTag{},
		&models.FileTagMapping{},
		&models.RecycleBin{},
		&models.FileIndex{},
		&models.StorageStats{},
	); err != nil {
		return err
	}

	// Create additional indexes for performance optimization
	// Note: Index syntax differs between SQLite and PostgreSQL

	dbType := db.Dialector.Name()

	// Files table - composite indexes for common queries
	if dbType == "postgres" {
		// PostgreSQL syntax
		if err := db.Exec(`
			CREATE INDEX IF NOT EXISTS idx_files_owner_deleted
			ON files(owner_id, is_deleted, updated_at DESC)
		`).Error; err != nil {
			return err
		}

		if err := db.Exec(`
			CREATE INDEX IF NOT EXISTS idx_files_parent_name_deleted
			ON files(parent_id, name, is_deleted) WHERE parent_id IS NOT NULL
		`).Error; err != nil {
			return err
		}
	} else {
		// SQLite syntax
		if err := db.Exec(`
			CREATE INDEX IF NOT EXISTS idx_files_owner_deleted
			ON files(owner_id, is_deleted, updated_at DESC)
		`).Error; err != nil {
			return err
		}

		if err := db.Exec(`
			CREATE INDEX IF NOT EXISTS idx_files_parent_name_deleted
			ON files(parent_id, name, is_deleted)
		`).Error; err != nil {
			return err
		}
	}

	// File index - for fast search queries
	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_file_index_search
		ON file_index(name_lower, extension, is_deleted)
	`).Error; err != nil {
		return err
	}

	// File activities - for recent activity queries
	if dbType == "postgres" {
		if err := db.Exec(`
			CREATE INDEX IF NOT EXISTS idx_file_activities_recent
			ON file_activities(created_at DESC, user_id, action)
		`).Error; err != nil {
			return err
		}
	} else {
		if err := db.Exec(`
			CREATE INDEX IF NOT EXISTS idx_file_activities_recent
			ON file_activities(created_at DESC, user_id, action)
		`).Error; err != nil {
			return err
		}
	}

	// Recycle bin - for auto cleanup
	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_recycle_bin_cleanup
		ON recycle_bin(auto_delete_at, deleted_by)
	`).Error; err != nil {
		return err
	}

	return nil
}

func down006(db *gorm.DB) error {
	// Drop indexes first
	db.Exec(`DROP INDEX IF EXISTS idx_files_owner_deleted`)
	db.Exec(`DROP INDEX IF EXISTS idx_files_parent_name_deleted`)
	db.Exec(`DROP INDEX IF EXISTS idx_file_index_search`)
	db.Exec(`DROP INDEX IF EXISTS idx_file_activities_recent`)
	db.Exec(`DROP INDEX IF EXISTS idx_recycle_bin_cleanup`)

	// Drop tables
	return db.Migrator().DropTable(
		&models.StorageStats{},
		&models.FileIndex{},
		&models.RecycleBin{},
		&models.FileTagMapping{},
		&models.FileTag{},
		&models.FileShare{},
		&models.FileActivity{},
		&models.FileVersion{},
		&models.FilePermission{},
		&models.File{},
	)
}
