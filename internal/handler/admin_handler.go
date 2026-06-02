package handler

import (
    "context"
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "urlshortener/internal/apperror"
    "urlshortener/internal/model"
)

type AdminService interface {
    GetAllLinks(ctx context.Context) ([]*model.URL, *apperror.AppError)
    DeleteLink(ctx context.Context, id string) *apperror.AppError
    GetAllUsers(ctx context.Context) ([]*model.User, *apperror.AppError)
    DeleteUser(ctx context.Context, id string) *apperror.AppError
}

type AdminHandler struct {
    service AdminService
}

func NewAdminHandler(service AdminService) *AdminHandler {
    return &AdminHandler{service: service}
}

func (h *AdminHandler) ListLinks(c *gin.Context) {
    urls, appErr := h.service.GetAllLinks(c.Request.Context())
    if appErr != nil {
        respondError(c, appErr)
        return
    }
    respondSuccess(c, http.StatusOK, "عملیات با موفقیت انجام شد", urls)
}

func (h *AdminHandler) DeleteLink(c *gin.Context) {
    id := c.Param("id")
    if id == "" {
        respondError(c, apperror.BadRequest("id is required"))
        return
    }

    if appErr := h.service.DeleteLink(
        c.Request.Context(), id); appErr != nil {
        return
    }

    respondSuccess(c, http.StatusOK, "لینک با موفقیت حذف شد", nil)
}

func (h *AdminHandler) ListUsers(c *gin.Context) {
    users, appErr := h.service.GetAllUsers(c.Request.Context())
    if appErr != nil {
        respondError(c, appErr)
        return
    }

    type safeUser struct {
        ID        string     `json:"id"`
        Email     string     `json:"email"`
        Role      model.Role `json:"role"`
        CreatedAt time.Time  `json:"created_at"`
    }

    result := make([]safeUser, len(users))
    for i, u := range users {
        result[i] = safeUser{
            ID:        u.ID,
            Email:     u.Email,
            Role:      u.Role,
            CreatedAt: u.CreatedAt,
        }
    }

    respondSuccess(c, http.StatusOK, "عملیات با موفقیت انجام شد", result)
}

func (h *AdminHandler) DeleteUser(c *gin.Context) {
    id := c.Param("id")
    if id == "" {
        respondError(c, apperror.BadRequest("id is required"))
        return
    }

    if appErr := h.service.DeleteUser(
        c.Request.Context(), id); appErr != nil {
        respondError(c, appErr)
        return
    }

    respondSuccess(c, http.StatusOK,
        "کاربر با موفقیت حذف شد", nil)
}