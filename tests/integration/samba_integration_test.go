package integration

import (
	"os"
	"testing"
	"time"

	"pnas/internal/app/dto"
	"pnas/internal/app/services"
	"pnas/internal/database"
	"pnas/internal/infrastructure/samba"
	"pnas/internal/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupSambaIntegrationTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	database.DB = db
	models.InitSambaModels(db)

	t.Cleanup(func() {
		sqlDB, err := db.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
		database.DB = nil
	})

	return db
}

// TestSambaControl_RealLinuxIntegration 在真实 Linux 环境中测试 Samba 管理
func TestSambaControl_RealLinuxIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// 使用真实的 SambaControl
	cli := samba.NewSambaControl()

	t.Run("GetServiceStatus", func(t *testing.T) {
		status, err := cli.GetServiceStatus()
		if err != nil {
			t.Fatalf("GetServiceStatus failed: %v", err)
		}

		t.Logf("Samba service status: running=%v, version=%s", status.IsRunning, status.Version)

		if status.IsRunning {
			if status.ProcessID == 0 {
				t.Error("Expected non-zero process ID when service is running")
			}
			t.Logf("Process ID: %d, Connected users: %d", status.ProcessID, status.ConnectedUsers)
		}
	})

	t.Run("ValidateConfig", func(t *testing.T) {
		err := cli.ValidateConfig()
		if err != nil {
			t.Fatalf("Config validation failed: %v", err)
		}
		t.Log("Configuration validation passed")
	})

	t.Run("ListShares", func(t *testing.T) {
		shares, err := cli.ListShares()
		if err != nil {
			t.Fatalf("ListShares failed: %v", err)
		}

		t.Logf("Found %d shares", len(shares))
		for _, share := range shares {
			t.Logf("  Share: %s -> %s (enabled: %v)", share.Name, share.Path, share.IsEnabled)
		}
	})

	t.Run("ListUsers", func(t *testing.T) {
		users, err := cli.ListUsers()
		if err != nil {
			t.Fatalf("ListUsers failed: %v", err)
		}

		t.Logf("Found %d Samba users", len(users))
		for _, user := range users {
			t.Logf("  User: %s (enabled: %v)", user.Username, user.IsEnabled)
		}
	})
}

// TestSambaControl_ShareLifecycle 测试 Samba 共享的完整生命周期
func TestSambaControl_ShareLifecycle(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	cli := samba.NewSambaControl()
	testShareName := "pnas-test-share"
	testPath := "/tmp/pnas-test-share"

	// 清理函数
	t.Cleanup(func() {
		_ = cli.DeleteShare(testShareName)
		_ = os.RemoveAll(testPath)
	})

	// 创建测试目录
	if err := os.MkdirAll(testPath, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	t.Run("CreateShare", func(t *testing.T) {
		options := map[string]string{
			"writable":   "yes",
			"browseable": "yes",
			"guest ok":   "no",
		}

		err := cli.CreateShare(testShareName, testPath, "PNAS Test Share", options)
		if err != nil {
			t.Fatalf("CreateShare failed: %v", err)
		}

		t.Logf("Successfully created share: %s", testShareName)
	})

	t.Run("VerifyShareExists", func(t *testing.T) {
		shares, err := cli.ListShares()
		if err != nil {
			t.Fatalf("ListShares failed: %v", err)
		}

		found := false
		for _, share := range shares {
			if share.Name == testShareName {
				found = true
				if share.Path != testPath {
					t.Errorf("Expected path %s, got %s", testPath, share.Path)
				}
				t.Logf("Verified share exists: %s -> %s", share.Name, share.Path)
				break
			}
		}

		if !found {
			t.Fatalf("Share %s not found in share list", testShareName)
		}
	})

	t.Run("DisableShare", func(t *testing.T) {
		err := cli.DisableShare(testShareName)
		if err != nil {
			t.Fatalf("DisableShare failed: %v", err)
		}

		t.Logf("Successfully disabled share: %s", testShareName)
	})

	t.Run("EnableShare", func(t *testing.T) {
		err := cli.EnableShare(testShareName)
		if err != nil {
			t.Fatalf("EnableShare failed: %v", err)
		}

		t.Logf("Successfully enabled share: %s", testShareName)
	})

	t.Run("DeleteShare", func(t *testing.T) {
		err := cli.DeleteShare(testShareName)
		if err != nil {
			t.Fatalf("DeleteShare failed: %v", err)
		}

		t.Logf("Successfully deleted share: %s", testShareName)

		// Verify share is removed
		shares, err := cli.ListShares()
		if err != nil {
			t.Fatalf("ListShares failed: %v", err)
		}

		for _, share := range shares {
			if share.Name == testShareName {
				t.Errorf("Share %s still exists after deletion", testShareName)
			}
		}
	})
}

// TestSambaControl_UserLifecycle 测试 Samba 用户的完整生命周期
func TestSambaControl_UserLifecycle(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	cli := samba.NewSambaControl()
	testUsername := "pnas-test-user"
	testPassword := "TestPass123!"

	// 清理函数
	t.Cleanup(func() {
		_ = cli.DeleteUser(testUsername)
		// 也清理系统用户（如果创建的话）
		// Note: 在生产环境中要谨慎删除系统用户
	})

	t.Run("CreateUser", func(t *testing.T) {
		err := cli.CreateUser(testUsername, testPassword)
		if err != nil {
			t.Fatalf("CreateUser failed: %v", err)
		}

		t.Logf("Successfully created user: %s", testUsername)
	})

	t.Run("VerifyUserExists", func(t *testing.T) {
		users, err := cli.ListUsers()
		if err != nil {
			t.Fatalf("ListUsers failed: %v", err)
		}

		found := false
		for _, user := range users {
			if user.Username == testUsername {
				found = true
				if !user.IsEnabled {
					t.Error("Expected newly created user to be enabled")
				}
				t.Logf("Verified user exists: %s (enabled: %v)", user.Username, user.IsEnabled)
				break
			}
		}

		if !found {
			t.Fatalf("User %s not found in user list", testUsername)
		}
	})

	t.Run("SetUserPassword", func(t *testing.T) {
		newPassword := "NewTestPass456!"
		err := cli.SetUserPassword(testUsername, newPassword)
		if err != nil {
			t.Fatalf("SetUserPassword failed: %v", err)
		}

		t.Logf("Successfully changed password for user: %s", testUsername)
	})

	t.Run("DisableUser", func(t *testing.T) {
		err := cli.DisableUser(testUsername)
		if err != nil {
			t.Fatalf("DisableUser failed: %v", err)
		}

		t.Logf("Successfully disabled user: %s", testUsername)

		// Wait a moment for changes to take effect
		time.Sleep(100 * time.Millisecond)

		// Verify user is disabled
		users, err := cli.ListUsers()
		if err != nil {
			t.Fatalf("ListUsers failed: %v", err)
		}

		for _, user := range users {
			if user.Username == testUsername {
				if user.IsEnabled {
					t.Error("Expected user to be disabled")
				}
				break
			}
		}
	})

	t.Run("EnableUser", func(t *testing.T) {
		err := cli.EnableUser(testUsername)
		if err != nil {
			t.Fatalf("EnableUser failed: %v", err)
		}

		t.Logf("Successfully enabled user: %s", testUsername)

		// Wait a moment for changes to take effect
		time.Sleep(100 * time.Millisecond)

		// Verify user is enabled
		users, err := cli.ListUsers()
		if err != nil {
			t.Fatalf("ListUsers failed: %v", err)
		}

		for _, user := range users {
			if user.Username == testUsername {
				if !user.IsEnabled {
					t.Error("Expected user to be enabled")
				}
				break
			}
		}
	})

	t.Run("DeleteUser", func(t *testing.T) {
		err := cli.DeleteUser(testUsername)
		if err != nil {
			t.Fatalf("DeleteUser failed: %v", err)
		}

		t.Logf("Successfully deleted user: %s", testUsername)

		// Verify user is removed
		users, err := cli.ListUsers()
		if err != nil {
			t.Fatalf("ListUsers failed: %v", err)
		}

		for _, user := range users {
			if user.Username == testUsername {
				t.Errorf("User %s still exists after deletion", testUsername)
			}
		}
	})
}

// TestSambaControl_ConfigManagement 测试 Samba 配置管理
func TestSambaControl_ConfigManagement(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	cli := samba.NewSambaControl()

	t.Run("BackupConfig", func(t *testing.T) {
		backupPath, err := cli.BackupConfig()
		if err != nil {
			t.Fatalf("BackupConfig failed: %v", err)
		}

		t.Logf("Successfully created backup: %s", backupPath)

		// Verify backup file exists
		if _, err := os.Stat(backupPath); os.IsNotExist(err) {
			t.Fatalf("Backup file does not exist: %s", backupPath)
		}

		// Clean up
		t.Cleanup(func() {
			_ = os.Remove(backupPath)
		})
	})

	t.Run("ReloadConfig", func(t *testing.T) {
		err := cli.ReloadConfig()
		if err != nil {
			t.Fatalf("ReloadConfig failed: %v", err)
		}

		t.Log("Successfully reloaded configuration")
	})
}

// TestSambaShareService_RealIntegration 测试 Samba 共享服务的集成功能
func TestSambaShareService_RealIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db := setupSambaIntegrationTestDB(t)

	// 使用真实的 SambaControl
	cli := samba.NewSambaControl()
	svc := services.NewSambaShareServiceWithDeps(db, cli)

	testShareName := "pnas-service-test"
	testPath := "/tmp/pnas-service-test"

	// 清理函数
	t.Cleanup(func() {
		// 从数据库中查找共享并删除
		shares, _, _ := svc.GetShares(1, 100, "")
		for _, share := range shares {
			if share.Name == testShareName {
				_ = svc.DeleteShare(share.ID)
				break
			}
		}
		_ = os.RemoveAll(testPath)
	})

	// 创建测试目录
	if err := os.MkdirAll(testPath, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	t.Run("CreateShareWithService", func(t *testing.T) {
		req := &dto.CreateSambaShareRequest{
			Name:    testShareName,
			Path:    testPath,
			Comment: "Service integration test share",
		}

		resp, err := svc.CreateShare(req)
		if err != nil {
			t.Fatalf("CreateShare failed: %v", err)
		}

		if resp.ID == "" {
			t.Fatalf("CreateShare returned empty ID")
		}

		t.Logf("Successfully created share via service: %s (ID: %s)", resp.Name, resp.ID)
	})

	t.Run("GetSharesFromService", func(t *testing.T) {
		shares, total, err := svc.GetShares(1, 10, "")
		if err != nil {
			t.Fatalf("GetShares failed: %v", err)
		}

		if total == 0 {
			t.Fatal("Expected at least one share")
		}

		found := false
		for _, share := range shares {
			if share.Name == testShareName {
				found = true
				t.Logf("Found share in service: %s -> %s", share.Name, share.Path)
				break
			}
		}

		if !found {
			t.Fatalf("Test share %s not found in service response", testShareName)
		}

		t.Logf("Service returned %d shares (total: %d)", len(shares), total)
	})

	t.Run("VerifyShareInSambaConfig", func(t *testing.T) {
		// 直接从 Samba 验证共享存在
		shares, err := cli.ListShares()
		if err != nil {
			t.Fatalf("ListShares failed: %v", err)
		}

		found := false
		for _, share := range shares {
			if share.Name == testShareName {
				found = true
				t.Logf("Verified share in Samba config: %s -> %s", share.Name, share.Path)
				break
			}
		}

		if !found {
			t.Fatalf("Test share %s not found in Samba configuration", testShareName)
		}
	})
}