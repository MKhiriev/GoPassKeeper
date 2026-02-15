package migrations

import (
	"database/sql"
	"embed"
	"fmt"

	"github.com/pressly/goose/v3"
)

//go:embed *.sql
var embedMigrations embed.FS

func Migrate(db *sql.DB) error {
	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("pgx"); err != nil {
		return fmt.Errorf("migration error setting dialect for db: %w", err)
	}

	if err := goose.Up(db, "."); err != nil {
		return fmt.Errorf("migration error: %w", err)
	}

	return nil
}
