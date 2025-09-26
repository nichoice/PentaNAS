package services

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"pnas/internal/shared/utils"
)

type iSCSICommandService struct{}

func NewiSCSICommandService() *iSCSICommandService {
	return &iSCSICommandService{}
}

// TargetCLI commands for iSCSI target management

// CreateTarget creates an iSCSI target using targetcli
func (s *iSCSICommandService) CreateTarget(iqn string) error {
	cmd := fmt.Sprintf("targetcli /iscsi create %s", iqn)
	_, err := utils.RunCommand(cmd)
	if err != nil {
		return fmt.Errorf("failed to create target %s: %w", iqn, err)
	}

	// Save configuration
	if err := s.saveConfig(); err != nil {
		return fmt.Errorf("failed to save configuration after creating target: %w", err)
	}

	return nil
}

// DeleteTarget deletes an iSCSI target using targetcli
func (s *iSCSICommandService) DeleteTarget(iqn string) error {
	cmd := fmt.Sprintf("targetcli /iscsi delete %s", iqn)
	_, err := utils.RunCommand(cmd)
	if err != nil {
		return fmt.Errorf("failed to delete target %s: %w", iqn, err)
	}

	// Save configuration
	if err := s.saveConfig(); err != nil {
		return fmt.Errorf("failed to save configuration after deleting target: %w", err)
	}

	return nil
}

// CreateLUN creates a LUN for a target
func (s *iSCSICommandService) CreateLUN(iqn string, lunID int, devicePath string, deviceType string) error {
	var cmd string

	switch deviceType {
	case "file":
		// Create fileio backstore
		backstoreName := fmt.Sprintf("fileio_%s_lun%d", strings.Replace(iqn, ":", "_", -1), lunID)
		cmd = fmt.Sprintf("targetcli /backstores/fileio create %s %s", backstoreName, devicePath)
		if _, err := utils.RunCommand(cmd); err != nil {
			return fmt.Errorf("failed to create fileio backstore: %w", err)
		}

		// Create LUN
		cmd = fmt.Sprintf("targetcli /iscsi/%s/tpg1/luns create /backstores/fileio/%s %d", iqn, backstoreName, lunID)

	case "block":
		// Create block backstore
		backstoreName := fmt.Sprintf("block_%s_lun%d", strings.Replace(iqn, ":", "_", -1), lunID)
		cmd = fmt.Sprintf("targetcli /backstores/block create %s %s", backstoreName, devicePath)
		if _, err := utils.RunCommand(cmd); err != nil {
			return fmt.Errorf("failed to create block backstore: %w", err)
		}

		// Create LUN
		cmd = fmt.Sprintf("targetcli /iscsi/%s/tpg1/luns create /backstores/block/%s %d", iqn, backstoreName, lunID)

	case "tcmu":
		// Create user backstore using tcmu-runner
		backstoreName := fmt.Sprintf("user_%s_lun%d", strings.Replace(iqn, ":", "_", -1), lunID)
		cmd = fmt.Sprintf("targetcli /backstores/user:tcmu create %s %s", backstoreName, devicePath)
		if _, err := utils.RunCommand(cmd); err != nil {
			return fmt.Errorf("failed to create tcmu backstore: %w", err)
		}

		// Create LUN
		cmd = fmt.Sprintf("targetcli /iscsi/%s/tpg1/luns create /backstores/user:tcmu/%s %d", iqn, backstoreName, lunID)

	default:
		return fmt.Errorf("unsupported device type: %s", deviceType)
	}

	if _, err := utils.RunCommand(cmd); err != nil {
		return fmt.Errorf("failed to create LUN %d for target %s: %w", lunID, iqn, err)
	}

	// Save configuration
	if err := s.saveConfig(); err != nil {
		return fmt.Errorf("failed to save configuration after creating LUN: %w", err)
	}

	return nil
}

// DeleteLUN deletes a LUN from a target
func (s *iSCSICommandService) DeleteLUN(iqn string, lunID int, backstoreName string) error {
	// Delete LUN
	cmd := fmt.Sprintf("targetcli /iscsi/%s/tpg1/luns delete %d", iqn, lunID)
	if _, err := utils.RunCommand(cmd); err != nil {
		return fmt.Errorf("failed to delete LUN %d from target %s: %w", lunID, iqn, err)
	}

	// Delete backstore (determine type from name)
	var backstorePath string
	if strings.HasPrefix(backstoreName, "fileio_") {
		backstorePath = fmt.Sprintf("/backstores/fileio/%s", backstoreName)
	} else if strings.HasPrefix(backstoreName, "block_") {
		backstorePath = fmt.Sprintf("/backstores/block/%s", backstoreName)
	} else if strings.HasPrefix(backstoreName, "user_") {
		backstorePath = fmt.Sprintf("/backstores/user:tcmu/%s", backstoreName)
	} else {
		return fmt.Errorf("unknown backstore type for %s", backstoreName)
	}

	cmd = fmt.Sprintf("targetcli %s delete", backstorePath)
	if _, err := utils.RunCommand(cmd); err != nil {
		return fmt.Errorf("failed to delete backstore %s: %w", backstoreName, err)
	}

	// Save configuration
	if err := s.saveConfig(); err != nil {
		return fmt.Errorf("failed to save configuration after deleting LUN: %w", err)
	}

	return nil
}

// CreateACL creates an ACL for a target
func (s *iSCSICommandService) CreateACL(iqn, initiatorIQN string) error {
	cmd := fmt.Sprintf("targetcli /iscsi/%s/tpg1/acls create %s", iqn, initiatorIQN)
	if _, err := utils.RunCommand(cmd); err != nil {
		return fmt.Errorf("failed to create ACL %s for target %s: %w", initiatorIQN, iqn, err)
	}

	// Save configuration
	if err := s.saveConfig(); err != nil {
		return fmt.Errorf("failed to save configuration after creating ACL: %w", err)
	}

	return nil
}

// DeleteACL deletes an ACL from a target
func (s *iSCSICommandService) DeleteACL(iqn, initiatorIQN string) error {
	cmd := fmt.Sprintf("targetcli /iscsi/%s/tpg1/acls delete %s", iqn, initiatorIQN)
	if _, err := utils.RunCommand(cmd); err != nil {
		return fmt.Errorf("failed to delete ACL %s from target %s: %w", initiatorIQN, iqn, err)
	}

	// Save configuration
	if err := s.saveConfig(); err != nil {
		return fmt.Errorf("failed to save configuration after deleting ACL: %w", err)
	}

	return nil
}

// SetACLAuth configures CHAP authentication for an ACL
func (s *iSCSICommandService) SetACLAuth(iqn, initiatorIQN, username, password string, mutualAuth bool, mutualUsername, mutualPassword string) error {
	// Set userid and password
	cmd := fmt.Sprintf("targetcli /iscsi/%s/tpg1/acls/%s set auth userid=%s", iqn, initiatorIQN, username)
	if _, err := utils.RunCommand(cmd); err != nil {
		return fmt.Errorf("failed to set userid for ACL %s: %w", initiatorIQN, err)
	}

	cmd = fmt.Sprintf("targetcli /iscsi/%s/tpg1/acls/%s set auth password=%s", iqn, initiatorIQN, password)
	if _, err := utils.RunCommand(cmd); err != nil {
		return fmt.Errorf("failed to set password for ACL %s: %w", initiatorIQN, err)
	}

	// Set mutual authentication if enabled
	if mutualAuth && mutualUsername != "" && mutualPassword != "" {
		cmd = fmt.Sprintf("targetcli /iscsi/%s/tpg1/acls/%s set auth mutual_userid=%s", iqn, initiatorIQN, mutualUsername)
		if _, err := utils.RunCommand(cmd); err != nil {
			return fmt.Errorf("failed to set mutual userid for ACL %s: %w", initiatorIQN, err)
		}

		cmd = fmt.Sprintf("targetcli /iscsi/%s/tpg1/acls/%s set auth mutual_password=%s", iqn, initiatorIQN, mutualPassword)
		if _, err := utils.RunCommand(cmd); err != nil {
			return fmt.Errorf("failed to set mutual password for ACL %s: %w", initiatorIQN, err)
		}
	}

	// Save configuration
	if err := s.saveConfig(); err != nil {
		return fmt.Errorf("failed to save configuration after setting ACL auth: %w", err)
	}

	return nil
}

// ClearACLAuth clears authentication for an ACL
func (s *iSCSICommandService) ClearACLAuth(iqn, initiatorIQN string) error {
	// Clear authentication
	cmd := fmt.Sprintf("targetcli /iscsi/%s/tpg1/acls/%s set auth userid=", iqn, initiatorIQN)
	utils.RunCommand(cmd) // Ignore errors

	cmd = fmt.Sprintf("targetcli /iscsi/%s/tpg1/acls/%s set auth password=", iqn, initiatorIQN)
	utils.RunCommand(cmd) // Ignore errors

	cmd = fmt.Sprintf("targetcli /iscsi/%s/tpg1/acls/%s set auth mutual_userid=", iqn, initiatorIQN)
	utils.RunCommand(cmd) // Ignore errors

	cmd = fmt.Sprintf("targetcli /iscsi/%s/tpg1/acls/%s set auth mutual_password=", iqn, initiatorIQN)
	utils.RunCommand(cmd) // Ignore errors

	// Save configuration
	if err := s.saveConfig(); err != nil {
		return fmt.Errorf("failed to save configuration after clearing ACL auth: %w", err)
	}

	return nil
}

// MapLUNToACL maps a LUN to an ACL with specific permissions
func (s *iSCSICommandService) MapLUNToACL(iqn, initiatorIQN string, lunID int, permission string) error {
	aclPath := fmt.Sprintf("/iscsi/%s/tpg1/acls/%s", iqn, initiatorIQN)
	lunPath := fmt.Sprintf("/iscsi/%s/tpg1/luns/%d", iqn, lunID)

	// Create mapped LUN
	cmd := fmt.Sprintf("targetcli %s create %s", aclPath, lunPath)
	if _, err := utils.RunCommand(cmd); err != nil {
		return fmt.Errorf("failed to map LUN %d to ACL %s: %w", lunID, initiatorIQN, err)
	}

	// Set permissions if not default (rw)
	if permission == "ro" {
		mappedLUNPath := fmt.Sprintf("%s/mapped_lun%d", aclPath, lunID)
		cmd = fmt.Sprintf("targetcli %s set write_protect=1", mappedLUNPath)
		if _, err := utils.RunCommand(cmd); err != nil {
			return fmt.Errorf("failed to set read-only permission for LUN %d on ACL %s: %w", lunID, initiatorIQN, err)
		}
	}

	// Save configuration
	if err := s.saveConfig(); err != nil {
		return fmt.Errorf("failed to save configuration after mapping LUN: %w", err)
	}

	return nil
}

// UnmapLUNFromACL unmaps a LUN from an ACL
func (s *iSCSICommandService) UnmapLUNFromACL(iqn, initiatorIQN string, lunID int) error {
	cmd := fmt.Sprintf("targetcli /iscsi/%s/tpg1/acls/%s delete mapped_lun%d", iqn, initiatorIQN, lunID)
	if _, err := utils.RunCommand(cmd); err != nil {
		return fmt.Errorf("failed to unmap LUN %d from ACL %s: %w", lunID, initiatorIQN, err)
	}

	// Save configuration
	if err := s.saveConfig(); err != nil {
		return fmt.Errorf("failed to save configuration after unmapping LUN: %w", err)
	}

	return nil
}

// TCMU-runner commands

// StartTCMURunner starts the tcmu-runner service
func (s *iSCSICommandService) StartTCMURunner() error {
	cmd := "systemctl start tcmu-runner"
	if _, err := utils.RunCommand(cmd); err != nil {
		return fmt.Errorf("failed to start tcmu-runner: %w", err)
	}
	return nil
}

// StopTCMURunner stops the tcmu-runner service
func (s *iSCSICommandService) StopTCMURunner() error {
	cmd := "systemctl stop tcmu-runner"
	if _, err := utils.RunCommand(cmd); err != nil {
		return fmt.Errorf("failed to stop tcmu-runner: %w", err)
	}
	return nil
}

// GetTCMURunnerStatus gets the status of tcmu-runner service
func (s *iSCSICommandService) GetTCMURunnerStatus() (string, error) {
	cmd := "systemctl is-active tcmu-runner"
	output, err := utils.RunCommand(cmd)
	if err != nil {
		return "inactive", nil
	}
	return strings.TrimSpace(output), nil
}

// Target service management

// StartTargetService starts the target service
func (s *iSCSICommandService) StartTargetService() error {
	cmd := "systemctl start target"
	if _, err := utils.RunCommand(cmd); err != nil {
		return fmt.Errorf("failed to start target service: %w", err)
	}
	return nil
}

// StopTargetService stops the target service
func (s *iSCSICommandService) StopTargetService() error {
	cmd := "systemctl stop target"
	if _, err := utils.RunCommand(cmd); err != nil {
		return fmt.Errorf("failed to stop target service: %w", err)
	}
	return nil
}

// GetTargetServiceStatus gets the status of target service
func (s *iSCSICommandService) GetTargetServiceStatus() (string, error) {
	cmd := "systemctl is-active target"
	output, err := utils.RunCommand(cmd)
	if err != nil {
		return "inactive", nil
	}
	return strings.TrimSpace(output), nil
}

// Configuration management

// saveConfig saves the current targetcli configuration
func (s *iSCSICommandService) saveConfig() error {
	cmd := "targetcli saveconfig"
	if _, err := utils.RunCommand(cmd); err != nil {
		return fmt.Errorf("failed to save targetcli configuration: %w", err)
	}
	return nil
}

// RestoreConfig restores configuration from a saved file
func (s *iSCSICommandService) RestoreConfig(configPath string) error {
	cmd := fmt.Sprintf("targetcli restoreconfig %s", configPath)
	if _, err := utils.RunCommand(cmd); err != nil {
		return fmt.Errorf("failed to restore configuration from %s: %w", configPath, err)
	}
	return nil
}

// ClearConfig clears all current configuration
func (s *iSCSICommandService) ClearConfig() error {
	cmd := "targetcli clearconfig confirm=True"
	if _, err := utils.RunCommand(cmd); err != nil {
		return fmt.Errorf("failed to clear configuration: %w", err)
	}
	return nil
}

// GetTargetInfo gets detailed information about all targets
func (s *iSCSICommandService) GetTargetInfo() (map[string]interface{}, error) {
	cmd := "targetcli ls json=true"
	output, err := utils.RunCommand(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to get target info: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		return nil, fmt.Errorf("failed to parse target info JSON: %w", err)
	}

	return result, nil
}

// GetActiveSessions gets active iSCSI sessions
func (s *iSCSICommandService) GetActiveSessions() ([]map[string]interface{}, error) {
	cmd := "targetcli sessions"
	output, err := utils.RunCommand(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to get active sessions: %w", err)
	}

	// Parse session information from output
	sessions := []map[string]interface{}{}
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "alias:") {
			continue
		}

		// Parse session line format: "session_id: initiator_name -> target_name"
		if strings.Contains(line, "->") {
			parts := strings.Split(line, "->")
			if len(parts) == 2 {
				leftPart := strings.TrimSpace(parts[0])
				targetName := strings.TrimSpace(parts[1])

				// Extract session ID and initiator name
				if colonIndex := strings.Index(leftPart, ":"); colonIndex > 0 {
					sessionID := strings.TrimSpace(leftPart[:colonIndex])
					initiatorName := strings.TrimSpace(leftPart[colonIndex+1:])

					session := map[string]interface{}{
						"session_id":     sessionID,
						"initiator_name": initiatorName,
						"target_name":    targetName,
					}
					sessions = append(sessions, session)
				}
			}
		}
	}

	return sessions, nil
}

// CreateFileBackedLUN creates a file-backed LUN with specified size
func (s *iSCSICommandService) CreateFileBackedLUN(filePath string, sizeBytes int64) error {
	// Create sparse file
	cmd := fmt.Sprintf("truncate -s %d %s", sizeBytes, filePath)
	if _, err := utils.RunCommand(cmd); err != nil {
		return fmt.Errorf("failed to create file %s with size %d: %w", filePath, sizeBytes, err)
	}

	// Set proper permissions
	cmd = fmt.Sprintf("chmod 600 %s", filePath)
	if _, err := utils.RunCommand(cmd); err != nil {
		return fmt.Errorf("failed to set permissions for %s: %w", filePath, err)
	}

	return nil
}

// ResizeFileBackedLUN resizes a file-backed LUN
func (s *iSCSICommandService) ResizeFileBackedLUN(filePath string, newSizeBytes int64) error {
	cmd := fmt.Sprintf("truncate -s %d %s", newSizeBytes, filePath)
	if _, err := utils.RunCommand(cmd); err != nil {
		return fmt.Errorf("failed to resize file %s to %d bytes: %w", filePath, newSizeBytes, err)
	}
	return nil
}

// GetLUNInfo gets information about a specific LUN
func (s *iSCSICommandService) GetLUNInfo(iqn string, lunID int) (map[string]interface{}, error) {
	cmd := fmt.Sprintf("targetcli /iscsi/%s/tpg1/luns/%d status", iqn, lunID)
	output, err := utils.RunCommand(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to get LUN %d info for target %s: %w", lunID, iqn, err)
	}

	// Parse output and return as map
	info := map[string]interface{}{
		"output": output,
	}

	return info, nil
}