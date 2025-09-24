package database

import (
	"fmt"
	"log"
	"pnas/internal/config"
	"pnas/internal/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// InitDatabase initializes the database connection
func InitDatabase(cfg *config.DatabaseConfig) error {
	var err error
	
	// Connect to SQLite database
	DB, err = gorm.Open(sqlite.Open(cfg.Path), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	
	// Run migrations
	if err := DB.AutoMigrate(
		&models.User{},
		&models.Role{},
		&models.UserRole{},
		&models.FileAuditLog{},
		&models.FileAccessStats{},
		&models.AuditSummary{},
	); err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	// Initialize Samba models
	models.InitSambaModels(DB)

	log.Println("Database connection established and migrations completed")
	return nil
}