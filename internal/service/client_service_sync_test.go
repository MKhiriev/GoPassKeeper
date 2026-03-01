// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Rasul Khiriev

package service

import (
	"context"
	"errors"
	"testing"

	"github.com/MKhiriev/go-pass-keeper/internal/adapter"
	"github.com/MKhiriev/go-pass-keeper/internal/mock"
	"github.com/MKhiriev/go-pass-keeper/internal/store"
	"github.com/MKhiriev/go-pass-keeper/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// stubPlanner — простой мок SyncService, не требует mockgen (избегаем цикл импортов).
type stubPlanner struct {
	plan models.SyncPlan
	err  error
}

func (s *stubPlanner) BuildSyncPlan(_ context.Context, _, _ []models.PrivateDataState) (models.SyncPlan, error) {
	return s.plan, s.err
}

// newTestSyncSvc — хелпер для создания clientSyncService с моками
func newTestSyncSvc(
	t *testing.T,
	ctrl *gomock.Controller,
) (
	*clientSyncService,
	*mock.MockLocalPrivateDataRepository,
	*mock.MockServerAdapter,
	*stubPlanner,
) {
	t.Helper()
	mockRepo := mock.NewMockLocalPrivateDataRepository(ctrl)
	mockAdapter := mock.NewMockServerAdapter(ctrl)
	planner := &stubPlanner{}

	storages := &store.ClientStorages{
		PrivateDataRepository: mockRepo,
	}

	svc := NewClientSyncService(storages, mockAdapter).(*clientSyncService)
	svc.planner = planner

	return svc, mockRepo, mockAdapter, planner
}

// ── FullSync ─────────────────────────────────────────────────────────────────

func TestClientSyncService_FullSync_EmptyPlan(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, mockRepo, mockAdapter, planner := newTestSyncSvc(t, ctrl)
	ctx := context.Background()
	userID := int64(1)

	serverStates := []models.PrivateDataState{{ClientSideID: "s1", Version: 1}}
	clientStates := []models.PrivateDataState{{ClientSideID: "s1", Version: 1}}

	mockAdapter.EXPECT().GetServerStates(ctx, userID).Return(serverStates, nil)
	mockRepo.EXPECT().GetAllStates(ctx, userID).Return(clientStates, nil)
	planner.plan = models.SyncPlan{} // пустой план — всё синхронизировано

	err := svc.FullSync(ctx, userID)
	require.NoError(t, err)
}

func TestClientSyncService_FullSync_DownloadAndUpload(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, mockRepo, mockAdapter, planner := newTestSyncSvc(t, ctrl)
	ctx := context.Background()
	userID := int64(1)

	// Сервер: есть «new-on-server» (v2), нет «new-on-client»
	// Клиент: нет «new-on-server», есть «new-on-client» (v1)
	serverStates := []models.PrivateDataState{
		{ClientSideID: "new-on-server", Version: 2, Hash: "srv-hash"},
	}
	clientStates := []models.PrivateDataState{
		{ClientSideID: "new-on-client", Version: 1, Hash: "cli-hash"},
	}

	// Planner решает: скачать серверный, загрузить клиентский
	planner.plan = models.SyncPlan{
		Download: []models.PrivateDataState{{ClientSideID: "new-on-server"}},
		Upload:   []models.PrivateDataState{{ClientSideID: "new-on-client"}},
	}

	mockAdapter.EXPECT().GetServerStates(ctx, userID).Return(serverStates, nil)
	mockRepo.EXPECT().GetAllStates(ctx, userID).Return(clientStates, nil)

	// Download
	downloaded := []models.PrivateData{{ClientSideID: "new-on-server", UserID: userID}}
	mockAdapter.EXPECT().Download(ctx, gomock.Any()).Return(downloaded, nil)
	mockRepo.EXPECT().SavePrivateData(ctx, userID, downloaded[0]).Return(nil)

	// Upload
	localItem := models.PrivateData{ClientSideID: "new-on-client", UserID: userID, Version: 1}
	mockRepo.EXPECT().GetPrivateData(ctx, "new-on-client", userID).Return(localItem, nil)
	mockAdapter.EXPECT().Upload(ctx, gomock.Any()).Return(nil)

	err := svc.FullSync(ctx, userID)
	require.NoError(t, err)
}

func TestClientSyncService_FullSync_UpdateAndDelete(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, mockRepo, mockAdapter, planner := newTestSyncSvc(t, ctrl)
	ctx := context.Background()
	userID := int64(1)

	// Клиент опередил сервер по «edited», а сервер пометил «removed» удалённым
	serverStates := []models.PrivateDataState{
		{ClientSideID: "edited", Version: 3, Hash: "old-hash"},
		{ClientSideID: "removed", Version: 4, Deleted: true},
	}
	clientStates := []models.PrivateDataState{
		{ClientSideID: "edited", Version: 4, Hash: "new-hash"},
		{ClientSideID: "removed", Version: 3},
	}

	planner.plan = models.SyncPlan{
		Update:       []models.PrivateDataState{{ClientSideID: "edited"}},
		DeleteClient: []models.PrivateDataState{{ClientSideID: "removed", Version: 4}},
	}

	mockAdapter.EXPECT().GetServerStates(ctx, userID).Return(serverStates, nil)
	mockRepo.EXPECT().GetAllStates(ctx, userID).Return(clientStates, nil)

	// Update
	mockRepo.EXPECT().GetPrivateData(ctx, "edited", userID).Return(models.PrivateData{
		ClientSideID: "edited", UserID: userID, Version: 4, Hash: "new-hash",
	}, nil)
	mockAdapter.EXPECT().Update(ctx, gomock.Any()).Return(nil)

	// DeleteClient
	mockRepo.EXPECT().DeletePrivateData(ctx, "removed", int64(4)).Return(nil)

	err := svc.FullSync(ctx, userID)
	require.NoError(t, err)
}

func TestClientSyncService_FullSync_DeleteServer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, mockRepo, mockAdapter, planner := newTestSyncSvc(t, ctrl)
	ctx := context.Background()
	userID := int64(1)

	// Клиент удалил элемент, сервер ещё не знает
	serverStates := []models.PrivateDataState{
		{ClientSideID: "del-srv", Version: 2},
	}
	clientStates := []models.PrivateDataState{
		{ClientSideID: "del-srv", Version: 3, Deleted: true},
	}

	planner.plan = models.SyncPlan{
		DeleteServer: []models.PrivateDataState{{ClientSideID: "del-srv"}},
	}

	mockAdapter.EXPECT().GetServerStates(ctx, userID).Return(serverStates, nil)
	mockRepo.EXPECT().GetAllStates(ctx, userID).Return(clientStates, nil)

	mockRepo.EXPECT().GetPrivateData(ctx, "del-srv", userID).Return(models.PrivateData{
		ClientSideID: "del-srv", UserID: userID, Version: 3, Deleted: true,
	}, nil)
	mockAdapter.EXPECT().Delete(ctx, gomock.Any()).Return(nil)

	err := svc.FullSync(ctx, userID)
	require.NoError(t, err)
}

func TestClientSyncService_FullSync_GetServerStatesError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, _, mockAdapter, _ := newTestSyncSvc(t, ctrl)
	ctx := context.Background()

	mockAdapter.EXPECT().GetServerStates(ctx, int64(1)).Return(nil, errors.New("network error"))

	err := svc.FullSync(ctx, 1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "get server states")
}

func TestClientSyncService_FullSync_GetLocalStatesError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, mockRepo, mockAdapter, _ := newTestSyncSvc(t, ctrl)
	ctx := context.Background()

	mockAdapter.EXPECT().GetServerStates(ctx, int64(1)).Return(nil, nil)
	mockRepo.EXPECT().GetAllStates(ctx, int64(1)).Return(nil, errors.New("db error"))

	err := svc.FullSync(ctx, 1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "get local states")
}

func TestClientSyncService_FullSync_BuildSyncPlanError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, mockRepo, mockAdapter, planner := newTestSyncSvc(t, ctrl)
	ctx := context.Background()

	mockAdapter.EXPECT().GetServerStates(ctx, int64(1)).Return(nil, nil)
	mockRepo.EXPECT().GetAllStates(ctx, int64(1)).Return(nil, nil)
	planner.err = errors.New("plan error")

	err := svc.FullSync(ctx, 1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "build sync plan")
}

func TestClientSyncService_FullSync_ExecutePlanError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, mockRepo, mockAdapter, planner := newTestSyncSvc(t, ctrl)
	ctx := context.Background()
	userID := int64(1)

	planner.plan = models.SyncPlan{
		Download: []models.PrivateDataState{{ClientSideID: "d1"}},
	}

	mockAdapter.EXPECT().GetServerStates(ctx, userID).Return(nil, nil)
	mockRepo.EXPECT().GetAllStates(ctx, userID).Return(nil, nil)
	// download упадёт
	mockAdapter.EXPECT().Download(ctx, gomock.Any()).Return(nil, errors.New("download failed"))

	err := svc.FullSync(ctx, userID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "execute sync plan")
}

// ── ExecutePlan: Download ────────────────────────────────────────────────────

func TestClientSyncService_ExecutePlan_DownloadSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, mockRepo, mockAdapter, _ := newTestSyncSvc(t, ctrl)
	ctx := context.Background()
	userID := int64(1)

	plan := models.SyncPlan{
		Download: []models.PrivateDataState{
			{ClientSideID: "d1"},
			{ClientSideID: "d2"},
		},
	}

	downloaded := []models.PrivateData{
		{ClientSideID: "d1", UserID: userID},
		{ClientSideID: "d2", UserID: userID},
	}

	mockAdapter.EXPECT().Download(ctx, gomock.Any()).DoAndReturn(
		func(_ context.Context, req models.DownloadRequest) ([]models.PrivateData, error) {
			assert.ElementsMatch(t, []string{"d1", "d2"}, req.ClientSideIDs)
			assert.Equal(t, 2, req.Length)
			return downloaded, nil
		},
	)
	mockRepo.EXPECT().SavePrivateData(ctx, userID, downloaded[0], downloaded[1]).Return(nil)

	err := svc.ExecutePlan(ctx, plan, userID)
	require.NoError(t, err)
}

func TestClientSyncService_ExecutePlan_DownloadError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, _, mockAdapter, _ := newTestSyncSvc(t, ctrl)
	ctx := context.Background()

	plan := models.SyncPlan{
		Download: []models.PrivateDataState{{ClientSideID: "d1"}},
	}

	mockAdapter.EXPECT().Download(ctx, gomock.Any()).Return(nil, errors.New("timeout"))

	err := svc.ExecutePlan(ctx, plan, 1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "error sync downloading data from server")
}

func TestClientSyncService_ExecutePlan_DownloadSaveError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, mockRepo, mockAdapter, _ := newTestSyncSvc(t, ctrl)
	ctx := context.Background()

	plan := models.SyncPlan{
		Download: []models.PrivateDataState{{ClientSideID: "d1"}},
	}

	mockAdapter.EXPECT().Download(ctx, gomock.Any()).Return([]models.PrivateData{{ClientSideID: "d1"}}, nil)
	mockRepo.EXPECT().SavePrivateData(ctx, int64(1), gomock.Any()).Return(errors.New("db write error"))

	err := svc.ExecutePlan(ctx, plan, 1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "error saving downloaded items locally")
}

// ── ExecutePlan: Upload ──────────────────────────────────────────────────────

func TestClientSyncService_ExecutePlan_UploadSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, mockRepo, mockAdapter, _ := newTestSyncSvc(t, ctrl)
	ctx := context.Background()
	userID := int64(1)

	plan := models.SyncPlan{
		Upload: []models.PrivateDataState{
			{ClientSideID: "u1"},
			{ClientSideID: "u2"},
		},
	}

	item1 := models.PrivateData{ClientSideID: "u1", UserID: userID, Version: 1}
	item2 := models.PrivateData{ClientSideID: "u2", UserID: userID, Version: 1}

	mockRepo.EXPECT().GetPrivateData(ctx, "u1", userID).Return(item1, nil)
	mockRepo.EXPECT().GetPrivateData(ctx, "u2", userID).Return(item2, nil)
	mockAdapter.EXPECT().Upload(ctx, gomock.Any()).DoAndReturn(
		func(_ context.Context, req models.UploadRequest) error {
			assert.Len(t, req.PrivateDataList, 2)
			assert.Equal(t, 2, req.Length)
			return nil
		},
	)

	err := svc.ExecutePlan(ctx, plan, userID)
	require.NoError(t, err)
}

func TestClientSyncService_ExecutePlan_UploadGetLocalItemError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, mockRepo, _, _ := newTestSyncSvc(t, ctrl)
	ctx := context.Background()

	plan := models.SyncPlan{
		Upload: []models.PrivateDataState{{ClientSideID: "u1"}},
	}

	mockRepo.EXPECT().GetPrivateData(ctx, "u1", int64(1)).Return(models.PrivateData{}, errors.New("not found"))

	err := svc.ExecutePlan(ctx, plan, 1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "error getting client item for upload u1")
}

func TestClientSyncService_ExecutePlan_UploadAdapterError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, mockRepo, mockAdapter, _ := newTestSyncSvc(t, ctrl)
	ctx := context.Background()
	userID := int64(1)

	plan := models.SyncPlan{
		Upload: []models.PrivateDataState{{ClientSideID: "u1"}},
	}

	mockRepo.EXPECT().GetPrivateData(ctx, "u1", userID).Return(models.PrivateData{ClientSideID: "u1"}, nil)
	mockAdapter.EXPECT().Upload(ctx, gomock.Any()).Return(errors.New("server error"))

	err := svc.ExecutePlan(ctx, plan, userID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "upload items in sync plan")
}

// ── ExecutePlan: Update ──────────────────────────────────────────────────────

func TestClientSyncService_ExecutePlan_UpdateSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, mockRepo, mockAdapter, _ := newTestSyncSvc(t, ctrl)
	ctx := context.Background()
	userID := int64(1)

	meta := models.CipheredMetadata("enc-meta")
	data := models.CipheredData("enc-data")

	plan := models.SyncPlan{
		Update: []models.PrivateDataState{{ClientSideID: "up1"}},
	}

	item := models.PrivateData{
		ClientSideID: "up1",
		UserID:       userID,
		Version:      3,
		Hash:         "hash123",
		Payload: models.PrivateDataPayload{
			Metadata: meta,
			Data:     data,
		},
	}

	mockRepo.EXPECT().GetPrivateData(ctx, "up1", userID).Return(item, nil)
	mockAdapter.EXPECT().Update(ctx, gomock.Any()).DoAndReturn(
		func(_ context.Context, req models.UpdateRequest) error {
			require.Len(t, req.PrivateDataUpdates, 1)
			upd := req.PrivateDataUpdates[0]
			assert.Equal(t, "up1", upd.ClientSideID)
			assert.Equal(t, int64(3), upd.Version)
			assert.Equal(t, "hash123", upd.UpdatedRecordHash)
			assert.Equal(t, &meta, upd.FieldsUpdate.Metadata)
			assert.Equal(t, &data, upd.FieldsUpdate.Data)
			return nil
		},
	)

	err := svc.ExecutePlan(ctx, plan, userID)
	require.NoError(t, err)
}

func TestClientSyncService_ExecutePlan_UpdateGetLocalError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, mockRepo, _, _ := newTestSyncSvc(t, ctrl)
	ctx := context.Background()

	plan := models.SyncPlan{
		Update: []models.PrivateDataState{{ClientSideID: "up1"}},
	}

	mockRepo.EXPECT().GetPrivateData(ctx, "up1", int64(1)).Return(models.PrivateData{}, errors.New("not found"))

	err := svc.ExecutePlan(ctx, plan, 1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "load local item for update up1")
}

func TestClientSyncService_ExecutePlan_UpdateServerError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, mockRepo, mockAdapter, _ := newTestSyncSvc(t, ctrl)
	ctx := context.Background()
	userID := int64(1)

	plan := models.SyncPlan{
		Update: []models.PrivateDataState{{ClientSideID: "up1"}},
	}

	mockRepo.EXPECT().GetPrivateData(ctx, "up1", userID).Return(models.PrivateData{
		ClientSideID: "up1", UserID: userID, Version: 1,
	}, nil)
	mockAdapter.EXPECT().Update(ctx, gomock.Any()).Return(errors.New("server error"))

	err := svc.ExecutePlan(ctx, plan, userID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "update server item up1")
}

// ── ExecutePlan: Update with conflict → refreshConflict ──────────────────────

func TestClientSyncService_ExecutePlan_UpdateConflict_RefreshSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, mockRepo, mockAdapter, _ := newTestSyncSvc(t, ctrl)
	ctx := context.Background()
	userID := int64(1)

	plan := models.SyncPlan{
		Update: []models.PrivateDataState{{ClientSideID: "up1"}},
	}

	item := models.PrivateData{ClientSideID: "up1", UserID: userID, Version: 2}
	refreshed := []models.PrivateData{{ClientSideID: "up1", UserID: userID, Version: 5}}

	mockRepo.EXPECT().GetPrivateData(ctx, "up1", userID).Return(item, nil)
	// Update возвращает конфликт
	mockAdapter.EXPECT().Update(ctx, gomock.Any()).Return(adapter.ErrConflict)
	// refreshConflict: скачиваем и сохраняем серверную версию
	mockAdapter.EXPECT().Download(ctx, gomock.Any()).DoAndReturn(
		func(_ context.Context, req models.DownloadRequest) ([]models.PrivateData, error) {
			assert.Equal(t, []string{"up1"}, req.ClientSideIDs)
			return refreshed, nil
		},
	)
	mockRepo.EXPECT().SavePrivateData(ctx, userID, refreshed[0]).Return(nil)

	err := svc.ExecutePlan(ctx, plan, userID)
	require.NoError(t, err)
}

func TestClientSyncService_ExecutePlan_UpdateConflict_DownloadError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, mockRepo, mockAdapter, _ := newTestSyncSvc(t, ctrl)
	ctx := context.Background()
	userID := int64(1)

	plan := models.SyncPlan{
		Update: []models.PrivateDataState{{ClientSideID: "up1"}},
	}

	mockRepo.EXPECT().GetPrivateData(ctx, "up1", userID).Return(models.PrivateData{
		ClientSideID: "up1", UserID: userID, Version: 2,
	}, nil)
	mockAdapter.EXPECT().Update(ctx, gomock.Any()).Return(adapter.ErrConflict)
	mockAdapter.EXPECT().Download(ctx, gomock.Any()).Return(nil, errors.New("download failed"))

	err := svc.ExecutePlan(ctx, plan, userID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "download conflict item up1")
}

func TestClientSyncService_ExecutePlan_UpdateConflict_EmptyDownload(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, mockRepo, mockAdapter, _ := newTestSyncSvc(t, ctrl)
	ctx := context.Background()
	userID := int64(1)

	plan := models.SyncPlan{
		Update: []models.PrivateDataState{{ClientSideID: "up1"}},
	}

	mockRepo.EXPECT().GetPrivateData(ctx, "up1", userID).Return(models.PrivateData{
		ClientSideID: "up1", UserID: userID, Version: 2,
	}, nil)
	mockAdapter.EXPECT().Update(ctx, gomock.Any()).Return(adapter.ErrConflict)
	// Сервер вернул пустой список — элемент удалён
	mockAdapter.EXPECT().Download(ctx, gomock.Any()).Return([]models.PrivateData{}, nil)

	err := svc.ExecutePlan(ctx, plan, userID)
	require.NoError(t, err)
}

// ── ExecutePlan: DeleteClient ────────────────────────────────────────────────

func TestClientSyncService_ExecutePlan_DeleteClientSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, mockRepo, _, _ := newTestSyncSvc(t, ctrl)
	ctx := context.Background()

	plan := models.SyncPlan{
		DeleteClient: []models.PrivateDataState{
			{ClientSideID: "dc1", Version: 5},
			{ClientSideID: "dc2", Version: 3},
		},
	}

	mockRepo.EXPECT().DeletePrivateData(ctx, "dc1", int64(5)).Return(nil)
	mockRepo.EXPECT().DeletePrivateData(ctx, "dc2", int64(3)).Return(nil)

	err := svc.ExecutePlan(ctx, plan, 1)
	require.NoError(t, err)
}

func TestClientSyncService_ExecutePlan_DeleteClientError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, mockRepo, _, _ := newTestSyncSvc(t, ctrl)
	ctx := context.Background()

	plan := models.SyncPlan{
		DeleteClient: []models.PrivateDataState{{ClientSideID: "dc1", Version: 5}},
	}

	mockRepo.EXPECT().DeletePrivateData(ctx, "dc1", int64(5)).Return(errors.New("db error"))

	err := svc.ExecutePlan(ctx, plan, 1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "delete on client for dc1")
}

// ── ExecutePlan: DeleteServer ────────────────────────────────────────────────

func TestClientSyncService_ExecutePlan_DeleteServerSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, mockRepo, mockAdapter, _ := newTestSyncSvc(t, ctrl)
	ctx := context.Background()
	userID := int64(1)

	plan := models.SyncPlan{
		DeleteServer: []models.PrivateDataState{{ClientSideID: "ds1"}},
	}

	item := models.PrivateData{ClientSideID: "ds1", UserID: userID, Version: 4}

	mockRepo.EXPECT().GetPrivateData(ctx, "ds1", userID).Return(item, nil)
	mockAdapter.EXPECT().Delete(ctx, gomock.Any()).DoAndReturn(
		func(_ context.Context, req models.DeleteRequest) error {
			require.Len(t, req.DeleteEntries, 1)
			assert.Equal(t, "ds1", req.DeleteEntries[0].ClientSideID)
			assert.Equal(t, int64(4), req.DeleteEntries[0].Version)
			assert.Equal(t, userID, req.UserID)
			return nil
		},
	)

	err := svc.ExecutePlan(ctx, plan, userID)
	require.NoError(t, err)
}

func TestClientSyncService_ExecutePlan_DeleteServerGetLocalError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, mockRepo, _, _ := newTestSyncSvc(t, ctrl)
	ctx := context.Background()

	plan := models.SyncPlan{
		DeleteServer: []models.PrivateDataState{{ClientSideID: "ds1"}},
	}

	mockRepo.EXPECT().GetPrivateData(ctx, "ds1", int64(1)).Return(models.PrivateData{}, errors.New("not found"))

	err := svc.ExecutePlan(ctx, plan, 1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "load local item for delete ds1")
}

func TestClientSyncService_ExecutePlan_DeleteServerAdapterError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, mockRepo, mockAdapter, _ := newTestSyncSvc(t, ctrl)
	ctx := context.Background()
	userID := int64(1)

	plan := models.SyncPlan{
		DeleteServer: []models.PrivateDataState{{ClientSideID: "ds1"}},
	}

	mockRepo.EXPECT().GetPrivateData(ctx, "ds1", userID).Return(models.PrivateData{
		ClientSideID: "ds1", UserID: userID, Version: 2,
	}, nil)
	mockAdapter.EXPECT().Delete(ctx, gomock.Any()).Return(errors.New("server error"))

	err := svc.ExecutePlan(ctx, plan, userID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "delete server item ds1")
}

// ── ExecutePlan: DeleteServer with conflict → refreshConflict ────────────────

func TestClientSyncService_ExecutePlan_DeleteServerConflict_RefreshSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, mockRepo, mockAdapter, _ := newTestSyncSvc(t, ctrl)
	ctx := context.Background()
	userID := int64(1)

	plan := models.SyncPlan{
		DeleteServer: []models.PrivateDataState{{ClientSideID: "ds1"}},
	}

	item := models.PrivateData{ClientSideID: "ds1", UserID: userID, Version: 3}
	refreshed := []models.PrivateData{{ClientSideID: "ds1", UserID: userID, Version: 6}}

	mockRepo.EXPECT().GetPrivateData(ctx, "ds1", userID).Return(item, nil)
	mockAdapter.EXPECT().Delete(ctx, gomock.Any()).Return(adapter.ErrConflict)
	mockAdapter.EXPECT().Download(ctx, gomock.Any()).Return(refreshed, nil)
	mockRepo.EXPECT().SavePrivateData(ctx, userID, refreshed[0]).Return(nil)

	err := svc.ExecutePlan(ctx, plan, userID)
	require.NoError(t, err)
}

func TestClientSyncService_ExecutePlan_DeleteServerConflict_RefreshError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, mockRepo, mockAdapter, _ := newTestSyncSvc(t, ctrl)
	ctx := context.Background()
	userID := int64(1)

	plan := models.SyncPlan{
		DeleteServer: []models.PrivateDataState{{ClientSideID: "ds1"}},
	}

	mockRepo.EXPECT().GetPrivateData(ctx, "ds1", userID).Return(models.PrivateData{
		ClientSideID: "ds1", UserID: userID, Version: 3,
	}, nil)
	mockAdapter.EXPECT().Delete(ctx, gomock.Any()).Return(adapter.ErrConflict)
	mockAdapter.EXPECT().Download(ctx, gomock.Any()).Return(nil, errors.New("network error"))

	err := svc.ExecutePlan(ctx, plan, userID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "download conflict item ds1")
}

// ── ExecutePlan: Mixed plan ──────────────────────────────────────────────────

func TestClientSyncService_ExecutePlan_MixedPlan(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, mockRepo, mockAdapter, _ := newTestSyncSvc(t, ctrl)
	ctx := context.Background()
	userID := int64(1)

	plan := models.SyncPlan{
		Download:     []models.PrivateDataState{{ClientSideID: "d1"}},
		Upload:       []models.PrivateDataState{{ClientSideID: "u1"}},
		Update:       []models.PrivateDataState{{ClientSideID: "up1"}},
		DeleteClient: []models.PrivateDataState{{ClientSideID: "dc1", Version: 2}},
		DeleteServer: []models.PrivateDataState{{ClientSideID: "ds1"}},
	}

	// Download
	mockAdapter.EXPECT().Download(ctx, gomock.Any()).Return(
		[]models.PrivateData{{ClientSideID: "d1", UserID: userID}}, nil,
	)
	mockRepo.EXPECT().SavePrivateData(ctx, userID, gomock.Any()).Return(nil)

	// Upload
	mockRepo.EXPECT().GetPrivateData(ctx, "u1", userID).Return(
		models.PrivateData{ClientSideID: "u1", UserID: userID}, nil,
	)
	mockAdapter.EXPECT().Upload(ctx, gomock.Any()).Return(nil)

	// Update
	mockRepo.EXPECT().GetPrivateData(ctx, "up1", userID).Return(
		models.PrivateData{ClientSideID: "up1", UserID: userID, Version: 3}, nil,
	)
	mockAdapter.EXPECT().Update(ctx, gomock.Any()).Return(nil)

	// DeleteClient
	mockRepo.EXPECT().DeletePrivateData(ctx, "dc1", int64(2)).Return(nil)

	// DeleteServer
	mockRepo.EXPECT().GetPrivateData(ctx, "ds1", userID).Return(
		models.PrivateData{ClientSideID: "ds1", UserID: userID, Version: 5}, nil,
	)
	mockAdapter.EXPECT().Delete(ctx, gomock.Any()).Return(nil)

	err := svc.ExecutePlan(ctx, plan, userID)
	require.NoError(t, err)
}

// ── ExecutePlan: Empty plan ──────────────────────────────────────────────────

func TestClientSyncService_ExecutePlan_EmptyPlan(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, _, _, _ := newTestSyncSvc(t, ctrl)
	ctx := context.Background()

	err := svc.ExecutePlan(ctx, models.SyncPlan{}, 1)
	require.NoError(t, err)
}

// ── collectIDs ───────────────────────────────────────────────────────────────

func TestCollectIDs(t *testing.T) {
	states := []models.PrivateDataState{
		{ClientSideID: "a"},
		{ClientSideID: "b"},
		{ClientSideID: "c"},
	}

	ids := collectIDs(states)
	assert.Equal(t, []string{"a", "b", "c"}, ids)
}

func TestCollectIDs_Empty(t *testing.T) {
	ids := collectIDs(nil)
	assert.Empty(t, ids)
}
