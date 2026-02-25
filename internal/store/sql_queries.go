package store

import (
	"context"
	"fmt"
	"strings"

	"github.com/MKhiriev/go-pass-keeper/internal/logger"
	"github.com/MKhiriev/go-pass-keeper/internal/utils"
	"github.com/MKhiriev/go-pass-keeper/models"
	sq "github.com/Masterminds/squirrel"
)

const (
	createUser = `
		INSERT INTO users (login, auth_hash, master_password_hint, name) 
    	VALUES ($1, $2, $3, $4) 
    	RETURNING user_id, login, auth_hash, master_password_hint, name, created_at;`

	findUserByLogin = `
		SELECT user_id, login, auth_hash, master_password_hint, name, created_at 
    	FROM users 
    	WHERE login = $1;`

	savePrivateData = `
		INSERT INTO ciphers (
			client_side_id,
			user_id,
			metadata,
			type,
			data,
			notes,
			additional_fields,
			version,
			hash,
			created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id;`

	getAllUserPrivateData = `
		SELECT
			id,
			user_id,
			type,
			metadata,
			data,
			notes,
			additional_fields,
			created_at,
			updated_at,
			version,
			client_side_id,
			hash,
			deleted
		FROM ciphers
		WHERE user_id = $1;`

	getAllUserDataState = `
		SELECT client_side_id, hash, version, deleted, updated_at
		FROM ciphers
		WHERE user_id = $1;`

	deletePrivateDataQuery = `
		WITH target_record AS (
			SELECT id, version
			FROM ciphers
			WHERE client_side_id = $1 AND user_id = $2
		),
		updated_record AS (
			UPDATE ciphers
			SET
				deleted = TRUE,
				updated_at = NOW(),
				version = version + 1
			WHERE client_side_id = $1
			  AND user_id = $2
			  AND version = $3
			RETURNING id
		)
		SELECT
			(SELECT id FROM updated_record)       AS updated_id,
			(SELECT version FROM target_record)   AS current_db_version;`
)

var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

// buildSelectAllUserDataQuery builds SELECT query for all user private data
// checked!
func buildSelectAllUserDataQuery(ctx context.Context, userID int64) (string, []any, error) {
	qb := psql.Select(
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
	).From("ciphers").Where(sq.Eq{"user_id": userID})

	query, args, err := qb.ToSql()
	if err != nil {
		return "", nil, fmt.Errorf("error building query for getting all user data: %w", err)
	}

	logger.FromContext(ctx).Debug().Str("query", query).Any("args", args).Msg("built select query")
	return query, args, nil
}

// buildGetPrivateDataQuery builds SELECT query with optional ID filter
// checked!
func buildGetPrivateDataQuery(ctx context.Context, req models.DownloadRequest) (string, []any, error) {
	qb := psql.
		Select(
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
		).
		From("ciphers").
		Where(sq.Eq{"user_id": req.UserID})

	if len(req.ClientSideIDs) > 0 {
		qb = qb.Where(sq.Eq{"client_side_id": req.ClientSideIDs})
	}

	query, args, err := qb.ToSql()
	if err != nil {
		return "", nil, fmt.Errorf("error building query for getting private data with filters: %w", err)
	}

	logger.FromContext(ctx).Debug().Str("query", query).Any("args", args).Msg("built get private data query")
	return query, args, nil
}

// checked!
func buildGetStatesSyncQuery(ctx context.Context, syncRequest models.SyncRequest) (string, []any, error) {
	qb := psql.
		Select(
			"client_side_id",
			"hash",
			"version",
			"deleted",
			"updated_at",
		).
		From("ciphers").
		Where(sq.Eq{"user_id": syncRequest.UserID})

	if len(syncRequest.ClientSideIDs) > 0 {
		qb = qb.Where(sq.Eq{"client_side_id": syncRequest.ClientSideIDs})
	}

	query, args, err := qb.ToSql()
	if err != nil {
		return "", nil, fmt.Errorf("error building query for getting private data states with filters: %w", err)
	}

	logger.FromContext(ctx).
		Debug().
		Str("query", query).
		Any("args", args).
		Msg("built get private data states query")

	return query, args, nil
}

// buildUpdateQuery dynamically builds UPDATE query with CTE for optimistic locking.
// Returns a query that always returns a row if the record exists,
// allowing to distinguish between NotFound and VersionConflict.
// checked!
func buildUpdateQuery(ctx context.Context, update models.PrivateDataUpdate) (string, []any, error) {
	userID, _ := utils.GetUserIDFromContext(ctx)

	setClauses := []string{
		"updated_at = NOW()",
		"version = version + 1",
	}
	args := []any{update.ClientSideID, userID} // $1 = client_side_id, $2 = user_id
	argIndex := 3

	if update.FieldsUpdate.Metadata != nil {
		setClauses = append(setClauses, fmt.Sprintf("metadata = $%d", argIndex))
		args = append(args, *update.FieldsUpdate.Metadata)
		argIndex++
	}

	if update.FieldsUpdate.Data != nil {
		setClauses = append(setClauses, fmt.Sprintf("data = $%d", argIndex))
		args = append(args, *update.FieldsUpdate.Data)
		argIndex++
	}

	if update.FieldsUpdate.Notes != nil {
		setClauses = append(setClauses, fmt.Sprintf("notes = $%d", argIndex))
		args = append(args, *update.FieldsUpdate.Notes)
		argIndex++
	}

	if update.FieldsUpdate.AdditionalFields != nil {
		setClauses = append(setClauses, fmt.Sprintf("additional_fields = $%d", argIndex))
		args = append(args, *update.FieldsUpdate.AdditionalFields)
		argIndex++
	}

	if update.UpdatedRecordHash != "" {
		setClauses = append(setClauses, fmt.Sprintf("hash = $%d", argIndex))
		args = append(args, update.UpdatedRecordHash)
		argIndex++
	}

	versionPlaceholder := fmt.Sprintf("$%d", argIndex)
	args = append(args, update.Version)

	query := fmt.Sprintf(`
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
		strings.Join(setClauses, ", "),
		versionPlaceholder,
	)

	logger.FromContext(ctx).
		Debug().
		Str("query", query).
		Any("args", args).
		Msg("built update query with CTE")

	return query, args, nil
}

// buildCreateUserQuery builds INSERT query with RETURNING clause
func buildCreateUserQuery(ctx context.Context, user models.User) (string, []any, error) {
	qb := psql.Insert("users").
		Columns("login", "auth_hash", "master_password_hint", "name").
		Values(user.Login, user.AuthHash, user.MasterPasswordHint, user.Name).
		Suffix("RETURNING user_id, login, auth_hash, master_password_hint, name, created_at")

	query, args, err := qb.ToSql()
	if err != nil {
		return "", nil, fmt.Errorf("error building create user query: %w", err)
	}

	logger.FromContext(ctx).Debug().Str("query", query).Any("args", args).Msg("built create user query")
	return query, args, nil
}

// buildFindUserByLoginQuery builds SELECT query for finding user by login
func buildFindUserByLoginQuery(ctx context.Context, login string) (string, []any, error) {
	qb := psql.Select("user_id", "login", "auth_hash", "master_password_hint", "name", "created_at").
		From("users").
		Where(sq.Eq{"login": login})

	query, args, err := qb.ToSql()
	if err != nil {
		return "", nil, fmt.Errorf("error building find user by login query: %w", err)
	}

	logger.FromContext(ctx).Debug().Str("query", query).Any("args", args).Msg("built find user by login query")
	return query, args, nil
}
