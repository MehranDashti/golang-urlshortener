package main

import (
    "context"
    "log/slog"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

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

    // Give in-flight requests 30 seconds to finish
    ctx, cancel := context.WithTimeout(
        context.Background(), 30*time.Second)
    defer cancel()

    if err := srv.Shutdown(ctx); err != nil {
        slog.Error("forced shutdown", "error", err)
        os.Exit(1)
    }

    slog.Info("server stopped cleanly")
}