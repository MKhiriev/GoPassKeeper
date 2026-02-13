package models

// DownloadRequest represents search criteria for querying vault items.
// Only unencrypted fields can be used for database-level filtering.
type DownloadRequest struct {
	// UserID filters records by owner.
	// Required in most cases to ensure data isolation.
	UserID int64 `json:"user_id,omitempty"`

	// IDs filters by specific record identifiers.
	// Useful for batch operations or direct lookups.
	IDs []int64 `json:"ids,omitempty"`

	// Types filters by data type (login, note, card, binary).
	// Allows users to view only specific categories of secrets.
	Types []DataType `json:"types,omitempty"`
}

// UpdateRequest represents a batch update request for vault items.
type UpdateRequest struct {
	PrivateDataUpdates []PrivateDataUpdate `json:"private_data_updates"`
}

// DeleteRequest represents criteria for deleting vault items.
// Supports both single and batch deletion operations.
type DeleteRequest struct {
	// UserID is the owner of the data to delete.
	UserID int64 `json:"user_id"`

	// IDs contains specific record identifiers to delete.
	IDs []int64 `json:"ids"`
}

// PrivateDataUpdate represents criteria for updating a single vault item.
// Only non-nil fields will be updated (partial update support).
type PrivateDataUpdate struct {
	// ID is the unique identifier of the record to update.
	// Required.
	ID int64 `json:"id"`

	// UserID is the owner of the data to update.
	// Required for data isolation and security.
	UserID int64 `json:"user_id"`

	// Metadata contains updated non-sensitive descriptive information.
	// If nil, the field will not be updated.
	Metadata *CipheredMetadata `json:"metadata,omitempty"`

	// Type defines the updated semantic type of the stored data.
	// If nil, the field will not be updated.
	Type *DataType `json:"type,omitempty"`

	// Data holds the updated encrypted payload.
	// If nil, the field will not be updated.
	Data *CipheredData `json:"data,omitempty"`

	// Notes contains updated optional user notes.
	// If nil, the field will not be updated.
	Notes *CipheredNotes `json:"notes,omitempty"`

	// AdditionalFields contains updated custom user-defined fields.
	// If nil, the field will not be updated.
	AdditionalFields *CipheredCustomFields `json:"additional_fields,omitempty"`
}
