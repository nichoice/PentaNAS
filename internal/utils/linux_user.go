package utils

import (
	"fmt"
	"os/exec"
	"strings"
)

// createLinuxUser uses nologin shell to create a new user
func CreateLinuxUser(username string) error {
	// Implementation here
	// check if user exists
	if userExists(username) {
		return fmt.Errorf("user %s already exists", username)
	}

	// create user
	cmd := exec.Command("useradd", "-m", "-s", "/sbin/nologin", username)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create user %s: %w", username, err)
	}

	return nil
}

// deleteUser deletes a user
func DeleteLinuxUser(username string) error {
	// Implementation here
	// check if user exists
	if !userExists(username) {
		return fmt.Errorf("user %s does not exist", username)
	}

	// delete user
	cmd := exec.Command("userdel", "-r", username)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to delete user %s: %w", username, err)
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

func EnableLinuxUser(username string) error {
	// Implementation here
	// check if user exists
	if !userExists(username) {
		return fmt.Errorf("user %s does not exist", username)
	}

	// enable user
	cmd := exec.Command("usermod", "-U", username)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to enable user %s: %w", username, err)
	}

	return nil
}

// disableLinuxUser disables a user
func DisableLinuxUser(username string) error {
	// Implementation here
	// check if user exists
	if !userExists(username) {
		return fmt.Errorf("user %s does not exist", username)
	}

	// disable user
	cmd := exec.Command("usermod", "-L", username)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to disable user %s: %w", username, err)
	}

	return nil
}

// setLinuxUserPassword sets the password for a user
func SetLinuxUserPassword(username, password string) error {
	// Implementation here
	// check if user exists
	if !userExists(username) {
		return fmt.Errorf("user %s does not exist", username)
	}

	// set password
	cmd := exec.Command("passwd", username)
	cmd.Stdin = strings.NewReader(password + "\n" + password + "\n")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to set password for user %s: %w", username, err)
	}

	return nil
}

// setSambaUserPassword sets the password for a user in Samba
func SetSambaUserPassword(username, password string) error {
	// Implementation here
	// check if user exists
	if !userExists(username) {
		return fmt.Errorf("user %s does not exist", username)
	}

	// set password
	cmd := exec.Command("smbpasswd", "-s", "-a", username)
	cmd.Stdin = strings.NewReader(password + "\n" + password + "\n")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to set password for user %s in Samba: %w", username, err)
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
