package service

import (
	"context"
	"fmt"

	"github.com/MKhiriev/go-pass-keeper/internal/utils"
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

func (v *privateDataValidationService) UploadPrivateData(ctx context.Context, privateData ...models.PrivateData) error {
	if len(privateData) == 0 {
		return ErrValidationNoPrivateDataProvided
	}

	userID, found := utils.GetUserIDFromContext(ctx)
	if !found {
		return ErrValidationNoUserID
	}

	for _, data := range privateData {
		if err := v.validator.Validate(ctx, data); err != nil {
			return fmt.Errorf("error during private data validation before saving: %w", err)
		}

		data.UserID = userID
	}

	return v.inner.UploadPrivateData(ctx, privateData...)
}

func (v *privateDataValidationService) DownloadPrivateData(ctx context.Context, downloadRequests ...models.DownloadRequest) ([]models.PrivateData, error) {
	if len(downloadRequests) == 0 {
		return nil, ErrValidationNoDownloadRequestsProvided
	}

	userID, found := utils.GetUserIDFromContext(ctx)
	if !found {
		return nil, ErrValidationNoUserID
	}

	for _, request := range downloadRequests {
		if err := v.validator.Validate(ctx, request); err != nil {
			return nil, fmt.Errorf("error during download request validation before saving: %w", err)
		}

		request.UserID = userID
	}

	return v.inner.DownloadPrivateData(ctx, downloadRequests...)
}

func (v *privateDataValidationService) DownloadAllPrivateData(ctx context.Context, userID int64) ([]models.PrivateData, error) {
	if userID == 0 {
		return nil, ErrValidationNoUserID
	}

	return v.inner.DownloadAllPrivateData(ctx, userID)
}

func (v *privateDataValidationService) UpdatePrivateData(ctx context.Context, updateRequests ...models.UpdateRequest) error {
	if len(updateRequests) == 0 {
		return ErrValidationNoUpdateRequestsProvided
	}

	userID, found := utils.GetUserIDFromContext(ctx)
	if !found {
		return ErrValidationNoUserID
	}

	for _, request := range updateRequests {
		if err := v.validator.Validate(ctx, request); err != nil {
			return fmt.Errorf("error during download request validation before saving: %w", err)
		}

		request.UserID = userID
	}

	return v.inner.UpdatePrivateData(ctx, updateRequests...)
}

func (v *privateDataValidationService) DeletePrivateData(ctx context.Context, deleteRequests ...models.DeleteRequest) error {
	if len(deleteRequests) == 0 {
		return ErrValidationNoDeleteRequestsProvided
	}

	userID, found := utils.GetUserIDFromContext(ctx)
	if !found {
		return ErrValidationNoUserID
	}

	for _, request := range deleteRequests {
		if err := v.validator.Validate(ctx, request); err != nil {
			return fmt.Errorf("error during download request validation before saving: %w", err)
		}

		request.UserID = userID
	}

	return v.inner.DeletePrivateData(ctx, deleteRequests...)
}

func (v *privateDataValidationService) Wrap(wrapper PrivateDataService) PrivateDataService {
	v.inner = wrapper
	return v
}
