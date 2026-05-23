package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type healthResponse struct {
	Status    string `json:"status"`
	Service   string `json:"service"`
	Timestamp string `json:"timestamp"`
}

type readinessResponse struct {
	Status    string `json:"status"`
	Service   string `json:"service"`
	Database  string `json:"database"`
	Timestamp string `json:"timestamp"`
}

type errorResponse struct {
	Error     string `json:"error"`
	Service   string `json:"service"`
	Timestamp string `json:"timestamp"`
}

func RegisterHealthRoutes(
	mux *http.ServeMux,
	db *pgxpool.Pool,
	logger *slog.Logger,
) {
	mux.HandleFunc("GET /healthz", handleHealth())
	mux.HandleFunc("GET /readyz", handleReadiness(db, logger))
}

func handleHealth() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		response := healthResponse{
			Status:    "ok",
			Service:   "opencops",
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		}

		writeJSON(w, http.StatusOK, response)
	}
}

func handleReadiness(db *pgxpool.Pool, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()

		if err := db.Ping(ctx); err != nil {
			logger.Error(
				"readiness check failed",
				slog.String("component", "postgres"),
				slog.String("error", err.Error()),
			)

			response := errorResponse{
				Error:     "database is not ready",
				Service:   "opencops",
				Timestamp: time.Now().UTC().Format(time.RFC3339),
			}

			writeJSON(w, http.StatusServiceUnavailable, response)
			return
		}

		response := readinessResponse{
			Status:    "ready",
			Service:   "opencops",
			Database:  "ok",
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		}

		writeJSON(w, http.StatusOK, response)
	}
}

func writeJSON(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}
