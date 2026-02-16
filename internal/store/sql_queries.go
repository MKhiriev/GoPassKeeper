package store

import (
	"fmt"
	"strings"

	"github.com/MKhiriev/go-pass-keeper/models"
)

const (
	createUser = `INSERT INTO users (login, master_password, master_password_hint, name) 
    VALUES ($1, $2, $3, $4) 
    RETURNING user_id, login, master_password, master_password_hint, name, created_at;`

	findUserByLogin = `SELECT user_id, login, master_password, master_password_hint, name, created_at 
    FROM users 
    WHERE login = $1;`

	getAllUserPrivateData = `SELECT *
		FROM ciphers
		WHERE user_id = $1;`
	getRequestedPrivateData = `SELECT *
		FROM ciphers
		WHERE user_id = $1`
	getRequestedPrivateDataWhereID   = ` AND id = ANY($%d)`
	getRequestedPrivateDataWhereType = ` AND type = ANY($%d)`

	savePrivateData = `INSERT INTO ciphers (
			id, 
			user_id, 
			metadata, 
			type, 
			data, 
			notes, 
			additional_fields, 
			created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8);`

	deletePrivateData = `DELETE FROM ciphers
		WHERE user_id = $1 AND id = ANY($2);`

	updatePrivateDataBase = `
		UPDATE ciphers
		SET updated_at = NOW()`
	updatePrivateDataWhere = `
		WHERE id = $1 AND user_id = $2`
)

// buildUpdateQuery dynamically builds UPDATE query
func (p *privateDataRepository) buildUpdateQuery(update models.PrivateDataUpdate) (string, []any, error) {
	queryBuilder := new(strings.Builder)
	queryBuilder.WriteString(updatePrivateDataBase)

	args := make([]any, 0, 7)
	setClauses := make([]string, 0, 5)
	argIndex := 1

	// Добавляем поля для обновления
	if update.Metadata != nil {
		setClauses = append(setClauses, fmt.Sprintf("metadata = $%d", argIndex))
		args = append(args, *update.Metadata)
		argIndex++
	}

	if update.Type != nil {
		setClauses = append(setClauses, fmt.Sprintf("type = $%d", argIndex))
		args = append(args, *update.Type)
		argIndex++
	}

	if update.Data != nil {
		setClauses = append(setClauses, fmt.Sprintf("data = $%d", argIndex))
		args = append(args, *update.Data)
		argIndex++
	}

	if update.Notes != nil {
		setClauses = append(setClauses, fmt.Sprintf("notes = $%d", argIndex))
		args = append(args, *update.Notes)
		argIndex++
	}

	if update.AdditionalFields != nil {
		setClauses = append(setClauses, fmt.Sprintf("additional_fields = $%d", argIndex))
		args = append(args, *update.AdditionalFields)
		argIndex++
	}

	// Если есть поля для обновления, добавляем их в запрос
	if len(setClauses) > 0 {
		queryBuilder.WriteString(", ")
		queryBuilder.WriteString(strings.Join(setClauses, ", "))
	}

	// Добавляем WHERE условие
	queryBuilder.WriteString(" ")
	queryBuilder.WriteString(updatePrivateDataWhere)

	// Добавляем ID и UserID в аргументы
	args = append(args, update.ID, update.UserID)

	return queryBuilder.String(), args, nil
}
