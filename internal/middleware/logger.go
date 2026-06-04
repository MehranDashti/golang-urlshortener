package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"urlshortener/internal/trace"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		defer func() {
			traceID := trace.FromContext(c.Request.Context())

			slog.Info("request",
				"method", c.Request.Method,
				"path", c.Request.URL.Path,
				"status", c.Writer.Status(),
				"duration", time.Since(start),
				"ip", c.ClientIP(),
				"trace_id", traceID,
			)
		}()

		c.Next()
	}
}
