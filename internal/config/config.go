package config

import(
    "fmt"
    "os"
    "strconv"
    "time"

    "github.com/joho/godotenv"
)

type Config struct {
	Port            string
	BaseURL         string
	DSN             string
    JWTSecret       string
    JWTDuration     time.Duration
    AccessTokenDuration   time.Duration
    RefreshTokenDuration  time.Duration
}

func Load() *Config {
	godotenv.Load()

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

	return &Config{
		Port:    os.Getenv("APP_PORT"),
		BaseURL: os.Getenv("APP_BASE_URL"),
        DSN:     dsn,
        JWTSecret:   os.Getenv("JWT_SECRET"),
        JWTDuration: time.Duration(hours) * time.Hour,
        AccessTokenDuration:  time.Duration(accessMinutes) * time.Minute,
        RefreshTokenDuration: time.Duration(refreshDays) * 24 * time.Hour,
	}
}