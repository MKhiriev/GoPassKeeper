// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Rasul Khiriev

package utils

import "github.com/google/uuid"

// UUIDGenerator creates string UUID values for application identifiers.
//
// The generator is stateless and safe to reuse across goroutines.
// Its [Generate] method prefers UUID version 7 (time-ordered) and falls
// back to a random UUID if v7 generation fails.
type UUIDGenerator struct{}

// NewUUIDGenerator returns a new [UUIDGenerator] instance.
//
// The returned generator has no internal mutable state; creating multiple
// instances is inexpensive.
func NewUUIDGenerator() *UUIDGenerator {
	return &UUIDGenerator{}
}

// Generate returns a UUID string suitable for use as a client-side identifier.
//
// It first attempts to create UUID v7 via [uuid.NewV7]. If that operation
// fails, it falls back to [uuid.NewString] (random UUID) to preserve
// availability and still return a valid UUID-formatted value.
func (g *UUIDGenerator) Generate() string {
	v7, err := uuid.NewV7()
	if err != nil {
		return uuid.NewString()
	}

	return v7.String()
}
