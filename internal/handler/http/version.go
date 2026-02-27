package http

import (
	"net/http"
)

func (h *Handler) getServerVersion(w http.ResponseWriter, r *http.Request) {
	serverVersion := h.services.AppInfoService.GetAppVersion(r.Context())

	w.Write([]byte(serverVersion))
}
