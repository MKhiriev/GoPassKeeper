// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Rasul Khiriev

package store

import (
	"context"
	"fmt"

	"github.com/MKhiriev/go-pass-keeper/internal/logger"
	"github.com/MKhiriev/go-pass-keeper/models"
)

type localPrivateDataRepository struct {
	*DB
	logger *logger.Logger
}

// NewLocalPrivateDataRepository constructs a [LocalPrivateDataRepository]
// backed by the provided SQLite [DB] connection. All query results are
// structured-logged using logger.
func NewLocalPrivateDataRepository(db *DB, logger *logger.Logger) LocalPrivateDataRepository {
	return &localPrivateDataRepository{
		DB:     db,
		logger: logger,
	}
}

// SavePrivateData implements [LocalPrivateDataRepository]. It upserts each item
// in data into the local SQLite store under userID. Upsert semantics allow both
// new records and server-downloaded updates to be stored without explicit
// conflict handling in the caller.
//
// Returns an error wrapping the driver error if any upsert fails.
func (l *localPrivateDataRepository) SavePrivateData(ctx context.Context, userID int64, data ...models.PrivateData) error {
	log := logger.FromContext(ctx)

	for _, item := range data {
		_, err := l.DB.ExecContext(ctx, saveSinglePrivateData,
			userID,
			item.Payload.Type,
			item.Payload.Metadata,
			item.Payload.Data,
			item.Payload.Notes,
			item.Payload.AdditionalFields,
			item.CreatedAt,
			item.UpdatedAt,
			item.Version,
			item.ClientSideID,
			item.Hash,
			item.Deleted,
		)
		if err != nil {
			log.Err(err).
				Str("func", "privateDataRepository.SavePrivateData").
				Int64("user_id", userID).
				Str("client_side_id", item.ClientSideID).
				Msg("failed to execute upsert for private data")
			return fmt.Errorf("failed to save private data (client_side_id=%s): %w", item.ClientSideID, err)
		}
	}

	return nil
}

// GetPrivateData implements [LocalPrivateDataRepository]. It returns the single
// vault item identified by clientSideID and userID. Returns an error if the
// item does not exist or if scanning the result row fails.
func (l *localPrivateDataRepository) GetPrivateData(ctx context.Context, clientSideID string, userID int64) (models.PrivateData, error) {
	log := logger.FromContext(ctx)

	var item models.PrivateData
	row := l.DB.QueryRowContext(ctx, getSinglePrivateData, userID, clientSideID)
	if row.Err() != nil {
		err := row.Err()
		log.Err(err).
			Str("func", "privateDataRepository.GetPrivateData").
			Int64("user_id", userID).
			Str("id", clientSideID).
			Msg("failed to execute query for getting requested private data")
		return models.PrivateData{}, fmt.Errorf("failed to query requested private data: %w", err)
	}

	scanErr := row.Scan(
		&item.UserID,
		&item.Payload.Type,
		&item.Payload.Metadata,
		&item.Payload.Data,
		&item.Payload.Notes,
		&item.Payload.AdditionalFields,
		&item.CreatedAt,
		&item.UpdatedAt,
		&item.Version,
		&item.ClientSideID,
		&item.Hash,
		&item.Deleted,
	)
	if scanErr != nil {
		log.Err(scanErr).
			Str("func", "privateDataRepository.GetPrivateData").
			Int64("user_id", userID).
			Msg("failed to scan private data row")
		return models.PrivateData{}, fmt.Errorf("failed to scan private data row: %w", scanErr)
	}

	return item, nil
}

// GetAllPrivateData implements [LocalPrivateDataRepository]. It returns all
// vault items owned by userID, including soft-deleted records. Returns an error
// if the query or any row-scan fails.
func (l *localPrivateDataRepository) GetAllPrivateData(ctx context.Context, userID int64) ([]models.PrivateData, error) {
	log := logger.FromContext(ctx)

	rows, err := l.DB.QueryContext(ctx, getAllPrivateData, userID)
	if err != nil {
		log.Err(err).
			Str("func", "privateDataRepository.GetAllPrivateData").
			Int64("user_id", userID).
			Msg("failed to execute query for getting all private data")
		return nil, fmt.Errorf("failed to query all private data: %w", err)
	}
	defer rows.Close()

	var items []models.PrivateData

	for rows.Next() {
		var item models.PrivateData

		scanErr := rows.Scan(
			&item.UserID,
			&item.Payload.Type,
			&item.Payload.Metadata,
			&item.Payload.Data,
			&item.Payload.Notes,
			&item.Payload.AdditionalFields,
			&item.CreatedAt,
			&item.UpdatedAt,
			&item.Version,
			&item.ClientSideID,
			&item.Hash,
			&item.Deleted,
		)
		if scanErr != nil {
			log.Err(scanErr).
				Str("func", "privateDataRepository.GetAllPrivateData").
				Int64("user_id", userID).
				Msg("failed to scan private data row")
			return nil, fmt.Errorf("failed to scan private data row: %w", scanErr)
		}

		items = append(items, item)
	}

	if rowsErr := rows.Err(); rowsErr != nil {
		log.Err(rowsErr).
			Str("func", "privateDataRepository.GetAllPrivateData").
			Int64("user_id", userID).
			Msg("error occurred during rows iteration")
		return nil, fmt.Errorf("error iterating private data rows: %w", rowsErr)
	}

	return items, nil
}

// GetAllStates implements [LocalPrivateDataRepository]. It returns lightweight
// state descriptors (ClientSideID, Hash, Version, Deleted, UpdatedAt) for all
// vault items owned by userID. Used by the sync planner to compare local and
// server states without loading encrypted payloads.
func (l *localPrivateDataRepository) GetAllStates(ctx context.Context, userID int64) ([]models.PrivateDataState, error) {
	log := logger.FromContext(ctx)

	rows, err := l.DB.QueryContext(ctx, getAllStates, userID)
	if err != nil {
		log.Err(err).
			Str("func", "privateDataRepository.GetAllStates").
			Int64("user_id", userID).
			Msg("failed to execute query for getting all states")
		return nil, fmt.Errorf("failed to query all states: %w", err)
	}
	defer rows.Close()

	var items []models.PrivateDataState

	for rows.Next() {
		var item models.PrivateDataState

		scanErr := rows.Scan(
			&item.Version,
			&item.ClientSideID,
			&item.Hash,
			&item.Deleted,
			&item.UpdatedAt,
		)
		if scanErr != nil {
			log.Err(scanErr).
				Str("func", "privateDataRepository.GetAllStates").
				Int64("user_id", userID).
				Msg("failed to scan private data state row")
			return nil, fmt.Errorf("failed to scan private data state row: %w", scanErr)
		}

		items = append(items, item)
	}

	if rowsErr := rows.Err(); rowsErr != nil {
		log.Err(rowsErr).
			Str("func", "privateDataRepository.GetAllStates").
			Int64("user_id", userID).
			Msg("error occurred during rows iteration")
		return nil, fmt.Errorf("error iterating private data state rows: %w", rowsErr)
	}

	return items, nil
}

// UpdatePrivateData implements [LocalPrivateDataRepository]. It overwrites the
// stored vault item with the values in data, identified by data.ClientSideID
// and data.UserID. The caller must populate Version, Hash, and UpdatedAt
// before calling this method. Returns an error if the UPDATE fails.
func (l *localPrivateDataRepository) UpdatePrivateData(ctx context.Context, data models.PrivateData) error {
	log := logger.FromContext(ctx)

	_, err := l.DB.ExecContext(ctx, updatePrivateData,
		data.Payload.Type,
		data.Payload.Metadata,
		data.Payload.Data,
		data.Payload.Notes,
		data.Payload.AdditionalFields,
		data.UpdatedAt,
		data.Version,
		data.Hash,
		data.Deleted,
		data.UserID,
		data.ClientSideID,
	)
	if err != nil {
		log.Err(err).
			Str("func", "privateDataRepository.UpdatePrivateData").
			Int64("user_id", data.UserID).
			Str("client_side_id", data.ClientSideID).
			Msg("failed to execute update for private data")
		return fmt.Errorf("failed to update private data (client_side_id=%s): %w", data.ClientSideID, err)
	}

	return nil
}

// DeletePrivateData implements [LocalPrivateDataRepository]. It soft-deletes
// the vault item identified by clientSideID and userID by setting its deleted
// flag. The record is retained so that the sync service can propagate the
// deletion to the server. Returns an error if the UPDATE fails.
func (l *localPrivateDataRepository) DeletePrivateData(ctx context.Context, clientSideID string, userID int64) error {
	log := logger.FromContext(ctx)

	_, err := l.DB.ExecContext(ctx, deletePrivateData, userID, clientSideID)
	if err != nil {
		log.Err(err).
			Str("func", "privateDataRepository.DeletePrivateData").
			Int64("user_id", userID).
			Str("client_side_id", clientSideID).
			Msg("failed to execute soft delete for private data")
		return fmt.Errorf("failed to delete private data (client_side_id=%s): %w", clientSideID, err)
	}

	return nil
}

// IncrementVersion implements [LocalPrivateDataRepository]. It increments the
// version counter of the vault item identified by clientSideID and userID by
// one. Called after a successful server-side write to keep the local record in
// sync with the server version. Returns an error if no matching record is found
// or if the UPDATE itself fails.
func (l *localPrivateDataRepository) IncrementVersion(ctx context.Context, clientSideID string, userID int64) error {
	log := logger.FromContext(ctx)

	result, err := l.DB.ExecContext(ctx, incrementVersion, clientSideID, userID)
	if err != nil {
		log.Err(err).
			Str("func", "privateDataRepository.IncrementVersion").
			Int64("user_id", userID).
			Str("client_side_id", clientSideID).
			Msg("failed to execute increment version for private data")
		return fmt.Errorf("failed to increment version (client_side_id=%s): %w", clientSideID, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Err(err).
			Str("func", "privateDataRepository.IncrementVersion").
			Int64("user_id", userID).
			Str("client_side_id", clientSideID).
			Msg("failed to get rows affected after increment version")
		return fmt.Errorf("failed to get rows affected (client_side_id=%s): %w", clientSideID, err)
	}

	if rowsAffected == 0 {
		log.Warn().
			Str("func", "privateDataRepository.IncrementVersion").
			Int64("user_id", userID).
			Str("client_side_id", clientSideID).
			Msg("no rows affected during increment version: record not found")
		return fmt.Errorf("failed to increment version: record not found (client_side_id=%s, user_id=%d)", clientSideID, userID)
	}

	return nil
}
