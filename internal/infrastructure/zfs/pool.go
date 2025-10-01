package zfs

import (
	"encoding/json"
	"fmt"
	"pnas/internal/utils"
	"strconv"
	"strings"
	"time"
)

// CreatePool creates a new ZFS pool
func (c *ZFSClient) CreatePool(name string, vdevs []VDevSpec, opts PoolOptions) error {
	// Validate pool name
	if err := utils.ValidatePath(name); err != nil {
		return fmt.Errorf("invalid pool name: %w", err)
	}

	args := []string{"create"}

	// Add ashift if specified
	if opts.Ashift > 0 {
		args = append(args, "-o", fmt.Sprintf("ashift=%d", opts.Ashift))
	}

	// Add properties
	for key, value := range opts.Properties {
		args = append(args, "-O", fmt.Sprintf("%s=%s", key, value))
	}

	// Add mountpoint
	if opts.Mountpoint != "" {
		args = append(args, "-m", opts.Mountpoint)
	}

	// Add force flag
	if opts.Force {
		args = append(args, "-f")
	}

	// Add pool name
	args = append(args, name)

	// Add vdev specifications
	for _, vdev := range vdevs {
		if vdev.Type != "stripe" {
			args = append(args, vdev.Type)
		}
		args = append(args, vdev.Devices...)
	}

	_, err := c.execZpool(args...)
	return err
}

// DestroyPool destroys a ZFS pool
func (c *ZFSClient) DestroyPool(name string, force bool) error {
	args := []string{"destroy"}
	if force {
		args = append(args, "-f")
	}
	args = append(args, name)

	_, err := c.execZpool(args...)
	return err
}

// ListPools lists all ZFS pools
func (c *ZFSClient) ListPools() ([]PoolInfo, error) {
	// Get pool list with properties
	result, err := c.execZpool("list", "-H", "-p", "-o",
		"name,size,alloc,free,capacity,health,dedup,altroot,version")
	if err != nil {
		// If no pools exist, return empty list
		if strings.Contains(result.Stderr, "no pools available") {
			return []PoolInfo{}, nil
		}
		return nil, err
	}

	var pools []PoolInfo
	lines := strings.Split(strings.TrimSpace(result.Stdout), "\n")

	for _, line := range lines {
		if line == "" {
			continue
		}

		fields := strings.Split(line, "\t")
		if len(fields) < 9 {
			continue
		}

		size, _ := strconv.ParseUint(fields[1], 10, 64)
		alloc, _ := strconv.ParseUint(fields[2], 10, 64)
		free, _ := strconv.ParseUint(fields[3], 10, 64)
		capacity, _ := strconv.ParseFloat(strings.TrimSuffix(fields[4], "%"), 64)

		pool := PoolInfo{
			Name:      fields[0],
			Size:      size,
			Allocated: alloc,
			Free:      free,
			Capacity:  capacity,
			Health:    fields[5],
			Dedup:     fields[6],
			AltRoot:   fields[7],
			Version:   fields[8],
		}

		pools = append(pools, pool)
	}

	return pools, nil
}

// GetPoolStatus gets detailed status of a pool
func (c *ZFSClient) GetPoolStatus(name string) (*PoolStatus, error) {
	// Get status with verbose output
	result, err := c.execZpool("status", name)
	if err != nil {
		return nil, err
	}

	status := &PoolStatus{
		Name:   name,
		Config: []VDevInfo{},
		Errors: []ErrorInfo{},
	}

	lines := strings.Split(result.Stdout, "\n")
	section := ""

	for _, line := range lines {
		line = strings.TrimRight(line, " \t")
		if line == "" {
			continue
		}

		// Parse different sections
		if strings.Contains(line, "state:") {
			status.State = strings.TrimSpace(strings.Split(line, ":")[1])
		} else if strings.Contains(line, "status:") {
			status.Status = strings.TrimSpace(strings.Split(line, ":")[1])
		} else if strings.Contains(line, "action:") {
			status.Action = strings.TrimSpace(strings.Split(line, ":")[1])
		} else if strings.Contains(line, "scan:") || strings.Contains(line, "scrub:") {
			// Parse scan information
			section = "scan"
			status.Scan = parseScanInfo(line)
		} else if strings.Contains(line, "config:") {
			section = "config"
		} else if strings.Contains(line, "errors:") {
			section = "errors"
		} else if section == "config" && strings.TrimSpace(line) != "" {
			// Parse vdev configuration
			// Skip header line
			if !strings.Contains(line, "NAME") {
				vdev := parseVDevLine(line)
				if vdev != nil {
					status.Config = append(status.Config, *vdev)
				}
			}
		}
	}

	return status, nil
}

// ExportPool exports a ZFS pool
func (c *ZFSClient) ExportPool(name string) error {
	_, err := c.execZpool("export", name)
	return err
}

// ImportPool imports a ZFS pool
func (c *ZFSClient) ImportPool(name string, opts ImportOptions) error {
	args := []string{"import"}

	if opts.Force {
		args = append(args, "-f")
	}

	if opts.Directory != "" {
		args = append(args, "-d", opts.Directory)
	}

	if opts.AltRoot != "" {
		args = append(args, "-R", opts.AltRoot)
	}

	if opts.Readonly {
		args = append(args, "-o", "readonly=on")
	}

	args = append(args, name)

	_, err := c.execZpool(args...)
	return err
}

// ScrubPool starts a scrub operation on a pool
func (c *ZFSClient) ScrubPool(name string) error {
	_, err := c.execZpool("scrub", name)
	return err
}

// GetScrubStatus gets the scrub status of a pool
func (c *ZFSClient) GetScrubStatus(name string) (*ScrubStatus, error) {
	status, err := c.GetPoolStatus(name)
	if err != nil {
		return nil, err
	}

	if status.Scan == nil {
		return &ScrubStatus{
			State: "none",
		}, nil
	}

	scrubStatus := &ScrubStatus{
		State:     status.Scan.State,
		Progress:  status.Scan.Progress,
		Scanned:   status.Scan.Scanned,
		ToScan:    status.Scan.ToScan,
		Errors:    status.Scan.Errors,
		Repaired:  status.Scan.Repaired,
		StartTime: status.Scan.StartTime,
	}

	if !status.Scan.EndTime.IsZero() {
		scrubStatus.EndTime = &status.Scan.EndTime
		scrubStatus.Duration = int64(status.Scan.EndTime.Sub(status.Scan.StartTime).Seconds())
	}

	return scrubStatus, nil
}

// AddCache adds a cache device (L2ARC) to a pool
func (c *ZFSClient) AddCache(pool string, device string) error {
	_, err := c.execZpool("add", pool, "cache", device)
	return err
}

// RemoveCache removes a cache device from a pool
func (c *ZFSClient) RemoveCache(pool string, device string) error {
	_, err := c.execZpool("remove", pool, device)
	return err
}

// AddLog adds a log device (SLOG) to a pool
func (c *ZFSClient) AddLog(pool string, device string) error {
	_, err := c.execZpool("add", pool, "log", device)
	return err
}

// RemoveLog removes a log device from a pool
func (c *ZFSClient) RemoveLog(pool string, device string) error {
	_, err := c.execZpool("remove", pool, device)
	return err
}

// Helper to parse scan information from status output
func parseScanInfo(line string) *ScanInfo {
	// Example: "scan: scrub repaired 0B in 00:00:03 with 0 errors on Sun Jan  1 00:00:00 2024"
	// Example: "scan: scrub in progress since Sun Jan  1 00:00:00 2024"

	info := &ScanInfo{}

	if strings.Contains(line, "scrub") {
		info.Function = "scrub"
	} else if strings.Contains(line, "resilver") {
		info.Function = "resilver"
	}

	if strings.Contains(line, "in progress") {
		info.State = "scanning"
	} else if strings.Contains(line, "repaired") {
		info.State = "finished"
	} else if strings.Contains(line, "canceled") {
		info.State = "canceled"
	}

	// Parse repaired bytes
	if strings.Contains(line, "repaired") {
		parts := strings.Split(line, "repaired")
		if len(parts) > 1 {
			sizeStr := strings.TrimSpace(strings.Split(parts[1], "in")[0])
			if size, err := parseSize(sizeStr); err == nil {
				info.Repaired = size
			}
		}
	}

	// Parse errors
	if strings.Contains(line, "errors") {
		parts := strings.Split(line, "with")
		if len(parts) > 1 {
			errorStr := strings.TrimSpace(strings.Split(parts[1], "errors")[0])
			if errors, err := strconv.Atoi(errorStr); err == nil {
				info.Errors = errors
			}
		}
	}

	return info
}

// Helper to parse vdev line from status output
func parseVDevLine(line string) *VDevInfo {
	// Example: "    mirror-0       ONLINE       0     0     0"
	// Example: "      sda          ONLINE       0     0     0"

	fields := strings.Fields(line)
	if len(fields) < 5 {
		return nil
	}

	vdev := &VDevInfo{
		Name:  fields[0],
		State: fields[1],
	}

	// Parse error counts
	if read, err := strconv.Atoi(fields[2]); err == nil {
		vdev.Read = read
	}
	if write, err := strconv.Atoi(fields[3]); err == nil {
		vdev.Write = write
	}
	if cksum, err := strconv.Atoi(fields[4]); err == nil {
		vdev.Cksum = cksum
	}

	// Determine vdev type from name
	if strings.Contains(vdev.Name, "mirror") {
		vdev.Type = "mirror"
	} else if strings.Contains(vdev.Name, "raidz3") {
		vdev.Type = "raidz3"
	} else if strings.Contains(vdev.Name, "raidz2") {
		vdev.Type = "raidz2"
	} else if strings.Contains(vdev.Name, "raidz") {
		vdev.Type = "raidz"
	} else if strings.Contains(vdev.Name, "cache") {
		vdev.Type = "cache"
	} else if strings.Contains(vdev.Name, "log") {
		vdev.Type = "log"
	} else {
		vdev.Type = "disk"
	}

	return vdev
}

// GetProperty gets a pool property value
func (c *ZFSClient) GetPoolProperty(pool string, property string) (string, error) {
	result, err := c.execZpool("get", "-H", "-o", "value", property, pool)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(result.Stdout), nil
}

// SetPoolProperty sets a pool property
func (c *ZFSClient) SetPoolProperty(pool string, property string, value string) error {
	_, err := c.execZpool("set", fmt.Sprintf("%s=%s", property, value), pool)
	return err
}
