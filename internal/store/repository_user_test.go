package store

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/MKhiriev/go-pass-keeper/internal/logger"
	"github.com/MKhiriev/go-pass-keeper/models"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

func newTestUserRepo(t *testing.T) (*userRepository, sqlmock.Sqlmock, *sql.DB) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	l := logger.NewLogger("test")
	repo := &userRepository{
		db:     &DB{DB: db, logger: l},
		logger: l,
	}
	return repo, mock, db
}

func pgError(code string) error {
	return &pgconn.PgError{Code: code}
}

func TestCreateUser_Success(t *testing.T) {
	repo, mock, db := newTestUserRepo(t)
	defer db.Close()

	ctx := context.Background()
	user := models.User{
		Login:              "john",
		MasterPassword:     "hash",
		MasterPasswordHint: "hint",
		Name:               "John",
	}

	now := time.Now()

	rows := sqlmock.
		NewRows([]string{"user_id", "login", "auth_hash", "master_password_hint", "name", "created_at"}).
		AddRow(1, user.Login, user.MasterPassword, user.MasterPasswordHint, user.Name, now)

	mock.ExpectQuery("INSERT INTO users").
		WithArgs(user.Login, user.MasterPassword, user.MasterPasswordHint, user.Name).
		WillReturnRows(rows)

	created, err := repo.CreateUser(ctx, user)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.UserID != 1 {
		t.Errorf("expected UserID=1, got %d", created.UserID)
	}
	if created.Login != user.Login {
		t.Errorf("expected login %s, got %s", user.Login, created.Login)
	}
}

func TestCreateUser_UniqueViolation(t *testing.T) {
	repo, mock, db := newTestUserRepo(t)
	defer db.Close()

	ctx := context.Background()
	user := models.User{Login: "john"}

	mock.ExpectQuery("INSERT INTO users").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(pgError(pgerrcode.UniqueViolation))

	_, err := repo.CreateUser(ctx, user)
	if !errors.Is(err, ErrLoginAlreadyExists) {
		t.Fatalf("expected ErrLoginAlreadyExists, got %v", err)
	}
}

func TestCreateUser_UnexpectedDBError(t *testing.T) {
	repo, mock, db := newTestUserRepo(t)
	defer db.Close()

	ctx := context.Background()
	user := models.User{Login: "john"}

	mock.ExpectQuery("INSERT INTO users").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(errors.New("db network error"))

	_, err := repo.CreateUser(ctx, user)
	if err == nil || !strings.Contains(err.Error(), "unexpected DB error") {
		t.Fatalf("expected wrapped unexpected DB error, got %v", err)
	}
}

func TestCreateUser_ScanError(t *testing.T) {
	repo, mock, db := newTestUserRepo(t)
	defer db.Close()

	ctx := context.Background()
	user := models.User{Login: "john"}

	rows := sqlmock.
		NewRows([]string{"user_id"}). // intentionally wrong shape → scan error
		AddRow(1)

	mock.ExpectQuery("INSERT INTO users").
		WillReturnRows(rows)

	_, err := repo.CreateUser(ctx, user)
	if err == nil {
		t.Fatal("expected scan error, got nil")
	}
}

func TestFindUserByLogin_Success(t *testing.T) {
	repo, mock, db := newTestUserRepo(t)
	defer db.Close()

	ctx := context.Background()
	user := models.User{Login: "john"}

	now := time.Now()
	rows := sqlmock.
		NewRows([]string{"user_id", "login", "auth_hash", "master_password_hint", "name", "created_at"}).
		AddRow(1, "john", "hash", "hint", "John", now)

	mock.ExpectQuery("SELECT user_id").
		WithArgs("john").
		WillReturnRows(rows)

	found, err := repo.FindUserByLogin(ctx, user)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found.Login != "john" {
		t.Errorf("expected login john, got %s", found.Login)
	}
}

func TestFindUserByLogin_NotFound(t *testing.T) {
	repo, mock, db := newTestUserRepo(t)
	defer db.Close()

	ctx := context.Background()
	user := models.User{Login: "john"}

	mock.ExpectQuery("SELECT user_id").
		WithArgs("john").
		WillReturnError(pgError(pgerrcode.NoDataFound))

	_, err := repo.FindUserByLogin(ctx, user)
	if !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}

func TestFindUserByLogin_UnexpectedError(t *testing.T) {
	repo, mock, db := newTestUserRepo(t)
	defer db.Close()

	ctx := context.Background()
	user := models.User{Login: "john"}

	mock.ExpectQuery("SELECT user_id").
		WithArgs("john").
		WillReturnError(errors.New("db failure"))

	_, err := repo.FindUserByLogin(ctx, user)
	if err == nil || !strings.Contains(err.Error(), "unexpected DB error") {
		t.Fatalf("expected wrapped unexpected DB error, got %v", err)
	}
}

func TestFindUserByLogin_ScanError(t *testing.T) {
	repo, mock, db := newTestUserRepo(t)
	defer db.Close()

	ctx := context.Background()
	user := models.User{Login: "john"}

	rows := sqlmock.NewRows([]string{"user_id"}).AddRow(1)

	mock.ExpectQuery("SELECT user_id").
		WithArgs("john").
		WillReturnRows(rows)

	_, err := repo.FindUserByLogin(ctx, user)
	if err == nil {
		t.Fatal("expected scan error, got nil")
	}
}
