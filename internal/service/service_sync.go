package service

import (
	"context"

	"github.com/MKhiriev/go-pass-keeper/models"
)

// syncService is the concrete implementation of SyncService.
// It performs a purely in-memory comparison of server and client
// PrivateDataState slices; no storage layer or logger is required
// because the operation is stateless and produces no side effects.
type syncService struct{}

// NewSyncService constructs a SyncService ready for use.
// Because BuildSyncPlan is a stateless, in-memory operation,
// no dependencies (storage, config, logger) are needed.
func NewSyncService() SyncService {
	return &syncService{}
}

// BuildSyncPlan implements SyncService.
//
// It builds two O(1) lookup indexes from the input slices, then makes
// two linear passes to classify every item into exactly one action category:
//
//   - Pass 1 (over serverData): handles items present on the server,
//     whether or not they also exist on the client.
//   - Pass 2 (over clientData): catches items that exist only on the
//     client and were therefore invisible in pass 1.
//
// ctx cancellation is checked at the start of each iteration so that
// callers can abort early when operating on large datasets.
func (s *syncService) BuildSyncPlan(
	ctx context.Context,
	serverData, clientData []models.PrivateDataState,
) (models.SyncPlan, error) {
	var plan models.SyncPlan

	// Build O(1) lookup indexes keyed by ClientSideID.
	clientIndex := make(map[string]models.PrivateDataState, len(clientData))
	for _, cd := range clientData {
		clientIndex[cd.ClientSideID] = cd
	}

	serverIndex := make(map[string]models.PrivateDataState, len(serverData))
	for _, sd := range serverData {
		serverIndex[sd.ClientSideID] = sd
	}

	// ── Pass 1: iterate over server records ─────────────────────────────────
	for _, sd := range serverData {
		if err := ctx.Err(); err != nil {
			return models.SyncPlan{}, err
		}

		cd, existsOnClient := clientIndex[sd.ClientSideID]

		if !existsOnClient {
			if !sd.Deleted {
				// Server has a live record the client has never seen → download.
				plan.Download = append(plan.Download, sd)
			}
			// sd.Deleted && !existsOnClient: the record was created and
			// deleted on the server before the client ever synced it — no action.
			continue
		}

		// Record exists on both sides: classify by version, then by state.
		switch {
		case sd.Version == cd.Version:
			// Same version — decide based on deletion flags and hash.
			switch {
			case sd.Deleted && cd.Deleted:
				// Both sides agree it is deleted → already in sync, no action.

			case sd.Deleted && !cd.Deleted:
				// Server soft-deleted the record; client still has a live copy
				// → remove from client.
				plan.DeleteClient = append(plan.DeleteClient, sd)

			case !sd.Deleted && cd.Deleted:
				// Client soft-deleted the record; server still has a live copy
				// → remove from server.
				plan.DeleteServer = append(plan.DeleteServer, cd)

			case sd.Hash == cd.Hash:
				// Version and hash match → records are identical, no action.

			default: // !sd.Deleted && !cd.Deleted && sd.Hash != cd.Hash
				// Same version but diverged hash: the client edited the record
				// locally without bumping the version (e.g. offline edit)
				// → push client's version to the server.
				plan.Update = append(plan.Update, cd)
			}

		case sd.Version > cd.Version:
			// Server is ahead of the client.
			if sd.Deleted {
				// Server's newer version is a soft-delete → remove from client.
				plan.DeleteClient = append(plan.DeleteClient, sd)
			} else {
				// Server has a newer live version → download to client.
				plan.Download = append(plan.Download, sd)
			}

		default: // sd.Version < cd.Version
			// Client is ahead of the server (offline edits accumulated locally).
			if cd.Deleted {
				// Client's newer version is a soft-delete → remove from server.
				plan.DeleteServer = append(plan.DeleteServer, cd)
			} else {
				// Client has a newer live version → push to server.
				plan.Update = append(plan.Update, cd)
			}
		}
	}

	// ── Pass 2: find client-only records (absent from server) ────────────────
	for _, cd := range clientData {
		if err := ctx.Err(); err != nil {
			return models.SyncPlan{}, err
		}

		if _, existsOnServer := serverIndex[cd.ClientSideID]; existsOnServer {
			// Already handled in pass 1.
			continue
		}

		if !cd.Deleted {
			// Live record that has never been pushed to the server → upload.
			plan.Upload = append(plan.Upload, cd)
		}
		// cd.Deleted && !existsOnServer: record was created and soft-deleted
		// locally before the first sync — server never knew about it, no action.
	}

	return plan, nil
}
