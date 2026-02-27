// Package utils provides general-purpose helper utilities
// used across different parts of the application.
// Includes tools for working with context, type-safe keys, hashing,
// HTTP response writing, HTTP client initialization, JWT token generation
// and validation, and other common operations.
package utils

import (
	"context"
)

// contextKey is a private type for context keys.
// Using a dedicated type instead of a plain string prevents key collisions
// with other packages that may use string-based keys in the context.
type contextKey string

// String returns the string representation of the context key.
// Implements the fmt.Stringer interface.
func (c contextKey) String() string {
	return string(c)
}

// UserIDCtxKey is the key used to store the user identifier in the context.
// Used together with GetUserIDFromContext for type-safe retrieval
// of the user ID from context.Context.
//
// Example of writing a value to the context:
//
//	ctx := context.WithValue(ctx, utils.UserIDCtxKey, int64(42))
var UserIDCtxKey = contextKey("userID")

// GetUserIDFromContext retrieves the user identifier from the context.
//
// Returns the user ID of type int64 and an ok flag:
//   - ok == true  — value is found and has the correct int64 type
//   - ok == false — value is missing or has an unexpected type
//
// Example usage:
//
//	userID, ok := utils.GetUserIDFromContext(ctx)
//	if !ok {
//	    // handle missing user in context
//	}
func GetUserIDFromContext(ctx context.Context) (int64, bool) {
	userID, ok := ctx.Value(UserIDCtxKey).(int64)
	return userID, ok
}
