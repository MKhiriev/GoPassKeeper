package store

import (
	"context"

	"github.com/MKhiriev/go-pass-keeper/models"
)

type privateDataRepository struct {
}

func NewPrivateDataRepository() PrivateDataRepository {
	return &privateDataRepository{}
}

func (p *privateDataRepository) SavePrivateData(ctx context.Context, data ...models.PrivateData) error {
	//TODO implement me
	panic("implement me")
}

func (p *privateDataRepository) GetPrivateData(ctx context.Context, downloadRequests ...models.DownloadRequest) ([]models.PrivateData, error) {
	//TODO implement me
	panic("implement me")
}

func (p *privateDataRepository) GetAllPrivateData(ctx context.Context, userID int64) ([]models.PrivateData, error) {
	//TODO implement me
	panic("implement me")
}

func (p *privateDataRepository) UpdatePrivateData(ctx context.Context, updateRequests ...models.UpdateRequest) error {
	//TODO implement me
	panic("implement me")
}

func (p *privateDataRepository) DeletePrivateData(ctx context.Context, deleteRequests ...models.DeleteRequest) error {
	//TODO implement me
	panic("implement me")
}
