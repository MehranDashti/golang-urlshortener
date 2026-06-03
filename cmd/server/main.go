package main

import (
    "strings"
    "context"
    "log/slog"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"
    _ "embed"

    "urlshortener/internal/config"
    "urlshortener/internal/database"
    "urlshortener/internal/handler"
    "urlshortener/internal/middleware"
    "urlshortener/internal/repository"
    "urlshortener/internal/router"
    "urlshortener/internal/service"
    "urlshortener/pkg/token"
)

func main() {
    cfg := config.Load()
    db  := database.Connect(cfg.DSN)

    slog.Info("server starting",
    "port", cfg.Port,
    "version", strings.TrimSpace(version))

    // Wire dependencies
    tokenManager := token.NewManager(
        cfg.JWTSecret,
        cfg.AccessTokenDuration,
        cfg.RefreshTokenDuration,
    )

    urlRepo      := repository.NewURLRepository(db)
    userRepo     := repository.NewUserRepository(db)
    urlService   := service.NewURLService(urlRepo)
    authService  := service.NewAuthService(userRepo, tokenManager)
    adminService := service.NewAdminService(urlRepo, userRepo)

    urlHandler   := handler.NewURLHandler(urlService, cfg.BaseURL)
    authHandler  := handler.NewAuthHandler(authService)
    adminHandler := handler.NewAdminHandler(adminService)

    authMiddleware := middleware.Auth(tokenManager)
    rateLimiter    := middleware.NewRateLimiter(60, time.Minute)

    r := router.Setup(
        urlHandler,
        authHandler,
        adminHandler,
        authMiddleware,
        rateLimiter.Middleware(),
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