// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Rasul Khiriev

// Package crypto implements the client-side zero-knowledge cryptography layer
// for GoPassKeeper.
//
// # Key hierarchy
//
// The package follows a three-level key hierarchy that ensures the server never
// sees plaintext data or any key that could recover it:
//
//  1. DEK (data-encryption key) — a random 256-bit AES key generated once per
//     registration. It encrypts and decrypts all vault payloads on the client.
//
//  2. KEK (key-encryption key) — derived from the user's master password and a
//     random salt using Argon2id. It wraps (encrypts) the DEK with AES-GCM so
//     that the encrypted DEK can be stored on the server safely.
//
//  3. AuthHash — SHA-256(KEK ‖ authSalt). Sent to the server as the
//     authentication credential. Because SHA-256 is one-way, the server cannot
//     recover the KEK from it.
//
// # Registration flow
//
//  1. [KeyChainService.GenerateEncryptionSalt] + [KeyChainService.GenerateDEK]
//  2. [KeyChainService.GenerateKEK](password, salt)
//  3. [KeyChainService.GetEncryptedDEK](DEK, KEK)  → stored on server
//  4. [KeyChainService.GenerateAuthHash](KEK, authSalt) → used as password hash
//
// # Login flow
//
//  1. Fetch salt from server
//  2. [KeyChainService.GenerateKEK](password, salt)
//  3. [KeyChainService.GenerateAuthHash](KEK, authSalt) → authenticate
//  4. [KeyChainService.DecryptDEK](encryptedDEK, KEK)  → recover DEK
package crypto

//go:generate mockgen -source=interfaces.go -destination=../mock/keychain_service_mock.go -package=mock

// KeyChainService is responsible for all client-side cryptography in the
// zero-knowledge scheme. It has no knowledge of the network, database, or
// user identity — its sole responsibility is to generate and protect keys.
//
// See the package documentation for a description of the key hierarchy and
// the registration / login flows.
type KeyChainService interface {
	// GenerateEncryptionSalt generates a cryptographically random 16-byte
	// (128-bit) salt. The salt is not a secret — it is stored in plaintext on
	// the server — but it ensures that identical master passwords produce
	// different KEKs for different users (or after a password change).
	// Called at step 1 of registration.
	GenerateEncryptionSalt() ([]byte, error)

	// GenerateDEK generates a cryptographically random 32-byte (256-bit)
	// data-encryption key. The DEK encrypts all user vault data and must
	// never leave the client in plaintext. Called at step 1 of registration.
	GenerateDEK() ([]byte, error)

	// GenerateKEK derives a 256-bit key-encryption key from masterPassword
	// and salt using Argon2id. The KEK exists only in client memory and is
	// never transmitted to the server. Called at step 2 of both registration
	// and login.
	GenerateKEK(masterPassword string, salt []byte) []byte

	// GetEncryptedDEK wraps DEK with KEK using AES-256-GCM. The returned
	// blob has the format: nonce (12 bytes) ‖ ciphertext. It is safe to
	// store on the server — without the KEK it is indistinguishable from
	// random bytes. Called at step 3 of registration.
	GetEncryptedDEK(DEK, KEK []byte) ([]byte, error)

	// GenerateAuthHash computes the authentication credential that is sent to
	// the server in place of the raw password. It returns SHA-256(KEK ‖
	// authSalt). The fixed authSalt distinguishes this hash from the KEK
	// itself. Because SHA-256 is one-way, the server cannot recover the KEK
	// from the hash. Called at step 4 of registration and step 3 of login.
	GenerateAuthHash(KEK []byte, authSalt string) []byte

	// DecryptDEK unwraps an encrypted DEK blob using KEK via AES-256-GCM.
	// The blob must be in the format produced by [KeyChainService.GetEncryptedDEK]:
	// nonce (12 bytes) ‖ ciphertext. Returns the original DEK, or an error
	// if KEK is wrong or the ciphertext is corrupted (authentication tag
	// mismatch). Called at step 4 of login.
	DecryptDEK(encryptedDEK, KEK []byte) ([]byte, error)

	// EncryptData serialises data to JSON and encrypts it with DEK using
	// AES-256-GCM. Returns a Base64-encoded blob (nonce ‖ ciphertext) that
	// is safe to store on the server.
	EncryptData(data any, DEK []byte) (string, error)

	// DecryptData decodes the Base64 blob produced by
	// [KeyChainService.EncryptData], decrypts it with DEK, and unmarshals
	// the resulting JSON into target (which must be a pointer, exactly as
	// required by [encoding/json.Unmarshal]). Returns an error if decoding,
	// decryption, or unmarshalling fails.
	DecryptData(encryptedB64 string, DEK []byte, target any) error
}
