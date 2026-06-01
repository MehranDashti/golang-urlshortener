package middleware

import (
    "net/http"
    "strings"

    "github.com/gin-gonic/gin"
    "urlshortener/pkg/token"
)

const UserIDKey = "userID"

func Auth(tokenManager *token.Manager) gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization header required"})
            c.Abort() 
            return
        }

        parts := strings.SplitN(authHeader, " ", 2)
        if len(parts) != 2 || parts[0] != "Bearer" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization format"})
            c.Abort()
            return
        }

        claims, err := tokenManager.Validate(parts[1])
        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
            c.Abort()
            return
        }

        c.Set(UserIDKey, claims.UserID)
        c.Next() 
    }
}