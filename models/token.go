package models

import (
	"fmt"
	"strconv"

	"github.com/golang-jwt/jwt/v5"
)

// Token wraps a JWT token with convenience accessors for authentication flows.
//
// It embeds [jwt.Token] for low-level token operations (signing, parsing)
// and [jwt.RegisteredClaims] for standard claim access (subject, expiry, etc.).
//
// SignedString holds the compact serialized form of the token (header.payload.signature)
// ready to be transmitted in HTTP headers or stored on the client side.
//
// UserID is a cached, parsed copy of the "sub" (subject) claim converted to int64.
// It is typically populated after a successful call to [Token.GetUserID] or
// during token construction and avoids repeated string-to-int parsing.
type Token struct {
	// Token is the underlying JWT token used for signing and claim inspection.
	// Excluded from JSON serialization because only the compact string form
	// is meaningful outside the server process.
	*jwt.Token `json:"-"`

	// RegisteredClaims provides access to the standard JWT claim set
	// (sub, exp, iat, nbf, iss, aud, jti) as defined by RFC 7519.
	jwt.RegisteredClaims

	// SignedString is the compact JWS representation of the token
	// (base64url-encoded header.payload.signature).
	// Excluded from JSON serialization; use [Token.String] to retrieve it.
	SignedString string `json:"-"`

	// UserID is the owner identifier extracted from the "sub" claim.
	// Excluded from JSON serialization; it is an internal server-side cache.
	UserID int64 `json:"-"`
}

// GetUserID extracts the user identifier from the token's "sub" (subject) claim,
// parses it as a base-10 int64, and returns the result.
//
// Returns an error if the subject claim is missing, empty, or cannot be
// converted to int64.
func (t *Token) GetUserID() (int64, error) {
	userIDString, err := t.GetSubject()
	if err != nil {
		return 0, fmt.Errorf("error extracting UserID from token: %w", err)
	}

	userID, err := strconv.ParseInt(userIDString, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("error converting UserID from token to int64: %w", err)
	}

	return userID, nil
}

// String returns the compact JWS serialization of the token
// (the signed, base64url-encoded header.payload.signature string).
// It implements the [fmt.Stringer] interface.
func (t *Token) String() string {
	return t.SignedString
}
