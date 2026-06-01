package handler

import (
	"github.com/gin-gonic/gin"

	"urlshortener/internal/apperror"
)

// APIResponse is the standard shape for every response in the app
type APIResponse struct {
    Success bool        `json:"success"`
    Code    int         `json:"code"`
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`  // omitempty = omit if nil
    Error   interface{} `json:"error,omitempty"` // omitempty = omit if nil
}

func respondSuccess(c *gin.Context, code int, message string, data interface{}) {
    c.JSON(code, APIResponse{
        Success: true,
        Code:    code,
        Message: message,
        Data:    data,
    })
}

func respondError(c *gin.Context, err *apperror.AppError) {
    c.JSON(err.Code, APIResponse{
        Success: false,
        Code:    err.Code,
        Message: err.Message,
        Error:   err.Details, // we'll add this below
    })
}