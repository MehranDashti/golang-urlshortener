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
    rateLimiter   gin.HandlerFunc, 
) *gin.Engine {
    r := gin.New()
    r.Use(middleware.Logger())
    r.Use(gin.Recovery())
    r.Use(rateLimiter) 

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
            client.GET("/links/paginated", urlHandler.ListLinksPaginated)
        }

        // Admin routes — JWT + admin role required
        admin := api.Group("/admin")
        admin.Use(authMiddleware, middleware.Admin())
        {
            admin.GET("/links", adminHandler.ListLinks)
            admin.DELETE("/links/:id", adminHandler.DeleteLink)
            admin.GET("/users", adminHandler.ListUsers)
            admin.DELETE("/users/:id", adminHandler.DeleteUser)
        }
    }

    return r
}