package service

import "errors"

var (
	ErrInvalidDataProvided = errors.New("invalid data provided")
	ErrWrongPassword       = errors.New("wrong password")

	ErrTokenIsExpired          = errors.New("token is expired")
	ErrTokenIsExpiredOrInvalid = errors.New("token is expired/invalid")

	ErrValidationNoPrivateDataProvided      = errors.New("no private data provided")
	ErrValidationNoDownloadRequestsProvided = errors.New("no download requests provided")
	ErrValidationNoUpdateRequestsProvided   = errors.New("no update requests provided")
	ErrValidationNoDeleteRequestsProvided   = errors.New("no delete requests provided")
	ErrValidationNoUserID                   = errors.New("no user ID for private data was given")

	ErrValidationNoClientIDsProvidedForSyncRequests   = errors.New("no client side IDs provided for sync request")
	ErrValidationEmptyClientIDProvidedForSyncRequests = errors.New("empty client side ID provided for sync request")

	ErrUnauthorizedAccessToDifferentUserData = errors.New("unauthorized access to different user's data")

	ErrVersionIsNotSpecified = errors.New("version is not specified")
)
