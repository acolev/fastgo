package config

import (
	"os"
	"strings"
)

type Config struct {
	APP_PORT  string
	DB_DSN    string
	REDIS_URL string
}

func Load() *Config {
	_ = loadDotEnv()

	return &Config{
		APP_PORT:  envOrDefault("APP_PORT", "3005"),
		DB_DSN:    os.Getenv("DB_DSN"),
		REDIS_URL: os.Getenv("REDIS_URL"),
	}
}

func (c *Config) AppAddr() string {
	port := strings.TrimSpace(c.APP_PORT)
	if port == "" {
		port = "3005"
	}

	if strings.Contains(port, ":") {
		return port
	}

	return ":" + port
}

func envOrDefault(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	return value
}
