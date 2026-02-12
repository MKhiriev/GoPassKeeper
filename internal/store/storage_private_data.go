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
	//TODO implement me
	panic("implement me")
}

func (p *privateDataStorage) Get(ctx context.Context, downloadRequests ...models.DownloadRequest) ([]models.PrivateData, error) {
	//TODO implement me
	panic("implement me")
}

func (p *privateDataStorage) GetAll(ctx context.Context) ([]models.PrivateData, error) {
	//TODO implement me
	panic("implement me")
}

func (p *privateDataStorage) Update(ctx context.Context, updateRequests ...models.UpdateRequest) error {
	//TODO implement me
	panic("implement me")
}

func (p *privateDataStorage) Delete(ctx context.Context, deleteRequests ...models.DeleteRequest) error {
	//TODO implement me
	panic("implement me")
}
