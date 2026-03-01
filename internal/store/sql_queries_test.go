// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Rasul Khiriev

package store

import (
	"context"
	"strings"
	"testing"

	"github.com/MKhiriev/go-pass-keeper/internal/utils"
	"github.com/MKhiriev/go-pass-keeper/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_buildSelectAllUserDataQuery_SQLContainsParts(t *testing.T) {
	ctx := context.Background()
	userID := int64(42)

	query, args, err := buildSelectAllUserDataQuery(ctx, userID)
	require.NoError(t, err)

	// args checks
	require.Len(t, args, 1)
	require.Equal(t, userID, args[0])

	// query checks (contains parts)
	q := strings.ToLower(query)

	require.Contains(t, q, "select")
	require.Contains(t, q, "from ciphers")
	require.Contains(t, q, "where")
	require.Contains(t, q, "user_id")

	// placeholder format should be $1 (Postgres)
	require.Contains(t, query, "$1")

	// columns presence (subset / key columns)
	require.Contains(t, q, "id")
	require.Contains(t, q, "client_side_id")
	require.Contains(t, q, "version")
	require.Contains(t, q, "deleted")
	require.Contains(t, q, "updated_at")
}

func Test_buildSelectAllUserDataQuery_SelectsAllExpectedColumns(t *testing.T) {
	ctx := context.Background()

	query, _, err := buildSelectAllUserDataQuery(ctx, 1)
	require.NoError(t, err)

	q := strings.ToLower(query)

	// Check that all expected columns are present in the SELECT section.
	// This is a "contains" check; it does not enforce order but catches regressions quickly.
	cols := []string{
		"id",
		"user_id",
		"type",
		"metadata",
		"data",
		"notes",
		"additional_fields",
		"created_at",
		"updated_at",
		"version",
		"client_side_id",
		"hash",
		"deleted",
	}
	for _, c := range cols {
		require.Contains(t, q, c)
	}
}

func Test_buildSelectAllUserDataQuery(t *testing.T) {
	tests := []struct {
		name       string
		userID     int64
		wantErr    bool
		checkQuery func(t *testing.T, query string, args []any)
	}{
		{
			name:    "success: valid user ID",
			userID:  42,
			wantErr: false,
			checkQuery: func(t *testing.T, query string, args []any) {
				// Check that all 13 expected columns are present.
				expectedColumns := []string{
					"id", "user_id", "type", "metadata", "data",
					"notes", "additional_fields", "created_at", "updated_at",
					"version", "client_side_id", "hash", "deleted",
				}
				for _, col := range expectedColumns {
					assert.True(t, strings.Contains(query, col),
						"query should contain column %q", col)
				}

				// Check query structure.
				assert.True(t, strings.Contains(strings.ToUpper(query), "SELECT"))
				assert.True(t, strings.Contains(strings.ToUpper(query), "FROM"))
				assert.True(t, strings.Contains(query, "ciphers"))
				assert.True(t, strings.Contains(strings.ToUpper(query), "WHERE"))
				assert.True(t, strings.Contains(query, "user_id"))

				// Check placeholder format ($1 for PostgreSQL).
				assert.True(t, strings.Contains(query, "$1"),
					"query should use $1 placeholder for PostgreSQL")

				// Check query arguments.
				require.Len(t, args, 1)
				assert.Equal(t, int64(42), args[0])
			},
		},
		{
			name:    "success: zero user ID",
			userID:  0,
			wantErr: false,
			checkQuery: func(t *testing.T, query string, args []any) {
				require.Len(t, args, 1)
				assert.Equal(t, int64(0), args[0],
					"zero user ID should be passed as-is (DB will return empty result)")
			},
		},
		{
			name:    "success: negative user ID",
			userID:  -1,
			wantErr: false,
			checkQuery: func(t *testing.T, query string, args []any) {
				// buildSelectAllUserDataQuery does not validate userID.
				// Validation is a service-layer concern; this function only builds SQL.
				require.Len(t, args, 1)
				assert.Equal(t, int64(-1), args[0])
			},
		},
		{
			name:    "success: large user ID",
			userID:  9999999999,
			wantErr: false,
			checkQuery: func(t *testing.T, query string, args []any) {
				require.Len(t, args, 1)
				assert.Equal(t, int64(9999999999), args[0])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			query, args, err := buildSelectAllUserDataQuery(ctx, tt.userID)

			if tt.wantErr {
				require.Error(t, err)
				assert.Empty(t, query)
				assert.Nil(t, args)
				return
			}

			require.NoError(t, err)
			assert.NotEmpty(t, query)
			assert.NotNil(t, args)

			if tt.checkQuery != nil {
				tt.checkQuery(t, query, args)
			}
		})
	}
}

func Test_buildGetPrivateDataQuery_SQLContainsParts(t *testing.T) {
	tests := []struct {
		name       string
		req        models.DownloadRequest
		checkQuery func(t *testing.T, query string, args []any)
	}{
		{
			name: "success: only userID filter (no ClientSideIDs)",
			req: models.DownloadRequest{
				UserID:        42,
				ClientSideIDs: nil,
			},
			checkQuery: func(t *testing.T, query string, args []any) {
				q := strings.ToLower(query)

				// Query structure.
				require.Contains(t, q, "select")
				require.Contains(t, q, "from ciphers")
				require.Contains(t, q, "where")
				require.Contains(t, q, "user_id")

				// Postgres placeholder
				require.Contains(t, query, "$1")

				// client_side_id filter must NOT be added.
				require.NotContains(t, q, "client_side_id =")
				require.NotContains(t, q, "client_side_id in")

				// Exactly one argument: userID.
				require.Len(t, args, 1)
				require.Equal(t, int64(42), args[0])
			},
		},
		{
			name: "success: userID + single ClientSideID",
			req: models.DownloadRequest{
				UserID:        42,
				ClientSideIDs: []string{"abc-123"},
			},
			checkQuery: func(t *testing.T, query string, args []any) {
				q := strings.ToLower(query)

				require.Contains(t, q, "select")
				require.Contains(t, q, "from ciphers")
				require.Contains(t, q, "where")
				require.Contains(t, q, "user_id")
				require.Contains(t, q, "client_side_id")

				// Two placeholders: $1 (user_id), $2 (client_side_id).
				require.Contains(t, query, "$1")
				require.Contains(t, query, "$2")

				// Two arguments.
				require.Len(t, args, 2)
				require.Equal(t, int64(42), args[0])
				require.Equal(t, "abc-123", args[1])
			},
		},
		{
			name: "success: userID + multiple ClientSideIDs",
			req: models.DownloadRequest{
				UserID:        42,
				ClientSideIDs: []string{"abc-123", "def-456", "ghi-789"},
			},
			checkQuery: func(t *testing.T, query string, args []any) {
				q := strings.ToLower(query)

				require.Contains(t, q, "client_side_id")

				// squirrel generates IN ($2,$3,$4) for a slice.
				require.Contains(t, query, "$2")
				require.Contains(t, query, "$3")
				require.Contains(t, query, "$4")

				// Four arguments: userID + 3 client_side_id values.
				require.Len(t, args, 4)
				require.Equal(t, int64(42), args[0])
				require.Equal(t, "abc-123", args[1])
				require.Equal(t, "def-456", args[2])
				require.Equal(t, "ghi-789", args[3])
			},
		},
		{
			name: "success: empty ClientSideIDs slice treated as no filter",
			req: models.DownloadRequest{
				UserID:        42,
				ClientSideIDs: []string{},
			},
			checkQuery: func(t *testing.T, query string, args []any) {
				q := strings.ToLower(query)

				// Empty slice: client_side_id filter is not added to WHERE.
				// client_side_id is present in SELECT, so check only the WHERE section.
				whereIdx := strings.Index(q, "where")
				require.NotEqual(t, -1, whereIdx, "query should contain WHERE clause")
				wherePart := q[whereIdx:]
				require.NotContains(t, wherePart, "client_side_id",
					"WHERE clause should not contain client_side_id filter for empty slice")

				// Only one argument: userID.
				require.Len(t, args, 1)
				require.Equal(t, int64(42), args[0])
			},
		},
		{
			name: "success: all expected columns present",
			req: models.DownloadRequest{
				UserID: 1,
			},
			checkQuery: func(t *testing.T, query string, args []any) {
				q := strings.ToLower(query)

				expectedCols := []string{
					"id", "user_id", "type", "metadata", "data",
					"notes", "additional_fields", "created_at", "updated_at",
					"version", "client_side_id", "hash", "deleted",
				}
				for _, col := range expectedCols {
					require.Contains(t, q, col, "query should contain column %q", col)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			query, args, err := buildGetPrivateDataQuery(ctx, tt.req)

			require.NoError(t, err)
			require.NotEmpty(t, query)
			require.NotNil(t, args)

			tt.checkQuery(t, query, args)
		})
	}
}

func Test_buildGetStatesSyncQuery_SQLContainsParts(t *testing.T) {
	tests := []struct {
		name       string
		req        models.SyncRequest
		checkQuery func(t *testing.T, query string, args []any)
	}{
		{
			name: "success: only userID filter (no ClientSideIDs)",
			req: models.SyncRequest{
				UserID:        42,
				ClientSideIDs: nil,
			},
			checkQuery: func(t *testing.T, query string, args []any) {
				q := strings.ToLower(query)

				// Query structure.
				require.Contains(t, q, "select")
				require.Contains(t, q, "from ciphers")
				require.Contains(t, q, "where")
				require.Contains(t, q, "user_id")

				// Postgres placeholder
				require.Contains(t, query, "$1")

				// WHERE must not contain a client_side_id filter.
				whereIdx := strings.Index(q, "where")
				require.NotEqual(t, -1, whereIdx)
				wherePart := q[whereIdx:]
				require.NotContains(t, wherePart, "client_side_id",
					"WHERE clause should not contain client_side_id filter when ClientSideIDs is nil")

				// Exactly one argument: userID.
				require.Len(t, args, 1)
				require.Equal(t, int64(42), args[0])
			},
		},
		{
			name: "success: empty ClientSideIDs slice treated as no filter",
			req: models.SyncRequest{
				UserID:        42,
				ClientSideIDs: []string{},
			},
			checkQuery: func(t *testing.T, query string, args []any) {
				q := strings.ToLower(query)

				// Empty slice: client_side_id filter is not added to WHERE.
				whereIdx := strings.Index(q, "where")
				require.NotEqual(t, -1, whereIdx)
				wherePart := q[whereIdx:]
				require.NotContains(t, wherePart, "client_side_id",
					"WHERE clause should not contain client_side_id filter for empty slice")

				// Only one argument: userID.
				require.Len(t, args, 1)
				require.Equal(t, int64(42), args[0])
			},
		},
		{
			name: "success: userID + single ClientSideID",
			req: models.SyncRequest{
				UserID:        42,
				ClientSideIDs: []string{"abc-123"},
			},
			checkQuery: func(t *testing.T, query string, args []any) {
				q := strings.ToLower(query)

				require.Contains(t, q, "where")
				require.Contains(t, q, "user_id")

				// WHERE contains a client_side_id filter.
				whereIdx := strings.Index(q, "where")
				wherePart := q[whereIdx:]
				require.Contains(t, wherePart, "client_side_id")

				// $1 (user_id), $2 (client_side_id)
				require.Contains(t, query, "$1")
				require.Contains(t, query, "$2")

				// Two arguments.
				require.Len(t, args, 2)
				require.Equal(t, int64(42), args[0])
				require.Equal(t, "abc-123", args[1])
			},
		},
		{
			name: "success: userID + multiple ClientSideIDs",
			req: models.SyncRequest{
				UserID:        42,
				ClientSideIDs: []string{"abc-123", "def-456", "ghi-789"},
			},
			checkQuery: func(t *testing.T, query string, args []any) {
				q := strings.ToLower(query)

				// WHERE contains an IN filter by client_side_id.
				whereIdx := strings.Index(q, "where")
				wherePart := q[whereIdx:]
				require.Contains(t, wherePart, "client_side_id")

				// squirrel generates IN ($2,$3,$4).
				require.Contains(t, query, "$2")
				require.Contains(t, query, "$3")
				require.Contains(t, query, "$4")

				// Four arguments: userID + 3 client_side_id values.
				require.Len(t, args, 4)
				require.Equal(t, int64(42), args[0])
				require.Equal(t, "abc-123", args[1])
				require.Equal(t, "def-456", args[2])
				require.Equal(t, "ghi-789", args[3])
			},
		},
		{
			name: "success: only expected 5 columns selected (not SELECT *)",
			req: models.SyncRequest{
				UserID: 1,
			},
			checkQuery: func(t *testing.T, query string, args []any) {
				q := strings.ToLower(query)

				// Extract SELECT section (before FROM).
				fromIdx := strings.Index(q, " from ")
				require.NotEqual(t, -1, fromIdx)
				selectPart := q[:fromIdx]

				// Check that exactly the required 5 columns are present.
				expectedCols := []string{
					"client_side_id",
					"hash",
					"version",
					"deleted",
					"updated_at",
				}
				for _, col := range expectedCols {
					require.Contains(t, selectPart, col,
						"SELECT part should contain column %q", col)
				}

				// Ensure this is not SELECT *.
				require.NotContains(t, selectPart, "*",
					"query should not use SELECT *")

				// Ensure no extra columns are included in SELECT.
				unexpectedCols := []string{"metadata", "data", "notes", "additional_fields"}
				for _, col := range unexpectedCols {
					require.NotContains(t, selectPart, col,
						"SELECT part should NOT contain column %q", col)
				}
			},
		},
		{
			name: "success: query is idempotent for same request",
			req: models.SyncRequest{
				UserID:        99,
				ClientSideIDs: []string{"x-1", "x-2"},
			},
			checkQuery: func(t *testing.T, query string, args []any) {
				query2, args2, err2 := buildGetStatesSyncQuery(context.Background(), models.SyncRequest{
					UserID:        99,
					ClientSideIDs: []string{"x-1", "x-2"},
				})
				require.NoError(t, err2)
				require.Equal(t, query, query2)
				require.Equal(t, args, args2)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			query, args, err := buildGetStatesSyncQuery(ctx, tt.req)

			require.NoError(t, err)
			require.NotEmpty(t, query)
			require.NotNil(t, args)

			tt.checkQuery(t, query, args)
		})
	}
}

func Test_buildUpdateQuery_SQLContainsParts(t *testing.T) {
	userID := int64(42)

	meta := models.CipheredMetadata("m1")
	data := models.CipheredData("d1")
	notes := models.CipheredNotes("n1")
	addl := models.CipheredCustomFields("a1")

	tests := []struct {
		name       string
		ctx        context.Context
		update     models.PrivateDataUpdate
		checkQuery func(t *testing.T, query string, args []any)
	}{
		{
			name: "success: no optional fields, no hash (version placeholder is $3)",
			ctx:  context.WithValue(context.Background(), utils.UserIDCtxKey, userID),
			update: models.PrivateDataUpdate{
				ClientSideID:      "csid-1",
				FieldsUpdate:      models.FieldsUpdate{},
				UpdatedRecordHash: "",
				Version:           7,
			},
			checkQuery: func(t *testing.T, query string, args []any) {
				q := strings.ToLower(query)

				// CTE structure
				require.Contains(t, q, "with target_record as")
				require.Contains(t, q, "updated_record as")
				require.Contains(t, q, "update ciphers")
				require.Contains(t, q, "from ciphers")
				require.Contains(t, q, "returning id")

				// Always sets these
				require.Contains(t, q, "updated_at = now()")
				require.Contains(t, q, "version = version + 1")

				// Filters / optimistic locking uses placeholders $1, $2, $3
				require.Contains(t, query, "client_side_id = $1")
				require.Contains(t, query, "user_id = $2")
				require.Contains(t, query, "version = $3") // "AND version = $3" in real query

				// No optional SET clauses
				// Note: use " data = $" (with leading space) to avoid matching "metadata = $"
				require.NotContains(t, q, "metadata = $")
				require.NotContains(t, q, " data = $")
				require.NotContains(t, q, "notes = $")
				require.NotContains(t, q, "additional_fields = $")
				require.NotContains(t, q, "hash = $")

				// Args: csid, userID, version
				require.Len(t, args, 3)
				require.Equal(t, "csid-1", args[0])
				require.Equal(t, userID, args[1])
				require.Equal(t, int64(7), args[2])
			},
		},
		{
			name: "success: all optional fields + hash (version placeholder is $8)",
			ctx:  context.WithValue(context.Background(), utils.UserIDCtxKey, userID),
			update: models.PrivateDataUpdate{
				ClientSideID: "csid-2",
				FieldsUpdate: models.FieldsUpdate{
					Metadata:         &meta,
					Data:             &data,
					Notes:            &notes,
					AdditionalFields: &addl,
				},
				UpdatedRecordHash: "hash-xyz",
				Version:           5,
			},
			checkQuery: func(t *testing.T, query string, args []any) {
				q := strings.ToLower(query)

				// SET placeholders are sequential: $3..$7, version is $8
				require.Contains(t, q, "metadata = $3")
				require.Contains(t, q, " data = $4") // leading space to avoid matching "metadata"
				require.Contains(t, q, "notes = $5")
				require.Contains(t, q, "additional_fields = $6")
				require.Contains(t, q, "hash = $7")
				require.Contains(t, q, "version = $8") // "AND version = $8" in real query

				// Args order: csid, userID, meta, data, notes, addl, hash, version
				require.Len(t, args, 8)
				require.Equal(t, "csid-2", args[0])
				require.Equal(t, userID, args[1])
				require.Equal(t, models.CipheredMetadata("m1"), args[2])
				require.Equal(t, models.CipheredData("d1"), args[3])
				require.Equal(t, models.CipheredNotes("n1"), args[4])
				require.Equal(t, models.CipheredCustomFields("a1"), args[5])
				require.Equal(t, "hash-xyz", args[6])
				require.Equal(t, int64(5), args[7])
			},
		},
		{
			name: "success: only metadata + hash (version placeholder is $5)",
			ctx:  context.WithValue(context.Background(), utils.UserIDCtxKey, userID),
			update: models.PrivateDataUpdate{
				ClientSideID: "csid-3",
				FieldsUpdate: models.FieldsUpdate{
					Metadata: &meta,
				},
				UpdatedRecordHash: "hash-only-meta",
				Version:           3,
			},
			checkQuery: func(t *testing.T, query string, args []any) {
				q := strings.ToLower(query)

				require.Contains(t, q, "metadata = $3")
				require.Contains(t, q, "hash = $4")
				require.Contains(t, q, "version = $5") // "AND version = $5" in real query

				require.NotContains(t, q, " data = $") // leading space to avoid matching "metadata"
				require.NotContains(t, q, "notes = $")
				require.NotContains(t, q, "additional_fields = $")

				require.Len(t, args, 5)
				require.Equal(t, "csid-3", args[0])
				require.Equal(t, userID, args[1])
				require.Equal(t, models.CipheredMetadata("m1"), args[2])
				require.Equal(t, "hash-only-meta", args[3])
				require.Equal(t, int64(3), args[4])
			},
		},
		{
			name: "success: data + notes (no hash) (version placeholder is $5)",
			ctx:  context.WithValue(context.Background(), utils.UserIDCtxKey, userID),
			update: models.PrivateDataUpdate{
				ClientSideID: "csid-4",
				FieldsUpdate: models.FieldsUpdate{
					Data:  &data,
					Notes: &notes,
				},
				UpdatedRecordHash: "",
				Version:           2,
			},
			checkQuery: func(t *testing.T, query string, args []any) {
				q := strings.ToLower(query)

				require.Contains(t, q, " data = $3") // leading space to avoid matching "metadata"
				require.Contains(t, q, "notes = $4")
				require.Contains(t, q, "version = $5") // "AND version = $5" in real query
				require.NotContains(t, q, "hash = $")

				require.Len(t, args, 5)
				require.Equal(t, "csid-4", args[0])
				require.Equal(t, userID, args[1])
				require.Equal(t, models.CipheredData("d1"), args[2])
				require.Equal(t, models.CipheredNotes("n1"), args[3])
				require.Equal(t, int64(2), args[4])
			},
		},
		{
			name: "success: idempotent for same ctx + update",
			ctx:  context.WithValue(context.Background(), utils.UserIDCtxKey, userID),
			update: models.PrivateDataUpdate{
				ClientSideID: "csid-5",
				FieldsUpdate: models.FieldsUpdate{
					Metadata: &meta,
				},
				UpdatedRecordHash: "h5",
				Version:           10,
			},
			checkQuery: func(t *testing.T, query string, args []any) {
				query2, args2, err2 := buildUpdateQuery(
					context.WithValue(context.Background(), utils.UserIDCtxKey, userID),
					models.PrivateDataUpdate{
						ClientSideID: "csid-5",
						FieldsUpdate: models.FieldsUpdate{
							Metadata: &meta,
						},
						UpdatedRecordHash: "h5",
						Version:           10,
					},
				)
				require.NoError(t, err2)
				require.Equal(t, query, query2)
				require.Equal(t, args, args2)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query, args, err := buildUpdateQuery(tt.ctx, tt.update)

			require.NoError(t, err)
			require.NotEmpty(t, query)
			require.NotNil(t, args)

			tt.checkQuery(t, query, args)
		})
	}
}
