package store

import (
	"context"

	"github.com/MKhiriev/go-pass-keeper/models"
)

type privateDataStorage struct {
	repository  PrivateDataRepository
	fileStorage PrivateDataFileStorage // todo for now we ignore it
}

func NewPrivateDataStorage(repository PrivateDataRepository, fileStorage PrivateDataFileStorage) PrivateDataStorage {
	return &privateDataStorage{
		repository:  repository,
		fileStorage: fileStorage,
	}
}

func (p *privateDataStorage) Save(ctx context.Context, data ...models.PrivateData) error {
	return p.repository.SavePrivateData(ctx, data...)
}

func (p *privateDataStorage) Get(ctx context.Context, downloadRequests models.DownloadRequest) ([]models.PrivateData, error) {
	return p.repository.GetPrivateData(ctx, downloadRequests)
}

func (p *privateDataStorage) GetAll(ctx context.Context, userID int64) ([]models.PrivateData, error) {
	return p.repository.GetAllPrivateData(ctx, userID)
}

func (p *privateDataStorage) Update(ctx context.Context, updateRequests models.UpdateRequest) error {
	return p.repository.UpdatePrivateData(ctx, updateRequests)
}

func (p *privateDataStorage) Delete(ctx context.Context, deleteRequests models.DeleteRequest) error {
	return p.repository.DeletePrivateData(ctx, deleteRequests)
}
