package store

import (
	"context"
	"fmt"

	"github.com/MKhiriev/go-pass-keeper/internal/logger"
	"github.com/MKhiriev/go-pass-keeper/models"
	"github.com/jackc/pgerrcode"
)

// userRepository is the PostgreSQL-backed implementation of [UserRepository].
// It handles user account creation and lookup against the "users" table.
//
// All methods obtain a context-scoped logger via [logger.FromContext] for
// structured, request-level tracing of database interactions.
type userRepository struct {
	logger *logger.Logger
	db     *DB
}

// NewUserRepository constructs a [UserRepository] backed by the provided
// database connection and logger.
//
// A debug-level log message is emitted at construction time to aid
// application startup diagnostics.
func NewUserRepository(db *DB, logger *logger.Logger) UserRepository {
	logger.Debug().Msg("creating user repository")
	return &userRepository{
		db:     db,
		logger: logger,
	}
}

// CreateUser persists a new user record and returns the fully populated
// [models.User] with server-assigned fields (UserID, CreatedAt).
//
// The INSERT uses the [createUser] prepared query which returns all columns
// via a RETURNING clause, so the caller receives the canonical database
// representation of the newly created account.
//
// Error handling:
//   - PostgreSQL unique_violation (23505) → [ErrLoginAlreadyExists].
//   - Any other driver-level error → wrapped as "unexpected DB error".
//   - Scan failure → returned directly.
func (r *userRepository) CreateUser(ctx context.Context, user models.User) (models.User, error) {
	log := logger.FromContext(ctx)

	row := r.db.QueryRowContext(ctx, createUser, user.Login, user.AuthHash, user.MasterPasswordHint, user.Name, user.EncryptionSalt, user.EncryptedMasterKey)

	// create user in db
	if err := row.Err(); err != nil {
		log.Err(err).Str("func", "*userRepository.CreateUser").Msg("error: row is nil")

		switch postgresError(err) {
		case pgerrcode.UniqueViolation:
			return models.User{}, ErrLoginAlreadyExists
		default:
			return models.User{}, fmt.Errorf("unexpected DB error: %w", err)
		}
	}

	// scan saved user from db
	if err := row.Scan(&user.UserID, &user.Login, &user.AuthHash, &user.MasterPasswordHint, &user.Name, &user.CreatedAt, &user.EncryptionSalt, &user.EncryptedMasterKey); err != nil {
		log.Err(err).Str("func", "*userRepository.CreateUser").Msg("error: scanning error")
		return models.User{}, err
	}

	return user, nil
}

// FindUserByLogin retrieves a user record whose Login matches the one
// provided in the input [models.User].
//
// The lookup uses the [findUserByLogin] prepared query and scans all
// persisted fields into a fresh [models.User] instance.
//
// Error handling:
//   - PostgreSQL no_data_found (P0002) → [ErrNoUserWasFound].
//   - Any other driver-level error → wrapped as "unexpected DB error".
//   - Scan failure (including [sql.ErrNoRows]) → returned directly.
func (r *userRepository) FindUserByLogin(ctx context.Context, user models.User) (models.User, error) {
	log := logger.FromContext(ctx)

	var foundUser models.User
	row := r.db.QueryRowContext(ctx, findUserByLogin, user.Login)

	// find user by login
	if err := row.Err(); err != nil {
		log.Err(err).Str("func", "*userRepository.CreateUser").Msg("error: row is nil")
		switch postgresError(err) {
		case pgerrcode.NoDataFound:
			return models.User{}, ErrNoUserWasFound
		default:
			return models.User{}, fmt.Errorf("unexpected DB error: %w", err)
		}
	}

	// scan found user from db
	if err := row.Scan(&foundUser.UserID, &foundUser.Login, &foundUser.AuthHash, &foundUser.MasterPasswordHint, &foundUser.Name, &foundUser.CreatedAt, &foundUser.EncryptionSalt, &foundUser.EncryptedMasterKey); err != nil {
		log.Err(err).Str("func", "*userRepository.CreateUser").Msg("error: scanning error")
		return models.User{}, err
	}

	return foundUser, nil
}
