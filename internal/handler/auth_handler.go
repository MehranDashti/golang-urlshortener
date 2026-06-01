package handler

import (
    "net/http"

    "github.com/gin-gonic/gin"
	"urlshortener/internal/model"
    "urlshortener/internal/apperror"
)

type AuthService interface {
    Signup(email, password string) (*model.User, *apperror.AppError)
    Login(email, password string) (string, *apperror.AppError)
}

type AuthHandler struct {
    service AuthService
}

func NewAuthHandler(service AuthService) *AuthHandler {
    return &AuthHandler{service: service}
}

func (h *AuthHandler) Signup(c *gin.Context) {
    var body struct {
        Email    string `json:"email"`
        Password string `json:"password"`
    }
    if err := c.ShouldBindJSON(&body); err != nil || body.Email == "" || body.Password == "" {
        respondError(c, apperror.BadRequest("email and password are required"))
        return
    }

    user, appErr := h.service.Signup(body.Email, body.Password)
    if appErr != nil {
        respondError(c, appErr)
        return
    }

    c.JSON(http.StatusCreated, gin.H{"message": "account created", "user": user})
}

func (h *AuthHandler) Login(c *gin.Context) {
    var body struct {
        Email    string `json:"email"`
        Password string `json:"password"`
    }
    if err := c.ShouldBindJSON(&body); err != nil || body.Email == "" || body.Password == "" {
        respondError(c, apperror.BadRequest("email and password are required"))
        return
    }

    tokenStr, appErr := h.service.Login(body.Email, body.Password)
    if appErr != nil {
        respondError(c, appErr)
        return
    }

    c.JSON(http.StatusOK, gin.H{"token": tokenStr})
}