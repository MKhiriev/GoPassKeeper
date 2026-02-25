package adapter

import "errors"

var (
	ErrUnauthorized    = errors.New("client unauthorized")
	ErrVersionConflict = errors.New("version conflict")
)
