package trace

import (
	"context"

	"github.com/google/uuid"
)

type contextKey string

const traceIDKey contextKey = "trace_id"

// NewTraceID generates a new unique trace ID.
func NewTraceID() string {
	return uuid.New().String()[:8] // short 8-char ID — readable in logs
}

// WithTraceID returns a new context with the trace ID embedded.
func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceIDKey, traceID)
}

// FromContext extracts the trace ID from context.
// Returns empty string if not set.
func FromContext(ctx context.Context) string {
	if id, ok := ctx.Value(traceIDKey).(string); ok {
		return id
	}
	return ""
}
