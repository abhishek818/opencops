package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	httpapi "github.com/abhishek818/opencops/internal/api/http"
	"github.com/abhishek818/opencops/internal/config"
	"github.com/abhishek818/opencops/internal/store/postgres"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	cfg, err := config.Load()
	if err != nil {
		logger.Error(
			"failed to load config",
			slog.String("error", err.Error()),
		)
		os.Exit(1)
	}

	logger.Info(
		"starting opencops",
		slog.String("env", cfg.Env),
		slog.String("http_addr", cfg.HTTPAddr),
	)

	rootCtx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	db, err := postgres.NewPool(rootCtx, cfg.DatabaseURL)
	if err != nil {
		logger.Error(
			"failed to connect postgres",
			slog.String("error", err.Error()),
		)
		os.Exit(1)
	}
	defer db.Close()

	server := httpapi.NewServer(cfg, db, logger)

	serverErrCh := make(chan error, 1)

	go func() {
		serverErrCh <- server.Start()
	}()

	select {
	case <-rootCtx.Done():
		logger.Info("shutdown signal received")

	case err := <-serverErrCh:
		if err != nil {
			logger.Error(
				"http server stopped with error",
				slog.String("error", err.Error()),
			)
			os.Exit(1)
		}
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error(
			"graceful shutdown failed",
			slog.String("error", err.Error()),
		)
		os.Exit(1)
	}

	logger.Info(
		"opencops stopped successfully",
		slog.String("stopped_at", time.Now().UTC().Format(time.RFC3339)),
	)
}
