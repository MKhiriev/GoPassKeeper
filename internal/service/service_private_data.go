// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Rasul Khiriev

package service

import (
	"context"

	"github.com/MKhiriev/go-pass-keeper/internal/config"
	"github.com/MKhiriev/go-pass-keeper/internal/logger"
	"github.com/MKhiriev/go-pass-keeper/internal/store"
	"github.com/MKhiriev/go-pass-keeper/models"
)

// privateDataService is the concrete implementation of PrivateDataService.
// It delegates all persistence operations to a PrivateDataStorage and is
// automatically wrapped with input-validation middleware by NewPrivateDataService.
type privateDataService struct {
	// privateDataRepository is the storage abstraction used to persist and
	// retrieve encrypted vault items.
	privateDataRepository store.PrivateDataStorage

	// logger is the structured logger used for diagnostic output.
	logger *logger.Logger
}

// NewPrivateDataService constructs a PrivateDataService that is pre-wrapped
// with validation middleware.
//
// Internally it creates a bare privateDataService and passes it through
// NewPrivateDataValidationService().Wrap(), so every public method call is
// validated before reaching the storage layer.
//
// The cfg parameter is accepted for future configuration needs (e.g. size
// limits) and is currently unused by the core service itself.
func NewPrivateDataService(privateDataRepository store.PrivateDataStorage, cfg config.App, logger *logger.Logger) PrivateDataService {
	service := &privateDataService{
		privateDataRepository: privateDataRepository,
		logger:                logger,
	}
	validationService := NewPrivateDataValidationService()

	return validationService.Wrap(service)
}

// UploadPrivateData persists all vault items in uploadRequest.PrivateDataList
// to the storage layer in a single call.
// Returns an error if the storage operation fails.
func (p *privateDataService) UploadPrivateData(ctx context.Context, uploadRequest models.UploadRequest) error {
	return p.privateDataRepository.Save(ctx, uploadRequest.PrivateDataList...)
}

// DownloadPrivateData retrieves the vault items identified by downloadRequests.
// Returns the matching items or an error if the storage query fails.
func (p *privateDataService) DownloadPrivateData(ctx context.Context, downloadRequests models.DownloadRequest) ([]models.PrivateData, error) {
	return p.privateDataRepository.Get(ctx, downloadRequests)
}

// DownloadAllPrivateData retrieves every vault item owned by userID.
// Returns the full collection or an error if the storage query fails.
func (p *privateDataService) DownloadAllPrivateData(ctx context.Context, userID int64) ([]models.PrivateData, error) {
	return p.privateDataRepository.GetAll(ctx, userID)
}

// DownloadUserPrivateDataStates returns lightweight state descriptors
// (client-side ID + version) for all vault items owned by userID.
// Clients use these states to detect which items are out of date and need
// to be re-downloaded.
// Returns the state list or an error if the storage query fails.
func (p *privateDataService) DownloadUserPrivateDataStates(ctx context.Context, userID int64) ([]models.PrivateDataState, error) {
	return p.privateDataRepository.GetAllStates(ctx, userID)
}

// DownloadSpecificUserPrivateDataStates returns state descriptors only for
// the vault items whose client-side IDs are listed in syncRequest.
// This is the targeted variant used during incremental sync operations.
// Returns the state list or an error if the storage query fails.
func (p *privateDataService) DownloadSpecificUserPrivateDataStates(ctx context.Context, syncRequest models.SyncRequest) ([]models.PrivateDataState, error) {
	return p.privateDataRepository.GetStates(ctx, syncRequest)
}

// UpdatePrivateData applies the batch of updates described by updateRequests
// to existing vault items in the storage layer.
// Returns an error if the storage operation fails.
func (p *privateDataService) UpdatePrivateData(ctx context.Context, updateRequests models.UpdateRequest) error {
	return p.privateDataRepository.Update(ctx, updateRequests)
}

// DeletePrivateData `soft deletes` the vault items listed in deleteRequests from the
// storage layer.
// Returns an error if the storage operation fails.
func (p *privateDataService) DeletePrivateData(ctx context.Context, deleteRequests models.DeleteRequest) error {
	return p.privateDataRepository.Delete(ctx, deleteRequests)
}
