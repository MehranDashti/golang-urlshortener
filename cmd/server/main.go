package main

import (
    "log"
    "urlshortener/internal/config"
    "urlshortener/internal/database"
    "urlshortener/internal/handler"
    "urlshortener/internal/repository"
    "urlshortener/internal/router"
    "urlshortener/internal/service"
)

func main() {
    cfg := config.Load()
    db := database.Connect(cfg.DSN)

    urlRepo := repository.NewURLRepository(db)
    urlService := service.NewURLService(urlRepo)     
    urlHandler := handler.NewURLHandler(urlService, cfg.BaseURL)

    r := router.Setup(urlHandler)

    log.Println("server starting on port", cfg.Port)
    if err := r.Run(":" + cfg.Port); err != nil {
        log.Fatal("server failed to start: ", err)
    }
}