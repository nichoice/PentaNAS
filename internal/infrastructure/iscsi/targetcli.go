package iscsi

import (
	"fmt"
	"runtime"
	"strings"
	"time"

	"pnas/internal/utils"
)

// CommandRunner defines the function signature used to execute shell commands.
type CommandRunner func(opts utils.ExecOptions, command string, args ...string) *utils.CmdResult

// Client exposes the subset of targetcli operations required by the services layer.
type Client interface {
	CreateTarget(iqn string) error
	DeleteTarget(iqn string) error
	EnsurePortal(iqn, ip string, port int) error
	EnableTarget(iqn string) error
	DisableTarget(iqn string) error

	// Backstore operations
	CreateFileBackstore(name, path string, size int64) error
	CreateBlockBackstore(name, device string) error
	DeleteBackstore(backendType, name string) error

	// LUN operations
	CreateLUN(iqn string, lun int, backstoreName, backstoreType string) error
	DeleteLUN(iqn string, lun int) error
}

const (
	defaultBinary  = "targetcli"
	defaultTimeout = 10 * time.Second
)

// TargetCLI executes targetcli commands on Linux hosts.
type TargetCLI struct {
	runner  CommandRunner
	binary  string
	isLinux func() bool
}

// NewTargetCLI constructs a TargetCLI backed by utils.ExecCommand.
func NewTargetCLI() *TargetCLI {
	return &TargetCLI{
		runner:  utils.ExecCommand,
		binary:  defaultBinary,
		isLinux: func() bool { return runtime.GOOS == "linux" },
	}
}

// WithRunner overrides the command runner, primarily for testing.
func (c *TargetCLI) WithRunner(runner CommandRunner) *TargetCLI {
	c.runner = runner
	return c
}

// WithPlatformDetector overrides the platform check used by ensureLinux, primarily for testing.
func (c *TargetCLI) WithPlatformDetector(detector func() bool) *TargetCLI {
	c.isLinux = detector
	return c
}

func (c *TargetCLI) exec(args ...string) *utils.CmdResult {
	return c.runner(utils.ExecOptions{Timeout: defaultTimeout}, c.binary, args...)
}

func (c *TargetCLI) ensureLinux() error {
	checker := c.isLinux
	if checker == nil {
		checker = func() bool { return runtime.GOOS == "linux" }
	}
	if !checker() {
		return fmt.Errorf("targetcli operations require a Linux host")
	}
	return nil
}

// CreateTarget creates an IQN under /iscsi. If the target already exists the call is treated as success.
func (c *TargetCLI) CreateTarget(iqn string) error {
	if err := c.ensureLinux(); err != nil {
		return err
	}
	res := c.exec("/iscsi", "create", iqn)
	if res.Error != nil {
		if strings.Contains(res.Stderr, "exists") {
			return nil
		}
		return fmt.Errorf("targetcli create target failed: %w; stderr: %s", res.Error, res.Stderr)
	}
	return nil
}

// DeleteTarget removes an IQN from targetcli, ignoring "not found" errors.
func (c *TargetCLI) DeleteTarget(iqn string) error {
	if err := c.ensureLinux(); err != nil {
		return err
	}
	res := c.exec("/iscsi", "delete", iqn)
	if res.Error != nil {
		if strings.Contains(res.Stderr, "No such path") || strings.Contains(res.Stderr, "not found") {
			return nil
		}
		return fmt.Errorf("targetcli delete target failed: %w; stderr: %s", res.Error, res.Stderr)
	}
	return nil
}

// EnsurePortal creates a network portal if it does not already exist.
func (c *TargetCLI) EnsurePortal(iqn, ip string, port int) error {
	if err := c.ensureLinux(); err != nil {
		return err
	}
	path := fmt.Sprintf("/iscsi/%s/tpg1/portals", iqn)
	res := c.exec(path, "create", ip, fmt.Sprintf("%d", port))
	if res.Error != nil {
		if strings.Contains(res.Stderr, "already exists") {
			return nil
		}
		return fmt.Errorf("targetcli create portal failed: %w; stderr: %s", res.Error, res.Stderr)
	}
	return nil
}

// EnableTarget enables the default TPG so the target becomes discoverable.
func (c *TargetCLI) EnableTarget(iqn string) error {
	if err := c.ensureLinux(); err != nil {
		return err
	}
	res := c.exec(fmt.Sprintf("/iscsi/%s/tpg1", iqn), "enable")
	if res.Error != nil {
		return fmt.Errorf("targetcli enable target failed: %w; stderr: %s", res.Error, res.Stderr)
	}
	return nil
}

// DisableTarget disables the default TPG, effectively stopping the target service.
func (c *TargetCLI) DisableTarget(iqn string) error {
	if err := c.ensureLinux(); err != nil {
		return err
	}
	res := c.exec(fmt.Sprintf("/iscsi/%s/tpg1", iqn), "disable")
	if res.Error != nil {
		if strings.Contains(res.Stderr, "already disabled") {
			return nil
		}
		return fmt.Errorf("targetcli disable target failed: %w; stderr: %s", res.Error, res.Stderr)
	}
	return nil
}

// CreateFileBackstore creates a file-based backstore for LUN usage.
func (c *TargetCLI) CreateFileBackstore(name, path string, size int64) error {
	if err := c.ensureLinux(); err != nil {
		return err
	}
	// targetcli /backstores/fileio create name=<name> file_or_dev=<path> size=<size>
	res := c.exec("/backstores/fileio", "create", fmt.Sprintf("name=%s", name),
		fmt.Sprintf("file_or_dev=%s", path), fmt.Sprintf("size=%d", size))
	if res.Error != nil {
		if strings.Contains(res.Stderr, "already exists") {
			return nil
		}
		return fmt.Errorf("targetcli create file backstore failed: %w; stderr: %s", res.Error, res.Stderr)
	}
	return nil
}

// CreateBlockBackstore creates a block device-based backstore.
func (c *TargetCLI) CreateBlockBackstore(name, device string) error {
	if err := c.ensureLinux(); err != nil {
		return err
	}
	// targetcli /backstores/block create name=<name> dev=<device>
	res := c.exec("/backstores/block", "create", fmt.Sprintf("name=%s", name),
		fmt.Sprintf("dev=%s", device))
	if res.Error != nil {
		if strings.Contains(res.Stderr, "already exists") {
			return nil
		}
		return fmt.Errorf("targetcli create block backstore failed: %w; stderr: %s", res.Error, res.Stderr)
	}
	return nil
}

// DeleteBackstore removes a backstore of specified type.
func (c *TargetCLI) DeleteBackstore(backendType, name string) error {
	if err := c.ensureLinux(); err != nil {
		return err
	}
	// targetcli /backstores/<type> delete <name>
	path := fmt.Sprintf("/backstores/%s", backendType)
	res := c.exec(path, "delete", name)
	if res.Error != nil {
		if strings.Contains(res.Stderr, "No such path") || strings.Contains(res.Stderr, "not found") {
			return nil
		}
		return fmt.Errorf("targetcli delete backstore failed: %w; stderr: %s", res.Error, res.Stderr)
	}
	return nil
}

// CreateLUN maps a backstore to a LUN under the specified target.
func (c *TargetCLI) CreateLUN(iqn string, lun int, backstoreName, backstoreType string) error {
	if err := c.ensureLinux(); err != nil {
		return err
	}
	// targetcli /iscsi/<iqn>/tpg1/luns create /backstores/<type>/<backstoreName> <lun>
	lunPath := fmt.Sprintf("/iscsi/%s/tpg1/luns", iqn)
	backstorePath := fmt.Sprintf("/backstores/%s/%s", backstoreType, backstoreName)
	res := c.exec(lunPath, "create", backstorePath, fmt.Sprintf("%d", lun))
	if res.Error != nil {
		if strings.Contains(res.Stderr, "already exists") {
			return nil
		}
		return fmt.Errorf("targetcli create LUN failed: %w; stderr: %s", res.Error, res.Stderr)
	}
	return nil
}

// DeleteLUN removes a LUN from the specified target.
func (c *TargetCLI) DeleteLUN(iqn string, lun int) error {
	if err := c.ensureLinux(); err != nil {
		return err
	}
	// targetcli /iscsi/<iqn>/tpg1/luns delete <lun>
	lunPath := fmt.Sprintf("/iscsi/%s/tpg1/luns", iqn)
	res := c.exec(lunPath, "delete", fmt.Sprintf("%d", lun))
	if res.Error != nil {
		if strings.Contains(res.Stderr, "No such path") || strings.Contains(res.Stderr, "not found") {
			return nil
		}
		return fmt.Errorf("targetcli delete LUN failed: %w; stderr: %s", res.Error, res.Stderr)
	}
	return nil
}
