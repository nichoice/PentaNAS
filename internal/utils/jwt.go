package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims represents the JWT claims
type Claims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// JWTConfig holds the JWT configuration
type JWTConfig struct {
	Secret string
	Expire int
}

var jwtConfig *JWTConfig

// InitJWT initializes the JWT configuration
func InitJWT(secret string, expire int) {
	jwtConfig = &JWTConfig{
		Secret: secret,
		Expire: expire,
	}
}

// GenerateToken generates a JWT token for a user
func GenerateToken(userID, username string) (string, error) {
	if jwtConfig == nil {
		// Default configuration
		jwtConfig = &JWTConfig{
			Secret: "pnas-secret-key",
			Expire: 24,
		}
	}
	
	// Get expiration time from config
	expireHours := jwtConfig.Expire
	if expireHours <= 0 {
		expireHours = 24 // Default to 24 hours
	}
	
	expirationTime := time.Now().Add(time.Duration(expireHours) * time.Hour)
	
	claims := &Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "pnas",
		},
	}
	
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	
	// Get secret key from config
	secretKey := jwtConfig.Secret
	if secretKey == "" {
		secretKey = "pnas-secret-key" // Default secret
	}
	
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}
	
	return tokenString, nil
}

// ParseToken parses and validates a JWT token
func ParseToken(tokenString string) (*Claims, error) {
	if jwtConfig == nil {
		// Default configuration
		jwtConfig = &JWTConfig{
			Secret: "pnas-secret-key",
			Expire: 24,
		}
	}
	
	// Get secret key from config
	secretKey := jwtConfig.Secret
	if secretKey == "" {
		secretKey = "pnas-secret-key" // Default secret
	}
	
	claims := &Claims{}
	
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})
	
	if err != nil {
		return nil, err
	}
	
	if !token.Valid {
		return nil, errors.New("invalid token")
	}
	
	return claims, nil
}