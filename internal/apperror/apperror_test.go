package apperror_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"urlshortener/internal/apperror"
)

func TestAppError(t *testing.T) {
	tests := []struct {
		name     string
		err      *apperror.AppError
		wantCode int
		wantMsg  string
	}{
		{"not found", apperror.NotFound("link missing"), 404, "link missing"},
		{"bad request", apperror.BadRequest("invalid"), 400, "invalid"},
		{"internal", apperror.Internal("db down"), 500, "db down"},
		{"unauthorized", apperror.Unauthorized("no token"), 401, "no token"},
		{"gone", apperror.Gone("link expired"), 410, "link expired"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantCode, tt.err.Code)
			assert.Equal(t, tt.wantMsg, tt.err.Message)
			assert.NotEmpty(t, tt.err.Error())
		})
	}
}

func TestAppError_ErrorsAs(t *testing.T) {
	// Wrap an AppError inside fmt.Errorf
	original := apperror.NotFound("url not found")
	wrapped := fmt.Errorf("service layer: %w", original)

	// errors.As should find the AppError through the chain
	var appErr *apperror.AppError
	assert.True(t, errors.As(wrapped, &appErr))
	assert.Equal(t, 404, appErr.Code)
	assert.Equal(t, "url not found", appErr.Message)
}

func TestAppError_Unwrap(t *testing.T) {
	underlying := errors.New("connection refused")
	wrapped := apperror.InternalWithErr("db failed", underlying)

	// errors.Is should find the underlying error through AppError.Unwrap()
	assert.True(t, errors.Is(wrapped, underlying))
}
