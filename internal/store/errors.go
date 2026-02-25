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

var (
	ErrBuildingSQLQuery     = errors.New("error building sql query")
	ErrExecutingQuery       = errors.New("error executing sql query")
	ErrBeginningTransaction = errors.New("failed to begin transaction")
	ErrCommitingTransaction = errors.New("failed to commit transaction")
	ErrPreparingStatement   = errors.New("failed to prepare statement")
	ErrExecutingStatement   = errors.New("failed to executing statement")
	ErrScanningRow          = errors.New("failed to scan private data row")
	ErrScanningRows         = errors.New("failed to scan private data rows")
)
