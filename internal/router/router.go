package router

import (
	"time"

	"github.com/gin-gonic/gin"
	"urlshortener/internal/handler"
	"urlshortener/internal/middleware"
)

func Setup(
	urlHandler *handler.URLHandler,
	authHandler *handler.AuthHandler,
	adminHandler *handler.AdminHandler,
	healthHandler *handler.HealthHandler,
	authMiddleware gin.HandlerFunc,
	globalLimiter gin.HandlerFunc,
	authLimiter gin.HandlerFunc,
	clientLimiter gin.HandlerFunc,
) *gin.Engine {
	r := gin.New()
	r.Use(middleware.Trace())
	r.Use(middleware.Logger())
	r.Use(gin.Recovery())
	r.Use(middleware.ErrorHandler())
	r.Use(globalLimiter)
	r.Use(middleware.Timeout(30 * time.Second))

	r.GET("/health", healthHandler.Check)
	r.GET("/healthz", healthHandler.Healthz)
	r.GET("/readyz", healthHandler.Readyz)	

	api := r.Group("/api/v1")
	{
		// Auth routes — tight limit (brute force protection)
		auth := api.Group("/auth")
		auth.Use(authLimiter)
		{
			auth.POST("/signup", authHandler.Signup)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.Refresh)
			auth.POST("/logout", authHandler.Logout)
		}

		api.GET("/:code", urlHandler.Redirect)

		// ── Client routes — JWT required ──────────────────────
		client := api.Group("/client")
		client.Use(authMiddleware)
		client.Use(clientLimiter) // ← after auth so userID is set
		{
			client.POST("/shorten", urlHandler.Shorten)
			client.POST("/shorten/bulk", urlHandler.BulkShorten)
			client.POST("/links/import", urlHandler.ImportLinks)
			client.GET("/links", urlHandler.ListLinks)
			client.GET("/links/paginated", urlHandler.ListLinksPaginated)
		}

		// ── Admin routes — JWT + admin role required ──────────
		admin := api.Group("/admin")
		admin.Use(authMiddleware, middleware.Admin())
		admin.Use(clientLimiter)
		{
			admin.GET("/links", adminHandler.ListLinks)
			admin.GET("/links/paginated", adminHandler.ListLinksPaginated)
			admin.GET("/links/export", adminHandler.ExportLinksCSV)
			admin.DELETE("/links/:id", adminHandler.DeleteLink)
			admin.GET("/users", adminHandler.ListUsers)
			admin.GET("/users/paginated", adminHandler.ListUsersPaginated)
			admin.DELETE("/users/:id", adminHandler.DeleteUser)
			admin.GET("/dashboard", adminHandler.Dashboard)
		}

		// ── Debug routes — only compiled with -tags pprof ────
		registerDebugRoutes(api)
	}

	return r
}
