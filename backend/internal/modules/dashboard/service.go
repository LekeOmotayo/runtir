package dashboard

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lekeOmotayo/runtir/internal/db"
	"github.com/lekeOmotayo/runtir/internal/utils"
)

type StatsDTO struct {
	TotalDocuments      int64             `json:"total_documents"`
	CompletedDocuments  int64             `json:"completed_documents"`
	QueuedDocuments     int64             `json:"queued_documents"`
	ProcessingDocuments int64             `json:"processing_documents"`
	FailedDocuments     int64             `json:"failed_documents"`
	RecentDocuments     []DocumentSummary `json:"recent_documents"`
}

type DocumentSummary struct {
	ID        uuid.UUID `json:"id"`
	Title     string    `json:"title"`
	FileName  string    `json:"file_name"`
	MimeType  string    `json:"mime_type"`
	FileSize  int64     `json:"file_size"`
	Status    string    `json:"status"`
	CreatedAt string    `json:"created_at"`
}

type Service struct {
	queries *db.Queries
}

func NewService(queries *db.Queries) *Service {
	return &Service{queries: queries}
}

// GetStats returns dashboard statistics for the active organisation
func (s *Service) GetStats(ctx context.Context, orgID uuid.UUID) (*StatsDTO, error) {
	// 1. Total count
	total, err := s.queries.CountDocumentsByOrg(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to count documents: %w", err)
	}

	// 2. Status breakdown
	statusCounts, err := s.queries.CountDocumentsByStatus(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to count documents by status: %w", err)
	}

	var completed, queued, processing, failed int64
	for _, row := range statusCounts {
		switch row.Status {
		case "completed":
			completed = row.Count
		case "queued":
			queued = row.Count
		case "processing":
			processing = row.Count
		case "failed":
			failed = row.Count
		}
	}

	// 3. Recent documents (last 5)
	recent, err := s.queries.GetRecentDocuments(ctx, db.GetRecentDocumentsParams{
		OrganisationID: orgID,
		Limit:          5,
	})
	if err != nil {
		recent = nil // non-fatal
	}

	var recentSummaries []DocumentSummary
	for _, doc := range recent {
		createdAt := ""
		if doc.CreatedAt.Valid {
			createdAt = doc.CreatedAt.Time.Format(time.RFC3339)
		}
		recentSummaries = append(recentSummaries, DocumentSummary{
			ID:        doc.ID,
			Title:     doc.Title,
			FileName:  doc.FileName,
			MimeType:  doc.MimeType,
			FileSize:  doc.FileSize,
			Status:    doc.Status,
			CreatedAt: createdAt,
		})
	}

	_ = utils.SuccessResponse // referenced

	return &StatsDTO{
		TotalDocuments:      total,
		CompletedDocuments:  completed,
		QueuedDocuments:     queued,
		ProcessingDocuments: processing,
		FailedDocuments:     failed,
		RecentDocuments:     recentSummaries,
	}, nil
}
