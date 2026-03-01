// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Rasul Khiriev

package service

import (
	"errors"
	"strings"

	"github.com/MKhiriev/go-pass-keeper/internal/adapter"
	"github.com/MKhiriev/go-pass-keeper/internal/app"
	"github.com/MKhiriev/go-pass-keeper/internal/store"
)

// mapAdapterError translates the adapter's transport error into a service business error
func mapAdapterError(err error) error {
	if err == nil {
		return nil
	}

	msg := extractBody(err)

	switch {
	case errors.Is(err, adapter.ErrBadRequest):
		switch msg {
		case app.MsgInvalidDataProvided:
			return ErrInvalidDataProvided
		case app.MsgNoPrivateDataProvided:
			return ErrValidationNoPrivateDataProvided
		case app.MsgNoDownloadRequestsProvided:
			return ErrValidationNoDownloadRequestsProvided
		case app.MsgNoUpdateRequestsProvided:
			return ErrValidationNoUpdateRequestsProvided
		case app.MsgNoDeleteRequestsProvided:
			return ErrValidationNoDeleteRequestsProvided
		case app.MsgNoUserIDProvided:
			return ErrValidationNoUserID
		case app.MsgNoClientIDsForSync:
			return ErrValidationNoClientIDsProvidedForSyncRequests
		case app.MsgEmptyClientIDForSync:
			return ErrValidationEmptyClientIDProvidedForSyncRequests
		case app.MsgVersionIsNotSpecified:
			return ErrVersionIsNotSpecified
		}

	case errors.Is(err, adapter.ErrUnauthorized):
		switch msg {
		case app.MsgInvalidLoginPassword:
			return ErrWrongPassword
		case app.MsgTokenIsExpired:
			return ErrTokenIsExpired
		case app.MsgTokenIsExpiredOrInvalid:
			return ErrTokenIsExpiredOrInvalid
		}

	case errors.Is(err, adapter.ErrForbidden):
		return ErrUnauthorizedAccessToDifferentUserData

	case errors.Is(err, adapter.ErrNotFound):
		return store.ErrPrivateDataNotFound

	case errors.Is(err, adapter.ErrConflict):
		switch msg {
		case app.MsgLoginAlreadyExists:
			return store.ErrLoginAlreadyExists
		case app.MsgVersionConflict:
			return store.ErrVersionConflict
		}

	case errors.Is(err, adapter.ErrBadGateway):
		switch msg {
		case app.MsgRegistrationFailed:
			return ErrRegisterOnServer
		case app.MsgLoginFailed:
			return ErrLoginOnServer
		}

	case errors.Is(err, adapter.ErrInternalServerError):
		return ErrTokenCreationFailed
	}

	return err
}

// extractBody extracts the body from a message of the form "bad request: <body>"
func extractBody(err error) string {
	msg := err.Error()
	if idx := strings.Index(msg, ": "); idx != -1 {
		return msg[idx+2:]
	}
	return msg
}
