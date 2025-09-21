package models

// File represents metadata for a file.
// Stored in DB and used across upload, download, deduplication.
type File struct {
	ID             int    // unique file ID
	UserID         int    // uploader's user ID
	Filename       string // original filename
	Filepath       string // physical path on disk
	Hash           string // file hash (for deduplication)
	Size           int64  // size in bytes
	MimeType       string // detected MIME type
	ReferenceCount int    // number of users referencing this file
	IsMaster       bool   // true if this is the original (non-duplicate) file
}
