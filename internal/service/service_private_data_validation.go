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

func (v *privateDataValidationService) UploadPrivateData(ctx context.Context, uploadRequest models.UploadRequest) error {
	if len(uploadRequest.PrivateData) == 0 {
		return ErrValidationNoPrivateDataProvided
	}

	userID, found := utils.GetUserIDFromContext(ctx)
	if !found {
		return ErrValidationNoUserID
	}

	for _, data := range uploadRequest.PrivateData {
		if data.UserID == 0 {
			data.UserID = userID
		}

		if err := v.validator.Validate(ctx, data); err != nil {
			return fmt.Errorf("error during private data validation before saving: %w", err)
		}
	}

	return v.inner.UploadPrivateData(ctx, uploadRequest)
}

func (v *privateDataValidationService) DownloadPrivateData(ctx context.Context, downloadRequests models.DownloadRequest) ([]models.PrivateData, error) {
	userID, found := utils.GetUserIDFromContext(ctx)
	if !found {
		return nil, ErrValidationNoUserID
	}

	if err := v.validator.Validate(ctx, downloadRequests); err != nil {
		return nil, fmt.Errorf("error during download request validation before downloading: %w", err)
	}

	if downloadRequests.UserID != userID {
		return nil, ErrUnauthorizedAccessToDifferentUserData
	}

	return v.inner.DownloadPrivateData(ctx, downloadRequests)
}

func (v *privateDataValidationService) DownloadAllPrivateData(ctx context.Context, userID int64) ([]models.PrivateData, error) {
	if userID == 0 {
		return nil, ErrValidationNoUserID
	}

	return v.inner.DownloadAllPrivateData(ctx, userID)
}

func (v *privateDataValidationService) DownloadUserPrivateDataStates(ctx context.Context, userID int64) ([]models.PrivateDataState, error) {
	if userID == 0 {
		return nil, ErrValidationNoUserID
	}

	return v.inner.DownloadUserPrivateDataStates(ctx, userID)
}

func (v *privateDataValidationService) DownloadSpecificUserPrivateDataStates(ctx context.Context, syncRequest models.SyncRequest) ([]models.PrivateDataState, error) {
	userID, found := utils.GetUserIDFromContext(ctx)
	if !found {
		return nil, ErrValidationNoUserID
	}

	if len(syncRequest.ClientSideIDs) == 0 {
		return nil, ErrValidationNoClientIDsProvidedForSyncRequests
	}

	if syncRequest.UserID != userID {
		return nil, ErrUnauthorizedAccessToDifferentUserData
	}

	for _, request := range syncRequest.ClientSideIDs {
		if request == "" {
			return nil, ErrValidationEmptyClientIDProvidedForSyncRequests
		}
	}

	return v.inner.DownloadSpecificUserPrivateDataStates(ctx, syncRequest)
}

func (v *privateDataValidationService) UpdatePrivateData(ctx context.Context, updateRequests models.UpdateRequest) error {
	userID, found := utils.GetUserIDFromContext(ctx)
	if !found {
		return ErrValidationNoUserID
	}

	if len(updateRequests.PrivateDataUpdates) == 0 {
		return ErrValidationNoUpdateRequestsProvided
	}

	if updateRequests.UserID != userID {
		return ErrUnauthorizedAccessToDifferentUserData
	}

	for _, request := range updateRequests.PrivateDataUpdates {
		if err := v.validator.Validate(ctx, request); err != nil {
			return fmt.Errorf("error during update request validation before updating: %w", err)
		}
	}

	return v.inner.UpdatePrivateData(ctx, updateRequests)
}

func (v *privateDataValidationService) DeletePrivateData(ctx context.Context, deleteRequests models.DeleteRequest) error {
	userID, found := utils.GetUserIDFromContext(ctx)
	if !found {
		return ErrValidationNoUserID
	}

	if err := v.validator.Validate(ctx, deleteRequests); err != nil {
		return fmt.Errorf("error during delete request validation before deleting: %w", err)
	}

	if deleteRequests.UserID != userID {
		return ErrUnauthorizedAccessToDifferentUserData
	}

	return v.inner.DeletePrivateData(ctx, deleteRequests)
}

func (v *privateDataValidationService) Wrap(wrapper PrivateDataService) PrivateDataService {
	v.inner = wrapper
	return v
}
