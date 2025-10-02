package database

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"gorm.io/gorm"
)

type Migration struct {
	ID        uint   `gorm:"primaryKey"`
	Version   string `gorm:"unique;not null"`
	Name      string `gorm:"not null"`
	AppliedAt time.Time
	Rollback  *string // SQL for rollback (nullable)
}

type MigrationFunc func(*gorm.DB) error

type MigrationFile struct {
	Version string
	Name    string
	Up      MigrationFunc
	Down    MigrationFunc
}

var migrations = make(map[string]*MigrationFile)

// RegisterMigration registers a migration
func RegisterMigration(version, name string, up, down MigrationFunc) {
	migrations[version] = &MigrationFile{
		Version: version,
		Name:    name,
		Up:      up,
		Down:    down,
	}
}

// Migrate runs all pending migrations
func Migrate(db *gorm.DB) error {
	// Ensure migrations table exists
	if err := db.AutoMigrate(&Migration{}); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get applied migrations
	var appliedMigrations []Migration
	if err := db.Order("version").Find(&appliedMigrations).Error; err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	appliedVersions := make(map[string]bool)
	for _, m := range appliedMigrations {
		appliedVersions[m.Version] = true
	}

	// Get all migration versions and sort them
	versions := make([]string, 0, len(migrations))
	for version := range migrations {
		versions = append(versions, version)
	}
	sort.Strings(versions)

	// Run pending migrations
	for _, version := range versions {
		if appliedVersions[version] {
			continue
		}

		migration := migrations[version]
		log.Printf("Running migration %s: %s", version, migration.Name)

		// Start transaction
		tx := db.Begin()
		if tx.Error != nil {
			return fmt.Errorf("failed to start transaction: %w", tx.Error)
		}

		// Run migration
		if err := migration.Up(tx); err != nil {
			tx.Rollback()
			return fmt.Errorf("migration %s failed: %w", version, err)
		}

		// Record migration
		migrationRecord := Migration{
			Version:   version,
			Name:      migration.Name,
			AppliedAt: time.Now(),
		}
		if err := tx.Create(&migrationRecord).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to record migration %s: %w", version, err)
		}

		// Commit transaction
		if err := tx.Commit().Error; err != nil {
			return fmt.Errorf("failed to commit migration %s: %w", version, err)
		}

		log.Printf("Migration %s completed successfully", version)
	}

	log.Println("All migrations completed")
	return nil
}

// Rollback rolls back the last migration or to a specific version
func Rollback(db *gorm.DB, targetVersion string) error {
	// Get applied migrations in reverse order
	var appliedMigrations []Migration
	var query *gorm.DB

	if targetVersion != "" {
		query = db.Where("version > ?", targetVersion).Order("version DESC")
	} else {
		query = db.Order("version DESC").Limit(1)
	}

	if err := query.Find(&appliedMigrations).Error; err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	if len(appliedMigrations) == 0 {
		log.Println("No migrations to rollback")
		return nil
	}

	// Rollback migrations
	for _, migrationRecord := range appliedMigrations {
		migration, exists := migrations[migrationRecord.Version]
		if !exists {
			return fmt.Errorf("migration %s not found in registered migrations", migrationRecord.Version)
		}

		if migration.Down == nil {
			return fmt.Errorf("migration %s does not support rollback", migrationRecord.Version)
		}

		log.Printf("Rolling back migration %s: %s", migrationRecord.Version, migrationRecord.Name)

		// Start transaction
		tx := db.Begin()
		if tx.Error != nil {
			return fmt.Errorf("failed to start transaction: %w", tx.Error)
		}

		// Run rollback
		if err := migration.Down(tx); err != nil {
			tx.Rollback()
			return fmt.Errorf("rollback %s failed: %w", migrationRecord.Version, err)
		}

		// Remove migration record
		if err := tx.Delete(&migrationRecord).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to remove migration record %s: %w", migrationRecord.Version, err)
		}

		// Commit transaction
		if err := tx.Commit().Error; err != nil {
			return fmt.Errorf("failed to commit rollback %s: %w", migrationRecord.Version, err)
		}

		log.Printf("Rollback %s completed successfully", migrationRecord.Version)
	}

	return nil
}

// Status shows migration status
func Status(db *gorm.DB) error {
	// Get applied migrations
	var appliedMigrations []Migration
	if err := db.Order("version").Find(&appliedMigrations).Error; err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	appliedVersions := make(map[string]*Migration)
	for i := range appliedMigrations {
		appliedVersions[appliedMigrations[i].Version] = &appliedMigrations[i]
	}

	// Get all migration versions and sort them
	versions := make([]string, 0, len(migrations))
	for version := range migrations {
		versions = append(versions, version)
	}
	sort.Strings(versions)

	fmt.Println("Migration Status:")
	fmt.Println("================")

	for _, version := range versions {
		migration := migrations[version]
		if applied, exists := appliedVersions[version]; exists {
			fmt.Printf("[✓] %s - %s (applied at %s)\n", version, migration.Name, applied.AppliedAt.Format("2006-01-02 15:04:05"))
		} else {
			fmt.Printf("[ ] %s - %s (pending)\n", version, migration.Name)
		}
	}

	return nil
}

// CreateMigration creates a new migration file
func CreateMigration(name string) error {
	if name == "" {
		return fmt.Errorf("migration name is required")
	}

	// Generate version (timestamp)
	version := time.Now().Format("20060102150405")

	// Clean name
	cleanName := strings.ReplaceAll(strings.ToLower(name), " ", "_")

	// Create migrations directory if it doesn't exist
	migrationsDir := "migrations"
	if err := os.MkdirAll(migrationsDir, 0755); err != nil {
		return fmt.Errorf("failed to create migrations directory: %w", err)
	}

	// Create migration file
	filename := fmt.Sprintf("%s_%s.go", version, cleanName)
	filepath := filepath.Join(migrationsDir, filename)

	content := fmt.Sprintf(`package migrations

import (
	"pnas/internal/database"
	"gorm.io/gorm"
)

func init() {
	database.RegisterMigration("%s", "%s", up_%s, down_%s)
}

func up_%s(db *gorm.DB) error {
	// TODO: Write your migration here
	return nil
}

func down_%s(db *gorm.DB) error {
	// TODO: Write your rollback here
	return nil
}
`, version, name, version, version, version, version)

	if err := os.WriteFile(filepath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to create migration file: %w", err)
	}

	fmt.Printf("Created migration: %s\n", filepath)
	return nil
}

// Reset drops all tables and re-runs all migrations
func Reset(db *gorm.DB) error {
	log.Println("Warning: This will drop all tables and data!")

	// Drop migrations table
	if err := db.Migrator().DropTable(&Migration{}); err != nil {
		log.Printf("Warning: failed to drop migrations table: %v", err)
	}

	// Get all table names using GORM migrator to support multiple dialects
	tables, err := db.Migrator().GetTables()
	if err != nil {
		return fmt.Errorf("failed to get table names: %w", err)
	}

	// Drop all tables
	for _, table := range tables {
		if err := db.Migrator().DropTable(table); err != nil {
			log.Printf("Warning: failed to drop table %s: %v", table, err)
		}
	}

	log.Println("All tables dropped, running migrations...")
	return Migrate(db)
}

// Fresh drops all tables and re-runs all migrations (alias for Reset)
func Fresh(db *gorm.DB) error {
	return Reset(db)
}
