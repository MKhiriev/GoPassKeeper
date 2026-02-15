package store

import (
	"database/sql"

	"github.com/MKhiriev/go-pass-keeper/internal/logger"
	"github.com/MKhiriev/go-pass-keeper/migrations"
)

type DB struct {
	*sql.DB
	errorClassificator ErrorClassificator
	logger             *logger.Logger
}

func (db *DB) Migrate() error {
	return migrations.Migrate(db.DB)
}
