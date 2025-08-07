package database

import (
	"pnas/internal/models"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"golang.org/x/crypto/bcrypt"
)

// AutoMigrate 自动迁移数据库表结构
func AutoMigrate(db *gorm.DB, logger *zap.Logger) error {
	logger.Info("开始数据库迁移...")

	// 定义需要迁移的模型（注意顺序，先迁移被引用的表）
	modelsToMigrate := []interface{}{
		&models.UserGroup{},    // 先迁移用户组
		&models.User{},         // 再迁移用户（引用用户组）
		&models.HealthCheck{},
	}

	// 执行迁移
	for _, model := range modelsToMigrate {
		if err := db.AutoMigrate(model); err != nil {
			logger.Error("数据库迁移失败", 
				zap.String("model", getModelName(model)),
				zap.Error(err),
			)
			return err
		}
		logger.Debug("模型迁移成功", zap.String("model", getModelName(model)))
	}

	logger.Info("数据库迁移完成")
	return nil
}

// getModelName 获取模型名称
func getModelName(model interface{}) string {
	switch model.(type) {
	case *models.User:
		return "User"
	case *models.UserGroup:
		return "UserGroup"
	case *models.HealthCheck:
		return "HealthCheck"
	default:
		return "Unknown"
	}
}

// SeedData 初始化种子数据（三员管理）
func SeedData(db *gorm.DB, logger *zap.Logger) error {
	logger.Info("开始初始化三员管理种子数据...")

	// 检查是否已有用户组数据
	var groupCount int64
	db.Model(&models.UserGroup{}).Count(&groupCount)
	
	if groupCount == 0 {
		// 创建三员管理用户组
		groups := []models.UserGroup{
			{
				Name:        "系统管理员组",
				Description: "负责系统配置、用户管理、系统维护等工作",
				Status:      1,
			},
			{
				Name:        "安全管理员组",
				Description: "负责安全策略制定、权限管理、安全审核等工作",
				Status:      1,
			},
			{
				Name:        "审计管理员组",
				Description: "负责系统审计、日志分析、合规检查等工作",
				Status:      1,
			},
			{
				Name:        "普通用户组",
				Description: "普通业务用户组",
				Status:      1,
			},
		}

		for _, group := range groups {
			if err := db.Create(&group).Error; err != nil {
				logger.Error("创建用户组失败", 
					zap.String("group_name", group.Name),
					zap.Error(err),
				)
				return err
			}
			logger.Info("用户组创建成功", zap.String("group_name", group.Name))
		}
	}

	// 检查是否已有用户数据
	var userCount int64
	db.Model(&models.User{}).Count(&userCount)
	
	if userCount == 0 {
		// 获取用户组ID
		var systemGroup, securityGroup, auditGroup models.UserGroup
		db.Where("name = ?", "系统管理员组").First(&systemGroup)
		db.Where("name = ?", "安全管理员组").First(&securityGroup)
		db.Where("name = ?", "审计管理员组").First(&auditGroup)

		// 创建三员管理默认用户（使用bcrypt哈希密码）
		users := []struct {
			username string
			password string
			userType models.UserType
			groupID  uint
		}{
			{"sysadmin", "admin123", models.UserTypeSystem, systemGroup.ID},
			{"secadmin", "admin123", models.UserTypeSecurity, securityGroup.ID},
			{"auditadmin", "admin123", models.UserTypeAudit, auditGroup.ID},
		}

		for _, u := range users {
			// 使用bcrypt哈希密码
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.password), bcrypt.DefaultCost)
			if err != nil {
				logger.Error("密码哈希失败", 
					zap.String("username", u.username),
					zap.Error(err),
				)
				return err
			}

			user := models.User{
				Username: u.username,
				Password: string(hashedPassword),
				Status:   models.UserStatusActive,
				UserType: u.userType,
				GroupID:  u.groupID,
			}

			if err := db.Create(&user).Error; err != nil {
				logger.Error("创建默认用户失败", 
					zap.String("username", u.username),
					zap.Error(err),
				)
				return err
			}
			logger.Info("默认用户创建成功", 
				zap.String("username", u.username),
				zap.Int("user_type", int(u.userType)),
			)
		}
	}

	logger.Info("三员管理种子数据初始化完成")
	return nil
}
