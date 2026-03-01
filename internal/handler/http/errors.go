// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Rasul Khiriev

package http

import "errors"

// Sentinel errors used by the authentication middleware when parsing the
// "Authorization" HTTP header. Callers can match against them with [errors.Is].
var (
	// ErrEmptyAuthorizationHeader is returned by the auth middleware when the
	// incoming request does not include an "Authorization" header at all.
	ErrEmptyAuthorizationHeader = errors.New("empty `Authorization` header")

	// ErrInvalidAuthorizationHeader is returned when the "Authorization"
	// header is present but cannot be split into at least two space-separated
	// parts (i.e. the token value is missing entirely).
	ErrInvalidAuthorizationHeader = errors.New("invalid `Authorization` header")

	// ErrEmptyToken is returned when the "Authorization" header contains the
	// expected scheme prefix but the token value itself is an empty string.
	ErrEmptyToken = errors.New("empty token in `Authorization` header")
)
