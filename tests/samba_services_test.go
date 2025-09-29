package main

import (
	"fmt"
	"sync"
	"testing"

	"pnas/internal/app/dto"
	"pnas/internal/app/services"
	"pnas/internal/database"
	"pnas/internal/infrastructure/samba"
	"pnas/internal/models"
	"pnas/internal/utils"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type sambaCommandCall struct {
	command string
	args    []string
}

type recordingSambaRunner struct {
	mu    sync.Mutex
	calls []sambaCommandCall
}

func newRecordingSambaRunner() *recordingSambaRunner {
	return &recordingSambaRunner{}
}

func (r *recordingSambaRunner) Run(_ utils.ExecOptions, command string, args ...string) *utils.CmdResult {
	r.mu.Lock()
	defer r.mu.Unlock()
	copyArgs := append([]string(nil), args...)
	r.calls = append(r.calls, sambaCommandCall{command: command, args: copyArgs})

	// Return appropriate mock responses based on command
	result := &utils.CmdResult{}

	switch command {
	case "testparm":
		if len(args) > 0 && args[0] == "-s" && args[1] == "--section-name" {
			result.Stdout = "testshare\nanothershare\n"
		} else if len(args) > 0 && args[0] == "-s" {
			result.Stdout = "Load smb config files from /etc/samba/smb.conf\nProcessed services file successfully.\n"
		}
	case "pdbedit":
		if len(args) > 0 && args[0] == "-L" {
			result.Stdout = "testuser:1001:Test User\nanotheruser:1002:Another User\n"
		}
	case "systemctl":
		if len(args) >= 2 && args[0] == "is-active" && args[1] == "smbd" {
			result.Stdout = "active\n"
		} else if len(args) >= 3 && args[0] == "show" && args[1] == "smbd" {
			result.Stdout = "1234\n"
		}
	case "smbd":
		if len(args) > 0 && args[0] == "--version" {
			result.Stdout = "Version 4.18.6\n"
		}
	case "smbstatus":
		if len(args) > 0 && args[0] == "-p" {
			result.Stdout = "PID     Username      Group         Machine            Protocol Version  Encryption           Signing\n---------------------------------------------------------------------------------------------------------------------------\n1234    testuser      testgroup     192.168.1.100      SMB3_11           -                    partial(AES-128-CMAC)\n"
		}
	case "id":
		// User exists check
		result.Stdout = "uid=1001(testuser) gid=1001(testuser) groups=1001(testuser)\n"
	}

	return result
}

func (r *recordingSambaRunner) Snapshot() []sambaCommandCall {
	r.mu.Lock()
	defer r.mu.Unlock()
	copyCalls := make([]sambaCommandCall, len(r.calls))
	copy(copyCalls, r.calls)
	return copyCalls
}

func (r *recordingSambaRunner) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.calls = nil
}

func newTestSambaControl() (*samba.SambaControl, *recordingSambaRunner) {
	runner := newRecordingSambaRunner()
	cli := samba.NewSambaControl().
		WithRunner(runner.Run).
		WithConfigPath("/tmp/test_smb.conf")
	return cli, runner
}

func expectSambaArgs(t *testing.T, call sambaCommandCall, expectedCommand string, expectedArgs []string) {
	t.Helper()
	if call.command != expectedCommand {
		t.Fatalf("expected command %s, got %s", expectedCommand, call.command)
	}
	if len(call.args) != len(expectedArgs) {
		t.Fatalf("expected args %v, got %v", expectedArgs, call.args)
	}
	for i, arg := range expectedArgs {
		if call.args[i] != arg {
			t.Fatalf("expected arg[%d] = %q, got %q", i, arg, call.args[i])
		}
	}
}

func setupSambaTestDB(t *testing.T) *gorm.DB {
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

func createTestShare(t *testing.T, svc *services.SambaShareService, name string) *dto.SambaShareResponse {
	t.Helper()

	req := &dto.CreateSambaShareRequest{
		Name:    name,
		Path:    "/tmp/testshare",
		Comment: "Test share",
	}

	resp, err := svc.CreateShare(req)
	if err != nil {
		t.Fatalf("CreateShare returned error: %v", err)
	}

	if resp.ID == "" {
		t.Fatalf("CreateShare returned empty ID")
	}

	return resp
}

// TestSambaControl_ShareManagement 测试 Samba 共享管理
func TestSambaControl_ShareManagement(t *testing.T) {
	cli, runner := newTestSambaControl()

	t.Run("CreateShare", func(t *testing.T) {
		runner.Reset()

		options := map[string]string{
			"writable":    "yes",
			"browseable":  "yes",
			"guest ok":    "no",
		}

		err := cli.CreateShare("testshare", "/tmp/testshare", "Test Share", options)
		if err != nil {
			t.Fatalf("CreateShare failed: %v", err)
		}

		calls := runner.Snapshot()
		// Should backup config, append share config, validate, and reload
		if len(calls) < 3 {
			t.Fatalf("expected at least 3 commands, got %d", len(calls))
		}

		// Check that testparm is called for validation
		foundValidation := false
		foundReload := false
		for _, call := range calls {
			if call.command == "testparm" && len(call.args) > 0 && call.args[0] == "-s" {
				foundValidation = true
			}
			if call.command == "smbcontrol" && len(call.args) >= 3 && call.args[0] == "all" && call.args[1] == "reload-config" {
				foundReload = true
			}
		}

		if !foundValidation {
			t.Error("Expected testparm validation command")
		}
		if !foundReload {
			t.Error("Expected smbcontrol reload command")
		}
	})

	t.Run("ListShares", func(t *testing.T) {
		runner.Reset()

		shares, err := cli.ListShares()
		if err != nil {
			t.Fatalf("ListShares failed: %v", err)
		}

		calls := runner.Snapshot()
		if len(calls) < 1 {
			t.Fatalf("expected at least 1 command, got %d", len(calls))
		}

		expectSambaArgs(t, calls[0], "testparm", []string{"-s", "--section-name"})

		// Should have found the mocked shares
		if len(shares) < 2 {
			t.Fatalf("expected at least 2 shares, got %d", len(shares))
		}

		t.Logf("Found shares: %+v", shares)
	})

	t.Run("DeleteShare", func(t *testing.T) {
		runner.Reset()

		err := cli.DeleteShare("testshare")
		if err != nil {
			t.Fatalf("DeleteShare failed: %v", err)
		}

		calls := runner.Snapshot()
		// Should backup config, remove share section, and reload
		if len(calls) < 3 {
			t.Fatalf("expected at least 3 commands, got %d", len(calls))
		}

		// Check for sed command to remove share
		foundSed := false
		for _, call := range calls {
			if call.command == "sed" {
				foundSed = true
				break
			}
		}

		if !foundSed {
			t.Error("Expected sed command to remove share")
		}
	})
}

// TestSambaControl_UserManagement 测试 Samba 用户管理
func TestSambaControl_UserManagement(t *testing.T) {
	cli, runner := newTestSambaControl()

	t.Run("CreateUser", func(t *testing.T) {
		runner.Reset()

		err := cli.CreateUser("testuser", "testpass")
		if err != nil {
			t.Fatalf("CreateUser failed: %v", err)
		}

		calls := runner.Snapshot()
		if len(calls) < 2 {
			t.Fatalf("expected at least 2 commands, got %d", len(calls))
		}

		// Should check if user exists and add to samba
		foundUserCheck := false
		foundSambaAdd := false
		for _, call := range calls {
			if call.command == "id" && len(call.args) == 1 && call.args[0] == "testuser" {
				foundUserCheck = true
			}
			if call.command == "sh" && len(call.args) >= 2 && call.args[0] == "-c" {
				if contains(call.args[1], "smbpasswd") && contains(call.args[1], "-a") {
					foundSambaAdd = true
				}
			}
		}

		if !foundUserCheck {
			t.Error("Expected user existence check")
		}
		if !foundSambaAdd {
			t.Error("Expected smbpasswd add command")
		}
	})

	t.Run("ListUsers", func(t *testing.T) {
		runner.Reset()

		users, err := cli.ListUsers()
		if err != nil {
			t.Fatalf("ListUsers failed: %v", err)
		}

		calls := runner.Snapshot()
		if len(calls) < 1 {
			t.Fatalf("expected at least 1 command, got %d", len(calls))
		}

		expectSambaArgs(t, calls[0], "pdbedit", []string{"-L"})

		// Should have found the mocked users
		if len(users) < 2 {
			t.Fatalf("expected at least 2 users, got %d", len(users))
		}

		t.Logf("Found users: %+v", users)
	})

	t.Run("SetUserPassword", func(t *testing.T) {
		runner.Reset()

		err := cli.SetUserPassword("testuser", "newpass")
		if err != nil {
			t.Fatalf("SetUserPassword failed: %v", err)
		}

		calls := runner.Snapshot()
		if len(calls) < 1 {
			t.Fatalf("expected at least 1 command, got %d", len(calls))
		}

		// Should use smbpasswd to set password
		foundPasswordSet := false
		for _, call := range calls {
			if call.command == "sh" && len(call.args) >= 2 && call.args[0] == "-c" {
				if contains(call.args[1], "smbpasswd") && contains(call.args[1], "testuser") {
					foundPasswordSet = true
				}
			}
		}

		if !foundPasswordSet {
			t.Error("Expected smbpasswd password set command")
		}
	})

	t.Run("DeleteUser", func(t *testing.T) {
		runner.Reset()

		err := cli.DeleteUser("testuser")
		if err != nil {
			t.Fatalf("DeleteUser failed: %v", err)
		}

		calls := runner.Snapshot()
		if len(calls) < 1 {
			t.Fatalf("expected at least 1 command, got %d", len(calls))
		}

		expectSambaArgs(t, calls[0], "smbpasswd", []string{"-x", "testuser"})
	})
}

// TestSambaControl_ServiceManagement 测试 Samba 服务管理
func TestSambaControl_ServiceManagement(t *testing.T) {
	cli, runner := newTestSambaControl()

	t.Run("GetServiceStatus", func(t *testing.T) {
		runner.Reset()

		status, err := cli.GetServiceStatus()
		if err != nil {
			t.Fatalf("GetServiceStatus failed: %v", err)
		}

		if !status.IsRunning {
			t.Error("Expected service to be running")
		}

		if status.ProcessID == 0 {
			t.Error("Expected non-zero process ID")
		}

		if status.Version == "" {
			t.Error("Expected version information")
		}

		calls := runner.Snapshot()
		if len(calls) < 3 {
			t.Fatalf("expected at least 3 commands for status check, got %d", len(calls))
		}

		t.Logf("Service status: %+v", status)
	})

	t.Run("ReloadConfig", func(t *testing.T) {
		runner.Reset()

		err := cli.ReloadConfig()
		if err != nil {
			t.Fatalf("ReloadConfig failed: %v", err)
		}

		calls := runner.Snapshot()
		if len(calls) < 1 {
			t.Fatalf("expected at least 1 command, got %d", len(calls))
		}

		expectSambaArgs(t, calls[0], "smbcontrol", []string{"all", "reload-config"})
	})

	t.Run("RestartService", func(t *testing.T) {
		runner.Reset()

		err := cli.RestartService()
		if err != nil {
			t.Fatalf("RestartService failed: %v", err)
		}

		calls := runner.Snapshot()
		if len(calls) < 1 {
			t.Fatalf("expected at least 1 command, got %d", len(calls))
		}

		expectSambaArgs(t, calls[0], "systemctl", []string{"restart", "smbd"})
	})

	t.Run("ValidateConfig", func(t *testing.T) {
		runner.Reset()

		err := cli.ValidateConfig()
		if err != nil {
			t.Fatalf("ValidateConfig failed: %v", err)
		}

		calls := runner.Snapshot()
		if len(calls) < 1 {
			t.Fatalf("expected at least 1 command, got %d", len(calls))
		}

		expectSambaArgs(t, calls[0], "testparm", []string{"-s"})
	})
}

// TestSambaShareService_CreateAndManage 测试 Samba 共享服务
func TestSambaShareService_CreateAndManage(t *testing.T) {
	db := setupSambaTestDB(t)
	cli, runner := newTestSambaControl()
	svc := services.NewSambaShareServiceWithDeps(db, cli)

	t.Run("CreateShare", func(t *testing.T) {
		runner.Reset()

		req := &dto.CreateSambaShareRequest{
			Name:    "testshare",
			Path:    "/tmp/testshare",
			Comment: "Test share for unit testing",
		}

		resp, err := svc.CreateShare(req)
		if err != nil {
			t.Fatalf("CreateShare failed: %v", err)
		}

		if resp.ID == "" {
			t.Fatalf("CreateShare returned empty ID")
		}

		if resp.Name != req.Name {
			t.Fatalf("expected name %s, got %s", req.Name, resp.Name)
		}

		t.Logf("Successfully created share: %s (ID: %s)", resp.Name, resp.ID)

		// Should have called Samba commands to create the share
		calls := runner.Snapshot()
		if len(calls) == 0 {
			t.Error("Expected Samba commands to be called")
		}
	})

	t.Run("GetShares", func(t *testing.T) {
		shares, total, err := svc.GetShares(1, 10, "")
		if err != nil {
			t.Fatalf("GetShares failed: %v", err)
		}

		if total == 0 {
			t.Fatal("Expected at least one share")
		}

		if len(shares) == 0 {
			t.Fatal("Expected at least one share in response")
		}

		t.Logf("Found %d shares (total: %d)", len(shares), total)
	})
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
		(len(s) > len(substr) && s[len(s)-len(substr)-1:len(s)-len(substr)] == " " && s[len(s)-len(substr):] == substr) ||
		(len(s) > len(substr) && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}