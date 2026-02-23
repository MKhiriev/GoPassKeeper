package models

// SyncResponse contains the server-side state of every vault item
// that belongs to the user. The client uses this information to
// reconcile its local database: download missing items, push local
// changes, and remove soft-deleted records.
type SyncResponse struct {
	// PrivateDataStates is the list of lightweight state descriptors
	// for each vault item. Each entry carries the hash, version, and
	// deletion flag — enough for the client to decide whether a full
	// fetch or push is needed.
	PrivateDataStates []PrivateDataState `json:"private_data_states"`

	// Length is the total number of entries in PrivateDataStates.
	// Provided for convenience so the client can pre-allocate
	// or validate the response without iterating the slice.
	Length int `json:"length"`
}
