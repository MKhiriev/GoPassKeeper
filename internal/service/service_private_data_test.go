// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Rasul Khiriev

package service

import (
	"context"
	"errors"
	"testing"

	"github.com/MKhiriev/go-pass-keeper/internal/logger"
	"github.com/MKhiriev/go-pass-keeper/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ─────────────────────────────────────────────
// Mock: store.PrivateDataStorage
// ─────────────────────────────────────────────

type mockPrivateDataStorage struct {
	saveFn         func(ctx context.Context, data ...*models.PrivateData) error
	getFn          func(ctx context.Context, req models.DownloadRequest) ([]models.PrivateData, error)
	getAllFn       func(ctx context.Context, userID int64) ([]models.PrivateData, error)
	getAllStatesFn func(ctx context.Context, userID int64) ([]models.PrivateDataState, error)
	getStatesFn    func(ctx context.Context, req models.SyncRequest) ([]models.PrivateDataState, error)
	updateFn       func(ctx context.Context, req models.UpdateRequest) error
	deleteFn       func(ctx context.Context, req models.DeleteRequest) error
}

func (m *mockPrivateDataStorage) Save(ctx context.Context, data ...*models.PrivateData) error {
	if m.saveFn != nil {
		return m.saveFn(ctx, data...)
	}
	return nil
}

func (m *mockPrivateDataStorage) Get(ctx context.Context, req models.DownloadRequest) ([]models.PrivateData, error) {
	if m.getFn != nil {
		return m.getFn(ctx, req)
	}
	return nil, nil
}

func (m *mockPrivateDataStorage) GetAll(ctx context.Context, userID int64) ([]models.PrivateData, error) {
	if m.getAllFn != nil {
		return m.getAllFn(ctx, userID)
	}
	return nil, nil
}

func (m *mockPrivateDataStorage) GetAllStates(ctx context.Context, userID int64) ([]models.PrivateDataState, error) {
	if m.getAllStatesFn != nil {
		return m.getAllStatesFn(ctx, userID)
	}
	return nil, nil
}

func (m *mockPrivateDataStorage) GetStates(ctx context.Context, req models.SyncRequest) ([]models.PrivateDataState, error) {
	if m.getStatesFn != nil {
		return m.getStatesFn(ctx, req)
	}
	return nil, nil
}

func (m *mockPrivateDataStorage) Update(ctx context.Context, req models.UpdateRequest) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, req)
	}
	return nil
}

func (m *mockPrivateDataStorage) Delete(ctx context.Context, req models.DeleteRequest) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, req)
	}
	return nil
}

// ─────────────────────────────────────────────
// Helper
// ─────────────────────────────────────────────

// newRawPrivateDataService bypasses the validation wrapper and returns the
// bare *privateDataService so we can test delegation in isolation.
func newRawPrivateDataService(storage *mockPrivateDataStorage) *privateDataService {
	return &privateDataService{
		privateDataRepository: storage,
		logger:                logger.Nop(),
	}
}

var errStorage = errors.New("storage error")

// ─────────────────────────────────────────────
// UploadPrivateData
// ─────────────────────────────────────────────

func TestPrivateDataService_UploadPrivateData_Success(t *testing.T) {
	items := []*models.PrivateData{{ClientSideID: "a"}, {ClientSideID: "b"}}
	storage := &mockPrivateDataStorage{
		saveFn: func(_ context.Context, data ...*models.PrivateData) error {
			assert.Equal(t, items, data)
			return nil
		},
	}
	svc := newRawPrivateDataService(storage)

	err := svc.UploadPrivateData(context.Background(), models.UploadRequest{
		PrivateDataList: items,
	})

	require.NoError(t, err)
}

func TestPrivateDataService_UploadPrivateData_StorageError(t *testing.T) {
	storage := &mockPrivateDataStorage{
		saveFn: func(_ context.Context, _ ...*models.PrivateData) error {
			return errStorage
		},
	}
	svc := newRawPrivateDataService(storage)

	err := svc.UploadPrivateData(context.Background(), models.UploadRequest{
		PrivateDataList: []*models.PrivateData{{ClientSideID: "x"}},
	})

	require.ErrorIs(t, err, errStorage)
}

func TestPrivateDataService_UploadPrivateData_EmptyList_DelegatesToStorage(t *testing.T) {
	called := false
	storage := &mockPrivateDataStorage{
		saveFn: func(_ context.Context, _ ...*models.PrivateData) error {
			called = true
			return nil
		},
	}
	svc := newRawPrivateDataService(storage)

	// privateDataService itself does NOT validate — that is the validation layer's job.
	err := svc.UploadPrivateData(context.Background(), models.UploadRequest{})

	require.NoError(t, err)
	assert.True(t, called, "Save must be called even for empty list — validation is not this layer's concern")
}

// ─────────────────────────────────────────────
// DownloadPrivateData
// ─────────────────────────────────────────────

func TestPrivateDataService_DownloadPrivateData_Success(t *testing.T) {
	expected := []models.PrivateData{{ClientSideID: "id-1"}}
	req := models.DownloadRequest{UserID: 1, ClientSideIDs: []string{"id-1"}}

	storage := &mockPrivateDataStorage{
		getFn: func(_ context.Context, r models.DownloadRequest) ([]models.PrivateData, error) {
			assert.Equal(t, req, r)
			return expected, nil
		},
	}
	svc := newRawPrivateDataService(storage)

	result, err := svc.DownloadPrivateData(context.Background(), req)

	require.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestPrivateDataService_DownloadPrivateData_StorageError(t *testing.T) {
	storage := &mockPrivateDataStorage{
		getFn: func(_ context.Context, _ models.DownloadRequest) ([]models.PrivateData, error) {
			return nil, errStorage
		},
	}
	svc := newRawPrivateDataService(storage)

	result, err := svc.DownloadPrivateData(context.Background(), models.DownloadRequest{UserID: 1})

	assert.Nil(t, result)
	require.ErrorIs(t, err, errStorage)
}

// ─────────────────────────────────────────────
// DownloadAllPrivateData
// ─────────────────────────────────────────────

func TestPrivateDataService_DownloadAllPrivateData_Success(t *testing.T) {
	expected := []models.PrivateData{{ClientSideID: "all-1"}}
	storage := &mockPrivateDataStorage{
		getAllFn: func(_ context.Context, userID int64) ([]models.PrivateData, error) {
			assert.Equal(t, int64(7), userID)
			return expected, nil
		},
	}
	svc := newRawPrivateDataService(storage)

	result, err := svc.DownloadAllPrivateData(context.Background(), 7)

	require.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestPrivateDataService_DownloadAllPrivateData_StorageError(t *testing.T) {
	storage := &mockPrivateDataStorage{
		getAllFn: func(_ context.Context, _ int64) ([]models.PrivateData, error) {
			return nil, errStorage
		},
	}
	svc := newRawPrivateDataService(storage)

	result, err := svc.DownloadAllPrivateData(context.Background(), 1)

	assert.Nil(t, result)
	require.ErrorIs(t, err, errStorage)
}

// ─────────────────────────────────────────────
// DownloadUserPrivateDataStates
// ─────────────────────────────────────────────

func TestPrivateDataService_DownloadUserPrivateDataStates_Success(t *testing.T) {
	expected := []models.PrivateDataState{{ClientSideID: "s-1"}}
	storage := &mockPrivateDataStorage{
		getAllStatesFn: func(_ context.Context, userID int64) ([]models.PrivateDataState, error) {
			assert.Equal(t, int64(3), userID)
			return expected, nil
		},
	}
	svc := newRawPrivateDataService(storage)

	result, err := svc.DownloadUserPrivateDataStates(context.Background(), 3)

	require.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestPrivateDataService_DownloadUserPrivateDataStates_StorageError(t *testing.T) {
	storage := &mockPrivateDataStorage{
		getAllStatesFn: func(_ context.Context, _ int64) ([]models.PrivateDataState, error) {
			return nil, errStorage
		},
	}
	svc := newRawPrivateDataService(storage)

	result, err := svc.DownloadUserPrivateDataStates(context.Background(), 1)

	assert.Nil(t, result)
	require.ErrorIs(t, err, errStorage)
}

// ─────────────────────────────────────────────
// DownloadSpecificUserPrivateDataStates
// ─────────────────────────────────────────────

func TestPrivateDataService_DownloadSpecificUserPrivateDataStates_Success(t *testing.T) {
	expected := []models.PrivateDataState{{ClientSideID: "spec-1"}}
	req := models.SyncRequest{UserID: 2, ClientSideIDs: []string{"spec-1"}}

	storage := &mockPrivateDataStorage{
		getStatesFn: func(_ context.Context, r models.SyncRequest) ([]models.PrivateDataState, error) {
			assert.Equal(t, req, r)
			return expected, nil
		},
	}
	svc := newRawPrivateDataService(storage)

	result, err := svc.DownloadSpecificUserPrivateDataStates(context.Background(), req)

	require.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestPrivateDataService_DownloadSpecificUserPrivateDataStates_StorageError(t *testing.T) {
	storage := &mockPrivateDataStorage{
		getStatesFn: func(_ context.Context, _ models.SyncRequest) ([]models.PrivateDataState, error) {
			return nil, errStorage
		},
	}
	svc := newRawPrivateDataService(storage)

	result, err := svc.DownloadSpecificUserPrivateDataStates(context.Background(), models.SyncRequest{UserID: 1})

	assert.Nil(t, result)
	require.ErrorIs(t, err, errStorage)
}

// ─────────────────────────────────────────────
// UpdatePrivateData
// ─────────────────────────────────────────────

func TestPrivateDataService_UpdatePrivateData_Success(t *testing.T) {
	req := models.UpdateRequest{
		UserID: 5,
		PrivateDataUpdates: []models.PrivateDataUpdate{
			{ClientSideID: "upd-1", Version: 1},
		},
		Length: 1,
	}
	storage := &mockPrivateDataStorage{
		updateFn: func(_ context.Context, r models.UpdateRequest) error {
			assert.Equal(t, req, r)
			return nil
		},
	}
	svc := newRawPrivateDataService(storage)

	err := svc.UpdatePrivateData(context.Background(), req)

	require.NoError(t, err)
}

func TestPrivateDataService_UpdatePrivateData_StorageError(t *testing.T) {
	storage := &mockPrivateDataStorage{
		updateFn: func(_ context.Context, _ models.UpdateRequest) error {
			return errStorage
		},
	}
	svc := newRawPrivateDataService(storage)

	err := svc.UpdatePrivateData(context.Background(), models.UpdateRequest{UserID: 1})

	require.ErrorIs(t, err, errStorage)
}

// ─────────────────────────────────────────────
// DeletePrivateData
// ─────────────────────────────────────────────

func TestPrivateDataService_DeletePrivateData_Success(t *testing.T) {
	req := models.DeleteRequest{
		UserID: 9,
		DeleteEntries: []models.DeleteEntry{
			{ClientSideID: "del-1", Version: 1},
		},
		Length: 1,
	}
	storage := &mockPrivateDataStorage{
		deleteFn: func(_ context.Context, r models.DeleteRequest) error {
			assert.Equal(t, req, r)
			return nil
		},
	}
	svc := newRawPrivateDataService(storage)

	err := svc.DeletePrivateData(context.Background(), req)

	require.NoError(t, err)
}

func TestPrivateDataService_DeletePrivateData_StorageError(t *testing.T) {
	storage := &mockPrivateDataStorage{
		deleteFn: func(_ context.Context, _ models.DeleteRequest) error {
			return errStorage
		},
	}
	svc := newRawPrivateDataService(storage)

	err := svc.DeletePrivateData(context.Background(), models.DeleteRequest{UserID: 1})

	require.ErrorIs(t, err, errStorage)
}
