package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	APP_NAME              string
	APP_ENV               string
	APP_PORT              string
	ENABLE_SWAGGER        bool
	ENABLE_METRICS        bool
	DB_DSN                string
	DB_LOG_LEVEL          string
	DB_READ_DSNS          []string
	DB_MAX_IDLE_CONNS     int
	DB_MAX_OPEN_CONNS     int
	DB_CONN_MAX_LIFETIME  time.Duration
	DB_CONN_MAX_IDLE_TIME time.Duration
	REDIS_URL             string
}

func Load() (*Config, error) {
	if err := loadDotEnv(); err != nil {
		return nil, fmt.Errorf("load .env: %w", err)
	}

	maxIdleConns, err := envIntOrDefault("DB_MAX_IDLE_CONNS", 10)
	if err != nil {
		return nil, err
	}

	maxOpenConns, err := envIntOrDefault("DB_MAX_OPEN_CONNS", 50)
	if err != nil {
		return nil, err
	}

	connMaxLifetime, err := envDurationOrDefault("DB_CONN_MAX_LIFETIME", time.Hour)
	if err != nil {
		return nil, err
	}

	connMaxIdleTime, err := envDurationOrDefault("DB_CONN_MAX_IDLE_TIME", 15*time.Minute)
	if err != nil {
		return nil, err
	}

	appEnv := envOrDefault("APP_ENV", "development")
	enableSwagger, err := envBoolOrDefault("ENABLE_SWAGGER", !isProductionEnv(appEnv))
	if err != nil {
		return nil, err
	}

	enableMetrics, err := envBoolOrDefault("ENABLE_METRICS", true)
	if err != nil {
		return nil, err
	}

	return &Config{
		APP_NAME:              envOrDefault("APP_NAME", "FastGo"),
		APP_ENV:               appEnv,
		APP_PORT:              envOrDefault("APP_PORT", "3005"),
		ENABLE_SWAGGER:        enableSwagger,
		ENABLE_METRICS:        enableMetrics,
		DB_DSN:                envValue("DB_DSN"),
		DB_LOG_LEVEL:          envValue("DB_LOG_LEVEL"),
		DB_READ_DSNS:          envList("DB_READ_DSNS"),
		DB_MAX_IDLE_CONNS:     maxIdleConns,
		DB_MAX_OPEN_CONNS:     maxOpenConns,
		DB_CONN_MAX_LIFETIME:  connMaxLifetime,
		DB_CONN_MAX_IDLE_TIME: connMaxIdleTime,
		REDIS_URL:             envValue("REDIS_URL"),
	}, nil
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
	value := envValue(key)
	if value == "" {
		return fallback
	}

	return value
}

func envValue(key string) string {
	return strings.TrimSpace(os.Getenv(key))
}

func envList(key string) []string {
	raw := envValue(key)
	if raw == "" {
		return nil
	}

	parts := strings.Split(raw, ",")
	values := make([]string, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		values = append(values, part)
	}

	return values
}

func envIntOrDefault(key string, fallback int) (int, error) {
	raw := envValue(key)
	if raw == "" {
		return fallback, nil
	}

	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("%s must be a valid integer: %w", key, err)
	}

	return value, nil
}

func envDurationOrDefault(key string, fallback time.Duration) (time.Duration, error) {
	raw := envValue(key)
	if raw == "" {
		return fallback, nil
	}

	value, err := time.ParseDuration(raw)
	if err != nil {
		return 0, fmt.Errorf("%s must be a valid duration: %w", key, err)
	}

	return value, nil
}

func envBoolOrDefault(key string, fallback bool) (bool, error) {
	raw := envValue(key)
	if raw == "" {
		return fallback, nil
	}

	value, err := strconv.ParseBool(raw)
	if err != nil {
		return false, fmt.Errorf("%s must be a valid boolean: %w", key, err)
	}

	return value, nil
}

func isProductionEnv(env string) bool {
	switch strings.ToLower(strings.TrimSpace(env)) {
	case "prod", "production":
		return true
	default:
		return false
	}
}
