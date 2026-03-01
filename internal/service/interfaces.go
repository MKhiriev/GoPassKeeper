// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Rasul Khiriev

// Package service defines the core business logic interfaces and service
// implementations for the go-pass-keeper application.
//
// The package is organized around three primary domains:
//   - Private data management: uploading, downloading, syncing, updating, and
//     deleting encrypted vault items owned by a user.
//   - Authentication: user registration, login, and JWT token lifecycle.
//   - Application metadata: exposing runtime information such as the app version.
//
// All service interfaces accept a context.Context as the first argument to
// support cancellation, deadlines, and request-scoped values (e.g. user ID).
//
// Middleware composition is supported via PrivateDataServiceWrapper, which
// allows decorating a PrivateDataService with cross-cutting concerns such as
// input validation or structured logging without modifying the core implementation.
package service

import (
	"context"

	"github.com/MKhiriev/go-pass-keeper/models"
)

// PrivateDataService defines the contract for managing encrypted private data
// (vault items) on behalf of authenticated users.
//
// All mutating operations accept structured request objects that carry the
// caller's user ID alongside the payload, enabling the service layer to enforce
// ownership and authorization rules independently of the transport layer.
type PrivateDataService interface {
	// UploadPrivateData persists a batch of new encrypted vault items described
	// by data. All items in the batch must belong to the same owner and must
	// have a zero version (initial creation semantics).
	// Returns an error if validation fails or the storage layer rejects the write.
	UploadPrivateData(ctx context.Context, data models.UploadRequest) error

	// DownloadPrivateData retrieves a filtered set of vault items matching the
	// criteria in downloadRequests (e.g. specific client-side IDs).
	// Returns the matching items or an error if the query fails.
	DownloadPrivateData(ctx context.Context, downloadRequests models.DownloadRequest) ([]models.PrivateData, error)

	// DownloadAllPrivateData retrieves every vault item owned by userID.
	// Returns the full collection or an error if the query fails.
	DownloadAllPrivateData(ctx context.Context, userID int64) ([]models.PrivateData, error)

	// DownloadUserPrivateDataStates returns lightweight state descriptors
	// (client-side ID + version) for all vault items owned by userID.
	// Clients use these states to detect which items need to be synced.
	DownloadUserPrivateDataStates(ctx context.Context, userID int64) ([]models.PrivateDataState, error)

	// DownloadSpecificUserPrivateDataStates returns state descriptors only for
	// the vault items whose client-side IDs are listed in syncRequest.
	// This is the targeted variant of DownloadUserPrivateDataStates used during
	// incremental sync operations.
	DownloadSpecificUserPrivateDataStates(ctx context.Context, syncRequest models.SyncRequest) ([]models.PrivateDataState, error)

	// UpdatePrivateData applies a batch of partial updates described by
	// updateRequests to existing vault items. Each update carries the expected
	// current version for optimistic concurrency control.
	// Returns an error if validation fails, a version conflict is detected, or
	// the storage layer rejects the write.
	UpdatePrivateData(ctx context.Context, updateRequests models.UpdateRequest) error

	// DeletePrivateData soft-deletes the vault items listed in deleteRequests.
	// Returns an error if validation fails or the storage layer rejects the operation.
	DeletePrivateData(ctx context.Context, deleteRequests models.DeleteRequest) error
}

// SyncService defines the contract for computing a client-server synchronisation plan.
//
// It operates purely on lightweight state descriptors (PrivateDataState) rather
// than full vault payloads, so no decryption is required at this stage.
// The resulting SyncPlan tells the caller exactly which items to download,
// upload, update, or delete on each side.
type SyncService interface {
	// BuildSyncPlan compares serverData (states fetched from the server) against
	// clientData (states read from the local database) and returns a SyncPlan
	// that classifies every item into one of five mutually exclusive action
	// categories:
	//
	//   Download     — fetch from server (new or newer version)
	//   Upload       — push to server   (exists only on client, never synced)
	//   Update       — write to server  (client version is ahead or hash diverged)
	//   DeleteClient — remove locally  (server holds a newer soft-deleted version)
	//   DeleteServer — remove remotely (client holds a newer soft-deleted version)
	//
	// Items that are identical on both sides produce no entry in the plan.
	// ctx is forwarded to allow cancellation of any I/O performed internally.
	BuildSyncPlan(ctx context.Context, serverData, clientData []models.PrivateDataState) (models.SyncPlan, error)
}

// AuthService defines the contract for user authentication and JWT token management.
//
// It covers the full authentication lifecycle: account creation, credential
// verification, token issuance, and token parsing/validation.
type AuthService interface {
	// RegisterUser creates a new user account using the credentials in user.
	// Returns the persisted user (with a server-assigned ID) or an error if the
	// login is already taken or the storage layer fails.
	RegisterUser(ctx context.Context, user models.User) (models.User, error)

	// Login verifies the credentials in user against the stored account.
	// Returns the authenticated user record or an error if the credentials are
	// invalid or the user does not exist.
	Login(ctx context.Context, user models.User) (models.User, error)

	// Params used to get user encryption salt
	Params(ctx context.Context, user models.User) (models.User, error)

	// CreateToken issues a signed JWT for the given user.
	// Returns the token model containing the raw token string and its claims,
	// or an error if token generation fails.
	CreateToken(ctx context.Context, user models.User) (models.Token, error)

	// ParseToken validates and parses the raw JWT string tokenString.
	// Returns the decoded token model on success, or an error if the token is
	// malformed, expired, or signed with an unexpected key.
	ParseToken(ctx context.Context, tokenString string) (models.Token, error)
}

// AppInfoService defines the contract for exposing application-level metadata.
type AppInfoService interface {
	// GetAppVersion returns the current semantic version string of the running
	// application (e.g. "1.2.3" or "dev").
	GetAppVersion(ctx context.Context) string
}

// PrivateDataServiceWrapper defines the middleware composition contract for
// PrivateDataService implementations.
//
// Implementations decorate an existing PrivateDataService to inject
// cross-cutting behavior — such as input validation, structured logging, or
// metrics collection — without modifying the underlying service logic.
//
// Usage pattern:
//
//	var svc PrivateDataService = NewCoreService(store)
//	svc = validationWrapper.Wrap(svc)
//	svc = loggingWrapper.Wrap(svc)
type PrivateDataServiceWrapper interface {
	// Wrap accepts an inner PrivateDataService and returns a new
	// PrivateDataService that applies additional behavior around each method call.
	Wrap(PrivateDataService) PrivateDataService
}
