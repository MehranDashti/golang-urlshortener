package config

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                 string
	BaseURL              string
	DSN                  string
	JWTSecret            string
	JWTDuration          time.Duration
	AccessTokenDuration  time.Duration
	RefreshTokenDuration time.Duration
	// Add to Config struct:
	RedisAddr     string
	RedisPassword string
	RedisDB       int

}

func Load() *Config {
	_ = godotenv.Load() // .env is optional — production uses real env vars

	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASS"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)

	hours, err := strconv.Atoi(os.Getenv("JWT_DURATION_HOURS"))
	if err != nil || hours == 0 {
		hours = 24
	}

	accessMinutes, err := strconv.Atoi(os.Getenv("ACCESS_TOKEN_DURATION_MINUTES"))
	if err != nil || accessMinutes == 0 {
		accessMinutes = 15
	}

	refreshDays, err := strconv.Atoi(os.Getenv("REFRESH_TOKEN_DURATION_DAYS"))
	if err != nil || refreshDays == 0 {
		refreshDays = 7
	}

	redisDB, err := strconv.Atoi(os.Getenv("REDIS_DB"))
	if err != nil {
		redisDB = 0
	}

	return &Config{
		Port:                 os.Getenv("APP_PORT"),
		BaseURL:              os.Getenv("APP_BASE_URL"),
		DSN:                  dsn,
		JWTSecret:            os.Getenv("JWT_SECRET"),
		JWTDuration:          time.Duration(hours) * time.Hour,
		AccessTokenDuration:  time.Duration(accessMinutes) * time.Minute,
		RefreshTokenDuration: time.Duration(refreshDays) * 24 * time.Hour,
		RedisAddr:            os.Getenv("REDIS_ADDR"),
		RedisPassword:        os.Getenv("REDIS_PASSWORD"),
		RedisDB:              redisDB,
	}
}

func (c *Config) Validate() error {
	var missing []string

	if c.Port == "" {
		missing = append(missing, "APP_PORT")
	}
	if c.BaseURL == "" {
		missing = append(missing, "APP_BASE_URL")
	}
	if c.JWTSecret == "" {
		missing = append(missing, "JWT_SECRET")
	}
	if c.DSN == "" {
		missing = append(missing, "DB_USER/DB_PASS/DB_HOST/DB_PORT/DB_NAME")
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required env vars: %s",
			strings.Join(missing, ", "))
	}

	// Warn about weak JWT secret in production
	if len(c.JWTSecret) < 32 {
		slog.Warn("JWT_SECRET is short — use at least 32 characters in production",
			"length", len(c.JWTSecret))
	}

	return nil
}
