package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port          string
	DatabaseURL   string
	Environment   string
	LogLevel      string
	CheckInterval int // seconds
}

func Load() *Config {
	return &Config{
		Port:          getEnv("PORT", ":7777"),
		DatabaseURL:   getEnv("DATABASE_URL", "postgres://localhost/pipeline_monitor?sslmode=disable"),
		Environment:   getEnv("ENVIRONMENT", "development"),
		LogLevel:      getEnv("LOG_LEVEL", "info"),
		CheckInterval: getEnvInt("CHECK_INTERVAL", 30),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}
