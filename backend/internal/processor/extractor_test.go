package processor

import (
	"context"
	"strings"
	"testing"
)

func TestChunkText(t *testing.T) {
	text := "The quick brown fox jumps over the lazy dog. Document ingestion and processing pipeline works smoothly."

	chunks := ChunkText(text, 5, 2)
	if len(chunks) == 0 {
		t.Fatalf("expected chunks, got 0")
	}

	for i, c := range chunks {
		if strings.TrimSpace(c) == "" {
			t.Errorf("chunk %d is empty", i)
		}
	}
}

func TestEstimateTokenCount(t *testing.T) {
	sample := "Hello world, this is a test"
	tokens := EstimateTokenCount(sample)
	if tokens <= 0 {
		t.Fatalf("expected positive token count, got %d", tokens)
	}
}

func TestExtractText_PlainText(t *testing.T) {
	content := "AfriOps is an enterprise document intelligence platform.\nIt enables AI question answering."
	reader := strings.NewReader(content)

	extracted, pages, err := ExtractText(context.Background(), "text/plain", reader)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(extracted, "AfriOps") {
		t.Errorf("expected extracted text to contain 'AfriOps', got '%s'", extracted)
	}

	if pages < 1 {
		t.Errorf("expected at least 1 page, got %d", pages)
	}
}
