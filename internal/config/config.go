package config

import (
	"fmt"
	"os"
	"time"
)

type Config struct {
	Env             string
	HTTPAddr        string
	DatabaseURL     string
	ShutdownTimeout time.Duration
}

func Load() (Config, error) {
	cfg := Config{
		Env:             getEnv("OPENCOPS_ENV", "local"),
		HTTPAddr:        getEnv("OPENCOPS_HTTP_ADDR", ":8080"),
		DatabaseURL:     getEnv("OPENCOPS_DATABASE_URL", "postgres://opencops:opencops@localhost:5432/opencops?sslmode=disable"),
		ShutdownTimeout: 10 * time.Second,
	}

	shutdownTimeoutRaw := getEnv("OPENCOPS_SHUTDOWN_TIMEOUT", "10s")

	shutdownTimeout, err := time.ParseDuration(shutdownTimeoutRaw)
	if err != nil {
		return Config{}, fmt.Errorf("invalid OPENCOPS_SHUTDOWN_TIMEOUT value %q: %w", shutdownTimeoutRaw, err)
	}

	cfg.ShutdownTimeout = shutdownTimeout

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
