package finder

// FileInfo holds metadata about a file including its computed content hash.
//
// The struct is JSON-serializable for output formatting.
type FileInfo struct {
	// Name is the base name of the file.
	Name string `json:"name"`
	// Path is the absolute path to the file.
	Path string `json:"path"`
	// Size is the file size in bytes.
	Size int64 `json:"size"`
	// Hash is the hexadecimal-encoded content hash.
	Hash string `json:"hash"`
}
