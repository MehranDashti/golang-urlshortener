package middleware

import (
    "log/slog"
    "time"

    "github.com/gin-gonic/gin"
)

func Logger() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()

        defer func() {
            slog.Info("request",
                "method",   c.Request.Method,
                "path",     c.Request.URL.Path,
                "status",   c.Writer.Status(),
                "duration", time.Since(start),
                "ip",       c.ClientIP(),
            )
        }()

        c.Next()
    }
}