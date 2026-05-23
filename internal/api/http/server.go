package http

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/abhishek818/opencops/internal/api/http/handlers"
	"github.com/abhishek818/opencops/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Server struct {
	httpServer *http.Server
	logger     *slog.Logger
}

func NewServer(
	cfg config.Config,
	db *pgxpool.Pool,
	logger *slog.Logger,
) *Server {
	mux := http.NewServeMux()

	handlers.RegisterHealthRoutes(mux, db, logger)

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           loggingMiddleware(logger, mux),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	return &Server{
		httpServer: server,
		logger:     logger,
	}
}

func (s *Server) Start() error {
	s.logger.Info(
		"starting http server",
		slog.String("addr", s.httpServer.Addr),
	)

	err := s.httpServer.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("http server failed: %w", err)
	}

	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("shutting down http server")

	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("http server shutdown failed: %w", err)
	}

	return nil
}

func loggingMiddleware(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startedAt := time.Now()

		next.ServeHTTP(w, r)

		logger.Info(
			"http request completed",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.String("remote_addr", r.RemoteAddr),
			slog.Duration("duration", time.Since(startedAt)),
		)
	})
}
