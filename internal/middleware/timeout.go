package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Timeout returns middleware that cancels the request context
// after the given duration.
// If the handler takes longer than d — client gets 503.
func Timeout(d time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Create a new context with deadline
		ctx, cancel := context.WithTimeout(
			c.Request.Context(), d)
		defer cancel()

		// Replace the request context with our timeout context
		c.Request = c.Request.WithContext(ctx)

		// Channel to know when the handler chain finishes
		done := make(chan struct{})

		// Run the handler chain in a goroutine
		go func() {
			c.Next()    // run all subsequent handlers
			close(done) // signal completion
		}()

		// select — wait for handler OR timeout
		select {
		case <-done:
			// Handler finished in time — nothing to do
		case <-ctx.Done():
			// Timeout exceeded
			c.AbortWithStatusJSON(http.StatusServiceUnavailable,
				map[string]interface{}{
					"success": false,
					"code":    503,
					"message": "request timeout — server too busy",
				})
		}
	}
}
