package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"testing"
)

func TestGenerateSalt_LengthAndRandomness(t *testing.T) {
	svc := NewKeyChainService()

	s1, err := svc.GenerateEncryptionSalt()
	if err != nil {
		t.Fatalf("GenerateEncryptionSalt error: %v", err)
	}
	s2, err := svc.GenerateEncryptionSalt()
	if err != nil {
		t.Fatalf("GenerateEncryptionSalt error: %v", err)
	}

	if len(s1) != 16 {
		t.Fatalf("salt length = %d, want 16", len(s1))
	}
	if len(s2) != 16 {
		t.Fatalf("salt length = %d, want 16", len(s2))
	}
	if bytes.Equal(s1, s2) {
		t.Fatalf("expected salts to differ, but they are equal")
	}
}

func TestGenerateDEK_LengthAndRandomness(t *testing.T) {
	svc := NewKeyChainService()

	d1, err := svc.GenerateDEK()
	if err != nil {
		t.Fatalf("GenerateDEK error: %v", err)
	}
	d2, err := svc.GenerateDEK()
	if err != nil {
		t.Fatalf("GenerateDEK error: %v", err)
	}

	if len(d1) != 32 {
		t.Fatalf("DEK length = %d, want 32", len(d1))
	}
	if len(d2) != 32 {
		t.Fatalf("DEK length = %d, want 32", len(d2))
	}
	if bytes.Equal(d1, d2) {
		t.Fatalf("expected DEKs to differ, but they are equal")
	}
}

func TestGenerateKEK_DeterministicForSameInputs(t *testing.T) {
	svc := NewKeyChainService()

	password := "correct horse battery staple"
	salt := bytes.Repeat([]byte{0xAB}, 16)

	k1 := svc.GenerateKEK(password, salt)
	k2 := svc.GenerateKEK(password, salt)

	if len(k1) != 32 {
		t.Fatalf("KEK length = %d, want 32", len(k1))
	}
	if !bytes.Equal(k1, k2) {
		t.Fatalf("expected KEKs to match for same password+salt")
	}
}

func TestGenerateKEK_DifferentSaltProducesDifferentKEK(t *testing.T) {
	svc := NewKeyChainService()

	password := "same password"
	salt1 := bytes.Repeat([]byte{0x01}, 16)
	salt2 := bytes.Repeat([]byte{0x02}, 16)

	k1 := svc.GenerateKEK(password, salt1)
	k2 := svc.GenerateKEK(password, salt2)

	if bytes.Equal(k1, k2) {
		t.Fatalf("expected different KEKs for different salts")
	}
}

func TestGenerateAuthHash_DeterministicAndSeparated(t *testing.T) {
	svc := NewKeyChainService()

	kek := bytes.Repeat([]byte{0x11}, 32)

	a1 := svc.GenerateAuthHash(kek, "auth-purpose")
	a2 := svc.GenerateAuthHash(kek, "auth-purpose")
	if !bytes.Equal(a1, a2) {
		t.Fatalf("expected AuthHash to be deterministic")
	}

	a3 := svc.GenerateAuthHash(kek, "other-purpose")
	if bytes.Equal(a1, a3) {
		t.Fatalf("expected AuthHash to differ for different authSalt")
	}
}

// Use a separate test to avoid confusing byte literals.
func TestGetEncryptedDEK_DecryptRoundTrip(t *testing.T) {
	svc := NewKeyChainService()

	dek := bytes.Repeat([]byte{0xDD}, 32)
	kek := bytes.Repeat([]byte{0x2A}, 32) // valid AES-256 key length

	blob, err := svc.GetEncryptedDEK(dek, kek)
	if err != nil {
		t.Fatalf("GetEncryptedDEK error: %v", err)
	}

	// Reconstruct AES-GCM and decrypt to verify round-trip.
	block, err := aes.NewCipher(kek)
	if err != nil {
		t.Fatalf("aes.NewCipher error: %v", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		t.Fatalf("cipher.NewGCM error: %v", err)
	}

	nonceSize := gcm.NonceSize()
	if len(blob) <= nonceSize {
		t.Fatalf("blob too short: got %d, want > %d", len(blob), nonceSize)
	}

	nonce := blob[:nonceSize]
	ct := blob[nonceSize:]

	plain, err := gcm.Open(nil, nonce, ct, nil)
	if err != nil {
		t.Fatalf("gcm.Open error: %v", err)
	}

	if !bytes.Equal(plain, dek) {
		t.Fatalf("decrypted DEK mismatch")
	}
}

func TestGetEncryptedDEK_NonceRandomness(t *testing.T) {
	svc := NewKeyChainService()

	dek := bytes.Repeat([]byte{0xDD}, 32)
	kek := bytes.Repeat([]byte{0x2A}, 32)

	blob1, err := svc.GetEncryptedDEK(dek, kek)
	if err != nil {
		t.Fatalf("GetEncryptedDEK error: %v", err)
	}
	blob2, err := svc.GetEncryptedDEK(dek, kek)
	if err != nil {
		t.Fatalf("GetEncryptedDEK error: %v", err)
	}

	block, _ := aes.NewCipher(kek)
	gcm, _ := cipher.NewGCM(block)
	nonceSize := gcm.NonceSize()

	n1 := blob1[:nonceSize]
	n2 := blob2[:nonceSize]

	if bytes.Equal(n1, n2) {
		t.Fatalf("expected different nonces for two encryptions")
	}

	// With different nonces, the full blobs should almost certainly differ.
	if bytes.Equal(blob1, blob2) {
		t.Fatalf("expected different ciphertext blobs for two encryptions")
	}
}
