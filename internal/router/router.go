package router

import (
	"net/http"
	"net/http/pprof"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"urlshortener/internal/handler"
	"urlshortener/internal/middleware"
)

func Setup(
    urlHandler    *handler.URLHandler,
    authHandler   *handler.AuthHandler,
    adminHandler  *handler.AdminHandler,
    healthHandler *handler.HealthHandler,
    authMiddleware gin.HandlerFunc,
    globalLimiter  gin.HandlerFunc,
    authLimiter    gin.HandlerFunc,
    clientLimiter  gin.HandlerFunc,
) *gin.Engine {
    r := gin.New()
    r.Use(middleware.Trace())   
    r.Use(middleware.Logger())
    r.Use(gin.Recovery())
    r.Use(globalLimiter)
    r.Use(middleware.Timeout(30 * time.Second))

    r.GET("/health", healthHandler.Check)
	
	api := r.Group("/api/v1")
	{
		// Auth routes — tight limit (brute force protection)
        auth := api.Group("/auth")
        auth.Use(authLimiter)
        {
            auth.POST("/signup",  authHandler.Signup)
            auth.POST("/login",   authHandler.Login)
            auth.POST("/refresh", authHandler.Refresh)
            auth.POST("/logout",  authHandler.Logout)
        }
		
		api.GET("/:code",         urlHandler.Redirect)

		// ── Client routes — JWT required ──────────────────────
		client := api.Group("/client")
        client.Use(authMiddleware)
        client.Use(clientLimiter)               // ← after auth so userID is set
        {
            client.POST("/shorten",          urlHandler.Shorten)
            client.POST("/shorten/bulk",     urlHandler.BulkShorten)
            client.POST("/links/import",     urlHandler.ImportLinks)
            client.GET("/links",             urlHandler.ListLinks)
            client.GET("/links/paginated",   urlHandler.ListLinksPaginated)
        }

		// ── Admin routes — JWT + admin role required ──────────
		admin := api.Group("/admin")
        admin.Use(authMiddleware, middleware.Admin())
        admin.Use(clientLimiter)
        {
            admin.GET("/links",             adminHandler.ListLinks)
            admin.GET("/links/paginated",   adminHandler.ListLinksPaginated)
            admin.GET("/links/export",      adminHandler.ExportLinksCSV)
            admin.DELETE("/links/:id",      adminHandler.DeleteLink)
            admin.GET("/users",             adminHandler.ListUsers)
            admin.GET("/users/paginated",   adminHandler.ListUsersPaginated)
            admin.DELETE("/users/:id",      adminHandler.DeleteUser)
            admin.GET("/dashboard",         adminHandler.Dashboard)
        }

		// ── Debug routes — development only ───────────────────
		if gin.Mode() == gin.DebugMode {
			debug := api.Group("/debug")
			{
				debug.GET("/goroutines", func(c *gin.Context) {
					c.JSON(http.StatusOK, gin.H{
						"goroutines": runtime.NumGoroutine(),
					})
				})
				debug.GET("/pprof/",          gin.WrapF(pprof.Index))
				debug.GET("/pprof/cmdline",   gin.WrapF(pprof.Cmdline))
				debug.GET("/pprof/profile",   gin.WrapF(pprof.Profile))
				debug.GET("/pprof/symbol",    gin.WrapF(pprof.Symbol))
				debug.GET("/pprof/trace",     gin.WrapF(pprof.Trace))
				debug.GET("/pprof/heap",      gin.WrapH(pprof.Handler("heap")))
				debug.GET("/pprof/goroutine", gin.WrapH(pprof.Handler("goroutine")))
				debug.GET("/pprof/block",     gin.WrapH(pprof.Handler("block")))
				debug.GET("/pprof/mutex",     gin.WrapH(pprof.Handler("mutex")))
			}
		}
	}

	return r
}