package middleware

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/xiaowumin-mark/AMLX/service"
)

const (
	CtxUserIDKey = "user_id"
	CtxRoleIDKey = "role_id"
	CtxEmailKey  = "email"
)

type AuthMiddleware struct {
	auth service.AuthService
}

func NewAuth(auth service.AuthService) *AuthMiddleware {
	return &AuthMiddleware{auth: auth}
}

func (m *AuthMiddleware) Required() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractBearerToken(c)
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
			c.Abort()
			return
		}
		claims, err := m.auth.ParseAccessToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}
		userID, err := parseSubject(claims.Subject)
		if err != nil || userID == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}
		c.Set(CtxUserIDKey, userID)
		c.Set(CtxRoleIDKey, claims.RoleID)
		c.Set(CtxEmailKey, claims.Email)
		c.Next()
	}
}

func RequirePermission(svc service.PermissionService, permName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleID, ok := getRoleID(c)
		if !ok || roleID == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}
		allowed, err := svc.HasPermission(c.Request.Context(), roleID, permName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "permission check failed"})
			c.Abort()
			return
		}
		if !allowed {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			c.Abort()
			return
		}
		c.Next()
	}
}

func GetUserID(c *gin.Context) (uint, bool) {
	value, ok := c.Get(CtxUserIDKey)
	if !ok {
		return 0, false
	}
	id, ok := value.(uint)
	return id, ok
}

func extractBearerToken(c *gin.Context) string {
	header := c.GetHeader("Authorization")
	if header == "" {
		return ""
	}
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 {
		return ""
	}
	if !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

func parseSubject(subject string) (uint, error) {
	parsed, err := strconv.ParseUint(subject, 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(parsed), nil
}

func getRoleID(c *gin.Context) (uint, bool) {
	value, ok := c.Get(CtxRoleIDKey)
	if !ok {
		return 0, false
	}
	roleID, ok := value.(uint)
	return roleID, ok
}
