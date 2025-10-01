package zfs

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// CreateDataset creates a new ZFS dataset (filesystem)
func (c *ZFSClient) CreateDataset(name string, opts DatasetOptions) error {
	args := []string{"create"}

	// Add type if volume
	if opts.Type == "volume" {
		return fmt.Errorf("use CreateVolume for volumes")
	}

	// Add parent creation flag
	if opts.CreateParents {
		args = append(args, "-p")
	}

	// Add mountpoint
	if opts.Mountpoint != "" {
		args = append(args, "-o", fmt.Sprintf("mountpoint=%s", opts.Mountpoint))
	}

	// Add quota
	if opts.Quota > 0 {
		args = append(args, "-o", fmt.Sprintf("quota=%d", opts.Quota))
	}

	// Add reservation
	if opts.Reservation > 0 {
		args = append(args, "-o", fmt.Sprintf("reservation=%d", opts.Reservation))
	}

	// Add compression
	if opts.Compression != "" {
		args = append(args, "-o", fmt.Sprintf("compression=%s", opts.Compression))
	}

	// Add dedup
	if opts.Dedup != "" {
		args = append(args, "-o", fmt.Sprintf("dedup=%s", opts.Dedup))
	}

	// Add encryption
	if opts.Encryption != "" {
		args = append(args, "-o", fmt.Sprintf("encryption=%s", opts.Encryption))

		// Add key location if specified
		if opts.KeyLocation != "" {
			if opts.KeyLocation == "prompt" {
				args = append(args, "-o", "keylocation=prompt")
			} else {
				args = append(args, "-o", fmt.Sprintf("keylocation=file://%s", opts.KeyLocation))
			}
		}

		// Add key format
		if opts.KeyFormat != "" {
			args = append(args, "-o", fmt.Sprintf("keyformat=%s", opts.KeyFormat))
		}
	}

	// Add custom properties
	for key, value := range opts.Properties {
		args = append(args, "-o", fmt.Sprintf("%s=%s", key, value))
	}

	args = append(args, name)

	_, err := c.execZfs(args...)
	return err
}

// DestroyDataset destroys a ZFS dataset
func (c *ZFSClient) DestroyDataset(name string, recursive bool) error {
	args := []string{"destroy"}

	if recursive {
		args = append(args, "-r")
	}

	args = append(args, name)

	_, err := c.execZfs(args...)
	return err
}

// ListDatasets lists all datasets in a pool or under a path
func (c *ZFSClient) ListDatasets(pool string) ([]DatasetInfo, error) {
	args := []string{"list", "-H", "-p", "-t", "filesystem", "-o",
		"name,type,used,avail,refer,mountpoint,compression,quota"}

	if pool != "" {
		args = append(args, "-r", pool)
	}

	result, err := c.execZfs(args...)
	if err != nil {
		return nil, err
	}

	var datasets []DatasetInfo
	lines := strings.Split(strings.TrimSpace(result.Stdout), "\n")

	for _, line := range lines {
		if line == "" {
			continue
		}

		fields := strings.Split(line, "\t")
		if len(fields) < 8 {
			continue
		}

		used, _ := strconv.ParseUint(fields[2], 10, 64)
		avail, _ := strconv.ParseUint(fields[3], 10, 64)
		refer, _ := strconv.ParseUint(fields[4], 10, 64)
		quota, _ := strconv.ParseUint(fields[7], 10, 64)

		dataset := DatasetInfo{
			Name:        fields[0],
			Type:        fields[1],
			Used:        used,
			Available:   avail,
			Refer:       refer,
			Mountpoint:  fields[5],
			Compression: fields[6],
			Quota:       quota,
		}

		datasets = append(datasets, dataset)
	}

	return datasets, nil
}

// GetDatasetProperties gets all properties of a dataset
func (c *ZFSClient) GetDatasetProperties(name string) (*DatasetProperties, error) {
	// Get all properties
	result, err := c.execZfs("get", "-H", "-p", "all", name)
	if err != nil {
		return nil, err
	}

	props := &DatasetProperties{
		Name:          name,
		AllProperties: make(map[string]string),
	}

	lines := strings.Split(strings.TrimSpace(result.Stdout), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		fields := strings.Split(line, "\t")
		if len(fields) < 3 {
			continue
		}

		property := fields[1]
		value := fields[2]

		props.AllProperties[property] = value

		// Parse common properties
		switch property {
		case "type":
			props.Type = value
		case "used":
			props.Used, _ = strconv.ParseUint(value, 10, 64)
		case "available":
			props.Available, _ = strconv.ParseUint(value, 10, 64)
		case "referenced":
			props.Referenced, _ = strconv.ParseUint(value, 10, 64)
		case "mountpoint":
			props.Mountpoint = value
		case "mounted":
			props.Mounted = parseBool(value)
		case "compression":
			props.Compression = value
		case "compressratio":
			props.CompressRatio = value
		case "dedup":
			props.Dedup = value
		case "encryption":
			props.Encryption = value
		case "encryptionroot":
			props.EncryptionRoot = value
		case "keystatus":
			props.KeyStatus = value
		case "quota":
			if value != "0" && value != "none" {
				props.Quota, _ = strconv.ParseUint(value, 10, 64)
			}
		case "reservation":
			if value != "0" && value != "none" {
				props.Reservation, _ = strconv.ParseUint(value, 10, 64)
			}
		case "recordsize":
			if val, err := strconv.Atoi(value); err == nil {
				props.RecordSize = val
			}
		case "readonly":
			props.ReadOnly = parseBool(value)
		case "atime":
			props.Atime = parseBool(value)
		case "sync":
			props.Sync = value
		}
	}

	return props, nil
}

// SetDatasetProperty sets a property on a dataset
func (c *ZFSClient) SetDatasetProperty(name string, property string, value string) error {
	return c.SetProperty(name, property, value)
}

// MountDataset mounts a dataset
func (c *ZFSClient) MountDataset(name string) error {
	_, err := c.execZfs("mount", name)
	return err
}

// UnmountDataset unmounts a dataset
func (c *ZFSClient) UnmountDataset(name string, force bool) error {
	args := []string{"unmount"}
	if force {
		args = append(args, "-f")
	}
	args = append(args, name)

	_, err := c.execZfs(args...)
	return err
}

// CreateVolume creates a new ZFS volume (zvol)
func (c *ZFSClient) CreateVolume(name string, size uint64, opts VolumeOptions) error {
	args := []string{"create"}

	// Add sparse flag
	if opts.Sparse {
		args = append(args, "-s")
	}

	// Add block size
	if opts.BlockSize > 0 {
		args = append(args, "-o", fmt.Sprintf("volblocksize=%d", opts.BlockSize))
	}

	// Add compression
	if opts.Compression != "" {
		args = append(args, "-o", fmt.Sprintf("compression=%s", opts.Compression))
	}

	// Add dedup
	if opts.Dedup != "" {
		args = append(args, "-o", fmt.Sprintf("dedup=%s", opts.Dedup))
	}

	// Add encryption
	if opts.Encryption != "" {
		args = append(args, "-o", fmt.Sprintf("encryption=%s", opts.Encryption))

		if opts.KeyLocation != "" {
			if opts.KeyLocation == "prompt" {
				args = append(args, "-o", "keylocation=prompt")
			} else {
				args = append(args, "-o", fmt.Sprintf("keylocation=file://%s", opts.KeyLocation))
			}
		}
	}

	// Add custom properties
	for key, value := range opts.Properties {
		args = append(args, "-o", fmt.Sprintf("%s=%s", key, value))
	}

	// Add volume type and size
	args = append(args, "-V", fmt.Sprintf("%d", size), name)

	_, err := c.execZfs(args...)
	return err
}

// DestroyVolume destroys a ZFS volume
func (c *ZFSClient) DestroyVolume(name string) error {
	_, err := c.execZfs("destroy", name)
	return err
}

// ListVolumes lists all volumes in a pool
func (c *ZFSClient) ListVolumes(pool string) ([]VolumeInfo, error) {
	args := []string{"list", "-H", "-p", "-t", "volume", "-o",
		"name,volsize,used,avail,volblocksize"}

	if pool != "" {
		args = append(args, "-r", pool)
	}

	result, err := c.execZfs(args...)
	if err != nil {
		// No volumes might not be an error
		if strings.Contains(result.Stderr, "no datasets available") {
			return []VolumeInfo{}, nil
		}
		return nil, err
	}

	var volumes []VolumeInfo
	lines := strings.Split(strings.TrimSpace(result.Stdout), "\n")

	for _, line := range lines {
		if line == "" {
			continue
		}

		fields := strings.Split(line, "\t")
		if len(fields) < 5 {
			continue
		}

		volsize, _ := strconv.ParseUint(fields[1], 10, 64)
		used, _ := strconv.ParseUint(fields[2], 10, 64)
		avail, _ := strconv.ParseUint(fields[3], 10, 64)
		volblock, _ := strconv.Atoi(fields[4])

		// Construct device path
		devicePath := fmt.Sprintf("/dev/zvol/%s", fields[0])

		volume := VolumeInfo{
			Name:       fields[0],
			VolSize:    volsize,
			Size:       volsize,
			Used:       used,
			Available:  avail,
			VolBlock:   volblock,
			DevicePath: devicePath,
		}

		volumes = append(volumes, volume)
	}

	return volumes, nil
}

// ResizeVolume resizes a ZFS volume
func (c *ZFSClient) ResizeVolume(name string, newSize uint64) error {
	_, err := c.execZfs("set", fmt.Sprintf("volsize=%d", newSize), name)
	return err
}

// GetProperty gets a property value from a dataset/volume
func (c *ZFSClient) GetProperty(dataset string, property string) (string, error) {
	result, err := c.execZfs("get", "-H", "-o", "value", property, dataset)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(result.Stdout), nil
}

// SetProperty sets a property on a dataset/volume
func (c *ZFSClient) SetProperty(dataset string, property string, value string) error {
	_, err := c.execZfs("set", fmt.Sprintf("%s=%s", property, value), dataset)
	return err
}

// InheritProperty inherits a property from parent
func (c *ZFSClient) InheritProperty(dataset string, property string) error {
	_, err := c.execZfs("inherit", property, dataset)
	return err
}
