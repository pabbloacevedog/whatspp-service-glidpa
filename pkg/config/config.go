package config

import (
	"os"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	// Application configuration
	AppEnv   string
	Port     string
	LogLevel string

	// CORS configuration
	CorsAllowedOrigins string

	// WhatsApp configuration
	WhatsAppSessionTimeout time.Duration

	// AI configuration
	GeminiAPIKey string

	// Database configuration
	PostgresURL string

	// Redis configuration
	RedisAddr string

	// JWT configuration
	JWTSecret  string
	JWTExpires time.Duration
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if it exists
	_ = godotenv.Load()

	// Parse WhatsApp session timeout
	whatsAppSessionTimeout, err := time.ParseDuration(getEnv("WHATSAPP_SESSION_TIMEOUT", "5m"))
	if err != nil {
		whatsAppSessionTimeout = 5 * time.Minute
	}

	// Parse JWT expiration time
	jwtExpires, err := time.ParseDuration(getEnv("JWT_EXPIRES", "1h"))
	if err != nil {
		jwtExpires = 1 * time.Hour
	}

	return &Config{
		// Application configuration
		AppEnv:   getEnv("APP_ENV", "development"),
		Port:     getEnv("PORT", "3000"),
		LogLevel: getEnv("LOG_LEVEL", "debug"),

		// CORS configuration
		CorsAllowedOrigins: getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:3000"),

		// WhatsApp configuration
		WhatsAppSessionTimeout: whatsAppSessionTimeout,

		// AI configuration
		GeminiAPIKey: getEnv("GEMINI_API_KEY", ""),

		// Database configuration
		PostgresURL: getEnv("POSTGRES_URL", "postgres://postgres:postgres@localhost:5432/whatsapp_service"),

		// Redis configuration
		RedisAddr: getEnv("REDIS_ADDR", "localhost:6379"),

		// JWT configuration
		JWTSecret:  getEnv("JWT_SECRET", "secret"),
		JWTExpires: jwtExpires,
	}, nil
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
