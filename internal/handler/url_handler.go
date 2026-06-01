package handler

import (
    "net/http"

    "github.com/gin-gonic/gin"
    "urlshortener/internal/apperror"
    "urlshortener/internal/model"
)

type URLService interface {
    ShortenURL(originalURL string) (*model.URL, *apperror.AppError)
    GetByShortCode(code string) (*model.URL, *apperror.AppError)
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
    }
    if err := c.ShouldBindJSON(&body); err != nil || body.URL == "" {
        respondError(c, apperror.BadRequest("valid url is required"))
        return
    }

    url, appErr := h.service.ShortenURL(body.URL)
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