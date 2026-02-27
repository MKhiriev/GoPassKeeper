// Package service defines the core business logic interfaces and service
// implementations for the go-pass-keeper application.
package service

import (
	"context"
	"fmt"

	"github.com/MKhiriev/go-pass-keeper/internal/utils"
	"github.com/MKhiriev/go-pass-keeper/internal/validators"
	"github.com/MKhiriev/go-pass-keeper/models"
)

// privateDataValidationService is a middleware implementation of
// PrivateDataService that performs authorization and data validation checks
// before delegating to an inner PrivateDataService.
//
// It ensures that:
//   - the authenticated user (from context) matches the target user in the
//     request;
//   - all required fields are present;
//   - payloads satisfy domain-level validation rules enforced by Validator.
type privateDataValidationService struct {
	// inner is the next PrivateDataService in the middleware chain that
	// receives requests after successful validation.
	inner PrivateDataService

	// validator performs structural and semantic validation of private data
	// and request DTOs.
	validator validators.Validator
}

// NewPrivateDataValidationService constructs a PrivateDataServiceWrapper that
// decorates any PrivateDataService with validation and authorization checks.
//
// The returned wrapper uses validators.NewPrivateDataValidator() internally
// and is typically applied in NewPrivateDataService.
func NewPrivateDataValidationService() PrivateDataServiceWrapper {
	return &privateDataValidationService{
		validator: validators.NewPrivateDataValidator(),
	}
}

// UploadPrivateData validates the uploadRequest before delegating to the inner
// service:
//
//   - ensures at least one private data item is provided;
//   - ensures a user ID is present in the context;
//   - ensures every item belongs to the authenticated user;
//   - validates each item using the configured validator.
//
// Returns an error if any validation step fails, otherwise forwards the call
// to inner.UploadPrivateData.
func (v *privateDataValidationService) UploadPrivateData(ctx context.Context, uploadRequest models.UploadRequest) error {
	if len(uploadRequest.PrivateDataList) == 0 {
		return ErrValidationNoPrivateDataProvided
	}

	userID, found := utils.GetUserIDFromContext(ctx)
	if !found {
		return ErrValidationNoUserID
	}

	for _, data := range uploadRequest.PrivateDataList {
		if data.UserID != userID {
			return ErrUnauthorizedAccessToDifferentUserData
		}

		if err := v.validator.Validate(ctx, data); err != nil {
			return fmt.Errorf("%w: %w", ErrInvalidDataProvided, err)
		}
	}

	return v.inner.UploadPrivateData(ctx, uploadRequest)
}

// DownloadPrivateData validates the downloadRequests before delegating to the
// inner service:
//
//   - ensures a user ID is present in the context;
//   - ensures the request's UserID matches the authenticated user;
//   - validates the request structure using the validator.
//
// Returns a slice of PrivateData or an error if validation or storage fails.
func (v *privateDataValidationService) DownloadPrivateData(ctx context.Context, downloadRequests models.DownloadRequest) ([]models.PrivateData, error) {
	userID, found := utils.GetUserIDFromContext(ctx)
	if !found {
		return nil, ErrValidationNoUserID
	}

	if downloadRequests.UserID != userID {
		return nil, ErrUnauthorizedAccessToDifferentUserData
	}

	if err := v.validator.Validate(ctx, downloadRequests); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidDataProvided, err)
	}

	return v.inner.DownloadPrivateData(ctx, downloadRequests)
}

// DownloadAllPrivateData validates that the requested userID matches the
// authenticated user before delegating to the inner service.
//
// Returns all private data items for the user or an error if validation fails.
func (v *privateDataValidationService) DownloadAllPrivateData(ctx context.Context, userID int64) ([]models.PrivateData, error) {
	if userID == 0 {
		return nil, ErrValidationNoUserID
	}

	userIDFromAuthToken, found := utils.GetUserIDFromContext(ctx)
	if !found {
		return nil, ErrValidationNoUserID
	}

	if userIDFromAuthToken != userID {
		return nil, ErrUnauthorizedAccessToDifferentUserData
	}

	return v.inner.DownloadAllPrivateData(ctx, userID)
}

// DownloadUserPrivateDataStates validates that the requested userID matches
// the authenticated user before delegating to the inner service.
//
// Returns state descriptors for all private data items of the user or an error
// if validation fails.
func (v *privateDataValidationService) DownloadUserPrivateDataStates(ctx context.Context, userID int64) ([]models.PrivateDataState, error) {
	if userID == 0 {
		return nil, ErrValidationNoUserID
	}

	userIDFromAuthToken, found := utils.GetUserIDFromContext(ctx)
	if !found {
		return nil, ErrValidationNoUserID
	}

	if userIDFromAuthToken != userID {
		return nil, ErrUnauthorizedAccessToDifferentUserData
	}

	return v.inner.DownloadUserPrivateDataStates(ctx, userID)
}

// DownloadSpecificUserPrivateDataStates validates the syncRequest before
// delegating to the inner service:
//
//   - ensures a user ID is present in the context;
//   - ensures that at least one client-side ID is provided;
//   - ensures the request's UserID matches the authenticated user;
//   - ensures no empty client-side IDs are present.
//
// Returns a slice of PrivateDataState or an error if validation fails.
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
			return nil, ErrInvalidDataProvided
		}
	}

	return v.inner.DownloadSpecificUserPrivateDataStates(ctx, syncRequest)
}

// UpdatePrivateData validates the updateRequests before delegating to the
// inner service:
//
//   - ensures a user ID is present in the context;
//   - ensures the request's UserID matches the authenticated user;
//   - ensures at least one update entry is provided;
//   - validates each update entry using the validator.
//
// Returns an error if validation fails.
func (v *privateDataValidationService) UpdatePrivateData(ctx context.Context, updateRequests models.UpdateRequest) error {
	userID, found := utils.GetUserIDFromContext(ctx)
	if !found {
		return ErrValidationNoUserID
	}

	if updateRequests.UserID != userID {
		return ErrUnauthorizedAccessToDifferentUserData
	}

	if len(updateRequests.PrivateDataUpdates) == 0 {
		return ErrValidationNoUpdateRequestsProvided
	}

	for _, dataUpdate := range updateRequests.PrivateDataUpdates {
		if err := v.validator.Validate(ctx, dataUpdate); err != nil {
			return fmt.Errorf("%w: %w", ErrInvalidDataProvided, err)
		}
	}

	return v.inner.UpdatePrivateData(ctx, updateRequests)
}

// DeletePrivateData validates the deleteRequests before delegating to the
// inner service:
//
//   - ensures a user ID is present in the context;
//   - ensures the request's UserID matches the authenticated user;
//   - validates the request using the validator.
//
// Returns an error if validation fails.
func (v *privateDataValidationService) DeletePrivateData(ctx context.Context, deleteRequests models.DeleteRequest) error {
	userID, found := utils.GetUserIDFromContext(ctx)
	if !found {
		return ErrValidationNoUserID
	}

	if deleteRequests.UserID != userID {
		return ErrUnauthorizedAccessToDifferentUserData
	}

	if err := v.validator.Validate(ctx, deleteRequests); err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidDataProvided, err)
	}

	return v.inner.DeletePrivateData(ctx, deleteRequests)
}

// Wrap sets the inner PrivateDataService that this validation middleware will
// delegate to and returns the decorated service.
//
// It is typically used in a composition chain:
//
//	var baseSvc PrivateDataService = newCoreService(...)
//	var wrapper = NewPrivateDataValidationService()
//	svcWithValidation := wrapper.Wrap(baseSvc)
func (v *privateDataValidationService) Wrap(wrapper PrivateDataService) PrivateDataService {
	v.inner = wrapper
	return v
}
