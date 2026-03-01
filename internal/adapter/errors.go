// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Rasul Khiriev

package adapter

import "errors"

// Sentinel errors produced by adapter implementations when the server returns a
// non-2xx HTTP status code. Callers should use [errors.Is] to distinguish them,
// e.g. [errors.Is](err, [ErrConflict]) to detect an optimistic-locking conflict.
var (
	// ErrBadRequest is returned when the server responds with HTTP 400,
	// indicating malformed or logically invalid request data.
	ErrBadRequest = errors.New("bad request")

	// ErrUnauthorized is returned when the server responds with HTTP 401,
	// indicating that the request lacks valid authentication credentials.
	ErrUnauthorized = errors.New("unauthorized")

	// ErrForbidden is returned when the server responds with HTTP 403,
	// indicating that the authenticated user does not have permission to
	// perform the requested operation.
	ErrForbidden = errors.New("forbidden")

	// ErrNotFound is returned when the server responds with HTTP 404,
	// indicating that the requested resource does not exist.
	ErrNotFound = errors.New("not found")

	// ErrConflict is returned when the server responds with HTTP 409,
	// indicating an optimistic-locking conflict: the version provided by the
	// client no longer matches the current server version.
	ErrConflict = errors.New("conflict")

	// ErrBadGateway is returned when the server responds with HTTP 502,
	// typically indicating an upstream service is unreachable.
	ErrBadGateway = errors.New("bad gateway")

	// ErrInternalServerError is returned when the server responds with
	// HTTP 500, indicating an unexpected server-side failure.
	ErrInternalServerError = errors.New("internal server error")
)
