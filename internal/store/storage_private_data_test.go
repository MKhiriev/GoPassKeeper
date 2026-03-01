// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Rasul Khiriev

package store

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/MKhiriev/go-pass-keeper/internal/config"
	"github.com/MKhiriev/go-pass-keeper/internal/logger"
	"github.com/MKhiriev/go-pass-keeper/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ─────────────────────────────────────────────
// Mock: PrivateDataRepository
// ─────────────────────────────────────────────

type mockPrivateDataRepository struct {
	saveErr         error
	getResult       []models.PrivateData
	getErr          error
	getAllResult    []models.PrivateData
	getAllErr       error
	allStatesResult []models.PrivateDataState
	allStatesErr    error
	statesResult    []models.PrivateDataState
	statesErr       error
	updateErr       error
	deleteErr       error
}

func (m *mockPrivateDataRepository) SavePrivateData(_ context.Context, _ ...*models.PrivateData) error {
	return m.saveErr
}
func (m *mockPrivateDataRepository) GetPrivateData(_ context.Context, _ models.DownloadRequest) ([]models.PrivateData, error) {
	return m.getResult, m.getErr
}
func (m *mockPrivateDataRepository) GetAllPrivateData(_ context.Context, _ int64) ([]models.PrivateData, error) {
	return m.getAllResult, m.getAllErr
}
func (m *mockPrivateDataRepository) GetAllStates(_ context.Context, _ int64) ([]models.PrivateDataState, error) {
	return m.allStatesResult, m.allStatesErr
}
func (m *mockPrivateDataRepository) GetStates(_ context.Context, _ models.SyncRequest) ([]models.PrivateDataState, error) {
	return m.statesResult, m.statesErr
}
func (m *mockPrivateDataRepository) UpdatePrivateData(_ context.Context, _ models.UpdateRequest) error {
	return m.updateErr
}
func (m *mockPrivateDataRepository) DeletePrivateData(_ context.Context, _ models.DeleteRequest) error {
	return m.deleteErr
}

// ─────────────────────────────────────────────
// Helper
// ─────────────────────────────────────────────

func newStorageWithMock(repo *mockPrivateDataRepository) *privateDataStorage {
	return &privateDataStorage{
		repository: repo,
		logger:     logger.Nop(),
	}
}

// ─────────────────────────────────────────────
// NewPrivateDataStorage
// ─────────────────────────────────────────────

func TestNewPrivateDataStorage_WithoutFileStorage(t *testing.T) {
	log := logger.Nop()

	sqlDB, _ := newTestDB(t)
	db := newDBFromSQL(sqlDB)

	// Pass empty configuration.
	storage := NewPrivateDataStorage(db, config.Storage{}, log)

	// If this fails, NewPrivateDataStorage returned nil.
	require.NotNil(t, storage)

	s, ok := storage.(*privateDataStorage)
	require.True(t, ok)
	assert.Nil(t, s.fileStorage)
}

func TestNewPrivateDataStorage_WithFileStorage(t *testing.T) {
	sqlDB, _ := newTestDB(t)
	db := newDBFromSQL(sqlDB)
	cfg := config.Storage{Files: config.Files{BinaryDataDir: "/tmp/data"}}

	storage := NewPrivateDataStorage(db, cfg, logger.Nop())

	s, ok := storage.(*privateDataStorage)
	require.True(t, ok)
	assert.NotNil(t, s.fileStorage, "expected fileStorage to be set when BinaryDataDir is non-empty")
}

// ─────────────────────────────────────────────
// Save
// ─────────────────────────────────────────────

func TestSave_Success(t *testing.T) {
	s := newStorageWithMock(&mockPrivateDataRepository{})

	err := s.Save(context.Background(), &models.PrivateData{ClientSideID: "abc"})

	assert.NoError(t, err)
}

func TestSave_Error(t *testing.T) {
	expected := errors.New("save failed")
	s := newStorageWithMock(&mockPrivateDataRepository{saveErr: expected})

	err := s.Save(context.Background(), &models.PrivateData{})

	assert.ErrorIs(t, err, expected)
}

func TestSave_MultipleItems(t *testing.T) {
	s := newStorageWithMock(&mockPrivateDataRepository{})

	err := s.Save(context.Background(),
		&models.PrivateData{ClientSideID: "a"},
		&models.PrivateData{ClientSideID: "b"},
		&models.PrivateData{ClientSideID: "c"},
	)

	assert.NoError(t, err)
}

func TestSave_NoItems(t *testing.T) {
	s := newStorageWithMock(&mockPrivateDataRepository{})

	err := s.Save(context.Background())

	assert.NoError(t, err)
}

// ─────────────────────────────────────────────
// Get
// ─────────────────────────────────────────────

func TestGet_Success(t *testing.T) {
	now := time.Now()
	expected := []models.PrivateData{{ClientSideID: "x", UpdatedAt: &now}}
	s := newStorageWithMock(&mockPrivateDataRepository{getResult: expected})

	result, err := s.Get(context.Background(), models.DownloadRequest{UserID: 1})

	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.Equal(t, "x", result[0].ClientSideID)
}

func TestGet_Error(t *testing.T) {
	expected := errors.New("get failed")
	s := newStorageWithMock(&mockPrivateDataRepository{getErr: expected})

	_, err := s.Get(context.Background(), models.DownloadRequest{})

	assert.ErrorIs(t, err, expected)
}

func TestGet_EmptyResult(t *testing.T) {
	s := newStorageWithMock(&mockPrivateDataRepository{getResult: []models.PrivateData{}})

	result, err := s.Get(context.Background(), models.DownloadRequest{UserID: 1})

	require.NoError(t, err)
	assert.Empty(t, result)
}

// ─────────────────────────────────────────────
// GetAll
// ─────────────────────────────────────────────

func TestGetAll_Success(t *testing.T) {
	expected := []models.PrivateData{{ClientSideID: "y"}, {ClientSideID: "z"}}
	s := newStorageWithMock(&mockPrivateDataRepository{getAllResult: expected})

	result, err := s.GetAll(context.Background(), 42)

	require.NoError(t, err)
	assert.Len(t, result, 2)
}

func TestGetAll_Error(t *testing.T) {
	expected := errors.New("getall failed")
	s := newStorageWithMock(&mockPrivateDataRepository{getAllErr: expected})

	_, err := s.GetAll(context.Background(), 1)

	assert.ErrorIs(t, err, expected)
}

func TestGetAll_Empty(t *testing.T) {
	s := newStorageWithMock(&mockPrivateDataRepository{getAllResult: []models.PrivateData{}})

	result, err := s.GetAll(context.Background(), 99)

	require.NoError(t, err)
	assert.Empty(t, result)
}

// ─────────────────────────────────────────────
// GetAllStates
// ─────────────────────────────────────────────

func TestGetAllStates_Success(t *testing.T) {
	now := time.Now()
	expected := []models.PrivateDataState{
		{ClientSideID: "a", Hash: "h1", Version: 1, UpdatedAt: &now},
	}
	s := newStorageWithMock(&mockPrivateDataRepository{allStatesResult: expected})

	result, err := s.GetAllStates(context.Background(), 1)

	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.Equal(t, "a", result[0].ClientSideID)
	assert.Equal(t, int64(1), result[0].Version)
}

func TestGetAllStates_Error(t *testing.T) {
	expected := errors.New("states failed")
	s := newStorageWithMock(&mockPrivateDataRepository{allStatesErr: expected})

	_, err := s.GetAllStates(context.Background(), 1)

	assert.ErrorIs(t, err, expected)
}

func TestGetAllStates_Empty(t *testing.T) {
	s := newStorageWithMock(&mockPrivateDataRepository{allStatesResult: []models.PrivateDataState{}})

	result, err := s.GetAllStates(context.Background(), 1)

	require.NoError(t, err)
	assert.Empty(t, result)
}

// ─────────────────────────────────────────────
// GetStates
// ─────────────────────────────────────────────

func TestGetStates_Success(t *testing.T) {
	expected := []models.PrivateDataState{{ClientSideID: "b", Version: 3}}
	s := newStorageWithMock(&mockPrivateDataRepository{statesResult: expected})

	result, err := s.GetStates(context.Background(), models.SyncRequest{
		UserID:        1,
		ClientSideIDs: []string{"b"},
	})

	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.Equal(t, "b", result[0].ClientSideID)
}

func TestGetStates_Error(t *testing.T) {
	expected := errors.New("getstates failed")
	s := newStorageWithMock(&mockPrivateDataRepository{statesErr: expected})

	_, err := s.GetStates(context.Background(), models.SyncRequest{})

	assert.ErrorIs(t, err, expected)
}

func TestGetStates_EmptyIDs(t *testing.T) {
	s := newStorageWithMock(&mockPrivateDataRepository{statesResult: []models.PrivateDataState{}})

	result, err := s.GetStates(context.Background(), models.SyncRequest{UserID: 1, ClientSideIDs: []string{}})

	require.NoError(t, err)
	assert.Empty(t, result)
}

// ─────────────────────────────────────────────
// Update
// ─────────────────────────────────────────────

func TestUpdate_Success(t *testing.T) {
	s := newStorageWithMock(&mockPrivateDataRepository{})

	err := s.Update(context.Background(), models.UpdateRequest{UserID: 1})

	assert.NoError(t, err)
}

func TestUpdate_Error(t *testing.T) {
	expected := errors.New("update failed")
	s := newStorageWithMock(&mockPrivateDataRepository{updateErr: expected})

	err := s.Update(context.Background(), models.UpdateRequest{})

	assert.ErrorIs(t, err, expected)
}

// ─────────────────────────────────────────────
// Delete
// ─────────────────────────────────────────────

func TestDelete_Success(t *testing.T) {
	s := newStorageWithMock(&mockPrivateDataRepository{})

	err := s.Delete(context.Background(), models.DeleteRequest{UserID: 1})

	assert.NoError(t, err)
}

func TestDelete_Error(t *testing.T) {
	expected := errors.New("delete failed")
	s := newStorageWithMock(&mockPrivateDataRepository{deleteErr: expected})

	err := s.Delete(context.Background(), models.DeleteRequest{})

	assert.ErrorIs(t, err, expected)
}

func TestDelete_EmptyEntries(t *testing.T) {
	s := newStorageWithMock(&mockPrivateDataRepository{})

	err := s.Delete(context.Background(), models.DeleteRequest{
		UserID:        1,
		DeleteEntries: []models.DeleteEntry{},
	})

	assert.NoError(t, err)
}
