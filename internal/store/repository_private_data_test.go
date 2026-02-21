package store

import (
	"context"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/MKhiriev/go-pass-keeper/internal/logger"
	"github.com/MKhiriev/go-pass-keeper/internal/utils"
	"github.com/MKhiriev/go-pass-keeper/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestPrivateDataRepo creates a privateDataRepository backed by a sqlmock DB.
func newTestPrivateDataRepo(t *testing.T) (*privateDataRepository, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	log := logger.NewLogger("test")
	repo := &privateDataRepository{
		DB:     &DB{DB: db, logger: log},
		logger: log,
	}
	return repo, mock
}

// ctxWithUserID returns a context carrying the given userID, as expected by buildUpdateQuery.
func ctxWithUserID(userID int64) context.Context {
	return context.WithValue(context.Background(), utils.UserIDCtxKey, userID)
}

func TestUpdateSingleRecord_VersionMatch(t *testing.T) {
	repo, mock := newTestPrivateDataRepo(t)

	const (
		userID       = int64(10)
		recordID     = int64(42)
		clientSideID = "550e8400-e29b-41d4-a716-446655440000"
	)

	createdAt := time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC)

	initialNotes := models.CipheredNotes("encrypted-notes-v1")
	initialFields := models.CipheredCustomFields("encrypted-fields-v1")

	existingRecord := &models.PrivateData{
		ID:           recordID,
		ClientSideID: clientSideID,
		UserID:       userID,
		Payload: models.PrivateDataPayload{
			Metadata:         models.CipheredMetadata("encrypted-metadata-v1"),
			Type:             models.LoginPassword,
			Data:             models.CipheredData("encrypted-data-v1"),
			Notes:            &initialNotes,
			AdditionalFields: &initialFields,
		},
		Version:   1,
		CreatedAt: &createdAt,
	}

	mock.ExpectExec(`INSERT INTO ciphers`).
		WillReturnResult(sqlmock.NewResult(recordID, 1))

	err := repo.saveSinglePrivateData(ctxWithUserID(userID), existingRecord)
	require.NoError(t, err, "setup: failed to insert initial record")

	newMetadata := models.CipheredMetadata("encrypted-metadata-v2")
	newType := models.LoginPassword
	newData := models.CipheredData("encrypted-data-v2")
	newNotes := models.CipheredNotes("encrypted-notes-v2")
	newFields := models.CipheredCustomFields("encrypted-fields-v2")

	update := models.PrivateDataUpdate{
		ID:               recordID,
		ClientSideID:     clientSideID,
		Metadata:         &newMetadata,
		Type:             &newType,
		Data:             &newData,
		Notes:            &newNotes,
		AdditionalFields: &newFields,
		Version:          2,
	}

	rows := sqlmock.NewRows([]string{"id", "version"}).
		AddRow(recordID, int64(2))

	mock.ExpectQuery(`UPDATE ciphers`).
		WithArgs(
			sqlmock.AnyArg(), // metadata
			sqlmock.AnyArg(), // type
			sqlmock.AnyArg(), // data
			sqlmock.AnyArg(), // notes
			sqlmock.AnyArg(), // additional_fields
			int64(2),         // version
			clientSideID,     // client_side_id
			userID,           // user_id
		).
		WillReturnRows(rows)

	err = repo.updateSingleRecord(ctxWithUserID(userID), update)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateSingleRecord_VersionMismatch(t *testing.T) {
	repo, mock := newTestPrivateDataRepo(t)

	const (
		userID        = int64(10)
		recordID      = int64(42)
		clientSideID  = "550e8400-e29b-41d4-a716-446655440000"
		serverVersion = int64(5)
	)

	createdAt := time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC)

	initialNotes := models.CipheredNotes("encrypted-notes-v1")
	initialFields := models.CipheredCustomFields("encrypted-fields-v1")

	existingRecord := &models.PrivateData{
		ID:           recordID,
		ClientSideID: clientSideID,
		UserID:       userID,
		Payload: models.PrivateDataPayload{
			Metadata:         models.CipheredMetadata("encrypted-metadata-v1"),
			Type:             models.LoginPassword,
			Data:             models.CipheredData("encrypted-data-v1"),
			Notes:            &initialNotes,
			AdditionalFields: &initialFields,
		},
		Version:   1,
		CreatedAt: &createdAt,
	}

	mock.ExpectExec(`INSERT INTO ciphers`).
		WillReturnResult(sqlmock.NewResult(recordID, 1))

	err := repo.saveSinglePrivateData(ctxWithUserID(userID), existingRecord)
	require.NoError(t, err, "setup: failed to insert initial record")

	newMetadata := models.CipheredMetadata("encrypted-metadata-stale")
	newType := models.LoginPassword
	newData := models.CipheredData("encrypted-data-stale")
	newNotes := models.CipheredNotes("encrypted-notes-stale")
	newFields := models.CipheredCustomFields("encrypted-fields-stale")

	update := models.PrivateDataUpdate{
		ID:               recordID,
		ClientSideID:     clientSideID,
		Metadata:         &newMetadata,
		Type:             &newType,
		Data:             &newData,
		Notes:            &newNotes,
		AdditionalFields: &newFields,
		Version:          2,
	}

	rows := sqlmock.NewRows([]string{"id", "version"}).
		AddRow(nil, serverVersion)

	mock.ExpectQuery(`UPDATE ciphers`).
		WithArgs(
			sqlmock.AnyArg(), // metadata
			sqlmock.AnyArg(), // type
			sqlmock.AnyArg(), // data
			sqlmock.AnyArg(), // notes
			sqlmock.AnyArg(), // additional_fields
			int64(2),         // version (client's stale version)
			clientSideID,     // client_side_id
			userID,           // user_id
		).
		WillReturnRows(rows)

	err = repo.updateSingleRecord(ctxWithUserID(userID), update)
	assert.ErrorIs(t, err, ErrVersionConflict)
	assert.NoError(t, mock.ExpectationsWereMet())
}
