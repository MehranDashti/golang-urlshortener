package handler

import (
    "context"
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "urlshortener/internal/apperror"
    "urlshortener/internal/middleware"
    "urlshortener/internal/model"
)

type URLService interface {
    ShortenURL(ctx context.Context, originalURL string,
        userID string, expiresAt *time.Time) (*model.URL, *apperror.AppError)
    GetByShortCode(ctx context.Context,
        code string) (*model.URL, *apperror.AppError)
    GetUserLinks(ctx context.Context,
        userID string) ([]*model.URL, *apperror.AppError)
    GetUserLinksPaginated(ctx context.Context, userID string,
        params model.PaginationParams) (*model.PaginatedResult, *apperror.AppError) // ← new
}

type URLHandler struct {
    service URLService
    baseURL string
}

func NewURLHandler(service URLService, baseURL string) *URLHandler {
    return &URLHandler{service: service, baseURL: baseURL}
}

func (h *URLHandler) Shorten(c *gin.Context) {
    var req ShortenRequest
    if appErr := bindAndValidate(c, &req); appErr != nil {
        respondError(c, appErr)
        return
    }

    userID, exists := c.Get(middleware.UserIDKey)
    if !exists {
        respondError(c, apperror.Unauthorized("not authenticated"))
        return
    }

    var expiresAt *time.Time
    if req.ExpiresAt != nil {
        t, err := time.Parse(time.RFC3339, *req.ExpiresAt)
        if err != nil {
            respondError(c, apperror.BadRequest("invalid expires_at format, use RFC3339: 2006-01-02T15:04:05Z"))
            return
        }
        expiresAt = &t
    }

    url, appErr := h.service.ShortenURL(
        c.Request.Context(), // ← pass request context
        req.URL,
        userID.(string),
        expiresAt,
    )
    if appErr != nil {
        respondError(c, appErr)
        return
    }

    respondSuccess(c, http.StatusCreated, "لینک کوتاه با موفقیت ساخته شد", gin.H{
        "short_url":    h.baseURL + "/" + url.ShortCode,
        "short_code":   url.ShortCode,
        "original_url": url.OriginalURL,
    })
}

func (h *URLHandler) Redirect(c *gin.Context) {
    code := c.Param("code")

    url, appErr := h.service.GetByShortCode(
        c.Request.Context(), code)
    if appErr != nil {
        respondError(c, appErr)
        return
    }

    c.Redirect(http.StatusMovedPermanently, url.OriginalURL)
}

func (h *URLHandler) ListLinks(c *gin.Context) {
    userID, exists := c.Get(middleware.UserIDKey)
    if !exists {
        respondError(c, apperror.Unauthorized("not authenticated"))
        return
    }
    urls, appErr := h.service.GetUserLinks(
        c.Request.Context(), userID.(string))
    if appErr != nil {
        respondError(c, appErr)
        return
    }

    respondSuccess(c, http.StatusOK, "عملیات با موفقیت انجام شد", urls)
}

func (h *URLHandler) ListLinksPaginated(c *gin.Context) {
    userID, exists := c.Get(middleware.UserIDKey)
    if !exists {
        respondError(c, apperror.Unauthorized("not authenticated"))
        return
    }

    params, appErr := parsePagination(c)
    if appErr != nil {
        respondError(c, appErr)
        return
    }

    result, appErr := h.service.GetUserLinksPaginated(
        c.Request.Context(), userID.(string), params)
    if appErr != nil {
        respondError(c, appErr)
        return
    }

    respondSuccess(c, http.StatusOK,
        "عملیات با موفقیت انجام شد", result)
}