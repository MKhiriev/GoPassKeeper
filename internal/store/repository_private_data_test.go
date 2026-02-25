package store

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"regexp"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/MKhiriev/go-pass-keeper/internal/logger"
	"github.com/MKhiriev/go-pass-keeper/internal/utils"
	"github.com/MKhiriev/go-pass-keeper/models"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const selectPrivateDataSQL = `SELECT id, user_id, type, metadata, data, notes, additional_fields, created_at, updated_at, version, client_side_id, hash, deleted FROM ciphers`

func newTestDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return db, mock
}

// NewDBFromSQL создаёт DB из существующего *sql.DB (для тестов).
func newDBFromSQL(db *sql.DB) *DB {
	return &DB{
		DB:                 db,
		errorClassificator: NewPostgresErrorClassifier(),
		logger:             logger.Nop(),
	}
}

func newTestRepo(t *testing.T, db *sql.DB) PrivateDataRepository {
	t.Helper()
	storeDB := newDBFromSQL(db)
	log := logger.Nop()
	return NewPrivateDataRepository(storeDB, log)
}

func testContext() context.Context {
	l := zerolog.Nop()
	return l.WithContext(context.Background())
}

var privateDataColumns = []string{
	"id", "user_id", "type", "metadata", "data", "notes",
	"additional_fields", "created_at", "updated_at", "version",
	"client_side_id", "hash", "deleted",
}

type privateDataRow struct {
	id               int64
	userID           int64
	dataType         models.DataType
	metadata         models.CipheredMetadata
	data             models.CipheredData
	notes            driver.Value // *models.CipheredNotes или nil
	additionalFields driver.Value // *models.CipheredCustomFields или nil
	createdAt        *time.Time
	updatedAt        *time.Time
	version          int64
	clientSideID     string
	hash             string
	deleted          bool
}

func (r privateDataRow) toArgs() []driver.Value {
	return []driver.Value{
		r.id, r.userID, r.dataType,
		r.metadata, r.data,
		r.notes, r.additionalFields,
		r.createdAt, r.updatedAt,
		r.version, r.clientSideID, r.hash, r.deleted,
	}
}

func TestGetPrivateData(t *testing.T) {
	now := time.Now().Truncate(time.Millisecond)
	notes := models.CipheredNotes("enc_notes")
	fields := models.CipheredCustomFields("enc_fields")

	type mockSetup struct {
		query    string
		args     []driver.Value
		rows     []privateDataRow
		queryErr error
		rowErr   error
		badCols  []string
	}

	type want struct {
		err       string
		resultLen int
		items     []models.PrivateData
	}

	tests := []struct {
		name string
		req  models.DownloadRequest
		mock mockSetup
		want want
	}{
		{
			name: "success: no client_side_ids filter",
			req:  models.DownloadRequest{UserID: 42},
			mock: mockSetup{
				query: selectPrivateDataSQL + ` WHERE user_id = $1`,
				args:  []driver.Value{int64(42)},
				rows: []privateDataRow{
					{
						id: 1, userID: 42,
						dataType: models.LoginPassword, metadata: "enc_meta", data: "enc_data",
						notes: nil, additionalFields: nil,
						createdAt: &now, updatedAt: &now,
						version: 1, clientSideID: "cid-1", hash: "hash1", deleted: false,
					},
					{
						id: 2, userID: 42,
						dataType: models.Text, metadata: "enc_meta2", data: "enc_data2",
						notes: notes, additionalFields: fields,
						createdAt: &now, updatedAt: &now,
						version: 2, clientSideID: "cid-2", hash: "hash2", deleted: false,
					},
				},
			},
			want: want{
				resultLen: 2,
				items: []models.PrivateData{
					{
						ID: 1, UserID: 42,
						Payload:   models.PrivateDataPayload{Type: models.LoginPassword, Metadata: "enc_meta", Data: "enc_data"},
						CreatedAt: &now, UpdatedAt: &now,
						Version: 1, ClientSideID: "cid-1", Hash: "hash1", Deleted: false,
					},
					{
						ID: 2, UserID: 42,
						Payload:   models.PrivateDataPayload{Type: models.Text, Metadata: "enc_meta2", Data: "enc_data2", Notes: &notes, AdditionalFields: &fields},
						CreatedAt: &now, UpdatedAt: &now,
						Version: 2, ClientSideID: "cid-2", Hash: "hash2", Deleted: false,
					},
				},
			},
		},
		{
			name: "success: with client_side_ids filter",
			req:  models.DownloadRequest{UserID: 42, ClientSideIDs: []string{"cid-1", "cid-3"}, Length: 2},
			mock: mockSetup{
				query: selectPrivateDataSQL + ` WHERE user_id = $1 AND client_side_id IN ($2,$3)`,
				args:  []driver.Value{int64(42), "cid-1", "cid-3"},
				rows: []privateDataRow{
					{
						id: 1, userID: 42,
						dataType: models.LoginPassword, metadata: "enc_meta", data: "enc_data",
						createdAt: &now, updatedAt: &now,
						version: 1, clientSideID: "cid-1", hash: "hash1",
					},
					{
						id: 3, userID: 42,
						dataType: models.BankCard, metadata: "enc_meta3", data: "enc_data3",
						createdAt: &now, updatedAt: nil, // NULL
						version: 1, clientSideID: "cid-3", hash: "hash3",
					},
				},
			},
			want: want{
				resultLen: 2,
				items: []models.PrivateData{
					{
						ID: 1, UserID: 42,
						Payload:   models.PrivateDataPayload{Type: models.LoginPassword, Metadata: "enc_meta", Data: "enc_data"},
						CreatedAt: &now, UpdatedAt: &now,
						Version: 1, ClientSideID: "cid-1", Hash: "hash1",
					},
					{
						ID: 3, UserID: 42,
						Payload:   models.PrivateDataPayload{Type: models.BankCard, Metadata: "enc_meta3", Data: "enc_data3"},
						CreatedAt: &now, UpdatedAt: nil,
						Version: 1, ClientSideID: "cid-3", Hash: "hash3",
					},
				},
			},
		},
		{
			name: "success: deleted record",
			req:  models.DownloadRequest{UserID: 42, ClientSideIDs: []string{"cid-del"}, Length: 1},
			mock: mockSetup{
				query: selectPrivateDataSQL + ` WHERE user_id = $1 AND client_side_id IN ($2)`,
				args:  []driver.Value{int64(42), "cid-del"},
				rows: []privateDataRow{
					{
						id: 5, userID: 42,
						dataType: models.Binary, metadata: "enc_meta5", data: "enc_data5",
						createdAt: &now, updatedAt: &now,
						version: 7, clientSideID: "cid-del", hash: "hash5", deleted: true,
					},
				},
			},
			want: want{
				resultLen: 1,
				items: []models.PrivateData{
					{
						ID: 5, UserID: 42,
						Payload:   models.PrivateDataPayload{Type: models.Binary, Metadata: "enc_meta5", Data: "enc_data5"},
						CreatedAt: &now, UpdatedAt: &now,
						Version: 7, ClientSideID: "cid-del", Hash: "hash5", Deleted: true,
					},
				},
			},
		},
		{
			name: "success: empty result",
			req:  models.DownloadRequest{UserID: 99},
			mock: mockSetup{
				query: selectPrivateDataSQL + ` WHERE user_id = $1`,
				args:  []driver.Value{int64(99)},
				rows:  []privateDataRow{},
			},
			want: want{resultLen: 0},
		},
		{
			name: "error: query execution fails",
			req:  models.DownloadRequest{UserID: 42},
			mock: mockSetup{
				query:    selectPrivateDataSQL + ` WHERE user_id = $1`,
				args:     []driver.Value{int64(42)},
				queryErr: errors.New("connection refused"),
			},
			want: want{err: "error executing sql query: connection refused"},
		},
		{
			name: "error: scan fails (wrong column count)",
			req:  models.DownloadRequest{UserID: 42},
			mock: mockSetup{
				query:   selectPrivateDataSQL + ` WHERE user_id = $1`,
				args:    []driver.Value{int64(42)},
				badCols: []string{"id", "user_id"},
				rows:    []privateDataRow{{id: 1, userID: 42}},
			},
			want: want{err: "failed to scan private data row"},
		},
		{
			name: "error: rows iteration error",
			req:  models.DownloadRequest{UserID: 42},
			mock: mockSetup{
				query: selectPrivateDataSQL + ` WHERE user_id = $1`,
				args:  []driver.Value{int64(42)},
				rows: []privateDataRow{
					{
						id: 1, userID: 42,
						dataType: models.LoginPassword, metadata: "enc_meta", data: "enc_data",
						createdAt: &now, updatedAt: &now,
						version: 1, clientSideID: "cid-1", hash: "hash1",
					},
				},
				rowErr: errors.New("network interruption"),
			},
			want: want{err: "failed to scan private data rows: network interruption"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db, mock := newTestDB(t)
			repo := newTestRepo(t, db)
			ctx := testContext()

			expectation := mock.ExpectQuery(regexp.QuoteMeta(tc.mock.query)).
				WithArgs(tc.mock.args...)

			if tc.mock.queryErr != nil {
				expectation.WillReturnError(tc.mock.queryErr)
			} else {
				cols := privateDataColumns
				if len(tc.mock.badCols) > 0 {
					cols = tc.mock.badCols
				}

				mockRows := sqlmock.NewRows(cols)
				for i, r := range tc.mock.rows {
					if len(tc.mock.badCols) > 0 {
						mockRows.AddRow(driver.Value(r.id), driver.Value(r.userID))
					} else {
						mockRows.AddRow(r.toArgs()...)
					}
					if tc.mock.rowErr != nil {
						mockRows.RowError(i, tc.mock.rowErr)
					}
				}
				expectation.WillReturnRows(mockRows)
			}

			result, err := repo.GetPrivateData(ctx, tc.req)

			if tc.want.err != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.want.err)
				assert.Nil(t, result)
				require.NoError(t, mock.ExpectationsWereMet())
				return
			}

			require.NoError(t, err)
			require.Len(t, result, tc.want.resultLen)

			for i, expected := range tc.want.items {
				got := result[i]

				assert.Equal(t, expected.ID, got.ID, "ID[%d]", i)
				assert.Equal(t, expected.UserID, got.UserID, "UserID[%d]", i)
				assert.Equal(t, expected.ClientSideID, got.ClientSideID, "ClientSideID[%d]", i)
				assert.Equal(t, expected.Version, got.Version, "Version[%d]", i)
				assert.Equal(t, expected.Hash, got.Hash, "Hash[%d]", i)
				assert.Equal(t, expected.Deleted, got.Deleted, "Deleted[%d]", i)

				assert.Equal(t, expected.Payload.Type, got.Payload.Type, "Payload.Type[%d]", i)
				assert.Equal(t, expected.Payload.Metadata, got.Payload.Metadata, "Payload.Metadata[%d]", i)
				assert.Equal(t, expected.Payload.Data, got.Payload.Data, "Payload.Data[%d]", i)
				assert.Equal(t, expected.Payload.Notes, got.Payload.Notes, "Payload.Notes[%d]", i)
				assert.Equal(t, expected.Payload.AdditionalFields, got.Payload.AdditionalFields, "Payload.AdditionalFields[%d]", i)

				if expected.CreatedAt == nil {
					assert.Nil(t, got.CreatedAt, "CreatedAt[%d] should be nil", i)
				} else {
					require.NotNil(t, got.CreatedAt, "CreatedAt[%d] should not be nil", i)
					assert.Equal(t, expected.CreatedAt.UTC(), got.CreatedAt.UTC(), "CreatedAt[%d]", i)
				}

				if expected.UpdatedAt == nil {
					assert.Nil(t, got.UpdatedAt, "UpdatedAt[%d] should be nil", i)
				} else {
					require.NotNil(t, got.UpdatedAt, "UpdatedAt[%d] should not be nil", i)
					assert.Equal(t, expected.UpdatedAt.UTC(), got.UpdatedAt.UTC(), "UpdatedAt[%d]", i)
				}
			}

			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestGetAllPrivateData(t *testing.T) {
	now := time.Now().Truncate(time.Millisecond)
	notes := models.CipheredNotes("enc_notes")
	fields := models.CipheredCustomFields("enc_fields")

	const query = `SELECT id, user_id, type, metadata, data, notes, additional_fields, created_at, updated_at, version, client_side_id, hash, deleted FROM ciphers WHERE user_id = $1;`

	type mockSetup struct {
		rows     []privateDataRow
		queryErr error
		rowErr   error
		badCols  []string
	}

	type want struct {
		err       string
		resultLen int
		items     []models.PrivateData
	}

	tests := []struct {
		name   string
		userID int64
		mock   mockSetup
		want   want
	}{
		{
			name:   "success: multiple records",
			userID: 42,
			mock: mockSetup{
				rows: []privateDataRow{
					{
						id: 1, userID: 42,
						dataType: models.LoginPassword, metadata: "enc_meta", data: "enc_data",
						notes: nil, additionalFields: nil,
						createdAt: &now, updatedAt: &now,
						version: 1, clientSideID: "cid-1", hash: "hash1", deleted: false,
					},
					{
						id: 2, userID: 42,
						dataType: models.Text, metadata: "enc_meta2", data: "enc_data2",
						notes: driver.Value(notes), additionalFields: driver.Value(fields),
						createdAt: &now, updatedAt: &now,
						version: 3, clientSideID: "cid-2", hash: "hash2", deleted: false,
					},
				},
			},
			want: want{
				resultLen: 2,
				items: []models.PrivateData{
					{
						ID: 1, UserID: 42,
						Payload:   models.PrivateDataPayload{Type: models.LoginPassword, Metadata: "enc_meta", Data: "enc_data"},
						CreatedAt: &now, UpdatedAt: &now,
						Version: 1, ClientSideID: "cid-1", Hash: "hash1", Deleted: false,
					},
					{
						ID: 2, UserID: 42,
						Payload:   models.PrivateDataPayload{Type: models.Text, Metadata: "enc_meta2", Data: "enc_data2", Notes: &notes, AdditionalFields: &fields},
						CreatedAt: &now, UpdatedAt: &now,
						Version: 3, ClientSideID: "cid-2", Hash: "hash2", Deleted: false,
					},
				},
			},
		},
		{
			name:   "success: deleted record included",
			userID: 42,
			mock: mockSetup{
				rows: []privateDataRow{
					{
						id: 5, userID: 42,
						dataType: models.Binary, metadata: "enc_meta5", data: "enc_data5",
						createdAt: &now, updatedAt: &now,
						version: 7, clientSideID: "cid-del", hash: "hash5", deleted: true,
					},
				},
			},
			want: want{
				resultLen: 1,
				items: []models.PrivateData{
					{
						ID: 5, UserID: 42,
						Payload:   models.PrivateDataPayload{Type: models.Binary, Metadata: "enc_meta5", Data: "enc_data5"},
						CreatedAt: &now, UpdatedAt: &now,
						Version: 7, ClientSideID: "cid-del", Hash: "hash5", Deleted: true,
					},
				},
			},
		},
		{
			name:   "success: nullable timestamps (updatedAt = NULL)",
			userID: 42,
			mock: mockSetup{
				rows: []privateDataRow{
					{
						id: 3, userID: 42,
						dataType: models.BankCard, metadata: "enc_meta3", data: "enc_data3",
						createdAt: &now, updatedAt: nil, // updated_at = NULL
						version: 1, clientSideID: "cid-3", hash: "hash3",
					},
				},
			},
			want: want{
				resultLen: 1,
				items: []models.PrivateData{
					{
						ID: 3, UserID: 42,
						Payload:   models.PrivateDataPayload{Type: models.BankCard, Metadata: "enc_meta3", Data: "enc_data3"},
						CreatedAt: &now, UpdatedAt: nil,
						Version: 1, ClientSideID: "cid-3", Hash: "hash3",
					},
				},
			},
		},
		{
			name:   "success: nullable timestamps (createdAt = NULL)",
			userID: 42,
			mock: mockSetup{
				rows: []privateDataRow{
					{
						id: 4, userID: 42,
						dataType: models.Text, metadata: "enc_meta4", data: "enc_data4",
						createdAt: nil, updatedAt: nil, // оба NULL
						version: 2, clientSideID: "cid-4", hash: "hash4",
					},
				},
			},
			want: want{
				resultLen: 1,
				items: []models.PrivateData{
					{
						ID: 4, UserID: 42,
						Payload:   models.PrivateDataPayload{Type: models.Text, Metadata: "enc_meta4", Data: "enc_data4"},
						CreatedAt: nil, UpdatedAt: nil,
						Version: 2, ClientSideID: "cid-4", Hash: "hash4",
					},
				},
			},
		},
		{
			name:   "success: empty result",
			userID: 99,
			mock:   mockSetup{rows: []privateDataRow{}},
			want:   want{resultLen: 0},
		},
		{
			name:   "error: query execution fails",
			userID: 42,
			mock: mockSetup{
				queryErr: errors.New("connection refused"),
			},
			want: want{err: "error executing sql query: connection refused"},
		},
		{
			name:   "error: scan fails (wrong column count)",
			userID: 42,
			mock: mockSetup{
				badCols: []string{"id", "user_id"},
				rows:    []privateDataRow{{id: 1, userID: 42}},
			},
			want: want{err: "failed to scan private data row"},
		},
		{
			name:   "error: rows iteration error",
			userID: 42,
			mock: mockSetup{
				rows: []privateDataRow{
					{
						id: 1, userID: 42,
						dataType: models.LoginPassword, metadata: "enc_meta", data: "enc_data",
						createdAt: &now, updatedAt: &now,
						version: 1, clientSideID: "cid-1", hash: "hash1",
					},
				},
				rowErr: errors.New("network interruption"),
			},
			want: want{err: "failed to scan private data rows"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db, mock := newTestDB(t)
			repo := newTestRepo(t, db)
			ctx := testContext()

			expectation := mock.ExpectQuery(regexp.QuoteMeta(query)).
				WithArgs(driver.Value(tc.userID))

			if tc.mock.queryErr != nil {
				expectation.WillReturnError(tc.mock.queryErr)
			} else {
				cols := privateDataColumns
				if len(tc.mock.badCols) > 0 {
					cols = tc.mock.badCols
				}

				mockRows := sqlmock.NewRows(cols)
				for i, r := range tc.mock.rows {
					if len(tc.mock.badCols) > 0 {
						mockRows.AddRow(driver.Value(r.id), driver.Value(r.userID))
					} else {
						mockRows.AddRow(r.toArgs()...)
					}
					if tc.mock.rowErr != nil {
						mockRows.RowError(i, tc.mock.rowErr)
					}
				}
				expectation.WillReturnRows(mockRows)
			}

			result, err := repo.GetAllPrivateData(ctx, tc.userID)

			if tc.want.err != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.want.err)
				assert.Nil(t, result)
				require.NoError(t, mock.ExpectationsWereMet())
				return
			}

			require.NoError(t, err)
			require.Len(t, result, tc.want.resultLen)

			for i, expected := range tc.want.items {
				got := result[i]

				assert.Equal(t, expected.ID, got.ID, "ID[%d]", i)
				assert.Equal(t, expected.UserID, got.UserID, "UserID[%d]", i)
				assert.Equal(t, expected.ClientSideID, got.ClientSideID, "ClientSideID[%d]", i)
				assert.Equal(t, expected.Version, got.Version, "Version[%d]", i)
				assert.Equal(t, expected.Hash, got.Hash, "Hash[%d]", i)
				assert.Equal(t, expected.Deleted, got.Deleted, "Deleted[%d]", i)

				assert.Equal(t, expected.Payload.Type, got.Payload.Type, "Payload.Type[%d]", i)
				assert.Equal(t, expected.Payload.Metadata, got.Payload.Metadata, "Payload.Metadata[%d]", i)
				assert.Equal(t, expected.Payload.Data, got.Payload.Data, "Payload.Data[%d]", i)
				assert.Equal(t, expected.Payload.Notes, got.Payload.Notes, "Payload.Notes[%d]", i)
				assert.Equal(t, expected.Payload.AdditionalFields, got.Payload.AdditionalFields, "Payload.AdditionalFields[%d]", i)

				// CreatedAt
				if expected.CreatedAt == nil {
					assert.Nil(t, got.CreatedAt, "CreatedAt[%d] should be nil", i)
				} else {
					require.NotNil(t, got.CreatedAt, "CreatedAt[%d] should not be nil", i)
					assert.Equal(t, expected.CreatedAt.UTC(), got.CreatedAt.UTC(), "CreatedAt[%d]", i)
				}

				// UpdatedAt
				if expected.UpdatedAt == nil {
					assert.Nil(t, got.UpdatedAt, "UpdatedAt[%d] should be nil", i)
				} else {
					require.NotNil(t, got.UpdatedAt, "UpdatedAt[%d] should not be nil", i)
					assert.Equal(t, expected.UpdatedAt.UTC(), got.UpdatedAt.UTC(), "UpdatedAt[%d]", i)
				}
			}

			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestGetAllStates(t *testing.T) {
	now := time.Now().Truncate(time.Millisecond)

	const query = `SELECT client_side_id, hash, version, deleted, updated_at FROM ciphers WHERE user_id = $1;`

	var stateColumns = []string{"client_side_id", "hash", "version", "deleted", "updated_at"}

	type stateRow struct {
		clientSideID string
		hash         string
		version      int64
		deleted      bool
		updatedAt    *time.Time
	}

	toArgs := func(r stateRow) []driver.Value {
		return []driver.Value{r.clientSideID, r.hash, r.version, r.deleted, r.updatedAt}
	}

	type mockSetup struct {
		rows     []stateRow
		queryErr error
		rowErr   error
		badCols  []string
	}

	type want struct {
		err       string
		resultLen int
		items     []models.PrivateDataState
	}

	tests := []struct {
		name   string
		userID int64
		mock   mockSetup
		want   want
	}{
		{
			name:   "success: multiple records",
			userID: 42,
			mock: mockSetup{
				rows: []stateRow{
					{clientSideID: "cid-1", hash: "hash1", version: 1, deleted: false, updatedAt: &now},
					{clientSideID: "cid-2", hash: "hash2", version: 3, deleted: false, updatedAt: &now},
					{clientSideID: "cid-3", hash: "hash3", version: 7, deleted: true, updatedAt: &now},
				},
			},
			want: want{
				resultLen: 3,
				items: []models.PrivateDataState{
					{ClientSideID: "cid-1", Hash: "hash1", Version: 1, Deleted: false, UpdatedAt: &now},
					{ClientSideID: "cid-2", Hash: "hash2", Version: 3, Deleted: false, UpdatedAt: &now},
					{ClientSideID: "cid-3", Hash: "hash3", Version: 7, Deleted: true, UpdatedAt: &now},
				},
			},
		},
		{
			name:   "success: deleted record included",
			userID: 42,
			mock: mockSetup{
				rows: []stateRow{
					{clientSideID: "cid-del", hash: "hash-del", version: 5, deleted: true, updatedAt: &now},
				},
			},
			want: want{
				resultLen: 1,
				items: []models.PrivateDataState{
					{ClientSideID: "cid-del", Hash: "hash-del", Version: 5, Deleted: true, UpdatedAt: &now},
				},
			},
		},
		{
			name:   "success: updatedAt = NULL",
			userID: 42,
			mock: mockSetup{
				rows: []stateRow{
					{clientSideID: "cid-null", hash: "hash-null", version: 2, deleted: false, updatedAt: nil},
				},
			},
			want: want{
				resultLen: 1,
				items: []models.PrivateDataState{
					{ClientSideID: "cid-null", Hash: "hash-null", Version: 2, Deleted: false, UpdatedAt: nil},
				},
			},
		},
		{
			name:   "success: empty result",
			userID: 99,
			mock:   mockSetup{rows: []stateRow{}},
			want:   want{resultLen: 0},
		},
		{
			name:   "error: query execution fails",
			userID: 42,
			mock: mockSetup{
				queryErr: errors.New("connection refused"),
			},
			want: want{err: "error executing sql query"},
		},
		{
			name:   "error: scan fails (wrong column count)",
			userID: 42,
			mock: mockSetup{
				badCols: []string{"client_side_id"},
				rows:    []stateRow{{clientSideID: "cid-1"}},
			},
			want: want{err: "failed to scan private data row"},
		},
		{
			name:   "error: rows iteration error",
			userID: 42,
			mock: mockSetup{
				rows: []stateRow{
					{clientSideID: "cid-1", hash: "hash1", version: 1, deleted: false, updatedAt: &now},
				},
				rowErr: errors.New("network interruption"),
			},
			want: want{err: "failed to scan private data rows"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db, mock := newTestDB(t)
			repo := newTestRepo(t, db)
			ctx := testContext()

			expectation := mock.ExpectQuery(regexp.QuoteMeta(query)).
				WithArgs(driver.Value(tc.userID))

			if tc.mock.queryErr != nil {
				expectation.WillReturnError(tc.mock.queryErr)
			} else {
				cols := stateColumns
				if len(tc.mock.badCols) > 0 {
					cols = tc.mock.badCols
				}

				mockRows := sqlmock.NewRows(cols)
				for i, r := range tc.mock.rows {
					if len(tc.mock.badCols) > 0 {
						mockRows.AddRow(driver.Value(r.clientSideID))
					} else {
						mockRows.AddRow(toArgs(r)...)
					}
					if tc.mock.rowErr != nil {
						mockRows.RowError(i, tc.mock.rowErr)
					}
				}
				expectation.WillReturnRows(mockRows)
			}

			result, err := repo.GetAllStates(ctx, tc.userID)

			if tc.want.err != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.want.err)
				assert.Nil(t, result)
				require.NoError(t, mock.ExpectationsWereMet())
				return
			}

			require.NoError(t, err)
			require.Len(t, result, tc.want.resultLen)

			for i, expected := range tc.want.items {
				got := result[i]

				assert.Equal(t, expected.ClientSideID, got.ClientSideID, "ClientSideID[%d]", i)
				assert.Equal(t, expected.Hash, got.Hash, "Hash[%d]", i)
				assert.Equal(t, expected.Version, got.Version, "Version[%d]", i)
				assert.Equal(t, expected.Deleted, got.Deleted, "Deleted[%d]", i)

				if expected.UpdatedAt == nil {
					assert.Nil(t, got.UpdatedAt, "UpdatedAt[%d] should be nil", i)
				} else {
					require.NotNil(t, got.UpdatedAt, "UpdatedAt[%d] should not be nil", i)
					assert.Equal(t, expected.UpdatedAt.UTC(), got.UpdatedAt.UTC(), "UpdatedAt[%d]", i)
				}
			}

			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestGetStates(t *testing.T) {
	now := time.Now().Truncate(time.Millisecond)

	const baseQuery = `SELECT client_side_id, hash, version, deleted, updated_at FROM ciphers WHERE user_id = $1`
	const queryWithIDs2 = `SELECT client_side_id, hash, version, deleted, updated_at FROM ciphers WHERE user_id = $1 AND client_side_id IN ($2,$3)`
	const queryWithIDs1 = `SELECT client_side_id, hash, version, deleted, updated_at FROM ciphers WHERE user_id = $1 AND client_side_id IN ($2)`

	var stateColumns = []string{"client_side_id", "hash", "version", "deleted", "updated_at"}

	type stateRow struct {
		clientSideID string
		hash         string
		version      int64
		deleted      bool
		updatedAt    *time.Time
	}

	toArgs := func(r stateRow) []driver.Value {
		return []driver.Value{r.clientSideID, r.hash, r.version, r.deleted, r.updatedAt}
	}

	type mockSetup struct {
		query    string
		args     []driver.Value
		rows     []stateRow
		queryErr error
		rowErr   error
		badCols  []string
	}

	type want struct {
		err       string
		resultLen int
		items     []models.PrivateDataState
	}

	tests := []struct {
		name string
		req  models.SyncRequest
		mock mockSetup
		want want
	}{
		{
			name: "success: no client_side_ids filter",
			req:  models.SyncRequest{UserID: 42},
			mock: mockSetup{
				query: baseQuery,
				args:  []driver.Value{int64(42)},
				rows: []stateRow{
					{clientSideID: "cid-1", hash: "hash1", version: 1, deleted: false, updatedAt: &now},
					{clientSideID: "cid-2", hash: "hash2", version: 3, deleted: false, updatedAt: &now},
				},
			},
			want: want{
				resultLen: 2,
				items: []models.PrivateDataState{
					{ClientSideID: "cid-1", Hash: "hash1", Version: 1, Deleted: false, UpdatedAt: &now},
					{ClientSideID: "cid-2", Hash: "hash2", Version: 3, Deleted: false, UpdatedAt: &now},
				},
			},
		},
		{
			name: "success: with client_side_ids filter (2 ids)",
			req: models.SyncRequest{
				UserID:        42,
				ClientSideIDs: []string{"cid-1", "cid-3"},
				Length:        2,
			},
			mock: mockSetup{
				query: queryWithIDs2,
				args:  []driver.Value{int64(42), "cid-1", "cid-3"},
				rows: []stateRow{
					{clientSideID: "cid-1", hash: "hash1", version: 1, deleted: false, updatedAt: &now},
					{clientSideID: "cid-3", hash: "hash3", version: 5, deleted: false, updatedAt: &now},
				},
			},
			want: want{
				resultLen: 2,
				items: []models.PrivateDataState{
					{ClientSideID: "cid-1", Hash: "hash1", Version: 1, Deleted: false, UpdatedAt: &now},
					{ClientSideID: "cid-3", Hash: "hash3", Version: 5, Deleted: false, UpdatedAt: &now},
				},
			},
		},
		{
			name: "success: with client_side_ids filter (1 id)",
			req: models.SyncRequest{
				UserID:        42,
				ClientSideIDs: []string{"cid-1"},
				Length:        1,
			},
			mock: mockSetup{
				query: queryWithIDs1,
				args:  []driver.Value{int64(42), "cid-1"},
				rows: []stateRow{
					{clientSideID: "cid-1", hash: "hash1", version: 2, deleted: false, updatedAt: &now},
				},
			},
			want: want{
				resultLen: 1,
				items: []models.PrivateDataState{
					{ClientSideID: "cid-1", Hash: "hash1", Version: 2, Deleted: false, UpdatedAt: &now},
				},
			},
		},
		{
			name: "success: deleted record included",
			req:  models.SyncRequest{UserID: 42},
			mock: mockSetup{
				query: baseQuery,
				args:  []driver.Value{int64(42)},
				rows: []stateRow{
					{clientSideID: "cid-del", hash: "hash-del", version: 7, deleted: true, updatedAt: &now},
				},
			},
			want: want{
				resultLen: 1,
				items: []models.PrivateDataState{
					{ClientSideID: "cid-del", Hash: "hash-del", Version: 7, Deleted: true, UpdatedAt: &now},
				},
			},
		},
		{
			name: "success: updatedAt = NULL",
			req:  models.SyncRequest{UserID: 42},
			mock: mockSetup{
				query: baseQuery,
				args:  []driver.Value{int64(42)},
				rows: []stateRow{
					{clientSideID: "cid-null", hash: "hash-null", version: 1, deleted: false, updatedAt: nil},
				},
			},
			want: want{
				resultLen: 1,
				items: []models.PrivateDataState{
					{ClientSideID: "cid-null", Hash: "hash-null", Version: 1, Deleted: false, UpdatedAt: nil},
				},
			},
		},
		{
			name: "success: empty result",
			req:  models.SyncRequest{UserID: 99},
			mock: mockSetup{
				query: baseQuery,
				args:  []driver.Value{int64(99)},
				rows:  []stateRow{},
			},
			want: want{resultLen: 0},
		},
		{
			name: "error: query execution fails",
			req:  models.SyncRequest{UserID: 42},
			mock: mockSetup{
				query:    baseQuery,
				args:     []driver.Value{int64(42)},
				queryErr: errors.New("connection refused"),
			},
			want: want{err: "error executing sql query"},
		},
		{
			name: "error: scan fails (wrong column count)",
			req:  models.SyncRequest{UserID: 42},
			mock: mockSetup{
				query:   baseQuery,
				args:    []driver.Value{int64(42)},
				badCols: []string{"client_side_id"},
				rows:    []stateRow{{clientSideID: "cid-1"}},
			},
			want: want{err: "failed to scan private data row"},
		},
		{
			name: "error: rows iteration error",
			req:  models.SyncRequest{UserID: 42},
			mock: mockSetup{
				query: baseQuery,
				args:  []driver.Value{int64(42)},
				rows: []stateRow{
					{clientSideID: "cid-1", hash: "hash1", version: 1, deleted: false, updatedAt: &now},
				},
				rowErr: errors.New("network interruption"),
			},
			want: want{err: "failed to scan private data rows"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db, mock := newTestDB(t)
			repo := newTestRepo(t, db)
			ctx := testContext()

			expectation := mock.ExpectQuery(regexp.QuoteMeta(tc.mock.query)).
				WithArgs(tc.mock.args...)

			if tc.mock.queryErr != nil {
				expectation.WillReturnError(tc.mock.queryErr)
			} else {
				cols := stateColumns
				if len(tc.mock.badCols) > 0 {
					cols = tc.mock.badCols
				}

				mockRows := sqlmock.NewRows(cols)
				for i, r := range tc.mock.rows {
					if len(tc.mock.badCols) > 0 {
						mockRows.AddRow(driver.Value(r.clientSideID))
					} else {
						mockRows.AddRow(toArgs(r)...)
					}
					if tc.mock.rowErr != nil {
						mockRows.RowError(i, tc.mock.rowErr)
					}
				}
				expectation.WillReturnRows(mockRows)
			}

			result, err := repo.GetStates(ctx, tc.req)

			if tc.want.err != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.want.err)
				assert.Nil(t, result)
				require.NoError(t, mock.ExpectationsWereMet())
				return
			}

			require.NoError(t, err)
			require.Len(t, result, tc.want.resultLen)

			for i, expected := range tc.want.items {
				got := result[i]

				assert.Equal(t, expected.ClientSideID, got.ClientSideID, "ClientSideID[%d]", i)
				assert.Equal(t, expected.Hash, got.Hash, "Hash[%d]", i)
				assert.Equal(t, expected.Version, got.Version, "Version[%d]", i)
				assert.Equal(t, expected.Deleted, got.Deleted, "Deleted[%d]", i)

				if expected.UpdatedAt == nil {
					assert.Nil(t, got.UpdatedAt, "UpdatedAt[%d] should be nil", i)
				} else {
					require.NotNil(t, got.UpdatedAt, "UpdatedAt[%d] should not be nil", i)
					assert.Equal(t, expected.UpdatedAt.UTC(), got.UpdatedAt.UTC(), "UpdatedAt[%d]", i)
				}
			}

			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestUpdateSingleRecord(t *testing.T) {
	ptrMeta := func(s string) *models.CipheredMetadata { v := models.CipheredMetadata(s); return &v }
	ptrData := func(s string) *models.CipheredData { v := models.CipheredData(s); return &v }
	ptrNotes := func(s string) *models.CipheredNotes { v := models.CipheredNotes(s); return &v }
	ptrFields := func(s string) *models.CipheredCustomFields { v := models.CipheredCustomFields(s); return &v }

	const userID = int64(42)

	testContextWithUser := func(uid int64) context.Context {
		l := zerolog.Nop()
		ctx := l.WithContext(context.Background())
		return context.WithValue(ctx, utils.UserIDCtxKey, uid)
	}

	buildExpectedQuery := func(setClauses, versionPlaceholder string) string {
		return fmt.Sprintf(`
       WITH target_record AS (
          SELECT id, version
          FROM ciphers
          WHERE client_side_id = $1 AND user_id = $2
       ),
       updated_record AS (
          UPDATE ciphers
          SET %s
          WHERE client_side_id = $1
            AND user_id = $2
            AND version = %s
          RETURNING id
       )
       SELECT
          (SELECT id FROM updated_record)      AS updated_id,
          (SELECT version FROM target_record)   AS current_db_version`,
			setClauses, versionPlaceholder)
	}

	fullQuery := buildExpectedQuery(
		"updated_at = NOW(), version = version + 1, metadata = $3, data = $4, notes = $5, additional_fields = $6, hash = $7",
		"$8",
	)
	metaOnlyQuery := buildExpectedQuery(
		"updated_at = NOW(), version = version + 1, metadata = $3, hash = $4",
		"$5",
	)
	hashOnlyQuery := buildExpectedQuery(
		"updated_at = NOW(), version = version + 1, hash = $3",
		"$4",
	)
	noPayloadQuery := buildExpectedQuery(
		"updated_at = NOW(), version = version + 1",
		"$3",
	)

	var cteColumns = []string{"updated_id", "current_db_version"}

	id1 := int64(1)
	ver5 := int64(5)

	type mockSetup struct {
		query            string
		args             []driver.Value
		updatedID        *int64
		currentDBVersion *int64
		queryErr         error
	}

	type want struct {
		err     error  // для errors.Is
		errWrap string // для Contains
		noErr   bool
	}

	tests := []struct {
		name   string
		update models.PrivateDataUpdate
		mock   mockSetup
		want   want
	}{
		{
			name: "success: all fields updated",
			update: models.PrivateDataUpdate{
				ClientSideID:      "cid-1",
				Version:           5,
				UpdatedRecordHash: "new-hash",
				FieldsUpdate: models.FieldsUpdate{
					Metadata:         ptrMeta("enc_meta"),
					Data:             ptrData("enc_data"),
					Notes:            ptrNotes("enc_notes"),
					AdditionalFields: ptrFields("enc_fields"),
				},
			},
			mock: mockSetup{
				query:            fullQuery,
				args:             []driver.Value{"cid-1", userID, "enc_meta", "enc_data", "enc_notes", "enc_fields", "new-hash", int64(5)},
				updatedID:        &id1,
				currentDBVersion: &ver5,
			},
			want: want{noErr: true},
		},
		{
			name: "success: only metadata updated",
			update: models.PrivateDataUpdate{
				ClientSideID:      "cid-1",
				Version:           5,
				UpdatedRecordHash: "new-hash",
				FieldsUpdate: models.FieldsUpdate{
					Metadata: ptrMeta("enc_meta"),
				},
			},
			mock: mockSetup{
				query:            metaOnlyQuery,
				args:             []driver.Value{"cid-1", userID, "enc_meta", "new-hash", int64(5)},
				updatedID:        &id1,
				currentDBVersion: &ver5,
			},
			want: want{noErr: true},
		},
		{
			name: "success: only hash updated (no payload fields)",
			update: models.PrivateDataUpdate{
				ClientSideID:      "cid-1",
				Version:           5,
				UpdatedRecordHash: "new-hash",
				FieldsUpdate:      models.FieldsUpdate{},
			},
			mock: mockSetup{
				query:            hashOnlyQuery,
				args:             []driver.Value{"cid-1", userID, "new-hash", int64(5)},
				updatedID:        &id1,
				currentDBVersion: &ver5,
			},
			want: want{noErr: true},
		},
		{
			name: "success: no payload fields and no hash (only updated_at + version bump)",
			update: models.PrivateDataUpdate{
				ClientSideID:      "cid-1",
				Version:           5,
				UpdatedRecordHash: "",
				FieldsUpdate:      models.FieldsUpdate{},
			},
			mock: mockSetup{
				query:            noPayloadQuery,
				args:             []driver.Value{"cid-1", userID, int64(5)},
				updatedID:        &id1,
				currentDBVersion: &ver5,
			},
			want: want{noErr: true},
		},
		{
			name: "error: record not found (both NULL)",
			update: models.PrivateDataUpdate{
				ClientSideID:      "cid-missing",
				Version:           1,
				UpdatedRecordHash: "hash",
			},
			mock: mockSetup{
				query:            hashOnlyQuery,
				args:             []driver.Value{"cid-missing", userID, "hash", int64(1)},
				updatedID:        nil,
				currentDBVersion: nil,
			},
			want: want{err: ErrPrivateDataNotFound},
		},
		{
			name: "error: version conflict (updatedID=NULL, currentDBVersion!=NULL)",
			update: models.PrivateDataUpdate{
				ClientSideID:      "cid-1",
				Version:           3,
				UpdatedRecordHash: "hash",
			},
			mock: mockSetup{
				query:            hashOnlyQuery,
				args:             []driver.Value{"cid-1", userID, "hash", int64(3)},
				updatedID:        nil,
				currentDBVersion: &ver5,
			},
			want: want{errWrap: ErrVersionConflict.Error()},
		},
		{
			name: "error: query execution fails",
			update: models.PrivateDataUpdate{
				ClientSideID:      "cid-1",
				Version:           5,
				UpdatedRecordHash: "hash",
			},
			mock: mockSetup{
				query:    hashOnlyQuery,
				args:     []driver.Value{"cid-1", userID, "hash", int64(5)},
				queryErr: errors.New("connection refused"),
			},
			want: want{errWrap: "error executing sql query"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db, mock := newTestDB(t)
			repo := newTestRepo(t, db).(*privateDataRepository)
			ctx := testContextWithUser(userID)

			if tc.mock.queryErr != nil {
				mock.ExpectQuery(regexp.QuoteMeta(tc.mock.query)).
					WithArgs(tc.mock.args...).
					WillReturnError(tc.mock.queryErr)
			} else {
				rows := sqlmock.NewRows(cteColumns).
					AddRow(
						driver.Value(tc.mock.updatedID),
						driver.Value(tc.mock.currentDBVersion),
					)
				mock.ExpectQuery(regexp.QuoteMeta(tc.mock.query)).
					WithArgs(tc.mock.args...).
					WillReturnRows(rows)
			}

			err := repo.updateSingleRecord(ctx, tc.update)

			switch {
			case tc.want.err != nil:
				require.ErrorIs(t, err, tc.want.err)
			case tc.want.errWrap != "":
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.want.errWrap)
			default:
				require.NoError(t, err)
			}

			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestUpdateSingleRecord_WithExistingRecord(t *testing.T) {
	ptrMeta := func(s string) *models.CipheredMetadata { v := models.CipheredMetadata(s); return &v }

	const userID = int64(42)

	testContextWithUser := func(uid int64) context.Context {
		l := zerolog.Nop()
		ctx := l.WithContext(context.Background())
		return context.WithValue(ctx, utils.UserIDCtxKey, uid)
	}

	// Запись, которая "уже есть в БД"
	// id=10, user_id=42, version=5, hash="old-hash", client_side_id="cid-existing"

	buildExpectedQuery := func(setClauses, versionPlaceholder string) string {
		return fmt.Sprintf(`
       WITH target_record AS (          SELECT id, version          FROM ciphers          WHERE client_side_id = $1 AND user_id = $2       ),       updated_record AS (          UPDATE ciphers          SET %s
          WHERE client_side_id = $1            AND user_id = $2            AND version = %s
          RETURNING id       )       SELECT          (SELECT id FROM updated_record)      AS updated_id,          (SELECT version FROM target_record)   AS current_db_version`,
			setClauses, versionPlaceholder)
	}

	var cteColumns = []string{"updated_id", "current_db_version"}

	t.Run("success: update hash on existing record, version matches", func(t *testing.T) {
		db, mock := newTestDB(t)
		repo := newTestRepo(t, db).(*privateDataRepository)
		ctx := testContextWithUser(userID)

		// Сначала убедимся что запись есть — мокаем SELECT для GetPrivateData
		selectRows := sqlmock.NewRows(privateDataColumns).
			AddRow(
				int64(10), int64(42), models.LoginPassword, "old_meta", "old_data", nil, nil,
				time.Now().UTC(), time.Now().UTC(), int64(5), "cid-existing", "old-hash", false,
			)
		mock.ExpectQuery(regexp.QuoteMeta(selectPrivateDataSQL+" WHERE user_id = $1 AND client_side_id IN ($2)")).
			WithArgs(int64(42), "cid-existing").
			WillReturnRows(selectRows)

		// Выполняем GetPrivateData — запись существует с version=5
		existing, err := repo.GetPrivateData(ctx, models.DownloadRequest{
			UserID:        userID,
			ClientSideIDs: []string{"cid-existing"},
			Length:        1,
		})
		require.NoError(t, err)
		require.Len(t, existing, 1)
		assert.Equal(t, int64(5), existing[0].Version)
		assert.Equal(t, "old-hash", existing[0].Hash)

		// Теперь обновляем — version=5 совпадает с БД, UPDATE пройдёт
		updatedID := int64(10)
		dbVersion := int64(5)

		query := buildExpectedQuery(
			"updated_at = NOW(), version = version + 1, metadata = $3, hash = $4",
			"$5",
		)

		updateRows := sqlmock.NewRows(cteColumns).
			AddRow(
				driver.Value(&updatedID), // updated_id != nil → UPDATE сработал
				driver.Value(&dbVersion), // current_db_version = 5
			)
		mock.ExpectQuery(regexp.QuoteMeta(query)).
			WithArgs("cid-existing", userID, "new_meta", "new-hash", int64(5)).
			WillReturnRows(updateRows)

		err = repo.updateSingleRecord(ctx, models.PrivateDataUpdate{
			ClientSideID:      "cid-existing",
			Version:           5, // совпадает с version в БД
			UpdatedRecordHash: "new-hash",
			FieldsUpdate: models.FieldsUpdate{
				Metadata: ptrMeta("new_meta"),
			},
		})
		require.NoError(t, err)

		// Мокаем повторный SELECT — версия должна быть 6, хэш обновлён
		selectAfterRows := sqlmock.NewRows(privateDataColumns).
			AddRow(
				int64(10), int64(42), models.LoginPassword, "new_meta", "old_data", nil, nil,
				time.Now().UTC(), time.Now().UTC(), int64(6), "cid-existing", "new-hash", false,
			)
		mock.ExpectQuery(regexp.QuoteMeta(selectPrivateDataSQL+" WHERE user_id = $1 AND client_side_id IN ($2)")).
			WithArgs(int64(42), "cid-existing").
			WillReturnRows(selectAfterRows)

		updated, err := repo.GetPrivateData(ctx, models.DownloadRequest{
			UserID:        userID,
			ClientSideIDs: []string{"cid-existing"},
			Length:        1,
		})
		require.NoError(t, err)
		require.Len(t, updated, 1)
		assert.Equal(t, int64(6), updated[0].Version, "version should be incremented")
		assert.Equal(t, "new-hash", updated[0].Hash, "hash should be updated")

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("error: version conflict on existing record", func(t *testing.T) {
		db, mock := newTestDB(t)
		repo := newTestRepo(t, db).(*privateDataRepository)
		ctx := testContextWithUser(userID)

		// Запись есть в БД с version=5
		selectRows := sqlmock.NewRows(privateDataColumns).
			AddRow(
				int64(10), int64(42), models.LoginPassword, "old_meta", "old_data", nil, nil,
				time.Now().UTC(), time.Now().UTC(), int64(5), "cid-existing", "old-hash", false,
			)
		mock.ExpectQuery(regexp.QuoteMeta(selectPrivateDataSQL+" WHERE user_id = $1 AND client_side_id IN ($2)")).
			WithArgs(int64(42), "cid-existing").
			WillReturnRows(selectRows)

		existing, err := repo.GetPrivateData(ctx, models.DownloadRequest{
			UserID:        userID,
			ClientSideIDs: []string{"cid-existing"},
			Length:        1,
		})
		require.NoError(t, err)
		require.Len(t, existing, 1)
		assert.Equal(t, int64(5), existing[0].Version)

		// Пытаемся обновить с version=3 — не совпадает с БД (5)
		dbVersion := int64(5)

		query := buildExpectedQuery(
			"updated_at = NOW(), version = version + 1, metadata = $3, hash = $4",
			"$5",
		)

		updateRows := sqlmock.NewRows(cteColumns).
			AddRow(
				driver.Value((*int64)(nil)), // updated_id = NULL → UPDATE не сработал
				driver.Value(&dbVersion),    // current_db_version = 5 → запись есть
			)
		mock.ExpectQuery(regexp.QuoteMeta(query)).
			WithArgs("cid-existing", userID, "new_meta", "new-hash", int64(3)).
			WillReturnRows(updateRows)

		err = repo.updateSingleRecord(ctx, models.PrivateDataUpdate{
			ClientSideID:      "cid-existing",
			Version:           3, // НЕ совпадает с version=5 в БД
			UpdatedRecordHash: "new-hash",
			FieldsUpdate: models.FieldsUpdate{
				Metadata: ptrMeta("new_meta"),
			},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), ErrVersionConflict.Error())

		// Проверяем что запись не изменилась — version по-прежнему 5
		selectAfterRows := sqlmock.NewRows(privateDataColumns).
			AddRow(
				int64(10), int64(42), models.LoginPassword, "old_meta", "old_data", nil, nil,
				time.Now().UTC(), time.Now().UTC(), int64(5), "cid-existing", "old-hash", false,
			)
		mock.ExpectQuery(regexp.QuoteMeta(selectPrivateDataSQL+" WHERE user_id = $1 AND client_side_id IN ($2)")).
			WithArgs(int64(42), "cid-existing").
			WillReturnRows(selectAfterRows)

		unchanged, err := repo.GetPrivateData(ctx, models.DownloadRequest{
			UserID:        userID,
			ClientSideIDs: []string{"cid-existing"},
			Length:        1,
		})
		require.NoError(t, err)
		require.Len(t, unchanged, 1)
		assert.Equal(t, int64(5), unchanged[0].Version, "version should remain unchanged")
		assert.Equal(t, "old-hash", unchanged[0].Hash, "hash should remain unchanged")

		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestUpdateMultipleRecords(t *testing.T) {
	ptrMeta := func(s string) *models.CipheredMetadata { v := models.CipheredMetadata(s); return &v }
	ptrData := func(s string) *models.CipheredData { v := models.CipheredData(s); return &v }

	const userID = int64(42)

	testContextWithUser := func(uid int64) context.Context {
		l := zerolog.Nop()
		ctx := l.WithContext(context.Background())
		return context.WithValue(ctx, utils.UserIDCtxKey, uid)
	}

	buildExpectedQuery := func(setClauses, versionPlaceholder string) string {
		return fmt.Sprintf(`
       WITH target_record AS (          SELECT id, version          FROM ciphers          WHERE client_side_id = $1 AND user_id = $2       ),       updated_record AS (          UPDATE ciphers          SET %s
          WHERE client_side_id = $1            AND user_id = $2            AND version = %s
          RETURNING id       )       SELECT          (SELECT id FROM updated_record)      AS updated_id,          (SELECT version FROM target_record)   AS current_db_version`,
			setClauses, versionPlaceholder)
	}

	hashOnlyQuery := buildExpectedQuery(
		"updated_at = NOW(), version = version + 1, hash = $3",
		"$4",
	)
	metaHashQuery := buildExpectedQuery(
		"updated_at = NOW(), version = version + 1, metadata = $3, hash = $4",
		"$5",
	)
	metaDataHashQuery := buildExpectedQuery(
		"updated_at = NOW(), version = version + 1, metadata = $3, data = $4, hash = $5",
		"$6",
	)

	var cteColumns = []string{"updated_id", "current_db_version"}

	id1 := int64(1)
	id2 := int64(2)
	id3 := int64(3)
	ver5 := int64(5)
	ver3 := int64(3)
	ver7 := int64(7)

	t.Run("success: update two records in transaction", func(t *testing.T) {
		db, mock := newTestDB(t)
		repo := newTestRepo(t, db).(*privateDataRepository)
		ctx := testContextWithUser(userID)

		updates := []models.PrivateDataUpdate{
			{
				ClientSideID:      "cid-1",
				Version:           5,
				UpdatedRecordHash: "hash-1-new",
				FieldsUpdate: models.FieldsUpdate{
					Metadata: ptrMeta("meta-1-new"),
				},
			},
			{
				ClientSideID:      "cid-2",
				Version:           3,
				UpdatedRecordHash: "hash-2-new",
				FieldsUpdate: models.FieldsUpdate{
					Metadata: ptrMeta("meta-2-new"),
					Data:     ptrData("data-2-new"),
				},
			},
		}

		mock.ExpectBegin()

		rows1 := sqlmock.NewRows(cteColumns).
			AddRow(driver.Value(&id1), driver.Value(&ver5))
		mock.ExpectQuery(regexp.QuoteMeta(metaHashQuery)).
			WithArgs("cid-1", userID, "meta-1-new", "hash-1-new", int64(5)).
			WillReturnRows(rows1)

		rows2 := sqlmock.NewRows(cteColumns).
			AddRow(driver.Value(&id2), driver.Value(&ver3))
		mock.ExpectQuery(regexp.QuoteMeta(metaDataHashQuery)).
			WithArgs("cid-2", userID, "meta-2-new", "data-2-new", "hash-2-new", int64(3)).
			WillReturnRows(rows2)

		mock.ExpectCommit()

		err := repo.updateMultipleRecords(ctx, updates)
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success: update three records", func(t *testing.T) {
		db, mock := newTestDB(t)
		repo := newTestRepo(t, db).(*privateDataRepository)
		ctx := testContextWithUser(userID)

		updates := []models.PrivateDataUpdate{
			{
				ClientSideID:      "cid-1",
				Version:           5,
				UpdatedRecordHash: "h1",
			},
			{
				ClientSideID:      "cid-2",
				Version:           3,
				UpdatedRecordHash: "h2",
			},
			{
				ClientSideID:      "cid-3",
				Version:           7,
				UpdatedRecordHash: "h3",
			},
		}

		mock.ExpectBegin()

		r1 := sqlmock.NewRows(cteColumns).AddRow(driver.Value(&id1), driver.Value(&ver5))
		mock.ExpectQuery(regexp.QuoteMeta(hashOnlyQuery)).
			WithArgs("cid-1", userID, "h1", int64(5)).
			WillReturnRows(r1)

		r2 := sqlmock.NewRows(cteColumns).AddRow(driver.Value(&id2), driver.Value(&ver3))
		mock.ExpectQuery(regexp.QuoteMeta(hashOnlyQuery)).
			WithArgs("cid-2", userID, "h2", int64(3)).
			WillReturnRows(r2)

		r3 := sqlmock.NewRows(cteColumns).AddRow(driver.Value(&id3), driver.Value(&ver7))
		mock.ExpectQuery(regexp.QuoteMeta(hashOnlyQuery)).
			WithArgs("cid-3", userID, "h3", int64(7)).
			WillReturnRows(r3)

		mock.ExpectCommit()

		err := repo.updateMultipleRecords(ctx, updates)
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("error: begin transaction fails", func(t *testing.T) {
		db, mock := newTestDB(t)
		repo := newTestRepo(t, db).(*privateDataRepository)
		ctx := testContextWithUser(userID)

		mock.ExpectBegin().WillReturnError(errors.New("cannot begin"))

		err := repo.updateMultipleRecords(ctx, []models.PrivateDataUpdate{
			{ClientSideID: "cid-1", Version: 1, UpdatedRecordHash: "h"},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to begin transaction")
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("error: version conflict on second record rolls back", func(t *testing.T) {
		db, mock := newTestDB(t)
		repo := newTestRepo(t, db).(*privateDataRepository)
		ctx := testContextWithUser(userID)

		updates := []models.PrivateDataUpdate{
			{
				ClientSideID:      "cid-1",
				Version:           5,
				UpdatedRecordHash: "h1",
			},
			{
				ClientSideID:      "cid-2",
				Version:           1, // БД имеет version=3 → конфликт
				UpdatedRecordHash: "h2",
			},
		}

		mock.ExpectBegin()

		// Первый UPDATE — успех
		r1 := sqlmock.NewRows(cteColumns).AddRow(driver.Value(&id1), driver.Value(&ver5))
		mock.ExpectQuery(regexp.QuoteMeta(hashOnlyQuery)).
			WithArgs("cid-1", userID, "h1", int64(5)).
			WillReturnRows(r1)

		// Второй UPDATE — version mismatch: updatedID=NULL, currentDBVersion=3
		r2 := sqlmock.NewRows(cteColumns).AddRow(driver.Value((*int64)(nil)), driver.Value(&ver3))
		mock.ExpectQuery(regexp.QuoteMeta(hashOnlyQuery)).
			WithArgs("cid-2", userID, "h2", int64(1)).
			WillReturnRows(r2)

		mock.ExpectRollback()

		err := repo.updateMultipleRecords(ctx, updates)
		require.Error(t, err)
		assert.Contains(t, err.Error(), ErrVersionConflict.Error())
		assert.Contains(t, err.Error(), "index 1")
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("error: record not found on first record", func(t *testing.T) {
		db, mock := newTestDB(t)
		repo := newTestRepo(t, db).(*privateDataRepository)
		ctx := testContextWithUser(userID)

		updates := []models.PrivateDataUpdate{
			{
				ClientSideID:      "cid-missing",
				Version:           1,
				UpdatedRecordHash: "h",
			},
			{
				ClientSideID:      "cid-2",
				Version:           3,
				UpdatedRecordHash: "h2",
			},
		}

		mock.ExpectBegin()

		// Первый UPDATE — запись не найдена: оба NULL
		r1 := sqlmock.NewRows(cteColumns).AddRow(driver.Value((*int64)(nil)), driver.Value((*int64)(nil)))
		mock.ExpectQuery(regexp.QuoteMeta(hashOnlyQuery)).
			WithArgs("cid-missing", userID, "h", int64(1)).
			WillReturnRows(r1)

		mock.ExpectRollback()

		err := repo.updateMultipleRecords(ctx, updates)
		require.ErrorIs(t, err, ErrPrivateDataNotFound)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("error: query execution fails on first record", func(t *testing.T) {
		db, mock := newTestDB(t)
		repo := newTestRepo(t, db).(*privateDataRepository)
		ctx := testContextWithUser(userID)

		updates := []models.PrivateDataUpdate{
			{
				ClientSideID:      "cid-1",
				Version:           5,
				UpdatedRecordHash: "h",
			},
		}

		mock.ExpectBegin()

		mock.ExpectQuery(regexp.QuoteMeta(hashOnlyQuery)).
			WithArgs("cid-1", userID, "h", int64(5)).
			WillReturnError(errors.New("connection lost"))

		mock.ExpectRollback()

		err := repo.updateMultipleRecords(ctx, updates)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "error executing sql query")
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("error: commit fails after successful updates", func(t *testing.T) {
		db, mock := newTestDB(t)
		repo := newTestRepo(t, db).(*privateDataRepository)
		ctx := testContextWithUser(userID)

		updates := []models.PrivateDataUpdate{
			{
				ClientSideID:      "cid-1",
				Version:           5,
				UpdatedRecordHash: "h1",
			},
		}

		mock.ExpectBegin()

		r1 := sqlmock.NewRows(cteColumns).AddRow(driver.Value(&id1), driver.Value(&ver5))
		mock.ExpectQuery(regexp.QuoteMeta(hashOnlyQuery)).
			WithArgs("cid-1", userID, "h1", int64(5)).
			WillReturnRows(r1)

		mock.ExpectCommit().WillReturnError(errors.New("commit failed"))

		err := repo.updateMultipleRecords(ctx, updates)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to commit transaction")
		require.NoError(t, mock.ExpectationsWereMet())
	})
}
