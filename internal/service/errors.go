package service

import "errors"

var (
	// ErrInvalidDataProvided is returned when the caller supplies a request object
	// that fails basic structural or semantic validation (e.g. missing required
	// fields, malformed values).
	ErrInvalidDataProvided = errors.New("invalid data provided")

	// ErrWrongPassword is returned by the authentication service when the supplied
	// password does not match the stored credential hash for the given user.
	ErrWrongPassword = errors.New("wrong password")

	// ErrTokenIsExpired is returned when a JWT has passed its expiration time (exp
	// claim) but is otherwise structurally valid.
	ErrTokenIsExpired = errors.New("token is expired")

	// ErrTokenIsExpiredOrInvalid is returned when a JWT cannot be trusted — either
	// because it has expired or because its signature / claims are invalid.
	ErrTokenIsExpiredOrInvalid = errors.New("token is expired/invalid")

	// ErrValidationNoPrivateDataProvided is returned when an upload or mutation
	// request contains an empty list of private data items.
	ErrValidationNoPrivateDataProvided = errors.New("no private data provided")

	// ErrValidationNoDownloadRequestsProvided is returned when a download request
	// carries no item identifiers to fetch.
	ErrValidationNoDownloadRequestsProvided = errors.New("no download requests provided")

	// ErrValidationNoUpdateRequestsProvided is returned when an update request
	// carries no update entries.
	ErrValidationNoUpdateRequestsProvided = errors.New("no update requests provided")

	// ErrValidationNoDeleteRequestsProvided is returned when a delete request
	// carries no entries to remove.
	ErrValidationNoDeleteRequestsProvided = errors.New("no delete requests provided")

	// ErrValidationNoUserID is returned when a request that requires an owner
	// identity is submitted without a valid user ID (zero or negative value).
	ErrValidationNoUserID = errors.New("no user ID for private data was given")

	// ErrValidationNoClientIDsProvidedForSyncRequests is returned when a sync
	// request contains an empty slice of client-side item IDs.
	ErrValidationNoClientIDsProvidedForSyncRequests = errors.New("no client side IDs provided for sync request")

	// ErrValidationEmptyClientIDProvidedForSyncRequests is returned when a sync
	// request contains at least one blank (empty string) client-side item ID.
	ErrValidationEmptyClientIDProvidedForSyncRequests = errors.New("empty client side ID provided for sync request")

	// ErrUnauthorizedAccessToDifferentUserData is returned when the authenticated
	// caller attempts to read or modify vault items that belong to another user.
	ErrUnauthorizedAccessToDifferentUserData = errors.New("unauthorized access to different user's data")

	// ErrVersionIsNotSpecified is returned when an update or delete entry omits the
	// required version field, which is needed for optimistic concurrency control.
	ErrVersionIsNotSpecified = errors.New("version is not specified")
)

// client errors
var (
	ErrRegisterOnServer = errors.New("register user on server")
	ErrLoginOnServer    = errors.New("login on server")
)
