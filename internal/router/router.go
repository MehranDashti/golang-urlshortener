package router

import (
    "github.com/gin-gonic/gin"
    "urlshortener/internal/handler"
)

func Setup(
    urlHandler *handler.URLHandler,
    authHandler *handler.AuthHandler,
    authMiddleware gin.HandlerFunc,
) *gin.Engine {
    r := gin.Default()

    // Public routes
    r.POST("/auth/signup", authHandler.Signup)
    r.POST("/auth/login", authHandler.Login)
    r.GET("/:code", urlHandler.Redirect)

    // Protected routes — JWT required
    client := r.Group("/client")
    client.Use(authMiddleware)
    {
        client.POST("/shorten", urlHandler.Shorten)
    }

    return r
}