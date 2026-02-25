package http

import (
	"errors"
	"net/http"

	"github.com/MKhiriev/go-pass-keeper/internal/service"
	"github.com/MKhiriev/go-pass-keeper/internal/store"
)

var errorStatusMap = map[error]int{
	service.ErrInvalidDataProvided:                          http.StatusBadRequest,
	service.ErrValidationNoUpdateRequestsProvided:           http.StatusBadRequest,
	service.ErrValidationNoClientIDsProvidedForSyncRequests: http.StatusBadRequest,
	service.ErrValidationNoUserID:                           http.StatusUnauthorized,
	service.ErrUnauthorizedAccessToDifferentUserData:        http.StatusForbidden,

	store.ErrLoginAlreadyExists:  http.StatusBadRequest,
	store.ErrNoUserWasFound:      http.StatusBadRequest,
	store.ErrPrivateDataNotSaved: http.StatusInternalServerError,
	store.ErrPrivateDataNotFound: http.StatusNotFound,
	store.ErrVersionConflict:     http.StatusConflict,

	store.ErrBuildingSQLQuery:     http.StatusInternalServerError,
	store.ErrExecutingQuery:       http.StatusInternalServerError,
	store.ErrBeginningTransaction: http.StatusInternalServerError,
	store.ErrCommitingTransaction: http.StatusInternalServerError,
	store.ErrPreparingStatement:   http.StatusInternalServerError,
	store.ErrExecutingStatement:   http.StatusInternalServerError,
	store.ErrScanningRow:          http.StatusInternalServerError,
	store.ErrScanningRows:         http.StatusInternalServerError,
}

func statusFromError(err error) int {
	for target, status := range errorStatusMap {
		if errors.Is(err, target) {
			return status
		}
	}
	return http.StatusInternalServerError
}
