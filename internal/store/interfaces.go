// Package store provides data-access abstractions and repository implementations
// for persisting and querying application domain objects (users, vault items).
//
// It defines repository interfaces, concrete PostgreSQL-backed implementations,
// query builders, error classification, and sentinel errors used across
// the storage layer.
package store

import (
	"context"

	"github.com/MKhiriev/go-pass-keeper/models"
)

// PrivateDataStorage is a high-level facade that combines database and
// (in the future) file-based operations for vault items.
//
// It is the primary interface consumed by the service layer; implementations
// may coordinate between [PrivateDataRepository] (relational storage) and
// [PrivateDataFileStorage] (binary/file storage) transparently.
type PrivateDataStorage interface {
	// Save persists one or more new vault items.
	// Each item must have Version == 0 and a unique ClientSideID within the user's vault.
	// Returns [ErrPrivateDataNotSaved] if the insert produces zero affected rows.
	Save(ctx context.Context, data ...*models.PrivateData) error

	// Get retrieves vault items that match the criteria specified in downloadRequests.
	// When ClientSideIDs is non-empty, only items with those identifiers are returned;
	// otherwise all items belonging to the user are returned.
	Get(ctx context.Context, downloadRequests models.DownloadRequest) ([]models.PrivateData, error)

	// GetAll returns every vault item owned by the given user,
	// including soft-deleted records.
	GetAll(ctx context.Context, userID int64) ([]models.PrivateData, error)

	// GetAllStates returns lightweight state descriptors for every vault item
	// owned by the given user. The result contains only identity and
	// change-detection fields (ClientSideID, Hash, Version, Deleted, UpdatedAt),
	// without encrypted payloads.
	GetAllStates(ctx context.Context, userID int64) ([]models.PrivateDataState, error)

	// GetStates returns lightweight state descriptors for vault items
	// whose ClientSideIDs are listed in syncRequest.
	// Used during synchronization to determine which items need to be
	// fetched, pushed, or removed on the client.
	GetStates(ctx context.Context, syncRequest models.SyncRequest) ([]models.PrivateDataState, error)

	// Update applies a batch of partial updates described in updateRequests.
	// Each update uses optimistic locking: the provided Version must match
	// the current database version, otherwise [ErrVersionConflict] is returned.
	Update(ctx context.Context, updateRequests models.UpdateRequest) error

	// Delete performs a soft-delete of one or more vault items described
	// in deleteRequests. The records remain in the database with the Deleted
	// flag set to true so that clients can detect the deletion during sync.
	Delete(ctx context.Context, deleteRequests models.DeleteRequest) error
}

// PrivateDataRepository defines the relational database access contract
// for vault items stored in the "ciphers" table.
//
// Method semantics mirror [PrivateDataStorage] but operate directly
// against the SQL database without involving file storage.
type PrivateDataRepository interface {
	// SavePrivateData inserts one or more new vault items into the database.
	// Returns [ErrPrivateDataNotSaved] if no rows were affected.
	SavePrivateData(ctx context.Context, data ...*models.PrivateData) error

	// GetPrivateData retrieves vault items matching the given download criteria.
	// Filtering is applied by UserID and, optionally, by ClientSideIDs.
	GetPrivateData(ctx context.Context, downloadRequests models.DownloadRequest) ([]models.PrivateData, error)

	// GetAllPrivateData returns all vault items belonging to the specified user.
	GetAllPrivateData(ctx context.Context, userID int64) ([]models.PrivateData, error)

	// GetAllStates returns lightweight state descriptors for all vault items
	// owned by the specified user, without encrypted payload fields.
	GetAllStates(ctx context.Context, userID int64) ([]models.PrivateDataState, error)

	// GetStates returns lightweight state descriptors for vault items
	// whose ClientSideIDs are listed in syncRequest.
	GetStates(ctx context.Context, syncRequest models.SyncRequest) ([]models.PrivateDataState, error)

	// UpdatePrivateData applies a batch of partial updates with optimistic
	// locking. Returns [ErrVersionConflict] on version mismatch or
	// [ErrPrivateDataNotFound] if a targeted record does not exist.
	UpdatePrivateData(ctx context.Context, updateRequests models.UpdateRequest) error

	// DeletePrivateData soft-deletes vault items identified in deleteRequests.
	DeletePrivateData(ctx context.Context, deleteRequests models.DeleteRequest) error
}

// PrivateDataFileStorage defines the contract for persisting and retrieving
// vault items as binary files outside the relational database.
//
// This interface exists to support future offloading of large encrypted
// payloads (e.g. file attachments) to a filesystem or object store,
// keeping the database optimized for metadata and text-sized fields.
type PrivateDataFileStorage interface {
	// SaveBinaryDataToFile writes one or more vault items to a file
	// identified by fileName. The exact storage backend (local FS, S3, etc.)
	// is determined by the implementation.
	SaveBinaryDataToFile(ctx context.Context, fileName string, data ...models.PrivateData) error

	// LoadBinaryDataFromFile reads and deserializes vault items
	// previously saved to the file identified by fileName.
	LoadBinaryDataFromFile(ctx context.Context, fileName string) ([]models.PrivateData, error)
}

// UserRepository defines the database access contract for user accounts.
type UserRepository interface {
	// CreateUser persists a new user record and returns the created entity
	// with server-assigned fields (e.g. ID) populated.
	// Returns [ErrLoginAlreadyExists] if the login is already taken.
	CreateUser(ctx context.Context, user models.User) (models.User, error)

	// FindUserByLogin retrieves a user record matching the Login field
	// of the provided user model.
	// Returns [ErrNoUserWasFound] if no matching record exists.
	FindUserByLogin(ctx context.Context, user models.User) (models.User, error)
}

// ErrorClassificator defines a strategy for categorizing errors produced
// by persistence layers (e.g. PostgreSQL driver errors) into well-known
// application-level classifications.
//
// Implementations inspect the underlying driver error (error codes, types)
// and return a corresponding [ErrorClassification] value that higher layers
// can switch on without coupling to a specific database driver.
type ErrorClassificator interface {
	// Classify maps an error into a predefined [ErrorClassification] enum.
	// If the error is not recognized, the implementation should return
	// a generic/unknown classification rather than panicking.
	Classify(err error) ErrorClassification
}
