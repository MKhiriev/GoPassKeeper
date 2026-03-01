// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Rasul Khiriev

package store

import (
	"database/sql"

	"github.com/MKhiriev/go-pass-keeper/internal/logger"
	"github.com/MKhiriev/go-pass-keeper/migrations"
)

// DB represents the primary database wrapper used by the application.
//
// It embeds *sql.DB to expose the standard database/sql API while extending
// it with infrastructure-specific dependencies such as:
//
//   - ErrorClassificator: used to normalize and classify low-level database
//     errors into domain-level errors.
//   - logger.Logger: used for structured logging of database operations.
//
// This struct acts as the root dependency for repository layers and
// migration execution.
type DB struct {
	// DB is the underlying SQL connection pool.
	// It is embedded to allow direct access to database/sql methods.
	*sql.DB

	// errorClassificator classifies database-specific errors
	// (e.g., constraint violations, not found, conflicts)
	// into higher-level application errors.
	errorClassificator ErrorClassificator

	// logger is used for structured logging of database-related events,
	// failures, and diagnostic information.
	logger *logger.Logger
}

// Migrate executes all pending database schema migrations.
//
// It delegates migration execution to the migrations package,
// applying all unapplied migration files in order.
//
// The method should typically be called during application startup
// to ensure the database schema is in sync with the expected version.
//
// Returns:
//   - nil if all migrations were applied successfully.
//   - an error if migration execution fails at any stage.
func (db *DB) Migrate() error {
	return migrations.Migrate(db.DB)
}
