package services

import (
	"crypto/md5"
	"crypto/rand"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"pnas/internal/app/dto"
	"pnas/internal/models"
)

type SambaAccountService struct {
	db *gorm.DB
}

func NewSambaAccountService(db *gorm.DB) *SambaAccountService {
	return &SambaAccountService{db: db}
}

// CreateAccount 创建Samba账号
func (s *SambaAccountService) CreateAccount(req *dto.CreateSambaAccountRequest) (*dto.SambaAccountResponse, error) {
	// 检查用户是否存在
	var user models.User
	if err := s.db.First(&user, "id = ?", req.UserID).Error; err != nil {
		return nil, fmt.Errorf("用户不存在: %w", err)
	}

	// 检查Samba用户名是否已存在
	var existingAccount models.SambaAccount
	if err := s.db.Where("samba_user = ?", req.SambaUser).First(&existingAccount).Error; err == nil {
		return nil, fmt.Errorf("Samba用户名 %s 已存在", req.SambaUser)
	}

	// 哈希密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("密码加密失败: %w", err)
	}

	// 创建Samba账号
	account := models.SambaAccount{
		Base:        models.Base{ID: uuid.New().String()},
		UserID:      req.UserID,
		SambaUser:   req.SambaUser,
		Password:    string(hashedPassword),
		Role:        models.SambaAccountRole(req.Role),
		IsEnabled:   true,
		Description: req.Description,
	}

	if req.IsEnabled != nil {
		account.IsEnabled = *req.IsEnabled
	}

	if err := s.db.Create(&account).Error; err != nil {
		return nil, fmt.Errorf("创建Samba账号失败: %w", err)
	}

	// 创建系统Samba用户
	if err := s.createSystemSambaUser(req.SambaUser, req.Password); err != nil {
		// 回滚数据库操作
		s.db.Delete(&account)
		return nil, fmt.Errorf("创建系统Samba用户失败: %w", err)
	}

	return s.GetAccount(account.ID)
}

// UpdateAccount 更新Samba账号
func (s *SambaAccountService) UpdateAccount(accountID string, req *dto.UpdateSambaAccountRequest) (*dto.SambaAccountResponse, error) {
	var account models.SambaAccount
	if err := s.db.First(&account, "id = ?", accountID).Error; err != nil {
		return nil, fmt.Errorf("Samba账号不存在: %w", err)
	}

	updates := make(map[string]interface{})

	// 更新密码
	if req.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, fmt.Errorf("密码加密失败: %w", err)
		}
		updates["password"] = string(hashedPassword)

		// 更新系统Samba用户密码
		if err := s.updateSystemSambaUserPassword(account.SambaUser, req.Password); err != nil {
			return nil, fmt.Errorf("更新系统Samba用户密码失败: %w", err)
		}
	}

	if req.Role != "" {
		updates["role"] = models.SambaAccountRole(req.Role)
	}

	if req.IsEnabled != nil {
		updates["is_enabled"] = *req.IsEnabled
		// 启用/禁用系统Samba用户
		if err := s.enableSystemSambaUser(account.SambaUser, *req.IsEnabled); err != nil {
			return nil, fmt.Errorf("启用/禁用系统Samba用户失败: %w", err)
		}
	}

	if req.Description != "" {
		updates["description"] = req.Description
	}

	if len(updates) > 0 {
		if err := s.db.Model(&account).Updates(updates).Error; err != nil {
			return nil, fmt.Errorf("更新Samba账号失败: %w", err)
		}
	}

	return s.GetAccount(accountID)
}

// GetAccount 获取Samba账号
func (s *SambaAccountService) GetAccount(accountID string) (*dto.SambaAccountResponse, error) {
	var account models.SambaAccount
	if err := s.db.Preload("User").Preload("ShareAccess.Share").First(&account, "id = ?", accountID).Error; err != nil {
		return nil, fmt.Errorf("Samba账号不存在: %w", err)
	}

	return s.toAccountResponse(&account), nil
}

// GetAccountByUsername 根据用户名获取Samba账号
func (s *SambaAccountService) GetAccountByUsername(username string) (*dto.SambaAccountResponse, error) {
	var account models.SambaAccount
	if err := s.db.Preload("User").Preload("ShareAccess.Share").Where("samba_user = ?", username).First(&account).Error; err != nil {
		return nil, fmt.Errorf("Samba账号不存在: %w", err)
	}

	return s.toAccountResponse(&account), nil
}

// ListAccounts 获取Samba账号列表
func (s *SambaAccountService) ListAccounts(offset, limit int, role string) ([]*dto.SambaAccountResponse, int64, error) {
	var accounts []models.SambaAccount
	var total int64

	query := s.db.Model(&models.SambaAccount{})
	if role != "" {
		query = query.Where("role = ?", role)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("统计Samba账号数量失败: %w", err)
	}

	if err := query.Preload("User").Preload("ShareAccess.Share").
		Offset(offset).Limit(limit).Find(&accounts).Error; err != nil {
		return nil, 0, fmt.Errorf("获取Samba账号列表失败: %w", err)
	}

	responses := make([]*dto.SambaAccountResponse, len(accounts))
	for i, account := range accounts {
		responses[i] = s.toAccountResponse(&account)
	}

	return responses, total, nil
}

// DeleteAccount 删除Samba账号
func (s *SambaAccountService) DeleteAccount(accountID string) error {
	var account models.SambaAccount
	if err := s.db.First(&account, "id = ?", accountID).Error; err != nil {
		return fmt.Errorf("Samba账号不存在: %w", err)
	}

	// 删除系统Samba用户
	if err := s.deleteSystemSambaUser(account.SambaUser); err != nil {
		return fmt.Errorf("删除系统Samba用户失败: %w", err)
	}

	// 删除相关的访问权限
	if err := s.db.Where("account_id = ?", accountID).Delete(&models.SambaShareAccess{}).Error; err != nil {
		return fmt.Errorf("删除Samba账号访问权限失败: %w", err)
	}

	// 删除账号
	if err := s.db.Delete(&account).Error; err != nil {
		return fmt.Errorf("删除Samba账号失败: %w", err)
	}

	return nil
}

// ValidateCredentials 验证Samba账号凭据
func (s *SambaAccountService) ValidateCredentials(username, password string) (*dto.SambaAccountResponse, error) {
	var account models.SambaAccount
	if err := s.db.Preload("User").Where("samba_user = ? AND is_enabled = ?", username, true).First(&account).Error; err != nil {
		return nil, fmt.Errorf("用户名或密码错误")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(account.Password), []byte(password)); err != nil {
		return nil, fmt.Errorf("用户名或密码错误")
	}

	// 更新最后登录时间
	now := time.Now()
	s.db.Model(&account).Update("last_login", &now)

	return s.toAccountResponse(&account), nil
}

// GetAccountsByRole 根据角色获取账号列表
func (s *SambaAccountService) GetAccountsByRole(role models.SambaAccountRole) ([]*dto.SambaAccountResponse, error) {
	var accounts []models.SambaAccount
	if err := s.db.Preload("User").Where("role = ? AND is_enabled = ?", role, true).Find(&accounts).Error; err != nil {
		return nil, fmt.Errorf("获取角色账号列表失败: %w", err)
	}

	responses := make([]*dto.SambaAccountResponse, len(accounts))
	for i, account := range accounts {
		responses[i] = s.toAccountResponse(&account)
	}

	return responses, nil
}

// 系统Samba用户管理方法

// createSystemSambaUser 创建系统Samba用户
func (s *SambaAccountService) createSystemSambaUser(username, password string) error {
	// 创建Linux用户（如果不存在）
	if err := s.createLinuxUser(username); err != nil {
		return fmt.Errorf("创建Linux用户失败: %w", err)
	}

	// 添加到Samba用户数据库
	cmd := exec.Command("smbpasswd", "-a", "-s", username)
	cmd.Stdin = strings.NewReader(password + "\n" + password + "\n")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("添加Samba用户失败: %s, %w", string(output), err)
	}

	return nil
}

// updateSystemSambaUserPassword 更新系统Samba用户密码
func (s *SambaAccountService) updateSystemSambaUserPassword(username, password string) error {
	cmd := exec.Command("smbpasswd", "-s", username)
	cmd.Stdin = strings.NewReader(password + "\n" + password + "\n")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("更新Samba用户密码失败: %s, %w", string(output), err)
	}
	return nil
}

// enableSystemSambaUser 启用/禁用系统Samba用户
func (s *SambaAccountService) enableSystemSambaUser(username string, enabled bool) error {
	var cmd *exec.Cmd
	if enabled {
		cmd = exec.Command("smbpasswd", "-e", username)
	} else {
		cmd = exec.Command("smbpasswd", "-d", username)
	}

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("启用/禁用Samba用户失败: %s, %w", string(output), err)
	}
	return nil
}

// deleteSystemSambaUser 删除系统Samba用户
func (s *SambaAccountService) deleteSystemSambaUser(username string) error {
	cmd := exec.Command("smbpasswd", "-x", username)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("删除Samba用户失败: %s, %w", string(output), err)
	}
	return nil
}

// createLinuxUser 创建Linux用户
func (s *SambaAccountService) createLinuxUser(username string) error {
	// 检查用户是否已存在
	cmd := exec.Command("id", username)
	if err := cmd.Run(); err == nil {
		return nil // 用户已存在
	}

	// 创建用户
	cmd = exec.Command("useradd", "-m", "-s", "/bin/false", username)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("创建Linux用户失败: %s, %w", string(output), err)
	}

	return nil
}

// generateRandomPassword 生成随机密码
func (s *SambaAccountService) generateRandomPassword(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*"
	b := make([]byte, length)
	rand.Read(b)
	for i := range b {
		b[i] = charset[b[i]%byte(len(charset))]
	}
	return string(b)
}

// generateNTLMHash 生成NTLM哈希（用于Samba）
func (s *SambaAccountService) generateNTLMHash(password string) string {
	h := md5.New()
	h.Write([]byte(password))
	return fmt.Sprintf("%x", h.Sum(nil))
}

// toAccountResponse 转换为响应格式
func (s *SambaAccountService) toAccountResponse(account *models.SambaAccount) *dto.SambaAccountResponse {
	response := &dto.SambaAccountResponse{
		ID:          account.ID,
		UserID:      account.UserID,
		SambaUser:   account.SambaUser,
		Role:        string(account.Role),
		IsEnabled:   account.IsEnabled,
		Description: account.Description,
		LastLogin:   account.LastLogin,
		CreatedAt:   account.CreatedAt,
		UpdatedAt:   account.UpdatedAt,
	}

	// 添加用户信息
	if account.User.ID != "" {
		response.User = &dto.UserResponse{
			ID:       uint(0), // 需要转换
			Username: account.User.Username,
		}
	}

	// 添加共享访问权限
	if len(account.ShareAccess) > 0 {
		response.ShareAccess = make([]dto.SambaShareAccessResponse, len(account.ShareAccess))
		for i, access := range account.ShareAccess {
			response.ShareAccess[i] = dto.SambaShareAccessResponse{
				ID:         access.ID,
				ShareID:    access.ShareID,
				AccountID:  access.AccountID,
				Permission: access.Permission,
				CreatedAt:  access.CreatedAt,
				UpdatedAt:  access.UpdatedAt,
			}
		}
	}

	return response
}