package store

import (
	"context"
	"fmt"
	"strings"

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

func (p *privateDataRepository) SavePrivateData(ctx context.Context, data ...*models.PrivateData) error {
	log := logger.FromContext(ctx)

	saveErr := p.savePrivateData(ctx, data...)
	if saveErr != nil {
		log.Err(saveErr).Str("func", "privateDataRepository.SavePrivateData").Msg("error saving private data to a repository")
		return saveErr
	}

	return nil
}

func (p *privateDataRepository) GetPrivateData(ctx context.Context, downloadRequest models.DownloadRequest) ([]models.PrivateData, error) {
	log := logger.FromContext(ctx)

	queryBuilder := new(strings.Builder)
	queryBuilder.Grow(len(getRequestedPrivateData) * 2)

	queryBuilder.WriteString(getRequestedPrivateData)

	args := []any{downloadRequest.UserID}
	argIndex := 2

	if len(downloadRequest.IDs) > 0 {
		queryBuilder.WriteString(fmt.Sprintf(getRequestedPrivateDataWhereID, argIndex))
		args = append(args, downloadRequest.IDs)
		argIndex++
	}

	if len(downloadRequest.Types) > 0 {
		queryBuilder.WriteString(fmt.Sprintf(getRequestedPrivateDataWhereType, argIndex))
		args = append(args, downloadRequest.Types)
		argIndex++
	}

	query := queryBuilder.String()

	rows, err := p.DB.QueryContext(ctx, query, args...)
	if err != nil {
		log.Err(err).
			Str("func", "privateDataRepository.GetPrivateData").
			Int64("user_id", downloadRequest.UserID).
			Int("ids_count", len(downloadRequest.IDs)).
			Int("types_count", len(downloadRequest.Types)).
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
			&item.Metadata,
			&item.Type,
			&item.Data,
			&item.Notes,
			&item.AdditionalFields,
			&item.CreatedAt,
			&item.UpdatedAt,
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

	// Предварительная аллокация с разумным capacity
	allData := make([]models.PrivateData, 0, 50)

	for rows.Next() {
		var data models.PrivateData

		scanErr := rows.Scan(
			&data.ID,
			&data.UserID,
			&data.Metadata,
			&data.Type,
			&data.Data,
			&data.Notes,
			&data.AdditionalFields,
			&data.CreatedAt,
			&data.UpdatedAt,
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

func (p *privateDataRepository) UpdatePrivateData(ctx context.Context, updateRequest models.UpdateRequest) error {
	log := logger.FromContext(ctx)

	if len(updateRequest.PrivateDataUpdates) == 0 {
		log.Warn().
			Str("func", "privateDataRepository.UpdatePrivateData").
			Msg("no update requests provided")
		return nil
	}

	// Оптимизация: одно обновление без транзакции
	if len(updateRequest.PrivateDataUpdates) == 1 {
		return p.updateSingleRecord(ctx, updateRequest.PrivateDataUpdates[0])
	}

	// Множественное обновление через транзакцию
	return p.updateMultipleRecords(ctx, updateRequest.PrivateDataUpdates)
}

func (p *privateDataRepository) DeletePrivateData(ctx context.Context, deleteRequest models.DeleteRequest) error {
	log := logger.FromContext(ctx)

	// Защита: не удаляем без указания конкретных ID
	if len(deleteRequest.IDs) == 0 {
		log.Warn().
			Str("func", "privateDataRepository.deleteSingleRequest").
			Int64("user_id", deleteRequest.UserID).
			Msg("no IDs provided for deletion, skipping")
		return nil
	}

	result, err := p.DB.ExecContext(ctx, deletePrivateData, deleteRequest.UserID, deleteRequest.IDs)
	if err != nil {
		log.Err(err).
			Str("func", "privateDataRepository.deleteSingleRequest").
			Int64("user_id", deleteRequest.UserID).
			Int("ids_count", len(deleteRequest.IDs)).
			Msg("failed to execute delete query")
		return fmt.Errorf("failed to delete private data: %w", err)
	}

	rowsAffected, rowsErr := result.RowsAffected()
	if rowsErr != nil {
		log.Err(rowsErr).
			Str("func", "privateDataRepository.deleteSingleRequest").
			Int64("user_id", deleteRequest.UserID).
			Msg("failed to get rows affected")
		return fmt.Errorf("failed to get rows affected: %w", rowsErr)
	}

	log.Info().
		Str("func", "privateDataRepository.deleteSingleRequest").
		Int64("user_id", deleteRequest.UserID).
		Int("ids_count", len(deleteRequest.IDs)).
		Int64("deleted", rowsAffected).
		Msg("successfully deleted private data")

	return nil
}

func (p *privateDataRepository) savePrivateData(ctx context.Context, data ...*models.PrivateData) error {
	if len(data) == 1 {
		return p.saveSinglePrivateData(ctx, data[0])
	}

	return p.saveMultiplePrivateData(ctx, data)
}

// saveSinglePrivateData saves one private data point
func (p *privateDataRepository) saveSinglePrivateData(ctx context.Context, data *models.PrivateData) error {
	log := logger.FromContext(ctx)

	result, err := p.DB.ExecContext(ctx, savePrivateData,
		data.UserID,
		data.Metadata,
		data.Type,
		data.Data,
		data.Notes,
		data.AdditionalFields,
		data.CreatedAt,
	)
	if err != nil {
		log.Err(err).
			Str("func", "privateDataRepository.saveSinglePrivateData").
			Int64("id", data.ID).
			Int64("user_id", data.UserID).
			Msg("failed to execute query for saving private data")
		return fmt.Errorf("failed to save private data: %w", err)
	}

	rowsAffected, rowsErr := result.RowsAffected()
	if rowsErr != nil {
		log.Err(rowsErr).
			Str("func", "privateDataRepository.saveSinglePrivateData").
			Int64("id", data.ID).
			Msg("failed to get rows affected")
		return fmt.Errorf("failed to get rows affected: %w", rowsErr)
	}

	if rowsAffected == 0 {
		log.Error().
			Str("func", "privateDataRepository.saveSinglePrivateData").
			Int64("id", data.ID).
			Int64("user_id", data.UserID).
			Msg("private data was not saved (0 rows affected)")
		return ErrPrivateDataNotSaved
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
			Int64("id", singleData.ID).
			Int64("user_id", singleData.UserID).
			Msg("saving private data in transaction")

		result, execErr := stmt.ExecContext(ctx,
			singleData.UserID,
			singleData.Metadata,
			singleData.Type,
			singleData.Data,
			singleData.Notes,
			singleData.AdditionalFields,
			singleData.CreatedAt,
		)
		if execErr != nil {
			log.Err(execErr).
				Str("func", "privateDataRepository.saveMultiplePrivateData").
				Int("iteration", idx+1).
				Int64("id", singleData.ID).
				Int64("user_id", singleData.UserID).
				Msg("failed to execute prepared statement")
			return fmt.Errorf("failed to save private data at index %d: %w", idx, execErr)
		}

		rowsAffected, rowsErr := result.RowsAffected()
		if rowsErr != nil {
			log.Err(rowsErr).
				Str("func", "privateDataRepository.saveMultiplePrivateData").
				Int("iteration", idx+1).
				Int64("id", singleData.ID).
				Msg("failed to get rows affected")
			return fmt.Errorf("failed to get rows affected at index %d: %w", idx, rowsErr)
		}

		if rowsAffected == 0 {
			log.Error().
				Str("func", "privateDataRepository.saveMultiplePrivateData").
				Int("iteration", idx+1).
				Int64("id", singleData.ID).
				Int64("user_id", singleData.UserID).
				Msg("private data was not saved (0 rows affected)")
			return ErrPrivateDataNotSaved
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

	query, args, err := p.buildUpdateQuery(ctx, update)
	if err != nil {
		log.Err(err).
			Str("func", "privateDataRepository.updateSingleRecord").
			Int64("id", update.ID).
			Msg("failed to build update query")
		return fmt.Errorf("failed to build update query: %w", err)
	}

	if len(args) == 2 {
		log.Warn().
			Str("func", "privateDataRepository.updateSingleRecord").
			Int64("id", update.ID).
			Msg("no fields to update, skipping")
		return nil
	}

	result, execErr := p.DB.ExecContext(ctx, query, args...)
	if execErr != nil {
		log.Err(execErr).
			Str("func", "privateDataRepository.updateSingleRecord").
			Int64("id", update.ID).
			Msg("failed to execute update query")
		return fmt.Errorf("failed to update private data: %w", execErr)
	}

	rowsAffected, rowsErr := result.RowsAffected()
	if rowsErr != nil {
		log.Err(rowsErr).
			Str("func", "privateDataRepository.updateSingleRecord").
			Int64("id", update.ID).
			Msg("failed to get rows affected")
		return fmt.Errorf("failed to get rows affected: %w", rowsErr)
	}

	if rowsAffected == 0 {
		log.Warn().
			Str("func", "privateDataRepository.updateSingleRecord").
			Int64("id", update.ID).
			Msg("no rows updated (record not found or access denied)")
		return ErrPrivateDataNotFound
	}

	log.Info().
		Str("func", "privateDataRepository.updateSingleRecord").
		Int64("id", update.ID).
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
		query, args, buildErr := p.buildUpdateQuery(ctx, update)
		if buildErr != nil {
			log.Err(buildErr).
				Str("func", "privateDataRepository.updateMultipleRecords").
				Int("iteration", idx+1).
				Int64("id", update.ID).
				Msg("failed to build update query")
			return fmt.Errorf("failed to build update query at index %d: %w", idx, buildErr)
		}

		log.Debug().
			Str("func", "privateDataRepository.updateMultipleRecords").
			Int("iteration", idx+1).
			Int("total", len(updates)).
			Int64("id", update.ID).
			Msg("updating private data in transaction")

		result, execErr := tx.ExecContext(ctx, query, args...)
		if execErr != nil {
			log.Err(execErr).
				Str("func", "privateDataRepository.updateMultipleRecords").
				Int("iteration", idx+1).
				Int64("id", update.ID).
				Msg("failed to execute update query")
			return fmt.Errorf("failed to update private data at index %d: %w", idx, execErr)
		}

		rowsAffected, rowsErr := result.RowsAffected()
		if rowsErr != nil {
			log.Err(rowsErr).
				Str("func", "privateDataRepository.updateMultipleRecords").
				Int("iteration", idx+1).
				Int64("id", update.ID).
				Msg("failed to get rows affected")
			return fmt.Errorf("failed to get rows affected at index %d: %w", idx, rowsErr)
		}

		if rowsAffected == 0 {
			log.Warn().
				Str("func", "privateDataRepository.updateMultipleRecords").
				Int("iteration", idx+1).
				Int64("id", update.ID).
				Msg("no rows updated (record not found or access denied)")
			return ErrPrivateDataNotFound
		}

		log.Debug().
			Str("func", "privateDataRepository.updateMultipleRecords").
			Int("iteration", idx+1).
			Int64("id", update.ID).
			Int64("updated", rowsAffected).
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
