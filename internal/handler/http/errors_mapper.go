package http

import (
	"errors"
	"net/http"

	"github.com/MKhiriev/go-pass-keeper/internal/service"
	"github.com/MKhiriev/go-pass-keeper/internal/store"
)

var errorStatusMap = map[error]int{
	service.ErrInvalidDataProvided:                            http.StatusBadRequest,
	service.ErrWrongPassword:                                  http.StatusUnauthorized,
	service.ErrTokenIsExpired:                                 http.StatusUnauthorized,
	service.ErrTokenIsExpiredOrInvalid:                        http.StatusUnauthorized,
	service.ErrValidationNoPrivateDataProvided:                http.StatusBadRequest,
	service.ErrValidationNoDownloadRequestsProvided:           http.StatusBadRequest,
	service.ErrValidationNoUpdateRequestsProvided:             http.StatusBadRequest,
	service.ErrValidationNoDeleteRequestsProvided:             http.StatusBadRequest,
	service.ErrValidationNoUserID:                             http.StatusBadRequest,
	service.ErrValidationNoClientIDsProvidedForSyncRequests:   http.StatusBadRequest,
	service.ErrValidationEmptyClientIDProvidedForSyncRequests: http.StatusBadRequest,
	service.ErrUnauthorizedAccessToDifferentUserData:          http.StatusForbidden,
	service.ErrVersionIsNotSpecified:                          http.StatusBadRequest,
	service.ErrRegisterOnServer:                               http.StatusBadGateway,
	service.ErrLoginOnServer:                                  http.StatusBadGateway,

	store.ErrLoginAlreadyExists:  http.StatusConflict,
	store.ErrNoUserWasFound:      http.StatusNotFound,
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
