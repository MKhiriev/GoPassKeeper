package models

import "time"

// PrivateData represents a single vault item.
// It is the primary persistence model for all sensitive user data.
// All confidential payloads are stored encrypted and opaque to the database.
type PrivateData struct {
	// ID is the unique identifier of the record in the database.
	ID int64 `json:"id"`

	// UserID is the owner of this private data entry.
	UserID int64 `json:"user_id"`

	// Metadata contains non-sensitive descriptive information
	// such as display name and folder placement.
	// Metadata is stored in DB as an encrypted string.
	Metadata CipheredMetadata `json:"metadata"`

	// Type defines the semantic type of the stored data
	// (e.g. login, secure note, card, binary reference).
	Type DataType `json:"type"`

	// Data holds the encrypted payload.
	// The database treats this field as an opaque string.
	// Data is stored in DB as an encrypted string.
	Data CipheredData `json:"data"`

	// Notes contains optional user notes.
	// Notes may be stored encrypted or in plain form depending on configuration.
	// Notes are stored in DB as an encrypted string.
	Notes *CipheredNotes `json:"notes,omitempty"`

	// AdditionalFields contains optional custom user-defined fields.
	// Each field is independently typed and encrypted.
	// AdditionalFields are stored in DB as an encrypted string.
	AdditionalFields *CipheredCustomFields `json:"fields,omitempty"`

	// CreatedAt is the timestamp when the record was created.
	CreatedAt *time.Time `json:"created_at"`

	// UpdatedAt is the timestamp of the last modification.
	UpdatedAt *time.Time `json:"updated_at"`
}

// Metadata describes non-secret attributes of a PrivateData item.
// These fields are used for organization and presentation only.
type Metadata struct {
	// Name is the human-readable display name of the item.
	Name string

	// Folder is an optional logical container used to group items.
	Folder *string
}

// Notes represents an optional textual annotation attached to PrivateData.
type Notes struct {
	// IsEncrypted indicates whether the notes content is encrypted.
	IsEncrypted bool

	// Notes contains the note text.
	// When IsEncrypted is true, this value is stored in encrypted form.
	Notes string
}

// CustomField represents a user-defined field attached to PrivateData.
// Each custom field has its own semantic type and encrypted value.
type CustomField struct {
	// Type defines the data type of the custom field.
	Type DataType `json:"type"`

	// Data contains the encrypted value of the custom field.
	Data CipheredData `json:"data"`
}

// TableName returns the name of the database table
// associated with the PrivateData model.
func (u *PrivateData) TableName() string {
	return "ciphers"
}
