package services

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"pnas/internal/app/dto"
	"pnas/internal/models"
)

type SambaTimeMachineService struct {
	db           *gorm.DB
	shareService *SambaShareService
}

func NewSambaTimeMachineService(db *gorm.DB, shareService *SambaShareService) *SambaTimeMachineService {
	return &SambaTimeMachineService{
		db:           db,
		shareService: shareService,
	}
}

// EnableTimeMachine 启用时间机器支持
func (s *SambaTimeMachineService) EnableTimeMachine(shareID string, quota int64) error {
	// 获取共享
	var share models.SambaShare
	if err := s.db.First(&share, "id = ?", shareID).Error; err != nil {
		return fmt.Errorf("共享不存在: %w", err)
	}

	// 检查路径是否存在
	if _, err := os.Stat(share.Path); os.IsNotExist(err) {
		return fmt.Errorf("共享路径不存在: %s", share.Path)
	}

	// 更新共享配置
	updates := map[string]interface{}{
		"enable_time_machine": true,
		"time_machine_quota":  quota,
	}

	if err := s.db.Model(&share).Updates(updates).Error; err != nil {
		return fmt.Errorf("启用时间机器失败: %w", err)
	}

	// 设置时间机器属性
	if err := s.setTimeMachineAttributes(share.Path); err != nil {
		return fmt.Errorf("设置时间机器属性失败: %w", err)
	}

	// 创建时间机器目录结构
	if err := s.createTimeMachineStructure(share.Path); err != nil {
		return fmt.Errorf("创建时间机器目录结构失败: %w", err)
	}

	return nil
}

// DisableTimeMachine 禁用时间机器支持
func (s *SambaTimeMachineService) DisableTimeMachine(shareID string) error {
	// 获取共享
	var share models.SambaShare
	if err := s.db.First(&share, "id = ?", shareID).Error; err != nil {
		return fmt.Errorf("共享不存在: %w", err)
	}

	// 更新共享配置
	updates := map[string]interface{}{
		"enable_time_machine": false,
		"time_machine_quota":  0,
	}

	if err := s.db.Model(&share).Updates(updates).Error; err != nil {
		return fmt.Errorf("禁用时间机器失败: %w", err)
	}

	// 移除时间机器属性
	if err := s.removeTimeMachineAttributes(share.Path); err != nil {
		return fmt.Errorf("移除时间机器属性失败: %w", err)
	}

	return nil
}

// GetTimeMachineShares 获取时间机器共享列表
func (s *SambaTimeMachineService) GetTimeMachineShares() ([]*dto.SambaShareResponse, error) {
	return s.shareService.GetTimeMachineShares()
}

// SetTimeMachineQuota 设置时间机器配额
func (s *SambaTimeMachineService) SetTimeMachineQuota(shareID string, quota int64) error {
	// 获取共享
	var share models.SambaShare
	if err := s.db.First(&share, "id = ?", shareID).Error; err != nil {
		return fmt.Errorf("共享不存在: %w", err)
	}

	if !share.EnableTimeMachine {
		return fmt.Errorf("共享未启用时间机器支持")
	}

	// 更新配额
	if err := s.db.Model(&share).Update("time_machine_quota", quota).Error; err != nil {
		return fmt.Errorf("设置时间机器配额失败: %w", err)
	}

	return nil
}

// GetTimeMachineStatus 获取时间机器状态
func (s *SambaTimeMachineService) GetTimeMachineStatus(shareID string) (map[string]interface{}, error) {
	// 获取共享
	var share models.SambaShare
	if err := s.db.First(&share, "id = ?", shareID).Error; err != nil {
		return nil, fmt.Errorf("共享不存在: %w", err)
	}

	if !share.EnableTimeMachine {
		return map[string]interface{}{
			"enabled": false,
		}, nil
	}

	// 获取磁盘使用情况
	diskUsage, err := s.getDiskUsage(share.Path)
	if err != nil {
		return nil, fmt.Errorf("获取磁盘使用情况失败: %w", err)
	}

	// 获取备份数量和最后备份时间
	backupInfo, err := s.getBackupInfo(share.Path)
	if err != nil {
		return nil, fmt.Errorf("获取备份信息失败: %w", err)
	}

	status := map[string]interface{}{
		"enabled":         true,
		"share_id":        shareID,
		"share_name":      share.Name,
		"share_path":      share.Path,
		"quota_mb":        share.TimeMachineQuota,
		"used_mb":         diskUsage["used_mb"],
		"available_mb":    diskUsage["available_mb"],
		"usage_percent":   diskUsage["usage_percent"],
		"backup_count":    backupInfo["backup_count"],
		"last_backup":     backupInfo["last_backup"],
		"oldest_backup":   backupInfo["oldest_backup"],
	}

	return status, nil
}

// CreateTimeMachineUser 创建专用的时间机器用户
func (s *SambaTimeMachineService) CreateTimeMachineUser(username, password string) (*dto.SambaAccountResponse, error) {
	accountService := NewSambaAccountService(s.db)

	// 首先创建系统用户
	var user models.User
	if err := s.db.Where("username = ?", username).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// 创建新用户
			user = models.User{
				Base:     models.Base{ID: uuid.New().String()},
				Username: username,
				Password: password, // 这里应该使用哈希后的密码
				IsActive: true,
				Remark:   "Time Machine dedicated user",
			}
			if err := s.db.Create(&user).Error; err != nil {
				return nil, fmt.Errorf("创建用户失败: %w", err)
			}
		} else {
			return nil, fmt.Errorf("查询用户失败: %w", err)
		}
	}

	// 创建Samba账号
	req := &dto.CreateSambaAccountRequest{
		UserID:      user.ID,
		SambaUser:   username,
		Password:    password,
		Role:        string(models.RoleSambaTimeMachine),
		Description: "Time Machine dedicated account",
	}

	return accountService.CreateAccount(req)
}

// ValidateTimeMachineSetup 验证时间机器设置
func (s *SambaTimeMachineService) ValidateTimeMachineSetup(shareID string) error {
	// 获取共享
	var share models.SambaShare
	if err := s.db.First(&share, "id = ?", shareID).Error; err != nil {
		return fmt.Errorf("共享不存在: %w", err)
	}

	if !share.EnableTimeMachine {
		return fmt.Errorf("共享未启用时间机器支持")
	}

	// 检查路径是否存在且可写
	if err := s.validatePath(share.Path); err != nil {
		return fmt.Errorf("路径验证失败: %w", err)
	}

	// 检查磁盘空间
	if share.TimeMachineQuota > 0 {
		diskUsage, err := s.getDiskUsage(share.Path)
		if err != nil {
			return fmt.Errorf("获取磁盘使用情况失败: %w", err)
		}

		availableMB := diskUsage["available_mb"].(int64)
		if availableMB < share.TimeMachineQuota {
			return fmt.Errorf("可用磁盘空间不足，需要 %d MB，可用 %d MB", share.TimeMachineQuota, availableMB)
		}
	}

	// 检查Samba配置
	if err := s.validateSambaConfig(); err != nil {
		return fmt.Errorf("Samba配置验证失败: %w", err)
	}

	return nil
}

// 私有方法

// setTimeMachineAttributes 设置时间机器扩展属性
func (s *SambaTimeMachineService) setTimeMachineAttributes(path string) error {
	// 设置com.apple.metadata:com_apple_backup_excludeItem扩展属性
	// 这告诉macOS这个目录是Time Machine目标
	cmd := exec.Command("setfattr", "-n", "user.com.apple.metadata:com_apple_backup_excludeItem", "-v", "com.apple.backupd", path)
	if _, err := cmd.CombinedOutput(); err != nil {
		// 如果setfattr不可用，尝试使用attr
		cmd = exec.Command("attr", "-s", "com.apple.metadata:com_apple_backup_excludeItem", "-V", "com.apple.backupd", path)
		if output2, err2 := cmd.CombinedOutput(); err2 != nil {
			return fmt.Errorf("设置扩展属性失败: %s, %w", string(output2), err2)
		}
	}

	return nil
}

// removeTimeMachineAttributes 移除时间机器扩展属性
func (s *SambaTimeMachineService) removeTimeMachineAttributes(path string) error {
	cmd := exec.Command("setfattr", "-x", "user.com.apple.metadata:com_apple_backup_excludeItem", path)
	if _, err := cmd.CombinedOutput(); err != nil {
		// 如果setfattr不可用，尝试使用attr
		cmd = exec.Command("attr", "-r", "com.apple.metadata:com_apple_backup_excludeItem", path)
		if output2, err2 := cmd.CombinedOutput(); err2 != nil {
			// 忽略移除不存在属性的错误
			if !strings.Contains(string(output2), "No such attribute") {
				return fmt.Errorf("移除扩展属性失败: %s, %w", string(output2), err2)
			}
		}
	}

	return nil
}

// createTimeMachineStructure 创建时间机器目录结构
func (s *SambaTimeMachineService) createTimeMachineStructure(path string) error {
	// 创建.com.apple.timemachine.supported文件
	supportedFile := filepath.Join(path, ".com.apple.timemachine.supported")
	file, err := os.Create(supportedFile)
	if err != nil {
		return fmt.Errorf("创建时间机器支持文件失败: %w", err)
	}
	file.Close()

	// 设置权限
	if err := os.Chmod(supportedFile, 0644); err != nil {
		return fmt.Errorf("设置文件权限失败: %w", err)
	}

	return nil
}

// getDiskUsage 获取磁盘使用情况
func (s *SambaTimeMachineService) getDiskUsage(path string) (map[string]interface{}, error) {
	cmd := exec.Command("df", "-BM", path)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("获取磁盘使用情况失败: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) < 2 {
		return nil, fmt.Errorf("解析df输出失败")
	}

	// 解析df输出
	fields := strings.Fields(lines[1])
	if len(fields) < 4 {
		return nil, fmt.Errorf("解析df输出失败")
	}

	totalStr := strings.TrimSuffix(fields[1], "M")
	usedStr := strings.TrimSuffix(fields[2], "M")
	availableStr := strings.TrimSuffix(fields[3], "M")

	total, _ := strconv.ParseInt(totalStr, 10, 64)
	used, _ := strconv.ParseInt(usedStr, 10, 64)
	available, _ := strconv.ParseInt(availableStr, 10, 64)

	usagePercent := float64(used) / float64(total) * 100

	return map[string]interface{}{
		"total_mb":       total,
		"used_mb":        used,
		"available_mb":   available,
		"usage_percent":  usagePercent,
	}, nil
}

// getBackupInfo 获取备份信息
func (s *SambaTimeMachineService) getBackupInfo(path string) (map[string]interface{}, error) {
	// 查找.sparsebundle文件（Time Machine备份文件）
	var backupCount int
	var lastBackup, oldestBackup *time.Time

	err := filepath.Walk(path, func(walkPath string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // 忽略访问错误
		}

		if strings.HasSuffix(info.Name(), ".sparsebundle") && info.IsDir() {
			backupCount++

			modTime := info.ModTime()
			if lastBackup == nil || modTime.After(*lastBackup) {
				lastBackup = &modTime
			}
			if oldestBackup == nil || modTime.Before(*oldestBackup) {
				oldestBackup = &modTime
			}
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("扫描备份文件失败: %w", err)
	}

	return map[string]interface{}{
		"backup_count":  backupCount,
		"last_backup":   lastBackup,
		"oldest_backup": oldestBackup,
	}, nil
}

// validatePath 验证路径
func (s *SambaTimeMachineService) validatePath(path string) error {
	// 检查路径是否存在
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("路径不存在: %w", err)
	}

	// 检查是否为目录
	if !info.IsDir() {
		return fmt.Errorf("路径不是目录")
	}

	// 检查是否可写
	testFile := filepath.Join(path, ".tm_write_test")
	file, err := os.Create(testFile)
	if err != nil {
		return fmt.Errorf("路径不可写: %w", err)
	}
	file.Close()
	os.Remove(testFile)

	return nil
}

// validateSambaConfig 验证Samba配置
func (s *SambaTimeMachineService) validateSambaConfig() error {
	// 检查Samba是否支持fruit模块
	cmd := exec.Command("smbd", "-b")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("检查Samba模块失败: %w", err)
	}

	if !strings.Contains(string(output), "fruit") {
		return fmt.Errorf("Samba不支持fruit模块，无法使用时间机器功能")
	}

	return nil
}

// CleanupExpiredBackups 清理过期备份
func (s *SambaTimeMachineService) CleanupExpiredBackups(shareID string, retentionDays int) error {
	// 获取共享
	var share models.SambaShare
	if err := s.db.First(&share, "id = ?", shareID).Error; err != nil {
		return fmt.Errorf("共享不存在: %w", err)
	}

	if !share.EnableTimeMachine {
		return fmt.Errorf("共享未启用时间机器支持")
	}

	cutoffTime := time.Now().AddDate(0, 0, -retentionDays)

	// 查找并删除过期的备份文件
	err := filepath.Walk(share.Path, func(walkPath string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // 忽略访问错误
		}

		if strings.HasSuffix(info.Name(), ".sparsebundle") && info.IsDir() {
			if info.ModTime().Before(cutoffTime) {
				if err := os.RemoveAll(walkPath); err != nil {
					return fmt.Errorf("删除过期备份失败: %w", err)
				}
			}
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("清理过期备份失败: %w", err)
	}

	return nil
}