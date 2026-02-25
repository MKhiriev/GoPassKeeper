package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/MKhiriev/go-pass-keeper/internal/logger"
	"github.com/MKhiriev/go-pass-keeper/internal/service"
	"github.com/MKhiriev/go-pass-keeper/internal/utils"
	"github.com/MKhiriev/go-pass-keeper/models"
)

type mockPrivateDataService struct {
	downloadUserStatesFn     func(ctx context.Context, userID int64) ([]models.PrivateDataState, error)
	downloadSpecificStatesFn func(ctx context.Context, req models.SyncRequest) ([]models.PrivateDataState, error)
}

func (m *mockPrivateDataService) UploadPrivateData(ctx context.Context, data models.UploadRequest) error {
	return nil
}
func (m *mockPrivateDataService) DownloadPrivateData(ctx context.Context, downloadRequests models.DownloadRequest) ([]models.PrivateData, error) {
	return nil, nil
}
func (m *mockPrivateDataService) DownloadAllPrivateData(ctx context.Context, userID int64) ([]models.PrivateData, error) {
	return nil, nil
}
func (m *mockPrivateDataService) DownloadUserPrivateDataStates(ctx context.Context, userID int64) ([]models.PrivateDataState, error) {
	return m.downloadUserStatesFn(ctx, userID)
}
func (m *mockPrivateDataService) DownloadSpecificUserPrivateDataStates(ctx context.Context, req models.SyncRequest) ([]models.PrivateDataState, error) {
	return m.downloadSpecificStatesFn(ctx, req)
}
func (m *mockPrivateDataService) UpdatePrivateData(ctx context.Context, updateRequests models.UpdateRequest) error {
	return nil
}
func (m *mockPrivateDataService) DeletePrivateData(ctx context.Context, deleteRequests models.DeleteRequest) error {
	return nil
}

func newHandlerWithPrivateDataService(pds service.PrivateDataService) *Handler {
	return &Handler{
		services: &service.Services{
			PrivateDataService: pds,
		},
		logger: logger.Nop(),
	}
}

func withUserID(ctx context.Context, userID int64) context.Context {
	return context.WithValue(ctx, utils.UserIDCtxKey, userID)
}

func normalizeStates(states []models.PrivateDataState) {
	for i := range states {
		if states[i].UpdatedAt != nil {
			t := states[i].UpdatedAt.UTC()
			states[i].UpdatedAt = &t
		}
	}
}

func TestGetClientServerDiff_Success(t *testing.T) {
	now := time.Unix(1700000000, 0).UTC()

	expected := []models.PrivateDataState{
		{
			ClientSideID: "id1",
			Hash:         "hash1",
			Version:      1,
			Deleted:      false,
			UpdatedAt:    &now,
		},
	}

	mockSvc := &mockPrivateDataService{
		downloadUserStatesFn: func(ctx context.Context, userID int64) ([]models.PrivateDataState, error) {
			return expected, nil
		},
	}

	h := newHandlerWithPrivateDataService(mockSvc)

	req := httptest.NewRequest(http.MethodGet, "/sync", nil)
	req = req.WithContext(withUserID(req.Context(), 1))

	rr := httptest.NewRecorder()
	h.getClientServerDiff(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var resp models.SyncResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}

	normalizeStates(resp.PrivateDataStates)
	normalizeStates(expected)

	if resp.Length != len(expected) {
		t.Fatalf("length mismatch")
	}

	if !reflect.DeepEqual(resp.PrivateDataStates, expected) {
		t.Fatalf("unexpected response body")
	}
}

func TestSyncSpecificUserData_Success(t *testing.T) {
	now := time.Unix(1700000000, 0).UTC()

	expected := []models.PrivateDataState{
		{
			ClientSideID: "id1",
			Hash:         "hash1",
			Version:      2,
			Deleted:      false,
			UpdatedAt:    &now,
		},
	}

	mockSvc := &mockPrivateDataService{
		downloadSpecificStatesFn: func(ctx context.Context, req models.SyncRequest) ([]models.PrivateDataState, error) {
			return expected, nil
		},
	}

	h := newHandlerWithPrivateDataService(mockSvc)

	body, _ := json.Marshal(models.SyncRequest{
		UserID:        1,
		ClientSideIDs: []string{"id1"},
		Length:        1,
	})

	req := httptest.NewRequest(http.MethodGet, "/sync/specific", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	h.syncSpecificUserData(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var resp models.SyncResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}

	normalizeStates(resp.PrivateDataStates)
	normalizeStates(expected)

	if resp.Length != len(expected) {
		t.Fatalf("length mismatch")
	}

	if !reflect.DeepEqual(resp.PrivateDataStates, expected) {
		t.Fatalf("unexpected response body")
	}
}

func TestGetClientServerDiff_ServiceError(t *testing.T) {
	mockSvc := &mockPrivateDataService{
		downloadUserStatesFn: func(ctx context.Context, userID int64) ([]models.PrivateDataState, error) {
			return nil, errors.New("service error")
		},
	}

	h := newHandlerWithPrivateDataService(mockSvc)

	req := httptest.NewRequest(http.MethodGet, "/sync", nil)
	req = req.WithContext(withUserID(req.Context(), 1))

	rr := httptest.NewRecorder()
	h.getClientServerDiff(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rr.Code)
	}
}

func TestSyncSpecificUserData_InvalidJSON(t *testing.T) {
	h := newHandlerWithPrivateDataService(&mockPrivateDataService{})

	req := httptest.NewRequest(http.MethodGet, "/sync/specific", bytes.NewBufferString("invalid"))
	rr := httptest.NewRecorder()

	h.syncSpecificUserData(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}
