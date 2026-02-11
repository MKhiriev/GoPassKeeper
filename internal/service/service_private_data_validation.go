package service

import (
	"context"
	"fmt"

	"github.com/MKhiriev/go-pass-keeper/internal/validators"
	"github.com/MKhiriev/go-pass-keeper/models"
)

type PrivateDataValidationService struct {
	inner     PrivateDataService
	validator validators.Validator
}

func NewPrivateDataValidationService() PrivateDataServiceWrapper {
	return &PrivateDataValidationService{
		validator: validators.NewPrivateDataValidator(),
	}
}

func (v *PrivateDataValidationService) UploadPrivateData(ctx context.Context, privateData models.PrivateData) error {
	// data in json should consist of:
	//  - Metadata
	//  - Type
	//  - Data
	//  - (not always) Notes
	//  - (not always) Additional Fields
	if err := v.validator.Validate(ctx, privateData); err != nil {
		return fmt.Errorf("error during private data validation before saving: %w", err)
	}

	// todo get user id from context and add to private data user id

	return v.inner.UploadPrivateData(ctx, privateData)
}

func (v *PrivateDataValidationService) DownloadPrivateData(ctx context.Context, data models.PrivateData) (models.PrivateData, error) {
	//TODO implement me
	panic("implement me")
}

func (v *PrivateDataValidationService) DownloadMultiplePrivateData(ctx context.Context, data []models.PrivateData) ([]models.PrivateData, error) {
	//TODO implement me
	panic("implement me")
}

func (v *PrivateDataValidationService) DownloadAllPrivateData(ctx context.Context) ([]models.PrivateData, error) {
	//TODO implement me
	panic("implement me")
}

func (v *PrivateDataValidationService) UpdatePrivateData(ctx context.Context, data models.PrivateData) error {
	//TODO implement me
	panic("implement me")
}

func (v *PrivateDataValidationService) DeletePrivateData(ctx context.Context, data models.PrivateData) error {
	//TODO implement me
	panic("implement me")
}

func (v *PrivateDataValidationService) Wrap(wrapper PrivateDataService) PrivateDataService {
	v.inner = wrapper
	return v
}
