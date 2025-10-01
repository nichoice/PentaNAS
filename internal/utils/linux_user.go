package utils

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// CreateLinuxUser uses nologin shell to create a new user
func CreateLinuxUser(username string) error {
	// Validate username to prevent command injection
	if err := ValidateUsername(username); err != nil {
		return fmt.Errorf("invalid username: %w", err)
	}

	// Check if user exists
	if userExists(username) {
		return fmt.Errorf("user %s already exists", username)
	}

	// Create user with validated username
	cmd := exec.Command("useradd", "-m", "-s", "/sbin/nologin", username)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create user %s: %w, stderr: %s", username, err, stderr.String())
	}

	return nil
}

// DeleteLinuxUser deletes a user
func DeleteLinuxUser(username string) error {
	// Validate username to prevent command injection
	if err := ValidateUsername(username); err != nil {
		return fmt.Errorf("invalid username: %w", err)
	}

	// Check if user exists
	if !userExists(username) {
		return fmt.Errorf("user %s does not exist", username)
	}

	// Delete user with validated username
	cmd := exec.Command("userdel", "-r", username)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to delete user %s: %w, stderr: %s", username, err, stderr.String())
	}

	return nil
}

// getLinuxUsers returns a list of all linux users
func GetLinuxUsers() ([]string, error) {
	cmd := exec.Command("cut", "-d:", "-f1", "/etc/passwd")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to get linux users: %w", err)
	}

	users := strings.Split(string(output), "\n")
	return users, nil
}

// EnableLinuxUser unlocks a user account
func EnableLinuxUser(username string) error {
	// Validate username to prevent command injection
	if err := ValidateUsername(username); err != nil {
		return fmt.Errorf("invalid username: %w", err)
	}

	// Check if user exists
	if !userExists(username) {
		return fmt.Errorf("user %s does not exist", username)
	}

	// Enable user
	cmd := exec.Command("usermod", "-U", username)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to enable user %s: %w, stderr: %s", username, err, stderr.String())
	}

	return nil
}

// DisableLinuxUser locks a user account
func DisableLinuxUser(username string) error {
	// Validate username to prevent command injection
	if err := ValidateUsername(username); err != nil {
		return fmt.Errorf("invalid username: %w", err)
	}

	// Check if user exists
	if !userExists(username) {
		return fmt.Errorf("user %s does not exist", username)
	}

	// Disable user
	cmd := exec.Command("usermod", "-L", username)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to disable user %s: %w, stderr: %s", username, err, stderr.String())
	}

	return nil
}

// SetLinuxUserPassword sets the password for a user using chpasswd (safer than passwd with stdin)
func SetLinuxUserPassword(username, password string) error {
	// Validate inputs to prevent command injection
	if err := ValidateUsername(username); err != nil {
		return fmt.Errorf("invalid username: %w", err)
	}

	if err := ValidatePassword(password); err != nil {
		return fmt.Errorf("invalid password: %w", err)
	}

	// Check if user exists
	if !userExists(username) {
		return fmt.Errorf("user %s does not exist", username)
	}

	// Use chpasswd which is safer than passwd for automation
	// Format: username:password
	cmd := exec.Command("chpasswd")
	cmd.Stdin = strings.NewReader(username + ":" + password)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to set password for user %s: %w, stderr: %s", username, err, stderr.String())
	}

	return nil
}

// SetSambaUserPassword sets the password for a user in Samba
func SetSambaUserPassword(username, password string) error {
	// Validate inputs to prevent command injection
	if err := ValidateUsername(username); err != nil {
		return fmt.Errorf("invalid username: %w", err)
	}

	if err := ValidatePassword(password); err != nil {
		return fmt.Errorf("invalid password: %w", err)
	}

	// Check if user exists
	if !userExists(username) {
		return fmt.Errorf("user %s does not exist", username)
	}

	// Set Samba password using stdin (-s for stdin mode, -a for add user)
	cmd := exec.Command("smbpasswd", "-s", "-a", username)
	cmd.Stdin = strings.NewReader(password + "\n" + password + "\n")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to set Samba password for user %s: %w, stderr: %s", username, err, stderr.String())
	}

	return nil
}

// userExists checks if a user exists
func userExists(username string) bool {
	cmd := exec.Command("id", "-u", username)
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}
