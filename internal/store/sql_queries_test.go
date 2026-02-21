package store

import (
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/MKhiriev/go-pass-keeper/internal/logger"
	"github.com/MKhiriev/go-pass-keeper/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newPrivateRepo creates a *privateDataRepository for query-building tests.
// No DB interaction is expected — sqlmock is only needed to satisfy the struct.
func newPrivateRepo(t *testing.T) *privateDataRepository {
	t.Helper()
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	log := logger.NewLogger("test")
	return &privateDataRepository{
		DB:     &DB{DB: db, logger: log},
		logger: log,
	}
}

// newUserRepo creates a *userRepository for query-building tests.
func newUserRepo(t *testing.T) *userRepository {
	t.Helper()
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	log := logger.NewLogger("test")
	return &userRepository{
		db:     &DB{DB: db, logger: log},
		logger: log,
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// buildSelectAllUserDataQuery
// ─────────────────────────────────────────────────────────────────────────────

func TestBuildSelectAllUserDataQuery(t *testing.T) {
	repo := newPrivateRepo(t)

	query, args, err := repo.buildSelectAllUserDataQuery(ctxWithUserID(10), int64(10))

	require.NoError(t, err)
	assert.Equal(t, []any{int64(10)}, args)
	assert.Contains(t, query, "SELECT *")
	assert.Contains(t, query, "FROM ciphers")
	assert.Contains(t, query, "user_id = $1")
}

// ─────────────────────────────────────────────────────────────────────────────
// buildUpdateQuery
// ─────────────────────────────────────────────────────────────────────────────

func TestBuildUpdateQuery(t *testing.T) {
	const (
		userID       = int64(10)
		clientSideID = "abc-123"
	)

	// Reusable field values.
	meta := models.CipheredMetadata("encrypted-metadata")
	dt := models.LoginPassword
	data := models.CipheredData("encrypted-data")
	notes := models.CipheredNotes("encrypted-notes")
	fields := models.CipheredCustomFields("encrypted-fields")

	tests := []struct {
		name     string
		update   models.PrivateDataUpdate
		wantArgs []any
		wantSQL  []string
		noSQL    []string // substrings that must NOT appear in the query
	}{
		{
			name:     "no_fields_version_zero",
			update:   models.PrivateDataUpdate{ClientSideID: clientSideID},
			wantArgs: []any{clientSideID, userID},
			wantSQL:  []string{"UPDATE ciphers", "updated_at = NOW()"},
			noSQL:    []string{"metadata =", "type =", "data =", "notes =", "additional_fields =", "version ="},
		},
		{
			name:     "only_metadata",
			update:   models.PrivateDataUpdate{ClientSideID: clientSideID, Metadata: &meta},
			wantArgs: []any{meta, clientSideID, userID},
			wantSQL:  []string{"metadata = $1", "client_side_id = $2", "user_id = $3"},
		},
		{
			name:     "only_type",
			update:   models.PrivateDataUpdate{ClientSideID: clientSideID, Type: &dt},
			wantArgs: []any{dt, clientSideID, userID},
			wantSQL:  []string{"type = $1", "client_side_id = $2", "user_id = $3"},
		},
		{
			name:     "only_data",
			update:   models.PrivateDataUpdate{ClientSideID: clientSideID, Data: &data},
			wantArgs: []any{data, clientSideID, userID},
			wantSQL:  []string{"data = $1", "client_side_id = $2", "user_id = $3"},
		},
		{
			name:     "only_notes",
			update:   models.PrivateDataUpdate{ClientSideID: clientSideID, Notes: &notes},
			wantArgs: []any{notes, clientSideID, userID},
			wantSQL:  []string{"notes = $1", "client_side_id = $2", "user_id = $3"},
		},
		{
			name:     "only_additional_fields",
			update:   models.PrivateDataUpdate{ClientSideID: clientSideID, AdditionalFields: &fields},
			wantArgs: []any{fields, clientSideID, userID},
			wantSQL:  []string{"additional_fields = $1", "client_side_id = $2", "user_id = $3"},
		},
		{
			name:     "only_version",
			update:   models.PrivateDataUpdate{ClientSideID: clientSideID, Version: int64(2)},
			wantArgs: []any{int64(2), clientSideID, userID},
			wantSQL:  []string{"version = $1", "client_side_id = $2", "user_id = $3"},
		},
		{
			// Version=0 is the zero value and must be excluded from SET.
			name:     "version_zero_excluded",
			update:   models.PrivateDataUpdate{ClientSideID: clientSideID, Metadata: &meta, Version: 0},
			wantArgs: []any{meta, clientSideID, userID},
			wantSQL:  []string{"metadata = $1"},
			noSQL:    []string{"version ="},
		},
		{
			name:     "metadata_and_version",
			update:   models.PrivateDataUpdate{ClientSideID: clientSideID, Metadata: &meta, Version: int64(3)},
			wantArgs: []any{meta, int64(3), clientSideID, userID},
			wantSQL:  []string{"metadata = $1", "version = $2", "client_side_id = $3", "user_id = $4"},
		},
		{
			name: "all_fields_with_version",
			update: models.PrivateDataUpdate{
				ClientSideID:     clientSideID,
				Metadata:         &meta,
				Type:             &dt,
				Data:             &data,
				Notes:            &notes,
				AdditionalFields: &fields,
				Version:          int64(5),
			},
			wantArgs: []any{meta, dt, data, notes, fields, int64(5), clientSideID, userID},
			wantSQL: []string{
				"metadata = $1", "type = $2", "data = $3",
				"notes = $4", "additional_fields = $5", "version = $6",
				"client_side_id = $7", "user_id = $8",
			},
		},
		{
			name: "all_fields_without_version",
			update: models.PrivateDataUpdate{
				ClientSideID:     clientSideID,
				Metadata:         &meta,
				Type:             &dt,
				Data:             &data,
				Notes:            &notes,
				AdditionalFields: &fields,
			},
			wantArgs: []any{meta, dt, data, notes, fields, clientSideID, userID},
			wantSQL: []string{
				"metadata = $1", "type = $2", "data = $3",
				"notes = $4", "additional_fields = $5",
				"client_side_id = $6", "user_id = $7",
			},
			noSQL: []string{"version ="},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newPrivateRepo(t)
			ctx := ctxWithUserID(userID)

			query, args, err := repo.buildUpdateQuery(ctx, tt.update)

			require.NoError(t, err)
			assert.Equal(t, tt.wantArgs, args)
			for _, s := range tt.wantSQL {
				assert.Contains(t, query, s, "expected substring in query: %q", s)
			}
			for _, s := range tt.noSQL {
				assert.NotContains(t, query, s, "unexpected substring in query: %q", s)
			}
		})
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// buildGetPrivateDataQuery
// ─────────────────────────────────────────────────────────────────────────────

func TestBuildGetPrivateDataQuery(t *testing.T) {
	const userID = int64(10)

	tests := []struct {
		name     string
		req      models.DownloadRequest
		wantArgs []any
		wantSQL  []string
		noSQL    []string
	}{
		{
			name:     "only_user_id",
			req:      models.DownloadRequest{UserID: userID},
			wantArgs: []any{userID},
			wantSQL:  []string{"SELECT *", "FROM ciphers", "user_id = $1"},
			noSQL:    []string{"id IN", "type IN"},
		},
		{
			// Separate .Where() calls → order preserved: user_id first, then id.
			name:     "with_ids",
			req:      models.DownloadRequest{UserID: userID, IDs: []int64{1, 2}},
			wantArgs: []any{userID, int64(1), int64(2)},
			wantSQL:  []string{"user_id = $1", "id IN ("},
		},
		{
			name:     "with_types",
			req:      models.DownloadRequest{UserID: userID, Types: []models.DataType{models.LoginPassword}},
			wantArgs: []any{userID, models.LoginPassword},
			wantSQL:  []string{"user_id = $1", "type IN ("},
			noSQL:    []string{"id IN"},
		},
		{
			name: "with_ids_and_types",
			req: models.DownloadRequest{
				UserID: userID,
				IDs:    []int64{5},
				Types:  []models.DataType{models.BankCard},
			},
			wantArgs: []any{userID, int64(5), models.BankCard},
			wantSQL:  []string{"user_id = $1", "id IN (", "type IN ("},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newPrivateRepo(t)
			ctx := ctxWithUserID(userID)

			query, args, err := repo.buildGetPrivateDataQuery(ctx, tt.req)

			require.NoError(t, err)
			assert.Equal(t, tt.wantArgs, args)
			for _, s := range tt.wantSQL {
				assert.Contains(t, query, s, "expected substring in query: %q", s)
			}
			for _, s := range tt.noSQL {
				assert.NotContains(t, query, s, "unexpected substring in query: %q", s)
			}
		})
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// buildDeletePrivateDataQuery
// ─────────────────────────────────────────────────────────────────────────────

func TestBuildDeletePrivateDataQuery(t *testing.T) {
	// sq.Eq{"user_id": ..., "id": ...} sorts keys alphabetically:
	// "id" < "user_id" → args = [...IDs, userID]
	tests := []struct {
		name     string
		req      models.DeleteRequest
		wantArgs []any
		wantSQL  []string
	}{
		{
			name:     "single_id",
			req:      models.DeleteRequest{UserID: int64(10), IDs: []int64{42}},
			wantArgs: []any{int64(42), int64(10)},
			wantSQL:  []string{"DELETE FROM ciphers", "id IN (", "user_id = $2"},
		},
		{
			name:     "multiple_ids",
			req:      models.DeleteRequest{UserID: int64(10), IDs: []int64{1, 2, 3}},
			wantArgs: []any{int64(1), int64(2), int64(3), int64(10)},
			wantSQL:  []string{"DELETE FROM ciphers", "id IN (", "user_id = $4"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newPrivateRepo(t)
			ctx := ctxWithUserID(int64(10))

			query, args, err := repo.buildDeletePrivateDataQuery(ctx, tt.req)

			require.NoError(t, err)
			assert.Equal(t, tt.wantArgs, args)
			for _, s := range tt.wantSQL {
				assert.Contains(t, query, s, "expected substring in query: %q", s)
			}
		})
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// buildCreateUserQuery
// ─────────────────────────────────────────────────────────────────────────────

func TestBuildCreateUserQuery(t *testing.T) {
	repo := newUserRepo(t)

	user := models.User{
		Login:              "john@example.com",
		MasterPassword:     "hashed-master-password",
		MasterPasswordHint: "favorite pet name",
		Name:               "John Doe",
	}

	query, args, err := repo.buildCreateUserQuery(ctxWithUserID(0), user)

	require.NoError(t, err)
	assert.Equal(t, []any{
		user.Login,
		user.MasterPassword,
		user.MasterPasswordHint,
		user.Name,
	}, args)
	assert.Contains(t, query, "INSERT INTO users")
	assert.Contains(t, query, "RETURNING user_id")
}

// ─────────────────────────────────────────────────────────────────────────────
// buildFindUserByLoginQuery
// ─────────────────────────────────────────────────────────────────────────────

func TestBuildFindUserByLoginQuery(t *testing.T) {
	repo := newUserRepo(t)
	login := "john@example.com"

	query, args, err := repo.buildFindUserByLoginQuery(ctxWithUserID(0), login)

	require.NoError(t, err)
	assert.Equal(t, []any{login}, args)
	assert.Contains(t, query, "SELECT user_id")
	assert.Contains(t, query, "FROM users")
	assert.Contains(t, query, "login = $1")
}
