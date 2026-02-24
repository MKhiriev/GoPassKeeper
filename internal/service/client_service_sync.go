package service

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/MKhiriev/go-pass-keeper/internal/adapter"
	"github.com/MKhiriev/go-pass-keeper/internal/store"
	"github.com/MKhiriev/go-pass-keeper/models"
)

type clientSyncService struct {
	localStore store.LocalStorage
	adapter    adapter.ServerAdapter
	planner    SyncService

	mu          sync.RWMutex
	serverState map[string]models.PrivateDataState
}

func NewClientSyncService(localStore store.LocalStorage, serverAdapter adapter.ServerAdapter) ClientSyncService {
	return &clientSyncService{
		localStore:  localStore,
		adapter:     serverAdapter,
		planner:     NewSyncService(),
		serverState: make(map[string]models.PrivateDataState),
	}
}

func (s *clientSyncService) FullSync(ctx context.Context, userID int64) error {
	serverStates, err := s.adapter.GetServerStates(ctx, userID)
	if err != nil {
		return fmt.Errorf("get server states: %w", err)
	}

	clientStates, err := s.localStore.GetAllStates(ctx, userID)
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
	s.mu.Lock()
	s.serverState = idx
	s.mu.Unlock()

	if err = s.ExecutePlan(ctx, plan); err != nil {
		return fmt.Errorf("execute sync plan: %w", err)
	}

	return nil
}

func (s *clientSyncService) ExecutePlan(ctx context.Context, plan models.SyncPlan) error {
	if len(plan.Download) > 0 {
		ids := collectIDs(plan.Download)
		items, err := s.adapter.Download(ctx, models.DownloadRequest{ClientSideIDs: ids, Length: len(ids)})
		if err != nil {
			return fmt.Errorf("download states in plan: %w", err)
		}
		batch := make([]*models.PrivateData, 0, len(items))
		for i := range items {
			item := items[i]
			batch = append(batch, &item)
		}
		if err = s.localStore.Save(ctx, batch...); err != nil {
			return fmt.Errorf("save downloaded items locally: %w", err)
		}
	}

	if len(plan.Upload) > 0 {
		payload := make([]*models.PrivateData, 0, len(plan.Upload))
		for _, st := range plan.Upload {
			item, err := s.localStore.Get(ctx, st.ClientSideID)
			if err != nil {
				return fmt.Errorf("get local upload item %s: %w", st.ClientSideID, err)
			}
			it := item
			payload = append(payload, &it)
		}
		if err := s.adapter.Upload(ctx, models.UploadRequest{PrivateDataList: payload, Length: len(payload)}); err != nil {
			return fmt.Errorf("upload items in sync plan: %w", err)
		}
	}

	for _, st := range plan.Update {
		if err := s.updateServer(ctx, st.ClientSideID); err != nil {
			return err
		}
	}

	for _, st := range plan.DeleteClient {
		if err := s.localStore.SoftDelete(ctx, st.ClientSideID, st.Version); err != nil {
			return fmt.Errorf("delete on client for %s: %w", st.ClientSideID, err)
		}
	}

	for _, st := range plan.DeleteServer {
		if err := s.deleteServer(ctx, st.ClientSideID); err != nil {
			return err
		}
	}

	return nil
}

func (s *clientSyncService) updateServer(ctx context.Context, clientSideID string) error {
	item, err := s.localStore.Get(ctx, clientSideID)
	if err != nil {
		return fmt.Errorf("load local item for update %s: %w", clientSideID, err)
	}

	meta := item.Payload.Metadata
	data := item.Payload.Data
	baseVersion := s.serverVersion(clientSideID, item.Version)
	req := models.UpdateRequest{
		UserID: item.UserID,
		PrivateDataUpdates: []models.PrivateDataUpdate{{
			ClientSideID:      item.ClientSideID,
			Version:           baseVersion,
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
	if !errors.Is(err, adapter.ErrVersionConflict) {
		return fmt.Errorf("update server item %s: %w", clientSideID, err)
	}

	return s.refreshConflict(ctx, item.UserID, clientSideID)
}

func (s *clientSyncService) deleteServer(ctx context.Context, clientSideID string) error {
	item, err := s.localStore.Get(ctx, clientSideID)
	if err != nil {
		return fmt.Errorf("load local item for delete %s: %w", clientSideID, err)
	}

	req := models.DeleteRequest{UserID: item.UserID, DeleteEntries: []models.DeleteEntry{{
		ClientSideID: clientSideID,
		Version:      s.serverVersion(clientSideID, item.Version),
	}}}

	err = s.adapter.Delete(ctx, req)
	if err == nil {
		return nil
	}
	if !errors.Is(err, adapter.ErrVersionConflict) {
		return fmt.Errorf("delete server item %s: %w", clientSideID, err)
	}

	return s.refreshConflict(ctx, item.UserID, clientSideID)
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

	batch := make([]*models.PrivateData, 0, len(items))
	for i := range items {
		item := items[i]
		batch = append(batch, &item)
	}
	if err = s.localStore.Save(ctx, batch...); err != nil {
		return fmt.Errorf("save conflict item %s: %w", clientSideID, err)
	}
	return nil
}

func (s *clientSyncService) serverVersion(clientSideID string, fallback int64) int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if st, ok := s.serverState[clientSideID]; ok {
		return st.Version
	}
	if fallback > 0 {
		return fallback - 1
	}
	return 0
}

func collectIDs(states []models.PrivateDataState) []string {
	ids := make([]string, 0, len(states))
	for _, st := range states {
		ids = append(ids, st.ClientSideID)
	}
	return ids
}
