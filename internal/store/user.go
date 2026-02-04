package store

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/MKhiriev/go-pass-keeper/internal/logger"
	"github.com/MKhiriev/go-pass-keeper/models"
	"github.com/jackc/pgerrcode"
)

type userRepository struct {
	logger *logger.Logger
	db     *sql.DB
}

func NewUserRepository(db *sql.DB, logger *logger.Logger) UserRepository {
	logger.Debug().Msg("UserRepository created")
	return &userRepository{
		db:     db,
		logger: logger,
	}
}

func (r *userRepository) CreateUser(ctx context.Context, user models.User) (models.User, error) {
	row := r.db.QueryRowContext(ctx, createUser, user.Login, user.Password)

	// create user in db
	if err := row.Err(); err != nil {
		r.logger.Err(err).Str("func", "*userRepository.CreateUser").Msg("error: row is nil")

		switch postgresError(err) {
		case pgerrcode.UniqueViolation:
			return models.User{}, ErrLoginAlreadyExists
		default:
			return models.User{}, fmt.Errorf("unexpected DB error: %w", err)
		}
	}

	// scan saved user from db
	if err := row.Scan(&user.UserID, &user.Login, &user.Password); err != nil {
		r.logger.Err(err).Str("func", "*userRepository.CreateUser").Msg("error: scanning error")
		return models.User{}, err
	}

	return user, nil
}

func (r *userRepository) FindUserByLogin(ctx context.Context, user models.User) (models.User, error) {
	var foundUser models.User
	row := r.db.QueryRowContext(ctx, findUserByLogin, user.Login)

	// find user by login
	if err := row.Err(); err != nil {
		r.logger.Err(err).Str("func", "*userRepository.CreateUser").Msg("error: row is nil")
		switch postgresError(err) {
		case pgerrcode.NoDataFound:
			return models.User{}, ErrUserNotFound
		default:
			return models.User{}, fmt.Errorf("unexpected DB error: %w", err)
		}
	}

	// scan found user from db
	if err := row.Scan(&foundUser.UserID, &foundUser.Login, &foundUser.Password); err != nil {
		r.logger.Err(err).Str("func", "*userRepository.CreateUser").Msg("error: scanning error")
		return models.User{}, err
	}

	return foundUser, nil
}
