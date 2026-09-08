package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Port             string
	Environment      string
	Version          string
	DatabaseURL      string
	MigrationsPath   string
	AllowedOrigins   []string
	ProfilingEnabled bool
	ProfilingAddress string
	ShutdownTimeout  time.Duration
}

func New(version string) (Config, error) {
	timeout, err := time.ParseDuration(env("SHUTDOWN_TIMEOUT", "10s"))
	if err != nil {
		return Config{}, fmt.Errorf("parse SHUTDOWN_TIMEOUT: %w", err)
	}
	port := env("PORT", "8080")
	if _, err := strconv.ParseUint(port, 10, 16); err != nil {
		return Config{}, fmt.Errorf("invalid PORT: %w", err)
	}
	profilingEnabled, err := strconv.ParseBool(env("PROFILING_ENABLED", "false"))
	if err != nil {
		return Config{}, fmt.Errorf("parse PROFILING_ENABLED: %w", err)
	}
	return Config{
		Port:             port,
		Environment:      env("ENVIRONMENT", "development"),
		Version:          env("APP_VERSION", version),
		DatabaseURL:      env("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/orders?sslmode=disable"),
		MigrationsPath:   env("MIGRATIONS_PATH", "db"),
		AllowedOrigins:   csvEnv("CORS_ALLOWED_ORIGINS", "http://localhost:5173"),
		ProfilingEnabled: profilingEnabled,
		ProfilingAddress: env("PROFILING_ADDRESS", "127.0.0.1:6060"),
		ShutdownTimeout:  timeout,
	}, nil
}

func csvEnv(name, fallback string) []string {
	values := strings.Split(env(name, fallback), ",")
	result := make([]string, 0, len(values))
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			result = append(result, value)
		}
	}
	return result
}

func env(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}
