package models

// lsblk 输出
type BlockDevice struct {
	Name       string        `json:"name"`
	Path       string        `json:"path"`
	Size       uint64        `json:"size"`
	Serial     string        `json:"serial"` // 使用指针处理可能为 null 的值
	Rota       bool          `json:"rota"`
	Model      string        `json:"model"`
	Vendor     string        `json:"vendor"`
	Type       string        `json:"type"`
	MajMin     string        `json:"maj:min"`
	Mountpoint string        `json:"mountpoint"`
	Fstype     *string       `json:"fstype"`
	Children   []BlockDevice `json:"children,omitempty"`
}

// Disk 磁盘模型
type Disk struct {
	Name         string  `json:"name"`           //磁盘名称
	Path         string  `json:"path"`           //磁盘路径
	Size         string  `json:"size"`           //磁盘大小(字节)
	UsedSize     uint64  `json:"used_size"`      //已使用大小(字节)
	AvailSize    uint64  `json:"avail_size"`     //可用大小(字节)
	UsageRate    float64 `json:"usage_rate"`     //使用率百分比
	Serial       string  `json:"serial"`         //序列号
	Rota         bool    `json:"rota"`           //是否为机械硬盘
	Model        string  `json:"model"`          //型号
	Vendor       string  `json:"vendor"`         //厂商
	Type         string  `json:"type"`           //磁盘类型
	MajMin       string  `json:"maj_min"`        //主次设备号
	MountPoint   string  `json:"mount_point"`    //挂载点
	FileSystem   string  `json:"file_system"`    //文件系统
	IsSystemDisk bool    `json:"is_system_disk"` //是否为系统盘
	IsOnline     bool    `json:"is_online"`      //是否在线
	Temperature  int     `json:"temperature"`    //温度
	Health       string  `json:"health"`         //健康状态
}

// 磁盘使用情况
type DiskUsage struct {
	Total      uint64  `json:"total"`       //总大小(字节)
	Used       uint64  `json:"used"`        //已使用大小(字节)
	Avail      uint64  `json:"avail"`       //可用大小(字节)
	UsePercent float64 `json:"use_percent"` //使用率百分比
}

// smart信息 smartctl -A devicepath
type SmartInfo struct {
	JSONFormatVersion []int `json:"json_format_version"`
	Smartctl          struct {
		Version      []int    `json:"version"`
		SvnRevision  string   `json:"svn_revision"`
		PlatformInfo string   `json:"platform_info"`
		BuildInfo    string   `json:"build_info"`
		Argv         []string `json:"argv"`
		ExitStatus   int      `json:"exit_status"`
	} `json:"smartctl"`
	Device struct {
		Name     string `json:"name"`
		InfoName string `json:"info_name"`
		Type     string `json:"type"`
		Protocol string `json:"protocol"`
	} `json:"device"`
	AtaSmartAttributes struct {
		Revision int `json:"revision"`
		Table    []struct {
			ID         int    `json:"id"`
			Name       string `json:"name"`
			Value      int    `json:"value"`
			Worst      int    `json:"worst"`
			Thresh     int    `json:"thresh"`
			WhenFailed string `json:"when_failed"`
			Flags      struct {
				Value         int    `json:"value"`
				String        string `json:"string"`
				Prefailure    bool   `json:"prefailure"`
				UpdatedOnline bool   `json:"updated_online"`
				Performance   bool   `json:"performance"`
				ErrorRate     bool   `json:"error_rate"`
				EventCount    bool   `json:"event_count"`
				AutoKeep      bool   `json:"auto_keep"`
			} `json:"flags"`
			Raw struct {
				Value  int    `json:"value"`
				String string `json:"string"`
			} `json:"raw"`
		} `json:"table"`
	} `json:"ata_smart_attributes"`
	PowerOnTime struct {
		Hours int `json:"hours"`
	} `json:"power_on_time"`
	PowerCycleCount int `json:"power_cycle_count"`
	Temperature     struct {
		Current int `json:"current"`
	} `json:"temperature"`
}

// 磁盘健康 smartctl -H devicepath
type SmartHealthInfo struct {
	JSONFormatVersion []int `json:"json_format_version"`
	Smartctl          struct {
		Version      []int    `json:"version"`
		SvnRevision  string   `json:"svn_revision"`
		PlatformInfo string   `json:"platform_info"`
		BuildInfo    string   `json:"build_info"`
		Argv         []string `json:"argv"`
		ExitStatus   int      `json:"exit_status"`
	} `json:"smartctl"`
	Device struct {
		Name     string `json:"name"`
		InfoName string `json:"info_name"`
		Type     string `json:"type"`
		Protocol string `json:"protocol"`
	} `json:"device"`
	SmartStatus struct {
		Passed bool `json:"passed"`
	} `json:"smart_status"`
}

func (b *BlockDevice) GetFstype() string {
	if b.Fstype == nil {
		return "无文件系统"
	}
	return *b.Fstype
}
