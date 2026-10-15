package storage

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

var (
	ErrInvalidPath = errors.New("invalid or unsafe file path")
	ErrFileNotFound = errors.New("file not found")
)

type LocalStorageService struct {
	baseDir string
}

func NewLocalStorageService(baseDir string) (*LocalStorageService, error) {
	absBase, err := filepath.Abs(baseDir)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve absolute path for base directory: %w", err)
	}

	if err := os.MkdirAll(absBase, 0755); err != nil {
		return nil, fmt.Errorf("failed to create upload directory %s: %w", absBase, err)
	}

	return &LocalStorageService{
		baseDir: absBase,
	}, nil
}

func (s *LocalStorageService) Save(ctx context.Context, orgID uuid.UUID, filename string, mimeType string, src io.Reader) (*StoredFile, error) {
	// 1. Create org-specific directory
	orgDir := filepath.Join(s.baseDir, orgID.String())
	if err := os.MkdirAll(orgDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create organization directory: %w", err)
	}

	// 2. Sanitize filename and create unique stored name
	cleanName := filepath.Base(filename)
	cleanName = strings.ReplaceAll(cleanName, " ", "_")
	uniquePrefix := uuid.New().String()
	storedFilename := fmt.Sprintf("%s_%s", uniquePrefix, cleanName)

	fullPath := filepath.Join(orgDir, storedFilename)
	relPath := filepath.Join(orgID.String(), storedFilename)

	dst, err := os.Create(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dst.Close()

	// 3. Stream data while calculating SHA-256 hash
	hasher := sha256.New()
	multiWriter := io.MultiWriter(dst, hasher)

	written, err := io.Copy(multiWriter, src)
	if err != nil {
		_ = os.Remove(fullPath)
		return nil, fmt.Errorf("failed to write file content: %w", err)
	}

	hashStr := hex.EncodeToString(hasher.Sum(nil))

	return &StoredFile{
		RelativePath: filepath.ToSlash(relPath),
		FullPath:     fullPath,
		FileSize:     written,
		FileHash:     hashStr,
		MimeType:     mimeType,
	}, nil
}

func (s *LocalStorageService) Open(ctx context.Context, relativePath string) (io.ReadCloser, error) {
	fullPath, err := s.resolveSafePath(relativePath)
	if err != nil {
		return nil, err
	}

	file, err := os.Open(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrFileNotFound
		}
		return nil, err
	}

	return file, nil
}

func (s *LocalStorageService) Delete(ctx context.Context, relativePath string) error {
	fullPath, err := s.resolveSafePath(relativePath)
	if err != nil {
		return err
	}

	if err := os.Remove(fullPath); err != nil && !os.IsNotExist(err) {
		return err
	}

	return nil
}

func (s *LocalStorageService) resolveSafePath(relativePath string) (string, error) {
	cleaned := filepath.Clean(relativePath)
	fullPath := filepath.Join(s.baseDir, cleaned)

	// Guard against path traversal attacks
	if !strings.HasPrefix(fullPath, s.baseDir) {
		return "", ErrInvalidPath
	}

	return fullPath, nil
}
