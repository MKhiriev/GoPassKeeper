package service

import (
	"context"
	"fmt"

	"github.com/MKhiriev/go-pass-keeper/internal/validators"
	"github.com/MKhiriev/go-pass-keeper/models"
)

type privateDataValidationService struct {
	inner     PrivateDataService
	validator validators.Validator
}

func NewPrivateDataValidationService() PrivateDataServiceWrapper {
	return &privateDataValidationService{
		validator: validators.NewPrivateDataValidator(),
	}
}

func (v *privateDataValidationService) UploadPrivateData(ctx context.Context, privateData models.PrivateData) error {
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

func (v *privateDataValidationService) DownloadPrivateData(ctx context.Context, data models.PrivateData) (models.PrivateData, error) {
	//TODO implement me
	panic("implement me")
}

func (v *privateDataValidationService) DownloadMultiplePrivateData(ctx context.Context, data []models.PrivateData) ([]models.PrivateData, error) {
	//TODO implement me
	panic("implement me")
}

func (v *privateDataValidationService) DownloadAllPrivateData(ctx context.Context) ([]models.PrivateData, error) {
	//TODO implement me
	panic("implement me")
}

func (v *privateDataValidationService) UpdatePrivateData(ctx context.Context, data models.PrivateData) error {
	//TODO implement me
	panic("implement me")
}

func (v *privateDataValidationService) DeletePrivateData(ctx context.Context, data models.PrivateData) error {
	//TODO implement me
	panic("implement me")
}

func (v *privateDataValidationService) Wrap(wrapper PrivateDataService) PrivateDataService {
	v.inner = wrapper
	return v
}
