package config

import (
	"os"
)

type Config struct {
	Port        string
	Environment string
	DatabaseURL string
	JWTSecret   string

	// Google SSO
	GoogleClientID     string
	GoogleClientSecret string

	// Apple SSO
	AppleTeamID     string
	AppleKeyID      string
	AppleClientID   string
	ApplePrivateKey string

	// Redis (for caching and real-time features)
	RedisURL string
}

func Load() *Config {
	return &Config{
		Port:        getEnv("PORT", "8080"),
		Environment: getEnv("ENVIRONMENT", "development"),
		DatabaseURL: getEnv("DATABASE_URL", "postgresql://postgres:114514@localhost:5432/postgres?sslmode=disable"),
		JWTSecret:   getEnv("JWT_SECRET", "your-secret-key-change-in-production"),

		GoogleClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),

		AppleTeamID:     getEnv("APPLE_TEAM_ID", ""),
		AppleKeyID:      getEnv("APPLE_KEY_ID", ""),
		AppleClientID:   getEnv("APPLE_CLIENT_ID", ""),
		ApplePrivateKey: getEnv("APPLE_PRIVATE_KEY", ""),

		RedisURL: getEnv("REDIS_URL", "redis://localhost:6379"),
	}
}

// LoadConfig 为 FX 依赖注入创建配置实例
func LoadConfig() *Config {
	return Load()
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
