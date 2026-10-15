package storage

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
)

func TestLocalStorageService(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "afriops_storage_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	svc, err := NewLocalStorageService(tempDir)
	if err != nil {
		t.Fatalf("failed to initialize LocalStorageService: %v", err)
	}

	ctx := context.Background()
	orgID := uuid.New()
	testContent := []byte("Hello AfriOps! This is a test document content for ingestion.")
	expectedHash := sha256.Sum256(testContent)
	expectedHashStr := hex.EncodeToString(expectedHash[:])

	// 1. Save file
	stored, err := svc.Save(ctx, orgID, "sample_doc.txt", "text/plain", bytes.NewReader(testContent))
	if err != nil {
		t.Fatalf("expected no error saving file, got: %v", err)
	}

	if stored.FileSize != int64(len(testContent)) {
		t.Errorf("expected file size %d, got %d", len(testContent), stored.FileSize)
	}
	if stored.FileHash != expectedHashStr {
		t.Errorf("expected hash %s, got %s", expectedHashStr, stored.FileHash)
	}
	if stored.MimeType != "text/plain" {
		t.Errorf("expected mime text/plain, got %s", stored.MimeType)
	}

	// 2. Open file
	reader, err := svc.Open(ctx, stored.RelativePath)
	if err != nil {
		t.Fatalf("expected no error opening stored file, got: %v", err)
	}
	readData, err := io.ReadAll(reader)
	_ = reader.Close()
	if err != nil {
		t.Fatalf("failed to read stored file: %v", err)
	}
	if !bytes.Equal(readData, testContent) {
		t.Fatalf("expected read content to match original content")
	}

	// 3. Prevent path traversal
	_, err = svc.Open(ctx, filepath.Join("..", "..", "windows", "system32", "cmd.exe"))
	if err == nil {
		t.Fatalf("expected error on path traversal attempt, got nil")
	}

	// 4. Delete file
	if err := svc.Delete(ctx, stored.RelativePath); err != nil {
		t.Fatalf("expected no error deleting file, got: %v", err)
	}

	// 5. Open after delete should fail
	_, err = svc.Open(ctx, stored.RelativePath)
	if err == nil {
		t.Fatalf("expected error opening deleted file, got nil")
	}
}
