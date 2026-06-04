package middleware

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"urlshortener/internal/apperror"
	"urlshortener/internal/trace"
)

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next() 

		if len(c.Errors) == 0 {
			return 
		}

		err := c.Errors.Last().Err 

		var appErr *apperror.AppError
		if errors.As(err, &appErr) {
			if appErr.Code >= http.StatusInternalServerError && appErr.Err != nil {
				slog.Error("internal error",
					"trace_id", trace.FromContext(c.Request.Context()),
					"path", c.Request.URL.Path,
					"error", appErr.Err,
				)
			}

			c.JSON(appErr.Code, gin.H{
				"success": false,
				"code":    appErr.Code,
				"message": appErr.Message,
				"error":   appErr.Details, 
			})
			return
		}
		slog.Error("unhandled error — not an AppError",
			"trace_id", trace.FromContext(c.Request.Context()),
			"path", c.Request.URL.Path,
			"error", err,
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"code":    http.StatusInternalServerError,
			"message": "internal server error",
			"error":   nil,
		})
	}
}