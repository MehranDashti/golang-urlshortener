package handler

import (
    "net/http"

    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
)

type HealthHandler struct {
    db *gorm.DB
}

func NewHealthHandler(db *gorm.DB) *HealthHandler {
    return &HealthHandler{db: db}
}

func (h *HealthHandler) Check(c *gin.Context) {
    // Check DB connectivity
    sqlDB, err := h.db.DB()
    if err != nil {
        c.JSON(http.StatusServiceUnavailable, gin.H{
            "status": "error",
            "db":     "cannot get sql.DB: " + err.Error(),
        })
        return
    }

    if err := sqlDB.PingContext(c.Request.Context()); err != nil {
        c.JSON(http.StatusServiceUnavailable, gin.H{
            "status": "error",
            "db":     "ping failed: " + err.Error(),
        })
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "status":  "ok",
        "db":      "ok",
        "version": "1.0.0",
    })
}