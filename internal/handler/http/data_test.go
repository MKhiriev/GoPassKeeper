package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/MKhiriev/go-pass-keeper/internal/logger"
	"github.com/MKhiriev/go-pass-keeper/internal/service"
	"github.com/MKhiriev/go-pass-keeper/internal/utils"
	"github.com/MKhiriev/go-pass-keeper/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ─────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────

// newHandlerForData builds a Handler with the given PrivateDataService mock,
// reusing the mocks already declared in routes_test.go.
func newHandlerForData(t *testing.T, svc service.PrivateDataService) *Handler {
	t.Helper()
	return newTestRouter_handler(t, svc)
}

// newTestRouter_handler returns a *Handler (not http.Handler) so individual
// handler methods can be called directly without going through the router.
func newTestRouter_handler(t *testing.T, svc service.PrivateDataService) *Handler {
	t.Helper()
	h := &Handler{
		logger: logger.Nop(),
		services: &service.Services{
			AuthService:        &mockAuthSvc{},
			AppInfoService:     &mockAppInfoSvc{},
			PrivateDataService: svc,
		},
	}
	return h
}

// encodeBody serialises v to JSON and returns it as an io.Reader.
func encodeBody(t *testing.T, v any) io.Reader {
	t.Helper()
	buf := &bytes.Buffer{}
	require.NoError(t, json.NewEncoder(buf).Encode(v))
	return buf
}

// ctxWithUser returns a context carrying the given userID.
// Declared here only if not already declared in another test file in the package.
// If ctxWithUserID is already declared in sync_test.go, remove this one.
func ctxWithUser(userID int64) context.Context {
	return context.WithValue(context.Background(), utils.UserIDCtxKey, userID)
}

// ─────────────────────────────────────────────
// upload
// ─────────────────────────────────────────────

func TestUpload_Success(t *testing.T) {
	called := false
	svc := &mockPrivateDataSvc{
		uploadFn: func(_ context.Context, req models.UploadRequest) error {
			called = true
			assert.Equal(t, int64(1), req.UserID)
			assert.Equal(t, 1, req.Length)
			return nil
		},
	}

	h := newHandlerForData(t, svc)
	body := models.UploadRequest{
		UserID:          1,
		PrivateDataList: []*models.PrivateData{{ClientSideID: "abc"}},
		Length:          1,
	}
	req := httptest.NewRequest(http.MethodPost, "/api/data/", encodeBody(t, body))
	rec := httptest.NewRecorder()

	h.upload(rec, req)

	assert.True(t, called, "UploadPrivateData should have been called")
	assert.Equal(t, http.StatusCreated, rec.Code)
}

func TestUpload_InvalidJSON(t *testing.T) {
	h := newHandlerForData(t, &mockPrivateDataSvc{})
	req := httptest.NewRequest(http.MethodPost, "/api/data/", strings.NewReader(`{bad json}`))
	rec := httptest.NewRecorder()

	h.upload(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "Invalid JSON was passed")
}

func TestUpload_EmptyBody(t *testing.T) {
	h := newHandlerForData(t, &mockPrivateDataSvc{})
	req := httptest.NewRequest(http.MethodPost, "/api/data/", strings.NewReader(""))
	rec := httptest.NewRecorder()

	h.upload(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestUpload_ServiceError(t *testing.T) {
	svc := &mockPrivateDataSvc{
		uploadFn: func(_ context.Context, _ models.UploadRequest) error {
			return errors.New("storage failure")
		},
	}

	h := newHandlerForData(t, svc)
	req := httptest.NewRequest(http.MethodPost, "/api/data/", encodeBody(t, models.UploadRequest{UserID: 1}))
	rec := httptest.NewRecorder()

	h.upload(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Contains(t, rec.Body.String(), "error uploading private data")
}

// ─────────────────────────────────────────────
// downloadMultiple
// ─────────────────────────────────────────────

func TestDownloadMultiple_Success(t *testing.T) {
	expected := []models.PrivateData{{ClientSideID: "id-1"}, {ClientSideID: "id-2"}}
	svc := &mockPrivateDataSvc{
		downloadFn: func(_ context.Context, req models.DownloadRequest) ([]models.PrivateData, error) {
			assert.Equal(t, int64(5), req.UserID)
			assert.Equal(t, []string{"id-1", "id-2"}, req.ClientSideIDs)
			return expected, nil
		},
	}

	h := newHandlerForData(t, svc)
	body := models.DownloadRequest{
		UserID:        5,
		ClientSideIDs: []string{"id-1", "id-2"},
		Length:        2,
	}
	req := httptest.NewRequest(http.MethodPost, "/api/data/download", encodeBody(t, body))
	rec := httptest.NewRecorder()

	h.downloadMultiple(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	var result []models.PrivateData
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&result))
	assert.Equal(t, expected, result)
}

func TestDownloadMultiple_EmptyResult(t *testing.T) {
	svc := &mockPrivateDataSvc{
		downloadFn: func(_ context.Context, _ models.DownloadRequest) ([]models.PrivateData, error) {
			return []models.PrivateData{}, nil
		},
	}

	h := newHandlerForData(t, svc)
	req := httptest.NewRequest(http.MethodPost, "/api/data/download",
		encodeBody(t, models.DownloadRequest{UserID: 1}))
	rec := httptest.NewRecorder()

	h.downloadMultiple(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	var result []models.PrivateData
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&result))
	assert.Empty(t, result)
}

func TestDownloadMultiple_InvalidJSON(t *testing.T) {
	h := newHandlerForData(t, &mockPrivateDataSvc{})
	req := httptest.NewRequest(http.MethodPost, "/api/data/download", strings.NewReader(`{bad json}`))
	rec := httptest.NewRecorder()

	h.downloadMultiple(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "Invalid JSON was passed")
}

func TestDownloadMultiple_ServiceError(t *testing.T) {
	svc := &mockPrivateDataSvc{
		downloadFn: func(_ context.Context, _ models.DownloadRequest) ([]models.PrivateData, error) {
			return nil, errors.New("db unavailable")
		},
	}

	h := newHandlerForData(t, svc)
	req := httptest.NewRequest(http.MethodPost, "/api/data/download",
		encodeBody(t, models.DownloadRequest{UserID: 1}))
	rec := httptest.NewRecorder()

	h.downloadMultiple(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Contains(t, rec.Body.String(), "error downloading private data")
}

// ─────────────────────────────────────────────
// downloadAllUserData
// ─────────────────────────────────────────────

func TestDownloadAllUserData_Success(t *testing.T) {
	expected := []models.PrivateData{{ClientSideID: "all-1"}, {ClientSideID: "all-2"}}
	svc := &mockPrivateDataSvc{
		downloadAllFn: func(_ context.Context, userID int64) ([]models.PrivateData, error) {
			assert.Equal(t, int64(42), userID)
			return expected, nil
		},
	}

	h := newHandlerForData(t, svc)
	req := httptest.NewRequest(http.MethodGet, "/api/data/all", nil).
		WithContext(ctxWithUser(42))
	rec := httptest.NewRecorder()

	h.downloadAllUserData(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	var result []models.PrivateData
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&result))
	assert.Equal(t, expected, result)
}

func TestDownloadAllUserData_EmptyResult(t *testing.T) {
	svc := &mockPrivateDataSvc{
		downloadAllFn: func(_ context.Context, _ int64) ([]models.PrivateData, error) {
			return []models.PrivateData{}, nil
		},
	}

	h := newHandlerForData(t, svc)
	req := httptest.NewRequest(http.MethodGet, "/api/data/all", nil).
		WithContext(ctxWithUser(1))
	rec := httptest.NewRecorder()

	h.downloadAllUserData(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	var result []models.PrivateData
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&result))
	assert.Empty(t, result)
}

func TestDownloadAllUserData_NoUserID(t *testing.T) {
	h := newHandlerForData(t, &mockPrivateDataSvc{})
	req := httptest.NewRequest(http.MethodGet, "/api/data/all", nil) // no userID in context
	rec := httptest.NewRecorder()

	h.downloadAllUserData(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "no user ID was given")
}

func TestDownloadAllUserData_ServiceError(t *testing.T) {
	svc := &mockPrivateDataSvc{
		downloadAllFn: func(_ context.Context, _ int64) ([]models.PrivateData, error) {
			return nil, errors.New("db unavailable")
		},
	}

	h := newHandlerForData(t, svc)
	req := httptest.NewRequest(http.MethodGet, "/api/data/all", nil).
		WithContext(ctxWithUser(1))
	rec := httptest.NewRecorder()

	h.downloadAllUserData(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Contains(t, rec.Body.String(), "error downloading private data")
}

// ─────────────────────────────────────────────
// update
// ─────────────────────────────────────────────

func TestUpdate_Success(t *testing.T) {
	called := false
	svc := &mockPrivateDataSvc{
		updateFn: func(_ context.Context, req models.UpdateRequest) error {
			called = true
			assert.Equal(t, int64(7), req.UserID)
			assert.Equal(t, 1, req.Length)
			return nil
		},
	}

	h := newHandlerForData(t, svc)
	body := models.UpdateRequest{
		UserID: 7,
		PrivateDataUpdates: []models.PrivateDataUpdate{
			{ClientSideID: "upd-1", Version: 2},
		},
		Length: 1,
	}
	req := httptest.NewRequest(http.MethodPut, "/api/data/update", encodeBody(t, body))
	rec := httptest.NewRecorder()

	h.update(rec, req)

	assert.True(t, called, "UpdatePrivateData should have been called")
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestUpdate_InvalidJSON(t *testing.T) {
	h := newHandlerForData(t, &mockPrivateDataSvc{})
	req := httptest.NewRequest(http.MethodPut, "/api/data/update", strings.NewReader(`{bad json}`))
	rec := httptest.NewRecorder()

	h.update(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "Invalid JSON was passed")
}

func TestUpdate_EmptyBody(t *testing.T) {
	h := newHandlerForData(t, &mockPrivateDataSvc{})
	req := httptest.NewRequest(http.MethodPut, "/api/data/update", strings.NewReader(""))
	rec := httptest.NewRecorder()

	h.update(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestUpdate_ServiceError(t *testing.T) {
	svc := &mockPrivateDataSvc{
		updateFn: func(_ context.Context, _ models.UpdateRequest) error {
			return errors.New("version conflict")
		},
	}

	h := newHandlerForData(t, svc)
	req := httptest.NewRequest(http.MethodPut, "/api/data/update",
		encodeBody(t, models.UpdateRequest{UserID: 1}))
	rec := httptest.NewRecorder()

	h.update(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Contains(t, rec.Body.String(), "error updating private data")
}

// ─────────────────────────────────────────────
// delete
// ─────────────────────────────────────────────

func TestDelete_Success(t *testing.T) {
	called := false
	svc := &mockPrivateDataSvc{
		deleteFn: func(_ context.Context, req models.DeleteRequest) error {
			called = true
			assert.Equal(t, int64(3), req.UserID)
			assert.Equal(t, 1, req.Length)
			return nil
		},
	}

	h := newHandlerForData(t, svc)
	body := models.DeleteRequest{
		UserID: 3,
		DeleteEntries: []models.DeleteEntry{
			{ClientSideID: "del-1", Version: 1},
		},
		Length: 1,
	}
	req := httptest.NewRequest(http.MethodDelete, "/api/data/delete", encodeBody(t, body))
	rec := httptest.NewRecorder()

	h.delete(rec, req)

	assert.True(t, called, "DeletePrivateData should have been called")
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestDelete_InvalidJSON(t *testing.T) {
	h := newHandlerForData(t, &mockPrivateDataSvc{})
	req := httptest.NewRequest(http.MethodDelete, "/api/data/delete", strings.NewReader(`{bad json}`))
	rec := httptest.NewRecorder()

	h.delete(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "Invalid JSON was passed")
}

func TestDelete_EmptyBody(t *testing.T) {
	h := newHandlerForData(t, &mockPrivateDataSvc{})
	req := httptest.NewRequest(http.MethodDelete, "/api/data/delete", strings.NewReader(""))
	rec := httptest.NewRecorder()

	h.delete(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestDelete_ServiceError(t *testing.T) {
	svc := &mockPrivateDataSvc{
		deleteFn: func(_ context.Context, _ models.DeleteRequest) error {
			return errors.New("not found")
		},
	}

	h := newHandlerForData(t, svc)
	req := httptest.NewRequest(http.MethodDelete, "/api/data/delete",
		encodeBody(t, models.DeleteRequest{UserID: 1}))
	rec := httptest.NewRecorder()

	h.delete(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Contains(t, rec.Body.String(), "error deleting private data")
}
