package utils

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/MKhiriev/go-pass-keeper/models"
	"github.com/golang-jwt/jwt/v5"
)

// GenerateJWTToken creates a signed HMAC-SHA256 JWT token with the given parameters.
//
// The token includes the following standard claims:
//   - Issuer    (iss): identifies the service that issued the token
//   - Subject   (sub): the user ID encoded as a string
//   - IssuedAt  (iat): the current time
//   - ExpiresAt (exp): the current time plus tokenDuration
//
// All parameters are required. Returns an error if any of them are empty or zero.
//
// Parameters:
//
//	issuer        - identifier of the token issuer (e.g. service name)
//	userID        - ID of the user the token is issued for
//	tokenDuration - how long the token remains valid
//	signKey       - secret key used to sign the token with HMAC-SHA256
//
// Returns:
//
//	models.Token - contains the signed token string and the jwt.Token object
//	error        - non-nil if parameters are invalid or signing fails
//
// Example usage:
//
//	token, err := utils.GenerateJWTToken("my-service", 42, time.Hour, "secret")
func GenerateJWTToken(issuer string, userID int64, tokenDuration time.Duration, signKey string) (models.Token, error) {
	if issuer == "" || tokenDuration == 0 || signKey == "" {
		return models.Token{}, errors.New("invalid params for generating JWT Token")
	}

	now := time.Now()
	claims := &jwt.RegisteredClaims{
		Issuer:    issuer,
		Subject:   strconv.FormatInt(userID, 10),
		ExpiresAt: jwt.NewNumericDate(now.Add(tokenDuration)),
		IssuedAt:  jwt.NewNumericDate(now),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(signKey))
	if err != nil {
		return models.Token{}, fmt.Errorf("error occurred during singing JWT token: %w", err)
	}

	return models.Token{Token: token, SignedString: tokenString}, nil
}

// ValidateAndParseJWTToken validates the given JWT token string and extracts its claims.
//
// Validation includes:
//   - Signature verification using the provided sign key
//   - Issuer (iss) claim check against the provided tokenIssuer
//   - Expiration (exp) claim check
//   - Subject (sub) claim presence and conversion to int64 UserID
//
// Parameters:
//
//	tokenString   - the raw signed JWT string to validate and parse
//	tokenSignKey  - secret key used to verify the token signature
//	tokenIssuer   - expected issuer value to validate against the iss claim
//
// Returns:
//
//	models.Token - contains the parsed jwt.Token object and the extracted UserID
//	error        - non-nil if validation fails, claims are missing, or subject cannot be parsed
//
// Example usage:
//
//	token, err := utils.ValidateAndParseJWTToken(rawToken, "secret", "my-service")
//	if err != nil {
//	    // handle invalid or expired token
//	}
func ValidateAndParseJWTToken(tokenString, tokenSignKey, tokenIssuer string) (models.Token, error) {
	token, err := jwt.ParseWithClaims(tokenString, &models.Token{}, func(token *jwt.Token) (any, error) {
		return []byte(tokenSignKey), nil
	}, jwt.WithIssuer(tokenIssuer))
	if err != nil {
		return models.Token{}, fmt.Errorf("error occurred validating and parsing token: %w", err)
	}

	userIDStr, err := token.Claims.GetSubject()
	if err != nil {
		return models.Token{}, fmt.Errorf("error occurred during getting subject from token: %w", err)
	}
	if userIDStr == "" {
		return models.Token{}, errors.New("empty subject error")
	}

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		return models.Token{}, fmt.Errorf("error occurred during converting subject to UserIDCtxKey: %w", err)
	}

	return models.Token{Token: token, UserID: userID}, err
}

func ParseBearerToken(authorizationHeader string) (string, error) {
	parts := strings.Split(strings.TrimSpace(authorizationHeader), " ")
	if len(parts) != 2 || parts[1] == "" {
		return "", errors.New("invalid authorization header")
	}
	return parts[1], nil
}

func ParseUserIDFromJWT(tokenString string) (int64, error) {
	token, _, err := jwt.NewParser().ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return 0, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, errors.New("invalid token claims")
	}

	sub, err := claims.GetSubject()
	if err != nil {
		return 0, err
	}

	id, err := strconv.ParseInt(sub, 10, 64)
	if err != nil {
		return 0, err
	}
	return id, nil
}
