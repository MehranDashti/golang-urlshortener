package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"urlshortener/internal/tokenstore"
	"urlshortener/pkg/token"
	"urlshortener/internal/apperror"
)

const UserIDKey = "userID"

// Note: UserRoleKey is declared in admin.go — same package, no redeclaration needed

func Auth(
	tokenManager *token.Manager,
	blacklist tokenstore.TokenBlacklist) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, map[string]interface{}{
				"success": false,
				"code":    401,
				"message": "authorization header required",
			})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, map[string]interface{}{
				"success": false,
				"code":    401,
				"message": "invalid authorization format",
			})
			c.Abort()
			return
		}

		claims, err := tokenManager.Validate(parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, map[string]interface{}{
				"success": false,
				"code":    401,
				"message": "invalid or expired token",
			})
			c.Abort()
			return
		}

		if claims.TokenType != token.AccessToken {
			c.JSON(http.StatusUnauthorized, map[string]interface{}{
				"success": false,
				"code":    401,
				"message": "invalid token type",
			})
			c.Abort()
			return
		}

		// Check blacklist — token revoked on logout?
		revoked, err := blacklist.IsRevoked(c.Request.Context(), claims.ID)
		if err != nil || revoked {
			_ = c.Error(apperror.Unauthorized("token has been revoked"))
			c.Abort()
			return
		}

		c.Set(UserIDKey, claims.UserID)
		c.Set(UserRoleKey, claims.Role)
		c.Next()
	}
}
