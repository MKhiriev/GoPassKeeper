package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/MKhiriev/go-pass-keeper/internal/adapter"
	"github.com/MKhiriev/go-pass-keeper/internal/store"
	"github.com/MKhiriev/go-pass-keeper/models"
)

type clientPrivateDataService struct {
	localStore store.LocalStorage
	adapter    adapter.ServerAdapter
	crypto     ClientCryptoService
	key        []byte
}

func NewClientPrivateDataService(localStore store.LocalStorage, serverAdapter adapter.ServerAdapter, crypto ClientCryptoService) ClientPrivateDataService {
	return &clientPrivateDataService{localStore: localStore, adapter: serverAdapter, crypto: crypto}
}

func (p *clientPrivateDataService) SetEncryptionKey(key []byte) {
	p.key = append([]byte(nil), key...)
}

func (p *clientPrivateDataService) Create(ctx context.Context, userID int64, plain models.PrivateDataPayload) error {
	encPayload, err := p.crypto.EncryptPayload(plain, p.key)
	if err != nil {
		return fmt.Errorf("encrypt payload for create: %w", err)
	}

	now := time.Now().UTC()
	item := &models.PrivateData{
		ClientSideID: uuid.NewString(),
		UserID:       userID,
		Payload:      encPayload,
		Hash:         p.crypto.ComputeHash(encPayload),
		Version:      0,
		CreatedAt:    &now,
		UpdatedAt:    &now,
	}

	if err = p.localStore.Save(ctx, item); err != nil {
		return fmt.Errorf("save created item to local store: %w", err)
	}

	if err = p.adapter.Upload(ctx, models.UploadRequest{UserID: userID, PrivateDataList: []*models.PrivateData{item}}); err != nil {
		return fmt.Errorf("upload created item to server: %w", err)
	}

	return nil
}

func (p *clientPrivateDataService) GetAll(ctx context.Context, userID int64) ([]models.PrivateData, error) {
	items, err := p.localStore.GetAll(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get all local items: %w", err)
	}

	decrypted := make([]models.PrivateData, 0, len(items))
	for _, item := range items {
		payload, err := p.crypto.DecryptPayload(item.Payload, p.key)
		if err != nil {
			return nil, fmt.Errorf("decrypt item %s: %w", item.ClientSideID, err)
		}
		item.Payload = payload
		decrypted = append(decrypted, item)
	}
	return decrypted, nil
}

func (p *clientPrivateDataService) Get(ctx context.Context, clientSideID string) (models.PrivateData, error) {
	item, err := p.localStore.Get(ctx, clientSideID)
	if err != nil {
		return models.PrivateData{}, fmt.Errorf("get local item: %w", err)
	}

	payload, err := p.crypto.DecryptPayload(item.Payload, p.key)
	if err != nil {
		return models.PrivateData{}, fmt.Errorf("decrypt local item: %w", err)
	}
	item.Payload = payload

	return item, nil
}

func (p *clientPrivateDataService) Update(ctx context.Context, data models.PrivateData) error {
	prev, err := p.localStore.Get(ctx, data.ClientSideID)
	if err != nil {
		return fmt.Errorf("load existing local item: %w", err)
	}

	encPayload, err := p.crypto.EncryptPayload(data.Payload, p.key)
	if err != nil {
		return fmt.Errorf("encrypt payload for update: %w", err)
	}

	now := time.Now().UTC()
	updated := prev
	updated.Payload = encPayload
	updated.Hash = p.crypto.ComputeHash(encPayload)
	updated.Version = prev.Version + 1
	updated.UpdatedAt = &now

	if err = p.localStore.Update(ctx, updated); err != nil {
		return fmt.Errorf("update local item: %w", err)
	}

	meta := updated.Payload.Metadata
	body := updated.Payload.Data
	req := models.UpdateRequest{
		UserID: updated.UserID,
		PrivateDataUpdates: []models.PrivateDataUpdate{{
			ClientSideID:      updated.ClientSideID,
			Version:           prev.Version,
			UpdatedRecordHash: updated.Hash,
			FieldsUpdate: models.FieldsUpdate{
				Metadata:         &meta,
				Data:             &body,
				Notes:            updated.Payload.Notes,
				AdditionalFields: updated.Payload.AdditionalFields,
			},
		}},
	}

	if err = p.adapter.Update(ctx, req); err != nil {
		return fmt.Errorf("update item on server: %w", err)
	}

	return nil
}

func (p *clientPrivateDataService) Delete(ctx context.Context, clientSideID string, version int64) error {
	item, err := p.localStore.Get(ctx, clientSideID)
	if err != nil {
		return fmt.Errorf("load item for delete: %w", err)
	}

	if err = p.localStore.SoftDelete(ctx, clientSideID, version+1); err != nil {
		return fmt.Errorf("soft delete local item: %w", err)
	}

	req := models.DeleteRequest{
		UserID: item.UserID,
		DeleteEntries: []models.DeleteEntry{{
			ClientSideID: clientSideID,
			Version:      version,
		}},
	}

	if err = p.adapter.Delete(ctx, req); err != nil {
		return fmt.Errorf("delete item on server: %w", err)
	}

	return nil
}
