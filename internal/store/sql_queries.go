package store

import (
	"context"
	"fmt"

	"github.com/MKhiriev/go-pass-keeper/internal/logger"
	"github.com/MKhiriev/go-pass-keeper/internal/utils"
	"github.com/MKhiriev/go-pass-keeper/models"
	sq "github.com/Masterminds/squirrel"
)

const (
	createUser = `INSERT INTO users (login, master_password, master_password_hint, name) 
    VALUES ($1, $2, $3, $4) 
    RETURNING user_id, login, master_password, master_password_hint, name, created_at;`

	findUserByLogin = `SELECT user_id, login, master_password, master_password_hint, name, created_at 
    FROM users 
    WHERE login = $1;`

	getRequestedPrivateData = `SELECT *
		FROM ciphers
		WHERE user_id = $1`
	getRequestedPrivateDataWhereID   = ` AND id = ANY($%d)`
	getRequestedPrivateDataWhereType = ` AND type = ANY($%d)`

	savePrivateData = `INSERT INTO ciphers (
			user_id, 
			metadata, 
			type, 
			data, 
			notes,
			additional_fields, 
            version,
			created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8);`

	deletePrivateData = `DELETE FROM ciphers
		WHERE user_id = $1 AND id = ANY($2);`

	// todo implement me!
	updatePrivateDataBase = `
		UPDATE ciphers
		SET updated_at = NOW()`
	updatePrivateDataWhere = `
        WHERE id = $%d AND user_id = $%d`
	updatePrivateDataWhereWithVersion = `
        WHERE id = $%d AND user_id = $%d AND version = $%d - 1`
)

var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

// buildSelectAllUserDataQuery builds SELECT query for all user private data
func (p *privateDataRepository) buildSelectAllUserDataQuery(ctx context.Context, userID int64) (string, []any, error) {
	qb := psql.Select("*").From("ciphers").Where(sq.Eq{"user_id": userID})

	query, args, err := qb.ToSql()
	if err != nil {
		return "", nil, fmt.Errorf("error building query for getting all user data: %w", err)
	}

	logger.FromContext(ctx).Info().Str("query", query).Any("args", args).Msg("built select query")
	return query, args, nil
}

// buildUpdateQuery dynamically builds UPDATE query
func (p *privateDataRepository) buildUpdateQuery(ctx context.Context, update models.PrivateDataUpdate) (string, []any, error) {
	userID, _ := utils.GetUserIDFromContext(ctx)

	qb := psql.Update("ciphers").
		Set("updated_at", sq.Expr("NOW()"))

	if update.Metadata != nil {
		qb = qb.Set("metadata", *update.Metadata)
	}
	if update.Type != nil {
		qb = qb.Set("type", *update.Type)
	}
	if update.Data != nil {
		qb = qb.Set("data", *update.Data)
	}
	if update.Notes != nil {
		qb = qb.Set("notes", *update.Notes)
	}
	if update.AdditionalFields != nil {
		qb = qb.Set("additional_fields", *update.AdditionalFields)
	}
	if update.Version != 0 {
		qb = qb.Set("version", update.Version)
	}

	qb = qb.Where(sq.Eq{"id": update.ID, "user_id": userID})

	query, args, err := qb.ToSql()
	if err != nil {
		return "", nil, err
	}

	logger.FromContext(ctx).Info().Str("query", query).Any("args", args).Msg("built update query")

	return query, args, nil
}
