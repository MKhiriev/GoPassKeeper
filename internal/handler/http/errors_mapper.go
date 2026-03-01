// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Rasul Khiriev

package http

import (
	"errors"
	"net/http"

	"github.com/MKhiriev/go-pass-keeper/internal/app"
	"github.com/MKhiriev/go-pass-keeper/internal/service"
	"github.com/MKhiriev/go-pass-keeper/internal/store"
)

type errorResponse struct {
	message string
	status  int
}

var errorStatusMap = map[error]errorResponse{
	service.ErrInvalidDataProvided:                            {message: app.MsgInvalidDataProvided, status: http.StatusBadRequest},
	service.ErrWrongPassword:                                  {message: app.MsgInvalidLoginPassword, status: http.StatusUnauthorized},
	service.ErrTokenCreationFailed:                            {message: app.MsgInternalServerError, status: http.StatusInternalServerError},
	service.ErrTokenIsExpired:                                 {message: app.MsgTokenIsExpired, status: http.StatusUnauthorized},
	service.ErrTokenIsExpiredOrInvalid:                        {message: app.MsgTokenIsExpiredOrInvalid, status: http.StatusUnauthorized},
	service.ErrValidationNoPrivateDataProvided:                {message: app.MsgNoPrivateDataProvided, status: http.StatusBadRequest},
	service.ErrValidationNoDownloadRequestsProvided:           {message: app.MsgNoDownloadRequestsProvided, status: http.StatusBadRequest},
	service.ErrValidationNoUpdateRequestsProvided:             {message: app.MsgNoUpdateRequestsProvided, status: http.StatusBadRequest},
	service.ErrValidationNoDeleteRequestsProvided:             {message: app.MsgNoDeleteRequestsProvided, status: http.StatusBadRequest},
	service.ErrValidationNoUserID:                             {message: app.MsgNoUserIDProvided, status: http.StatusBadRequest},
	service.ErrValidationNoClientIDsProvidedForSyncRequests:   {message: app.MsgNoClientIDsForSync, status: http.StatusBadRequest},
	service.ErrValidationEmptyClientIDProvidedForSyncRequests: {message: app.MsgEmptyClientIDForSync, status: http.StatusBadRequest},
	service.ErrUnauthorizedAccessToDifferentUserData:          {message: app.MsgAccessDenied, status: http.StatusForbidden},
	service.ErrVersionIsNotSpecified:                          {message: app.MsgVersionIsNotSpecified, status: http.StatusBadRequest},
	service.ErrRegisterOnServer:                               {message: app.MsgRegistrationFailed, status: http.StatusBadGateway},
	service.ErrLoginOnServer:                                  {message: app.MsgLoginFailed, status: http.StatusBadGateway},

	store.ErrLoginAlreadyExists:  {message: app.MsgLoginAlreadyExists, status: http.StatusConflict},
	store.ErrNoUserWasFound:      {message: app.MsgInvalidLoginPassword, status: http.StatusUnauthorized},
	store.ErrPrivateDataNotSaved: {message: app.MsgInternalServerError, status: http.StatusInternalServerError},
	store.ErrPrivateDataNotFound: {message: app.MsgDataNotFound, status: http.StatusNotFound},
	store.ErrVersionConflict:     {message: app.MsgVersionConflict, status: http.StatusConflict},

	store.ErrBuildingSQLQuery:     {message: app.MsgInternalServerError, status: http.StatusInternalServerError},
	store.ErrExecutingQuery:       {message: app.MsgInternalServerError, status: http.StatusInternalServerError},
	store.ErrBeginningTransaction: {message: app.MsgInternalServerError, status: http.StatusInternalServerError},
	store.ErrCommitingTransaction: {message: app.MsgInternalServerError, status: http.StatusInternalServerError},
	store.ErrPreparingStatement:   {message: app.MsgInternalServerError, status: http.StatusInternalServerError},
	store.ErrExecutingStatement:   {message: app.MsgInternalServerError, status: http.StatusInternalServerError},
	store.ErrScanningRow:          {message: app.MsgInternalServerError, status: http.StatusInternalServerError},
	store.ErrScanningRows:         {message: app.MsgInternalServerError, status: http.StatusInternalServerError},
}

func responseFromError(err error) errorResponse {
	for target, resp := range errorStatusMap {
		if errors.Is(err, target) {
			return resp
		}
	}
	return errorResponse{message: app.MsgInternalServerError, status: http.StatusInternalServerError}
}
