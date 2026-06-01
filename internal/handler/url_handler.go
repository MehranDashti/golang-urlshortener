package handler

import (
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "urlshortener/internal/middleware"
    "urlshortener/internal/apperror"
    "urlshortener/internal/model"
)

type URLService interface {
    ShortenURL(originalURL string, userID string, expiresAt *time.Time) (*model.URL, *apperror.AppError)
    GetByShortCode(code string) (*model.URL, *apperror.AppError)
    GetUserLinks(userID string) ([]*model.URL, *apperror.AppError)
}

type URLHandler struct {
    service URLService
    baseURL string
}

func NewURLHandler(service URLService, baseURL string) *URLHandler {
    return &URLHandler{service: service, baseURL: baseURL}
}

func respondError(c *gin.Context, err *apperror.AppError) {
    c.JSON(err.Code, gin.H{"error": err.Message})
}

func (h *URLHandler) Shorten(c *gin.Context) {
    var body struct {
        URL string `json:"url"`
        ExpiresAt *string `json:"expires_at"`
    }
    if err := c.ShouldBindJSON(&body); err != nil || body.URL == "" {
        respondError(c, apperror.BadRequest("valid url is required"))
        return
    }

    userID, exists := c.Get(middleware.UserIDKey)
    if !exists {
        respondError(c, apperror.Unauthorized("not authenticated"))
        return
    }

    var expiresAt *time.Time
    if body.ExpiresAt != nil {
        t, err := time.Parse(time.RFC3339, *body.ExpiresAt)
        if err != nil {
            respondError(c, apperror.BadRequest("invalid expires_at format, use RFC3339: 2006-01-02T15:04:05Z"))
            return
        }
        expiresAt = &t
    }

    url, appErr := h.service.ShortenURL(body.URL, userID.(string), expiresAt)
    if appErr != nil {
        respondError(c, appErr)
        return
    }

    c.JSON(http.StatusCreated, gin.H{
        "short_url":    h.baseURL + "/" + url.ShortCode,
        "short_code":   url.ShortCode,
        "original_url": url.OriginalURL,
    })
}

func (h *URLHandler) Redirect(c *gin.Context) {
    code := c.Param("code")

    url, appErr := h.service.GetByShortCode(code)
    if appErr != nil {
        respondError(c, appErr)
        return
    }

    c.Redirect(http.StatusMovedPermanently, url.OriginalURL)
}

func (h *URLHandler) ListLinks(c *gin.Context) {
    userID, exists := c.Get(middleware.UserIDKey)  // from auth middleware via context
    if !exists {
        respondError(c, apperror.Unauthorized("not authenticated"))
        return
    }
    urls, appErr := h.service.GetUserLinks(userID.(string))
    if appErr != nil {
        respondError(c, appErr)
        return
    }
    c.JSON(http.StatusOK, gin.H{"links": urls})
}