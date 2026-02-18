package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"hash"
	"sync"
)

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
func Hash(data []byte) []byte {
	h := hasherPool.Get().(hash.Hash)
	h.Reset()

	h.Write(data)
	sum := h.Sum(nil)

	h.Reset()
	hasherPool.Put(h)

	return sum
}
