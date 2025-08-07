package services

import (
	"errors"
	"pnas/internal/config"
	"pnas/internal/models"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

// JWTClaims JWT声明结构体
type JWTClaims struct {
	UserID   uint             `json:"user_id"`
	Username string           `json:"username"`
	UserType models.UserType  `json:"user_type"`
	GroupID  uint             `json:"group_id"`
	jwt.RegisteredClaims
}

// JWTService JWT服务
type JWTService struct {
	config *config.JWTConfig
	logger *zap.Logger
}

// NewJWTService 创建JWT服务实例
func NewJWTService(config *config.JWTConfig, logger *zap.Logger) *JWTService {
	return &JWTService{
		config: config,
		logger: logger,
	}
}

// GenerateToken 生成JWT Token
func (s *JWTService) GenerateToken(user *models.User) (string, error) {
	now := time.Now()
	claims := JWTClaims{
		UserID:   user.ID,
		Username: user.Username,
		UserType: user.UserType,
		GroupID:  user.GroupID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.config.JWT.Issuer,
			Subject:   user.Username,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.config.GetExpirationTime())),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.config.JWT.SecretKey))
	if err != nil {
		s.logger.Error("生成JWT Token失败", 
			zap.String("username", user.Username),
			zap.Error(err),
		)
		return "", err
	}

	s.logger.Debug("JWT Token生成成功", 
		zap.String("username", user.Username),
		zap.Uint("user_id", user.ID),
		zap.Int("user_type", int(user.UserType)),
	)

	return tokenString, nil
}

// ValidateToken 验证JWT Token
func (s *JWTService) ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("无效的签名方法")
		}
		return []byte(s.config.JWT.SecretKey), nil
	})

	if err != nil {
		s.logger.Debug("JWT Token验证失败", zap.Error(err))
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		s.logger.Debug("JWT Token验证成功", 
			zap.String("username", claims.Username),
			zap.Uint("user_id", claims.UserID),
		)
		return claims, nil
	}

	return nil, errors.New("无效的Token")
}

// RefreshToken 刷新JWT Token
func (s *JWTService) RefreshToken(tokenString string) (string, error) {
	claims, err := s.ValidateToken(tokenString)
	if err != nil {
		return "", err
	}

	// 检查Token是否即将过期（在过期前1小时内可以刷新）
	if time.Until(claims.ExpiresAt.Time) > time.Hour {
		return "", errors.New("Token尚未到刷新时间")
	}

	// 创建新的Token
	now := time.Now()
	newClaims := JWTClaims{
		UserID:   claims.UserID,
		Username: claims.Username,
		UserType: claims.UserType,
		GroupID:  claims.GroupID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.config.JWT.Issuer,
			Subject:   claims.Username,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.config.GetExpirationTime())),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, newClaims)
	newTokenString, err := token.SignedString([]byte(s.config.JWT.SecretKey))
	if err != nil {
		s.logger.Error("刷新JWT Token失败", 
			zap.String("username", claims.Username),
			zap.Error(err),
		)
		return "", err
	}

	s.logger.Info("JWT Token刷新成功", 
		zap.String("username", claims.Username),
		zap.Uint("user_id", claims.UserID),
	)

	return newTokenString, nil
}