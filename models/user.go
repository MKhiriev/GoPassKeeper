package models

import "time"

// User represents an account entity used for authentication and authorization.
// It contains identity attributes and credential-related data.
// Sensitive fields must never be exposed outside trusted boundaries.
type User struct {
	// UserID is the internal unique identifier of the user.
	// It is not exposed via JSON and is used only at the persistence layer.
	UserID int64 `json:"-"`

	// Login is the unique user login identifier.
	// Typically used during authentication.
	Login string `json:"login"`

	// Name is the display name of the user.
	// It is non-sensitive and may be shown in UI.
	Name string `json:"name"`

	// MasterPassword stores the user's master password representation.
	// This value MUST be a derived value (hash/KDF output), never plaintext.
	// It is used only for authentication and key derivation.
	MasterPassword string `json:"master_password"`

	// MasterPasswordHint is an optional user-provided hint.
	// It is never exposed via JSON and must not contain sensitive data.
	MasterPasswordHint string `json:"-"`

	// CreatedAt is the timestamp when the user account was created.
	// Used for auditing and lifecycle management.
	CreatedAt time.Time `json:"created_at"`
}

// TableName returns the name of the database table
// associated with the User model.
func (u User) TableName() string {
	return "users"
}
