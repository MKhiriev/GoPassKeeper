package store

import (
	"context"
	"fmt"

	"github.com/MKhiriev/go-pass-keeper/internal/logger"
	"github.com/MKhiriev/go-pass-keeper/models"
)

type privateDataRepository struct {
	*DB
	logger *logger.Logger
}

func NewPrivateDataRepository(db *DB, logger *logger.Logger) PrivateDataRepository {
	return &privateDataRepository{
		DB:     db,
		logger: logger,
	}
}

func (p *privateDataRepository) GetPrivateData(ctx context.Context, downloadRequest models.DownloadRequest) ([]models.PrivateData, error) {
	log := logger.FromContext(ctx)

	query, args, err := buildGetPrivateDataQuery(ctx, downloadRequest)
	if err != nil {
		log.Err(err).
			Str("func", "privateDataRepository.GetPrivateData").
			Int64("user_id", downloadRequest.UserID).
			Msg("failed to create query")
		return nil, err
	}

	rows, err := p.DB.QueryContext(ctx, query, args...)
	if err != nil {
		log.Err(err).
			Str("func", "privateDataRepository.GetPrivateData").
			Int64("user_id", downloadRequest.UserID).
			Int("client side ids count", len(downloadRequest.ClientSideIDs)).
			Msg("failed to execute query for getting requested private data")
		return nil, fmt.Errorf("failed to query requested private data: %w", err)
	}
	defer rows.Close()

	results := make([]models.PrivateData, 0, 50)

	for rows.Next() {
		var item models.PrivateData

		scanErr := rows.Scan(
			&item.ID,
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
				Int64("user_id", downloadRequest.UserID).
				Msg("failed to scan private data row")
			return nil, fmt.Errorf("failed to scan private data row: %w", scanErr)
		}

		results = append(results, item)
	}

	if rowsErr := rows.Err(); rowsErr != nil {
		log.Err(rowsErr).
			Str("func", "privateDataRepository.GetPrivateData").
			Int64("user_id", downloadRequest.UserID).
			Msg("error occurred during rows iteration")
		return nil, fmt.Errorf("error iterating private data rows: %w", rowsErr)
	}

	return results, nil
}

// GetAllPrivateData retrieves all vault items for a specific user.
// Returns an empty slice if no records found.
func (p *privateDataRepository) GetAllPrivateData(ctx context.Context, userID int64) ([]models.PrivateData, error) {
	log := logger.FromContext(ctx)

	rows, queryErr := p.DB.QueryContext(ctx, getAllUserPrivateData, userID)
	if queryErr != nil {
		log.Err(queryErr).
			Str("func", "privateDataRepository.GetAllPrivateData").
			Int64("user_id", userID).
			Msg("failed to execute query for getting all user private data")
		return nil, fmt.Errorf("failed to query user private data: %w", queryErr)
	}
	defer rows.Close()

	allData := make([]models.PrivateData, 0, 50)

	for rows.Next() {
		var data models.PrivateData

		scanErr := rows.Scan(
			&data.ID,
			&data.UserID,
			&data.Payload.Type,
			&data.Payload.Metadata,
			&data.Payload.Data,
			&data.Payload.Notes,
			&data.Payload.AdditionalFields,
			&data.CreatedAt,
			&data.UpdatedAt,
			&data.Version,
			&data.ClientSideID,
			&data.Hash,
			&data.Deleted,
		)
		if scanErr != nil {
			log.Err(scanErr).
				Str("func", "privateDataRepository.GetAllPrivateData").
				Int64("user_id", userID).
				Msg("failed to scan cipher row")
			return nil, fmt.Errorf("failed to scan cipher row: %w", scanErr)
		}

		allData = append(allData, data)
	}

	if err := rows.Err(); err != nil {
		log.Err(err).
			Str("func", "privateDataRepository.GetAllPrivateData").
			Int64("user_id", userID).
			Msg("error occurred during rows iteration")
		return nil, fmt.Errorf("error iterating cipher rows: %w", err)
	}

	return allData, nil
}

// checked!
func (p *privateDataRepository) GetAllStates(ctx context.Context, userID int64) ([]models.PrivateDataState, error) {
	log := logger.FromContext(ctx)

	rows, queryErr := p.DB.QueryContext(ctx, getAllUserDataState, userID)
	if queryErr != nil {
		log.Err(queryErr).
			Str("func", "privateDataRepository.GetAllStates").
			Int64("user_id", userID).
			Msg("failed to execute query for getting all user private data")
		return nil, fmt.Errorf("failed to query user private data states: %w", queryErr)
	}
	defer rows.Close()

	dataStates := make([]models.PrivateDataState, 0, 50)

	for rows.Next() {
		var data models.PrivateDataState

		scanErr := rows.Scan(
			&data.ClientSideID,
			&data.Hash,
			&data.Version,
			&data.Deleted,
			&data.UpdatedAt,
		)
		if scanErr != nil {
			log.Err(scanErr).
				Str("func", "privateDataRepository.GetAllStates").
				Int64("user_id", userID).
				Msg("failed to scan cipher row")
			return nil, fmt.Errorf("failed to scan cipher row: %w", scanErr)
		}

		dataStates = append(dataStates, data)
	}

	if err := rows.Err(); err != nil {
		log.Err(err).
			Str("func", "privateDataRepository.GetAllStates").
			Int64("user_id", userID).
			Msg("error occurred during rows iteration")
		return nil, fmt.Errorf("error iterating cipher rows: %w", err)
	}

	return dataStates, nil
}

func (p *privateDataRepository) GetStates(ctx context.Context, syncRequest models.SyncRequest) ([]models.PrivateDataState, error) {
	log := logger.FromContext(ctx)

	userID := syncRequest.UserID

	query, args, err := buildGetStatesSyncQuery(ctx, syncRequest)
	if err != nil {
		log.Err(err).
			Str("func", "privateDataRepository.GetStates").
			Int64("user_id", userID).
			Msg("failed to create query")
		return nil, err
	}

	rows, queryErr := p.DB.QueryContext(ctx, query, args...)
	if queryErr != nil {
		log.Err(queryErr).
			Str("func", "privateDataRepository.GetStates").
			Int64("user_id", userID).
			Msg("failed to execute query for getting all user private data")
		return nil, fmt.Errorf("failed to query user private data: %w", queryErr)
	}
	defer rows.Close()

	allData := make([]models.PrivateDataState, 0, 50)

	for rows.Next() {
		var data models.PrivateDataState

		scanErr := rows.Scan(
			&data.ClientSideID,
			&data.Hash,
			&data.Version,
			&data.Deleted,
			&data.UpdatedAt,
		)
		if scanErr != nil {
			log.Err(scanErr).
				Str("func", "privateDataRepository.GetStates").
				Int64("user_id", userID).
				Msg("failed to scan cipher row")
			return nil, fmt.Errorf("failed to scan cipher row: %w", scanErr)
		}

		allData = append(allData, data)
	}

	if err := rows.Err(); err != nil {
		log.Err(err).
			Str("func", "privateDataRepository.GetStates").
			Int64("user_id", userID).
			Msg("error occurred during rows iteration")
		return nil, fmt.Errorf("error iterating cipher rows: %w", err)
	}

	return allData, nil
}

func (p *privateDataRepository) UpdatePrivateData(ctx context.Context, updateRequest models.UpdateRequest) error {
	log := logger.FromContext(ctx)

	if len(updateRequest.PrivateDataUpdates) == 0 {
		log.Warn().
			Str("func", "privateDataRepository.UpdatePrivateData").
			Msg("no update requests provided")
		return nil
	}

	if len(updateRequest.PrivateDataUpdates) == 1 {
		return p.updateSingleRecord(ctx, updateRequest.PrivateDataUpdates[0])
	}

	return p.updateMultipleRecords(ctx, updateRequest.PrivateDataUpdates)
}

func (p *privateDataRepository) DeletePrivateData(ctx context.Context, deleteRequest models.DeleteRequest) error {
	log := logger.FromContext(ctx)

	if len(deleteRequest.DeleteEntries) == 0 {
		log.Warn().
			Str("func", "privateDataRepository.DeletePrivateData").
			Msg("no delete requests provided")
		return nil
	}

	if len(deleteRequest.DeleteEntries) == 1 {
		return p.deleteSingleRecord(ctx, deleteRequest)
	}

	return p.deleteMultipleRecords(ctx, deleteRequest)
}

func (p *privateDataRepository) deleteSingleRecord(ctx context.Context, deleteRequest models.DeleteRequest) error {
	log := logger.FromContext(ctx)

	entry := deleteRequest.DeleteEntries[0]

	log.Debug().
		Str("func", "privateDataRepository.deleteSingleRecord").
		Str("client_side_id", entry.ClientSideID).
		Int64("user_id", deleteRequest.UserID).
		Msg("soft-deleting single private data record")

	var updatedID *int64
	var currentDBVersion *int64

	queryRowErr := p.DB.QueryRowContext(ctx, deletePrivateDataQuery, entry.ClientSideID, deleteRequest.UserID, entry.Version).Scan(&updatedID, &currentDBVersion)
	if queryRowErr != nil {
		log.Err(queryRowErr).
			Str("func", "privateDataRepository.deleteSingleRecord").
			Str("client_side_id", entry.ClientSideID).
			Msg("failed to execute soft delete query")
		return fmt.Errorf("failed to delete private data: %w", queryRowErr)
	}

	// not found: target_record empty -> both NULL
	if currentDBVersion == nil {
		log.Warn().
			Str("func", "privateDataRepository.deleteSingleRecord").
			Str("client_side_id", entry.ClientSideID).
			Msg("record not found")
		return ErrPrivateDataNotFound
	}

	// found but not updated -> version mismatch
	if updatedID == nil {
		log.Error().
			Str("func", "privateDataRepository.deleteSingleRecord").
			Str("client_side_id", entry.ClientSideID).
			Int64("db_version", *currentDBVersion).
			Int64("provided_version", entry.Version).
			Msg("optimistic lock failed: version mismatch on delete")
		return ErrVersionConflict
	}

	log.Info().
		Str("func", "privateDataRepository.deleteSingleRecord").
		Str("client_side_id", entry.ClientSideID).
		Int64("deleted_id", *updatedID).
		Msg("successfully soft-deleted single private data record")

	return nil
}

func (p *privateDataRepository) deleteMultipleRecords(ctx context.Context, deleteRequest models.DeleteRequest) error {
	log := logger.FromContext(ctx)

	tx, err := p.DB.BeginTx(ctx, nil)
	if err != nil {
		log.Err(err).
			Str("func", "privateDataRepository.DeletePrivateData").
			Int("entries_count", len(deleteRequest.DeleteEntries)).
			Msg("failed to begin transaction")
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	for idx, entry := range deleteRequest.DeleteEntries {
		log.Debug().
			Str("func", "privateDataRepository.DeletePrivateData").
			Int("iteration", idx+1).
			Int("total", len(deleteRequest.DeleteEntries)).
			Str("client_side_id", entry.ClientSideID).
			Msg("soft-deleting private data in transaction")

		var updatedID *int64
		var currentDBVersion *int64

		queryRowErr := tx.QueryRowContext(ctx, deletePrivateDataQuery, entry.ClientSideID, deleteRequest.UserID, entry.Version).Scan(&updatedID, &currentDBVersion)
		if queryRowErr != nil {
			log.Err(queryRowErr).
				Str("func", "privateDataRepository.DeletePrivateData").
				Int("iteration", idx+1).
				Str("client_side_id", entry.ClientSideID).
				Msg("failed to execute soft delete query")
			return fmt.Errorf("failed to delete private data at index %d: %w", idx, queryRowErr)
		}

		// not found: target_record empty -> both NULL
		if currentDBVersion == nil {
			log.Warn().
				Str("func", "privateDataRepository.DeletePrivateData").
				Int("iteration", idx+1).
				Str("client_side_id", entry.ClientSideID).
				Msg("record not found")
			return ErrPrivateDataNotFound
		}

		// found but not updated -> version mismatch
		if updatedID == nil {
			log.Error().
				Str("func", "privateDataRepository.DeletePrivateData").
				Int("iteration", idx+1).
				Str("client_side_id", entry.ClientSideID).
				Int64("db_version", *currentDBVersion).
				Int64("provided_version", entry.Version).
				Msg("optimistic lock failed: version mismatch on delete")
			return fmt.Errorf("failed to delete private data at index %d: %w", idx, ErrVersionConflict)
		}

		log.Debug().
			Str("func", "privateDataRepository.DeletePrivateData").
			Int("iteration", idx+1).
			Str("client_side_id", entry.ClientSideID).
			Int64("deleted_id", *updatedID).
			Msg("soft-deleted record in current iteration")
	}

	if commitErr := tx.Commit(); commitErr != nil {
		log.Err(commitErr).
			Str("func", "privateDataRepository.DeletePrivateData").
			Int("entries_count", len(deleteRequest.DeleteEntries)).
			Msg("failed to commit transaction")
		return fmt.Errorf("failed to commit transaction: %w", commitErr)
	}

	log.Info().
		Str("func", "privateDataRepository.DeletePrivateData").
		Int64("user_id", deleteRequest.UserID).
		Int("entries_count", len(deleteRequest.DeleteEntries)).
		Msg("successfully soft-deleted private data")

	return nil
}

func (p *privateDataRepository) SavePrivateData(ctx context.Context, data ...*models.PrivateData) error {
	if len(data) == 1 {
		return p.saveSinglePrivateData(ctx, data[0])
	}

	return p.saveMultiplePrivateData(ctx, data)
}

// saveSinglePrivateData saves one private data point
func (p *privateDataRepository) saveSinglePrivateData(ctx context.Context, data *models.PrivateData) error {
	log := logger.FromContext(ctx)

	log.Debug().
		Str("client_side_id", data.ClientSideID).
		Int64("user_id", data.UserID).
		Msg("saving single private data record")

	// Используем QueryRowContext + Scan, чтобы получить созданный ID
	err := p.DB.QueryRowContext(ctx, savePrivateData,
		data.ClientSideID,
		data.UserID,
		data.Payload.Metadata,
		data.Payload.Type,
		data.Payload.Data,
		data.Payload.Notes,
		data.Payload.AdditionalFields,
		data.Version,
		data.Hash,
		data.CreatedAt,
	).Scan(&data.ID)

	if err != nil {
		log.Err(err).
			Str("func", "privateDataRepository.saveSinglePrivateData").
			Str("client_side_id", data.ClientSideID).
			Int64("user_id", data.UserID).
			Msg("failed to save private data")

		// Проверка на дубликат (если клиент прислал тот же client_side_id)
		// Можно добавить проверку на pq.Error с кодом 23505 (unique_violation)
		return fmt.Errorf("failed to save private data: %w", err)
	}

	return nil
}

// saveMultiplePrivateData saves one private data point using transaction and prepared statement
func (p *privateDataRepository) saveMultiplePrivateData(ctx context.Context, data []*models.PrivateData) error {
	log := logger.FromContext(ctx)

	tx, err := p.DB.BeginTx(ctx, nil)
	if err != nil {
		log.Err(err).
			Str("func", "privateDataRepository.saveMultiplePrivateData").
			Int("count", len(data)).
			Msg("failed to begin transaction")
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, savePrivateData)
	if err != nil {
		log.Err(err).
			Str("func", "privateDataRepository.saveMultiplePrivateData").
			Int("count", len(data)).
			Msg("failed to prepare statement")
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for idx, singleData := range data {
		log.Debug().
			Str("func", "privateDataRepository.saveMultiplePrivateData").
			Int("iteration", idx+1).
			Int("total", len(data)).
			Str("client_side_id", singleData.ClientSideID).
			Int64("user_id", singleData.UserID).
			Msg("saving private data in transaction")

		queryErr := stmt.QueryRowContext(ctx,
			singleData.ClientSideID,
			singleData.UserID,
			singleData.Payload.Metadata,
			singleData.Payload.Type,
			singleData.Payload.Data,
			singleData.Payload.Notes,
			singleData.Payload.AdditionalFields,
			singleData.Version,
			singleData.Hash,
			singleData.CreatedAt,
		).Scan(&singleData.ID)

		if queryErr != nil {
			log.Err(err).
				Str("func", "privateDataRepository.saveMultiplePrivateData").
				Int("iteration", idx+1).
				Str("client_side_id", singleData.ClientSideID).
				Msg("failed to execute prepared statement")
			return fmt.Errorf("failed to save private data at index %d: %w", idx, queryErr)
		}
	}

	if commitErr := tx.Commit(); commitErr != nil {
		log.Err(commitErr).
			Str("func", "privateDataRepository.saveMultiplePrivateData").
			Int("count", len(data)).
			Msg("failed to commit transaction")
		return fmt.Errorf("failed to commit transaction: %w", commitErr)
	}

	return nil
}

// updateSingleRecord updates singe record without transaction
func (p *privateDataRepository) updateSingleRecord(ctx context.Context, update models.PrivateDataUpdate) error {
	log := logger.FromContext(ctx)

	query, args, err := buildUpdateQuery(ctx, update)
	if err != nil {
		log.Err(err).
			Str("func", "privateDataRepository.updateSingleRecord").
			Str("id", update.ClientSideID).
			Msg("failed to build update query")
		return fmt.Errorf("failed to build update query: %w", err)
	}

	if len(args) == 2 {
		log.Warn().
			Str("func", "privateDataRepository.updateSingleRecord").
			Str("id", update.ClientSideID).
			Msg("no fields to update, skipping")
		return nil
	}

	var updatedID *int64
	var currentDBVersion *int64

	queryRowErr := p.DB.QueryRowContext(ctx, query, args...).Scan(&updatedID, &currentDBVersion)
	if queryRowErr != nil {
		log.Err(queryRowErr).
			Str("func", "privateDataRepository.updateSingleRecord").
			Str("id", update.ClientSideID).
			Msg("failed to execute update query")
		return fmt.Errorf("failed to update private data: %w", queryRowErr)
	}

	// no cipher record found: `target_record` is empty - both fields are NULL
	if currentDBVersion == nil {
		log.Warn().
			Str("func", "privateDataRepository.updateSingleRecord").
			Str("id", update.ClientSideID).
			Msg("record not found")
		return ErrPrivateDataNotFound
	}

	// cipher record found, but UPDATE didn't work - version mismatch
	if updatedID == nil {
		log.Warn().
			Str("func", "privateDataRepository.updateSingleRecord").
			Str("id", update.ClientSideID).
			Int64("db_version", *currentDBVersion).
			Int64("provided_version", update.Version).
			Msg("optimistic lock failed: version mismatch")
		return fmt.Errorf("failed to update private data: %w", ErrVersionConflict)
	}

	log.Info().
		Str("func", "privateDataRepository.updateSingleRecord").
		Str("id", update.ClientSideID).
		Msg("successfully updated private data")

	return nil
}

// updateMultipleRecords updates multiple records with transaction
func (p *privateDataRepository) updateMultipleRecords(ctx context.Context, updates []models.PrivateDataUpdate) error {
	log := logger.FromContext(ctx)

	tx, err := p.DB.BeginTx(ctx, nil)
	if err != nil {
		log.Err(err).
			Str("func", "privateDataRepository.updateMultipleRecords").
			Int("updates_count", len(updates)).
			Msg("failed to begin transaction")
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	for idx, update := range updates {
		query, args, buildErr := buildUpdateQuery(ctx, update)
		if buildErr != nil {
			log.Err(buildErr).
				Str("func", "privateDataRepository.updateMultipleRecords").
				Int("iteration", idx+1).
				Str("id", update.ClientSideID).
				Msg("failed to build update query")
			return fmt.Errorf("failed to build update query at index %d: %w", idx, buildErr)
		}

		log.Debug().
			Str("func", "privateDataRepository.updateMultipleRecords").
			Int("iteration", idx+1).
			Int("total", len(updates)).
			Str("id", update.ClientSideID).
			Msg("updating private data in transaction")

		var updatedID *int64
		var currentDBVersion *int64

		queryRowErr := tx.QueryRowContext(ctx, query, args...).Scan(&updatedID, &currentDBVersion)
		if queryRowErr != nil {
			log.Err(queryRowErr).
				Str("func", "privateDataRepository.updateMultipleRecords").
				Str("id", update.ClientSideID).
				Int("iteration", idx+1).
				Msg("failed to execute update query")
			return fmt.Errorf("failed to update private data at index %d: %w", idx, queryRowErr)
		}

		// no cipher record found: `target_record` is empty - both fields are NULL
		if currentDBVersion == nil {
			log.Warn().
				Str("func", "privateDataRepository.updateMultipleRecords").
				Int("iteration", idx+1).
				Str("id", update.ClientSideID).
				Msg("record not found")
			return ErrPrivateDataNotFound
		}

		// cipher record found, but UPDATE didn't work - version mismatch
		if updatedID == nil {
			log.Error().
				Str("func", "privateDataRepository.updateMultipleRecords").
				Int("iteration", idx+1).
				Str("id", update.ClientSideID).
				Int64("db_version", *currentDBVersion).
				Int64("provided_version", update.Version).
				Msg("optimistic lock failed: version mismatch")
			return fmt.Errorf("failed to update private data at index %d: %w", idx, ErrVersionConflict)
		}

		log.Debug().
			Str("func", "privateDataRepository.updateMultipleRecords").
			Int("iteration", idx+1).
			Str("id", update.ClientSideID).
			Msg("updated record in current iteration")
	}

	if commitErr := tx.Commit(); commitErr != nil {
		log.Err(commitErr).
			Str("func", "privateDataRepository.updateMultipleRecords").
			Int("updates_count", len(updates)).
			Msg("failed to commit transaction")
		return fmt.Errorf("failed to commit transaction: %w", commitErr)
	}

	log.Info().
		Str("func", "privateDataRepository.updateMultipleRecords").
		Int("updates_count", len(updates)).
		Msg("successfully updated multiple private data records")

	return nil
}
