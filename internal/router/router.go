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
    r := gin.New()
    r.Use(middleware.Logger())
    r.Use(gin.Recovery())

    // All routes live under /api/v1
    api := r.Group("/api/v1")
    {
        // Public routes
        api.POST("/auth/signup", authHandler.Signup)
        api.POST("/auth/login", authHandler.Login)
        api.POST("/auth/refresh", authHandler.Refresh)
        api.GET("/:code", urlHandler.Redirect)

        // Client routes — JWT required
        client := api.Group("/client")
        client.Use(authMiddleware)
        {
            client.POST("/shorten", urlHandler.Shorten)
            client.GET("/links", urlHandler.ListLinks)
        }

        // Admin routes — JWT + admin role required
        admin := api.Group("/admin")
        admin.Use(authMiddleware, middleware.Admin())
        {
            admin.GET("/links", adminHandler.ListLinks)
            admin.DELETE("/links/:id", adminHandler.DeleteLink)
            admin.GET("/users", adminHandler.ListUsers)
        }
    }

    return r
}