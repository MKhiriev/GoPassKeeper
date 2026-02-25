package store

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	_ "github.com/mattn/go-sqlite3"

	"github.com/MKhiriev/go-pass-keeper/internal/config"
	"github.com/MKhiriev/go-pass-keeper/internal/logger"
)

func NewConnectSQLite(ctx context.Context, cfg config.ClientDB, log *logger.Logger) (*DB, error) {
	// db will be in file
	if err := createLocalDBFileIfNotExists(cfg.DSN); err != nil {
		log.Err(err).Str("func", "NewConnectSQLite").Msg("error creating database file")
		return nil, fmt.Errorf("error creating database file")
	}

	conn, err := sql.Open("sqlite3", cfg.DSN)
	if err != nil {
		log.Err(err).Str("func", "NewConnectSQLite").Msg("error connecting database")
		return nil, fmt.Errorf("error opening connection to DB")
	}

	// ping database
	err = conn.PingContext(ctx)
	if err != nil {
		log.Err(err).Str("func", "NewConnectSQLite").Msg("error connecting database (ping)")
		return nil, err
	}
	log.Debug().Str("func", "NewConnectSQLite").Msg("connected to database successfully")

	// construct a DB struct
	db := &DB{
		DB:     conn,
		logger: log,
	}

	return db, nil
}

func createLocalDBFileIfNotExists(dbFile string) error {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		// if not found - create
		f, err := os.Create(dbFile)
		if err != nil {
			return fmt.Errorf("error creating DB file: %w", err)
		}
		f.Close()
	}

	// file already exists
	return nil
}
