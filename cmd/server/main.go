package main

import (
	"context"
	_ "embed"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"urlshortener/internal/cache"
	"urlshortener/internal/config"
	"urlshortener/internal/database"
	"urlshortener/internal/handler"
	"urlshortener/internal/metrics"
	"urlshortener/internal/middleware"
	"urlshortener/internal/repository"
	"urlshortener/internal/router"
	"urlshortener/internal/service"
	"urlshortener/internal/tokenstore"
	"urlshortener/pkg/token"
)

var blacklist tokenstore.TokenBlacklist = tokenstore.NewBlacklist()

func main() {
	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		slog.Error("invalid configuration", "error", err)
		os.Exit(1)
	}

	db := database.Connect(cfg.DSN)

	redisCache := cache.NewRedisCache(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
	if err := redisCache.Ping(context.Background()); err != nil {
		slog.Warn("redis unavailable — falling back to in-memory", "error", err)
	} else {
		blacklist = tokenstore.NewRedisBlacklist(redisCache)
	}

	slog.Info("server starting",
		"port", cfg.Port,
		"version", strings.TrimSpace(version))

	// Wire dependencies
	tokenManager := token.NewManager(
		cfg.JWTSecret,
		cfg.AccessTokenDuration,
		cfg.RefreshTokenDuration,
	)

	// Create a cancellable context for background workers
	workerCtx, workerCancel := context.WithCancel(
		context.Background())
	defer workerCancel() // cancels all workers on shutdown

	transactor := service.NewDBTransactor(db)

	urlRepo := repository.NewURLRepository(db)
	userRepo := repository.NewUserRepository(db)
	urlService := service.NewURLService(urlRepo, redisCache, workerCtx)
	authService := service.NewAuthService(
		userRepo, tokenManager, blacklist)
	adminService := service.NewAdminService(urlRepo, userRepo, transactor)

	urlHandler := handler.NewURLHandler(urlService, cfg.BaseURL)
	authHandler := handler.NewAuthHandler(authService)
	adminHandler := handler.NewAdminHandler(adminService)
	healthHandler := handler.NewHealthHandler(db)

	authMiddleware := middleware.Auth(tokenManager, blacklist)

	// Three limiters with different configs
	globalLimiter := middleware.NewRateLimiter(100, time.Minute) // 100/min per IP
	authLimiter := middleware.NewRateLimiter(10, time.Minute)    // 10/min per IP — brute force
	clientLimiter := middleware.NewRateLimiter(60, time.Minute)  // 60/min per user

	r := router.Setup(
		urlHandler,
		authHandler,
		adminHandler,
		healthHandler,
		authMiddleware,
		globalLimiter.Middleware(),
		authLimiter.Middleware(),
		clientLimiter.Middleware(),
	)

	// Create http.Server manually — we need it for Shutdown()
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,

		// Timeouts — protect against slow clients
		ReadTimeout:  10 * time.Second, // max time to read request
		WriteTimeout: 30 * time.Second, // max time to write response
		IdleTimeout:  60 * time.Second, // max keep-alive idle time
	}

	// Start server in a goroutine — non-blocking
	// so we can listen for signals on the main goroutine
	go func() {
		slog.Info("server starting", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil &&
			err != http.ErrServerClosed {
			// ErrServerClosed is expected on shutdown — not a real error
			slog.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	// Start metrics server on a separate port (pure net/http, no Gin)  ← HERE
	metricsServer := metrics.NewServer(":9090")
	go func() {
		slog.Info("metrics server starting", "port", 9090)
		if err := metricsServer.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("metrics server failed", "error", err)
		}
	}()

	// Block until we receive SIGINT or SIGTERM
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit // blocks here

	slog.Info("shutdown signal received", "signal", sig)

	// Give in-flight requests 30s — but FORCE exit if shutdown itself hangs
	shutdownCtx, cancel := context.WithTimeout(
		context.Background(), 30*time.Second)
	defer cancel()

	// Channel to know when Shutdown() finishes
	done := make(chan struct{})

	go func() {
		if err := srv.Shutdown(shutdownCtx); err != nil {
			slog.Error("shutdown error", "error", err)
		}
		close(done) // signal that shutdown completed
	}()

	// select — wait for EITHER shutdown to finish OR timeout
	select {
	case <-done:
		slog.Info("server stopped cleanly")
	case <-shutdownCtx.Done():
		slog.Error("shutdown timed out — forcing exit")
		os.Exit(1)
	}

	// Drain remaining click increments after HTTP server stopped
	urlService.Close()
	slog.Info("click worker stopped")
}
