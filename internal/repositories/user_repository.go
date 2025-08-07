package repositories

import (
	"pnas/internal/models"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// userRepository 用户仓库实现
type userRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewUserRepository 创建用户仓库实例
func NewUserRepository(db *gorm.DB, logger *zap.Logger) UserRepository {
	return &userRepository{
		db:     db,
		logger: logger,
	}
}

// Create 创建用户
func (r *userRepository) Create(user *models.User) error {
	if err := r.db.Create(user).Error; err != nil {
		r.logger.Error("创建用户失败", 
			zap.String("username", user.Username),
			zap.Error(err),
		)
		return err
	}

	r.logger.Info("用户创建成功", 
		zap.Uint("user_id", user.ID),
		zap.String("username", user.Username),
	)
	return nil
}

// GetByID 根据ID获取用户
func (r *userRepository) GetByID(id uint) (*models.User, error) {
	var user models.User
	if err := r.db.First(&user, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			r.logger.Debug("用户不存在", zap.Uint("user_id", id))
		} else {
			r.logger.Error("查询用户失败", zap.Uint("user_id", id), zap.Error(err))
		}
		return nil, err
	}
	return &user, nil
}

// GetByUsername 根据用户名获取用户
func (r *userRepository) GetByUsername(username string) (*models.User, error) {
	var user models.User
	if err := r.db.Where("username = ?", username).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			r.logger.Debug("用户不存在", zap.String("username", username))
		} else {
			r.logger.Error("查询用户失败", zap.String("username", username), zap.Error(err))
		}
		return nil, err
	}
	return &user, nil
}

// GetByUserType 根据用户类型获取用户列表
func (r *userRepository) GetByUserType(userType models.UserType, offset, limit int) ([]*models.User, int64, error) {
	var users []*models.User
	var total int64

	// 获取总数
	if err := r.db.Model(&models.User{}).Where("user_type = ?", userType).Count(&total).Error; err != nil {
		r.logger.Error("查询用户类型总数失败", zap.Int("user_type", int(userType)), zap.Error(err))
		return nil, 0, err
	}

	// 获取用户列表
	if err := r.db.Preload("Group").Where("user_type = ?", userType).Offset(offset).Limit(limit).Find(&users).Error; err != nil {
		r.logger.Error("查询用户类型列表失败", zap.Int("user_type", int(userType)), zap.Error(err))
		return nil, 0, err
	}

	r.logger.Debug("用户类型列表查询成功", 
		zap.Int("user_type", int(userType)),
		zap.Int("count", len(users)),
		zap.Int64("total", total),
	)

	return users, total, nil
}

// GetByGroupID 根据用户组ID获取用户列表
func (r *userRepository) GetByGroupID(groupID uint, offset, limit int) ([]*models.User, int64, error) {
	var users []*models.User
	var total int64

	// 获取总数
	if err := r.db.Model(&models.User{}).Where("group_id = ?", groupID).Count(&total).Error; err != nil {
		r.logger.Error("查询用户组总数失败", zap.Uint("group_id", groupID), zap.Error(err))
		return nil, 0, err
	}

	// 获取用户列表
	if err := r.db.Preload("Group").Where("group_id = ?", groupID).Offset(offset).Limit(limit).Find(&users).Error; err != nil {
		r.logger.Error("查询用户组列表失败", zap.Uint("group_id", groupID), zap.Error(err))
		return nil, 0, err
	}

	r.logger.Debug("用户组列表查询成功", 
		zap.Uint("group_id", groupID),
		zap.Int("count", len(users)),
		zap.Int64("total", total),
	)

	return users, total, nil
}

// Update 更新用户
func (r *userRepository) Update(user *models.User) error {
	if err := r.db.Save(user).Error; err != nil {
		r.logger.Error("更新用户失败", 
			zap.Uint("user_id", user.ID),
			zap.Error(err),
		)
		return err
	}

	r.logger.Info("用户更新成功", 
		zap.Uint("user_id", user.ID),
		zap.String("username", user.Username),
	)
	return nil
}

// Delete 删除用户（软删除）
func (r *userRepository) Delete(id uint) error {
	if err := r.db.Delete(&models.User{}, id).Error; err != nil {
		r.logger.Error("删除用户失败", zap.Uint("user_id", id), zap.Error(err))
		return err
	}

	r.logger.Info("用户删除成功", zap.Uint("user_id", id))
	return nil
}

// List 获取用户列表
func (r *userRepository) List(offset, limit int) ([]*models.User, int64, error) {
	var users []*models.User
	var total int64

	// 获取总数
	if err := r.db.Model(&models.User{}).Count(&total).Error; err != nil {
		r.logger.Error("查询用户总数失败", zap.Error(err))
		return nil, 0, err
	}

	// 获取用户列表（包含用户组信息）
	if err := r.db.Preload("Group").Offset(offset).Limit(limit).Find(&users).Error; err != nil {
		r.logger.Error("查询用户列表失败", zap.Error(err))
		return nil, 0, err
	}

	r.logger.Debug("用户列表查询成功", 
		zap.Int("count", len(users)),
		zap.Int64("total", total),
	)

	return users, total, nil
}
