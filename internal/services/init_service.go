package services

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"pnas/internal/database"
	"pnas/internal/models"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const (
	DefaultAdminUsername = "admin"
	DefaultAdminPassword = "admin123456"
)

// InitService handles first-time initialization
type InitService struct {
	db *gorm.DB
}

// NewInitService creates a new init service
func NewInitService() *InitService {
	return &InitService{
		db: database.DB,
	}
}

// InitializeSystem performs first-time system initialization
func (s *InitService) InitializeSystem() error {
	log.Println("开始检查系统初始化状态...")

	// 检查并创建默认角色
	if err := s.createDefaultRoles(); err != nil {
		return fmt.Errorf("创建默认角色失败: %w", err)
	}

	// 检查并创建默认超级管理员
	if err := s.createDefaultAdmin(); err != nil {
		return fmt.Errorf("创建默认管理员失败: %w", err)
	}

	log.Println("系统初始化检查完成")
	return nil
}

// createDefaultRoles creates default roles if they don't exist
func (s *InitService) createDefaultRoles() error {
	defaultRoles := []models.Role{
		{
			Base:        models.Base{ID: s.generateID()},
			Name:        models.RoleSuperAdmin,
			Description: "超级管理员，拥有系统所有权限",
		},
		{
			Base:        models.Base{ID: s.generateID()},
			Name:        models.RoleNormalUser,
			Description: "普通用户，拥有基本功能权限",
		},
		{
			Base:        models.Base{ID: s.generateID()},
			Name:        models.RoleAuditUser,
			Description: "审计用户，拥有审计查看权限",
		},
		{
			Base:        models.Base{ID: s.generateID()},
			Name:        models.RoleOpsUser,
			Description: "运维用户，拥有系统运维权限",
		},
	}

	for _, role := range defaultRoles {
		// 检查角色是否已存在
		var existingRole models.Role
		if err := s.db.Where("name = ?", role.Name).First(&existingRole).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				// 角色不存在，创建新角色
				if err := s.db.Create(&role).Error; err != nil {
					return fmt.Errorf("创建角色 %s 失败: %w", role.Name, err)
				}
				log.Printf("成功创建默认角色: %s", role.Name)
			} else {
				return fmt.Errorf("查询角色 %s 失败: %w", role.Name, err)
			}
		} else {
			log.Printf("角色 %s 已存在，跳过创建", role.Name)
		}
	}

	return nil
}

// createDefaultAdmin creates default admin user if no users exist
func (s *InitService) createDefaultAdmin() error {
	// 检查是否已有用户存在
	var userCount int64
	if err := s.db.Model(&models.User{}).Count(&userCount).Error; err != nil {
		return fmt.Errorf("查询用户数量失败: %w", err)
	}

	if userCount > 0 {
		log.Printf("系统中已存在 %d 个用户，跳过默认管理员创建", userCount)
		return nil
	}

	// 生成密码哈希
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(DefaultAdminPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("生成密码哈希失败: %w", err)
	}

	// 创建默认管理员用户
	adminUser := models.User{
		Base:     models.Base{ID: s.generateID()},
		Username: DefaultAdminUsername,
		Password: string(hashedPassword),
		IsActive: true,
		Remark:   "系统默认超级管理员账号",
	}

	if err := s.db.Create(&adminUser).Error; err != nil {
		return fmt.Errorf("创建默认管理员用户失败: %w", err)
	}

	// 获取超级管理员角色
	var superAdminRole models.Role
	if err := s.db.Where("name = ?", models.RoleSuperAdmin).First(&superAdminRole).Error; err != nil {
		return fmt.Errorf("查询超级管理员角色失败: %w", err)
	}

	// 分配超级管理员角色给默认用户
	userRole := models.UserRole{
		Base:   models.Base{ID: s.generateID()},
		UserID: adminUser.ID,
		RoleID: superAdminRole.ID,
	}

	if err := s.db.Create(&userRole).Error; err != nil {
		return fmt.Errorf("分配角色失败: %w", err)
	}

	log.Printf("成功创建默认超级管理员账号: %s (密码: %s)", DefaultAdminUsername, DefaultAdminPassword)
	log.Println("警告: 请在首次登录后立即修改默认密码!")

	return nil
}

// generateID generates a simple ID (you might want to use UUID in production)
func (s *InitService) generateID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}