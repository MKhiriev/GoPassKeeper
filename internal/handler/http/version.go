// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Rasul Khiriev

package http

import (
	"net/http"
)

func (h *Handler) getServerVersion(w http.ResponseWriter, r *http.Request) {
	serverVersion := h.services.AppInfoService.GetAppVersion(r.Context())

	w.Write([]byte(serverVersion))
}
