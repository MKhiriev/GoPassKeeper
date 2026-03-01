// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Rasul Khiriev

// Package app contains shared application-layer constants used across the
// GoPassKeeper server handlers and middleware.
//
// All Msg* constants are human-readable message strings that are written into
// HTTP response bodies or log entries to describe the outcome of an operation.
// Keeping them in one place ensures consistent wording throughout the API.
package app

const (
	// MsgInvalidDataProvided is returned when the request body cannot be
	// decoded or fails basic validation (e.g. missing required fields).
	MsgInvalidDataProvided = "invalid data provided"

	// MsgInvalidLoginPassword is returned when the supplied login/password
	// combination does not match any existing user record.
	MsgInvalidLoginPassword = "invalid login/password"

	// MsgInternalServerError is returned when an unexpected server-side
	// failure occurs that the client cannot resolve.
	MsgInternalServerError = "internal server error"

	// MsgTokenIsExpired is returned when a JWT bearer token is syntactically
	// valid but its expiry time has passed.
	MsgTokenIsExpired = "token is expired"

	// MsgTokenIsExpiredOrInvalid is returned when a JWT bearer token is
	// either expired or cannot be verified (e.g. wrong signature).
	MsgTokenIsExpiredOrInvalid = "token is expired or invalid"

	// MsgNoPrivateDataProvided is returned when an upload or create request
	// contains an empty vault-item list.
	MsgNoPrivateDataProvided = "no private data provided"

	// MsgNoDownloadRequestsProvided is returned when a download request
	// contains no client-side IDs to fetch.
	MsgNoDownloadRequestsProvided = "no download requests provided"

	// MsgNoUpdateRequestsProvided is returned when an update request contains
	// an empty list of vault-item changes.
	MsgNoUpdateRequestsProvided = "no update requests provided"

	// MsgNoDeleteRequestsProvided is returned when a delete request contains
	// an empty list of entries to remove.
	MsgNoDeleteRequestsProvided = "no delete requests provided"

	// MsgNoUserIDProvided is returned when a handler requires a user ID (e.g.
	// extracted from the JWT claim) but none is present in the request
	// context.
	MsgNoUserIDProvided = "no user ID provided"

	// MsgNoClientIDsForSync is returned when a sync request arrives with no
	// client-side IDs to reconcile.
	MsgNoClientIDsForSync = "no client IDs provided for sync"

	// MsgEmptyClientIDForSync is returned when at least one entry in a sync
	// request has a blank (empty string) client-side ID.
	MsgEmptyClientIDForSync = "empty client ID provided for sync"

	// MsgAccessDenied is returned when the authenticated user attempts to
	// access or modify a resource that belongs to a different user.
	MsgAccessDenied = "access denied"

	// MsgVersionIsNotSpecified is returned when an update or delete request
	// omits the version field required for optimistic-locking checks.
	MsgVersionIsNotSpecified = "version is not specified"

	// MsgRegistrationFailed is returned when the registration handler
	// encounters an unexpected error that prevents account creation.
	MsgRegistrationFailed = "registration failed"

	// MsgLoginFailed is returned when the login handler encounters an
	// unexpected error that prevents issuing a session token.
	MsgLoginFailed = "login failed"

	// MsgLoginAlreadyExists is returned when a registration attempt is
	// rejected because the requested login is already in use.
	MsgLoginAlreadyExists = "login already exists"

	// MsgDataNotFound is returned when a read, update, or delete operation
	// targets a vault item that does not exist for the current user.
	MsgDataNotFound = "data not found"

	// MsgVersionConflict is returned when an optimistic-locking check fails:
	// the version supplied by the client no longer matches the server's
	// current version. The client should sync before retrying.
	MsgVersionConflict = "version conflict, please sync"
)
