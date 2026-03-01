// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Rasul Khiriev

package utils

import (
	"context"
	"testing"
)

func TestContextKeyString(t *testing.T) {
	key := contextKey("testKey")
	if key.String() != "testKey" {
		t.Errorf("expected 'testKey', got '%s'", key.String())
	}
}

func TestUserIDCtxKey(t *testing.T) {
	if UserIDCtxKey.String() != "userID" {
		t.Errorf("expected 'userID', got '%s'", UserIDCtxKey.String())
	}
}

func TestGetUserIDFromContext_Success(t *testing.T) {
	ctx := context.WithValue(context.Background(), UserIDCtxKey, int64(42))

	userID, ok := GetUserIDFromContext(ctx)

	if !ok {
		t.Fatal("expected ok=true, got false")
	}
	if userID != 42 {
		t.Errorf("expected userID=42, got %d", userID)
	}
}

func TestGetUserIDFromContext_Missing(t *testing.T) {
	ctx := context.Background()

	userID, ok := GetUserIDFromContext(ctx)

	if ok {
		t.Fatal("expected ok=false, got true")
	}
	if userID != 0 {
		t.Errorf("expected userID=0, got %d", userID)
	}
}

func TestGetUserIDFromContext_WrongType(t *testing.T) {
	ctx := context.WithValue(context.Background(), UserIDCtxKey, "not-an-int64")

	userID, ok := GetUserIDFromContext(ctx)

	if ok {
		t.Fatal("expected ok=false for wrong type, got true")
	}
	if userID != 0 {
		t.Errorf("expected userID=0, got %d", userID)
	}
}

func TestGetUserIDFromContext_ZeroValue(t *testing.T) {
	ctx := context.WithValue(context.Background(), UserIDCtxKey, int64(0))

	userID, ok := GetUserIDFromContext(ctx)

	if !ok {
		t.Fatal("expected ok=true for zero value, got false")
	}
	if userID != 0 {
		t.Errorf("expected userID=0, got %d", userID)
	}
}

func TestGetUserIDFromContext_NegativeValue(t *testing.T) {
	ctx := context.WithValue(context.Background(), UserIDCtxKey, int64(-1))

	userID, ok := GetUserIDFromContext(ctx)

	if !ok {
		t.Fatal("expected ok=true, got false")
	}
	if userID != -1 {
		t.Errorf("expected userID=-1, got %d", userID)
	}
}

func TestGetUserIDFromContext_DifferentKey(t *testing.T) {
	otherKey := contextKey("otherKey")
	ctx := context.WithValue(context.Background(), otherKey, int64(99))

	userID, ok := GetUserIDFromContext(ctx)

	if ok {
		t.Fatal("expected ok=false for different key, got true")
	}
	if userID != 0 {
		t.Errorf("expected userID=0, got %d", userID)
	}
}
