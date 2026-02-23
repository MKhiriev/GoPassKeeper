package validators

import "errors"

var (
	// ErrUnsupportedType is returned when a value of an unsupported type
	// is passed to a validator that cannot handle it.
	ErrUnsupportedType = errors.New("unsupported type for validation")

	// ErrUnknownField is returned when a field name provided for validation
	// does not match any known or expected field.
	ErrUnknownField = errors.New("unknown field for validation")

	// ErrInvalidUserID is returned when the provided user ID
	// is missing, zero, or otherwise invalid.
	ErrInvalidUserID = errors.New("invalid user ID")

	// ErrInvalidClientSideID is returned when the client-generated record ID
	// is missing or does not meet the required format.
	ErrInvalidClientSideID = errors.New("invalid client side id")

	// ErrInvalidHash is returned when the hash of a record
	// is empty, malformed, or fails integrity checks.
	ErrInvalidHash = errors.New("invalid hash")

	// ErrInvalidUpdatedRecordHash is returned when the hash provided
	// for an updated record is missing or does not match the expected value.
	ErrInvalidUpdatedRecordHash = errors.New("invalid updated record hash")

	// ErrEmptyMetadata is returned when a required metadata field
	// is not provided in the request or entity.
	ErrEmptyMetadata = errors.New("metadata is required")

	// ErrEmptyData is returned when a required data payload
	// is absent or empty in the request or entity.
	ErrEmptyData = errors.New("data is required")

	// ErrInvalidType is returned when the data type field
	// contains an unrecognized or unsupported value.
	ErrInvalidType = errors.New("invalid data type")

	// ErrEmptyIDs is returned when an operation requires a non-empty list
	// of record IDs but an empty slice is provided.
	ErrEmptyIDs = errors.New("IDs list cannot be empty")

	// ErrNoDeleteEntries is returned when a delete operation is requested
	// but the list of entries to delete is empty.
	ErrNoDeleteEntries = errors.New("delete entries list cannot be empty")

	// ErrNoFieldsToUpdate is returned when an update request is submitted
	// without specifying any fields to be changed.
	ErrNoFieldsToUpdate = errors.New("at least one field must be provided for update")

	// ErrEmptyPrivateData is returned when an operation requires a non-empty
	// list of private data entries but an empty slice is provided.
	ErrEmptyPrivateData = errors.New("private data list cannot be empty")

	// ErrEmptyUpdates is returned when an update operation is requested
	// but the list of updates to apply is empty.
	ErrEmptyUpdates = errors.New("updates list cannot be empty")

	// ErrInvalidVersion is returned when the version field of a record
	// is not matching version from updating,deleting request.
	ErrInvalidVersion = errors.New("invalid Version")

	// ErrInvalidUpdateVersion is returned when the version field provided
	// in an update request is not zero.
	ErrInvalidUpdateVersion = errors.New("invalid Update Version")
)
