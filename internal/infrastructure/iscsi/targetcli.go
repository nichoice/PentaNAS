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
}

const (
	defaultBinary  = "targetcli"
	defaultTimeout = 10 * time.Second
)

// TargetCLI executes targetcli commands on Linux hosts.
type TargetCLI struct {
	runner CommandRunner
	binary string
}

// NewTargetCLI constructs a TargetCLI backed by utils.ExecCommand.
func NewTargetCLI() *TargetCLI {
	return &TargetCLI{
		runner: utils.ExecCommand,
		binary: defaultBinary,
	}
}

// WithRunner overrides the command runner, primarily for testing.
func (c *TargetCLI) WithRunner(runner CommandRunner) *TargetCLI {
	c.runner = runner
	return c
}

func (c *TargetCLI) exec(args ...string) *utils.CmdResult {
	return c.runner(utils.ExecOptions{Timeout: defaultTimeout}, c.binary, args...)
}

func (c *TargetCLI) ensureLinux() error {
	if runtime.GOOS != "linux" {
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
