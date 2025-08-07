package repositories

import (
	"pnas/internal/models"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// healthCheckRepository 健康检查仓库实现
type healthCheckRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewHealthCheckRepository 创建健康检查仓库实例
func NewHealthCheckRepository(db *gorm.DB, logger *zap.Logger) HealthCheckRepository {
	return &healthCheckRepository{
		db:     db,
		logger: logger,
	}
}

// Create 创建健康检查记录
func (r *healthCheckRepository) Create(healthCheck *models.HealthCheck) error {
	if err := r.db.Create(healthCheck).Error; err != nil {
		r.logger.Error("创建健康检查记录失败", 
			zap.String("client_ip", healthCheck.ClientIP),
			zap.Error(err),
		)
		return err
	}

	r.logger.Debug("健康检查记录创建成功", 
		zap.Uint("id", healthCheck.ID),
		zap.String("client_ip", healthCheck.ClientIP),
	)
	return nil
}

// GetByID 根据ID获取健康检查记录
func (r *healthCheckRepository) GetByID(id uint) (*models.HealthCheck, error) {
	var healthCheck models.HealthCheck
	if err := r.db.First(&healthCheck, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			r.logger.Debug("健康检查记录不存在", zap.Uint("id", id))
		} else {
			r.logger.Error("查询健康检查记录失败", zap.Uint("id", id), zap.Error(err))
		}
		return nil, err
	}
	return &healthCheck, nil
}

// List 获取健康检查记录列表
func (r *healthCheckRepository) List(offset, limit int) ([]*models.HealthCheck, int64, error) {
	var healthChecks []*models.HealthCheck
	var total int64

	// 获取总数
	if err := r.db.Model(&models.HealthCheck{}).Count(&total).Error; err != nil {
		r.logger.Error("查询健康检查记录总数失败", zap.Error(err))
		return nil, 0, err
	}

	// 获取记录列表
	if err := r.db.Order("created_at DESC").Offset(offset).Limit(limit).Find(&healthChecks).Error; err != nil {
		r.logger.Error("查询健康检查记录列表失败", zap.Error(err))
		return nil, 0, err
	}

	r.logger.Debug("健康检查记录列表查询成功", 
		zap.Int("count", len(healthChecks)),
		zap.Int64("total", total),
	)

	return healthChecks, total, nil
}

// GetStats 获取健康检查统计信息
func (r *healthCheckRepository) GetStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 总记录数
	var total int64
	if err := r.db.Model(&models.HealthCheck{}).Count(&total).Error; err != nil {
		r.logger.Error("查询健康检查总数失败", zap.Error(err))
		return nil, err
	}
	stats["total"] = total

	// 今日记录数
	var todayCount int64
	if err := r.db.Model(&models.HealthCheck{}).
		Where("DATE(created_at) = CURRENT_DATE").
		Count(&todayCount).Error; err != nil {
		r.logger.Error("查询今日健康检查数失败", zap.Error(err))
		return nil, err
	}
	stats["today"] = todayCount

	// 按状态统计
	var statusStats []struct {
		Status string `json:"status"`
		Count  int64  `json:"count"`
	}
	if err := r.db.Model(&models.HealthCheck{}).
		Select("status, COUNT(*) as count").
		Group("status").
		Scan(&statusStats).Error; err != nil {
		r.logger.Error("查询状态统计失败", zap.Error(err))
		return nil, err
	}
	stats["by_status"] = statusStats

	r.logger.Debug("健康检查统计查询成功", zap.Any("stats", stats))
	return stats, nil
}