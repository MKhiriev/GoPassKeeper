package service

import (
	"context"
	"errors"
	"testing"

	"github.com/MKhiriev/go-pass-keeper/internal/utils"
	"github.com/MKhiriev/go-pass-keeper/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ─────────────────────────────────────────────
// Mocks
// ─────────────────────────────────────────────

type mockInnerService struct {
	uploadFn           func(ctx context.Context, req models.UploadRequest) error
	downloadFn         func(ctx context.Context, req models.DownloadRequest) ([]models.PrivateData, error)
	downloadAllFn      func(ctx context.Context, userID int64) ([]models.PrivateData, error)
	downloadStatesFn   func(ctx context.Context, userID int64) ([]models.PrivateDataState, error)
	downloadSpecificFn func(ctx context.Context, req models.SyncRequest) ([]models.PrivateDataState, error)
	updateFn           func(ctx context.Context, req models.UpdateRequest) error
	deleteFn           func(ctx context.Context, req models.DeleteRequest) error
}

func (m *mockInnerService) UploadPrivateData(ctx context.Context, req models.UploadRequest) error {
	if m.uploadFn != nil {
		return m.uploadFn(ctx, req)
	}
	return nil
}
func (m *mockInnerService) DownloadPrivateData(ctx context.Context, req models.DownloadRequest) ([]models.PrivateData, error) {
	if m.downloadFn != nil {
		return m.downloadFn(ctx, req)
	}
	return nil, nil
}
func (m *mockInnerService) DownloadAllPrivateData(ctx context.Context, userID int64) ([]models.PrivateData, error) {
	if m.downloadAllFn != nil {
		return m.downloadAllFn(ctx, userID)
	}
	return nil, nil
}
func (m *mockInnerService) DownloadUserPrivateDataStates(ctx context.Context, userID int64) ([]models.PrivateDataState, error) {
	if m.downloadStatesFn != nil {
		return m.downloadStatesFn(ctx, userID)
	}
	return nil, nil
}
func (m *mockInnerService) DownloadSpecificUserPrivateDataStates(ctx context.Context, req models.SyncRequest) ([]models.PrivateDataState, error) {
	if m.downloadSpecificFn != nil {
		return m.downloadSpecificFn(ctx, req)
	}
	return nil, nil
}
func (m *mockInnerService) UpdatePrivateData(ctx context.Context, req models.UpdateRequest) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, req)
	}
	return nil
}
func (m *mockInnerService) DeletePrivateData(ctx context.Context, req models.DeleteRequest) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, req)
	}
	return nil
}

type mockValidator struct {
	validateFn func(ctx context.Context, i any, fields ...string) error
}

func (m *mockValidator) Validate(ctx context.Context, i any, fields ...string) error {
	if m.validateFn != nil {
		return m.validateFn(ctx, i, fields...)
	}
	return nil
}

// ─────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────

func newValidationService(inner PrivateDataService, v *mockValidator) *privateDataValidationService {
	return &privateDataValidationService{
		inner:     inner,
		validator: v,
	}
}

func ctxWithUserID(id int64) context.Context {
	return context.WithValue(context.Background(), utils.UserIDCtxKey, id)
}

var errValidation = errors.New("validation failed")

// ─────────────────────────────────────────────
// UploadPrivateData
// ─────────────────────────────────────────────

func TestValidation_UploadPrivateData_NoData(t *testing.T) {
	svc := newValidationService(nil, nil)
	err := svc.UploadPrivateData(context.Background(), models.UploadRequest{PrivateDataList: nil})
	assert.ErrorIs(t, err, ErrValidationNoPrivateDataProvided)
}

func TestValidation_UploadPrivateData_NoUserIDInCtx(t *testing.T) {
	svc := newValidationService(nil, nil)
	req := models.UploadRequest{PrivateDataList: []*models.PrivateData{{}}}
	err := svc.UploadPrivateData(context.Background(), req)
	assert.ErrorIs(t, err, ErrValidationNoUserID)
}

func TestValidation_UploadPrivateData_Unauthorized(t *testing.T) {
	svc := newValidationService(nil, nil)
	req := models.UploadRequest{
		PrivateDataList: []*models.PrivateData{{UserID: 999}}, // different from ctx
	}
	err := svc.UploadPrivateData(ctxWithUserID(1), req)
	assert.ErrorIs(t, err, ErrUnauthorizedAccessToDifferentUserData)
}

func TestValidation_UploadPrivateData_ValidatorError(t *testing.T) {
	v := &mockValidator{
		validateFn: func(_ context.Context, _ any, _ ...string) error { return errValidation },
	}
	svc := newValidationService(nil, v)
	req := models.UploadRequest{
		PrivateDataList: []*models.PrivateData{{UserID: 1}},
	}
	err := svc.UploadPrivateData(ctxWithUserID(1), req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "error during private data validation")
	assert.True(t, errors.Is(err, errValidation))
}

func TestValidation_UploadPrivateData_Success(t *testing.T) {
	called := false
	inner := &mockInnerService{
		uploadFn: func(_ context.Context, _ models.UploadRequest) error {
			called = true
			return nil
		},
	}
	svc := newValidationService(inner, &mockValidator{})
	req := models.UploadRequest{
		PrivateDataList: []*models.PrivateData{{UserID: 1}},
	}
	err := svc.UploadPrivateData(ctxWithUserID(1), req)
	assert.NoError(t, err)
	assert.True(t, called)
}

// ─────────────────────────────────────────────
// DownloadPrivateData
// ─────────────────────────────────────────────

func TestValidation_DownloadPrivateData_Unauthorized(t *testing.T) {
	svc := newValidationService(nil, nil)
	req := models.DownloadRequest{UserID: 999}
	_, err := svc.DownloadPrivateData(ctxWithUserID(1), req)
	assert.ErrorIs(t, err, ErrUnauthorizedAccessToDifferentUserData)
}

func TestValidation_DownloadPrivateData_ValidatorError(t *testing.T) {
	v := &mockValidator{
		validateFn: func(_ context.Context, _ any, _ ...string) error { return errValidation },
	}
	svc := newValidationService(nil, v)
	req := models.DownloadRequest{UserID: 1}
	_, err := svc.DownloadPrivateData(ctxWithUserID(1), req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "error during download request validation")
}

// ─────────────────────────────────────────────
// DownloadAllPrivateData
// ─────────────────────────────────────────────

func TestValidation_DownloadAllPrivateData_UserIDZero(t *testing.T) {
	svc := newValidationService(nil, nil)
	_, err := svc.DownloadAllPrivateData(context.Background(), 0)
	assert.ErrorIs(t, err, ErrValidationNoUserID)
}

func TestValidation_DownloadAllPrivateData_Unauthorized(t *testing.T) {
	svc := newValidationService(nil, nil)
	_, err := svc.DownloadAllPrivateData(ctxWithUserID(1), 2)
	assert.ErrorIs(t, err, ErrUnauthorizedAccessToDifferentUserData)
}

// ─────────────────────────────────────────────
// DownloadSpecificUserPrivateDataStates
// ─────────────────────────────────────────────

func TestValidation_DownloadSpecific_NoIDs(t *testing.T) {
	svc := newValidationService(nil, nil)
	req := models.SyncRequest{UserID: 1, ClientSideIDs: nil}
	_, err := svc.DownloadSpecificUserPrivateDataStates(ctxWithUserID(1), req)
	assert.ErrorIs(t, err, ErrValidationNoClientIDsProvidedForSyncRequests)
}

func TestValidation_DownloadSpecific_EmptyIDInList(t *testing.T) {
	svc := newValidationService(nil, nil)
	req := models.SyncRequest{UserID: 1, ClientSideIDs: []string{"id1", ""}}
	_, err := svc.DownloadSpecificUserPrivateDataStates(ctxWithUserID(1), req)
	assert.ErrorIs(t, err, ErrValidationEmptyClientIDProvidedForSyncRequests)
}

// ─────────────────────────────────────────────
// UpdatePrivateData
// ─────────────────────────────────────────────

func TestValidation_UpdatePrivateData_NoUpdates(t *testing.T) {
	svc := newValidationService(nil, nil)
	req := models.UpdateRequest{UserID: 1, PrivateDataUpdates: nil}
	err := svc.UpdatePrivateData(ctxWithUserID(1), req)
	assert.ErrorIs(t, err, ErrValidationNoUpdateRequestsProvided)
}

func TestValidation_UpdatePrivateData_ValidatorError(t *testing.T) {
	v := &mockValidator{
		validateFn: func(_ context.Context, _ any, _ ...string) error { return errValidation },
	}
	svc := newValidationService(nil, v)
	req := models.UpdateRequest{
		UserID:             1,
		PrivateDataUpdates: []models.PrivateDataUpdate{{}},
	}
	err := svc.UpdatePrivateData(ctxWithUserID(1), req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "error during update request validation")
}

// ─────────────────────────────────────────────
// DeletePrivateData
// ─────────────────────────────────────────────

func TestValidation_DeletePrivateData_ValidatorError(t *testing.T) {
	v := &mockValidator{
		validateFn: func(_ context.Context, _ any, _ ...string) error { return errValidation },
	}
	svc := newValidationService(nil, v)
	req := models.DeleteRequest{UserID: 1}
	err := svc.DeletePrivateData(ctxWithUserID(1), req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "error during delete request validation")
}

func TestValidation_DeletePrivateData_Success(t *testing.T) {
	called := false
	inner := &mockInnerService{
		deleteFn: func(_ context.Context, _ models.DeleteRequest) error {
			called = true
			return nil
		},
	}
	svc := newValidationService(inner, &mockValidator{})
	req := models.DeleteRequest{UserID: 1}
	err := svc.DeletePrivateData(ctxWithUserID(1), req)
	assert.NoError(t, err)
	assert.True(t, called)
}
