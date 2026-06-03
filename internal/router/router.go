package router

import (
    "net/http"
    "runtime"
    "time"
    
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
    r.Use(middleware.Timeout(30 * time.Second))

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
            client.POST("/shorten/bulk", urlHandler.BulkShorten)

            client.GET("/links", urlHandler.ListLinks)
            client.GET("/links/paginated", urlHandler.ListLinksPaginated)
        }

        // Admin routes — JWT + admin role required
        admin := api.Group("/admin")
        admin.Use(authMiddleware, middleware.Admin())
        {
            admin.GET("/links", adminHandler.ListLinks)
            admin.GET("/links/export", adminHandler.ExportLinksCSV)
            client.POST("/links/import", urlHandler.ImportLinks)
            admin.DELETE("/links/:id", adminHandler.DeleteLink)
            
            admin.GET("/users", adminHandler.ListUsers)
            admin.DELETE("/users/:id", adminHandler.DeleteUser)

            admin.GET("/dashboard", adminHandler.Dashboard)
        }
    }
    if gin.Mode() == gin.DebugMode {
        api.GET("/debug/goroutines", func(c *gin.Context) {
            buf := make([]byte, 1<<20)
            n := runtime.Stack(buf, true)
            c.Data(http.StatusOK, "text/plain", buf[:n])
        })

        api.GET("/debug/goroutine-count", func(c *gin.Context) {
            c.JSON(http.StatusOK, gin.H{
                "goroutines": runtime.NumGoroutine(),
            })
        })
    }

    return r
}