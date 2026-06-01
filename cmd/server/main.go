package main

import (
    "log"

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
    db := database.Connect(cfg.DSN)

    // Token manager
    tokenManager := token.NewManager(cfg.JWTSecret, cfg.JWTDuration)

    // Repositories
    urlRepo := repository.NewURLRepository(db)
    userRepo := repository.NewUserRepository(db)

    // Services
    urlService := service.NewURLService(urlRepo)
    authService := service.NewAuthService(userRepo, tokenManager)

    // Handlers
    urlHandler := handler.NewURLHandler(urlService, cfg.BaseURL)
    authHandler := handler.NewAuthHandler(authService)

    // Auth middleware
    authMiddleware := middleware.Auth(tokenManager)

    r := router.Setup(urlHandler, authHandler, authMiddleware)

    log.Println("server starting on port", cfg.Port)
    if err := r.Run(":" + cfg.Port); err != nil {
        log.Fatal("server failed to start: ", err)
    }
}