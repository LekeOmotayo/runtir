package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port            string
	AppEnv          string
	DatabaseURL     string
	JWTSecret       string
	JWTExpiryHours  int
	CookieName      string
	CookieSecure    bool
	CookieDomain    string
	UploadDir       string
	MaxUploadSizeMB int64
}

func LoadConfig() *Config {
	// Attempt to load .env file if present
	if err := godotenv.Load(); err != nil {
		log.Println("Notice: .env file not found or could not be loaded, using system environment variables")
	}

	appEnv := getEnv("APP_ENV", "development")
	isProd := appEnv == "production"

	jwtExpiryHours, _ := strconv.Atoi(getEnv("JWT_EXPIRY_HOURS", "24"))
	if jwtExpiryHours <= 0 {
		jwtExpiryHours = 24
	}

	maxUploadSizeMB, _ := strconv.ParseInt(getEnv("MAX_UPLOAD_SIZE_MB", "25"), 10, 64)
	if maxUploadSizeMB <= 0 {
		maxUploadSizeMB = 25
	}

	cookieSecure := isProd
	if secureStr := os.Getenv("COOKIE_SECURE"); secureStr != "" {
		cookieSecure = secureStr == "true" || secureStr == "1"
	}

	return &Config{
		Port:            getEnv("PORT", "8080"),
		AppEnv:          appEnv,
		DatabaseURL:     getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/runtir_db?sslmode=disable"),
		JWTSecret:       getEnv("JWT_SECRET", "afriops-super-secret-jwt-key-2026-very-secure"),
		JWTExpiryHours:  jwtExpiryHours,
		CookieName:      getEnv("COOKIE_NAME", "access_token"),
		CookieSecure:    cookieSecure,
		CookieDomain:    getEnv("COOKIE_DOMAIN", ""),
		UploadDir:       getEnv("UPLOAD_DIR", "./uploads"),
		MaxUploadSizeMB: maxUploadSizeMB,
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}