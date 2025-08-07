package repositories

import (
	"pnas/internal/models"
)

// UserRepository 用户仓库接口
type UserRepository interface {
	Create(user *models.User) error
	GetByID(id uint) (*models.User, error)
	GetByUsername(username string) (*models.User, error)
	Update(user *models.User) error
	Delete(id uint) error
	List(offset, limit int) ([]*models.User, int64, error)
	GetByUserType(userType models.UserType, offset, limit int) ([]*models.User, int64, error)
	GetByGroupID(groupID uint, offset, limit int) ([]*models.User, int64, error)
}

// UserGroupRepository 用户组仓库接口
type UserGroupRepository interface {
	Create(group *models.UserGroup) error
	GetByID(id uint) (*models.UserGroup, error)
	GetByName(name string) (*models.UserGroup, error)
	Update(group *models.UserGroup) error
	Delete(id uint) error
	List(offset, limit int) ([]*models.UserGroup, int64, error)
	GetWithUsers(id uint) (*models.UserGroup, error)
}

// HealthCheckRepository 健康检查仓库接口
type HealthCheckRepository interface {
	Create(healthCheck *models.HealthCheck) error
	GetByID(id uint) (*models.HealthCheck, error)
	List(offset, limit int) ([]*models.HealthCheck, int64, error)
	GetStats() (map[string]interface{}, error)
}
