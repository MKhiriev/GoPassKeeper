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

func NewPrivateDataService(privateDataRepository store.PrivateDataStorage, cfg config.App, logger *logger.Logger) PrivateDataService {
	service := &privateDataService{
		privateDataRepository: privateDataRepository,
		logger:                logger,
	}
	validationService := NewPrivateDataValidationService()

	return validationService.Wrap(service)
}

func (p *privateDataService) UploadPrivateData(ctx context.Context, uploadRequest models.UploadRequest) error {
	return p.privateDataRepository.Save(ctx, uploadRequest.PrivateData...)
}

func (p *privateDataService) DownloadPrivateData(ctx context.Context, downloadRequests models.DownloadRequest) ([]models.PrivateData, error) {
	return p.privateDataRepository.Get(ctx, downloadRequests)
}

func (p *privateDataService) DownloadAllPrivateData(ctx context.Context, userID int64) ([]models.PrivateData, error) {
	return p.privateDataRepository.GetAll(ctx, userID)
}

func (p *privateDataService) DownloadUserPrivateDataStates(ctx context.Context, userID int64) ([]models.PrivateDataState, error) {
	return p.privateDataRepository.GetAllStates(ctx, userID)
}

func (p *privateDataService) DownloadSpecificUserPrivateDataStates(ctx context.Context, syncRequest models.SyncRequest) ([]models.PrivateDataState, error) {
	return p.privateDataRepository.GetStates(ctx, syncRequest)
}

func (p *privateDataService) UpdatePrivateData(ctx context.Context, updateRequests models.UpdateRequest) error {
	return p.privateDataRepository.Update(ctx, updateRequests)
}

func (p *privateDataService) DeletePrivateData(ctx context.Context, deleteRequests models.DeleteRequest) error {
	return p.privateDataRepository.Delete(ctx, deleteRequests)
}
