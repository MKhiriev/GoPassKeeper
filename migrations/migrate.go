// Package migrations manages database schema migrations for the application.
// It uses the goose migration library with embedded SQL files,
// ensuring that all migration files are compiled into the binary
// and applied automatically at startup without requiring external file access.
package migrations

import (
	"database/sql"
	"embed"
	"fmt"

	"github.com/pressly/goose/v3"
)

// embedMigrations holds all *.sql migration files embedded into the binary
// at compile time via the go:embed directive.
// This ensures migrations are always available regardless of the working directory
// or deployment environment.
//
//go:embed *.sql
var embedMigrations embed.FS

// Migrate applies all pending database migrations using the goose library.
//
// It configures goose to use the embedded filesystem and the pgx dialect,
// then runs all unapplied migrations in ascending order.
//
// This function is intended to be called once at application startup,
// before the database is used by any other component.
//
// Parameters:
//
//	db - an open *sql.DB connection to the target PostgreSQL database
//
// Returns:
//
//	error - non-nil if setting the dialect or applying migrations fails
//
// Example usage:
//
//	if err := migrations.Migrate(db); err != nil {
//	    log.Fatalf("failed to run migrations: %v", err)
//	}
func Migrate(db *sql.DB) error {
	if db == nil {
		return fmt.Errorf("migration error: db is nil")
	}

	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("pgx"); err != nil {
		return fmt.Errorf("migration error setting dialect for db: %w", err)
	}

	if err := goose.Up(db, "."); err != nil {
		return fmt.Errorf("migration error: %w", err)
	}

	return nil
}
