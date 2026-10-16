package document

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/lekeOmotayo/runtir/internal/db"
)

func TestCreateHighlightedSnippet(t *testing.T) {
	content := "AfriOps enables companies across Africa to upload, process, search and analyze financial reports with ease."

	tests := []struct {
		name     string
		content  string
		query    string
		maxLen   int
		contains string
	}{
		{
			name:     "Finds exact keyword",
			content:  content,
			query:    "financial reports",
			maxLen:   50,
			contains: "financial reports",
		},
		{
			name:     "Case-insensitive matching",
			content:  content,
			query:    "AFRICA",
			maxLen:   40,
			contains: "Africa",
		},
		{
			name:     "Term not found fallback",
			content:  content,
			query:    "nonexistent",
			maxLen:   30,
			contains: "AfriOps",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			snippet := createHighlightedSnippet(tt.content, tt.query, tt.maxLen)
			if !strings.Contains(strings.ToLower(snippet), strings.ToLower(tt.contains)) {
				t.Errorf("expected snippet to contain '%s', got '%s'", tt.contains, snippet)
			}
		})
	}
}

func TestToDocumentDTO(t *testing.T) {
	docID := uuid.New()
	orgID := uuid.New()
	userID := uuid.New()
	now := time.Now()

	doc := db.Document{
		ID:             docID,
		OrganisationID: orgID,
		UploadedBy:     pgtype.UUID{Bytes: userID, Valid: true},
		Title:          "Annual Report",
		FileName:       "report.pdf",
		FilePath:       "org/report.pdf",
		FileSize:       1024,
		MimeType:       "application/pdf",
		FileHash:       "hash123",
		Status:         "completed",
		IsArchived:     true,
		ArchivedAt:     pgtype.Timestamptz{Time: now, Valid: true},
		CreatedAt:      pgtype.Timestamptz{Time: now, Valid: true},
		UpdatedAt:      pgtype.Timestamptz{Time: now, Valid: true},
	}

	dto := toDocumentDTO(doc)

	if dto.ID != docID {
		t.Errorf("expected doc ID %v, got %v", docID, dto.ID)
	}
	if dto.OrganisationID != orgID {
		t.Errorf("expected org ID %v, got %v", orgID, dto.OrganisationID)
	}
	if dto.UploadedBy != userID {
		t.Errorf("expected user ID %v, got %v", userID, dto.UploadedBy)
	}
	if !dto.IsArchived {
		t.Errorf("expected is_archived to be true")
	}
	if dto.ArchivedAt == "" {
		t.Errorf("expected archived_at timestamp string, got empty")
	}
	if dto.Status != "completed" {
		t.Errorf("expected status 'completed', got %s", dto.Status)
	}
}
