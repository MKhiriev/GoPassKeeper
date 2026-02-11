package service

import (
	"context"

	"github.com/MKhiriev/go-pass-keeper/internal/config"
	"github.com/MKhiriev/go-pass-keeper/internal/logger"
	"github.com/MKhiriev/go-pass-keeper/internal/store"
	"github.com/MKhiriev/go-pass-keeper/models"
)

type privateDataService struct {
	privateDataRepository store.PrivateDataStorage

	logger *logger.Logger
}

func NewPrivateDataService(privateDataRepository store.PrivateDataStorage, cfg config.DBConfig, logger *logger.Logger) PrivateDataService {
	return &privateDataService{
		privateDataRepository: privateDataRepository,
		logger:                logger,
	}
}

func (p *privateDataService) UploadPrivateData(ctx context.Context, privateData models.PrivateData) error {
	return p.privateDataRepository.Save(ctx, privateData)
}

func (p *privateDataService) DownloadPrivateData(ctx context.Context, data models.PrivateData) (models.PrivateData, error) {
	//TODO implement me
	panic("implement me")
}

func (p *privateDataService) DownloadMultiplePrivateData(ctx context.Context, data []models.PrivateData) ([]models.PrivateData, error) {
	//TODO implement me
	panic("implement me")
}

func (p *privateDataService) DownloadAllPrivateData(ctx context.Context) ([]models.PrivateData, error) {
	//TODO implement me
	panic("implement me")
}

func (p *privateDataService) UpdatePrivateData(ctx context.Context, data models.PrivateData) error {
	//TODO implement me
	panic("implement me")
}

func (p *privateDataService) DeletePrivateData(ctx context.Context, data models.PrivateData) error {
	//TODO implement me
	panic("implement me")
}
