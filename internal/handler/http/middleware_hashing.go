package http

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"

	"github.com/MKhiriev/go-pass-keeper/internal/utils"
	"github.com/MKhiriev/go-pass-keeper/models"
)

func (h *Handler) uploadHashing(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			PrivateDataList []*models.PrivateData `json:"private_data_list"`
			Hash            string                `json:"hash"`
		}

		h.logger.Debug().Str("func", "*Handler.uploadHashing").Msg("checking hash begins")

		// read bytes from body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			h.logger.Err(err).Str("func", "*Handler.uploadHashing").Msg("failed to read request body")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		// restore request body
		r.Body = io.NopCloser(bytes.NewReader(body))

		// Decode JSON from []byte
		if err := json.NewDecoder(bytes.NewReader(body)).Decode(&req); err != nil {
			h.logger.Err(err).Str("func", "*Handler.uploadHashing").Msg("failed to decode JSON")
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		// Serialize PrivateDataList back to JSON for hashing
		payloadBytes, err := json.Marshal(req.PrivateDataList)
		if err != nil {
			h.logger.Err(err).Str("func", "*Handler.uploadHashing").Msg("failed to marshal PrivateDataList")
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}

		// Calculate hash from JSON Payload
		hashedBody := hex.EncodeToString(utils.Hash(payloadBytes))
		if hashedBody != req.Hash {
			h.logger.Error().Str("func", "*Handler.uploadHashing").
				Str("hash from request", req.Hash).
				Str("hashed body", hashedBody).
				Msg("hashes are not equal")
			http.Error(w, "Integrity check failed", http.StatusBadRequest)
			return
		}

		h.logger.Debug().Str("func", "*Handler.uploadHashing").
			Str("hash from request", req.Hash).
			Str("hashed body", hashedBody).
			Msg("hashes are equal")

		next.ServeHTTP(w, r)
	})
}

func (h *Handler) updateHashing(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			PrivateDataUpdates []models.PrivateDataUpdate `json:"private_data_updates"`
			Hash               string                     `json:"hash"`
		}

		h.logger.Debug().Str("func", "*Handler.updateHashing").Msg("checking hash begins")

		// read bytes from body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			h.logger.Err(err).Str("func", "*Handler.updateHashing").Msg("failed to read request body")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		// restore request body
		r.Body = io.NopCloser(bytes.NewReader(body))

		// Decode JSON from []byte
		if err := json.NewDecoder(bytes.NewReader(body)).Decode(&req); err != nil {
			h.logger.Err(err).Str("func", "*Handler.updateHashing").Msg("failed to decode JSON")
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		// Serialize Payload back to JSON for hashing
		payloadBytes, err := json.Marshal(req.PrivateDataUpdates)
		if err != nil {
			h.logger.Err(err).Str("func", "*Handler.updateHashing").Msg("failed to marshal payload")
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}

		// Calculate hash from JSON Payload
		hashedBody := hex.EncodeToString(utils.Hash(payloadBytes))
		if hashedBody != req.Hash {
			h.logger.Error().Str("func", "*Handler.updateHashing").
				Str("hash from request", req.Hash).
				Str("hashed body", hashedBody).
				Msg("hashes are not equal")
			http.Error(w, "Integrity check failed", http.StatusBadRequest)
			return
		}

		h.logger.Debug().Str("func", "*Handler.updateHashing").
			Str("hash from request", req.Hash).
			Str("hashed body", hashedBody).
			Msg("hashes are equal")

		next.ServeHTTP(w, r)
	})
}
