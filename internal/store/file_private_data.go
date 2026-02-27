package store

import (
	"context"
	"errors"

	"github.com/MKhiriev/go-pass-keeper/models"
)

// privateDataFileStorage is the default implementation of
// [PrivateDataFileStorage]. It is intended to persist large or binary
// vault payloads outside the relational database — for example, on a local
// filesystem or in a cloud object store — so that the database only holds
// lightweight metadata and encrypted text fields.
//
// This implementation is currently a stub; all methods return
// "not implemented" errors and will be completed in a future iteration.
type privateDataFileStorage struct {
}

// NewPrivateDataFileStorage constructs a new [PrivateDataFileStorage] instance.
//
// The returned implementation is currently a no-op stub.
// Once finalized, it will accept configuration for the target storage
// backend (local path, S3 bucket, etc.).
func NewPrivateDataFileStorage() PrivateDataFileStorage {
	return &privateDataFileStorage{}
}

// SaveBinaryDataToFile persists one or more [models.PrivateData] records
// to a file identified by fileName.
//
// Parameters:
//   - ctx: cancellation and deadline context for the I/O operation.
//   - fileName: logical name (or path) of the destination file.
//   - data: one or more vault items whose binary payloads will be written.
//
// Returns an error if the write fails or the storage backend is unavailable.
//
// NOTE: This method is not yet implemented and will return an error
// on every call until the file storage backend is finalized.
func (p *privateDataFileStorage) SaveBinaryDataToFile(ctx context.Context, fileName string, data ...models.PrivateData) error {
	// TODO implement me!
	return errors.New("not implemented")
}

// LoadBinaryDataFromFile reads previously saved [models.PrivateData] records
// from a file identified by fileName.
//
// Parameters:
//   - ctx: cancellation and deadline context for the I/O operation.
//   - fileName: logical name (or path) of the source file.
//
// Returns the deserialized vault items and nil on success, or nil and an error
// if the file does not exist, is corrupted, or the storage backend is unavailable.
//
// NOTE: This method is not yet implemented and will return an error
// on every call until the file storage backend is finalized.
func (p *privateDataFileStorage) LoadBinaryDataFromFile(ctx context.Context, fileName string) ([]models.PrivateData, error) {
	// TODO implement me!
	return nil, errors.New("not implemented")
}
