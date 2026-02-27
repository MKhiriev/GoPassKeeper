package store

import "errors"

// Sentinel errors returned by repository methods to signal well-known failure
// conditions. Callers should use [errors.Is] to match against these values.
var (
	// ErrLoginAlreadyExists is returned when an attempt to register a new user
	// fails because a user with the same login already exists in the database.
	ErrLoginAlreadyExists = errors.New("login already exists")

	// ErrNoUserWasFound is returned when a query expected to match at least one
	// user record produces an empty result set.
	ErrNoUserWasFound = errors.New("no user was found")

	// ErrPrivateDataNotSaved is returned when an INSERT of one or more vault
	// items completes without error but the number of affected rows is zero,
	// indicating that no data was actually persisted.
	ErrPrivateDataNotSaved = errors.New("private data was not saved")

	// ErrPrivateDataNotFound is returned when a query or update targets a vault
	// item (identified by client_side_id and user_id) that does not exist
	// in the database.
	ErrPrivateDataNotFound = errors.New("private data was not found")

	// ErrVersionConflict is returned when an optimistic-locking check fails:
	// the version supplied by the client does not match the current version
	// stored in the database, meaning another device has modified the record
	// since the client last synchronized.
	ErrVersionConflict = errors.New("private data version conflict occurred")
)

// Low-level database operation errors. These are returned (or wrapped) by
// repository methods when a SQL-level operation fails before any domain logic
// can be applied.
var (
	// ErrBuildingSQLQuery is returned when constructing a parameterised SQL
	// query fails (e.g. invalid argument count or unsupported type).
	ErrBuildingSQLQuery = errors.New("error building sql query")

	// ErrExecutingQuery is returned when executing a SELECT or similar
	// read-only query against the database fails.
	ErrExecutingQuery = errors.New("error executing sql query")

	// ErrBeginningTransaction is returned when the database driver cannot
	// start a new transaction.
	ErrBeginningTransaction = errors.New("failed to begin transaction")

	// ErrCommitingTransaction is returned when committing an open transaction
	// fails. The transaction is considered rolled back at this point.
	ErrCommitingTransaction = errors.New("failed to commit transaction")

	// ErrPreparingStatement is returned when a SQL statement cannot be
	// prepared (e.g. syntax error or connection issue).
	ErrPreparingStatement = errors.New("failed to prepare statement")

	// ErrExecutingStatement is returned when executing a prepared DML
	// statement (INSERT, UPDATE, DELETE) fails.
	ErrExecutingStatement = errors.New("failed to executing statement")

	// ErrScanningRow is returned when scanning column values from a single
	// result row into a destination struct fails.
	ErrScanningRow = errors.New("failed to scan private data row")

	// ErrScanningRows is returned when scanning column values during
	// multi-row iteration fails, typically mid-result-set.
	ErrScanningRows = errors.New("failed to scan private data rows")
)
