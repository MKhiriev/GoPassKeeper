package service

import (
	"context"
	"fmt"
	"time"

	"github.com/MKhiriev/go-pass-keeper/internal/adapter"
	"github.com/MKhiriev/go-pass-keeper/internal/store"
	"github.com/MKhiriev/go-pass-keeper/internal/utils"
	"github.com/MKhiriev/go-pass-keeper/models"
)

type clientPrivateDataService struct {
	localStore        *store.ClientStorages
	adapter           adapter.ServerAdapter
	crypto            ClientCryptoService
	clientIDGenerator *utils.UUIDGenerator
}

func NewClientPrivateDataService(localStore *store.ClientStorages, serverAdapter adapter.ServerAdapter, crypto ClientCryptoService) ClientPrivateDataService {
	return &clientPrivateDataService{localStore: localStore, adapter: serverAdapter, crypto: crypto, clientIDGenerator: utils.NewUUIDGenerator()}
}

func (p *clientPrivateDataService) SetEncryptionKey(key []byte) {
	p.crypto.SetEncryptionKey(key)
}

func (p *clientPrivateDataService) Create(ctx context.Context, userID int64, plain models.DecipheredPayload) error {
	encPayload, err := p.crypto.EncryptPayload(plain)
	if err != nil {
		return fmt.Errorf("encrypt payload for create: %w", err)
	}

	clientSideID := p.clientIDGenerator.Generate()
	now := time.Now().UTC()

	hash, err := p.crypto.ComputeHash(encPayload)
	if err != nil {
		return fmt.Errorf("compute hash with encrypted payload for create: %w", err)
	}

	item := models.PrivateData{
		ClientSideID: clientSideID,
		UserID:       userID,
		Payload:      encPayload,
		Hash:         hash,
		Version:      0,
		CreatedAt:    &now,
	}

	if err = p.localStore.PrivateDataRepository.SavePrivateData(ctx, userID, item); err != nil {
		return fmt.Errorf("save created item to local store: %w", err)
	}

	if err = p.adapter.Upload(ctx, models.UploadRequest{UserID: userID, PrivateDataList: []*models.PrivateData{&item}}); err != nil {
		return fmt.Errorf("upload created item to server: %w", err)
	} else {
		incrementErr := p.localStore.PrivateDataRepository.IncrementVersion(ctx, clientSideID, userID)
		if incrementErr != nil {
			return fmt.Errorf("error incrementing version locally: %w", incrementErr)
		}
	}

	return nil
}

func (p *clientPrivateDataService) GetAll(ctx context.Context, userID int64) ([]models.DecipheredPayload, error) {
	items, err := p.localStore.PrivateDataRepository.GetAllPrivateData(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get all local items: %w", err)
	}

	decrypted := make([]models.DecipheredPayload, 0, len(items))
	for _, item := range items {
		payload, err := p.crypto.DecryptPayload(item.Payload)
		if err != nil {
			return nil, fmt.Errorf("decrypt item %s: %w", item.ClientSideID, err)
		}

		decrypted = append(decrypted, payload)
	}

	return decrypted, nil
}

func (p *clientPrivateDataService) Get(ctx context.Context, clientSideID string, userID int64) (models.DecipheredPayload, error) {
	item, err := p.localStore.PrivateDataRepository.GetPrivateData(ctx, clientSideID, userID)
	if err != nil {
		return models.DecipheredPayload{}, fmt.Errorf("get local item: %w", err)
	}

	payload, err := p.crypto.DecryptPayload(item.Payload)
	if err != nil {
		return models.DecipheredPayload{}, fmt.Errorf("decrypt local item: %w", err)
	}

	return payload, nil
}

func (p *clientPrivateDataService) Update(ctx context.Context, data models.DecipheredPayload) error {
	prev, err := p.localStore.PrivateDataRepository.GetPrivateData(ctx, data.ClientSideID, data.UserID)
	if err != nil {
		return fmt.Errorf("load existing local item: %w", err)
	}

	encPayload, err := p.crypto.EncryptPayload(data)
	if err != nil {
		return fmt.Errorf("encrypt payload for update: %w", err)
	}

	hash, err := p.crypto.ComputeHash(encPayload)
	if err != nil {
		return fmt.Errorf("compute hash with encrypted payload for create: %w", err)
	}

	now := time.Now().UTC()
	updated := prev
	updated.Payload = encPayload
	updated.Hash = hash
	updated.UpdatedAt = &now

	if err = p.localStore.PrivateDataRepository.UpdatePrivateData(ctx, updated); err != nil {
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
	} else {
		incrementErr := p.localStore.PrivateDataRepository.IncrementVersion(ctx, prev.ClientSideID, prev.UserID)
		if incrementErr != nil {
			return fmt.Errorf("error incrementing version locally: %w", incrementErr)
		}
	}

	return nil
}

func (p *clientPrivateDataService) Delete(ctx context.Context, clientSideID string, userID int64) error {
	item, err := p.localStore.PrivateDataRepository.GetPrivateData(ctx, clientSideID, userID)
	if err != nil {
		return fmt.Errorf("load item for delete: %w", err)
	}

	if err = p.localStore.PrivateDataRepository.DeletePrivateData(ctx, clientSideID, userID); err != nil {
		return fmt.Errorf("soft delete local item: %w", err)
	}

	req := models.DeleteRequest{
		UserID: item.UserID,
		DeleteEntries: []models.DeleteEntry{{
			ClientSideID: clientSideID,
			Version:      item.Version,
		}},
	}

	if err = p.adapter.Delete(ctx, req); err != nil {
		return fmt.Errorf("delete item on server: %w", err)
	} else {
		incrementErr := p.localStore.PrivateDataRepository.IncrementVersion(ctx, clientSideID, userID)
		if incrementErr != nil {
			return fmt.Errorf("error incrementing version locally: %w", incrementErr)
		}
	}

	return nil
}
