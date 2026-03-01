// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Rasul Khiriev

package http

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"

	"github.com/MKhiriev/go-pass-keeper/internal/logger"
	"github.com/MKhiriev/go-pass-keeper/internal/utils"
	"github.com/MKhiriev/go-pass-keeper/models"
)

// uploadHashing is an HTTP middleware that verifies the transport integrity
// of a vault-item upload request before forwarding it to the next handler.
//
// The middleware expects the request body to be a JSON object with the
// following structure:
//
//	{
//	    "private_data_list": [ ... ],  // list of vault items to upload
//	    "hash": "<hex-encoded HMAC>"   // integrity checksum
//	}
//
// Integrity verification proceeds as follows:
//  1. The raw request body is read and immediately restored so that
//     downstream handlers can read it again without re-seeking.
//  2. The body is decoded into the expected JSON shape.
//  3. The "private_data_list" field is re-serialised to JSON and hashed
//     via [utils.Hash]. The result is hex-encoded.
//  4. The computed hash is compared against the "hash" field supplied by
//     the client. If they differ, the request is rejected with
//     HTTP 400 Bad Request and an "Integrity check failed" message.
//
// On success the original request (with the restored body) is passed to next.
// All intermediate errors are logged via the context-scoped logger obtained
// from [logger.FromRequest].
func uploadHashing(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log := logger.FromRequest(r)

		// uploadRequest is the expected shape of an upload request body.
		// It is defined inline because it is not shared with any other handler.
		var req struct {
			PrivateDataList []*models.PrivateData `json:"private_data_list"`
			Hash            string                `json:"hash"`
		}

		log.Debug().Str("func", "*Handler.uploadHashing").Msg("checking hash begins")

		// Read the entire body into memory so it can be decoded and then
		// restored for downstream handlers.
		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Err(err).Str("func", "*Handler.uploadHashing").Msg("failed to read request body")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		// Restore the body so that the next handler can read it from the start.
		r.Body = io.NopCloser(bytes.NewReader(body))

		// Decode the JSON body into the expected request structure.
		if err := json.NewDecoder(bytes.NewReader(body)).Decode(&req); err != nil {
			log.Err(err).Str("func", "*Handler.uploadHashing").Msg("failed to decode JSON")
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		// Re-serialise only the payload field to obtain a canonical byte
		// representation that is independent of the surrounding JSON envelope.
		payloadBytes, err := json.Marshal(req.PrivateDataList)
		if err != nil {
			log.Err(err).Str("func", "*Handler.uploadHashing").Msg("failed to marshal PrivateDataList")
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}

		// Compute the expected hash and compare it against the client-supplied value.
		hashedBody := hex.EncodeToString(utils.Hash(payloadBytes))
		if hashedBody != req.Hash {
			log.Error().Str("func", "*Handler.uploadHashing").
				Str("hash from request", req.Hash).
				Str("hashed body", hashedBody).
				Msg("hashes are not equal")
			http.Error(w, "Integrity check failed", http.StatusBadRequest)
			return
		}

		log.Debug().Str("func", "*Handler.uploadHashing").
			Str("hash from request", req.Hash).
			Str("hashed body", hashedBody).
			Msg("hashes are equal")

		next.ServeHTTP(w, r)
	})
}

// updateHashing is an HTTP middleware that verifies the transport integrity
// of a vault-item update request before forwarding it to the next handler.
//
// The middleware expects the request body to be a JSON object with the
// following structure:
//
//	{
//	    "private_data_updates": [ ... ],  // list of partial update descriptors
//	    "hash": "<hex-encoded HMAC>"      // integrity checksum
//	}
//
// Integrity verification proceeds as follows:
//  1. The raw request body is read and immediately restored so that
//     downstream handlers can read it again without re-seeking.
//  2. The body is decoded into the expected JSON shape.
//  3. The "private_data_updates" field is re-serialised to JSON and hashed
//     via [utils.Hash]. The result is hex-encoded.
//  4. The computed hash is compared against the "hash" field supplied by
//     the client. If they differ, the request is rejected with
//     HTTP 400 Bad Request and an "Integrity check failed" message.
//
// On success the original request (with the restored body) is passed to next.
// All intermediate errors are logged via the context-scoped logger obtained
// from [logger.FromRequest].
func updateHashing(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log := logger.FromRequest(r)

		// updateRequest is the expected shape of an update request body.
		// It is defined inline because it is not shared with any other handler.
		var req struct {
			PrivateDataUpdates []models.PrivateDataUpdate `json:"private_data_updates"`
			Hash               string                     `json:"hash"`
		}

		log.Debug().Str("func", "*Handler.updateHashing").Msg("checking hash begins")

		// Read the entire body into memory so it can be decoded and then
		// restored for downstream handlers.
		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Err(err).Str("func", "*Handler.updateHashing").Msg("failed to read request body")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		// Restore the body so that the next handler can read it from the start.
		r.Body = io.NopCloser(bytes.NewReader(body))

		// Decode the JSON body into the expected request structure.
		if err := json.NewDecoder(bytes.NewReader(body)).Decode(&req); err != nil {
			log.Err(err).Str("func", "*Handler.updateHashing").Msg("failed to decode JSON")
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		// Re-serialise only the payload field to obtain a canonical byte
		// representation that is independent of the surrounding JSON envelope.
		payloadBytes, err := json.Marshal(req.PrivateDataUpdates)
		if err != nil {
			log.Err(err).Str("func", "*Handler.updateHashing").Msg("failed to marshal payload")
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}

		// Compute the expected hash and compare it against the client-supplied value.
		hashedBody := hex.EncodeToString(utils.Hash(payloadBytes))
		if hashedBody != req.Hash {
			log.Error().Str("func", "*Handler.updateHashing").
				Str("hash from request", req.Hash).
				Str("hashed body", hashedBody).
				Msg("hashes are not equal")
			http.Error(w, "Integrity check failed", http.StatusBadRequest)
			return
		}

		log.Debug().Str("func", "*Handler.updateHashing").
			Str("hash from request", req.Hash).
			Str("hashed body", hashedBody).
			Msg("hashes are equal")

		next.ServeHTTP(w, r)
	})
}
