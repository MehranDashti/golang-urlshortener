package handler

import (
    "net/http"

    "github.com/gin-gonic/gin"

    "urlshortener/internal/model"
    "urlshortener/internal/repository"
    "urlshortener/internal/util"
)

type URLHandler struct {
    repo    *repository.URLRepository
    baseURL string
}

func NewURLHandler(repo *repository.URLRepository, baseURL string) *URLHandler {
    return &URLHandler{
        repo:    repo,
        baseURL: baseURL,
    }
}

func (h *URLHandler) Shorten(c *gin.Context) {
    var body struct {
        URL string `json:"url"`
    }

    if err := c.ShouldBindJSON(&body); err != nil || body.URL == "" {
        c.JSON(http.StatusBadRequest, gin.H{
            "error": "valid url is required",
        })
        return
    }

    url := &model.URL{
        OriginalURL: body.URL,
        ShortCode:   util.GenerateShortCode(),
    }

    if err := h.repo.Create(url); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": "could not create short url",
        })
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

    url, err := h.repo.FindByShortCode(code)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": "something went wrong",
        })
        return
    }

    if url == nil {
        c.JSON(http.StatusNotFound, gin.H{
            "error": "short url not found",
        })
        return
    }

    h.repo.IncrementClicks(url.ID)
    c.Redirect(http.StatusMovedPermanently, url.OriginalURL)
}