package repositories

import (
	"pnas/internal/models"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// userGroupRepository 用户组仓库实现
type userGroupRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewUserGroupRepository 创建用户组仓库实例
func NewUserGroupRepository(db *gorm.DB, logger *zap.Logger) UserGroupRepository {
	return &userGroupRepository{
		db:     db,
		logger: logger,
	}
}

// Create 创建用户组
func (r *userGroupRepository) Create(group *models.UserGroup) error {
	if err := r.db.Create(group).Error; err != nil {
		r.logger.Error("创建用户组失败", 
			zap.String("group_name", group.Name),
			zap.Error(err),
		)
		return err
	}

	r.logger.Info("用户组创建成功", 
		zap.Uint("group_id", group.ID),
		zap.String("group_name", group.Name),
	)
	return nil
}

// GetByID 根据ID获取用户组
func (r *userGroupRepository) GetByID(id uint) (*models.UserGroup, error) {
	var group models.UserGroup
	if err := r.db.First(&group, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			r.logger.Debug("用户组不存在", zap.Uint("group_id", id))
		} else {
			r.logger.Error("查询用户组失败", zap.Uint("group_id", id), zap.Error(err))
		}
		return nil, err
	}
	return &group, nil
}

// GetByName 根据名称获取用户组
func (r *userGroupRepository) GetByName(name string) (*models.UserGroup, error) {
	var group models.UserGroup
	if err := r.db.Where("name = ?", name).First(&group).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			r.logger.Debug("用户组不存在", zap.String("group_name", name))
		} else {
			r.logger.Error("查询用户组失败", zap.String("group_name", name), zap.Error(err))
		}
		return nil, err
	}
	return &group, nil
}

// Update 更新用户组
func (r *userGroupRepository) Update(group *models.UserGroup) error {
	if err := r.db.Save(group).Error; err != nil {
		r.logger.Error("更新用户组失败", 
			zap.Uint("group_id", group.ID),
			zap.Error(err),
		)
		return err
	}

	r.logger.Info("用户组更新成功", 
		zap.Uint("group_id", group.ID),
		zap.String("group_name", group.Name),
	)
	return nil
}

// Delete 删除用户组（软删除）
func (r *userGroupRepository) Delete(id uint) error {
	if err := r.db.Delete(&models.UserGroup{}, id).Error; err != nil {
		r.logger.Error("删除用户组失败", zap.Uint("group_id", id), zap.Error(err))
		return err
	}

	r.logger.Info("用户组删除成功", zap.Uint("group_id", id))
	return nil
}

// List 获取用户组列表
func (r *userGroupRepository) List(offset, limit int) ([]*models.UserGroup, int64, error) {
	var groups []*models.UserGroup
	var total int64

	// 获取总数
	if err := r.db.Model(&models.UserGroup{}).Count(&total).Error; err != nil {
		r.logger.Error("查询用户组总数失败", zap.Error(err))
		return nil, 0, err
	}

	// 获取用户组列表
	if err := r.db.Offset(offset).Limit(limit).Find(&groups).Error; err != nil {
		r.logger.Error("查询用户组列表失败", zap.Error(err))
		return nil, 0, err
	}

	r.logger.Debug("用户组列表查询成功", 
		zap.Int("count", len(groups)),
		zap.Int64("total", total),
	)

	return groups, total, nil
}

// GetWithUsers 获取用户组及其用户列表
func (r *userGroupRepository) GetWithUsers(id uint) (*models.UserGroup, error) {
	var group models.UserGroup
	if err := r.db.Preload("Users").First(&group, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			r.logger.Debug("用户组不存在", zap.Uint("group_id", id))
		} else {
			r.logger.Error("查询用户组及用户失败", zap.Uint("group_id", id), zap.Error(err))
		}
		return nil, err
	}

	r.logger.Debug("用户组及用户查询成功", 
		zap.Uint("group_id", group.ID),
		zap.String("group_name", group.Name),
		zap.Int("user_count", len(group.Users)),
	)

	return &group, nil
}