package http

import (
	"errors"
	"net/http"

	"github.com/MKhiriev/go-pass-keeper/internal/service"
	"github.com/MKhiriev/go-pass-keeper/internal/store"
)

type errorResponse struct {
	message string
	status  int
}

var errorStatusMap = map[error]errorResponse{
	// service errors
	service.ErrInvalidDataProvided:                            {message: "invalid data provided", status: http.StatusBadRequest},
	service.ErrWrongPassword:                                  {message: "wrong password", status: http.StatusUnauthorized},
	service.ErrTokenCreationFailed:                            {message: "internal server error", status: http.StatusInternalServerError},
	service.ErrTokenIsExpired:                                 {message: "token is expired", status: http.StatusUnauthorized},
	service.ErrTokenIsExpiredOrInvalid:                        {message: "token is expired or invalid", status: http.StatusUnauthorized},
	service.ErrValidationNoPrivateDataProvided:                {message: "no private data provided", status: http.StatusBadRequest},
	service.ErrValidationNoDownloadRequestsProvided:           {message: "no download requests provided", status: http.StatusBadRequest},
	service.ErrValidationNoUpdateRequestsProvided:             {message: "no update requests provided", status: http.StatusBadRequest},
	service.ErrValidationNoDeleteRequestsProvided:             {message: "no delete requests provided", status: http.StatusBadRequest},
	service.ErrValidationNoUserID:                             {message: "no user ID provided", status: http.StatusBadRequest},
	service.ErrValidationNoClientIDsProvidedForSyncRequests:   {message: "no client IDs provided for sync", status: http.StatusBadRequest},
	service.ErrValidationEmptyClientIDProvidedForSyncRequests: {message: "empty client ID provided for sync", status: http.StatusBadRequest},
	service.ErrUnauthorizedAccessToDifferentUserData:          {message: "access denied", status: http.StatusForbidden},
	service.ErrVersionIsNotSpecified:                          {message: "version is not specified", status: http.StatusBadRequest},
	service.ErrRegisterOnServer:                               {message: "registration failed", status: http.StatusBadGateway},
	service.ErrLoginOnServer:                                  {message: "login failed", status: http.StatusBadGateway},

	// store errors
	store.ErrLoginAlreadyExists:  {message: "login already exists", status: http.StatusConflict},
	store.ErrNoUserWasFound:      {message: "user not found", status: http.StatusNotFound},
	store.ErrPrivateDataNotSaved: {message: "internal server error", status: http.StatusInternalServerError},
	store.ErrPrivateDataNotFound: {message: "data not found", status: http.StatusNotFound},
	store.ErrVersionConflict:     {message: "version conflict, please sync", status: http.StatusConflict},

	// store internal errors
	store.ErrBuildingSQLQuery:     {message: "internal server error", status: http.StatusInternalServerError},
	store.ErrExecutingQuery:       {message: "internal server error", status: http.StatusInternalServerError},
	store.ErrBeginningTransaction: {message: "internal server error", status: http.StatusInternalServerError},
	store.ErrCommitingTransaction: {message: "internal server error", status: http.StatusInternalServerError},
	store.ErrPreparingStatement:   {message: "internal server error", status: http.StatusInternalServerError},
	store.ErrExecutingStatement:   {message: "internal server error", status: http.StatusInternalServerError},
	store.ErrScanningRow:          {message: "internal server error", status: http.StatusInternalServerError},
	store.ErrScanningRows:         {message: "internal server error", status: http.StatusInternalServerError},
}

func responseFromError(err error) errorResponse {
	for target, resp := range errorStatusMap {
		if errors.Is(err, target) {
			return resp
		}
	}
	return errorResponse{message: "internal server error", status: http.StatusInternalServerError}
}
