package store

import (
	"context"

	"github.com/MKhiriev/go-pass-keeper/models"
)

type privateDataFileStorage struct {
}

func NewPrivateDataFileStorage() PrivateDataFileStorage {
	return &privateDataFileStorage{}
}

func (p *privateDataFileStorage) SaveBinaryDataToFile(ctx context.Context, fileName string, data ...models.PrivateData) error {
	//TODO implement me
	panic("implement me")
}

func (p *privateDataFileStorage) LoadBinaryDataFromFile(ctx context.Context, fileName string) ([]models.PrivateData, error) {
	//TODO implement me
	panic("implement me")
}
