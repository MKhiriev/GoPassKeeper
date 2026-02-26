package service

import (
	"context"
	"time"

	"github.com/MKhiriev/go-pass-keeper/models"
)

// ClientCryptoService defines the client-side contract for encrypting and decrypting
// vault payloads using the data-encryption key (DEK) derived from the user's master
// password. The key must be set via SetEncryptionKey before calling any other method.
type ClientCryptoService interface {
	// SetEncryptionKey stores the DEK that will be used for all subsequent
	// Encrypt/Decrypt operations. It is called once after a successful login.
	SetEncryptionKey(key []byte)

	// EncryptPayload encrypts a plaintext vault payload and returns the
	// ciphered representation ready for local storage or server upload.
	// Returns an error if encryption of any field fails.
	EncryptPayload(plain models.DecipheredPayload) (models.PrivateDataPayload, error)

	// DecryptPayload decrypts a ciphered vault payload retrieved from local
	// storage or the server and returns the plaintext representation.
	// Returns an error if decryption of any field fails.
	DecryptPayload(cipher models.PrivateDataPayload) (models.DecipheredPayload, error)

	// ComputeHash computes a deterministic hash of the given payload value
	// (typically a models.PrivateDataPayload) for use in sync conflict detection.
	// Returns the hash as a hex/base64 string or an error if serialisation fails.
	ComputeHash(payload any) (string, error)
}

// ClientAuthService defines the client-side contract for user registration and
// authentication. Implementations are responsible for key derivation and for
// communicating with the remote server adapter.
type ClientAuthService interface {
	// Register creates a new account on the server for the given user.
	// It derives a key-encryption key (KEK) from the master password, generates
	// a data-encryption key (DEK), encrypts the DEK with the KEK, and persists
	// the resulting credential bundle on the server.
	// Returns an error if key generation, encryption, or the server call fails.
	Register(ctx context.Context, user models.User) error

	// Login authenticates the user against the server.
	// It fetches the user's encryption salt, derives the KEK, computes the auth
	// hash, retrieves the encrypted DEK from the server, decrypts it, and
	// initialises the crypto service with the plaintext DEK.
	// Returns the server-assigned user ID and the plaintext DEK, or an error if
	// any step fails.
	Login(ctx context.Context, user models.User) (userID int64, encryptionKey []byte, err error)
}

// ClientPrivateDataService defines the client-side contract for managing vault items.
// All CRUD operations work against the local database and propagate changes to the
// server in the same call.
type ClientPrivateDataService interface {
	// SetEncryptionKey forwards the DEK to the underlying ClientCryptoService.
	// Must be called before any Create/Get/Update/Delete operation.
	SetEncryptionKey(key []byte)

	// Create encrypts plain, assigns a new client-side UUID, saves the item to
	// the local store, and uploads it to the server.
	// Returns an error if encryption, local save, or server upload fails.
	Create(ctx context.Context, userID int64, plain models.DecipheredPayload) error

	// GetAll loads every non-deleted vault item for userID from the local store,
	// decrypts each one, and returns the plaintext collection.
	// Returns an error if the local query or any decryption fails.
	GetAll(ctx context.Context, userID int64) ([]models.DecipheredPayload, error)

	// Get loads the single vault item identified by clientSideID from the local
	// store, decrypts it, and returns the plaintext payload.
	// Returns an error if the item is not found or decryption fails.
	Get(ctx context.Context, clientSideID string, userID int64) (models.DecipheredPayload, error)

	// Update encrypts the modified data payload, persists it to the local store,
	// and pushes the update to the server. On server success the local version
	// counter is incremented.
	// Returns an error if encryption, local update, or server update fails.
	Update(ctx context.Context, data models.DecipheredPayload) error

	// Delete soft-deletes the vault item in the local store and sends a delete
	// request to the server. On server success the local version counter is
	// incremented.
	// Returns an error if the local delete or server delete fails.
	Delete(ctx context.Context, clientSideID string, userID int64) error
}

// ClientSyncService defines the client-side contract for synchronising the local
// vault with the remote server.
type ClientSyncService interface {
	// FullSync performs a complete bidirectional synchronisation for the given
	// user: it fetches server and client state descriptors, builds a sync plan,
	// and executes all required download, upload, update, and delete operations.
	// Returns an error if any step of the sync fails.
	FullSync(ctx context.Context, userID int64) error

	// ExecutePlan carries out the actions described in plan for the given user.
	// Each action category (Download, Upload, Update, DeleteClient, DeleteServer)
	// is executed in order. Returns the first error encountered, if any.
	ExecutePlan(ctx context.Context, plan models.SyncPlan, userID int64) error
}

// ClientSyncJob defines the contract for a background sync worker that
// periodically calls FullSync for the authenticated user.
type ClientSyncJob interface {
	// Start launches the background sync goroutine. It syncs every interval,
	// defaulting to 5 minutes if interval is zero or negative. Any previously
	// running job is stopped before the new one begins.
	Start(ctx context.Context, userID int64, interval time.Duration)

	// Stop signals the background goroutine to exit and blocks until it has
	// fully terminated.
	Stop()
}
