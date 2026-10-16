package processor

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/lekeOmotayo/runtir/internal/db"
	"github.com/lekeOmotayo/runtir/internal/storage"
)

type Service struct {
	queries *db.Queries
	storage storage.StorageService
}

func NewService(queries *db.Queries, storage storage.StorageService) *Service {
	return &Service{
		queries: queries,
		storage: storage,
	}
}

// ProcessDocument runs the full extraction + chunking pipeline for a single document
func (s *Service) ProcessDocument(ctx context.Context, orgID, docID uuid.UUID) error {
	log.Printf("[processor] Starting processing for document %s", docID)

	// 1. Mark document as processing
	_, err := s.queries.UpdateDocumentStatus(ctx, db.UpdateDocumentStatusParams{
		ID:             docID,
		OrganisationID: orgID,
		Status:         "processing",
		ErrorMessage:   pgtype.Text{Valid: false},
	})
	if err != nil {
		return fmt.Errorf("failed to update document status to processing: %w", err)
	}

	// 2. Find the pending processing job
	jobs, err := s.queries.ListProcessingJobsByDoc(ctx, db.ListProcessingJobsByDocParams{
		DocumentID:     docID,
		OrganisationID: orgID,
	})
	if err != nil || len(jobs) == 0 {
		return fmt.Errorf("no processing job found for document %s", docID)
	}
	job := jobs[0]

	// 3. Fetch document record to get file path and MIME type
	doc, err := s.queries.GetDocumentByID(ctx, db.GetDocumentByIDParams{
		ID:             docID,
		OrganisationID: orgID,
	})
	if err != nil {
		s.markFailed(ctx, docID, orgID, job.ID, "Document record not found")
		return fmt.Errorf("document not found: %w", err)
	}

	// 4. Mark job as started
	now := pgtype.Timestamptz{Time: time.Now(), Valid: true}
	_, _ = s.queries.UpdateProcessingJobStatus(ctx, db.UpdateProcessingJobStatusParams{
		ID:           job.ID,
		Status:       "processing",
		Attempts:     job.Attempts + 1,
		ErrorMessage: pgtype.Text{Valid: false},
		StartedAt:    now,
		CompletedAt:  pgtype.Timestamptz{Valid: false},
	})

	// 5. Open the file from storage
	fileReader, err := s.storage.Open(ctx, doc.FilePath)
	if err != nil {
		s.markFailed(ctx, docID, orgID, job.ID, fmt.Sprintf("Failed to open file: %v", err))
		return err
	}
	defer fileReader.Close()

	// 6. Extract text from the document
	rawText, pageCount, err := ExtractText(ctx, doc.MimeType, fileReader)
	if err != nil {
		s.markFailed(ctx, docID, orgID, job.ID, fmt.Sprintf("Text extraction failed: %v", err))
		return err
	}

	wordCount := countWords(rawText)

	// 7. Save extraction record
	_, err = s.queries.CreateExtraction(ctx, db.CreateExtractionParams{
		DocumentID:       docID,
		RawText:          rawText,
		PageCount:        int32(pageCount),
		WordCount:        int32(wordCount),
		ExtractionMethod: "standard",
	})
	if err != nil {
		s.markFailed(ctx, docID, orgID, job.ID, fmt.Sprintf("Failed to save extraction: %v", err))
		return err
	}

	// 8. Chunk the text
	chunks := ChunkText(rawText, 800, 100)
	for i, chunkContent := range chunks {
		tokenCount := EstimateTokenCount(chunkContent)
		_, err := s.queries.CreateChunk(ctx, db.CreateChunkParams{
			DocumentID:     docID,
			OrganisationID: orgID,
			ChunkIndex:     int32(i),
			Content:        chunkContent,
			TokenCount:     int32(tokenCount),
			PageNumber:     pgtype.Int4{Valid: false},
		})
		if err != nil {
			log.Printf("[processor] Warning: failed to save chunk %d for doc %s: %v", i, docID, err)
		}
	}

	log.Printf("[processor] Document %s: extracted %d words, %d pages, %d chunks", docID, wordCount, pageCount, len(chunks))

	// 9. Mark document and job as completed
	completedAt := pgtype.Timestamptz{Time: time.Now(), Valid: true}
	_, _ = s.queries.UpdateProcessingJobStatus(ctx, db.UpdateProcessingJobStatusParams{
		ID:           job.ID,
		Status:       "completed",
		Attempts:     job.Attempts + 1,
		ErrorMessage: pgtype.Text{Valid: false},
		StartedAt:    now,
		CompletedAt:  completedAt,
	})

	_, _ = s.queries.UpdateDocumentStatus(ctx, db.UpdateDocumentStatusParams{
		ID:             docID,
		OrganisationID: orgID,
		Status:         "completed",
		ErrorMessage:   pgtype.Text{Valid: false},
	})

	log.Printf("[processor] Document %s processing completed", docID)
	return nil
}

// markFailed sets document and job status to failed with an error message
func (s *Service) markFailed(ctx context.Context, docID, orgID, jobID uuid.UUID, errMsg string) {
	log.Printf("[processor] Document %s failed: %s", docID, errMsg)

	_, _ = s.queries.UpdateDocumentStatus(ctx, db.UpdateDocumentStatusParams{
		ID:             docID,
		OrganisationID: orgID,
		Status:         "failed",
		ErrorMessage:   pgtype.Text{String: errMsg, Valid: true},
	})

	jobs, _ := s.queries.ListProcessingJobsByDoc(ctx, db.ListProcessingJobsByDocParams{
		DocumentID:     docID,
		OrganisationID: orgID,
	})
	if len(jobs) > 0 {
		job := jobs[0]
		_, _ = s.queries.UpdateProcessingJobStatus(ctx, db.UpdateProcessingJobStatusParams{
			ID:           job.ID,
			Status:       "failed",
			Attempts:     job.Attempts + 1,
			ErrorMessage: pgtype.Text{String: errMsg, Valid: true},
			StartedAt:    pgtype.Timestamptz{Valid: false},
			CompletedAt:  pgtype.Timestamptz{Time: time.Now(), Valid: true},
		})
	}
	_ = os.Stderr // suppress unused warning
	_ = jobID
}
