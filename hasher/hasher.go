// Package hasher provides interfaces and implementations for computing
// content hashes from data streams.
//
// The package defines a Hasher interface that abstracts the hashing algorithm,
// allowing different implementations for various file types. Built-in implementations
// include:
//   - DefaultHasher: computes SHA-256 hash of raw file content
//   - FlacHasher: computes SHA-256 hash of decoded FLAC audio samples
package hasher

import (
	"io"
)

// Hasher defines the interface for computing hashes from data streams.
//
// Implementations should be stateless and safe for concurrent use.
// The caller is responsible for managing the lifecycle of the provided reader.
type Hasher interface {
	// Hash reads all data from r and returns the computed hash bytes.
	// Returns an error if reading fails or the data format is invalid.
	Hash(r io.Reader) ([]byte, error)
}
