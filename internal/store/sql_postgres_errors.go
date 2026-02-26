package store

import (
	"errors"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

// ErrorClassification is the result type returned by [ErrorClassificator.Classify]
// and [PostgresErrorClassifier.Classify]. It indicates whether a failed database
// operation should be retried or abandoned.
type ErrorClassification int

// PostgresErrorClassifier implements [ErrorClassificator] for PostgreSQL.
// It inspects the pgconn error code returned by the pgx driver and maps it
// to a [ErrorClassification] value.
type PostgresErrorClassifier struct{}

var (
	// ErrNotFound is a sentinel error used when a queried resource does not
	// exist in the database.
	ErrNotFound = errors.New("metric is not found")
)

const (
	// NonRetryable indicates that the failed operation should not be retried.
	// This is the default classification for unrecognised errors, constraint
	// violations, syntax errors, and data exceptions.
	NonRetryable ErrorClassification = iota

	// Retryable indicates that the failed operation may succeed if attempted
	// again (e.g. after a transient connection loss or a deadlock rollback).
	Retryable
)

// NewPostgresErrorClassifier constructs a [PostgresErrorClassifier] ready for use.
func NewPostgresErrorClassifier() *PostgresErrorClassifier {
	return &PostgresErrorClassifier{}
}

// Classify implements [ErrorClassificator]. It attempts to unwrap err as a
// *pgconn.PgError and delegates to [ClassifyPgError]. If err is nil or is not
// a PostgreSQL driver error, [NonRetryable] is returned.
func (c *PostgresErrorClassifier) Classify(err error) ErrorClassification {
	if err == nil {
		return NonRetryable
	}

	// Attempt to unwrap to a pgconn.PgError.
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return ClassifyPgError(pgErr)
	}

	// Default: treat unrecognised errors as non-retryable.
	return NonRetryable
}

// ClassifyPgError maps a *pgconn.PgError to an [ErrorClassification] based on
// the PostgreSQL error code.
// See https://www.postgresql.org/docs/current/errcodes-appendix.html for the
// full list of PostgreSQL error codes.
//
// Retryable codes:
//   - Class 08 — connection exceptions (08000, 08003, 08006)
//   - Class 40 — transaction rollback, serialization failure, deadlock (40000, 40001, 40P01)
//   - Class 57 — cannot connect now (57P03)
//
// NonRetryable codes:
//   - Class 22 — data exceptions
//   - Class 23 — integrity constraint violations
//   - Class 42 — syntax errors and access rule violations
//
// Any code not listed above is classified as [NonRetryable].
func ClassifyPgError(pgErr *pgconn.PgError) ErrorClassification {
	switch pgErr.Code {
	// Class 08 — connection exceptions
	case pgerrcode.ConnectionException,
		pgerrcode.ConnectionDoesNotExist,
		pgerrcode.ConnectionFailure:
		return Retryable

	// Class 40 — transaction rollback
	case pgerrcode.TransactionRollback, // 40000
		pgerrcode.SerializationFailure, // 40001
		pgerrcode.DeadlockDetected:     // 40P01
		return Retryable

	// Class 57 — operator intervention
	case pgerrcode.CannotConnectNow: // 57P03
		return Retryable
	}

	switch pgErr.Code {
	// Class 22 — data exceptions
	case pgerrcode.DataException,
		pgerrcode.NullValueNotAllowedDataException:
		return NonRetryable

	// Class 23 — integrity constraint violations
	case pgerrcode.IntegrityConstraintViolation,
		pgerrcode.RestrictViolation,
		pgerrcode.NotNullViolation,
		pgerrcode.ForeignKeyViolation,
		pgerrcode.UniqueViolation,
		pgerrcode.CheckViolation:
		return NonRetryable

	// Class 42 — syntax errors or access rule violations
	case pgerrcode.SyntaxErrorOrAccessRuleViolation,
		pgerrcode.SyntaxError,
		pgerrcode.UndefinedColumn,
		pgerrcode.UndefinedTable,
		pgerrcode.UndefinedFunction:
		return NonRetryable
	}

	// Default: treat unrecognised codes as non-retryable.
	return NonRetryable
}
