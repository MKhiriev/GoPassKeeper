// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Rasul Khiriev

package service

import (
	"context"
	"testing"

	"github.com/MKhiriev/go-pass-keeper/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ─────────────────────────────────────────────────────────────────────────────
// Helper
// ─────────────────────────────────────────────────────────────────────────────

// st is a shorthand constructor for PrivateDataState used only in tests.
func st(id string, version int64, hash string, deleted bool) models.PrivateDataState {
	return models.PrivateDataState{
		ClientSideID: id,
		Version:      version,
		Hash:         hash,
		Deleted:      deleted,
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// BuildSyncPlan — decision matrix (table-driven)
// ─────────────────────────────────────────────────────────────────────────────

// TestSyncService_BuildSyncPlan_DecisionMatrix covers every cell of the
// classification table for a single record.  Each sub-test is named after the
// condition it exercises so failures are immediately self-documenting.
func TestSyncService_BuildSyncPlan_DecisionMatrix(t *testing.T) {
	const (
		id   = "record-1"
		hash = "abc123"
		newH = "xyz789" // a different hash, simulating a local edit
	)

	tests := []struct {
		name       string
		serverData []models.PrivateDataState
		clientData []models.PrivateDataState
		wantPlan   models.SyncPlan
	}{
		// ── Records present only on the server ───────────────────────────────

		{
			name:       "ServerOnly/Alive → Download",
			serverData: []models.PrivateDataState{st(id, 1, hash, false)},
			clientData: nil,
			wantPlan:   models.SyncPlan{Download: []models.PrivateDataState{st(id, 1, hash, false)}},
		},
		{
			name:       "ServerOnly/Deleted → NoAction",
			serverData: []models.PrivateDataState{st(id, 1, hash, true)},
			clientData: nil,
			wantPlan:   models.SyncPlan{},
		},

		// ── Records present only on the client ───────────────────────────────

		{
			name:       "ClientOnly/Alive → Upload",
			serverData: nil,
			clientData: []models.PrivateDataState{st(id, 1, hash, false)},
			wantPlan:   models.SyncPlan{Upload: []models.PrivateDataState{st(id, 1, hash, false)}},
		},
		{
			name:       "ClientOnly/Deleted → NoAction",
			serverData: nil,
			clientData: []models.PrivateDataState{st(id, 1, hash, true)},
			wantPlan:   models.SyncPlan{},
		},

		// ── Same version ─────────────────────────────────────────────────────

		{
			name:       "SameVersion/BothDeleted → NoAction",
			serverData: []models.PrivateDataState{st(id, 2, hash, true)},
			clientData: []models.PrivateDataState{st(id, 2, hash, true)},
			wantPlan:   models.SyncPlan{},
		},
		{
			name:       "SameVersion/ServerDeleted/ClientAlive → DeleteClient",
			serverData: []models.PrivateDataState{st(id, 2, hash, true)},
			clientData: []models.PrivateDataState{st(id, 2, hash, false)},
			wantPlan:   models.SyncPlan{DeleteClient: []models.PrivateDataState{st(id, 2, hash, true)}},
		},
		{
			name:       "SameVersion/ServerAlive/ClientDeleted → DeleteServer",
			serverData: []models.PrivateDataState{st(id, 2, hash, false)},
			clientData: []models.PrivateDataState{st(id, 2, hash, true)},
			wantPlan:   models.SyncPlan{DeleteServer: []models.PrivateDataState{st(id, 2, hash, true)}},
		},
		{
			name:       "SameVersion/SameHash/BothAlive → NoAction",
			serverData: []models.PrivateDataState{st(id, 3, hash, false)},
			clientData: []models.PrivateDataState{st(id, 3, hash, false)},
			wantPlan:   models.SyncPlan{},
		},
		{
			name:       "SameVersion/DiffHash/BothAlive → Update",
			serverData: []models.PrivateDataState{st(id, 3, hash, false)},
			clientData: []models.PrivateDataState{st(id, 3, newH, false)},
			wantPlan:   models.SyncPlan{Update: []models.PrivateDataState{st(id, 3, newH, false)}},
		},

		// ── Server version is newer (server ahead) ───────────────────────────

		{
			name:       "ServerNewer/ServerDeleted/ClientAlive → DeleteClient",
			serverData: []models.PrivateDataState{st(id, 5, hash, true)},
			clientData: []models.PrivateDataState{st(id, 3, hash, false)},
			wantPlan:   models.SyncPlan{DeleteClient: []models.PrivateDataState{st(id, 5, hash, true)}},
		},
		{
			name:       "ServerNewer/ServerAlive/ClientAlive → Download",
			serverData: []models.PrivateDataState{st(id, 5, newH, false)},
			clientData: []models.PrivateDataState{st(id, 3, hash, false)},
			wantPlan:   models.SyncPlan{Download: []models.PrivateDataState{st(id, 5, newH, false)}},
		},
		{
			// Server has a live newer version; client had already soft-deleted
			// the stale copy.  Server's more recent state wins → Download.
			name:       "ServerNewer/ServerAlive/ClientDeleted → Download",
			serverData: []models.PrivateDataState{st(id, 5, newH, false)},
			clientData: []models.PrivateDataState{st(id, 3, hash, true)},
			wantPlan:   models.SyncPlan{Download: []models.PrivateDataState{st(id, 5, newH, false)}},
		},

		// ── Client version is newer (offline edits) ──────────────────────────

		{
			name:       "ClientNewer/ClientDeleted/ServerAlive → DeleteServer",
			serverData: []models.PrivateDataState{st(id, 3, hash, false)},
			clientData: []models.PrivateDataState{st(id, 5, hash, true)},
			wantPlan:   models.SyncPlan{DeleteServer: []models.PrivateDataState{st(id, 5, hash, true)}},
		},
		{
			name:       "ClientNewer/ClientAlive/ServerAlive → Update",
			serverData: []models.PrivateDataState{st(id, 3, hash, false)},
			clientData: []models.PrivateDataState{st(id, 5, newH, false)},
			wantPlan:   models.SyncPlan{Update: []models.PrivateDataState{st(id, 5, newH, false)}},
		},
		{
			// Client has a live newer version; the server had soft-deleted the
			// stale copy.  Client's more recent state wins → Update.
			name:       "ClientNewer/ClientAlive/ServerDeleted → Update",
			serverData: []models.PrivateDataState{st(id, 3, hash, true)},
			clientData: []models.PrivateDataState{st(id, 5, newH, false)},
			wantPlan:   models.SyncPlan{Update: []models.PrivateDataState{st(id, 5, newH, false)}},
		},
	}

	svc := NewSyncService()

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			plan, err := svc.BuildSyncPlan(context.Background(), tc.serverData, tc.clientData)

			require.NoError(t, err)
			assert.Equal(t, tc.wantPlan.Download, plan.Download, "Download mismatch")
			assert.Equal(t, tc.wantPlan.Upload, plan.Upload, "Upload mismatch")
			assert.Equal(t, tc.wantPlan.Update, plan.Update, "Update mismatch")
			assert.Equal(t, tc.wantPlan.DeleteClient, plan.DeleteClient, "DeleteClient mismatch")
			assert.Equal(t, tc.wantPlan.DeleteServer, plan.DeleteServer, "DeleteServer mismatch")
		})
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// BuildSyncPlan — edge cases
// ─────────────────────────────────────────────────────────────────────────────

func TestSyncService_BuildSyncPlan_BothEmpty(t *testing.T) {
	svc := NewSyncService()

	plan, err := svc.BuildSyncPlan(context.Background(), nil, nil)

	require.NoError(t, err)
	assert.Nil(t, plan.Download)
	assert.Nil(t, plan.Upload)
	assert.Nil(t, plan.Update)
	assert.Nil(t, plan.DeleteClient)
	assert.Nil(t, plan.DeleteServer)
}

func TestSyncService_BuildSyncPlan_ContextCancelled(t *testing.T) {
	// A large enough slice ensures the cancellation check fires before
	// the loop finishes naturally.
	const n = 1000
	serverData := make([]models.PrivateDataState, n)
	for i := range serverData {
		serverData[i] = st("id", 1, "h", false)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	svc := NewSyncService()
	_, err := svc.BuildSyncPlan(ctx, serverData, nil)

	require.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
}

// ─────────────────────────────────────────────────────────────────────────────
// BuildSyncPlan — realistic mixed scenario
// ─────────────────────────────────────────────────────────────────────────────

// TestSyncService_BuildSyncPlan_MixedScenario simulates a realistic sync
// session in which different records each fall into a different action category.
//
// Vault state:
//
//	"pass-1"  server v3 alive,  client v3 same hash  → NoAction      (in sync)
//	"pass-2"  server v4 alive,  client v2 alive       → Download      (server newer)
//	"pass-3"  server v2 alive,  client v5 alive       → Update        (offline edit)
//	"pass-4"  server v3 deleted,client v1 alive       → DeleteClient  (server soft-deleted)
//	"pass-5"  server v1 alive,  client v4 deleted     → DeleteServer  (client soft-deleted)
//	"pass-6"  server only,      alive                 → Download      (new on server)
//	"pass-7"  server only,      deleted               → NoAction      (deleted before sync)
//	"pass-8"  client only,      alive                 → Upload        (new on client)
//	"pass-9"  client only,      deleted               → NoAction      (local create+delete)
//	"pass-10" server v2 alive,  client v2 diff hash   → Update        (local edit, same ver)
func TestSyncService_BuildSyncPlan_MixedScenario(t *testing.T) {
	serverData := []models.PrivateDataState{
		st("pass-1", 3, "h1", false),
		st("pass-2", 4, "h2new", false),
		st("pass-3", 2, "h3", false),
		st("pass-4", 3, "h4", true),
		st("pass-5", 1, "h5", false),
		st("pass-6", 1, "h6", false),
		st("pass-7", 2, "h7", true),
		st("pass-10", 2, "h10", false),
	}
	clientData := []models.PrivateDataState{
		st("pass-1", 3, "h1", false),
		st("pass-2", 2, "h2old", false),
		st("pass-3", 5, "h3new", false),
		st("pass-4", 1, "h4", false),
		st("pass-5", 4, "h5", true),
		st("pass-8", 1, "h8", false),
		st("pass-9", 1, "h9", true),
		st("pass-10", 2, "h10edited", false),
	}

	svc := NewSyncService()
	plan, err := svc.BuildSyncPlan(context.Background(), serverData, clientData)

	require.NoError(t, err)

	// Download: pass-2 (server newer) + pass-6 (server only, alive)
	assert.ElementsMatch(t, []models.PrivateDataState{
		st("pass-2", 4, "h2new", false),
		st("pass-6", 1, "h6", false),
	}, plan.Download, "Download")

	// Upload: pass-8 (client only, alive)
	assert.ElementsMatch(t, []models.PrivateDataState{
		st("pass-8", 1, "h8", false),
	}, plan.Upload, "Upload")

	// Update: pass-3 (client version ahead) + pass-10 (same version, hash diverged)
	assert.ElementsMatch(t, []models.PrivateDataState{
		st("pass-3", 5, "h3new", false),
		st("pass-10", 2, "h10edited", false),
	}, plan.Update, "Update")

	// DeleteClient: pass-4 (server soft-deleted a newer version)
	assert.ElementsMatch(t, []models.PrivateDataState{
		st("pass-4", 3, "h4", true),
	}, plan.DeleteClient, "DeleteClient")

	// DeleteServer: pass-5 (client soft-deleted a newer version)
	assert.ElementsMatch(t, []models.PrivateDataState{
		st("pass-5", 4, "h5", true),
	}, plan.DeleteServer, "DeleteServer")
}
