package router

import (
    "github.com/gin-gonic/gin"
    "urlshortener/internal/handler"
    "urlshortener/internal/middleware"
)

func Setup(
    urlHandler *handler.URLHandler,
    authHandler *handler.AuthHandler,
    adminHandler *handler.AdminHandler,
    authMiddleware gin.HandlerFunc,
) *gin.Engine {
    r := gin.Default()

    // Public routes
    r.POST("/auth/signup", authHandler.Signup)
    r.POST("/auth/login", authHandler.Login)
    r.GET("/:code", urlHandler.Redirect)

    // Client routes — JWT required
    client := r.Group("/client")
    client.Use(authMiddleware)
    {
        client.POST("/shorten", urlHandler.Shorten)
        client.GET("/links", urlHandler.ListLinks)
    }

    // Admin routes — JWT + admin role required
    admin := r.Group("/admin")
    admin.Use(authMiddleware, middleware.Admin())
    {
        admin.GET("/links", adminHandler.ListLinks)
        admin.DELETE("/links/:id", adminHandler.DeleteLink)
        admin.GET("/users", adminHandler.ListUsers)
    }

    return r
}