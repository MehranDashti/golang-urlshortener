package handler

import (
    "github.com/gin-gonic/gin"
    "urlshortener/internal/apperror"
)

type APIResponse struct {
    Success bool        `json:"success"`
    Code    int         `json:"code"`
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
    Error   interface{} `json:"error,omitempty"`
}

// respondData is generic — T is inferred from data at call site
func respondData[T any](c *gin.Context, code int, message string, data T) {
    c.JSON(code, APIResponse{
        Success: true,
        Code:    code,
        Message: message,
        Data:    data,
    })
}

// respondSuccess kept as alias for existing call sites
// delegates to respondData
func respondSuccess(c *gin.Context, code int, message string, data interface{}) {
    respondData(c, code, message, data)
}

func respondError(c *gin.Context, err *apperror.AppError) {
    c.JSON(err.Code, APIResponse{
        Success: false,
        Code:    err.Code,
        Message: err.Message,
        Error:   err.Details,
    })
}