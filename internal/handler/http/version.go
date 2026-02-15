package http

import (
	"net/http"
)

func (h *Handler) getServerVersion(w http.ResponseWriter, r *http.Request) {
	serverVersion := h.services.AppInfoService.GetAppVersion(r.Context())

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(serverVersion))
}
