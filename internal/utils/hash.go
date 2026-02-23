package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"hash"
	"sync"
)

// hasherPool is a package-level pool of reusable HMAC-SHA256 hash instances.
// Must be initialized via InitHasherPool before use.
var hasherPool sync.Pool

// Hasher provides keyed HMAC-SHA256 hashing for metrics.
// It stores the hash key as a byte slice to avoid repeated conversions.
type Hasher struct {
	hashKey []byte
}

// InitHasherPool initializes a sync.Pool of HMAC-SHA256 hashers.
// Each hasher in the pool is configured with the provided hash key.
//
// Purpose:
//   - Avoid repeated allocations of new hash.Hash instances
//   - Reduce GC pressure in high-throughput hashing paths
//
// Parameters:
//
//	hashKey - key used for all HMAC operations
//
// Example usage:
//
//	utils.InitHasherPool("my-secret-key")
func InitHasherPool(hashKey string) {
	hasherPool = sync.Pool{
		New: func() any {
			return hmac.New(sha256.New, []byte(hashKey))
		},
	}
}

// Hash computes an HMAC-SHA256 signature over the given byte slice
// using a hasher pulled from the global hasher pool.
//
// Behavior:
//   - Retrieves a hash.Hash instance from sync.Pool
//   - Resets it, writes the data, computes the sum
//   - Resets again and returns it to the pool
//
// Parameters:
//
//	data - arbitrary byte slice to be hashed
//
// Returns:
//
//	[]byte - HMAC-SHA256 digest
//
// Example usage:
//
//	digest := utils.Hash([]byte("some data"))
func Hash(data []byte) []byte {
	h := hasherPool.Get().(hash.Hash)
	h.Reset()

	h.Write(data)
	sum := h.Sum(nil)

	h.Reset()
	hasherPool.Put(h)

	return sum
}

// HashString computes an HMAC-SHA256 signature over the given string
// using the provided hash key and returns the result as a hex-encoded string.
//
// Unlike Hash, this function does not use the global hasher pool and
// creates a new HMAC instance on each call. Suitable for one-off hashing
// where pool initialization is not desired.
//
// Parameters:
//
//	data    - string to be hashed
//	hashKey - secret key used for the HMAC operation
//
// Returns:
//
//	string - hex-encoded HMAC-SHA256 digest
//
// Example usage:
//
//	signature := utils.HashString("some data", "my-secret-key")
func HashString(data string, hashKey string) string {
	return hex.EncodeToString(hashString([]byte(data), hashKey))
}

// hashString computes an HMAC-SHA256 digest over the given byte slice
// using the provided hash key.
//
// This is an internal helper used by HashString.
// A new HMAC instance is created on each call.
//
// Parameters:
//
//	data    - byte slice to be hashed
//	hashKey - secret key used for the HMAC operation
//
// Returns:
//
//	[]byte - raw HMAC-SHA256 digest
func hashString(data []byte, hashKey string) []byte {
	hasher := hmac.New(sha256.New, []byte(hashKey))
	hasher.Write(data)
	return hasher.Sum(nil)
}
