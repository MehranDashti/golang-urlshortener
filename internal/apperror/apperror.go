package apperror

import "fmt"

type AppError struct {
    Code    int
    Message string
    Details interface{} // validation errors, extra info
}

func (e *AppError) Error() string {
    return fmt.Sprintf("code=%d message=%s", e.Code, e.Message)
}

func NotFound(message string) *AppError {
    return &AppError{Code: 404, Message: message}
}

func BadRequest(message string) *AppError {
    return &AppError{Code: 400, Message: message}
}

func BadRequestWithDetails(message string, details interface{}) *AppError {
    return &AppError{Code: 400, Message: message, Details: details}
}

func Internal(message string) *AppError {
    return &AppError{Code: 500, Message: message}
}

func Unauthorized(message string) *AppError {
    return &AppError{Code: 401, Message: message}
}

func Gone(message string) *AppError {
    return &AppError{Code: 410, Message: message}
}