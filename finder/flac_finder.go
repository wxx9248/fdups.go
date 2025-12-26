package finder

import (
	"os"
	"path/filepath"
	"strings"

	"fdups/hasher"
)

// flacFinder finds duplicate FLAC audio files by comparing decoded audio content.
// It only processes files with the .flac extension and computes hashes based on
// the raw PCM samples, ignoring metadata differences.
type flacFinder struct {
	*baseFinder
}

// NewFlacFinder creates a Finder that processes only FLAC files in the
// target directory.
//
// Unlike NewDefaultFinder, this finder hashes the decoded audio samples
// rather than raw file bytes. This means two FLAC files with identical
// audio but different metadata or encoding parameters will be detected
// as duplicates.
func NewFlacFinder(targetDirectory string) Finder {
	return &flacFinder{
		baseFinder: newBaseFinder(
			targetDirectory,
			hasher.NewFlacHasher(),
			acceptFlacFiles,
		),
	}
}

// acceptFlacFiles is a FileFilter that accepts only .flac files (case-insensitive).
func acceptFlacFiles(path string, _ os.FileInfo) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".flac"
}
