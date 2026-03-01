// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Rasul Khiriev

package store

import (
	"context"

	"github.com/MKhiriev/go-pass-keeper/models"
)

//go:generate mockgen -source=client_interfaces.go -destination=../mock/client_store_mock.go -package=mock

// LocalPrivateDataRepository is the low-level SQLite-backed repository used by
// the client application to persist and query encrypted vault items locally.
//
// All methods operate on a single user's data identified by userID. Changes are
// stored in the local database only; synchronisation with the server is handled
// by a separate service layer.
type LocalPrivateDataRepository interface {
	// SavePrivateData inserts or upserts one or more vault items for userID.
	// An upsert is used so that items downloaded from the server can be stored
	// without conflicting with locally created records that share the same
	// ClientSideID.
	SavePrivateData(ctx context.Context, userID int64, data ...models.PrivateData) error

	// GetPrivateData retrieves the single vault item identified by clientSideID
	// and userID. Returns an error if the item does not exist or a database
	// error occurs.
	GetPrivateData(ctx context.Context, clientSideID string, userID int64) (models.PrivateData, error)

	// GetAllPrivateData returns every vault item owned by userID, including
	// soft-deleted records. Used by the service layer to decrypt and present
	// the full vault to the user.
	GetAllPrivateData(ctx context.Context, userID int64) ([]models.PrivateData, error)

	// GetAllStates returns lightweight state descriptors (ClientSideID, Hash,
	// Version, Deleted, UpdatedAt) for all vault items owned by userID.
	// Used by the sync planner to compare local and server states without
	// transferring full encrypted payloads.
	GetAllStates(ctx context.Context, userID int64) ([]models.PrivateDataState, error)

	// UpdatePrivateData overwrites an existing vault item in the local database
	// with the field values contained in data. The caller is responsible for
	// populating Version, Hash, and UpdatedAt before calling this method.
	UpdatePrivateData(ctx context.Context, data models.PrivateData) error

	// DeletePrivateData performs a soft-delete of the vault item identified by
	// clientSideID and userID, setting its deleted flag so that the sync
	// service can propagate the deletion to the server.
	DeletePrivateData(ctx context.Context, clientSideID string, userID int64) error

	// IncrementVersion increments the local version counter of the vault item
	// identified by clientSideID and userID by one. Called after a successful
	// server-side write (update or delete) to keep the local version in sync
	// with the server. Returns an error if the record is not found.
	IncrementVersion(ctx context.Context, clientSideID string, userID int64) error
}
