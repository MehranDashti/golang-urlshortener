package router

import (
    "github.com/gin-gonic/gin"

    "urlshortener/internal/handler"
)

func Setup(urlHandler *handler.URLHandler) *gin.Engine {
    r := gin.Default()

    r.POST("/shorten", urlHandler.Shorten)
    r.GET("/:code", urlHandler.Redirect)

    return r
}