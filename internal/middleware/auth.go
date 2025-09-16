package middleware

import (
	"net/http"
	"pnas/internal/utils"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware is a middleware for JWT authentication
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization header is required"})
			c.Abort()
			return
		}

		// Check if the header starts with "Bearer "
		if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization header must start with 'Bearer '"})
			c.Abort()
			return
		}

		// Extract token
		tokenString := authHeader[7:]

		// Parse and validate token
		claims, err := utils.ParseToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		// Set user information in context
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)

		// Continue with the next middleware or handler
		c.Next()
	}
}

// OptionalAuthMiddleware is a middleware for optional JWT authentication
// It will set user information in context if token is valid, but won't reject requests without valid token
func OptionalAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// No token, continue without setting user info
			c.Next()
			return
		}

		// Check if the header starts with "Bearer "
		if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
			// Invalid format, continue without setting user info
			c.Next()
			return
		}

		// Extract token
		tokenString := authHeader[7:]

		// Parse and validate token
		claims, err := utils.ParseToken(tokenString)
		if err != nil {
			// Invalid token, continue without setting user info
			c.Next()
			return
		}

		// Set user information in context
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)

		// Continue with the next middleware or handler
		c.Next()
	}
}