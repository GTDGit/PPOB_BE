package middleware

import (
	"strings"
	"time"

	"github.com/GTDGit/PPOB_BE/internal/domain"
	"github.com/GTDGit/PPOB_BE/internal/repository"
	"github.com/GTDGit/PPOB_BE/pkg/jwt"
	"github.com/gin-gonic/gin"
)

const (
	AuthorizationHeader = "Authorization"
	BearerSchema        = "Bearer"
	UserIDKey           = "user_id"
	DeviceIDKey         = "device_id"
	ClaimsKey           = "claims"
)

// JWTAuth returns a middleware that validates JWT tokens and ensures the device session is still active.
func JWTAuth(secret string, sessionRepo repository.SessionRepository) gin.HandlerFunc {
	jwtGen := jwt.NewGenerator(secret, 15*time.Minute, 30*24*time.Hour)
	return func(c *gin.Context) {
		// Get authorization header
		authHeader := c.GetHeader(AuthorizationHeader)
		if authHeader == "" {
			respondError(c, domain.ErrUnauthorized)
			c.Abort()
			return
		}

		// Parse Bearer token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != BearerSchema {
			respondError(c, domain.ErrUnauthorized)
			c.Abort()
			return
		}

		tokenString := parts[1]

		// Validate access token
		claims, err := jwtGen.ValidateAccessToken(tokenString)
		if err != nil {
			if err == jwt.ErrExpiredToken {
				respondError(c, domain.ErrTokenExpired)
			} else {
				respondError(c, domain.ErrInvalidToken)
			}
			c.Abort()
			return
		}

		// Enforce active session on every protected request so logout revokes access immediately.
		if sessionRepo != nil {
			session, err := sessionRepo.FindByUserIDAndDeviceID(c.Request.Context(), claims.UserID, claims.DeviceID)
			if err != nil || session == nil {
				respondError(c, domain.ErrUnauthorized)
				c.Abort()
				return
			}
		}

		// Set claims in context
		c.Set(UserIDKey, claims.UserID)
		c.Set(DeviceIDKey, claims.DeviceID)
		c.Set(ClaimsKey, claims)

		c.Next()
	}
}

// GetUserID returns user ID from context
func GetUserID(c *gin.Context) string {
	if userID, exists := c.Get(UserIDKey); exists {
		return userID.(string)
	}
	return ""
}

// GetDeviceID returns device ID from context
func GetDeviceID(c *gin.Context) string {
	if deviceID, exists := c.Get(DeviceIDKey); exists {
		return deviceID.(string)
	}
	return ""
}

// GetClaims returns JWT claims from context
func GetClaims(c *gin.Context) *jwt.Claims {
	if claims, exists := c.Get(ClaimsKey); exists {
		return claims.(*jwt.Claims)
	}
	return nil
}

// respondError sends error response
func respondError(c *gin.Context, appErr *domain.AppError) {
	c.JSON(appErr.HTTPStatus, domain.ErrorResponse(appErr, GetRequestID(c)))
}
