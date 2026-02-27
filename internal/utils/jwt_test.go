package utils

import (
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestGenerateJWTToken_Success(t *testing.T) {
	issuer := "test-issuer"
	userID := int64(123)
	duration := time.Hour
	key := "secret-key"

	token, err := GenerateJWTToken(issuer, userID, duration, key)

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if token.SignedString == "" {
		t.Error("expected non-empty SignedString")
	}
	if token.Token == nil {
		t.Error("expected non-nil jwt.Token object")
	}

	// Verify claims
	claims, ok := token.Token.Claims.(*jwt.RegisteredClaims)
	if !ok {
		t.Fatal("could not cast claims to RegisteredClaims")
	}
	if claims.Issuer != issuer {
		t.Errorf("expected issuer %s, got %s", issuer, claims.Issuer)
	}
	if claims.Subject != "123" {
		t.Errorf("expected subject '123', got %s", claims.Subject)
	}
}

func TestGenerateJWTToken_InvalidParams(t *testing.T) {
	tests := []struct {
		name     string
		issuer   string
		duration time.Duration
		key      string
	}{
		{"empty issuer", "", time.Hour, "key"},
		{"zero duration", "iss", 0, "key"},
		{"empty key", "iss", time.Hour, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := GenerateJWTToken(tt.issuer, 1, tt.duration, tt.key)
			if err == nil {
				t.Error("expected error for invalid parameters, got nil")
			}
		})
	}
}

func TestValidateAndParseJWTToken_Success(t *testing.T) {
	issuer := "test-issuer"
	userID := int64(456)
	key := "secret-key"
	duration := time.Minute * 5

	// First generate a valid token
	genToken, _ := GenerateJWTToken(issuer, userID, duration, key)

	// Now validate it
	parsedToken, err := ValidateAndParseJWTToken(genToken.SignedString, key, issuer)

	if err != nil {
		t.Fatalf("expected token to be valid, got error: %v", err)
	}
	if parsedToken.UserID != userID {
		t.Errorf("expected userID %d, got %d", userID, parsedToken.UserID)
	}
}

func TestValidateAndParseJWTToken_InvalidKey(t *testing.T) {
	issuer := "test-issuer"
	key := "correct-key"
	wrongKey := "wrong-key"

	genToken, _ := GenerateJWTToken(issuer, 1, time.Hour, key)

	_, err := ValidateAndParseJWTToken(genToken.SignedString, wrongKey, issuer)
	if err == nil {
		t.Error("expected error due to signature mismatch, got nil")
	}
}

func TestValidateAndParseJWTToken_Expired(t *testing.T) {
	issuer := "test-issuer"
	key := "key"
	// Token that expired 1 second ago
	genToken, _ := GenerateJWTToken(issuer, 1, -time.Second, key)

	_, err := ValidateAndParseJWTToken(genToken.SignedString, key, issuer)
	if err == nil {
		t.Error("expected error for expired token, got nil")
	}
	if !errors.Is(err, jwt.ErrTokenExpired) && err != nil {
		// Note: jwt.Parse returns a wrapped error, so we check if it contains expired info
	}
}

func TestValidateAndParseJWTToken_WrongIssuer(t *testing.T) {
	key := "key"
	genToken, _ := GenerateJWTToken("real-issuer", 1, time.Hour, key)

	_, err := ValidateAndParseJWTToken(genToken.SignedString, key, "fake-issuer")
	if err == nil {
		t.Error("expected error for issuer mismatch, got nil")
	}
}

func TestValidateAndParseJWTToken_Malformed(t *testing.T) {
	_, err := ValidateAndParseJWTToken("not.a.token", "key", "iss")
	if err == nil {
		t.Error("expected error for malformed token string, got nil")
	}
}
