// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Rasul Khiriev

package store

import (
	"context"

	"github.com/MKhiriev/go-pass-keeper/internal/config"
	"github.com/MKhiriev/go-pass-keeper/internal/logger"
	"github.com/MKhiriev/go-pass-keeper/models"
)

// privateDataStorage is the default implementation of [PrivateDataStorage].
//
// It acts as a high-level orchestration layer that delegates relational
// operations to a [PrivateDataRepository] and may optionally coordinate
// binary/file persistence through [PrivateDataFileStorage].
//
// At present, all vault-item operations are routed to the repository.
// File storage support is conditionally initialized and reserved for
// future use cases involving large encrypted payloads.
type privateDataStorage struct {
	// repository provides all relational database operations
	// against the "ciphers" table.
	repository PrivateDataRepository

	// fileStorage optionally provides binary/file-based persistence.
	// If configuration does not enable file storage, this field remains nil.
	fileStorage PrivateDataFileStorage

	// logger is used for structured diagnostic logging at the storage layer.
	logger *logger.Logger
}

// NewPrivateDataStorage constructs a fully configured implementation of
// [PrivateDataStorage].
//
// The function initializes:
//   - a relational repository backed by the provided [DB],
//   - an optional file storage component if enabled by configuration.
//
// Parameters:
//   - db: the initialized database wrapper used for repository construction.
//   - cfg: storage-related configuration, including binary data directory.
//   - logger: structured logger used for diagnostic output.
//
// Behavior:
//   - Always initializes a [PrivateDataRepository].
//   - Initializes [PrivateDataFileStorage] only if cfg.Files.BinaryDataDir
//     is non-empty.
//
// Returns:
//   - A ready-to-use [PrivateDataStorage] instance suitable for injection
//     into the service layer.
func NewPrivateDataStorage(db *DB, cfg config.Storage, logger *logger.Logger) PrivateDataStorage {
	logger.Debug().Msg("creating private data storage")

	storage := new(privateDataStorage)

	repository := NewPrivateDataRepository(db, logger)
	storage.repository = repository
	storage.logger = logger

	if cfg.Files.BinaryDataDir != "" {
		fileStorage := NewPrivateDataFileStorage()
		storage.fileStorage = fileStorage
	}

	return storage
}

// Save persists one or more new vault items.
//
// This method delegates directly to [PrivateDataRepository.SavePrivateData].
// Each item must:
//
//   - belong to the authenticated user (resolved at repository level),
//   - have Version == 0 (initial insert),
//   - have a unique ClientSideID within the user's vault.
//
// Returns:
//   - nil if all records were successfully stored.
//   - An error if validation fails or the repository reports persistence issues.
func (p *privateDataStorage) Save(ctx context.Context, data ...*models.PrivateData) error {
	return p.repository.SavePrivateData(ctx, data...)
}

// Get retrieves vault items matching the criteria defined in downloadRequests.
//
// If downloadRequests.ClientSideIDs is non-empty, only matching records
// are returned. Otherwise, all items owned by the user are returned.
//
// The method delegates to [PrivateDataRepository.GetPrivateData].
//
// Returns:
//   - A slice of fully populated [models.PrivateData] records.
//   - An error if the repository layer fails.
func (p *privateDataStorage) Get(
	ctx context.Context,
	downloadRequests models.DownloadRequest,
) ([]models.PrivateData, error) {
	return p.repository.GetPrivateData(ctx, downloadRequests)
}

// GetAll returns every vault item belonging to the specified user.
//
// This includes both active and soft-deleted records.
// The method is typically used for full data export or administrative tasks.
//
// Delegates to [PrivateDataRepository.GetAllPrivateData].
func (p *privateDataStorage) GetAll(
	ctx context.Context,
	userID int64,
) ([]models.PrivateData, error) {
	return p.repository.GetAllPrivateData(ctx, userID)
}

// GetAllStates returns lightweight state descriptors for all vault items
// owned by the specified user.
//
// The returned slice contains only identity and change-detection fields
// (ClientSideID, Hash, Version, Deleted, UpdatedAt).
// Encrypted payload fields are intentionally excluded.
//
// This method is used during full synchronization cycles to determine
// divergence between client and server state.
//
// Delegates to [PrivateDataRepository.GetAllStates].
func (p *privateDataStorage) GetAllStates(
	ctx context.Context,
	userID int64,
) ([]models.PrivateDataState, error) {
	return p.repository.GetAllStates(ctx, userID)
}

// GetStates returns lightweight state descriptors for a subset of vault items
// specified in syncRequest.
//
// This is typically used during incremental synchronization, where the client
// sends a list of known ClientSideIDs and the server responds with current
// state information for conflict detection and reconciliation.
//
// Delegates to [PrivateDataRepository.GetStates].
func (p *privateDataStorage) GetStates(
	ctx context.Context,
	syncRequest models.SyncRequest,
) ([]models.PrivateDataState, error) {
	return p.repository.GetStates(ctx, syncRequest)
}

// Update applies a batch of partial updates to existing vault items.
//
// Each update entry carries its own Version for optimistic concurrency control.
// The repository will:
//
//   - Verify that the provided Version matches the current database version.
//   - Increment Version upon successful update.
//   - Return [ErrVersionConflict] if a mismatch is detected.
//
// Delegates to [PrivateDataRepository.UpdatePrivateData].
func (p *privateDataStorage) Update(
	ctx context.Context,
	updateRequests models.UpdateRequest,
) error {
	return p.repository.UpdatePrivateData(ctx, updateRequests)
}

// Delete performs a soft-delete of one or more vault items.
//
// The underlying repository:
//
//   - Sets Deleted = true,
//   - Increments Version,
//   - Preserves the row for synchronization consistency.
//
// Physical deletion is intentionally avoided to ensure clients can
// detect removals during synchronization cycles.
//
// Delegates to [PrivateDataRepository.DeletePrivateData].
func (p *privateDataStorage) Delete(
	ctx context.Context,
	deleteRequests models.DeleteRequest,
) error {
	return p.repository.DeletePrivateData(ctx, deleteRequests)
}
