package service

import "errors"

var (
	ErrInvalidDataProvided = errors.New("invalid data provided")
	ErrWrongPassword       = errors.New("wrong password")

	ErrTokenIsExpired = errors.New("token is expired")

	ErrValidationNoPrivateDataProvided      = errors.New("no private data provided")
	ErrValidationNoDownloadRequestsProvided = errors.New("no download requests provided")
	ErrValidationNoUpdateRequestsProvided   = errors.New("no update requests provided")
	ErrValidationNoDeleteRequestsProvided   = errors.New("no delete requests provided")
	ErrValidationNoUserID                   = errors.New("no user ID for private data was given")
)
