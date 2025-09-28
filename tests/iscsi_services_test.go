package main

import (
	"fmt"
	"sync"
	"testing"

	"pnas/internal/app/dto"
	"pnas/internal/app/services"
	"pnas/internal/database"
	"pnas/internal/infrastructure/iscsi"
	"pnas/internal/models"
	"pnas/internal/utils"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type commandCall struct {
	command string
	args    []string
}

type recordingRunner struct {
	mu    sync.Mutex
	calls []commandCall
}

func newRecordingRunner() *recordingRunner {
	return &recordingRunner{}
}

func (r *recordingRunner) Run(_ utils.ExecOptions, command string, args ...string) *utils.CmdResult {
	r.mu.Lock()
	defer r.mu.Unlock()
	copyArgs := append([]string(nil), args...)
	r.calls = append(r.calls, commandCall{command: command, args: copyArgs})
	return &utils.CmdResult{}
}

func (r *recordingRunner) Snapshot() []commandCall {
	r.mu.Lock()
	defer r.mu.Unlock()
	copyCalls := make([]commandCall, len(r.calls))
	copy(copyCalls, r.calls)
	return copyCalls
}

func (r *recordingRunner) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.calls = nil
}

func newTestTargetCLI() (*iscsi.TargetCLI, *recordingRunner) {
	runner := newRecordingRunner()
	cli := iscsi.NewTargetCLI().
		WithRunner(runner.Run).
		WithPlatformDetector(func() bool { return true })
	return cli, runner
}

func expectArgs(t *testing.T, call commandCall, expected []string) {
	t.Helper()
	if len(call.args) != len(expected) {
		t.Fatalf("expected args %v, got %v", expected, call.args)
	}
	for i, arg := range expected {
		if call.args[i] != arg {
			t.Fatalf("expected arg[%d] = %q, got %q", i, arg, call.args[i])
		}
	}
}

func portalPath(iqn string) string {
	return fmt.Sprintf("/iscsi/%s/tpg1/portals", iqn)
}

func tpgPath(iqn string) string {
	return fmt.Sprintf("/iscsi/%s/tpg1", iqn)
}

func setupISCSITestDB(t *testing.T) *gorm.DB {
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

func createTestTarget(t *testing.T, svc *services.ISCSITargetService, name string) *dto.ISCSITargetResponse {
	t.Helper()

	req := &dto.CreateISCSITargetRequest{
		Name:      name,
		Alias:     "primary",
		Comment:   "test target",
		IsEnabled: true,
	}

	resp, err := svc.CreateTarget(req)
	if err != nil {
		t.Fatalf("CreateTarget returned error: %v", err)
	}

	if resp.ID == "" {
		t.Fatalf("CreateTarget returned empty ID")
	}

	return resp
}

func TestISCSITargetService_CreateAndFetch(t *testing.T) {
	db := setupISCSITestDB(t)
	cli, runner := newTestTargetCLI()
	svc := services.NewISCSITargetServiceWithDeps(db, cli)

	created := createTestTarget(t, svc, "iqn.2024-01.com.pnas:test1")

	if created.Status != models.ISCSIStatusActive {
		t.Fatalf("expected created target status %s, got %s", models.ISCSIStatusActive, created.Status)
	}

	calls := runner.Snapshot()
	if len(calls) != 3 {
		t.Fatalf("expected 3 targetcli commands, got %d", len(calls))
	}

	expectArgs(t, calls[0], []string{"/iscsi", "create", created.Name})
	expectArgs(t, calls[1], []string{portalPath(created.Name), "create", "0.0.0.0", "3260"})
	expectArgs(t, calls[2], []string{tpgPath(created.Name), "enable"})

	runner.Reset()

	targets, total, err := svc.GetTargets(1, 10, "")
	if err != nil {
		t.Fatalf("GetTargets returned error: %v", err)
	}

	if total != 1 {
		t.Fatalf("expected total 1, got %d", total)
	}

	if len(targets) != 1 {
		t.Fatalf("expected 1 target, got %d", len(targets))
	}

	if targets[0].Name != created.Name {
		t.Fatalf("expected target name %s, got %s", created.Name, targets[0].Name)
	}

	fetched, err := svc.GetTargetByID(created.ID)
	if err != nil {
		t.Fatalf("GetTargetByID returned error: %v", err)
	}

	if fetched.ID != created.ID {
		t.Fatalf("expected ID %s, got %s", created.ID, fetched.ID)
	}
}

func TestISCSITargetService_UpdateAndDelete(t *testing.T) {
	db := setupISCSITestDB(t)
	cli, runner := newTestTargetCLI()
	svc := services.NewISCSITargetServiceWithDeps(db, cli)

	created := createTestTarget(t, svc, "iqn.2024-01.com.pnas:test2")
	if len(runner.Snapshot()) != 3 {
		t.Fatalf("expected initial target creation to issue 3 commands")
	}
	runner.Reset()

	newAlias := "updated"
	disabled := false
	req := &dto.UpdateISCSITargetRequest{
		Alias:     &newAlias,
		IsEnabled: &disabled,
	}

	updated, err := svc.UpdateTarget(created.ID, req)
	if err != nil {
		t.Fatalf("UpdateTarget returned error: %v", err)
	}

	if updated.Alias != newAlias {
		t.Fatalf("expected alias %s, got %s", newAlias, updated.Alias)
	}

	if updated.IsEnabled != disabled {
		t.Fatalf("expected isEnabled %v, got %v", disabled, updated.IsEnabled)
	}

	updateCalls := runner.Snapshot()
	if len(updateCalls) != 1 {
		t.Fatalf("expected 1 targetcli command during update, got %d", len(updateCalls))
	}
	expectArgs(t, updateCalls[0], []string{tpgPath(created.Name), "disable"})
	runner.Reset()

	if err := svc.DeleteTarget(created.ID); err != nil {
		t.Fatalf("DeleteTarget returned error: %v", err)
	}

	deleteCalls := runner.Snapshot()
	if len(deleteCalls) != 2 {
		t.Fatalf("expected 2 targetcli commands during delete, got %d", len(deleteCalls))
	}
	expectArgs(t, deleteCalls[0], []string{tpgPath(created.Name), "disable"})
	expectArgs(t, deleteCalls[1], []string{"/iscsi", "delete", created.Name})

	if _, err := svc.GetTargetByID(created.ID); err == nil {
		t.Fatalf("expected error when fetching deleted target")
	}
}

func TestISCSITargetService_StartAndStop(t *testing.T) {
	db := setupISCSITestDB(t)
	cli, runner := newTestTargetCLI()
	svc := services.NewISCSITargetServiceWithDeps(db, cli)

	created := createTestTarget(t, svc, "iqn.2024-01.com.pnas:test3")
	if len(runner.Snapshot()) != 3 {
		t.Fatalf("expected initial target creation to issue 3 commands")
	}
	runner.Reset()

	if err := svc.StartTarget(created.ID); err != nil {
		t.Fatalf("StartTarget returned error: %v", err)
	}
	startCalls := runner.Snapshot()
	if len(startCalls) != 1 {
		t.Fatalf("expected 1 targetcli command during start, got %d", len(startCalls))
	}
	expectArgs(t, startCalls[0], []string{tpgPath(created.Name), "enable"})
	runner.Reset()

	started, err := svc.GetTargetByID(created.ID)
	if err != nil {
		t.Fatalf("GetTargetByID returned error: %v", err)
	}

	if started.Status != models.ISCSIStatusActive {
		t.Fatalf("expected status %s, got %s", models.ISCSIStatusActive, started.Status)
	}

	if err := svc.StopTarget(created.ID); err != nil {
		t.Fatalf("StopTarget returned error: %v", err)
	}
	stopCalls := runner.Snapshot()
	if len(stopCalls) != 1 {
		t.Fatalf("expected 1 targetcli command during stop, got %d", len(stopCalls))
	}
	expectArgs(t, stopCalls[0], []string{tpgPath(created.Name), "disable"})

	stopped, err := svc.GetTargetByID(created.ID)
	if err != nil {
		t.Fatalf("GetTargetByID returned error: %v", err)
	}

	if stopped.Status != models.ISCSIStatusInactive {
		t.Fatalf("expected status %s, got %s", models.ISCSIStatusInactive, stopped.Status)
	}

}
