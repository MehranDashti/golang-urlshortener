package middleware

import (
    "net/http"

    "github.com/gin-gonic/gin"
    "urlshortener/internal/model"
)

const UserRoleKey = "userRole"

func Admin() gin.HandlerFunc {
    return func(c *gin.Context) {
        role, exists := c.Get(UserRoleKey)
        if !exists {
            c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
            c.Abort()
            return
        }

        if role != model.RoleAdmin {
            c.JSON(http.StatusForbidden, gin.H{"error": "admin access required"})
            c.Abort()
            return
        }

        c.Next()
    }
}