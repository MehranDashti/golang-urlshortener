package apperror

import (
	"errors"
	"fmt"
	"net/http"
)

// AppError is our application error — carries HTTP code + message.
type AppError struct {
	Code    int
	Message string
	Details interface{}
	Err     error // wrapped underlying error — NEW
}

// Error implements the error interface.
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("code=%d message=%s: %v",
			e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("code=%d message=%s", e.Code, e.Message)
}

// Unwrap allows errors.Is and errors.As to look inside AppError.
// This is the key method that makes the chain work.
func (e *AppError) Unwrap() error {
	return e.Err
}

// IsAppError checks if any error in the chain is an *AppError.
func IsAppError(err error) (*AppError, bool) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr, true
	}
	return nil, false
}

// Constructor functions — clean API unchanged.
func NotFound(message string) *AppError {
	return &AppError{Code: http.StatusNotFound, Message: message}
}

func BadRequest(message string) *AppError {
	return &AppError{Code: http.StatusBadRequest, Message: message}
}

func BadRequestWithDetails(message string,
	details interface{}) *AppError {
	return &AppError{
		Code:    http.StatusBadRequest,
		Message: message,
		Details: details,
	}
}

func Internal(message string) *AppError {
	return &AppError{
		Code:    http.StatusInternalServerError,
		Message: message,
	}
}

// InternalWithErr wraps an underlying error — for logging
func InternalWithErr(message string, err error) *AppError {
	return &AppError{
		Code:    http.StatusInternalServerError,
		Message: message,
		Err:     err, // preserved for logging, not exposed to client
	}
}

func Unauthorized(message string) *AppError {
	return &AppError{Code: http.StatusUnauthorized, Message: message}
}

func Gone(message string) *AppError {
	return &AppError{Code: http.StatusGone, Message: message}
}

func Forbidden(message string) *AppError {
	return &AppError{Code: http.StatusForbidden, Message: message}
}
