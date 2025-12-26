package finder

import (
	"os"

	"fdups/hasher"
)

// defaultFinder finds duplicate files by computing SHA-256 hashes
// of raw file content. It processes all files regardless of type.
type defaultFinder struct {
	*baseFinder
}

// NewDefaultFinder creates a Finder that processes all files in the
// target directory using SHA-256 hashing.
//
// This is the general-purpose finder suitable for any file type.
func NewDefaultFinder(targetDirectory string) Finder {
	return &defaultFinder{
		baseFinder: newBaseFinder(
			targetDirectory,
			hasher.NewDefaultHasher(),
			acceptAllFiles,
		),
	}
}

// acceptAllFiles is a FileFilter that accepts all files.
func acceptAllFiles(string, os.FileInfo) bool {
	return true
}
