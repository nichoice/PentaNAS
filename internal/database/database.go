package database

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"pnas/internal/config"
)

var DB *gorm.DB

// InitDatabase initializes the database connection based on configuration
func InitDatabase(cfg *config.DatabaseConfig) error {
	var err error
	var dialector gorm.Dialector

	// Configure logger
	gormLogger := logger.Default.LogMode(logger.Info)

	// Select database driver based on configuration
	switch cfg.Type {
	case "postgres", "postgresql":
		log.Println("Connecting to PostgreSQL database...")
		dsn := buildPostgresDSN(cfg)
		dialector = postgres.Open(dsn)

	case "sqlite", "sqlite3", "":
		log.Println("Connecting to SQLite database...")
		dialector = sqlite.Open(cfg.Path)

	default:
		return fmt.Errorf("unsupported database type: %s (supported: sqlite, postgres)", cfg.Type)
	}

	// Open database connection
	DB, err = gorm.Open(dialector, &gorm.Config{
		Logger: gormLogger,
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Get underlying SQL database
	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %w", err)
	}

	// Set connection pool settings
	if cfg.MaxOpenConns > 0 {
		sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	}
	if cfg.MaxIdleConns > 0 {
		sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	}
	if cfg.ConnMaxLifetime > 0 {
		sqlDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second)
	}

	// Test connection
	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	log.Printf("Database connection established (type: %s)", cfg.Type)
	return nil
}

// buildPostgresDSN builds PostgreSQL connection string
func buildPostgresDSN(cfg *config.DatabaseConfig) string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.Password,
		cfg.DBName,
		cfg.SSLMode,
	)
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

// Close closes the database connection
func Close() error {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}

// GetDatabaseType returns the current database type
func GetDatabaseType() string {
	if DB == nil {
		return "unknown"
	}

	switch DB.Dialector.Name() {
	case "postgres":
		return "postgres"
	case "sqlite":
		return "sqlite"
	default:
		return DB.Dialector.Name()
	}
}

// IsPostgreSQL checks if current database is PostgreSQL
func IsPostgreSQL() bool {
	return GetDatabaseType() == "postgres"
}

// IsSQLite checks if current database is SQLite
func IsSQLite() bool {
	return GetDatabaseType() == "sqlite"
}
