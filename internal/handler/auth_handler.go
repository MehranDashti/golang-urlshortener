package handler

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"urlshortener/internal/apperror"
	"urlshortener/internal/model"
	"urlshortener/internal/service"
)

type AuthService interface {
	Signup(ctx context.Context, email,
		password string) (*model.User, *apperror.AppError)
	Login(ctx context.Context, email,
		password string) (*service.TokenPair, *apperror.AppError)
	Refresh(refreshToken string) (*service.TokenPair, *apperror.AppError)
	Logout(accessToken string) *apperror.AppError // ← new
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
		_ = c.Error(appErr)
		return
	}

	user, appErr := h.service.Signup(
		c.Request.Context(), req.Email, req.Password)
	if appErr != nil {
		_ = c.Error(appErr)
		return
	}

	respondData(c, http.StatusCreated, "حساب کاربری با موفقیت ساخته شد", gin.H{
		"id":    user.ID,
		"email": user.Email,
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if appErr := bindAndValidate(c, &req); appErr != nil {
		_ = c.Error(appErr)
		return
	}

	pair, appErr := h.service.Login(
		c.Request.Context(), req.Email, req.Password)
	if appErr != nil {
		_ = c.Error(appErr)
		return
	}

	respondData(c, http.StatusOK, "ورود با موفقیت انجام شد", gin.H{
		"access_token":  pair.AccessToken,
		"refresh_token": pair.RefreshToken,
	})
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	var body struct {
		RefreshToken string `json:"refresh_token" validate:"required"`
	}
	if appErr := bindAndValidate(c, &body); appErr != nil {
		_ = c.Error(appErr)
		return
	}

	pair, appErr := h.service.Refresh(body.RefreshToken)
	if appErr != nil {
		_ = c.Error(appErr)
		return
	}

	respondData(c, http.StatusOK, "توکن با موفقیت تجدید شد", gin.H{
		"access_token":  pair.AccessToken,
		"refresh_token": pair.RefreshToken,
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	// Extract token from header
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		_ = c.Error(apperror.BadRequest(
			"authorization header required"))
		return
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		_ = c.Error(apperror.BadRequest(
			"invalid authorization format"))
		return
	}

	if appErr := h.service.Logout(parts[1]); appErr != nil {
		_ = c.Error(appErr)
		return
	}

	respondSuccess(c, http.StatusOK,
		"با موفقیت خارج شدید", nil)
}
