package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config holds all application configuration settings
type Config struct {
	Port              string
	DatabaseURL       string
	SupabaseJWTSecret string
	GinMode           string
	AllowedOrigins    string
}

// LoadConfig loads configuration from environment variables or .env file
func LoadConfig() (*Config, error) {
	// Attempt to load .env file, ignore error if missing (e.g. in production)
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found or error reading .env file; falling back to system environment variables")
	}

	port := getEnv("PORT", "8080")
	dbURL := getEnv("DATABASE_URL", "")
	jwtSecret := getEnv("SUPABASE_JWT_SECRET", "")
	ginMode := getEnv("GIN_MODE", "debug")
	allowedOrigins := getEnv("ALLOWED_ORIGINS", "http://localhost:3000,http://127.0.0.1:3000")

	return &Config{
		Port:              port,
		DatabaseURL:       dbURL,
		SupabaseJWTSecret: jwtSecret,
		GinMode:           ginMode,
		AllowedOrigins:    allowedOrigins,
	}, nil
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		return val
	}
	return fallback
}
