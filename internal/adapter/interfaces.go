// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Rasul Khiriev

// Package adapter provides transport-layer abstractions for communicating with
// the GoPassKeeper server.
//
// The primary abstraction is [ServerAdapter], which decouples the service layer
// from the underlying protocol. The package currently ships an HTTP/REST
// implementation ([NewHTTPServerAdapter]); a gRPC implementation is reserved
// for future use in grpc.go.
//
// Error values defined in errors.go are mapped from HTTP status codes by
// mapHTTPError so that callers can use [errors.Is] for transport-agnostic error
// handling (e.g. [ErrConflict] for 409, [ErrUnauthorized] for 401).
package adapter

import (
	"context"

	"github.com/MKhiriev/go-pass-keeper/models"
)

//go:generate mockgen -source=interfaces.go -destination=../mock/server_adapter_mock.go -package=mock

// ServerAdapter defines transport-agnostic communication with the GoPassKeeper
// server. Implementations are responsible for serialisation, authentication
// header management, and mapping transport-level errors to the sentinel values
// defined in this package.
type ServerAdapter interface {
	// SetToken stores the bearer token that will be attached to all subsequent
	// authenticated requests. It should be called immediately after a
	// successful Register or Login.
	SetToken(token string)

	// Token returns the bearer token currently stored in the adapter, or an
	// empty string if no token has been set yet.
	Token() string

	// Register sends a registration request to the server with the provided
	// user credentials. On success it stores the returned bearer token via
	// SetToken and returns the user value. Returns an error if the request
	// fails or the server responds with a non-2xx status.
	Register(ctx context.Context, user models.User) (models.User, error)

	// RequestSalt fetches the encryption salt that was stored for user.Login
	// during registration. The salt is needed to derive the KEK before the
	// auth hash can be computed for Login. Returns a partial [models.User]
	// containing only Login and EncryptionSalt.
	RequestSalt(ctx context.Context, user models.User) (models.User, error)

	// Login authenticates the user with the server using the pre-computed auth
	// hash. On success it stores the returned bearer token via SetToken and
	// returns the fully populated server-side user record (including the
	// encrypted master key). Returns an error if the request fails or the
	// server responds with a non-2xx status.
	Login(ctx context.Context, user models.User) (models.User, error)

	// Upload sends one or more new vault items to the server in a single
	// request. A transport integrity hash covering the payload is computed and
	// attached to the request automatically. Returns an error if the request
	// or the server response indicates failure.
	Upload(ctx context.Context, req models.UploadRequest) error

	// Download retrieves vault items identified by req.ClientSideIDs from the
	// server. Returns the full [models.PrivateData] slice, including encrypted
	// payloads. Returns an error if the request fails or the response cannot
	// be decoded.
	Download(ctx context.Context, req models.DownloadRequest) ([]models.PrivateData, error)

	// Update pushes a batch of partial vault-item updates to the server. A
	// transport integrity hash is computed automatically. Returns [ErrConflict]
	// (wrapped) if the server detects an optimistic-locking conflict, or
	// another error if the request fails.
	Update(ctx context.Context, req models.UpdateRequest) error

	// Delete sends a soft-delete request for one or more vault items to the
	// server. Returns [ErrConflict] (wrapped) on a version conflict, or
	// another error if the request fails.
	Delete(ctx context.Context, req models.DeleteRequest) error

	// GetServerStates fetches lightweight state descriptors
	// (ClientSideID, Hash, Version, Deleted, UpdatedAt) for all vault items
	// owned by userID from the server. Used by the sync planner to compare
	// server and client state without downloading full encrypted payloads.
	GetServerStates(ctx context.Context, userID int64) ([]models.PrivateDataState, error)
}
