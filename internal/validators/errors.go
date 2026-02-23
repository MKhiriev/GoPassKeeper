package validators

import "errors"

var (
	ErrUnsupportedType = errors.New("unsupported type for validation")
	ErrUnknownField    = errors.New("unknown field for validation")

	ErrInvalidUserID            = errors.New("invalid user ID")
	ErrInvalidClientSideID      = errors.New("invalid client side id")
	ErrInvalidHash              = errors.New("invalid hash")
	ErrInvalidUpdatedRecordHash = errors.New("invalid updated record hash")
	ErrEmptyMetadata            = errors.New("metadata is required")
	ErrEmptyData                = errors.New("data is required")
	ErrInvalidType              = errors.New("invalid data type")
	ErrEmptyIDs                 = errors.New("IDs list cannot be empty")
	ErrNoDeleteEntries          = errors.New("delete entries list cannot be empty")
	ErrNoFieldsToUpdate         = errors.New("at least one field must be provided for update")
	ErrEmptyPrivateData         = errors.New("private data list cannot be empty")
	ErrEmptyUpdates             = errors.New("updates list cannot be empty")
	ErrInvalidVersion           = errors.New("invalid Version")
	ErrInvalidUpdateVersion     = errors.New("invalid Update Version")
)
