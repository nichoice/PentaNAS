package services

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"pnas/internal/app/dto"
	"pnas/internal/models"
)

type SambaRecycleService struct {
	db *gorm.DB
}

func NewSambaRecycleService(db *gorm.DB) *SambaRecycleService {
	return &SambaRecycleService{db: db}
}

// AddRecycleBinItem 添加回收站条目
func (s *SambaRecycleService) AddRecycleBinItem(shareID, originalPath, deletedBy string) (*dto.SambaRecycleBinItemResponse, error) {
	// 获取共享信息
	var share models.SambaShare
	if err := s.db.First(&share, "id = ?", shareID).Error; err != nil {
		return nil, fmt.Errorf("共享不存在: %w", err)
	}

	if !share.EnableRecycleBin {
		return nil, fmt.Errorf("共享未启用回收站功能")
	}

	// 检查原文件是否存在
	fileInfo, err := os.Stat(originalPath)
	if err != nil {
		return nil, fmt.Errorf("原文件不存在: %w", err)
	}

	// 生成回收站路径
	recyclePath, err := s.generateRecyclePath(share.RecycleBinPath, originalPath)
	if err != nil {
		return nil, fmt.Errorf("生成回收站路径失败: %w", err)
	}

	// 移动文件到回收站
	if err := s.moveToRecycleBin(originalPath, recyclePath); err != nil {
		return nil, fmt.Errorf("移动文件到回收站失败: %w", err)
	}

	// 创建回收站记录
	item := models.RecycleBinItem{
		Base:         models.Base{ID: uuid.New().String()},
		ShareID:      shareID,
		OriginalPath: originalPath,
		RecyclePath:  recyclePath,
		DeletedBy:    deletedBy,
		DeletedAt:    time.Now(),
		FileSize:     s.getFileSize(fileInfo),
		FileType:     s.getFileType(originalPath),
		IsDirectory:  fileInfo.IsDir(),
	}

	// 设置过期时间（30天后）
	expiresAt := time.Now().AddDate(0, 0, 30)
	item.ExpiresAt = &expiresAt

	if err := s.db.Create(&item).Error; err != nil {
		// 如果数据库操作失败，尝试恢复文件
		s.restoreFromRecycleBin(recyclePath, originalPath)
		return nil, fmt.Errorf("创建回收站记录失败: %w", err)
	}

	return s.GetRecycleBinItem(item.ID)
}

// GetRecycleBinItem 获取回收站条目
func (s *SambaRecycleService) GetRecycleBinItem(itemID string) (*dto.SambaRecycleBinItemResponse, error) {
	var item models.RecycleBinItem
	if err := s.db.Preload("Share").First(&item, "id = ?", itemID).Error; err != nil {
		return nil, fmt.Errorf("回收站条目不存在: %w", err)
	}

	return s.toRecycleBinItemResponse(&item), nil
}

// ListRecycleBinItems 获取回收站条目列表
func (s *SambaRecycleService) ListRecycleBinItems(shareID string, offset, limit int) ([]*dto.SambaRecycleBinItemResponse, int64, error) {
	var items []models.RecycleBinItem
	var total int64

	query := s.db.Model(&models.RecycleBinItem{})
	if shareID != "" {
		query = query.Where("share_id = ?", shareID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("统计回收站条目数量失败: %w", err)
	}

	if err := query.Preload("Share").
		Order("deleted_at DESC").
		Offset(offset).Limit(limit).Find(&items).Error; err != nil {
		return nil, 0, fmt.Errorf("获取回收站条目列表失败: %w", err)
	}

	responses := make([]*dto.SambaRecycleBinItemResponse, len(items))
	for i, item := range items {
		responses[i] = s.toRecycleBinItemResponse(&item)
	}

	return responses, total, nil
}

// RestoreRecycleBinItem 恢复回收站条目
func (s *SambaRecycleService) RestoreRecycleBinItem(req *dto.RestoreRecycleBinItemRequest) error {
	var item models.RecycleBinItem
	if err := s.db.First(&item, "id = ?", req.ItemID).Error; err != nil {
		return fmt.Errorf("回收站条目不存在: %w", err)
	}

	// 确定恢复路径
	restorePath := item.OriginalPath
	if req.RestorePath != "" {
		restorePath = req.RestorePath
	}

	// 检查恢复路径是否已存在
	if _, err := os.Stat(restorePath); err == nil {
		// 文件已存在，生成新名称
		restorePath = s.generateUniqueRestorePath(restorePath)
	}

	// 恢复文件
	if err := s.restoreFromRecycleBin(item.RecyclePath, restorePath); err != nil {
		return fmt.Errorf("恢复文件失败: %w", err)
	}

	// 删除回收站记录
	if err := s.db.Delete(&item).Error; err != nil {
		return fmt.Errorf("删除回收站记录失败: %w", err)
	}

	return nil
}

// PermanentDeleteRecycleBinItem 永久删除回收站条目
func (s *SambaRecycleService) PermanentDeleteRecycleBinItem(itemID string) error {
	var item models.RecycleBinItem
	if err := s.db.First(&item, "id = ?", itemID).Error; err != nil {
		return fmt.Errorf("回收站条目不存在: %w", err)
	}

	// 永久删除文件
	if err := s.permanentDelete(item.RecyclePath); err != nil {
		return fmt.Errorf("永久删除文件失败: %w", err)
	}

	// 删除回收站记录
	if err := s.db.Delete(&item).Error; err != nil {
		return fmt.Errorf("删除回收站记录失败: %w", err)
	}

	return nil
}

// EmptyRecycleBin 清空回收站
func (s *SambaRecycleService) EmptyRecycleBin(shareID string) error {
	// 获取共享信息
	var share models.SambaShare
	if err := s.db.First(&share, "id = ?", shareID).Error; err != nil {
		return fmt.Errorf("共享不存在: %w", err)
	}

	// 获取所有回收站条目
	var items []models.RecycleBinItem
	if err := s.db.Where("share_id = ?", shareID).Find(&items).Error; err != nil {
		return fmt.Errorf("获取回收站条目失败: %w", err)
	}

	// 永久删除所有文件
	for _, item := range items {
		if err := s.permanentDelete(item.RecyclePath); err != nil {
			// 记录错误但继续处理其他文件
			fmt.Printf("删除文件失败 %s: %v\n", item.RecyclePath, err)
		}
	}

	// 删除所有回收站记录
	if err := s.db.Where("share_id = ?", shareID).Delete(&models.RecycleBinItem{}).Error; err != nil {
		return fmt.Errorf("删除回收站记录失败: %w", err)
	}

	return nil
}

// CleanupExpiredItems 清理过期条目
func (s *SambaRecycleService) CleanupExpiredItems() error {
	// 获取所有过期条目
	var items []models.RecycleBinItem
	if err := s.db.Where("expires_at < ?", time.Now()).Find(&items).Error; err != nil {
		return fmt.Errorf("获取过期条目失败: %w", err)
	}

	// 永久删除过期文件
	for _, item := range items {
		if err := s.permanentDelete(item.RecyclePath); err != nil {
			// 记录错误但继续处理其他文件
			fmt.Printf("删除过期文件失败 %s: %v\n", item.RecyclePath, err)
		}
	}

	// 删除过期记录
	if err := s.db.Where("expires_at < ?", time.Now()).Delete(&models.RecycleBinItem{}).Error; err != nil {
		return fmt.Errorf("删除过期记录失败: %w", err)
	}

	return nil
}

// GetRecycleBinStatistics 获取回收站统计信息
func (s *SambaRecycleService) GetRecycleBinStatistics(shareID string) (map[string]interface{}, error) {
	query := s.db.Model(&models.RecycleBinItem{})
	if shareID != "" {
		query = query.Where("share_id = ?", shareID)
	}

	var totalCount int64
	var totalSize int64
	var directoryCount int64
	var fileCount int64

	// 统计总数
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, fmt.Errorf("统计总数失败: %w", err)
	}

	// 统计总大小
	var result struct {
		TotalSize int64
	}
	if err := query.Select("COALESCE(SUM(file_size), 0) as total_size").Scan(&result).Error; err != nil {
		return nil, fmt.Errorf("统计总大小失败: %w", err)
	}
	totalSize = result.TotalSize

	// 统计目录和文件数量
	if err := query.Where("is_directory = ?", true).Count(&directoryCount).Error; err != nil {
		return nil, fmt.Errorf("统计目录数量失败: %w", err)
	}
	fileCount = totalCount - directoryCount

	// 统计即将过期的条目（7天内过期）
	var expiringCount int64
	expiringSoon := time.Now().AddDate(0, 0, 7)
	if err := query.Where("expires_at <= ?", expiringSoon).Count(&expiringCount).Error; err != nil {
		return nil, fmt.Errorf("统计即将过期条目失败: %w", err)
	}

	return map[string]interface{}{
		"total_count":     totalCount,
		"total_size":      totalSize,
		"directory_count": directoryCount,
		"file_count":      fileCount,
		"expiring_count":  expiringCount,
	}, nil
}

// SetRecycleBinRetention 设置回收站保留期限
func (s *SambaRecycleService) SetRecycleBinRetention(shareID string, retentionDays int) error {
	// 更新现有条目的过期时间
	newExpiresAt := time.Now().AddDate(0, 0, retentionDays)
	if err := s.db.Model(&models.RecycleBinItem{}).
		Where("share_id = ?", shareID).
		Update("expires_at", newExpiresAt).Error; err != nil {
		return fmt.Errorf("更新回收站条目过期时间失败: %w", err)
	}

	return nil
}

// 私有方法

// generateRecyclePath 生成回收站路径
func (s *SambaRecycleService) generateRecyclePath(recycleBinPath, originalPath string) (string, error) {
	// 确保回收站目录存在
	if err := os.MkdirAll(recycleBinPath, 0755); err != nil {
		return "", fmt.Errorf("创建回收站目录失败: %w", err)
	}

	// 生成唯一的回收站文件名
	fileName := filepath.Base(originalPath)
	timestamp := time.Now().Format("20060102-150405")
	recycleName := fmt.Sprintf("%s_%s_%s", fileName, timestamp, uuid.New().String()[:8])

	return filepath.Join(recycleBinPath, recycleName), nil
}

// moveToRecycleBin 移动文件到回收站
func (s *SambaRecycleService) moveToRecycleBin(originalPath, recyclePath string) error {
	// 确保回收站目录存在
	recycleDir := filepath.Dir(recyclePath)
	if err := os.MkdirAll(recycleDir, 0755); err != nil {
		return fmt.Errorf("创建回收站目录失败: %w", err)
	}

	// 移动文件
	if err := os.Rename(originalPath, recyclePath); err != nil {
		return fmt.Errorf("移动文件失败: %w", err)
	}

	return nil
}

// restoreFromRecycleBin 从回收站恢复文件
func (s *SambaRecycleService) restoreFromRecycleBin(recyclePath, restorePath string) error {
	// 确保恢复目录存在
	restoreDir := filepath.Dir(restorePath)
	if err := os.MkdirAll(restoreDir, 0755); err != nil {
		return fmt.Errorf("创建恢复目录失败: %w", err)
	}

	// 移动文件
	if err := os.Rename(recyclePath, restorePath); err != nil {
		return fmt.Errorf("恢复文件失败: %w", err)
	}

	return nil
}

// permanentDelete 永久删除文件
func (s *SambaRecycleService) permanentDelete(path string) error {
	return os.RemoveAll(path)
}

// generateUniqueRestorePath 生成唯一的恢复路径
func (s *SambaRecycleService) generateUniqueRestorePath(originalPath string) string {
	dir := filepath.Dir(originalPath)
	fileName := filepath.Base(originalPath)
	ext := filepath.Ext(fileName)
	nameWithoutExt := strings.TrimSuffix(fileName, ext)

	for i := 1; ; i++ {
		newName := fmt.Sprintf("%s_restored_%d%s", nameWithoutExt, i, ext)
		newPath := filepath.Join(dir, newName)
		if _, err := os.Stat(newPath); os.IsNotExist(err) {
			return newPath
		}
	}
}

// getFileSize 获取文件大小
func (s *SambaRecycleService) getFileSize(info os.FileInfo) int64 {
	if info.IsDir() {
		// 对于目录，可能需要递归计算大小，这里简化处理
		return 0
	}
	return info.Size()
}

// getFileType 获取文件类型
func (s *SambaRecycleService) getFileType(path string) string {
	ext := filepath.Ext(path)
	if ext == "" {
		return "unknown"
	}
	return strings.ToLower(ext[1:]) // 去掉点号
}

// toRecycleBinItemResponse 转换为回收站条目响应格式
func (s *SambaRecycleService) toRecycleBinItemResponse(item *models.RecycleBinItem) *dto.SambaRecycleBinItemResponse {
	response := &dto.SambaRecycleBinItemResponse{
		ID:           item.ID,
		ShareID:      item.ShareID,
		OriginalPath: item.OriginalPath,
		RecyclePath:  item.RecyclePath,
		DeletedBy:    item.DeletedBy,
		DeletedAt:    item.DeletedAt,
		FileSize:     item.FileSize,
		FileType:     item.FileType,
		IsDirectory:  item.IsDirectory,
		ExpiresAt:    item.ExpiresAt,
		CreatedAt:    item.CreatedAt,
		UpdatedAt:    item.UpdatedAt,
	}

	// 添加共享信息
	if item.Share.ID != "" {
		response.Share = &dto.SambaShareResponse{
			ID:   item.Share.ID,
			Name: item.Share.Name,
			Path: item.Share.Path,
		}
	}

	return response
}

// CreateRecycleBinCleanupJob 创建回收站清理任务
func (s *SambaRecycleService) CreateRecycleBinCleanupJob() {
	// 这里可以集成定时任务框架，比如cron
	// 示例：每天凌晨2点清理过期条目
	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := s.CleanupExpiredItems(); err != nil {
					fmt.Printf("清理过期回收站条目失败: %v\n", err)
				}
			}
		}
	}()
}
