package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/MKhiriev/go-pass-keeper/internal/logger"
	"github.com/MKhiriev/go-pass-keeper/internal/utils"
	"github.com/MKhiriev/go-pass-keeper/models"
	"github.com/go-chi/chi/v5"
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
	log := logger.FromRequest(r)

	idParam := chi.URLParam(r, "id")
	typeParam := chi.URLParam(r, "type")

	id, err1 := strconv.ParseInt(idParam, 10, 64)
	dType, err2 := strconv.ParseInt(typeParam, 10, 64)
	if err1 != nil || err2 != nil {
		log.Error().Str("func", "*Handler.download").
			AnErr("id error", err1).
			AnErr("type error", err2).
			Msg("invalid id/type passed")
		http.Error(w, "invalid id/type passed", http.StatusBadRequest)
		return
	}

	dataToDownload := models.PrivateData{
		ID:   id,
		Type: models.DataType(dType),
	}

	requestedData, err := h.services.PrivateDataService.DownloadPrivateData(r.Context(), dataToDownload)
	if err != nil {
		// todo add error classification later
		log.Err(err).Str("func", "*Handler.download").Msg("error downloading private data")
		http.Error(w, "error downloading private data", http.StatusInternalServerError)
		return
	}

	utils.WriteJSON(w, requestedData, http.StatusOK)
}

func (h *Handler) downloadMultiple(w http.ResponseWriter, r *http.Request) {
	log := logger.FromRequest(r)

	var dataArrayFromBody []models.PrivateData
	if err := json.NewDecoder(r.Body).Decode(&dataArrayFromBody); err != nil {
		log.Err(err).Str("func", "*Handler.downloadMultiple").Msg("Invalid JSON was passed")
		http.Error(w, "Invalid JSON was passed", http.StatusBadRequest)
		return
	}

	requestedData, err := h.services.PrivateDataService.DownloadMultiplePrivateData(r.Context(), dataArrayFromBody)
	if err != nil {
		// todo add error classification later
		log.Err(err).Str("func", "*Handler.downloadMultiple").Msg("error downloading private data")
		http.Error(w, "error downloading private data", http.StatusInternalServerError)
		return
	}

	utils.WriteJSON(w, requestedData, http.StatusOK)
}

func (h *Handler) downloadAllUserData(w http.ResponseWriter, r *http.Request) {
	log := logger.FromRequest(r)

	requestedData, err := h.services.PrivateDataService.DownloadAllPrivateData(r.Context())
	if err != nil {
		// todo add error classification later
		log.Err(err).Str("func", "*Handler.downloadAllUserData").Msg("error downloading all private user data")
		http.Error(w, "error downloading private data", http.StatusInternalServerError)
		return
	}

	utils.WriteJSON(w, requestedData, http.StatusOK)
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	log := logger.FromRequest(r)

	var dataArrayFromBody []models.PrivateData
	if err := json.NewDecoder(r.Body).Decode(&dataArrayFromBody); err != nil {
		log.Err(err).Str("func", "*Handler.update").Msg("Invalid JSON was passed")
		http.Error(w, "Invalid JSON was passed", http.StatusBadRequest)
		return
	}

	err := h.services.PrivateDataService.UpdatePrivateData(r.Context(), dataArrayFromBody...)
	if err != nil {
		// todo add error classification later
		log.Err(err).Str("func", "*Handler.update").Msg("error updating private data")
		http.Error(w, "error updating private data", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	log := logger.FromRequest(r)

	var dataArrayFromBody []models.PrivateData
	if err := json.NewDecoder(r.Body).Decode(&dataArrayFromBody); err != nil {
		log.Err(err).Str("func", "*Handler.update").Msg("Invalid JSON was passed")
		http.Error(w, "Invalid JSON was passed", http.StatusBadRequest)
		return
	}

	err := h.services.PrivateDataService.DeletePrivateData(r.Context(), dataArrayFromBody...)
	if err != nil {
		// todo add error classification later
		log.Err(err).Str("func", "*Handler.update").Msg("error deleting private data")
		http.Error(w, "error deleting private data", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
