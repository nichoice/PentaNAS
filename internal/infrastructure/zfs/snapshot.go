package zfs

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// CreateSnapshot creates a snapshot of a dataset
func (c *ZFSClient) CreateSnapshot(dataset string, snapName string, recursive bool) error {
	args := []string{"snapshot"}

	if recursive {
		args = append(args, "-r")
	}

	fullName := fmt.Sprintf("%s@%s", dataset, snapName)
	args = append(args, fullName)

	_, err := c.execZfs(args...)
	return err
}

// DestroySnapshot destroys a snapshot
func (c *ZFSClient) DestroySnapshot(name string, recursive bool) error {
	args := []string{"destroy"}

	if recursive {
		args = append(args, "-r")
	}

	args = append(args, name)

	_, err := c.execZfs(args...)
	return err
}

// ListSnapshots lists snapshots of a dataset
func (c *ZFSClient) ListSnapshots(dataset string) ([]SnapshotInfo, error) {
	args := []string{"list", "-H", "-p", "-t", "snapshot", "-o",
		"name,used,refer,creation", "-s", "creation"}

	if dataset != "" {
		args = append(args, "-r", dataset)
	}

	result, err := c.execZfs(args...)
	if err != nil {
		// No snapshots might not be an error
		if strings.Contains(result.Stderr, "no datasets available") {
			return []SnapshotInfo{}, nil
		}
		return nil, err
	}

	var snapshots []SnapshotInfo
	lines := strings.Split(strings.TrimSpace(result.Stdout), "\n")

	for _, line := range lines {
		if line == "" {
			continue
		}

		fields := strings.Split(line, "\t")
		if len(fields) < 4 {
			continue
		}

		// Parse snapshot name (dataset@snapshot)
		nameParts := strings.Split(fields[0], "@")
		if len(nameParts) != 2 {
			continue
		}

		used, _ := strconv.ParseUint(fields[1], 10, 64)
		refer, _ := strconv.ParseUint(fields[2], 10, 64)

		// Parse creation time (Unix timestamp)
		createTimestamp, _ := strconv.ParseInt(fields[3], 10, 64)
		createTime := time.Unix(createTimestamp, 0)

		snapshot := SnapshotInfo{
			Name:       fields[0],
			Dataset:    nameParts[0],
			SnapName:   nameParts[1],
			Used:       used,
			Referenced: refer,
			CreateTime: createTime,
		}

		snapshots = append(snapshots, snapshot)
	}

	return snapshots, nil
}

// RollbackSnapshot rolls back a dataset to a snapshot
func (c *ZFSClient) RollbackSnapshot(name string) error {
	// Use -r flag to destroy more recent snapshots
	_, err := c.execZfs("rollback", "-r", name)
	return err
}

// CloneSnapshot creates a clone from a snapshot
func (c *ZFSClient) CloneSnapshot(snapshot string, target string) error {
	_, err := c.execZfs("clone", snapshot, target)
	return err
}

// SendSnapshot sends a snapshot to stdout (for backup/replication)
func (c *ZFSClient) SendSnapshot(snapshot string, incremental string) ([]byte, error) {
	args := []string{"send"}

	if incremental != "" {
		args = append(args, "-i", incremental)
	}

	args = append(args, snapshot)

	result, err := c.execZfs(args...)
	if err != nil {
		return nil, err
	}

	return []byte(result.Stdout), nil
}

// ReceiveSnapshot receives a snapshot stream
func (c *ZFSClient) ReceiveSnapshot(target string, data []byte, force bool) error {
	args := []string{"receive"}

	if force {
		args = append(args, "-F")
	}

	args = append(args, target)

	result := utils.ExecCommand(utils.ExecOptions{
		Timeout: 300 * time.Second,
		Input:   string(data),
	}, "zfs", args...)

	if result.Error != nil {
		return fmt.Errorf("zfs receive failed: %w, stderr: %s", result.Error, result.Stderr)
	}

	return nil
}

// DiffSnapshots shows differences between two snapshots
func (c *ZFSClient) DiffSnapshots(snapshot1 string, snapshot2 string) (string, error) {
	var args []string

	if snapshot2 != "" {
		// Compare two snapshots
		args = []string{"diff", snapshot1, snapshot2}
	} else {
		// Compare snapshot with current state
		args = []string{"diff", snapshot1}
	}

	result, err := c.execZfs(args...)
	if err != nil {
		return "", err
	}

	return result.Stdout, nil
}

// HoldSnapshot creates a user hold on a snapshot (prevents deletion)
func (c *ZFSClient) HoldSnapshot(tag string, snapshot string) error {
	_, err := c.execZfs("hold", tag, snapshot)
	return err
}

// ReleaseSnapshot releases a user hold from a snapshot
func (c *ZFSClient) ReleaseSnapshot(tag string, snapshot string) error {
	_, err := c.execZfs("release", tag, snapshot)
	return err
}

// ListHolds lists holds on a snapshot
func (c *ZFSClient) ListHolds(snapshot string) ([]string, error) {
	result, err := c.execZfs("holds", "-H", snapshot)
	if err != nil {
		return nil, err
	}

	var holds []string
	lines := strings.Split(strings.TrimSpace(result.Stdout), "\n")

	for _, line := range lines {
		if line == "" {
			continue
		}

		fields := strings.Split(line, "\t")
		if len(fields) > 1 {
			holds = append(holds, fields[1])
		}
	}

	return holds, nil
}
