package middleware

import (
	"github.com/gin-gonic/gin"
	"urlshortener/internal/trace"
)

const TraceIDHeader = "X-Trace-ID"

func Trace() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID := c.GetHeader(TraceIDHeader)
		if traceID == "" {
			traceID = trace.NewTraceID()
		}

		ctx := trace.WithTraceID(c.Request.Context(), traceID)
		c.Request = c.Request.WithContext(ctx)

		c.Set("trace_id", traceID)
		c.Header(TraceIDHeader, traceID)
		c.Next()
	}
}
