package integration

import (
	"testing"

	"pnas/internal/app/dto"
	"pnas/internal/app/services"
	"pnas/internal/database"
	"pnas/internal/infrastructure/iscsi"
	"pnas/internal/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)


func setupIntegrationTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	database.DB = db
	models.InitISCSIModels(db)

	t.Cleanup(func() {
		sqlDB, err := db.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
		database.DB = nil
	})

	return db
}

// TestISCSITargetService_Integration 在 Linux 环境中测试 iSCSI Target 创建
func TestISCSITargetService_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db := setupIntegrationTestDB(t)

	// 使用真实的 TargetCLI，不注入模拟的 runner
	cli := iscsi.NewTargetCLI()
	svc := services.NewISCSITargetServiceWithDeps(db, cli)

	testIQN := "iqn.2024-01.com.pnas:integration-test"

	t.Run("CreateTarget", func(t *testing.T) {
		req := &dto.CreateISCSITargetRequest{
			Name:      testIQN,
			Alias:     "integration-test",
			Comment:   "Integration test target",
			IsEnabled: true,
		}

		resp, err := svc.CreateTarget(req)
		if err != nil {
			t.Fatalf("CreateTarget failed: %v", err)
		}

		if resp.ID == "" {
			t.Fatalf("CreateTarget returned empty ID")
		}

		if resp.Status != models.ISCSIStatusActive {
			t.Fatalf("expected status %s, got %s", models.ISCSIStatusActive, resp.Status)
		}

		t.Logf("Successfully created target: %s (ID: %s)", resp.Name, resp.ID)
	})

	t.Run("VerifyTargetInSystem", func(t *testing.T) {
		// 在此处可以添加系统级验证
		// 例如检查 /sys/kernel/config/target/iscsi/ 目录
		// 或者使用 targetcli ls 命令验证
		t.Log("Target verification should be done manually or with system commands")
	})

	t.Run("StopTarget", func(t *testing.T) {
		targets, _, err := svc.GetTargets(1, 10, "")
		if err != nil {
			t.Fatalf("GetTargets failed: %v", err)
		}

		if len(targets) == 0 {
			t.Fatal("No targets found")
		}

		targetID := targets[0].ID
		err = svc.StopTarget(targetID)
		if err != nil {
			t.Fatalf("StopTarget failed: %v", err)
		}

		t.Logf("Successfully stopped target: %s", targetID)
	})

	t.Run("DeleteTarget", func(t *testing.T) {
		targets, _, err := svc.GetTargets(1, 10, "")
		if err != nil {
			t.Fatalf("GetTargets failed: %v", err)
		}

		if len(targets) == 0 {
			t.Fatal("No targets found")
		}

		targetID := targets[0].ID
		err = svc.DeleteTarget(targetID)
		if err != nil {
			t.Fatalf("DeleteTarget failed: %v", err)
		}

		t.Logf("Successfully deleted target: %s", targetID)
	})
}

// TestISCSITargetService_SystemValidation 测试系统级别的验证
func TestISCSITargetService_SystemValidation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping system validation test in short mode")
	}

	db := setupIntegrationTestDB(t)
	cli := iscsi.NewTargetCLI()
	svc := services.NewISCSITargetServiceWithDeps(db, cli)

	testIQN := "iqn.2024-01.com.pnas:system-validation"

	// 创建目标
	req := &dto.CreateISCSITargetRequest{
		Name:      testIQN,
		Alias:     "system-validation",
		Comment:   "System validation test",
		IsEnabled: true,
	}

	resp, err := svc.CreateTarget(req)
	if err != nil {
		t.Fatalf("CreateTarget failed: %v", err)
	}

	// 清理函数
	t.Cleanup(func() {
		_ = svc.DeleteTarget(resp.ID)
	})

	t.Run("VerifyTargetExists", func(t *testing.T) {
		// 这里可以添加系统命令来验证目标是否真实存在
		// 例如：targetcli ls /iscsi
		t.Logf("Target created with ID: %s, Name: %s", resp.ID, resp.Name)
		t.Log("Manual verification needed: run 'sudo targetcli ls /iscsi' to verify target exists")
	})
}

// TestISCSILUNService_Integration 测试 LUN 和 backstore 功能
func TestISCSILUNService_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db := setupIntegrationTestDB(t)

	// 使用真实的 TargetCLI
	cli := iscsi.NewTargetCLI()
	targetSvc := services.NewISCSITargetServiceWithDeps(db, cli)
	lunSvc := services.NewISCSILUNServiceWithDeps(db, cli)

	testIQN := "iqn.2024-01.com.pnas:lun-test"

	// 首先创建一个 Target
	targetReq := &dto.CreateISCSITargetRequest{
		Name:      testIQN,
		Alias:     "lun-test",
		Comment:   "LUN integration test target",
		IsEnabled: true,
	}

	targetResp, err := targetSvc.CreateTarget(targetReq)
	if err != nil {
		t.Fatalf("CreateTarget failed: %v", err)
	}

	// 清理函数
	t.Cleanup(func() {
		_ = targetSvc.DeleteTarget(targetResp.ID)
	})

	t.Run("CreateFileLUN", func(t *testing.T) {
		lunReq := &dto.CreateISCSILUNRequest{
			Name:       "test-file-lun",
			DeviceType: models.ISCSIDeviceFile,
			Size:       100 * 1024 * 1024, // 100MB
			Comment:    "Test file-based LUN",
			IsEnabled:  true,
			BlockSize:  512,
		}

		lunResp, err := lunSvc.CreateLUN(lunReq)
		if err != nil {
			t.Fatalf("CreateLUN failed: %v", err)
		}

		if lunResp.ID == "" {
			t.Fatalf("CreateLUN returned empty ID")
		}

		t.Logf("Successfully created file LUN: %s (ID: %s)", lunResp.Name, lunResp.ID)

		// 映射 LUN 到 Target
		mappingResp, err := lunSvc.MapLUNToTarget(lunResp.ID, targetResp.ID, 0)
		if err != nil {
			t.Fatalf("MapLUNToTarget failed: %v", err)
		}

		t.Logf("Successfully mapped LUN to target: LUN %d", mappingResp.LUN)
	})

	t.Run("CreateBlockLUN", func(t *testing.T) {
		// 注意：这个测试需要真实的块设备
		// 在实际测试环境中，你可能需要创建一个 loop 设备
		t.Skip("Block device test requires real block device - enable manually with valid device path")

		lunReq := &dto.CreateISCSILUNRequest{
			Name:       "test-block-lun",
			DeviceType: models.ISCSIDeviceBlock,
			DevicePath: "/dev/loop0", // 需要替换为真实的块设备路径
			Comment:    "Test block-based LUN",
			IsEnabled:  true,
			BlockSize:  4096,
		}

		lunResp, err := lunSvc.CreateLUN(lunReq)
		if err != nil {
			t.Fatalf("CreateLUN failed: %v", err)
		}

		t.Logf("Successfully created block LUN: %s (ID: %s)", lunResp.Name, lunResp.ID)

		// 映射到现有 Target
		mappingResp, err := lunSvc.MapLUNToTarget(lunResp.ID, targetResp.ID, 1)
		if err != nil {
			t.Fatalf("MapLUNToTarget failed: %v", err)
		}

		t.Logf("Successfully mapped block LUN to target: LUN %d", mappingResp.LUN)
	})

	t.Run("VerifyTargetWithLUN", func(t *testing.T) {
		t.Log("Target with LUN created. Manual verification:")
		t.Logf("  sudo targetcli ls /iscsi/%s/tpg1/luns", testIQN)
		t.Log("  sudo targetcli ls /backstores/fileio")
		t.Log("  sudo targetcli ls /backstores/block")
		t.Log("Expected: File LUN mapped to LUN 0")
	})
}
