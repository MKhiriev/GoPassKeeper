package models

// DataType defines the semantic type of encrypted data
// stored inside PrivateData.Data.
// The value determines how the decrypted payload must be interpreted.
type DataType int

const (
	// LoginPassword represents authentication credentials
	// such as username, password, URIs, and optional TOTP secret.
	LoginPassword DataType = 1

	// Text represents arbitrary textual data
	// stored as a secure note or free-form secret.
	Text DataType = 2

	// Binary represents encrypted binary attachments.
	// The actual file content is stored separately as an encrypted blob;
	// this structure contains only metadata and encryption references.
	Binary DataType = 3

	// BankCard represents payment card information.
	// All fields are considered highly sensitive and always encrypted.
	BankCard DataType = 4
)

// LoginData represents decrypted login credentials.
// This structure is serialized to JSON and stored encrypted
// inside PrivateData.Data when DataType is LoginPassword.
type LoginData struct {
	// Username is the login identifier used for authentication.
	Username string `json:"username"`

	// Password is the secret credential associated with the username.
	Password string `json:"password"`

	// URIs defines one or more resources where the credentials apply.
	URIs []LoginURI `json:"uris,omitempty"`

	// TOTP contains an optional time-based one-time password seed.
	// When present, it is used to generate 2FA codes.
	TOTP *string `json:"totp,omitempty"`
}

// LoginURI represents a single resource matching rule
// associated with a login entry.
type LoginURI struct {
	// URI is the target resource (domain, URL, or application identifier).
	URI string `json:"uri"`

	// Match defines the matching strategy used to associate
	// the login with the given URI.
	Match int `json:"match"`
}

// TextData represents decrypted free-form textual content.
// Used for secure notes or arbitrary secret text.
type TextData struct {
	// Text contains the textual payload.
	Text string `json:"text"`
}

// BinaryData represents metadata for an encrypted binary attachment.
// The actual binary content is stored outside the main data record.
type BinaryData struct {
	// ID is the unique identifier of the binary object in blob storage.
	ID string `json:"id"`

	// FileName is the original name of the attached file.
	FileName string `json:"fileName"`

	// Size is the size of the file in bytes.
	Size int64 `json:"size"`

	// Key is the encryption key or key reference
	// used to encrypt and decrypt the binary content.
	Key string `json:"key"`
}

// BankCardData represents decrypted payment card information.
// This structure is serialized and stored encrypted.
type BankCardData struct {
	// CardholderName is the name printed on the card.
	CardholderName string `json:"cardholderName"`

	// Number is the primary account number (PAN) of the card.
	Number string `json:"number"`

	// Brand identifies the card network (e.g. Visa, MasterCard).
	Brand string `json:"brand"`

	// ExpMonth is the card expiration month.
	ExpMonth string `json:"expMonth"`

	// ExpYear is the card expiration year.
	ExpYear string `json:"expYear"`

	// Code is the card security code (CVV/CVC).
	Code string `json:"code"`
}
