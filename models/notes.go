package models

// Notes represents an optional textual annotation attached to PrivateData.
type Notes struct {
	// IsEncrypted indicates whether the notes content is encrypted.
	IsEncrypted bool

	// Notes contains the note text.
	// When IsEncrypted is true, this value is stored in encrypted form.
	Notes string
}
