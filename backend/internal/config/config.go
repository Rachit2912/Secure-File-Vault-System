package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all .env file values :
type Config struct {
	Port         string
	DBUrl        string
	JWTKey       string
	UserQuotaMB  int
	ApiRateLimit int
}

// AppConfig will be populated on app booting :
var AppConfig Config

// LoadConfig loads environment variables from `.env` into AppConfig. :
func LoadConfig() {
	// load from .env :
	if err := godotenv.Load(); err != nil {
		log.Println("ℹ️ No .env file found, relying on environment variables.")
	}

	// read with fallbacks :
	port := getEnv("PORT", "8080")
	dbURL := getEnv("DB_URL", "postgres://filevault_db:filevault_db@db:5432/filevault?sslmode=disable")
	jwtKey := getEnv("JWT_KEY", "supersecret")
	userQuotaMB := getEnvAsInt("USER_QUOTA_MB", 10)
	apiRateLimit := getEnvAsInt("API_RATE_LIMIT", 10)

	AppConfig = Config{
		Port:         port,
		DBUrl:        dbURL,
		JWTKey:       jwtKey,
		UserQuotaMB:  userQuotaMB,
		ApiRateLimit: apiRateLimit,
	}
}

// getEnv fetches env var or returns fallback :
func getEnv(key string, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

// getEnvAsInt fetches env var as int, with fallback if parse fails :
func getEnvAsInt(key string, fallback int) int {
	valStr := getEnv(key, "")
	if val, err := strconv.Atoi(valStr); err == nil {
		return val
	}
	return fallback
}
