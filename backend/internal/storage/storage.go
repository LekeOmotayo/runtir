package storage

import (
	"context"
	"io"

	"github.com/google/uuid"
)

// StoredFile represents metadata of a successfully saved file
type StoredFile struct {
	RelativePath string
	FullPath     string
	FileSize     int64
	FileHash     string // SHA-256 hex
	MimeType     string
}

// StorageService defines the contract for document file operations
type StorageService interface {
	// Save stores a file stream for a specific organisation and returns stored file details
	Save(ctx context.Context, orgID uuid.UUID, filename string, mimeType string, src io.Reader) (*StoredFile, error)

	// Open retrieves a readable stream for a file by its relative path
	Open(ctx context.Context, relativePath string) (io.ReadCloser, error)

	// Delete removes a file by its relative path
	Delete(ctx context.Context, relativePath string) error
}
