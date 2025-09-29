package services

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"pnas/internal/app/dto"
	"pnas/internal/database"
	"pnas/internal/infrastructure/samba"
	"pnas/internal/models"
)

type SambaShareService struct {
	db  *gorm.DB
	cli samba.Client
}

func NewSambaShareService() *SambaShareService {
	return NewSambaShareServiceWithDeps(database.DB, samba.NewSambaControl())
}

func NewSambaShareServiceWithDeps(db *gorm.DB, cli samba.Client) *SambaShareService {
	return &SambaShareService{db: db, cli: cli}
}

// CreateShare 创建Samba共享
func (s *SambaShareService) CreateShare(req *dto.CreateSambaShareRequest) (*dto.SambaShareResponse, error) {
	// 检查共享名是否已存在
	var existingShare models.SambaShare
	if err := s.db.Where("name = ?", req.Name).First(&existingShare).Error; err == nil {
		return nil, fmt.Errorf("共享名 %s 已存在", req.Name)
	}

	// 检查路径是否存在，不存在则创建
	if err := s.ensurePathExists(req.Path); err != nil {
		return nil, fmt.Errorf("创建共享路径失败: %w", err)
	}

	// 创建共享
	share := models.SambaShare{
		Base:               models.Base{ID: uuid.New().String()},
		Name:               req.Name,
		Path:               req.Path,
		Comment:            req.Comment,
		IsEnabled:          true,
		AllowGuest:         false,
		GuestOnly:          false,
		Browseable:         true,
		Writable:           false,
		CreateMask:         "0664",
		DirectoryMask:      "0775",
		EnableTimeMachine:  false,
		TimeMachineQuota:   0,
		EnableRecycleBin:   false,
		EnableMultiChannel: false,
	}

	// 设置可选字段
	if req.IsEnabled != nil {
		share.IsEnabled = *req.IsEnabled
	}
	if req.AllowGuest != nil {
		share.AllowGuest = *req.AllowGuest
	}
	if req.GuestOnly != nil {
		share.GuestOnly = *req.GuestOnly
	}
	if req.Browseable != nil {
		share.Browseable = *req.Browseable
	}
	if req.Writable != nil {
		share.Writable = *req.Writable
	}
	if req.CreateMask != "" {
		share.CreateMask = req.CreateMask
	}
	if req.DirectoryMask != "" {
		share.DirectoryMask = req.DirectoryMask
	}
	if req.ForceCreateMode != "" {
		share.ForceCreateMode = req.ForceCreateMode
	}
	if req.ForceDirectoryMode != "" {
		share.ForceDirectoryMode = req.ForceDirectoryMode
	}

	// 设置高级功能
	if req.EnableTimeMachine != nil {
		share.EnableTimeMachine = *req.EnableTimeMachine
	}
	if req.TimeMachineQuota != nil {
		share.TimeMachineQuota = *req.TimeMachineQuota
	}
	if req.EnableRecycleBin != nil {
		share.EnableRecycleBin = *req.EnableRecycleBin
		// 如果启用回收站但未指定路径，则创建默认回收站路径
		if *req.EnableRecycleBin && req.RecycleBinPath == "" {
			share.RecycleBinPath = filepath.Join(req.Path, ".recycle")
		} else {
			share.RecycleBinPath = req.RecycleBinPath
		}
	}
	if req.EnableMultiChannel != nil {
		share.EnableMultiChannel = *req.EnableMultiChannel
	}

	// 创建回收站目录
	if share.EnableRecycleBin && share.RecycleBinPath != "" {
		if err := s.ensurePathExists(share.RecycleBinPath); err != nil {
			return nil, fmt.Errorf("创建回收站路径失败: %w", err)
		}
	}

	if err := s.db.Create(&share).Error; err != nil {
		return nil, fmt.Errorf("创建Samba共享失败: %w", err)
	}

	return s.GetShare(share.ID)
}

// UpdateShare 更新Samba共享
func (s *SambaShareService) UpdateShare(shareID string, req *dto.UpdateSambaShareRequest) (*dto.SambaShareResponse, error) {
	var share models.SambaShare
	if err := s.db.First(&share, "id = ?", shareID).Error; err != nil {
		return nil, fmt.Errorf("Samba共享不存在: %w", err)
	}

	updates := make(map[string]interface{})

	if req.Comment != "" {
		updates["comment"] = req.Comment
	}
	if req.IsEnabled != nil {
		updates["is_enabled"] = *req.IsEnabled
	}
	if req.AllowGuest != nil {
		updates["allow_guest"] = *req.AllowGuest
	}
	if req.GuestOnly != nil {
		updates["guest_only"] = *req.GuestOnly
	}
	if req.Browseable != nil {
		updates["browseable"] = *req.Browseable
	}
	if req.Writable != nil {
		updates["writable"] = *req.Writable
	}
	if req.CreateMask != "" {
		updates["create_mask"] = req.CreateMask
	}
	if req.DirectoryMask != "" {
		updates["directory_mask"] = req.DirectoryMask
	}
	if req.ForceCreateMode != "" {
		updates["force_create_mode"] = req.ForceCreateMode
	}
	if req.ForceDirectoryMode != "" {
		updates["force_directory_mode"] = req.ForceDirectoryMode
	}

	// 更新高级功能
	if req.EnableTimeMachine != nil {
		updates["enable_time_machine"] = *req.EnableTimeMachine
	}
	if req.TimeMachineQuota != nil {
		updates["time_machine_quota"] = *req.TimeMachineQuota
	}
	if req.EnableRecycleBin != nil {
		updates["enable_recycle_bin"] = *req.EnableRecycleBin
		if *req.EnableRecycleBin {
			recyclePath := req.RecycleBinPath
			if recyclePath == "" {
				recyclePath = filepath.Join(share.Path, ".recycle")
			}
			updates["recycle_bin_path"] = recyclePath
			// 创建回收站目录
			if err := s.ensurePathExists(recyclePath); err != nil {
				return nil, fmt.Errorf("创建回收站路径失败: %w", err)
			}
		} else {
			updates["recycle_bin_path"] = ""
		}
	}
	if req.EnableMultiChannel != nil {
		updates["enable_multi_channel"] = *req.EnableMultiChannel
	}

	if len(updates) > 0 {
		if err := s.db.Model(&share).Updates(updates).Error; err != nil {
			return nil, fmt.Errorf("更新Samba共享失败: %w", err)
		}
	}

	return s.GetShare(shareID)
}

// GetShare 获取Samba共享
func (s *SambaShareService) GetShare(shareID string) (*dto.SambaShareResponse, error) {
	var share models.SambaShare
	if err := s.db.Preload("ShareAccess.Account.User").First(&share, "id = ?", shareID).Error; err != nil {
		return nil, fmt.Errorf("Samba共享不存在: %w", err)
	}

	return s.toShareResponse(&share), nil
}

// GetShareByName 根据名称获取Samba共享
func (s *SambaShareService) GetShareByName(name string) (*dto.SambaShareResponse, error) {
	var share models.SambaShare
	if err := s.db.Preload("ShareAccess.Account.User").Where("name = ?", name).First(&share).Error; err != nil {
		return nil, fmt.Errorf("Samba共享不存在: %w", err)
	}

	return s.toShareResponse(&share), nil
}

// ListShares 获取Samba共享列表
func (s *SambaShareService) ListShares(offset, limit int, enabled *bool) ([]*dto.SambaShareResponse, int64, error) {
	var shares []models.SambaShare
	var total int64

	query := s.db.Model(&models.SambaShare{})
	if enabled != nil {
		query = query.Where("is_enabled = ?", *enabled)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("统计Samba共享数量失败: %w", err)
	}

	if err := query.Preload("ShareAccess.Account.User").
		Offset(offset).Limit(limit).Find(&shares).Error; err != nil {
		return nil, 0, fmt.Errorf("获取Samba共享列表失败: %w", err)
	}

	responses := make([]*dto.SambaShareResponse, len(shares))
	for i, share := range shares {
		responses[i] = s.toShareResponse(&share)
	}

	return responses, total, nil
}

// DeleteShare 删除Samba共享
func (s *SambaShareService) DeleteShare(shareID string) error {
	var share models.SambaShare
	if err := s.db.First(&share, "id = ?", shareID).Error; err != nil {
		return fmt.Errorf("Samba共享不存在: %w", err)
	}

	// 删除相关的访问权限
	if err := s.db.Where("share_id = ?", shareID).Delete(&models.SambaShareAccess{}).Error; err != nil {
		return fmt.Errorf("删除共享访问权限失败: %w", err)
	}

	// 删除回收站条目
	if err := s.db.Where("share_id = ?", shareID).Delete(&models.RecycleBinItem{}).Error; err != nil {
		return fmt.Errorf("删除回收站条目失败: %w", err)
	}

	// 删除共享
	if err := s.db.Delete(&share).Error; err != nil {
		return fmt.Errorf("删除Samba共享失败: %w", err)
	}

	return nil
}

// SetShareAccess 设置共享访问权限
func (s *SambaShareService) SetShareAccess(req *dto.SetSambaShareAccessRequest) (*dto.SambaShareAccessResponse, error) {
	// 检查共享是否存在
	var share models.SambaShare
	if err := s.db.First(&share, "id = ?", req.ShareID).Error; err != nil {
		return nil, fmt.Errorf("Samba共享不存在: %w", err)
	}

	// 检查账号是否存在
	var account models.SambaAccount
	if err := s.db.First(&account, "id = ?", req.AccountID).Error; err != nil {
		return nil, fmt.Errorf("Samba账号不存在: %w", err)
	}

	// 检查是否已存在权限记录
	var access models.SambaShareAccess
	if err := s.db.Where("share_id = ? AND account_id = ?", req.ShareID, req.AccountID).First(&access).Error; err == nil {
		// 更新权限
		if err := s.db.Model(&access).Update("permission", req.Permission).Error; err != nil {
			return nil, fmt.Errorf("更新共享访问权限失败: %w", err)
		}
	} else {
		// 创建新权限记录
		access = models.SambaShareAccess{
			Base:       models.Base{ID: uuid.New().String()},
			ShareID:    req.ShareID,
			AccountID:  req.AccountID,
			Permission: req.Permission,
		}
		if err := s.db.Create(&access).Error; err != nil {
			return nil, fmt.Errorf("创建共享访问权限失败: %w", err)
		}
	}

	// 获取完整的权限信息
	if err := s.db.Preload("Share").Preload("Account.User").First(&access, "id = ?", access.ID).Error; err != nil {
		return nil, fmt.Errorf("获取共享访问权限失败: %w", err)
	}

	return s.toShareAccessResponse(&access), nil
}

// RemoveShareAccess 移除共享访问权限
func (s *SambaShareService) RemoveShareAccess(shareID, accountID string) error {
	if err := s.db.Where("share_id = ? AND account_id = ?", shareID, accountID).Delete(&models.SambaShareAccess{}).Error; err != nil {
		return fmt.Errorf("移除共享访问权限失败: %w", err)
	}
	return nil
}

// GetShareAccess 获取共享访问权限列表
func (s *SambaShareService) GetShareAccess(shareID string) ([]*dto.SambaShareAccessResponse, error) {
	var accessList []models.SambaShareAccess
	if err := s.db.Preload("Account.User").Where("share_id = ?", shareID).Find(&accessList).Error; err != nil {
		return nil, fmt.Errorf("获取共享访问权限失败: %w", err)
	}

	responses := make([]*dto.SambaShareAccessResponse, len(accessList))
	for i, access := range accessList {
		responses[i] = s.toShareAccessResponse(&access)
	}

	return responses, nil
}

// GetTimeMachineShares 获取时间机器共享列表
func (s *SambaShareService) GetTimeMachineShares() ([]*dto.SambaShareResponse, error) {
	var shares []models.SambaShare
	if err := s.db.Where("enable_time_machine = ? AND is_enabled = ?", true, true).Find(&shares).Error; err != nil {
		return nil, fmt.Errorf("获取时间机器共享列表失败: %w", err)
	}

	responses := make([]*dto.SambaShareResponse, len(shares))
	for i, share := range shares {
		responses[i] = s.toShareResponse(&share)
	}

	return responses, nil
}

// GetRecycleBinShares 获取启用回收站的共享列表
func (s *SambaShareService) GetRecycleBinShares() ([]*dto.SambaShareResponse, error) {
	var shares []models.SambaShare
	if err := s.db.Where("enable_recycle_bin = ? AND is_enabled = ?", true, true).Find(&shares).Error; err != nil {
		return nil, fmt.Errorf("获取回收站共享列表失败: %w", err)
	}

	responses := make([]*dto.SambaShareResponse, len(shares))
	for i, share := range shares {
		responses[i] = s.toShareResponse(&share)
	}

	return responses, nil
}

// ValidateSharePath 验证共享路径
func (s *SambaShareService) ValidateSharePath(path string) error {
	// 检查路径是否为绝对路径
	if !filepath.IsAbs(path) {
		return fmt.Errorf("路径必须是绝对路径")
	}

	// 检查路径是否存在且可访问
	if _, err := os.Stat(path); err != nil {
		return fmt.Errorf("路径不存在或无法访问: %w", err)
	}

	// 检查是否与现有共享路径冲突
	var shares []models.SambaShare
	if err := s.db.Find(&shares).Error; err != nil {
		return fmt.Errorf("检查路径冲突失败: %w", err)
	}

	for _, share := range shares {
		// 检查是否是子路径或父路径
		if strings.HasPrefix(path, share.Path) || strings.HasPrefix(share.Path, path) {
			return fmt.Errorf("路径与现有共享 %s 冲突", share.Name)
		}
	}

	return nil
}

// 辅助方法

// ensurePathExists 确保路径存在
func (s *SambaShareService) ensurePathExists(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll(path, 0755); err != nil {
			return fmt.Errorf("创建目录失败: %w", err)
		}
	}
	return nil
}

// toShareResponse 转换为共享响应格式
func (s *SambaShareService) toShareResponse(share *models.SambaShare) *dto.SambaShareResponse {
	response := &dto.SambaShareResponse{
		ID:                 share.ID,
		Name:               share.Name,
		Path:               share.Path,
		Comment:            share.Comment,
		IsEnabled:          share.IsEnabled,
		AllowGuest:         share.AllowGuest,
		GuestOnly:          share.GuestOnly,
		Browseable:         share.Browseable,
		Writable:           share.Writable,
		CreateMask:         share.CreateMask,
		DirectoryMask:      share.DirectoryMask,
		ForceCreateMode:    share.ForceCreateMode,
		ForceDirectoryMode: share.ForceDirectoryMode,
		EnableTimeMachine:  share.EnableTimeMachine,
		TimeMachineQuota:   share.TimeMachineQuota,
		EnableRecycleBin:   share.EnableRecycleBin,
		RecycleBinPath:     share.RecycleBinPath,
		EnableMultiChannel: share.EnableMultiChannel,
		CreatedAt:          share.CreatedAt,
		UpdatedAt:          share.UpdatedAt,
	}

	// 添加访问权限信息
	if len(share.ShareAccess) > 0 {
		response.ShareAccess = make([]dto.SambaShareAccessResponse, len(share.ShareAccess))
		for i, access := range share.ShareAccess {
			response.ShareAccess[i] = *s.toShareAccessResponse(&access)
		}
	}

	return response
}

// toShareAccessResponse 转换为访问权限响应格式
func (s *SambaShareService) toShareAccessResponse(access *models.SambaShareAccess) *dto.SambaShareAccessResponse {
	response := &dto.SambaShareAccessResponse{
		ID:         access.ID,
		ShareID:    access.ShareID,
		AccountID:  access.AccountID,
		Permission: access.Permission,
		CreatedAt:  access.CreatedAt,
		UpdatedAt:  access.UpdatedAt,
	}

	// 添加共享信息
	if access.Share.ID != "" {
		response.Share = s.toShareResponse(&access.Share)
	}

	// 添加账号信息
	if access.Account.ID != "" {
		response.Account = &dto.SambaAccountResponse{
			ID:        access.Account.ID,
			UserID:    access.Account.UserID,
			SambaUser: access.Account.SambaUser,
			Role:      string(access.Account.Role),
			IsEnabled: access.Account.IsEnabled,
		}
	}

	return response
}