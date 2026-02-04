package middleware

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/GTDGit/PPOB_BE/internal/domain"
	"github.com/GTDGit/PPOB_BE/pkg/redis"
)

const (
	TempTokenHeader = "Authorization"
	TempTokenKey    = "temp_token"
)

// TempTokenAuth returns a middleware that validates temporary tokens
func TempTokenAuth(redisClient *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get Authorization header
		authHeader := c.GetHeader(TempTokenHeader)
		if authHeader == "" {
			respondError(c, domain.ErrInvalidTokenError)
			c.Abort()
			return
		}

		// Parse Bearer token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			respondError(c, domain.ErrInvalidTokenError)
			c.Abort()
			return
		}

		tempToken := parts[1]

		// Get temp token from Redis
		key := redis.TempTokenKey(tempToken)
		var tokenData domain.TempToken
		if err := redisClient.GetJSON(c.Request.Context(), key, &tokenData); err != nil {
			respondError(c, domain.ErrInvalidTokenError)
			c.Abort()
			return
		}

		// Check expiration
		if time.Now().Unix() > tokenData.ExpiresAt {
			redisClient.Del(c.Request.Context(), key)
			respondError(c, domain.ErrTempTokenExpiredError)
			c.Abort()
			return
		}

		// Set token data in context
		c.Set(TempTokenKey, &tokenData)

		c.Next()
	}
}

// GetTempToken returns temp token data from context
func GetTempToken(c *gin.Context) *domain.TempToken {
	if tokenData, exists := c.Get(TempTokenKey); exists {
		return tokenData.(*domain.TempToken)
	}
	return nil
}
