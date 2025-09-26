package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port         int
	Domain       string
	LogLevel     string
	ReadTimeout  int
	WriteTimeout int
}

func Load() *Config {
	return &Config{
		Port:         getEnvInt("PORT", 80),
		Domain:       getEnv("DOMAIN", "servit.app"),
		LogLevel:     getEnv("LOG_LEVEL", "info"),
		ReadTimeout:  getEnvInt("READ_TIMEOUT", 30),
		WriteTimeout: getEnvInt("WRITE_TIMEOUT", 30),
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
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
