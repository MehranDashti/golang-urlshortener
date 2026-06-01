package handler

import (
    "net/http"

    "github.com/gin-gonic/gin"
    "urlshortener/internal/apperror"
    "urlshortener/internal/model"
    "urlshortener/internal/service"
)

type AuthService interface {
    Signup(email, password string) (*model.User, *apperror.AppError)
    Login(email, password string) (*service.TokenPair, *apperror.AppError)
    Refresh(refreshToken string) (*service.TokenPair, *apperror.AppError)
}

type AuthHandler struct {
    service AuthService
}

func NewAuthHandler(service AuthService) *AuthHandler {
    return &AuthHandler{service: service}
}

func (h *AuthHandler) Signup(c *gin.Context) {
    var req SignupRequest
    if appErr := bindAndValidate(c, &req); appErr != nil {
        respondError(c, appErr)
        return
    }

    user, appErr := h.service.Signup(req.Email, req.Password)
    if appErr != nil {
        respondError(c, appErr)
        return
    }

    respondSuccess(c, http.StatusCreated, "حساب کاربری با موفقیت ساخته شد", gin.H{
        "id":    user.ID,
        "email": user.Email,
    })
}

func (h *AuthHandler) Login(c *gin.Context) {
    var req LoginRequest
    if appErr := bindAndValidate(c, &req); appErr != nil {
        respondError(c, appErr)
        return
    }

    pair, appErr := h.service.Login(req.Email, req.Password)
    if appErr != nil {
        respondError(c, appErr)
        return
    }

    respondSuccess(c, http.StatusOK, "ورود با موفقیت انجام شد", gin.H{
        "access_token":  pair.AccessToken,
        "refresh_token": pair.RefreshToken,
    })
}

func (h *AuthHandler) Refresh(c *gin.Context) {
    var body struct {
        RefreshToken string `json:"refresh_token" validate:"required"`
    }
    if appErr := bindAndValidate(c, &body); appErr != nil {
        respondError(c, appErr)
        return
    }

    pair, appErr := h.service.Refresh(body.RefreshToken)
    if appErr != nil {
        respondError(c, appErr)
        return
    }

    respondSuccess(c, http.StatusOK, "توکن با موفقیت تجدید شد", gin.H{
        "access_token":  pair.AccessToken,
        "refresh_token": pair.RefreshToken,
    })
}