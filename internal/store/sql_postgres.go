package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/MKhiriev/go-pass-keeper/internal/config"
	"github.com/MKhiriev/go-pass-keeper/internal/logger"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// NewConnectPostgres opens a PostgreSQL connection using the pgx stdlib driver
// and the DSN supplied in cfg. It configures the connection pool, verifies
// reachability with a ping, and returns a [DB] value wired to a
// [PostgresErrorClassifier] for driver-level error classification.
//
// Returns an error if the driver cannot be opened, the ping fails, or the
// connection string is invalid.
func NewConnectPostgres(ctx context.Context, cfg config.DB, log *logger.Logger) (*DB, error) {
	// establish connection
	conn, err := sql.Open("pgx", cfg.DSN)
	if err != nil {
		log.Err(err).Str("func", "NewConnectPostgres").Msg("error occurred during database connection")
		return nil, fmt.Errorf("error occured during database connection: %w", err)
	}

	// setup connections
	conn.SetMaxOpenConns(10)
	conn.SetMaxOpenConns(4)

	// ping database
	err = conn.PingContext(ctx)
	if err != nil {
		log.Err(err).Str("func", "NewConnectPostgres").Msg("error connecting database (ping)")
		return nil, err
	}
	log.Debug().Str("func", "NewConnectPostgres").Msg("connected to database successfully")

	// construct a DB struct
	db := &DB{
		DB:                 conn,
		logger:             log,
		errorClassificator: NewPostgresErrorClassifier(),
	}

	return db, nil
}

func postgresError(err error) string {
	var pgErr *pgconn.PgError
	// if postgres returns error
	if errors.As(err, &pgErr) {
		return pgErr.Code
	}

	return ""
}
