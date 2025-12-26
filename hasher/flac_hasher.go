package hasher

import (
	"crypto/sha256"
	"hash"
	"io"

	"github.com/mewkiz/flac"
)

type flacHasher struct{}

// NewFlacHasher returns a Hasher that computes SHA-256 hashes of decoded
// FLAC audio samples.
//
// Unlike NewDefaultHasher which hashes raw file bytes, this hasher decodes
// the FLAC stream and hashes only the PCM audio samples. This means two
// FLAC files with identical audio but different metadata or encoding
// parameters will produce the same hash.
//
// The hasher expects the reader to contain valid FLAC data. Returns an
// error if the input is not valid FLAC format.
func NewFlacHasher() Hasher {
	return &flacHasher{}
}

func (h *flacHasher) Hash(r io.Reader) ([]byte, error) {
	stream, err := flac.New(r)
	if err != nil {
		return nil, err
	}
	defer stream.Close()

	hash := sha256.New()
	if err := h.hashAudioFrames(stream, hash); err != nil {
		return nil, err
	}
	return hash.Sum(nil), nil
}

func (h *flacHasher) hashAudioFrames(stream *flac.Stream, hash hash.Hash) error {
	for {
		frame, err := stream.ParseNext()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		frame.Hash(hash)
	}
}
