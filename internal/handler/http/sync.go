package http

import (
	"encoding/json"
	"net/http"

	"github.com/MKhiriev/go-pass-keeper/internal/logger"
	"github.com/MKhiriev/go-pass-keeper/internal/utils"
	"github.com/MKhiriev/go-pass-keeper/models"
)

func (h *Handler) getClientServerDiff(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromRequest(r)

	userID, found := utils.GetUserIDFromContext(ctx)
	if !found {
		log.Error().Str("func", "*Handler.downloadAllUserData").Msg("no user ID was given")
		http.Error(w, "no user ID was given", http.StatusBadRequest)
		return
	}

	privateDataStates, err := h.services.PrivateDataService.DownloadUserPrivateDataStates(ctx, userID)
	if err != nil {
		log.Error().Str("func", "*Handler.getClientServerDiff").Msg("error getting user private data states")
		http.Error(w, "error getting user private data states", statusFromError(err))
		return
	}

	response := models.SyncResponse{
		PrivateDataStates: privateDataStates,
		Length:            len(privateDataStates),
	}

	utils.WriteJSON(w, response, http.StatusOK)
}

func (h *Handler) syncSpecificUserData(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromRequest(r)

	var syncRequest models.SyncRequest
	if err := json.NewDecoder(r.Body).Decode(&syncRequest); err != nil {
		log.Err(err).Str("func", "*Handler.syncSpecificUserData").Msg("Invalid JSON was passed")
		http.Error(w, "Invalid JSON was passed", http.StatusBadRequest)
		return
	}

	privateDataStates, err := h.services.PrivateDataService.DownloadSpecificUserPrivateDataStates(ctx, syncRequest)
	if err != nil {
		log.Error().Str("func", "*Handler.getClientServerDiff").Msg("error getting specific user private data states")
		http.Error(w, "error getting specific user private data states", statusFromError(err))
		return
	}

	response := models.SyncResponse{
		PrivateDataStates: privateDataStates,
		Length:            len(privateDataStates),
	}

	utils.WriteJSON(w, response, http.StatusOK)
}
