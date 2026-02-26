package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/MKhiriev/go-pass-keeper/internal/adapter"
	"github.com/MKhiriev/go-pass-keeper/internal/store"
	"github.com/MKhiriev/go-pass-keeper/models"
)

type clientSyncService struct {
	localStore *store.ClientStorages
	adapter    adapter.ServerAdapter
	planner    SyncService
}

func NewClientSyncService(localStore *store.ClientStorages, serverAdapter adapter.ServerAdapter) ClientSyncService {
	return &clientSyncService{
		localStore: localStore,
		adapter:    serverAdapter,
		planner:    NewSyncService(),
	}
}

func (s *clientSyncService) FullSync(ctx context.Context, userID int64) error {
	if userID <= 0 {
		return fmt.Errorf("full sync: invalid user id")
	}

	serverStates, err := s.adapter.GetServerStates(ctx, userID)
	if err != nil {
		return fmt.Errorf("get server states: %w", err)
	}

	clientStates, err := s.localStore.PrivateDataRepository.GetAllStates(ctx, userID)
	if err != nil {
		return fmt.Errorf("get local states: %w", err)
	}

	plan, err := s.planner.BuildSyncPlan(ctx, serverStates, clientStates)
	if err != nil {
		return fmt.Errorf("build sync plan: %w", err)
	}

	idx := make(map[string]models.PrivateDataState, len(serverStates))
	for _, st := range serverStates {
		idx[st.ClientSideID] = st
	}

	if err = s.ExecutePlan(ctx, plan, userID); err != nil {
		return fmt.Errorf("execute sync plan: %w", err)
	}

	return nil
}

func (s *clientSyncService) ExecutePlan(ctx context.Context, plan models.SyncPlan, userID int64) error {
	if userID <= 0 {
		return fmt.Errorf("execute sync plan: invalid user id")
	}

	if len(plan.Download) > 0 {
		if err := s.downloadFromServer(ctx, plan, userID); err != nil {
			return err
		}
	}

	if len(plan.Upload) > 0 {
		if err := s.uploadToServer(ctx, plan, userID); err != nil {
			return err
		}
	}

	for _, st := range plan.Update {
		if err := s.updateServerData(ctx, st.ClientSideID, userID); err != nil {
			return err
		}
	}

	for _, st := range plan.DeleteClient {
		if err := s.deleteFromClient(ctx, st.ClientSideID, st.Version); err != nil {
			return err
		}
	}

	for _, st := range plan.DeleteServer {
		if err := s.deleteFromServer(ctx, st.ClientSideID, userID); err != nil {
			return err
		}
	}

	return nil
}

func (s *clientSyncService) downloadFromServer(ctx context.Context, plan models.SyncPlan, userID int64) error {
	ids := collectIDs(plan.Download)

	downloadedData, err := s.adapter.Download(ctx, models.DownloadRequest{
		UserID:        userID,
		ClientSideIDs: ids,
		Length:        len(ids),
	})
	if err != nil {
		return fmt.Errorf("error sync downloading data from server: %w", err)
	}

	if err = s.localStore.PrivateDataRepository.SavePrivateData(ctx, userID, downloadedData...); err != nil {
		return fmt.Errorf("error saving downloaded items locally: %w", err)
	}

	return nil
}

func (s *clientSyncService) uploadToServer(ctx context.Context, plan models.SyncPlan, userID int64) error {
	payload := make([]*models.PrivateData, 0, len(plan.Upload))

	for _, st := range plan.Upload {
		item, err := s.localStore.PrivateDataRepository.GetPrivateData(ctx, st.ClientSideID, userID)
		if err != nil {
			return fmt.Errorf("error getting client item for upload %s: %w", st.ClientSideID, err)
		}

		it := item
		payload = append(payload, &it)
	}

	if err := s.adapter.Upload(ctx, models.UploadRequest{
		UserID:          userID,
		PrivateDataList: payload,
		Length:          len(payload),
	}); err != nil {
		return fmt.Errorf("upload items in sync plan: %w", err)
	}

	return nil
}

func (s *clientSyncService) updateServerData(ctx context.Context, clientSideID string, userID int64) error {
	item, err := s.localStore.PrivateDataRepository.GetPrivateData(ctx, clientSideID, userID)
	if err != nil {
		return fmt.Errorf("load local item for update %s: %w", clientSideID, err)
	}

	meta := item.Payload.Metadata
	data := item.Payload.Data
	req := models.UpdateRequest{
		UserID: userID,
		PrivateDataUpdates: []models.PrivateDataUpdate{{
			ClientSideID:      item.ClientSideID,
			Version:           item.Version,
			UpdatedRecordHash: item.Hash,
			FieldsUpdate: models.FieldsUpdate{
				Metadata:         &meta,
				Data:             &data,
				Notes:            item.Payload.Notes,
				AdditionalFields: item.Payload.AdditionalFields,
			},
		}},
	}

	err = s.adapter.Update(ctx, req)
	if err == nil {
		return nil
	}
	if !errors.Is(err, adapter.ErrConflict) {
		return fmt.Errorf("update server item %s: %w", clientSideID, err)
	}

	return s.refreshConflict(ctx, userID, clientSideID)
}

func (s *clientSyncService) deleteFromClient(ctx context.Context, clientSideID string, version int64) error {
	if err := s.localStore.PrivateDataRepository.DeletePrivateData(ctx, clientSideID, version); err != nil {
		return fmt.Errorf("delete on client for %s: %w", clientSideID, err)
	}

	return nil
}

func (s *clientSyncService) deleteFromServer(ctx context.Context, clientSideID string, userID int64) error {
	item, err := s.localStore.PrivateDataRepository.GetPrivateData(ctx, clientSideID, userID)
	if err != nil {
		return fmt.Errorf("load local item for delete %s: %w", clientSideID, err)
	}

	req := models.DeleteRequest{UserID: userID, DeleteEntries: []models.DeleteEntry{{
		ClientSideID: clientSideID,
		Version:      item.Version,
	}}}

	err = s.adapter.Delete(ctx, req)
	if err == nil {
		return nil
	}
	if !errors.Is(err, adapter.ErrConflict) {
		return fmt.Errorf("delete server item %s: %w", clientSideID, err)
	}

	return s.refreshConflict(ctx, userID, clientSideID)
}

func (s *clientSyncService) refreshConflict(ctx context.Context, userID int64, clientSideID string) error {
	req := models.DownloadRequest{UserID: userID, ClientSideIDs: []string{clientSideID}, Length: 1}
	items, err := s.adapter.Download(ctx, req)
	if err != nil {
		return fmt.Errorf("download conflict item %s: %w", clientSideID, err)
	}
	if len(items) == 0 {
		return nil
	}

	if err = s.localStore.PrivateDataRepository.SavePrivateData(ctx, userID, items...); err != nil {
		return fmt.Errorf("save conflict item %s: %w", clientSideID, err)
	}
	return nil
}

func collectIDs(states []models.PrivateDataState) []string {
	ids := make([]string, 0, len(states))
	for _, st := range states {
		ids = append(ids, st.ClientSideID)
	}
	return ids
}
