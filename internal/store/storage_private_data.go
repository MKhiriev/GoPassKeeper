package store

import (
	"context"

	"github.com/MKhiriev/go-pass-keeper/internal/config"
	"github.com/MKhiriev/go-pass-keeper/internal/logger"
	"github.com/MKhiriev/go-pass-keeper/models"
)

type privateDataStorage struct {
	repository  PrivateDataRepository
	fileStorage PrivateDataFileStorage // todo for now we ignore it

	logger *logger.Logger
}

func NewPrivateDataStorage(db *DB, cfg config.Storage, logger *logger.Logger) PrivateDataStorage {
	logger.Debug().Msg("creating private data storage")
	storage := new(privateDataStorage)

	repository := NewPrivateDataRepository(db, logger)
	storage.repository = repository

	if cfg.Files.BinaryDataDir != "" {
		fileStorage := NewPrivateDataFileStorage()
		storage.fileStorage = fileStorage
	}

	return storage
}

func (p *privateDataStorage) Save(ctx context.Context, data ...*models.PrivateData) error {
	return p.repository.SavePrivateData(ctx, data...)
}

func (p *privateDataStorage) Get(ctx context.Context, downloadRequests models.DownloadRequest) ([]models.PrivateData, error) {
	return p.repository.GetPrivateData(ctx, downloadRequests)
}

func (p *privateDataStorage) GetAll(ctx context.Context, userID int64) ([]models.PrivateData, error) {
	return p.repository.GetAllPrivateData(ctx, userID)
}

func (p *privateDataStorage) GetAllStates(ctx context.Context, userID int64) ([]models.PrivateDataState, error) {
	return p.repository.GetAllStates(ctx, userID)
}

func (p *privateDataStorage) GetStates(ctx context.Context, syncRequest models.SyncRequest) ([]models.PrivateDataState, error) {
	return p.repository.GetStates(ctx, syncRequest)
}

func (p *privateDataStorage) Update(ctx context.Context, updateRequests models.UpdateRequest) error {
	return p.repository.UpdatePrivateData(ctx, updateRequests)
}

func (p *privateDataStorage) Delete(ctx context.Context, deleteRequests models.DeleteRequest) error {
	return p.repository.DeletePrivateData(ctx, deleteRequests)
}
