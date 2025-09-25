package database

import (
	"fmt"
	"log"
	"pnas/internal/config"

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

	log.Println("Database connection established")
	return nil
}

// InitDatabaseWithMigrations initializes the database and runs migrations
func InitDatabaseWithMigrations(cfg *config.DatabaseConfig) error {
	if err := InitDatabase(cfg); err != nil {
		return err
	}

	// Run migrations
	if err := Migrate(DB); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Println("Database migrations completed")
	return nil
}