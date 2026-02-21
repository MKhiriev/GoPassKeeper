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

	getAllUserPrivateData = `SELECT *
		FROM ciphers
		WHERE user_id = $1;`
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
		return "", nil, fmt.Errorf("error building query for updating user data: %w", err)
	}

	logger.FromContext(ctx).Info().Str("query", query).Any("args", args).Msg("built update query")

	return query, args, nil
}

// buildGetPrivateDataQuery builds SELECT query with optional ID and Type filters
func (p *privateDataRepository) buildGetPrivateDataQuery(ctx context.Context, req models.DownloadRequest) (string, []any, error) {
	qb := psql.Select("*").From("ciphers").Where(sq.Eq{"user_id": req.UserID})

	if len(req.IDs) > 0 {
		qb = qb.Where(sq.Eq{"id": req.IDs})
	}
	if len(req.Types) > 0 {
		qb = qb.Where(sq.Eq{"type": req.Types})
	}

	query, args, err := qb.ToSql()
	if err != nil {
		return "", nil, fmt.Errorf("error building query for getting private data with filters: %w", err)
	}

	logger.FromContext(ctx).Info().Str("query", query).Any("args", args).Msg("built get private data query")
	return query, args, nil
}

// buildDeletePrivateDataQuery builds DELETE query for specific IDs
func (p *privateDataRepository) buildDeletePrivateDataQuery(ctx context.Context, req models.DeleteRequest) (string, []any, error) {
	qb := psql.Delete("ciphers").Where(sq.Eq{"user_id": req.UserID, "id": req.IDs})

	query, args, err := qb.ToSql()
	if err != nil {
		return "", nil, fmt.Errorf("error building delete query: %w", err)
	}

	logger.FromContext(ctx).Info().Str("query", query).Any("args", args).Msg("built delete query")
	return query, args, nil
}

// buildCreateUserQuery builds INSERT query with RETURNING clause
func (r *userRepository) buildCreateUserQuery(ctx context.Context, user models.User) (string, []any, error) {
	qb := psql.Insert("users").
		Columns("login", "master_password", "master_password_hint", "name").
		Values(user.Login, user.MasterPassword, user.MasterPasswordHint, user.Name).
		Suffix("RETURNING user_id, login, master_password, master_password_hint, name, created_at")

	query, args, err := qb.ToSql()
	if err != nil {
		return "", nil, fmt.Errorf("error building create user query: %w", err)
	}

	logger.FromContext(ctx).Info().Str("query", query).Any("args", args).Msg("built create user query")
	return query, args, nil
}

// buildFindUserByLoginQuery builds SELECT query for finding user by login
func (r *userRepository) buildFindUserByLoginQuery(ctx context.Context, login string) (string, []any, error) {
	qb := psql.Select("user_id", "login", "master_password", "master_password_hint", "name", "created_at").
		From("users").
		Where(sq.Eq{"login": login})

	query, args, err := qb.ToSql()
	if err != nil {
		return "", nil, fmt.Errorf("error building find user by login query: %w", err)
	}

	logger.FromContext(ctx).Info().Str("query", query).Any("args", args).Msg("built find user by login query")
	return query, args, nil
}
