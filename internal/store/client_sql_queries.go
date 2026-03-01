// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Rasul Khiriev

package store

const (
	saveSinglePrivateData = `
		INSERT INTO ciphers (
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
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12);`

	getSinglePrivateData = `
		SELECT
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
		WHERE user_id = $1 AND client_side_id = $2;`

	getAllPrivateData = `
		SELECT
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
		WHERE user_id = $1 AND deleted=false;`

	getAllStates = `
		SELECT
			version,
			client_side_id,
			hash,
			deleted,
			updated_at
		FROM ciphers
		WHERE user_id = $1;`

	updatePrivateData = `
		UPDATE ciphers SET
			type              = $1,
			metadata          = $2,
			data              = $3,
			notes             = $4,
			additional_fields = $5,
			updated_at        = $6,
			version           = $7,
			hash              = $8,
			deleted           = $9
		WHERE user_id = $10 AND client_side_id = $11;`

	deletePrivateData = `
		UPDATE ciphers SET
			deleted    = true,
			updated_at = CURRENT_TIMESTAMP
		WHERE user_id = $1 AND client_side_id = $2;`

	incrementVersion = `
		UPDATE ciphers
		SET version = version + 1
		WHERE client_side_id = $1
		  AND user_id = $2;`
)
