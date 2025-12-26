package hasher

import (
	"crypto/sha256"
	"io"
)

// defaultHasher computes SHA-256 hashes of raw file content.
type defaultHasher struct{}

// NewDefaultHasher returns a Hasher that computes SHA-256 hashes.
//
// The returned hasher streams data through the hash function without
// loading the entire content into memory, making it suitable for large files.
func NewDefaultHasher() Hasher {
	return &defaultHasher{}
}

func (h *defaultHasher) Hash(r io.Reader) ([]byte, error) {
	hash := sha256.New()
	if _, err := io.Copy(hash, r); err != nil {
		return nil, err
	}
	return hash.Sum(nil), nil
}
