package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"pnas/internal/infrastructure/zfs"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var zfsCmd = &cobra.Command{
	Use:   "zfs",
	Short: "ZFS storage management commands",
	Long:  `Manage ZFS pools, datasets, volumes, snapshots, and encryption`,
}

// Pool commands
var poolCmd = &cobra.Command{
	Use:   "pool",
	Short: "Manage ZFS pools",
}

var poolListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all ZFS pools",
	Run: func(cmd *cobra.Command, args []string) {
		client := zfs.NewZFSClient()
		pools, err := client.ListPools()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tSIZE\tALLOC\tFREE\tCAPACITY\tHEALTH\tDEDUP")

		for _, pool := range pools {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%.1f%%\t%s\t%s\n",
				pool.Name,
				formatSize(pool.Size),
				formatSize(pool.Allocated),
				formatSize(pool.Free),
				pool.Capacity,
				pool.Health,
				pool.Dedup,
			)
		}
		w.Flush()
	},
}

var poolStatusCmd = &cobra.Command{
	Use:   "status [pool]",
	Short: "Show pool status",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := zfs.NewZFSClient()
		status, err := client.GetPoolStatus(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Pool: %s\n", status.Name)
		fmt.Printf("State: %s\n", status.State)
		if status.Status != "" {
			fmt.Printf("Status: %s\n", status.Status)
		}
		if status.Action != "" {
			fmt.Printf("Action: %s\n", status.Action)
		}

		if status.Scan != nil {
			fmt.Printf("\nScan:\n")
			fmt.Printf("  Function: %s\n", status.Scan.Function)
			fmt.Printf("  State: %s\n", status.Scan.State)
			fmt.Printf("  Progress: %.2f%%\n", status.Scan.Progress)
			if status.Scan.Errors > 0 {
				fmt.Printf("  Errors: %d\n", status.Scan.Errors)
			}
		}

		fmt.Printf("\nConfiguration:\n")
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tSTATE\tREAD\tWRITE\tCKSUM")
		for _, vdev := range status.Config {
			fmt.Fprintf(w, "%s\t%s\t%d\t%d\t%d\n",
				vdev.Name, vdev.State, vdev.Read, vdev.Write, vdev.Cksum)
		}
		w.Flush()
	},
}

var poolCreateCmd = &cobra.Command{
	Use:   "create [pool] [vdev-spec]",
	Short: "Create a new ZFS pool",
	Long: `Create a new ZFS pool with specified vdev configuration.

Examples:
  # Single disk
  pnas zfs pool create mypool /dev/sdb

  # Mirror
  pnas zfs pool create mypool mirror /dev/sdb /dev/sdc

  # RAIDZ (RAID5)
  pnas zfs pool create mypool raidz /dev/sdb /dev/sdc /dev/sdd

  # RAIDZ2 (RAID6)
  pnas zfs pool create mypool raidz2 /dev/sdb /dev/sdc /dev/sdd /dev/sde`,
	Args: cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		poolName := args[0]
		vdevArgs := args[1:]

		// Parse vdev specification
		vdevs := parseVDevSpec(vdevArgs)

		ashift, _ := cmd.Flags().GetInt("ashift")
		compression, _ := cmd.Flags().GetString("compression")
		mountpoint, _ := cmd.Flags().GetString("mountpoint")
		force, _ := cmd.Flags().GetBool("force")

		opts := zfs.PoolOptions{
			Ashift:     ashift,
			Mountpoint: mountpoint,
			Force:      force,
			Properties: make(map[string]string),
		}

		if compression != "" {
			opts.Properties["compression"] = compression
		}

		client := zfs.NewZFSClient()
		if err := client.CreatePool(poolName, vdevs, opts); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating pool: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Successfully created pool: %s\n", poolName)
	},
}

var poolDestroyCmd = &cobra.Command{
	Use:   "destroy [pool]",
	Short: "Destroy a ZFS pool",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		force, _ := cmd.Flags().GetBool("force")

		client := zfs.NewZFSClient()
		if err := client.DestroyPool(args[0], force); err != nil {
			fmt.Fprintf(os.Stderr, "Error destroying pool: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Successfully destroyed pool: %s\n", args[0])
	},
}

var poolScrubCmd = &cobra.Command{
	Use:   "scrub [pool]",
	Short: "Start a scrub on a pool",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := zfs.NewZFSClient()
		if err := client.ScrubPool(args[0]); err != nil {
			fmt.Fprintf(os.Stderr, "Error starting scrub: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Scrub started on pool: %s\n", args[0])
	},
}

// Dataset commands
var datasetCmd = &cobra.Command{
	Use:   "dataset",
	Short: "Manage ZFS datasets",
}

var datasetListCmd = &cobra.Command{
	Use:   "list [pool]",
	Short: "List datasets",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		pool := ""
		if len(args) > 0 {
			pool = args[0]
		}

		client := zfs.NewZFSClient()
		datasets, err := client.ListDatasets(pool)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tUSED\tAVAIL\tREFER\tMOUNTPOINT\tCOMPRESS")

		for _, ds := range datasets {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
				ds.Name,
				formatSize(ds.Used),
				formatSize(ds.Available),
				formatSize(ds.Refer),
				ds.Mountpoint,
				ds.Compression,
			)
		}
		w.Flush()
	},
}

var datasetCreateCmd = &cobra.Command{
	Use:   "create [dataset]",
	Short: "Create a ZFS dataset",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		mountpoint, _ := cmd.Flags().GetString("mountpoint")
		compression, _ := cmd.Flags().GetString("compression")
		quota, _ := cmd.Flags().GetString("quota")
		encryption, _ := cmd.Flags().GetString("encryption")
		keyfile, _ := cmd.Flags().GetString("keyfile")

		opts := zfs.DatasetOptions{
			Mountpoint:  mountpoint,
			Compression: compression,
			Properties:  make(map[string]string),
		}

		if quota != "" {
			size, err := parseHumanSize(quota)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Invalid quota: %v\n", err)
				os.Exit(1)
			}
			opts.Quota = size
		}

		if encryption != "" {
			opts.Encryption = encryption
			opts.KeyFormat = "raw"
			if keyfile != "" {
				opts.KeyLocation = keyfile
			} else {
				opts.KeyLocation = "prompt"
			}
		}

		client := zfs.NewZFSClient()
		if err := client.CreateDataset(name, opts); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating dataset: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Successfully created dataset: %s\n", name)
	},
}

var datasetDestroyCmd = &cobra.Command{
	Use:   "destroy [dataset]",
	Short: "Destroy a ZFS dataset",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		recursive, _ := cmd.Flags().GetBool("recursive")

		client := zfs.NewZFSClient()
		if err := client.DestroyDataset(args[0], recursive); err != nil {
			fmt.Fprintf(os.Stderr, "Error destroying dataset: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Successfully destroyed dataset: %s\n", args[0])
	},
}

// Snapshot commands
var snapshotCmd = &cobra.Command{
	Use:   "snapshot",
	Short: "Manage ZFS snapshots",
}

var snapshotListCmd = &cobra.Command{
	Use:   "list [dataset]",
	Short: "List snapshots",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		dataset := ""
		if len(args) > 0 {
			dataset = args[0]
		}

		client := zfs.NewZFSClient()
		snapshots, err := client.ListSnapshots(dataset)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tUSED\tREFER\tCREATED")

		for _, snap := range snapshots {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
				snap.Name,
				formatSize(snap.Used),
				formatSize(snap.Referenced),
				snap.CreateTime.Format("2006-01-02 15:04"),
			)
		}
		w.Flush()
	},
}

var snapshotCreateCmd = &cobra.Command{
	Use:   "create [dataset@snapshot]",
	Short: "Create a snapshot",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		recursive, _ := cmd.Flags().GetBool("recursive")

		// Parse dataset@snapshot
		parts := parseSnapshotName(args[0])
		if parts == nil {
			fmt.Fprintf(os.Stderr, "Invalid snapshot name. Use format: dataset@snapshot\n")
			os.Exit(1)
		}

		client := zfs.NewZFSClient()
		if err := client.CreateSnapshot(parts[0], parts[1], recursive); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating snapshot: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Successfully created snapshot: %s\n", args[0])
	},
}

var snapshotDestroyCmd = &cobra.Command{
	Use:   "destroy [snapshot]",
	Short: "Destroy a snapshot",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		recursive, _ := cmd.Flags().GetBool("recursive")

		client := zfs.NewZFSClient()
		if err := client.DestroySnapshot(args[0], recursive); err != nil {
			fmt.Fprintf(os.Stderr, "Error destroying snapshot: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Successfully destroyed snapshot: %s\n", args[0])
	},
}

var snapshotRollbackCmd = &cobra.Command{
	Use:   "rollback [snapshot]",
	Short: "Rollback to a snapshot",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := zfs.NewZFSClient()
		if err := client.RollbackSnapshot(args[0]); err != nil {
			fmt.Fprintf(os.Stderr, "Error rolling back: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Successfully rolled back to: %s\n", args[0])
	},
}

// Volume commands
var volumeCmd = &cobra.Command{
	Use:   "volume",
	Short: "Manage ZFS volumes (zvols)",
}

var volumeListCmd = &cobra.Command{
	Use:   "list [pool]",
	Short: "List volumes",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		pool := ""
		if len(args) > 0 {
			pool = args[0]
		}

		client := zfs.NewZFSClient()
		volumes, err := client.ListVolumes(pool)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tSIZE\tUSED\tAVAIL\tDEVICE")

		for _, vol := range volumes {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
				vol.Name,
				formatSize(vol.Size),
				formatSize(vol.Used),
				formatSize(vol.Available),
				vol.DevicePath,
			)
		}
		w.Flush()
	},
}

var volumeCreateCmd = &cobra.Command{
	Use:   "create [volume] [size]",
	Short: "Create a ZFS volume",
	Long: `Create a ZFS volume (zvol) with specified size.

Examples:
  pnas zfs volume create mypool/vol1 10G
  pnas zfs volume create mypool/vol2 500M --sparse`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		sizeStr := args[1]

		size, err := parseHumanSize(sizeStr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid size: %v\n", err)
			os.Exit(1)
		}

		sparse, _ := cmd.Flags().GetBool("sparse")
		compression, _ := cmd.Flags().GetString("compression")
		blocksize, _ := cmd.Flags().GetInt("blocksize")

		opts := zfs.VolumeOptions{
			Sparse:      sparse,
			Compression: compression,
			BlockSize:   blocksize,
			Properties:  make(map[string]string),
		}

		client := zfs.NewZFSClient()
		if err := client.CreateVolume(name, size, opts); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating volume: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Successfully created volume: %s (%s)\n", name, formatSize(size))
	},
}

// Helper functions
func init() {
	rootCmd.AddCommand(zfsCmd)

	// Pool commands
	zfsCmd.AddCommand(poolCmd)
	poolCmd.AddCommand(poolListCmd)
	poolCmd.AddCommand(poolStatusCmd)
	poolCmd.AddCommand(poolCreateCmd)
	poolCmd.AddCommand(poolDestroyCmd)
	poolCmd.AddCommand(poolScrubCmd)

	poolCreateCmd.Flags().Int("ashift", 12, "Sector size (9=512B, 12=4K, 13=8K)")
	poolCreateCmd.Flags().String("compression", "lz4", "Compression algorithm")
	poolCreateCmd.Flags().String("mountpoint", "", "Mountpoint")
	poolCreateCmd.Flags().Bool("force", false, "Force creation")

	poolDestroyCmd.Flags().Bool("force", false, "Force destruction")

	// Dataset commands
	zfsCmd.AddCommand(datasetCmd)
	datasetCmd.AddCommand(datasetListCmd)
	datasetCmd.AddCommand(datasetCreateCmd)
	datasetCmd.AddCommand(datasetDestroyCmd)

	datasetCreateCmd.Flags().String("mountpoint", "", "Mountpoint")
	datasetCreateCmd.Flags().String("compression", "lz4", "Compression algorithm")
	datasetCreateCmd.Flags().String("quota", "", "Quota (e.g., 100G)")
	datasetCreateCmd.Flags().String("encryption", "", "Encryption algorithm (aes-256-gcm)")
	datasetCreateCmd.Flags().String("keyfile", "", "Key file location")

	datasetDestroyCmd.Flags().Bool("recursive", false, "Recursive destroy")

	// Snapshot commands
	zfsCmd.AddCommand(snapshotCmd)
	snapshotCmd.AddCommand(snapshotListCmd)
	snapshotCmd.AddCommand(snapshotCreateCmd)
	snapshotCmd.AddCommand(snapshotDestroyCmd)
	snapshotCmd.AddCommand(snapshotRollbackCmd)

	snapshotCreateCmd.Flags().Bool("recursive", false, "Recursive snapshot")
	snapshotDestroyCmd.Flags().Bool("recursive", false, "Recursive destroy")

	// Volume commands
	zfsCmd.AddCommand(volumeCmd)
	volumeCmd.AddCommand(volumeListCmd)
	volumeCmd.AddCommand(volumeCreateCmd)

	volumeCreateCmd.Flags().Bool("sparse", false, "Create sparse volume")
	volumeCreateCmd.Flags().String("compression", "lz4", "Compression algorithm")
	volumeCreateCmd.Flags().Int("blocksize", 8192, "Block size")
}

func parseVDevSpec(args []string) []zfs.VDevSpec {
	var vdevs []zfs.VDevSpec
	var currentVDev *zfs.VDevSpec

	for _, arg := range args {
		if arg == "mirror" || arg == "raidz" || arg == "raidz2" || arg == "raidz3" {
			if currentVDev != nil {
				vdevs = append(vdevs, *currentVDev)
			}
			currentVDev = &zfs.VDevSpec{
				Type:    arg,
				Devices: []string{},
			}
		} else {
			if currentVDev == nil {
				currentVDev = &zfs.VDevSpec{
					Type:    "stripe",
					Devices: []string{},
				}
			}
			currentVDev.Devices = append(currentVDev.Devices, arg)
		}
	}

	if currentVDev != nil {
		vdevs = append(vdevs, *currentVDev)
	}

	return vdevs
}

func parseSnapshotName(name string) []string {
	for i, c := range name {
		if c == '@' {
			return []string{name[:i], name[i+1:]}
		}
	}
	return nil
}

func parseHumanSize(sizeStr string) (uint64, error) {
	multiplier := uint64(1)
	value := sizeStr

	if len(sizeStr) < 2 {
		return 0, fmt.Errorf("invalid size format")
	}

	suffix := sizeStr[len(sizeStr)-1]
	switch suffix {
	case 'K', 'k':
		multiplier = 1024
		value = sizeStr[:len(sizeStr)-1]
	case 'M', 'm':
		multiplier = 1024 * 1024
		value = sizeStr[:len(sizeStr)-1]
	case 'G', 'g':
		multiplier = 1024 * 1024 * 1024
		value = sizeStr[:len(sizeStr)-1]
	case 'T', 't':
		multiplier = 1024 * 1024 * 1024 * 1024
		value = sizeStr[:len(sizeStr)-1]
	}

	var num float64
	_, err := fmt.Sscanf(value, "%f", &num)
	if err != nil {
		return 0, err
	}

	return uint64(num * float64(multiplier)), nil
}

func formatSize(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// JSON output command
var jsonCmd = &cobra.Command{
	Use:   "json",
	Short: "Output in JSON format",
}

var jsonPoolListCmd = &cobra.Command{
	Use:   "pools",
	Short: "List pools in JSON format",
	Run: func(cmd *cobra.Command, args []string) {
		client := zfs.NewZFSClient()
		pools, err := client.ListPools()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		data, _ := json.MarshalIndent(pools, "", "  ")
		fmt.Println(string(data))
	},
}

func init() {
	zfsCmd.AddCommand(jsonCmd)
	jsonCmd.AddCommand(jsonPoolListCmd)
}
