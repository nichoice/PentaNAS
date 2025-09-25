package cmd

import (
	"fmt"
	"log"
	"pnas/internal/config"
	"pnas/internal/database"
	"strings"

	"github.com/spf13/cobra"
)

import _ "pnas/migrations" // Import to register migrations

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "数据库迁移管理",
	Long:  "管理数据库迁移，包括运行迁移、回滚、查看状态等操作",
}

var migrateUpCmd = &cobra.Command{
	Use:   "up",
	Short: "运行所有待执行的迁移",
	Long:  "运行所有尚未执行的数据库迁移",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runWithDB(func() error {
			return database.Migrate(database.DB)
		})
	},
}

var migrateDownCmd = &cobra.Command{
	Use:   "down [version]",
	Short: "回滚迁移",
	Long:  "回滚到指定版本或回滚最后一个迁移。如果不指定版本，将回滚最后一个迁移",
	RunE: func(cmd *cobra.Command, args []string) error {
		var targetVersion string
		if len(args) > 0 {
			targetVersion = args[0]
		}

		return runWithDB(func() error {
			return database.Rollback(database.DB, targetVersion)
		})
	},
}

var migrateStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "查看迁移状态",
	Long:  "显示所有迁移的状态，包括已应用和待执行的迁移",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runWithDB(func() error {
			return database.Status(database.DB)
		})
	},
}

var migrateCreateCmd = &cobra.Command{
	Use:   "create [name]",
	Short: "创建新的迁移文件",
	Long:  "创建一个新的迁移文件模板",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return database.CreateMigration(strings.Join(args, " "))
	},
}

var migrateResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "重置数据库",
	Long:  "删除所有表并重新运行所有迁移（危险操作）",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Ask for confirmation
		fmt.Print("警告：这将删除所有数据！确认要继续吗？(y/N): ")
		var response string
		fmt.Scanln(&response)

		if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
			fmt.Println("操作已取消")
			return nil
		}

		return runWithDB(func() error {
			return database.Reset(database.DB)
		})
	},
}

var migrateFreshCmd = &cobra.Command{
	Use:   "fresh",
	Short: "刷新数据库",
	Long:  "删除所有表并重新运行所有迁移（reset的别名）",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Ask for confirmation
		fmt.Print("警告：这将删除所有数据！确认要继续吗？(y/N): ")
		var response string
		fmt.Scanln(&response)

		if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
			fmt.Println("操作已取消")
			return nil
		}

		return runWithDB(func() error {
			return database.Fresh(database.DB)
		})
	},
}

// runWithDB initializes the database and runs the given function
func runWithDB(fn func() error) error {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Printf("Warning: Failed to load config: %v", err)
		// Use default database config
		cfg = &config.Config{
			Database: config.DatabaseConfig{
				Path: "pnas.db",
			},
		}
	}

	// Initialize database
	if err := database.InitDatabase(&cfg.Database); err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}

	// Run the function
	return fn()
}

func init() {
	// Add migrate command to root
	rootCmd.AddCommand(migrateCmd)

	// Add subcommands to migrate
	migrateCmd.AddCommand(migrateUpCmd)
	migrateCmd.AddCommand(migrateDownCmd)
	migrateCmd.AddCommand(migrateStatusCmd)
	migrateCmd.AddCommand(migrateCreateCmd)
	migrateCmd.AddCommand(migrateResetCmd)
	migrateCmd.AddCommand(migrateFreshCmd)
}