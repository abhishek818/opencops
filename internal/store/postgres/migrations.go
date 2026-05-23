package postgres

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"log/slog"
	"time"

	"github.com/pressly/goose/v3"

	// This blank import registers the "pgx" driver for database/sql.
	_ "github.com/jackc/pgx/v5/stdlib"
)

// migrationsFS embeds all SQL migration files into the Go binary.
// This makes Docker and single-binary deployment easier because the app does not
// depend on external migration files at runtime.
//
//go:embed migrations/*.sql
var migrationsFS embed.FS

func RunMigrations(
	ctx context.Context,
	databaseURL string,
	logger *slog.Logger,
) error {
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("set goose dialect: %w", err)
	}

	goose.SetBaseFS(migrationsFS)

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return fmt.Errorf("open migration database connection: %w", err)
	}
	defer db.Close()

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(5 * time.Minute)

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := db.PingContext(pingCtx); err != nil {
		return fmt.Errorf("ping migration database connection: %w", err)
	}

	logger.Info("running database migrations")

	if err := goose.UpContext(ctx, db, "migrations"); err != nil {
		return fmt.Errorf("run database migrations: %w", err)
	}

	logger.Info("database migrations completed")

	return nil
}
