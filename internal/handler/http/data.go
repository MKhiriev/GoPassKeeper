package http

import (
	"encoding/json"
	"net/http"

	"github.com/MKhiriev/go-pass-keeper/internal/logger"
	"github.com/MKhiriev/go-pass-keeper/internal/utils"
	"github.com/MKhiriev/go-pass-keeper/models"
)

func (h *Handler) upload(w http.ResponseWriter, r *http.Request) {
	log := logger.FromRequest(r)

	var uploadRequest models.UploadRequest
	if err := json.NewDecoder(r.Body).Decode(&uploadRequest); err != nil {
		log.Err(err).Str("func", "*Handler.upload").Msg("Invalid JSON was passed")
		http.Error(w, "Invalid JSON was passed", http.StatusBadRequest)
		return
	}

	err := h.services.PrivateDataService.UploadPrivateData(r.Context(), uploadRequest)
	if err != nil {
		log.Err(err).Str("func", "*Handler.upload").Msg("error uploading private data")
		resp := responseFromError(err)
		http.Error(w, resp.message, resp.status)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *Handler) downloadMultiple(w http.ResponseWriter, r *http.Request) {
	log := logger.FromRequest(r)

	var dataArrayFromBody models.DownloadRequest
	if err := json.NewDecoder(r.Body).Decode(&dataArrayFromBody); err != nil {
		log.Err(err).Str("func", "*Handler.downloadMultiple").Msg("Invalid JSON was passed")
		resp := responseFromError(err)
		http.Error(w, resp.message, resp.status)
		return
	}

	requestedData, err := h.services.PrivateDataService.DownloadPrivateData(r.Context(), dataArrayFromBody)
	if err != nil {
		log.Err(err).Str("func", "*Handler.downloadMultiple").Msg("error downloading private data")
		resp := responseFromError(err)
		http.Error(w, resp.message, resp.status)
		return
	}

	utils.WriteJSON(w, requestedData, http.StatusOK)
}

func (h *Handler) downloadAllUserData(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromRequest(r)

	userID, found := utils.GetUserIDFromContext(ctx)
	if !found {
		log.Error().Str("func", "*Handler.downloadAllUserData").Msg("no user ID was given")
		http.Error(w, "no user ID was given", http.StatusBadRequest)
		return
	}

	requestedData, err := h.services.PrivateDataService.DownloadAllPrivateData(ctx, userID)
	if err != nil {
		log.Err(err).Str("func", "*Handler.downloadAllUserData").Msg("error downloading all private user data")
		resp := responseFromError(err)
		http.Error(w, resp.message, resp.status)
		return
	}

	utils.WriteJSON(w, requestedData, http.StatusOK)
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	log := logger.FromRequest(r)

	var dataArrayFromBody models.UpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&dataArrayFromBody); err != nil {
		log.Err(err).Str("func", "*Handler.update").Msg("Invalid JSON was passed")
		http.Error(w, "Invalid JSON was passed", http.StatusBadRequest)
		return
	}

	err := h.services.PrivateDataService.UpdatePrivateData(r.Context(), dataArrayFromBody)
	if err != nil {
		log.Err(err).Str("func", "*Handler.update").Msg("error updating private data")
		resp := responseFromError(err)
		http.Error(w, resp.message, resp.status)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	log := logger.FromRequest(r)

	var dataArrayFromBody models.DeleteRequest
	if err := json.NewDecoder(r.Body).Decode(&dataArrayFromBody); err != nil {
		log.Err(err).Str("func", "*Handler.update").Msg("Invalid JSON was passed")
		http.Error(w, "Invalid JSON was passed", http.StatusBadRequest)
		return
	}

	err := h.services.PrivateDataService.DeletePrivateData(r.Context(), dataArrayFromBody)
	if err != nil {
		log.Err(err).Str("func", "*Handler.update").Msg("error deleting private data")
		resp := responseFromError(err)
		http.Error(w, resp.message, resp.status)
		return
	}

	w.WriteHeader(http.StatusOK)
}
