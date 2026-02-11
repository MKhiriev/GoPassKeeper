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

func (p *privateDataRepository) SavePrivateData(ctx context.Context, data models.PrivateData) error {
	//TODO implement me
	panic("implement me")
}

func (p *privateDataRepository) SaveAllPrivateData(ctx context.Context, data []models.PrivateData) error {
	//TODO implement me
	panic("implement me")
}

func (p *privateDataRepository) GetPrivateData(ctx context.Context, data models.PrivateData) (models.PrivateData, error) {
	//TODO implement me
	panic("implement me")
}

func (p *privateDataRepository) GetAllPrivateData(ctx context.Context) ([]models.PrivateData, error) {
	//TODO implement me
	panic("implement me")
}
