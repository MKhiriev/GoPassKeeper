// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Rasul Khiriev

package service

import (
	"context"
	"errors"
	"testing"

	"github.com/MKhiriev/go-pass-keeper/internal/mock"
	"github.com/MKhiriev/go-pass-keeper/internal/store"
	"github.com/MKhiriev/go-pass-keeper/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// newTestPrivateDataSvc — хелпер для создания сервиса с моками
func newTestPrivateDataSvc(
	t *testing.T,
	ctrl *gomock.Controller,
) (
	ClientPrivateDataService,
	*mock.MockLocalPrivateDataRepository,
	*mock.MockServerAdapter,
	*mock.MockClientCryptoService,
) {
	t.Helper()
	mockRepo := mock.NewMockLocalPrivateDataRepository(ctrl)
	mockAdapter := mock.NewMockServerAdapter(ctrl)
	mockCrypto := mock.NewMockClientCryptoService(ctrl)

	storages := &store.ClientStorages{
		PrivateDataRepository: mockRepo,
	}
	svc := NewClientPrivateDataService(storages, mockAdapter, mockCrypto)
	return svc, mockRepo, mockAdapter, mockCrypto
}

// ── Create ───────────────────────────────────────────────────────────────────

func TestClientPrivateDataService_Create_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, mockRepo, mockAdapter, mockCrypto := newTestPrivateDataSvc(t, ctrl)
	ctx := context.Background()
	userID := int64(1)

	plain := models.DecipheredPayload{
		UserID:   userID,
		Metadata: models.Metadata{Name: "test"},
	}
	encPayload := models.PrivateDataPayload{}

	mockCrypto.EXPECT().EncryptPayload(plain).Return(encPayload, nil)
	mockCrypto.EXPECT().ComputeHash(encPayload).Return("hash123", nil)
	mockRepo.EXPECT().SavePrivateData(ctx, userID, gomock.Any()).Return(nil)
	mockAdapter.EXPECT().Upload(ctx, gomock.Any()).Return(nil)

	err := svc.Create(ctx, userID, plain)
	require.NoError(t, err)
}

func TestClientPrivateDataService_Create_EncryptError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, _, _, mockCrypto := newTestPrivateDataSvc(t, ctrl)
	ctx := context.Background()
	plain := models.DecipheredPayload{UserID: 1}

	mockCrypto.EXPECT().EncryptPayload(plain).Return(models.PrivateDataPayload{}, errors.New("aes fail"))

	err := svc.Create(ctx, 1, plain)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "encrypt payload for create")
}

func TestClientPrivateDataService_Create_HashError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, _, _, mockCrypto := newTestPrivateDataSvc(t, ctrl)
	ctx := context.Background()
	plain := models.DecipheredPayload{UserID: 1}
	encPayload := models.PrivateDataPayload{}

	mockCrypto.EXPECT().EncryptPayload(plain).Return(encPayload, nil)
	mockCrypto.EXPECT().ComputeHash(encPayload).Return("", errors.New("hash fail"))

	err := svc.Create(ctx, 1, plain)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "compute hash")
}

func TestClientPrivateDataService_Create_SaveLocalError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, mockRepo, _, mockCrypto := newTestPrivateDataSvc(t, ctrl)
	ctx := context.Background()
	userID := int64(1)
	plain := models.DecipheredPayload{UserID: userID}
	encPayload := models.PrivateDataPayload{}

	mockCrypto.EXPECT().EncryptPayload(plain).Return(encPayload, nil)
	mockCrypto.EXPECT().ComputeHash(encPayload).Return("hash", nil)
	mockRepo.EXPECT().SavePrivateData(ctx, userID, gomock.Any()).Return(errors.New("db error"))

	err := svc.Create(ctx, userID, plain)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "save created item to local store")
}

func TestClientPrivateDataService_Create_UploadError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, mockRepo, mockAdapter, mockCrypto := newTestPrivateDataSvc(t, ctrl)
	ctx := context.Background()
	userID := int64(1)
	plain := models.DecipheredPayload{UserID: userID}
	encPayload := models.PrivateDataPayload{}

	mockCrypto.EXPECT().EncryptPayload(plain).Return(encPayload, nil)
	mockCrypto.EXPECT().ComputeHash(encPayload).Return("hash", nil)
	mockRepo.EXPECT().SavePrivateData(ctx, userID, gomock.Any()).Return(nil)
	mockAdapter.EXPECT().Upload(ctx, gomock.Any()).Return(errors.New("network error"))

	// IncrementVersion НЕ должен вызываться при ошибке Upload
	err := svc.Create(ctx, userID, plain)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "upload created item to server")
}

// ── GetAll ───────────────────────────────────────────────────────────────────

func TestClientPrivateDataService_GetAll_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, mockRepo, _, mockCrypto := newTestPrivateDataSvc(t, ctrl)
	ctx := context.Background()
	userID := int64(1)

	encPayload := models.PrivateDataPayload{}
	items := []models.PrivateData{
		{ClientSideID: "id1", UserID: userID, Payload: encPayload},
		{ClientSideID: "id2", UserID: userID, Payload: encPayload},
	}
	decrypted := models.DecipheredPayload{ClientSideID: "id1", UserID: userID}

	mockRepo.EXPECT().GetAllPrivateData(ctx, userID).Return(items, nil)
	mockCrypto.EXPECT().DecryptPayload(encPayload).Return(decrypted, nil).Times(2)

	got, err := svc.GetAll(ctx, userID)
	require.NoError(t, err)
	assert.Len(t, got, 2)
}

func TestClientPrivateDataService_GetAll_RepoError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, mockRepo, _, _ := newTestPrivateDataSvc(t, ctrl)
	ctx := context.Background()

	mockRepo.EXPECT().GetAllPrivateData(ctx, int64(1)).Return(nil, errors.New("db error"))

	_, err := svc.GetAll(ctx, 1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "get all local items")
}

func TestClientPrivateDataService_GetAll_DecryptError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, mockRepo, _, mockCrypto := newTestPrivateDataSvc(t, ctrl)
	ctx := context.Background()
	userID := int64(1)
	encPayload := models.PrivateDataPayload{}

	mockRepo.EXPECT().GetAllPrivateData(ctx, userID).Return([]models.PrivateData{
		{ClientSideID: "id1", Payload: encPayload},
	}, nil)
	mockCrypto.EXPECT().DecryptPayload(encPayload).Return(models.DecipheredPayload{}, errors.New("decrypt fail"))

	_, err := svc.GetAll(ctx, userID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "decrypt item id1")
}

// ── Get ──────────────────────────────────────────────────────────────────────

func TestClientPrivateDataService_Get_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, mockRepo, _, mockCrypto := newTestPrivateDataSvc(t, ctrl)
	ctx := context.Background()
	userID := int64(1)
	clientSideID := "id1"
	encPayload := models.PrivateDataPayload{}
	want := models.DecipheredPayload{ClientSideID: clientSideID, UserID: userID}

	mockRepo.EXPECT().GetPrivateData(ctx, clientSideID, userID).Return(
		models.PrivateData{ClientSideID: clientSideID, Payload: encPayload}, nil,
	)
	mockCrypto.EXPECT().DecryptPayload(encPayload).Return(want, nil)

	got, err := svc.Get(ctx, clientSideID, userID)
	require.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestClientPrivateDataService_Get_RepoError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, mockRepo, _, _ := newTestPrivateDataSvc(t, ctrl)
	ctx := context.Background()

	mockRepo.EXPECT().GetPrivateData(ctx, "id1", int64(1)).Return(models.PrivateData{}, errors.New("not found"))

	_, err := svc.Get(ctx, "id1", 1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "get local item")
}

func TestClientPrivateDataService_Get_DecryptError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, mockRepo, _, mockCrypto := newTestPrivateDataSvc(t, ctrl)
	ctx := context.Background()
	encPayload := models.PrivateDataPayload{}

	mockRepo.EXPECT().GetPrivateData(ctx, "id1", int64(1)).Return(
		models.PrivateData{ClientSideID: "id1", Payload: encPayload}, nil,
	)
	mockCrypto.EXPECT().DecryptPayload(encPayload).Return(models.DecipheredPayload{}, errors.New("decrypt fail"))

	_, err := svc.Get(ctx, "id1", 1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "decrypt local item")
}

// ── Update ───────────────────────────────────────────────────────────────────

func TestClientPrivateDataService_Update_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, mockRepo, mockAdapter, mockCrypto := newTestPrivateDataSvc(t, ctrl)
	ctx := context.Background()

	data := models.DecipheredPayload{
		ClientSideID: "id1",
		UserID:       1,
		Metadata:     models.Metadata{Name: "updated"},
	}
	prevItem := models.PrivateData{ClientSideID: "id1", UserID: 1, Version: 5}
	encPayload := models.PrivateDataPayload{}

	mockRepo.EXPECT().GetPrivateData(ctx, "id1", int64(1)).Return(prevItem, nil)
	mockCrypto.EXPECT().EncryptPayload(data).Return(encPayload, nil)
	mockCrypto.EXPECT().ComputeHash(encPayload).Return("newhash", nil)
	mockRepo.EXPECT().UpdatePrivateData(ctx, gomock.Any()).Return(nil)
	mockAdapter.EXPECT().Update(ctx, gomock.Any()).Return(nil)
	mockRepo.EXPECT().IncrementVersion(ctx, "id1", int64(1)).Return(nil)

	err := svc.Update(ctx, data)
	require.NoError(t, err)
}

func TestClientPrivateDataService_Update_GetPrevError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, mockRepo, _, _ := newTestPrivateDataSvc(t, ctrl)
	ctx := context.Background()
	data := models.DecipheredPayload{ClientSideID: "id1", UserID: 1}

	mockRepo.EXPECT().GetPrivateData(ctx, "id1", int64(1)).Return(models.PrivateData{}, errors.New("not found"))

	err := svc.Update(ctx, data)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "load existing local item")
}

func TestClientPrivateDataService_Update_ServerError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, mockRepo, mockAdapter, mockCrypto := newTestPrivateDataSvc(t, ctrl)
	ctx := context.Background()
	data := models.DecipheredPayload{ClientSideID: "id1", UserID: 1}
	prevItem := models.PrivateData{ClientSideID: "id1", UserID: 1, Version: 3}
	encPayload := models.PrivateDataPayload{}

	mockRepo.EXPECT().GetPrivateData(ctx, "id1", int64(1)).Return(prevItem, nil)
	mockCrypto.EXPECT().EncryptPayload(data).Return(encPayload, nil)
	mockCrypto.EXPECT().ComputeHash(encPayload).Return("hash", nil)
	mockRepo.EXPECT().UpdatePrivateData(ctx, gomock.Any()).Return(nil)
	mockAdapter.EXPECT().Update(ctx, gomock.Any()).Return(errors.New("conflict"))

	// IncrementVersion НЕ должен вызываться
	err := svc.Update(ctx, data)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "update item on server")
}

// ── Delete ───────────────────────────────────────────────────────────────────

func TestClientPrivateDataService_Delete_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, mockRepo, mockAdapter, _ := newTestPrivateDataSvc(t, ctrl)
	ctx := context.Background()
	id := "id1"
	uid := int64(1)
	item := models.PrivateData{ClientSideID: id, UserID: uid, Version: 10}

	mockRepo.EXPECT().GetPrivateData(ctx, id, uid).Return(item, nil)
	mockRepo.EXPECT().DeletePrivateData(ctx, id, uid).Return(nil)
	mockAdapter.EXPECT().Delete(ctx, gomock.Any()).Return(nil)
	mockRepo.EXPECT().IncrementVersion(ctx, id, uid).Return(nil)

	err := svc.Delete(ctx, id, uid)
	require.NoError(t, err)
}

func TestClientPrivateDataService_Delete_GetError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, mockRepo, _, _ := newTestPrivateDataSvc(t, ctrl)
	ctx := context.Background()

	mockRepo.EXPECT().GetPrivateData(ctx, "id1", int64(1)).Return(models.PrivateData{}, errors.New("not found"))

	err := svc.Delete(ctx, "id1", 1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "load item for delete")
}

func TestClientPrivateDataService_Delete_SoftDeleteError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, mockRepo, _, _ := newTestPrivateDataSvc(t, ctrl)
	ctx := context.Background()
	id := "id1"
	uid := int64(1)

	mockRepo.EXPECT().GetPrivateData(ctx, id, uid).Return(models.PrivateData{ClientSideID: id, UserID: uid}, nil)
	mockRepo.EXPECT().DeletePrivateData(ctx, id, uid).Return(errors.New("db error"))

	// adapter.Delete НЕ должен вызываться
	err := svc.Delete(ctx, id, uid)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "soft delete local item")
}

func TestClientPrivateDataService_Delete_ServerError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, mockRepo, mockAdapter, _ := newTestPrivateDataSvc(t, ctrl)
	ctx := context.Background()
	id := "id1"
	uid := int64(1)

	mockRepo.EXPECT().GetPrivateData(ctx, id, uid).Return(models.PrivateData{ClientSideID: id, UserID: uid, Version: 2}, nil)
	mockRepo.EXPECT().DeletePrivateData(ctx, id, uid).Return(nil)
	mockAdapter.EXPECT().Delete(ctx, gomock.Any()).Return(errors.New("server error"))

	// IncrementVersion НЕ должен вызываться
	err := svc.Delete(ctx, id, uid)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "delete item on server")
}
