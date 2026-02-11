package validators

import "errors"

var (
	ErrUnsupportedType = errors.New("unsupported type for validation")
	ErrUnknownField    = errors.New("unknown field for validation")
)
