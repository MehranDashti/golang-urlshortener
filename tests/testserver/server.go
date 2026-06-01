package testserver

import (
    "fmt"
    "os"
    "path/filepath"
    "runtime"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/joho/godotenv"
    "gorm.io/driver/mysql"
    "gorm.io/gorm"
    "urlshortener/internal/handler"
    "urlshortener/internal/middleware"
    "urlshortener/internal/model"
    "urlshortener/internal/repository"
    "urlshortener/internal/router"
    "urlshortener/internal/service"
    "urlshortener/pkg/token"
)

type TestServer struct {
    Router *gin.Engine
    DB     *gorm.DB
}

func New() *TestServer {
    _, filename, _, _ := runtime.Caller(0)
    root := filepath.Join(filepath.Dir(filename), "../..")
    godotenv.Load(filepath.Join(root, ".env.testing"))

    gin.SetMode(gin.TestMode)

    dsn := fmt.Sprintf(
        "%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
        os.Getenv("DB_USER"),
        os.Getenv("DB_PASS"),
        os.Getenv("DB_HOST"),
        os.Getenv("DB_PORT"),
        os.Getenv("DB_NAME"),
    )

    db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
    if err != nil {
        panic("failed to connect to test database: " + err.Error())
    }

    // Run migrations on test DB
    db.AutoMigrate(&model.User{}, &model.URL{})

    tokenManager := token.NewManager(
        "test-secret-key",
        15*time.Minute,
        7*24*time.Hour,
    )

    urlRepo     := repository.NewURLRepository(db)
    userRepo    := repository.NewUserRepository(db)
    urlService  := service.NewURLService(urlRepo)
    authService := service.NewAuthService(userRepo, tokenManager)
    adminService := service.NewAdminService(urlRepo, userRepo)

    urlHandler   := handler.NewURLHandler(urlService, "http://localhost:8080")
    authHandler  := handler.NewAuthHandler(authService)
    adminHandler := handler.NewAdminHandler(adminService)

    authMiddleware := middleware.Auth(tokenManager)

    r := router.Setup(urlHandler, authHandler, adminHandler, authMiddleware)

    return &TestServer{Router: r, DB: db}
}

// CleanDB truncates all tables between tests — like Laravel's RefreshDatabase
func (s *TestServer) CleanDB() {
    s.DB.Exec("SET FOREIGN_KEY_CHECKS = 0")
    s.DB.Exec("TRUNCATE TABLE urls")
    s.DB.Exec("TRUNCATE TABLE users")
    s.DB.Exec("SET FOREIGN_KEY_CHECKS = 1")
}