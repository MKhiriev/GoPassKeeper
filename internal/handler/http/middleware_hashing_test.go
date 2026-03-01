// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Rasul Khiriev

package http

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/MKhiriev/go-pass-keeper/internal/utils"
	"github.com/MKhiriev/go-pass-keeper/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Helpers ---

func makeUploadBody(t *testing.T, list []*models.PrivateData, hash string) []byte {
	t.Helper()
	body, err := json.Marshal(struct {
		PrivateDataList []*models.PrivateData `json:"private_data_list"`
		Hash            string                `json:"hash"`
	}{
		PrivateDataList: list,
		Hash:            hash,
	})
	require.NoError(t, err)
	return body
}

func makeUpdateBody(t *testing.T, updates []models.PrivateDataUpdate, hash string) []byte {
	t.Helper()
	body, err := json.Marshal(struct {
		PrivateDataUpdates []models.PrivateDataUpdate `json:"private_data_updates"`
		Hash               string                     `json:"hash"`
	}{
		PrivateDataUpdates: updates,
		Hash:               hash,
	})
	require.NoError(t, err)
	return body
}

func computeHash(t *testing.T, v any) string {
	t.Helper()
	b, err := json.Marshal(v)
	require.NoError(t, err)
	return hex.EncodeToString(utils.Hash(b))
}

func samplePrivateData() []*models.PrivateData {
	now := time.Now()
	return []*models.PrivateData{
		{
			ID:           1,
			ClientSideID: "client-id-1",
			UserID:       42,
			Hash:         "somehash",
			Version:      1,
			CreatedAt:    &now,
			UpdatedAt:    &now,
		},
	}
}

func samplePrivateDataUpdates() []models.PrivateDataUpdate {
	return []models.PrivateDataUpdate{
		{
			ClientSideID:      "client-id-1",
			UpdatedRecordHash: "newhash",
			Version:           1,
		},
	}
}

// --- uploadHashing tests ---

func TestUploadHashing_TableTest(t *testing.T) {
	utils.InitHasherPool("test-secret-key")

	validList := samplePrivateData()
	validHash := computeHash(t, validList)
	emptyList := []*models.PrivateData{}
	emptyHash := computeHash(t, emptyList)

	tests := []struct {
		name           string
		body           []byte
		expectedStatus int
	}{
		{
			name:           "valid hash with data",
			body:           makeUploadBody(t, validList, validHash),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "valid hash with empty list",
			body:           makeUploadBody(t, emptyList, emptyHash),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid hash - wrong value",
			body:           makeUploadBody(t, validList, "0000000000000000000000000000000000000000000000000000000000000000"),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid hash - empty string",
			body:           makeUploadBody(t, validList, ""),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid JSON body",
			body:           []byte(`not-json`),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "hash mismatch - tampered data",
			body:           makeUploadBody(t, validList, computeHash(t, emptyList)), // hash is for an empty list while payload is non-empty
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nextCalled := false
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				w.WriteHeader(http.StatusOK)
			})

			middleware := uploadHashing(next)
			req := httptest.NewRequest(http.MethodPost, "/upload", bytes.NewReader(tt.body))
			rr := httptest.NewRecorder()
			middleware.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.expectedStatus == http.StatusOK {
				assert.True(t, nextCalled, "next handler should be called")
			} else {
				assert.False(t, nextCalled, "next handler should NOT be called")
			}
		})
	}
}

func TestUploadHashing_MultipleSequentialRequests(t *testing.T) {
	utils.InitHasherPool("test-secret-key")

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	middleware := uploadHashing(next)

	for i := 0; i < 5; i++ {
		list := samplePrivateData()
		list[0].Version = int64(i)
		hash := computeHash(t, list)
		body := makeUploadBody(t, list, hash)

		req := httptest.NewRequest(http.MethodPost, "/upload", bytes.NewReader(body))
		rr := httptest.NewRecorder()
		middleware.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code, "request %d failed", i)
	}
}

func TestUploadHashing_ConcurrentRequests(t *testing.T) {
	utils.InitHasherPool("test-secret-key")

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	middleware := uploadHashing(next)

	const n = 50
	var wg sync.WaitGroup
	wg.Add(n)

	for i := 0; i < n; i++ {
		go func(i int) {
			defer wg.Done()
			list := samplePrivateData()
			list[0].Version = int64(i)
			hash := computeHash(t, list)
			body := makeUploadBody(t, list, hash)

			req := httptest.NewRequest(http.MethodPost, "/upload", bytes.NewReader(body))
			rr := httptest.NewRecorder()
			middleware.ServeHTTP(rr, req)

			assert.Equal(t, http.StatusOK, rr.Code, "goroutine %d failed", i)
		}(i)
	}

	wg.Wait()
}

func TestUploadHashing_BodyRestoredForNextHandler(t *testing.T) {
	utils.InitHasherPool("test-secret-key")

	list := samplePrivateData()
	hash := computeHash(t, list)
	originalBody := makeUploadBody(t, list, hash)

	var bodyReadByNext []byte
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Middleware must restore the body; read it twice.
		b1, err := io.ReadAll(r.Body)
		require.NoError(t, err, "first read failed")

		// Second read should be empty (NopCloser does not rewind).
		b2, err := io.ReadAll(r.Body)
		require.NoError(t, err, "second read failed")
		assert.Empty(t, b2, "second read should be empty")

		bodyReadByNext = b1
		w.WriteHeader(http.StatusOK)
	})

	middleware := uploadHashing(next)
	req := httptest.NewRequest(http.MethodPost, "/upload", bytes.NewReader(originalBody))
	rr := httptest.NewRecorder()
	middleware.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, originalBody, bodyReadByNext, "next handler should receive full original body")
}

// --- updateHashing tests ---

func TestUpdateHashing_TableTest(t *testing.T) {
	utils.InitHasherPool("test-secret-key")

	validUpdates := samplePrivateDataUpdates()
	validHash := computeHash(t, validUpdates)
	emptyUpdates := []models.PrivateDataUpdate{}
	emptyHash := computeHash(t, emptyUpdates)

	tests := []struct {
		name           string
		body           []byte
		expectedStatus int
	}{
		{
			name:           "valid hash with updates",
			body:           makeUpdateBody(t, validUpdates, validHash),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "valid hash with empty updates",
			body:           makeUpdateBody(t, emptyUpdates, emptyHash),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid hash - wrong value",
			body:           makeUpdateBody(t, validUpdates, "0000000000000000000000000000000000000000000000000000000000000000"),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid hash - empty string",
			body:           makeUpdateBody(t, validUpdates, ""),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid JSON body",
			body:           []byte(`not-json`),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "hash mismatch - tampered data",
			body:           makeUpdateBody(t, validUpdates, computeHash(t, emptyUpdates)),
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nextCalled := false
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				w.WriteHeader(http.StatusOK)
			})

			middleware := updateHashing(next)
			req := httptest.NewRequest(http.MethodPost, "/update", bytes.NewReader(tt.body))
			rr := httptest.NewRecorder()
			middleware.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.expectedStatus == http.StatusOK {
				assert.True(t, nextCalled, "next handler should be called")
			} else {
				assert.False(t, nextCalled, "next handler should NOT be called")
			}
		})
	}
}

func TestUpdateHashing_MultipleSequentialRequests(t *testing.T) {
	utils.InitHasherPool("test-secret-key")

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	middleware := updateHashing(next)

	for i := 0; i < 5; i++ {
		updates := samplePrivateDataUpdates()
		updates[0].Version = int64(i)
		hash := computeHash(t, updates)
		body := makeUpdateBody(t, updates, hash)

		req := httptest.NewRequest(http.MethodPost, "/update", bytes.NewReader(body))
		rr := httptest.NewRecorder()
		middleware.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code, "request %d failed", i)
	}
}

func TestUpdateHashing_ConcurrentRequests(t *testing.T) {
	utils.InitHasherPool("test-secret-key")

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	middleware := updateHashing(next)

	const n = 50
	var wg sync.WaitGroup
	wg.Add(n)

	for i := 0; i < n; i++ {
		go func(i int) {
			defer wg.Done()
			updates := samplePrivateDataUpdates()
			updates[0].Version = int64(i)
			hash := computeHash(t, updates)
			body := makeUpdateBody(t, updates, hash)

			req := httptest.NewRequest(http.MethodPost, "/update", bytes.NewReader(body))
			rr := httptest.NewRecorder()
			middleware.ServeHTTP(rr, req)

			assert.Equal(t, http.StatusOK, rr.Code, "goroutine %d failed", i)
		}(i)
	}

	wg.Wait()
}

func TestUpdateHashing_BodyRestoredForNextHandler(t *testing.T) {
	utils.InitHasherPool("test-secret-key")

	updates := samplePrivateDataUpdates()
	hash := computeHash(t, updates)
	originalBody := makeUpdateBody(t, updates, hash)

	var bodyReadByNext []byte
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b1, err := io.ReadAll(r.Body)
		require.NoError(t, err, "first read failed")

		// Second read should be empty because the body is not rewound.
		b2, err := io.ReadAll(r.Body)
		require.NoError(t, err, "second read failed")
		assert.Empty(t, b2, "second read should be empty")

		bodyReadByNext = b1
		w.WriteHeader(http.StatusOK)
	})

	middleware := updateHashing(next)
	req := httptest.NewRequest(http.MethodPost, "/update", bytes.NewReader(originalBody))
	rr := httptest.NewRecorder()
	middleware.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, originalBody, bodyReadByNext, "next handler should receive full original body")
}
