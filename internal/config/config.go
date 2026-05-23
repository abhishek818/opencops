package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Env             string
	HTTPAddr        string
	DatabaseURL     string
	ShutdownTimeout time.Duration
	AutoMigrate     bool
}

func Load() (Config, error) {
	cfg := Config{
		Env:             getEnv("OPENCOPS_ENV", "local"),
		HTTPAddr:        getEnv("OPENCOPS_HTTP_ADDR", ":8080"),
		DatabaseURL:     getEnv("OPENCOPS_DATABASE_URL", "postgres://opencops:opencops@localhost:5432/opencops?sslmode=disable"),
		ShutdownTimeout: 10 * time.Second,
		AutoMigrate:     true,
	}

	shutdownTimeoutRaw := getEnv("OPENCOPS_SHUTDOWN_TIMEOUT", "10s")

	shutdownTimeout, err := time.ParseDuration(shutdownTimeoutRaw)
	if err != nil {
		return Config{}, fmt.Errorf("invalid OPENCOPS_SHUTDOWN_TIMEOUT value %q: %w", shutdownTimeoutRaw, err)
	}

	autoMigrateRaw := getEnv("OPENCOPS_AUTO_MIGRATE", "true")

	autoMigrate, err := strconv.ParseBool(autoMigrateRaw)
	if err != nil {
		return Config{}, fmt.Errorf("invalid OPENCOPS_AUTO_MIGRATE value %q: %w", autoMigrateRaw, err)
	}

	cfg.ShutdownTimeout = shutdownTimeout
	cfg.AutoMigrate = autoMigrate

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
