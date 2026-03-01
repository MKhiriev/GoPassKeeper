// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Rasul Khiriev

// Package validators provides abstractions for input validation and
// enforcement of business rules across the application.
//
// Core concepts:
//   - Validator: generic interface to validate arbitrary values or structures.
//     Supports optional field-level scoping for targeted validation.
//
// Usage patterns:
//  1. Implement Validator to encode domain-specific validation logic.
//  2. Inject Validator implementations into services or handlers.
//  3. Call Validate with context, value, and optional field names to enforce rules.
//
// This package decouples validation logic from transport layers and storage,
// enabling reusable, composable, and testable validation strategies.
package validators

import "context"

// Validator defines a generic validation interface for arbitrary input values.
// Implementations may perform structural validation, semantic checks,
// cross-field rules.
type Validator interface {

	// Validate validates the provided input and optionally
	// restricts validation to specific named fields.
	Validate(context.Context, any, ...string) error
}
