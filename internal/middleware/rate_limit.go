package middleware

import (
	"context"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/GTDGit/PPOB_BE/internal/domain"
	"github.com/GTDGit/PPOB_BE/pkg/redis"
)

func itoa(i int) string {
	return strconv.Itoa(i)
}

// RateLimitConfig holds rate limit configuration
type RateLimitConfig struct {
	Limit  int           // Max requests
	Window time.Duration // Time window
}

// RateLimit returns a middleware that limits request rate
func RateLimit(redisClient *redis.Client, config RateLimitConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		endpoint := c.FullPath()

		key := redis.RateLimitKey(clientIP, endpoint)
		ctx := context.Background()

		// Get current count
		count, err := redisClient.Incr(ctx, key).Result()
		if err != nil {
			// If Redis fails, allow request
			c.Next()
			return
		}

		// Set expiration on first request
		if count == 1 {
			redisClient.Expire(ctx, key, config.Window)
		}

		// Check if over limit
		if count > int64(config.Limit) {
			respondError(c, domain.ErrRateLimitedError)
			c.Abort()
			return
		}

		// Set headers
		c.Header("X-RateLimit-Limit", itoa(config.Limit))
		c.Header("X-RateLimit-Remaining", itoa(config.Limit-int(count)))

		c.Next()
	}
}

// DefaultRateLimitConfig returns default rate limit config
func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		Limit:  100,
		Window: time.Minute,
	}
}

// StrictRateLimitConfig returns strict rate limit for sensitive endpoints
func StrictRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		Limit:  10,
		Window: time.Minute,
	}
}
