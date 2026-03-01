// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Rasul Khiriev

package adapter

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MKhiriev/go-pass-keeper/internal/config"
	"github.com/MKhiriev/go-pass-keeper/internal/logger"
	"github.com/MKhiriev/go-pass-keeper/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestAdapter создаёт httpServerAdapter, направленный на тестовый сервер
func newTestAdapter(t *testing.T, serverURL string) *httpServerAdapter {
	t.Helper()
	log := logger.NewClientLogger("test")
	adapterCfg := config.ClientAdapter{HTTPAddress: serverURL}
	appCfg := config.ClientApp{HashKey: "testhashkey"}

	a, err := NewHTTPServerAdapter(adapterCfg, appCfg, log)
	require.NoError(t, err)
	return a.(*httpServerAdapter)
}

// ── Register ────────────────────────────────────────────────────────────────

func TestRegister_Success(t *testing.T) {
	user := models.User{Login: "alice", Name: "Alice"}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/api/auth/register", r.URL.Path)

		w.Header().Set("Authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxfQ.signature")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	a := newTestAdapter(t, srv.URL)
	got, err := a.Register(context.Background(), user)

	require.NoError(t, err)
	assert.Equal(t, user.Login, got.Login)
	assert.NotEmpty(t, a.Token())
}

func TestRegister_Conflict(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
		_, _ = w.Write([]byte("login already exists"))
	}))
	defer srv.Close()

	a := newTestAdapter(t, srv.URL)
	_, err := a.Register(context.Background(), models.User{Login: "alice"})

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrConflict)
}

func TestRegister_InternalServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("internal server error"))
	}))
	defer srv.Close()

	a := newTestAdapter(t, srv.URL)
	_, err := a.Register(context.Background(), models.User{Login: "alice"})

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInternalServerError)
}

// ── RequestSalt ──────────────────────────────────────────────────────────────

func TestRequestSalt_Success(t *testing.T) {
	want := models.User{Login: "alice", EncryptionSalt: "somesalt"}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/auth/params", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(want)
	}))
	defer srv.Close()

	a := newTestAdapter(t, srv.URL)
	got, err := a.RequestSalt(context.Background(), models.User{Login: "alice"})

	require.NoError(t, err)
	assert.Equal(t, want.Login, got.Login)
	assert.Equal(t, want.EncryptionSalt, got.EncryptionSalt)
}

func TestRequestSalt_Unauthorized(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("invalid login/password"))
	}))
	defer srv.Close()

	a := newTestAdapter(t, srv.URL)
	_, err := a.RequestSalt(context.Background(), models.User{Login: "alice"})

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrUnauthorized)
}

// ── Login ────────────────────────────────────────────────────────────────────

func TestLogin_Success(t *testing.T) {
	want := models.User{Login: "alice", Name: "Alice"}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/auth/login", r.URL.Path)
		w.Header().Set("Authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxfQ.signature")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(want)
	}))
	defer srv.Close()

	a := newTestAdapter(t, srv.URL)
	got, err := a.Login(context.Background(), models.User{Login: "alice"})

	require.NoError(t, err)
	assert.Equal(t, want.Login, got.Login)
	assert.NotEmpty(t, a.Token())
}

func TestLogin_Unauthorized(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("invalid login/password"))
	}))
	defer srv.Close()

	a := newTestAdapter(t, srv.URL)
	_, err := a.Login(context.Background(), models.User{Login: "alice"})

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrUnauthorized)
}

func TestLogin_BadGateway(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte("login on server failed"))
	}))
	defer srv.Close()

	a := newTestAdapter(t, srv.URL)
	_, err := a.Login(context.Background(), models.User{Login: "alice"})

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrBadGateway)
}

// ── Upload ───────────────────────────────────────────────────────────────────

func TestUpload_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/api/data/", r.URL.Path)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	a := newTestAdapter(t, srv.URL)
	a.SetToken("sometoken")

	err := a.Upload(context.Background(), models.UploadRequest{
		UserID:          1,
		PrivateDataList: []*models.PrivateData{},
	})
	require.NoError(t, err)
}

func TestUpload_Forbidden(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte("access denied"))
	}))
	defer srv.Close()

	a := newTestAdapter(t, srv.URL)
	err := a.Upload(context.Background(), models.UploadRequest{UserID: 1})

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrForbidden)
}

func TestUpload_Conflict(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
		_, _ = w.Write([]byte("version conflict"))
	}))
	defer srv.Close()

	a := newTestAdapter(t, srv.URL)
	err := a.Upload(context.Background(), models.UploadRequest{UserID: 1})

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrConflict)
}

// ── Download ─────────────────────────────────────────────────────────────────

func TestDownload_Success(t *testing.T) {
	want := []models.PrivateData{{ClientSideID: "abc-123", UserID: 1}}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/data/download", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(want)
	}))
	defer srv.Close()

	a := newTestAdapter(t, srv.URL)
	a.SetToken("sometoken")

	got, err := a.Download(context.Background(), models.DownloadRequest{
		UserID:        1,
		ClientSideIDs: []string{"abc-123"},
	})

	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, want[0].ClientSideID, got[0].ClientSideID)
}

func TestDownload_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("data not found"))
	}))
	defer srv.Close()

	a := newTestAdapter(t, srv.URL)
	_, err := a.Download(context.Background(), models.DownloadRequest{UserID: 1})

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrNotFound)
}

// ── Update ───────────────────────────────────────────────────────────────────

func TestUpdate_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)
		assert.Equal(t, "/api/data/update", r.URL.Path)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	a := newTestAdapter(t, srv.URL)
	a.SetToken("sometoken")

	err := a.Update(context.Background(), models.UpdateRequest{UserID: 1})
	require.NoError(t, err)
}

func TestUpdate_Conflict(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
		_, _ = w.Write([]byte("version conflict"))
	}))
	defer srv.Close()

	a := newTestAdapter(t, srv.URL)
	err := a.Update(context.Background(), models.UpdateRequest{UserID: 1})

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrConflict)
}

func TestUpdate_BadRequest(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("no update requests provided"))
	}))
	defer srv.Close()

	a := newTestAdapter(t, srv.URL)
	err := a.Update(context.Background(), models.UpdateRequest{UserID: 1})

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrBadRequest)
}

// ── Delete ───────────────────────────────────────────────────────────────────

func TestDelete_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		assert.Equal(t, "/api/data/delete", r.URL.Path)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	a := newTestAdapter(t, srv.URL)
	a.SetToken("sometoken")

	err := a.Delete(context.Background(), models.DeleteRequest{UserID: 1})
	require.NoError(t, err)
}

func TestDelete_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("data not found"))
	}))
	defer srv.Close()

	a := newTestAdapter(t, srv.URL)
	err := a.Delete(context.Background(), models.DeleteRequest{UserID: 1})

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrNotFound)
}

// ── GetServerStates ──────────────────────────────────────────────────────────

func TestGetServerStates_Success(t *testing.T) {
	want := models.SyncResponse{
		PrivateDataStates: []models.PrivateDataState{
			{ClientSideID: "abc-123", Version: 2},
		},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/api/sync/", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(want)
	}))
	defer srv.Close()

	a := newTestAdapter(t, srv.URL)
	a.SetToken("sometoken")

	got, err := a.GetServerStates(context.Background(), 1)

	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, want.PrivateDataStates[0].ClientSideID, got[0].ClientSideID)
	assert.Equal(t, want.PrivateDataStates[0].Version, got[0].Version)
}

func TestGetServerStates_Unauthorized(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("token is expired"))
	}))
	defer srv.Close()

	a := newTestAdapter(t, srv.URL)
	_, err := a.GetServerStates(context.Background(), 1)

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrUnauthorized)
}

// ── normalizeBaseURL ─────────────────────────────────────────────────────────

func TestNormalizeBaseURL(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"valid http", "http://localhost:8080", "http://localhost:8080", false},
		{"no scheme", "localhost:8080", "http://localhost:8080", false},
		{"trailing slash", "http://localhost:8080/", "http://localhost:8080", false},
		{"empty", "", "", true},
		{"no host", "http://", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizeBaseURL(tt.input)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
