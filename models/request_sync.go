// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Rasul Khiriev

package models

// SyncRequest is sent by the client to initiate a synchronization cycle.
// The client provides the list of all known item identifiers so that
// the server can determine which records were created, updated, or deleted
// since the last sync.
type SyncRequest struct {
	// UserID is the owner of the vault being synchronized.
	UserID int64 `json:"user_id"`

	// ClientSideIDs is the full set of item identifiers currently
	// present in the client's local database.
	// The server compares this list against its own records to detect
	// items that are missing on the client (new or restored)
	// or missing on the server (locally created or orphaned).
	ClientSideIDs []string `json:"client_side_ids"`

	// Length is the total number of entries in ClientSideIDs.
	Length int `json:"length"`
}
