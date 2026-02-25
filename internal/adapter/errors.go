package adapter

import "errors"

var (
	ErrBadRequest          = errors.New("bad request")
	ErrUnauthorized        = errors.New("unauthorized")
	ErrForbidden           = errors.New("forbidden")
	ErrNotFound            = errors.New("not found")
	ErrConflict            = errors.New("conflict")
	ErrBadGateway          = errors.New("bad gateway")
	ErrInternalServerError = errors.New("internal server error")
)
