package samba

import (
	"fmt"
	"os"
	"strings"
	"time"

	"pnas/internal/utils"
)

// CommandRunner defines the function signature used to execute shell commands.
type CommandRunner func(opts utils.ExecOptions, command string, args ...string) *utils.CmdResult

// Client exposes Samba management operations required by the services layer.
type Client interface {
	// Share management
	CreateShare(name, path, comment string, options map[string]string) error
	DeleteShare(name string) error
	EnableShare(name string) error
	DisableShare(name string) error
	ListShares() ([]ShareInfo, error)

	// User management
	CreateUser(username, password string) error
	DeleteUser(username string) error
	SetUserPassword(username, password string) error
	EnableUser(username string) error
	DisableUser(username string) error
	ListUsers() ([]UserInfo, error)

	// Service management
	ReloadConfig() error
	RestartService() error
	GetServiceStatus() (*ServiceStatus, error)

	// Configuration management
	BackupConfig() (string, error)
	RestoreConfig(backupPath string) error
	ValidateConfig() error
}

// ShareInfo represents information about a Samba share
type ShareInfo struct {
	Name        string            `json:"name"`
	Path        string            `json:"path"`
	Comment     string            `json:"comment"`
	IsEnabled   bool              `json:"is_enabled"`
	Options     map[string]string `json:"options"`
}

// UserInfo represents information about a Samba user
type UserInfo struct {
	Username    string    `json:"username"`
	IsEnabled   bool      `json:"is_enabled"`
	LastLogin   *time.Time `json:"last_login,omitempty"`
}

// ServiceStatus represents Samba service status
type ServiceStatus struct {
	IsRunning     bool      `json:"is_running"`
	ProcessID     int       `json:"process_id"`
	StartTime     *time.Time `json:"start_time,omitempty"`
	Version       string    `json:"version"`
	ActiveShares  int       `json:"active_shares"`
	ConnectedUsers int      `json:"connected_users"`
}

const (
	defaultTimeout = 30 * time.Second
	smbConfigPath  = "/etc/samba/smb.conf"
	smbBackupDir   = "/var/backups/samba"
)

// SambaControl manages Samba configuration and services on Linux hosts.
type SambaControl struct {
	runner     CommandRunner
	configPath string
	backupDir  string
}

// NewSambaControl constructs a SambaControl backed by utils.ExecCommand.
func NewSambaControl() *SambaControl {
	return &SambaControl{
		runner:     utils.ExecCommand,
		configPath: smbConfigPath,
		backupDir:  smbBackupDir,
	}
}

// WithRunner overrides the command runner, primarily for testing.
func (c *SambaControl) WithRunner(runner CommandRunner) *SambaControl {
	c.runner = runner
	return c
}

// WithConfigPath overrides the Samba configuration path.
func (c *SambaControl) WithConfigPath(path string) *SambaControl {
	c.configPath = path
	return c
}

func (c *SambaControl) exec(command string, args ...string) *utils.CmdResult {
	return c.runner(utils.ExecOptions{Timeout: defaultTimeout}, command, args...)
}

// CreateShare creates a new Samba share
func (c *SambaControl) CreateShare(name, path, comment string, options map[string]string) error {
	// Check if path exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("path does not exist: %s", path)
	}

	// Backup current config
	if err := c.backupCurrentConfig(); err != nil {
		return fmt.Errorf("failed to backup config: %w", err)
	}

	// Create share section
	shareConfig := fmt.Sprintf(`
[%s]
	path = %s
	comment = %s
	browseable = yes
	writable = yes
	guest ok = no
	create mask = 0664
	directory mask = 0775
`, name, path, comment)

	// Add custom options
	for key, value := range options {
		shareConfig += fmt.Sprintf("	%s = %s\n", key, value)
	}

	// Append to config file
	res := c.exec("sh", "-c", fmt.Sprintf("echo '%s' >> %s", shareConfig, c.configPath))
	if res.Error != nil {
		return fmt.Errorf("failed to add share to config: %w; stderr: %s", res.Error, res.Stderr)
	}

	// Validate configuration
	if err := c.ValidateConfig(); err != nil {
		// Restore backup if validation fails
		c.restoreLastBackup()
		return fmt.Errorf("share configuration is invalid: %w", err)
	}

	// Reload Samba configuration
	return c.ReloadConfig()
}

// DeleteShare removes a Samba share
func (c *SambaControl) DeleteShare(name string) error {
	// Backup current config
	if err := c.backupCurrentConfig(); err != nil {
		return fmt.Errorf("failed to backup config: %w", err)
	}

	// Remove share section using sed
	res := c.exec("sed", "-i", fmt.Sprintf("/^\\[%s\\]/,/^\\[.*\\]/{//!d; /^\\[%s\\]/d;}", name, name), c.configPath)
	if res.Error != nil {
		return fmt.Errorf("failed to remove share from config: %w; stderr: %s", res.Error, res.Stderr)
	}

	// Reload Samba configuration
	return c.ReloadConfig()
}

// EnableShare enables a Samba share (remove disabled directive)
func (c *SambaControl) EnableShare(name string) error {
	res := c.exec("sed", "-i", fmt.Sprintf("/^\\[%s\\]/,/^\\[.*\\]/{s/^.*available.*=.*no.*$/	available = yes/;}", name), c.configPath)
	if res.Error != nil {
		return fmt.Errorf("failed to enable share: %w; stderr: %s", res.Error, res.Stderr)
	}
	return c.ReloadConfig()
}

// DisableShare disables a Samba share
func (c *SambaControl) DisableShare(name string) error {
	res := c.exec("sed", "-i", fmt.Sprintf("/^\\[%s\\]/,/^\\[.*\\]/{/available.*=/d; /^\\[%s\\]/a\\	available = no", name, name), c.configPath)
	if res.Error != nil {
		return fmt.Errorf("failed to disable share: %w; stderr: %s", res.Error, res.Stderr)
	}
	return c.ReloadConfig()
}

// ListShares lists all configured Samba shares
func (c *SambaControl) ListShares() ([]ShareInfo, error) {
	res := c.exec("testparm", "-s", "--section-name")
	if res.Error != nil {
		return nil, fmt.Errorf("failed to list shares: %w; stderr: %s", res.Error, res.Stderr)
	}

	var shares []ShareInfo
	lines := strings.Split(strings.TrimSpace(res.Stdout), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && line != "global" && line != "homes" && line != "printers" {
			// Get share details
			shareInfo, err := c.getShareInfo(line)
			if err != nil {
				continue // Skip invalid shares
			}
			shares = append(shares, *shareInfo)
		}
	}

	return shares, nil
}

// CreateUser creates a new Samba user
func (c *SambaControl) CreateUser(username, password string) error {
	// Add user to system if not exists
	res := c.exec("id", username)
	if res.Error != nil {
		// User doesn't exist, create it
		res = c.exec("useradd", "-M", "-s", "/sbin/nologin", username)
		if res.Error != nil {
			return fmt.Errorf("failed to create system user: %w; stderr: %s", res.Error, res.Stderr)
		}
	}

	// Add user to Samba
	res = c.exec("sh", "-c", fmt.Sprintf("echo -e '%s\\n%s' | smbpasswd -a -s %s", password, password, username))
	if res.Error != nil {
		return fmt.Errorf("failed to create samba user: %w; stderr: %s", res.Error, res.Stderr)
	}

	return nil
}

// DeleteUser removes a Samba user
func (c *SambaControl) DeleteUser(username string) error {
	res := c.exec("smbpasswd", "-x", username)
	if res.Error != nil {
		return fmt.Errorf("failed to delete samba user: %w; stderr: %s", res.Error, res.Stderr)
	}
	return nil
}

// SetUserPassword sets password for a Samba user
func (c *SambaControl) SetUserPassword(username, password string) error {
	res := c.exec("sh", "-c", fmt.Sprintf("echo -e '%s\\n%s' | smbpasswd -s %s", password, password, username))
	if res.Error != nil {
		return fmt.Errorf("failed to set user password: %w; stderr: %s", res.Error, res.Stderr)
	}
	return nil
}

// EnableUser enables a Samba user
func (c *SambaControl) EnableUser(username string) error {
	res := c.exec("smbpasswd", "-e", username)
	if res.Error != nil {
		return fmt.Errorf("failed to enable user: %w; stderr: %s", res.Error, res.Stderr)
	}
	return nil
}

// DisableUser disables a Samba user
func (c *SambaControl) DisableUser(username string) error {
	res := c.exec("smbpasswd", "-d", username)
	if res.Error != nil {
		return fmt.Errorf("failed to disable user: %w; stderr: %s", res.Error, res.Stderr)
	}
	return nil
}

// ListUsers lists all Samba users
func (c *SambaControl) ListUsers() ([]UserInfo, error) {
	res := c.exec("pdbedit", "-L")
	if res.Error != nil {
		return nil, fmt.Errorf("failed to list users: %w; stderr: %s", res.Error, res.Stderr)
	}

	var users []UserInfo
	lines := strings.Split(strings.TrimSpace(res.Stdout), "\n")

	for _, line := range lines {
		parts := strings.Split(line, ":")
		if len(parts) >= 1 {
			username := strings.TrimSpace(parts[0])
			if username != "" {
				users = append(users, UserInfo{
					Username:  username,
					IsEnabled: !strings.Contains(line, "[D]"), // [D] means disabled
				})
			}
		}
	}

	return users, nil
}

// ReloadConfig reloads Samba configuration
func (c *SambaControl) ReloadConfig() error {
	res := c.exec("smbcontrol", "all", "reload-config")
	if res.Error != nil {
		return fmt.Errorf("failed to reload config: %w; stderr: %s", res.Error, res.Stderr)
	}
	return nil
}

// RestartService restarts Samba service
func (c *SambaControl) RestartService() error {
	res := c.exec("systemctl", "restart", "smbd")
	if res.Error != nil {
		return fmt.Errorf("failed to restart samba service: %w; stderr: %s", res.Error, res.Stderr)
	}
	return nil
}

// GetServiceStatus gets Samba service status
func (c *SambaControl) GetServiceStatus() (*ServiceStatus, error) {
	// Check if service is running
	res := c.exec("systemctl", "is-active", "smbd")
	isRunning := res.Error == nil && strings.TrimSpace(res.Stdout) == "active"

	status := &ServiceStatus{
		IsRunning: isRunning,
	}

	if isRunning {
		// Get process ID
		res = c.exec("systemctl", "show", "smbd", "--property=MainPID", "--value")
		if res.Error == nil {
			fmt.Sscanf(strings.TrimSpace(res.Stdout), "%d", &status.ProcessID)
		}

		// Get version
		res = c.exec("smbd", "--version")
		if res.Error == nil {
			status.Version = strings.TrimSpace(res.Stdout)
		}

		// Get connected users count
		res = c.exec("smbstatus", "-p")
		if res.Error == nil {
			lines := strings.Split(res.Stdout, "\n")
			// Count non-header lines
			count := 0
			for i, line := range lines {
				if i > 2 && strings.TrimSpace(line) != "" { // Skip header lines
					count++
				}
			}
			status.ConnectedUsers = count
		}
	}

	return status, nil
}

// BackupConfig creates a backup of current Samba configuration
func (c *SambaControl) BackupConfig() (string, error) {
	// Ensure backup directory exists
	if err := os.MkdirAll(c.backupDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create backup directory: %w", err)
	}

	timestamp := time.Now().Format("20060102-150405")
	backupPath := fmt.Sprintf("%s/smb.conf.backup.%s", c.backupDir, timestamp)

	res := c.exec("cp", c.configPath, backupPath)
	if res.Error != nil {
		return "", fmt.Errorf("failed to backup config: %w; stderr: %s", res.Error, res.Stderr)
	}

	return backupPath, nil
}

// RestoreConfig restores Samba configuration from backup
func (c *SambaControl) RestoreConfig(backupPath string) error {
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return fmt.Errorf("backup file does not exist: %s", backupPath)
	}

	res := c.exec("cp", backupPath, c.configPath)
	if res.Error != nil {
		return fmt.Errorf("failed to restore config: %w; stderr: %s", res.Error, res.Stderr)
	}

	return c.ReloadConfig()
}

// ValidateConfig validates Samba configuration
func (c *SambaControl) ValidateConfig() error {
	res := c.exec("testparm", "-s")
	if res.Error != nil {
		return fmt.Errorf("configuration validation failed: %w; stderr: %s", res.Error, res.Stderr)
	}
	return nil
}

// Helper methods

func (c *SambaControl) backupCurrentConfig() error {
	_, err := c.BackupConfig()
	return err
}

func (c *SambaControl) restoreLastBackup() error {
	// Find the most recent backup
	res := c.exec("ls", "-t", c.backupDir)
	if res.Error != nil {
		return fmt.Errorf("no backups found")
	}

	lines := strings.Split(strings.TrimSpace(res.Stdout), "\n")
	if len(lines) == 0 {
		return fmt.Errorf("no backups found")
	}

	latestBackup := fmt.Sprintf("%s/%s", c.backupDir, lines[0])
	return c.RestoreConfig(latestBackup)
}

func (c *SambaControl) getShareInfo(shareName string) (*ShareInfo, error) {
	res := c.exec("testparm", "-s", "--section-name", shareName)
	if res.Error != nil {
		return nil, fmt.Errorf("failed to get share info: %w", res.Error)
	}

	shareInfo := &ShareInfo{
		Name:    shareName,
		Options: make(map[string]string),
	}

	lines := strings.Split(res.Stdout, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])

				switch key {
				case "path":
					shareInfo.Path = value
				case "comment":
					shareInfo.Comment = value
				case "available":
					shareInfo.IsEnabled = value != "no"
				default:
					shareInfo.Options[key] = value
				}
			}
		}
	}

	return shareInfo, nil
}