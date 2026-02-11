package http

import (
	"encoding/json"
	"net/http"

	"github.com/MKhiriev/go-pass-keeper/internal/logger"
	"github.com/MKhiriev/go-pass-keeper/models"
)

func (h *Handler) upload(w http.ResponseWriter, r *http.Request) {
	log := logger.FromRequest(r)

	var dataFromBody models.PrivateData
	if err := json.NewDecoder(r.Body).Decode(&dataFromBody); err != nil {
		log.Err(err).Str("func", "*Handler.upload").Msg("Invalid JSON was passed")
		http.Error(w, "Invalid JSON was passed", http.StatusBadRequest)
		return
	}

	err := h.services.PrivateDataService.UploadPrivateData(r.Context(), dataFromBody)
	if err != nil {
		// todo add error classification later
		log.Err(err).Str("func", "*Handler.upload").Msg("error uploading private data")
		http.Error(w, "error uploading private data", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *Handler) download(w http.ResponseWriter, r *http.Request) {
	// 1. Get {type} and {id} url params
	// 2. Validate params
	// * get userID from context
	// 3. give data needed to the service to search for needed data
	// 4. check if errors are returned
	// 5. return response
}

func (h *Handler) downloadMultiple(w http.ResponseWriter, r *http.Request) {
	// 1. Get needed data in JSON array
	// 2. Validate each data point
	// * get userID from context
	// 3. give data needed to the service to search for needed data
	// 4. check if errors are returned
	// 5. return response
}

func (h *Handler) downloadAllUserData(w http.ResponseWriter, r *http.Request) {
	// * get userID from context
	// 1. call service method for getting all data
	// 2. check if errors are returned
	// 3. return response
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	// 1. Get needed data in JSON array
	// 2. Validate each data point
}

func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	// * get userID from context
	// 1. call service method for getting all data
	// 2. check if errors are returned
	// 3. return response
}
