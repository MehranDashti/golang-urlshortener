package config

import(
	"fmt"
    "os"

    "github.com/joho/godotenv"
)

type Config struct {
	Port string
	BaseURL string
	DSN string
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

	return &Config{
		Port:    os.Getenv("APP_PORT"),
		BaseURL: os.Getenv("APP_BASE_URL"),
        DSN:     dsn,
	}
}