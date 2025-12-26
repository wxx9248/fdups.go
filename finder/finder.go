// Package finder provides functionality for finding duplicate files
// in a directory tree by computing and comparing content hashes.
//
// The package uses a concurrent worker pool to process files in parallel,
// making it efficient for large directory trees. Different finder
// implementations support various file types and hashing strategies:
//   - DefaultFinder: processes all files using SHA-256 hash
//   - FlacFinder: processes only FLAC files, hashing decoded audio content
package finder

// Finder defines the interface for duplicate file detection.
//
// Implementations recursively scan a directory tree, compute content hashes,
// and group files by their hash values.
type Finder interface {
	// Find scans the target directory and returns files grouped by hash.
	// The returned map uses hash strings as keys, with slices of FileInfo
	// for all files sharing that hash. Files appearing alone in a group
	// have no duplicates.
	//
	// Returns an error if directory traversal or file processing fails.
	Find() (error, map[string][]FileInfo)
}
