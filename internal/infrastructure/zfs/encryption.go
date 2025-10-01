package zfs

import (
	"fmt"
	"strings"
	"time"
)

// LoadKey loads the encryption key for a dataset
func (c *ZFSClient) LoadKey(dataset string, keyLocation string) error {
	args := []string{"load-key"}

	// If key location is specified, use it
	if keyLocation != "" && keyLocation != "prompt" {
		args = append(args, "-L", fmt.Sprintf("file://%s", keyLocation))
	}

	args = append(args, dataset)

	result := utils.ExecCommand(utils.ExecOptions{
		Timeout: 60 * time.Second,
	}, "zfs", args...)

	if result.Error != nil {
		return fmt.Errorf("failed to load key: %w, stderr: %s", result.Error, result.Stderr)
	}

	return nil
}

// UnloadKey unloads the encryption key for a dataset
func (c *ZFSClient) UnloadKey(dataset string, recursive bool) error {
	args := []string{"unload-key"}

	if recursive {
		args = append(args, "-r")
	}

	args = append(args, dataset)

	_, err := c.execZfs(args...)
	return err
}

// ChangeKey changes the encryption key for a dataset
func (c *ZFSClient) ChangeKey(dataset string, keyLocation string) error {
	args := []string{"change-key"}

	if keyLocation != "" && keyLocation != "prompt" {
		args = append(args, "-o", fmt.Sprintf("keylocation=file://%s", keyLocation))
	}

	args = append(args, dataset)

	result := utils.ExecCommand(utils.ExecOptions{
		Timeout: 60 * time.Second,
	}, "zfs", args...)

	if result.Error != nil {
		return fmt.Errorf("failed to change key: %w, stderr: %s", result.Error, result.Stderr)
	}

	return nil
}

// GetKeyStatus gets the key status of an encrypted dataset
func (c *ZFSClient) GetKeyStatus(dataset string) (string, error) {
	result, err := c.execZfs("get", "-H", "-o", "value", "keystatus", dataset)
	if err != nil {
		return "", err
	}

	status := strings.TrimSpace(result.Stdout)
	return status, nil
}

// GetEncryptionRoot gets the encryption root dataset
func (c *ZFSClient) GetEncryptionRoot(dataset string) (string, error) {
	result, err := c.execZfs("get", "-H", "-o", "value", "encryptionroot", dataset)
	if err != nil {
		return "", err
	}

	root := strings.TrimSpace(result.Stdout)
	if root == "-" {
		return "", nil
	}

	return root, nil
}

// CreateEncryptedDataset creates a new encrypted dataset with specified encryption algorithm
func (c *ZFSClient) CreateEncryptedDataset(name string, encryption string, keyFormat string, keyLocation string, opts DatasetOptions) error {
	// Set encryption options
	opts.Encryption = encryption
	opts.KeyFormat = keyFormat
	opts.KeyLocation = keyLocation

	return c.CreateDataset(name, opts)
}

// GetEncryptionAlgorithm gets the encryption algorithm used for a dataset
func (c *ZFSClient) GetEncryptionAlgorithm(dataset string) (string, error) {
	result, err := c.execZfs("get", "-H", "-o", "value", "encryption", dataset)
	if err != nil {
		return "", err
	}

	algo := strings.TrimSpace(result.Stdout)
	if algo == "off" || algo == "-" {
		return "", nil
	}

	return algo, nil
}

// GetKeyLocation gets the key location for an encrypted dataset
func (c *ZFSClient) GetKeyLocation(dataset string) (string, error) {
	result, err := c.execZfs("get", "-H", "-o", "value", "keylocation", dataset)
	if err != nil {
		return "", err
	}

	location := strings.TrimSpace(result.Stdout)
	if location == "-" {
		return "", nil
	}

	// Remove file:// prefix if present
	location = strings.TrimPrefix(location, "file://")

	return location, nil
}

// SetKeyLocation sets the key location for an encrypted dataset
func (c *ZFSClient) SetKeyLocation(dataset string, keyLocation string) error {
	location := keyLocation
	if keyLocation != "prompt" {
		location = fmt.Sprintf("file://%s", keyLocation)
	}

	return c.SetProperty(dataset, "keylocation", location)
}

// IsEncrypted checks if a dataset is encrypted
func (c *ZFSClient) IsEncrypted(dataset string) (bool, error) {
	encryption, err := c.GetEncryptionAlgorithm(dataset)
	if err != nil {
		return false, err
	}

	return encryption != "", nil
}

// GetEncryptionProperties gets all encryption-related properties
func (c *ZFSClient) GetEncryptionProperties(dataset string) (map[string]string, error) {
	properties := []string{
		"encryption",
		"encryptionroot",
		"keystatus",
		"keylocation",
		"keyformat",
	}

	result := make(map[string]string)

	for _, prop := range properties {
		value, err := c.GetProperty(dataset, prop)
		if err != nil {
			continue
		}

		if value != "-" && value != "none" {
			result[prop] = value
		}
	}

	return result, nil
}
