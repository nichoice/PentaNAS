package services

import (
	"errors"
	"pnas/internal/models"
	"pnas/internal/repositories"

	"golang.org/x/crypto/bcrypt"
	"go.uber.org/zap"
)

// AuthService 认证服务
type AuthService struct {
	userRepo   repositories.UserRepository
	jwtService *JWTService
	logger     *zap.Logger
}

// NewAuthService 创建认证服务实例
func NewAuthService(userRepo repositories.UserRepository, jwtService *JWTService, logger *zap.Logger) *AuthService {
	return &AuthService{
		userRepo:   userRepo,
		jwtService: jwtService,
		logger:     logger,
	}
}

// Login 用户登录
func (s *AuthService) Login(username, password string) (string, *models.User, error) {
	// 查找用户
	user, err := s.userRepo.GetByUsername(username)
	if err != nil {
		s.logger.Warn("登录失败：用户不存在", 
			zap.String("username", username),
			zap.Error(err),
		)
		return "", nil, errors.New("用户名或密码错误")
	}

	// 检查用户状态
	if user.Status != models.UserStatusActive {
		s.logger.Warn("登录失败：用户状态异常", 
			zap.String("username", username),
			zap.Int("status", int(user.Status)),
		)
		return "", nil, errors.New("用户账户已被禁用或锁定")
	}

	// 检查用户类型：普通用户不能登录
	if user.UserType == models.UserTypeNormal {
		s.logger.Warn("登录失败：普通用户不允许登录", 
			zap.String("username", username),
			zap.Int("user_type", int(user.UserType)),
		)
		return "", nil, errors.New("普通用户不允许登录系统")
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		s.logger.Warn("登录失败：密码错误", 
			zap.String("username", username),
			zap.Error(err),
		)
		return "", nil, errors.New("用户名或密码错误")
	}

	// 生成JWT Token
	token, err := s.jwtService.GenerateToken(user)
	if err != nil {
		s.logger.Error("登录失败：生成Token失败", 
			zap.String("username", username),
			zap.Error(err),
		)
		return "", nil, errors.New("登录失败，请稍后重试")
	}

	s.logger.Info("用户登录成功", 
		zap.String("username", username),
		zap.Uint("user_id", user.ID),
		zap.Int("user_type", int(user.UserType)),
	)

	return token, user, nil
}

// HashPassword 密码哈希
func (s *AuthService) HashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("密码哈希失败", zap.Error(err))
		return "", err
	}
	return string(hashedBytes), nil
}

// ValidateToken 验证Token
func (s *AuthService) ValidateToken(tokenString string) (*JWTClaims, error) {
	return s.jwtService.ValidateToken(tokenString)
}

// RefreshToken 刷新Token
func (s *AuthService) RefreshToken(tokenString string) (string, error) {
	return s.jwtService.RefreshToken(tokenString)
}