// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Rasul Khiriev

package migrations

import (
	"database/sql"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestMigrate_DBError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	_ = mock // не используем напрямую, goose сам будет ходить в DB

	err = Migrate(db)
	if err == nil {
		t.Fatal("expected error from Migrate, got nil")
	}

	if !strings.Contains(err.Error(), "migration error") {
		t.Errorf("expected wrapped migration error, got: %v", err)
	}
}

func TestMigrate_NilDB(t *testing.T) {
	var db *sql.DB

	err := Migrate(db)
	if err == nil {
		t.Fatal("expected error when db is nil, got nil")
	}

	if !strings.Contains(err.Error(), "db is nil") {
		t.Errorf("expected 'db is nil' error, got: %v", err)
	}
}
