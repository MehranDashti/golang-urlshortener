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

// Readyz checks all dependencies — used by Kubernetes readiness probe.
// Fails → pod removed from load balancer, NOT restarted.
func (h *HealthHandler) Readyz(c *gin.Context) {
    checks := map[string]string{}
    healthy := true

    // Check DB
    sqlDB, err := h.db.DB()
    if err != nil || sqlDB.Ping() != nil {
        checks["database"] = "unavailable"
        healthy = false
    } else {
        checks["database"] = "ok"
    }

    if !healthy {
        c.JSON(http.StatusServiceUnavailable, gin.H{
            "status": "unavailable",
            "checks": checks,
        })
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "status": "ready",
        "checks": checks,
    })
}

// Healthz is the liveness probe — just confirms the process is alive.
// Never checks dependencies — a DB outage should NOT restart the pod.
func (h *HealthHandler) Healthz(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{"status": "alive"})
}