package services

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"text/template"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"pnas/internal/app/dto"
	"pnas/internal/models"
)

type SambaConfigService struct {
	db *gorm.DB
}

func NewSambaConfigService(db *gorm.DB) *SambaConfigService {
	return &SambaConfigService{db: db}
}

// UpdateGlobalConfig 更新全局配置
func (s *SambaConfigService) UpdateGlobalConfig(req *dto.UpdateSambaGlobalConfigRequest) (*dto.SambaGlobalConfigResponse, error) {
	var config models.SambaGlobalConfig

	// 查找现有配置，如果不存在则创建默认配置
	if err := s.db.Where("is_active = ?", true).First(&config).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			config = s.createDefaultGlobalConfig()
		} else {
			return nil, fmt.Errorf("获取全局配置失败: %w", err)
		}
	}

	// 更新配置
	updates := make(map[string]interface{})

	if req.Version != "" {
		updates["version"] = models.SambaVersion(req.Version)
	}
	if req.ServerString != "" {
		updates["server_string"] = req.ServerString
	}
	if req.Workgroup != "" {
		updates["workgroup"] = req.Workgroup
	}
	if req.NetbiosName != "" {
		updates["netbios_name"] = req.NetbiosName
	}
	if req.SecurityLevel != "" {
		updates["security_level"] = req.SecurityLevel
	}
	if req.EncryptPasswords != nil {
		updates["encrypt_passwords"] = *req.EncryptPasswords
	}
	if req.PassdbBackend != "" {
		updates["passdb_backend"] = req.PassdbBackend
	}

	// 网络配置
	if req.Interfaces != "" {
		updates["interfaces"] = req.Interfaces
	}
	if req.BindInterfacesOnly != nil {
		updates["bind_interfaces_only"] = *req.BindInterfacesOnly
	}
	if req.SocketOptions != "" {
		updates["socket_options"] = req.SocketOptions
	}

	// 日志配置
	if req.LogLevel != nil {
		updates["log_level"] = *req.LogLevel
	}
	if req.LogFile != "" {
		updates["log_file"] = req.LogFile
	}
	if req.MaxLogSize != nil {
		updates["max_log_size"] = *req.MaxLogSize
	}

	// 性能配置
	if req.DeadTime != nil {
		updates["dead_time"] = *req.DeadTime
	}
	if req.GetWDCacheTime != nil {
		updates["getwd_cache_time"] = *req.GetWDCacheTime
	}
	if req.LPQCacheTime != nil {
		updates["lpq_cache_time"] = *req.LPQCacheTime
	}
	if req.MaxConnections != nil {
		updates["max_connections"] = *req.MaxConnections
	}

	// 多通道配置
	if req.EnableMultiChannel != nil {
		updates["enable_multi_channel"] = *req.EnableMultiChannel
	}
	if req.MaxChannels != nil {
		updates["max_channels"] = *req.MaxChannels
	}

	// 其他配置
	if req.MapToGuest != "" {
		updates["map_to_guest"] = req.MapToGuest
	}
	if req.GuestAccount != "" {
		updates["guest_account"] = req.GuestAccount
	}
	if req.HostsAllow != "" {
		updates["hosts_allow"] = req.HostsAllow
	}
	if req.HostsDeny != "" {
		updates["hosts_deny"] = req.HostsDeny
	}

	// 审计配置
	if req.EnableAuditing != nil {
		updates["enable_auditing"] = *req.EnableAuditing
	}
	if req.AuditPrefix != "" {
		updates["audit_prefix"] = req.AuditPrefix
	}
	if req.FullAuditPrefix != "" {
		updates["full_audit_prefix"] = req.FullAuditPrefix
	}

	if len(updates) > 0 {
		if config.ID == "" {
			// 新建配置
			config.ID = uuid.New().String()
			if err := s.db.Create(&config).Error; err != nil {
				return nil, fmt.Errorf("创建全局配置失败: %w", err)
			}
		}

		if err := s.db.Model(&config).Updates(updates).Error; err != nil {
			return nil, fmt.Errorf("更新全局配置失败: %w", err)
		}
	}

	return s.GetGlobalConfig()
}

// GetGlobalConfig 获取全局配置
func (s *SambaConfigService) GetGlobalConfig() (*dto.SambaGlobalConfigResponse, error) {
	var config models.SambaGlobalConfig
	if err := s.db.Where("is_active = ?", true).First(&config).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// 创建默认配置
			config = s.createDefaultGlobalConfig()
			if err := s.db.Create(&config).Error; err != nil {
				return nil, fmt.Errorf("创建默认全局配置失败: %w", err)
			}
		} else {
			return nil, fmt.Errorf("获取全局配置失败: %w", err)
		}
	}

	return s.toGlobalConfigResponse(&config), nil
}

// GenerateConfig 生成Samba配置文件
func (s *SambaConfigService) GenerateConfig() (*dto.SambaConfigGenerateResponse, error) {
	// 获取全局配置
	globalConfig, err := s.GetGlobalConfig()
	if err != nil {
		return nil, fmt.Errorf("获取全局配置失败: %w", err)
	}

	// 获取所有启用的共享
	var shares []models.SambaShare
	if err := s.db.Where("is_enabled = ?", true).Find(&shares).Error; err != nil {
		return nil, fmt.Errorf("获取共享列表失败: %w", err)
	}

	// 生成配置内容
	configContent, err := s.generateConfigContent(globalConfig, shares)
	if err != nil {
		return nil, fmt.Errorf("生成配置内容失败: %w", err)
	}

	// 计算MD5校验和
	hash := md5.Sum([]byte(configContent))
	checksumMD5 := fmt.Sprintf("%x", hash)

	return &dto.SambaConfigGenerateResponse{
		ConfigContent: configContent,
		FilePath:      "/etc/samba/smb.conf",
		GeneratedAt:   time.Now(),
		ChecksumMD5:   checksumMD5,
	}, nil
}

// WriteConfig 写入配置文件
func (s *SambaConfigService) WriteConfig(configPath string) error {
	configResp, err := s.GenerateConfig()
	if err != nil {
		return fmt.Errorf("生成配置失败: %w", err)
	}

	// 备份现有配置
	if err := s.backupConfig(configPath); err != nil {
		return fmt.Errorf("备份配置文件失败: %w", err)
	}

	// 写入新配置
	if err := ioutil.WriteFile(configPath, []byte(configResp.ConfigContent), 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %w", err)
	}

	// 验证配置文件语法
	if err := s.validateConfig(configPath); err != nil {
		return fmt.Errorf("配置文件验证失败: %w", err)
	}

	return nil
}

// ReloadConfig 重载Samba配置
func (s *SambaConfigService) ReloadConfig() error {
	cmd := exec.Command("systemctl", "reload", "smbd")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("重载Samba配置失败: %s, %w", string(output), err)
	}
	return nil
}

// ValidateConfig 验证配置文件
func (s *SambaConfigService) validateConfig(configPath string) error {
	cmd := exec.Command("testparm", "-s", configPath)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("配置文件验证失败: %s, %w", string(output), err)
	}
	return nil
}

// GetSambaVersion 获取Samba版本
func (s *SambaConfigService) GetSambaVersion() (string, error) {
	cmd := exec.Command("smbd", "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("获取Samba版本失败: %w", err)
	}

	version := strings.TrimSpace(string(output))
	// 解析版本号，例如 "Version 4.18.6"
	parts := strings.Fields(version)
	if len(parts) >= 2 {
		return parts[1], nil
	}

	return version, nil
}

// 私有方法

// createDefaultGlobalConfig 创建默认全局配置
func (s *SambaConfigService) createDefaultGlobalConfig() models.SambaGlobalConfig {
	return models.SambaGlobalConfig{
		Base:                 models.Base{ID: uuid.New().String()},
		Version:              models.SambaVersion4_18,
		ServerString:         "PNAS Samba Server",
		Workgroup:            "WORKGROUP",
		NetbiosName:          "PNAS",
		SecurityLevel:        "user",
		EncryptPasswords:     true,
		PassdbBackend:        "tdbsam",
		Interfaces:           "",
		BindInterfacesOnly:   false,
		SocketOptions:        "TCP_NODELAY IPTOS_LOWDELAY SO_RCVBUF=524288 SO_SNDBUF=524288",
		LogLevel:             1,
		LogFile:              "/var/log/samba/samba.log",
		MaxLogSize:           5000,
		DeadTime:             15,
		GetWDCacheTime:       10,
		LPQCacheTime:         10,
		MaxConnections:       0,
		EnableMultiChannel:   false,
		MaxChannels:          4,
		MapToGuest:           "Never",
		GuestAccount:         "nobody",
		HostsAllow:           "",
		HostsDeny:            "",
		EnableAuditing:       false,
		AuditPrefix:          "",
		FullAuditPrefix:      "",
		IsActive:             true,
	}
}

// generateConfigContent 生成配置文件内容
func (s *SambaConfigService) generateConfigContent(globalConfig *dto.SambaGlobalConfigResponse, shares []models.SambaShare) (string, error) {
	tmpl := `# Samba configuration file
# Generated by PNAS at {{.GeneratedTime}}

[global]
	# Basic Configuration
	server string = {{.GlobalConfig.ServerString}}
	workgroup = {{.GlobalConfig.Workgroup}}
	{{- if .GlobalConfig.NetbiosName}}
	netbios name = {{.GlobalConfig.NetbiosName}}
	{{- end}}
	security = {{.GlobalConfig.SecurityLevel}}
	encrypt passwords = {{if .GlobalConfig.EncryptPasswords}}yes{{else}}no{{end}}
	passdb backend = {{.GlobalConfig.PassdbBackend}}

	# Network Configuration
	{{- if .GlobalConfig.Interfaces}}
	interfaces = {{.GlobalConfig.Interfaces}}
	{{- end}}
	bind interfaces only = {{if .GlobalConfig.BindInterfacesOnly}}yes{{else}}no{{end}}
	{{- if .GlobalConfig.SocketOptions}}
	socket options = {{.GlobalConfig.SocketOptions}}
	{{- end}}

	# Logging Configuration
	log level = {{.GlobalConfig.LogLevel}}
	log file = {{.GlobalConfig.LogFile}}
	max log size = {{.GlobalConfig.MaxLogSize}}

	# Performance Configuration
	dead time = {{.GlobalConfig.DeadTime}}
	getwd cache = {{if gt .GlobalConfig.GetWDCacheTime 0}}yes{{else}}no{{end}}
	lpq cache time = {{.GlobalConfig.LPQCacheTime}}
	{{- if gt .GlobalConfig.MaxConnections 0}}
	max connections = {{.GlobalConfig.MaxConnections}}
	{{- end}}

	# Multi-channel Configuration
	{{- if .GlobalConfig.EnableMultiChannel}}
	server multi channel support = yes
	server max channels = {{.GlobalConfig.MaxChannels}}
	{{- else}}
	server multi channel support = no
	{{- end}}

	# Guest Access Configuration
	map to guest = {{.GlobalConfig.MapToGuest}}
	guest account = {{.GlobalConfig.GuestAccount}}

	# Access Control
	{{- if .GlobalConfig.HostsAllow}}
	hosts allow = {{.GlobalConfig.HostsAllow}}
	{{- end}}
	{{- if .GlobalConfig.HostsDeny}}
	hosts deny = {{.GlobalConfig.HostsDeny}}
	{{- end}}

	# Auditing Configuration
	{{- if .GlobalConfig.EnableAuditing}}
	vfs objects = full_audit
	{{- if .GlobalConfig.AuditPrefix}}
	full_audit:prefix = {{.GlobalConfig.AuditPrefix}}
	{{- end}}
	full_audit:success = open opendir read write rename unlink mkdir rmdir
	full_audit:failure = open opendir read write rename unlink mkdir rmdir
	full_audit:facility = local5
	full_audit:priority = notice
	{{- end}}

	# Other Configuration
	unix extensions = yes
	wide links = no
	follow symlinks = yes

{{range .Shares}}
[{{.Name}}]
	{{- if .Comment}}
	comment = {{.Comment}}
	{{- end}}
	path = {{.Path}}
	browseable = {{if .Browseable}}yes{{else}}no{{end}}
	writable = {{if .Writable}}yes{{else}}no{{end}}
	guest ok = {{if .AllowGuest}}yes{{else}}no{{end}}
	{{- if .GuestOnly}}
	guest only = yes
	{{- end}}
	create mask = {{.CreateMask}}
	directory mask = {{.DirectoryMask}}
	{{- if .ForceCreateMode}}
	force create mode = {{.ForceCreateMode}}
	{{- end}}
	{{- if .ForceDirectoryMode}}
	force directory mode = {{.ForceDirectoryMode}}
	{{- end}}

	{{- if .EnableRecycleBin}}
	# Recycle Bin Configuration
	vfs objects = {{if $.GlobalConfig.EnableAuditing}}full_audit {{end}}recycle
	recycle:repository = {{.RecycleBinPath}}
	recycle:keeptree = yes
	recycle:versions = yes
	recycle:touch = yes
	recycle:directory_mode = 0755
	recycle:subdir_mode = 0700
	{{- end}}

	{{- if .EnableTimeMachine}}
	# Time Machine Configuration
	fruit:aapl = yes
	fruit:time machine = yes
	{{- if gt .TimeMachineQuota 0}}
	fruit:time machine max size = {{.TimeMachineQuota}}M
	{{- end}}
	vfs objects = {{if .EnableRecycleBin}}recycle {{end}}{{if $.GlobalConfig.EnableAuditing}}full_audit {{end}}catia fruit streams_xattr
	{{- end}}

	{{- if .EnableMultiChannel}}
	# Multi-channel support for this share
	smb encrypt = desired
	{{- end}}

{{end}}`

	data := struct {
		GeneratedTime time.Time
		GlobalConfig  *dto.SambaGlobalConfigResponse
		Shares        []models.SambaShare
	}{
		GeneratedTime: time.Now(),
		GlobalConfig:  globalConfig,
		Shares:        shares,
	}

	t, err := template.New("smb.conf").Parse(tmpl)
	if err != nil {
		return "", fmt.Errorf("解析模板失败: %w", err)
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("执行模板失败: %w", err)
	}

	return buf.String(), nil
}

// backupConfig 备份配置文件
func (s *SambaConfigService) backupConfig(configPath string) error {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil // 配置文件不存在，无需备份
	}

	backupPath := fmt.Sprintf("%s.backup.%s", configPath, time.Now().Format("20060102-150405"))

	input, err := ioutil.ReadFile(configPath)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(backupPath, input, 0644)
	if err != nil {
		return err
	}

	return nil
}

// toGlobalConfigResponse 转换为全局配置响应格式
func (s *SambaConfigService) toGlobalConfigResponse(config *models.SambaGlobalConfig) *dto.SambaGlobalConfigResponse {
	return &dto.SambaGlobalConfigResponse{
		ID:                   config.ID,
		Version:              string(config.Version),
		ServerString:         config.ServerString,
		Workgroup:            config.Workgroup,
		NetbiosName:          config.NetbiosName,
		SecurityLevel:        config.SecurityLevel,
		EncryptPasswords:     config.EncryptPasswords,
		PassdbBackend:        config.PassdbBackend,
		Interfaces:           config.Interfaces,
		BindInterfacesOnly:   config.BindInterfacesOnly,
		SocketOptions:        config.SocketOptions,
		LogLevel:             config.LogLevel,
		LogFile:              config.LogFile,
		MaxLogSize:           config.MaxLogSize,
		DeadTime:             config.DeadTime,
		GetWDCacheTime:       config.GetWDCacheTime,
		LPQCacheTime:         config.LPQCacheTime,
		MaxConnections:       config.MaxConnections,
		EnableMultiChannel:   config.EnableMultiChannel,
		MaxChannels:          config.MaxChannels,
		MapToGuest:           config.MapToGuest,
		GuestAccount:         config.GuestAccount,
		HostsAllow:           config.HostsAllow,
		HostsDeny:            config.HostsDeny,
		EnableAuditing:       config.EnableAuditing,
		AuditPrefix:          config.AuditPrefix,
		FullAuditPrefix:      config.FullAuditPrefix,
		IsActive:             config.IsActive,
		CreatedAt:            config.CreatedAt,
		UpdatedAt:            config.UpdatedAt,
	}
}