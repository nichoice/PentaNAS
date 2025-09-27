package main

import (
	"testing"

	"pnas/internal/app/dto"
	"pnas/internal/app/services"
	"pnas/internal/database"
	"pnas/internal/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type fakeTargetCLI struct {
	created []string
	deleted []string
	enabled map[string]bool
}

func newFakeTargetCLI() *fakeTargetCLI {
	return &fakeTargetCLI{enabled: make(map[string]bool)}
}

func (f *fakeTargetCLI) CreateTarget(iqn string) error {
	f.created = append(f.created, iqn)
	return nil
}

func (f *fakeTargetCLI) DeleteTarget(iqn string) error {
	f.deleted = append(f.deleted, iqn)
	delete(f.enabled, iqn)
	return nil
}

func (f *fakeTargetCLI) EnsurePortal(iqn, ip string, port int) error {
	return nil
}

func (f *fakeTargetCLI) EnableTarget(iqn string) error {
	f.enabled[iqn] = true
	return nil
}

func (f *fakeTargetCLI) DisableTarget(iqn string) error {
	f.enabled[iqn] = false
	return nil
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
	cli := newFakeTargetCLI()
	svc := services.NewISCSITargetServiceWithDeps(db, cli)

	created := createTestTarget(t, svc, "iqn.2024-01.com.pnas:test1")

	if created.Status != models.ISCSIStatusActive {
		t.Fatalf("expected created target status %s, got %s", models.ISCSIStatusActive, created.Status)
	}

	if !cli.enabled[created.Name] {
		t.Fatalf("expected target %s to be enabled after creation", created.Name)
	}

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
	cli := newFakeTargetCLI()
	svc := services.NewISCSITargetServiceWithDeps(db, cli)

	created := createTestTarget(t, svc, "iqn.2024-01.com.pnas:test2")

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

	if cli.enabled[created.Name] {
		t.Fatalf("expected CLI target %s to be disabled after update", created.Name)
	}

	if err := svc.DeleteTarget(created.ID); err != nil {
		t.Fatalf("DeleteTarget returned error: %v", err)
	}

	if len(cli.deleted) != 1 || cli.deleted[0] != created.Name {
		t.Fatalf("expected target %s to be deleted in CLI", created.Name)
	}

	if _, err := svc.GetTargetByID(created.ID); err == nil {
		t.Fatalf("expected error when fetching deleted target")
	}
}

func TestISCSITargetService_StartAndStop(t *testing.T) {
	db := setupISCSITestDB(t)
	cli := newFakeTargetCLI()
	svc := services.NewISCSITargetServiceWithDeps(db, cli)

	created := createTestTarget(t, svc, "iqn.2024-01.com.pnas:test3")

	if err := svc.StartTarget(created.ID); err != nil {
		t.Fatalf("StartTarget returned error: %v", err)
	}

	started, err := svc.GetTargetByID(created.ID)
	if err != nil {
		t.Fatalf("GetTargetByID returned error: %v", err)
	}

	if started.Status != models.ISCSIStatusActive {
		t.Fatalf("expected status %s, got %s", models.ISCSIStatusActive, started.Status)
	}

	if !cli.enabled[created.Name] {
		t.Fatalf("expected target %s to be enabled in CLI", created.Name)
	}

	if err := svc.StopTarget(created.ID); err != nil {
		t.Fatalf("StopTarget returned error: %v", err)
	}

	stopped, err := svc.GetTargetByID(created.ID)
	if err != nil {
		t.Fatalf("GetTargetByID returned error: %v", err)
	}

	if stopped.Status != models.ISCSIStatusInactive {
		t.Fatalf("expected status %s, got %s", models.ISCSIStatusInactive, stopped.Status)
	}

	if cli.enabled[created.Name] {
		t.Fatalf("expected target %s to be disabled in CLI", created.Name)
	}
}
