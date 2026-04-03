package middleware

import (
	"strings"

	"github.com/GTDGit/PPOB_BE/internal/domain"
	"github.com/GTDGit/PPOB_BE/internal/repository"
	"github.com/GTDGit/PPOB_BE/pkg/jwt"
	"github.com/gin-gonic/gin"
)

const (
	AdminIDKey          = "admin_id"
	AdminSessionIDKey   = "admin_session_id"
	AdminPermissionsKey = "admin_permissions"
	AdminUserKey        = "admin_user"
)

func AdminJWTAuth(secret string, adminRepo *repository.AdminRepository) gin.HandlerFunc {
	jwtGen := jwt.NewGenerator(secret, 0, 0)

	return func(c *gin.Context) {
		authHeader := c.GetHeader(AuthorizationHeader)
		if authHeader == "" {
			c.JSON(401, domain.ErrorResponse(domain.NewError("ADMIN_UNAUTHORIZED", "Akses admin tidak diizinkan", 401), GetRequestID(c)))
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != BearerSchema {
			c.JSON(401, domain.ErrorResponse(domain.NewError("ADMIN_UNAUTHORIZED", "Format token admin tidak valid", 401), GetRequestID(c)))
			c.Abort()
			return
		}

		claims, err := jwtGen.ValidateAccessToken(parts[1])
		if err != nil {
			c.JSON(401, domain.ErrorResponse(domain.NewError("ADMIN_TOKEN_INVALID", "Token admin tidak valid", 401), GetRequestID(c)))
			c.Abort()
			return
		}

		session, err := adminRepo.FindSessionByID(c.Request.Context(), claims.DeviceID)
		if err != nil || session == nil || session.AdminUserID != claims.UserID {
			c.JSON(401, domain.ErrorResponse(domain.NewError("ADMIN_SESSION_INVALID", "Sesi admin tidak valid", 401), GetRequestID(c)))
			c.Abort()
			return
		}

		admin, err := adminRepo.FindAdminByID(c.Request.Context(), claims.UserID)
		if err != nil || admin == nil || !admin.IsActive || admin.Status != domain.AdminStatusActive {
			c.JSON(403, domain.ErrorResponse(domain.NewError("ADMIN_DISABLED", "Akun admin tidak aktif", 403), GetRequestID(c)))
			c.Abort()
			return
		}

		_ = adminRepo.TouchSession(c.Request.Context(), session.ID)
		c.Set(AdminIDKey, admin.ID)
		c.Set(AdminSessionIDKey, session.ID)
		c.Set(AdminPermissionsKey, admin.Permissions)
		c.Set(AdminUserKey, admin)
		c.Next()
	}
}

func AdminRequirePermissions(required ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		rawPermissions, exists := c.Get(AdminPermissionsKey)
		if !exists {
			c.JSON(403, domain.ErrorResponse(domain.NewError("ADMIN_FORBIDDEN", "Permission admin tidak tersedia", 403), GetRequestID(c)))
			c.Abort()
			return
		}

		permissions, ok := rawPermissions.([]string)
		if !ok {
			c.JSON(403, domain.ErrorResponse(domain.NewError("ADMIN_FORBIDDEN", "Permission admin tidak valid", 403), GetRequestID(c)))
			c.Abort()
			return
		}

		owned := map[string]bool{}
		for _, permission := range permissions {
			owned[permission] = true
		}

		for _, permission := range required {
			if !owned[permission] {
				c.JSON(403, domain.ErrorResponse(domain.NewError("ADMIN_FORBIDDEN", "Anda tidak memiliki akses ke fitur ini", 403), GetRequestID(c)))
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

func GetAdminID(c *gin.Context) string {
	value, ok := c.Get(AdminIDKey)
	if !ok {
		return ""
	}
	if adminID, ok := value.(string); ok {
		return adminID
	}
	return ""
}

func GetAdminSessionID(c *gin.Context) string {
	value, ok := c.Get(AdminSessionIDKey)
	if !ok {
		return ""
	}
	if sessionID, ok := value.(string); ok {
		return sessionID
	}
	return ""
}
