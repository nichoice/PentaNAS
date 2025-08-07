package tests

import (
	"log"
	"os"
	"pnas/internal/config"
	"pnas/internal/database"
	"pnas/internal/repositories"
	"testing"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

var (
	testDB     *gorm.DB
	testLogger *zap.Logger
	userRepo   repositories.UserRepository
	groupRepo  repositories.UserGroupRepository
	healthRepo repositories.HealthCheckRepository
)

// TestMain 测试主函数，在所有测试运行前后执行
func TestMain(m *testing.M) {
	// 设置测试环境
	os.Setenv("GO_ENV", "test")
	
	// 初始化测试日志器
	var err error
	testLogger, err = config.InitLogger("../config/logger-dev.yml")
	if err != nil {
		log.Fatalf("初始化测试日志器失败: %v", err)
	}
	
	// 初始化测试数据库（内存数据库）
	testDB, err = config.InitDatabase("../config/database.yml", testLogger)
	if err != nil {
		log.Fatalf("初始化测试数据库失败: %v", err)
	}
	
	// 运行数据库迁移
	if err := database.AutoMigrate(testDB, testLogger); err != nil {
		log.Fatalf("测试数据库迁移失败: %v", err)
	}
	
	// 初始化仓库
	userRepo = repositories.NewUserRepository(testDB, testLogger)
	groupRepo = repositories.NewUserGroupRepository(testDB, testLogger)
	healthRepo = repositories.NewHealthCheckRepository(testDB, testLogger)
	
	// 运行测试
	code := m.Run()
	
	// 清理资源
	testLogger.Sync()
	
	os.Exit(code)
}

// cleanupDatabase 清理测试数据库
func cleanupDatabase() {
	testDB.Exec("DELETE FROM users")
	testDB.Exec("DELETE FROM user_groups")
	testDB.Exec("DELETE FROM health_checks")
}