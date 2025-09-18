package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

var (
	requestCounts = make(map[string]int)
	mu            sync.Mutex
)

func RateLimitMiddleware(maxRequests int, duration time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		mu.Lock()
		defer mu.Unlock()

		count, exists := requestCounts[clientIP]
		if !exists {
			requestCounts[clientIP] = 1
			go func() {
				time.Sleep(duration)
				mu.Lock()
				delete(requestCounts, clientIP)
				mu.Unlock()
			}()
		} else if count >= maxRequests {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"message": "Too many requests",
			})
			return
		} else {
			requestCounts[clientIP]++
		}

		c.Next()
	}
}
