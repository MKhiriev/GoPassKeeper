package store

import (
	"context"
	"fmt"

	"github.com/MKhiriev/go-pass-keeper/internal/logger"
	"github.com/MKhiriev/go-pass-keeper/models"
)

// privateDataRepository is the PostgreSQL-backed implementation of
// [PrivateDataRepository]. It executes all vault-item CRUD operations
// directly against the "ciphers" table using the embedded [*DB] connection.
//
// Every public method obtains a context-scoped logger via
// [logger.FromContext] so that all database interactions are traced
// with structured fields (user_id, client_side_id, iteration index, etc.).
type privateDataRepository struct {
	*DB
	logger *logger.Logger
}

// NewPrivateDataRepository constructs a [PrivateDataRepository] backed by
// the provided database connection and logger.
//
// The logger parameter is stored for fallback logging; most methods prefer
// the context-scoped logger obtained via [logger.FromContext].
func NewPrivateDataRepository(db *DB, logger *logger.Logger) PrivateDataRepository {
	return &privateDataRepository{
		DB:     db,
		logger: logger,
	}
}

// GetPrivateData retrieves vault items that match the criteria in downloadRequest.
//
// Filtering is always applied by UserID. When downloadRequest.ClientSideIDs is
// non-empty, an additional IN-clause narrows the result to those identifiers only.
//
// Returns the matched items or an error if the query fails, a row cannot be
// scanned, or an iteration error is detected after the result set is exhausted.
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
		return nil, fmt.Errorf("%w: %w", ErrExecutingQuery, err)
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
			return nil, fmt.Errorf("%w: %w", ErrScanningRow, scanErr)
		}

		results = append(results, item)
	}

	if rowsErr := rows.Err(); rowsErr != nil {
		log.Err(rowsErr).
			Str("func", "privateDataRepository.GetPrivateData").
			Int64("user_id", downloadRequest.UserID).
			Msg("error occurred during rows iteration")
		return nil, fmt.Errorf("%w: %w", ErrScanningRows, rowsErr)
	}

	return results, nil
}

// GetAllPrivateData retrieves every vault item owned by the given user,
// including soft-deleted records.
//
// Returns an empty slice when no records are found.
func (p *privateDataRepository) GetAllPrivateData(ctx context.Context, userID int64) ([]models.PrivateData, error) {
	log := logger.FromContext(ctx)

	rows, queryErr := p.DB.QueryContext(ctx, getAllUserPrivateData, userID)
	if queryErr != nil {
		log.Err(queryErr).
			Str("func", "privateDataRepository.GetAllPrivateData").
			Int64("user_id", userID).
			Msg("failed to execute query for getting all user private data")
		return nil, fmt.Errorf("%w: %w", ErrExecutingQuery, queryErr)
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
			return nil, fmt.Errorf("%w: %w", ErrScanningRow, scanErr)
		}

		allData = append(allData, data)
	}

	if err := rows.Err(); err != nil {
		log.Err(err).
			Str("func", "privateDataRepository.GetAllPrivateData").
			Int64("user_id", userID).
			Msg("error occurred during rows iteration")
		return nil, fmt.Errorf("%w: %w", ErrScanningRows, err)
	}

	return allData, nil
}

// GetAllStates returns lightweight [models.PrivateDataState] descriptors for
// every vault item owned by the given user.
//
// The result contains only identity and change-detection fields
// (ClientSideID, Hash, Version, Deleted, UpdatedAt) — no encrypted payloads.
// This is the primary method used at the start of a sync cycle when the
// client needs a full picture of the server-side state.
func (p *privateDataRepository) GetAllStates(ctx context.Context, userID int64) ([]models.PrivateDataState, error) {
	log := logger.FromContext(ctx)

	rows, queryErr := p.DB.QueryContext(ctx, getAllUserDataState, userID)
	if queryErr != nil {
		log.Err(queryErr).
			Str("func", "privateDataRepository.GetAllStates").
			Int64("user_id", userID).
			Msg("failed to execute query for getting all user private data")
		return nil, fmt.Errorf("%w: %w", ErrExecutingQuery, queryErr)
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
			return nil, fmt.Errorf("%w: %w", ErrScanningRow, scanErr)
		}

		dataStates = append(dataStates, data)
	}

	if err := rows.Err(); err != nil {
		log.Err(err).
			Str("func", "privateDataRepository.GetAllStates").
			Int64("user_id", userID).
			Msg("error occurred during rows iteration")
		return nil, fmt.Errorf("%w: %w", ErrScanningRows, err)
	}

	return dataStates, nil
}

// GetStates returns lightweight [models.PrivateDataState] descriptors for
// vault items whose ClientSideIDs are listed in syncRequest.
//
// This method is used during incremental synchronization: the client sends
// the IDs it knows about, and the server responds with the current state of
// each, allowing the client to decide what to fetch, push, or delete locally.
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
		return nil, fmt.Errorf("%w: %w", ErrExecutingQuery, queryErr)
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
			return nil, fmt.Errorf("%w: %w", ErrScanningRow, scanErr)
		}

		allData = append(allData, data)
	}

	if rowsErr := rows.Err(); rowsErr != nil {
		log.Err(rowsErr).
			Str("func", "privateDataRepository.GetStates").
			Int64("user_id", userID).
			Msg("error occurred during rows iteration")
		return nil, fmt.Errorf("%w: %w", ErrScanningRows, rowsErr)
	}

	return allData, nil
}

// UpdatePrivateData applies a batch of partial updates described in updateRequest.
//
// Routing strategy:
//   - Zero updates → no-op (returns nil with a warning log).
//   - Exactly one update → delegates to [updateSingleRecord] (no transaction overhead).
//   - Two or more updates → delegates to [updateMultipleRecords] (wrapped in a transaction).
//
// Each individual update uses optimistic locking: the provided Version must
// match the current database version, otherwise [ErrVersionConflict] is returned.
// If the target record does not exist, [ErrPrivateDataNotFound] is returned.
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

// DeletePrivateData performs a soft-delete of one or more vault items
// described in deleteRequest.
//
// Routing strategy mirrors [UpdatePrivateData]:
//   - Zero entries → no-op.
//   - One entry → [deleteSingleRecord] (no transaction).
//   - Two or more → [deleteMultipleRecords] (transaction).
//
// Soft-delete sets the "deleted" flag to true and bumps the version,
// preserving the record so that clients can detect the deletion during sync.
// Optimistic locking is enforced: [ErrVersionConflict] on mismatch,
// [ErrPrivateDataNotFound] if the record does not exist.
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

// deleteSingleRecord soft-deletes a single vault item without opening
// a database transaction.
//
// It executes the CTE-based delete query ([deletePrivateDataQuery]) that
// returns both the updated row ID and the current database version,
// enabling the caller to distinguish between "not found" (both NULL)
// and "version conflict" (updatedID NULL, currentDBVersion non-NULL).
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
		return fmt.Errorf("%w: %w", ErrExecutingQuery, queryRowErr)
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

// deleteMultipleRecords soft-deletes two or more vault items inside a single
// database transaction.
//
// The transaction is rolled back automatically (via defer) if any individual
// delete fails — either because the record is not found, a version conflict
// is detected, or the query itself errors. The commit is attempted only after
// all entries have been processed successfully.
func (p *privateDataRepository) deleteMultipleRecords(ctx context.Context, deleteRequest models.DeleteRequest) error {
	log := logger.FromContext(ctx)

	tx, err := p.DB.BeginTx(ctx, nil)
	if err != nil {
		log.Err(err).
			Str("func", "privateDataRepository.DeletePrivateData").
			Int("entries_count", len(deleteRequest.DeleteEntries)).
			Msg("failed to begin transaction")
		return fmt.Errorf("%w: %w", ErrBeginningTransaction, err)
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
			return fmt.Errorf("%w: %w", ErrExecutingQuery, queryRowErr)
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
		return fmt.Errorf("%w: %w", ErrCommitingTransaction, commitErr)
	}

	log.Info().
		Str("func", "privateDataRepository.DeletePrivateData").
		Int64("user_id", deleteRequest.UserID).
		Int("entries_count", len(deleteRequest.DeleteEntries)).
		Msg("successfully soft-deleted private data")

	return nil
}

// SavePrivateData persists one or more new vault items.
//
// Routing strategy:
//   - Exactly one item → [saveSinglePrivateData] (plain INSERT, no transaction).
//   - Two or more items → [saveMultiplePrivateData] (transaction with a prepared statement).
//
// On success each [models.PrivateData.ID] is populated with the server-assigned
// primary key returned by the INSERT … RETURNING id clause.
func (p *privateDataRepository) SavePrivateData(ctx context.Context, data ...*models.PrivateData) error {
	if len(data) == 1 {
		return p.saveSinglePrivateData(ctx, data[0])
	}

	return p.saveMultiplePrivateData(ctx, data)
}

// saveSinglePrivateData inserts a single vault item without opening a
// transaction.
//
// The generated database ID is written back into data.ID via the
// INSERT … RETURNING id clause.
func (p *privateDataRepository) saveSinglePrivateData(ctx context.Context, data *models.PrivateData) error {
	log := logger.FromContext(ctx)

	log.Debug().
		Str("client_side_id", data.ClientSideID).
		Int64("user_id", data.UserID).
		Msg("saving single private data record")

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

		return fmt.Errorf("%w: %w", ErrExecutingQuery, err)
	}

	return nil
}

// saveMultiplePrivateData inserts two or more vault items inside a single
// database transaction using a prepared statement for efficiency.
//
// The prepared statement is created once from [savePrivateData] and reused
// for every item. Each generated database ID is written back into the
// corresponding [models.PrivateData.ID] field.
//
// The transaction is rolled back automatically (via defer) if any individual
// insert fails; the commit is attempted only after all items succeed.
func (p *privateDataRepository) saveMultiplePrivateData(ctx context.Context, data []*models.PrivateData) error {
	log := logger.FromContext(ctx)

	tx, err := p.DB.BeginTx(ctx, nil)
	if err != nil {
		log.Err(err).
			Str("func", "privateDataRepository.saveMultiplePrivateData").
			Int("count", len(data)).
			Msg("failed to begin transaction")
		return fmt.Errorf("%w: %w", ErrBeginningTransaction, err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, savePrivateData)
	if err != nil {
		log.Err(err).
			Str("func", "privateDataRepository.saveMultiplePrivateData").
			Int("count", len(data)).
			Msg("failed to prepare statement")
		return fmt.Errorf("%w: %w", ErrPreparingStatement, err)
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
			return fmt.Errorf("%w: %w", ErrExecutingStatement, err)
		}
	}

	if commitErr := tx.Commit(); commitErr != nil {
		log.Err(commitErr).
			Str("func", "privateDataRepository.saveMultiplePrivateData").
			Int("count", len(data)).
			Msg("failed to commit transaction")
		return fmt.Errorf("%w: %w", ErrCommitingTransaction, commitErr)
	}

	return nil
}

// updateSingleRecord applies a partial update to a single vault item
// without opening a database transaction.
//
// The method builds a dynamic UPDATE query via [buildUpdateQuery], executes
// it, and inspects the CTE result to determine the outcome:
//   - Both updatedID and currentDBVersion are non-NULL → success.
//   - currentDBVersion is NULL → record not found ([ErrPrivateDataNotFound]).
//   - updatedID is NULL but currentDBVersion is non-NULL → version mismatch ([ErrVersionConflict]).
//
// If the built query contains only the two mandatory positional args
// (client_side_id, user_id) and no SET clauses beyond the default
// (updated_at, version bump), the method returns nil immediately as a no-op.
func (p *privateDataRepository) updateSingleRecord(ctx context.Context, update models.PrivateDataUpdate) error {
	log := logger.FromContext(ctx)

	query, args, err := buildUpdateQuery(ctx, update)
	if err != nil {
		log.Err(err).
			Str("func", "privateDataRepository.updateSingleRecord").
			Str("id", update.ClientSideID).
			Msg("failed to build update query")
		return err
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
		return fmt.Errorf("%w: %w", ErrExecutingQuery, queryRowErr)
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

// updateMultipleRecords applies partial updates to two or more vault items
// inside a single database transaction.
//
// Each update is built dynamically via [buildUpdateQuery] and executed
// within the transaction. The CTE result is inspected identically to
// [updateSingleRecord] (not found vs. version conflict).
//
// The transaction is rolled back automatically (via defer) if any individual
// update fails; the commit is attempted only after all updates succeed.
func (p *privateDataRepository) updateMultipleRecords(ctx context.Context, updates []models.PrivateDataUpdate) error {
	log := logger.FromContext(ctx)

	tx, err := p.DB.BeginTx(ctx, nil)
	if err != nil {
		log.Err(err).
			Str("func", "privateDataRepository.updateMultipleRecords").
			Int("updates_count", len(updates)).
			Msg("failed to begin transaction")
		return fmt.Errorf("%w: %w", ErrBeginningTransaction, err)
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
			return err
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
			return fmt.Errorf("%w: %w", ErrExecutingQuery, queryRowErr)
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
		return fmt.Errorf("%w: %w", ErrCommitingTransaction, commitErr)
	}

	log.Info().
		Str("func", "privateDataRepository.updateMultipleRecords").
		Int("updates_count", len(updates)).
		Msg("successfully updated multiple private data records")

	return nil
}
