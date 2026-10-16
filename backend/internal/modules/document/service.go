package document

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/lekeOmotayo/runtir/internal/config"
	"github.com/lekeOmotayo/runtir/internal/db"
	"github.com/lekeOmotayo/runtir/internal/storage"
)

var (
	ErrUnsupportedFileType = errors.New("unsupported file format. Allowed formats: PDF, TXT, DOCX, DOC, MD")
	ErrFileTooLarge        = errors.New("file size exceeds maximum allowed limit")
	ErrDocumentNotFound    = errors.New("document not found")
	ErrAlreadyProcessing   = errors.New("document is already being processed")
)

var allowedExtensions = map[string]string{
	".pdf":  "application/pdf",
	".txt":  "text/plain",
	".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
	".doc":  "application/msword",
	".md":   "text/markdown",
}

type DocumentDTO struct {
	ID             uuid.UUID         `json:"id"`
	OrganisationID uuid.UUID         `json:"organisation_id"`
	UploadedBy     uuid.UUID         `json:"uploaded_by,omitempty"`
	Title          string            `json:"title"`
	FileName       string            `json:"file_name"`
	FilePath       string            `json:"file_path"`
	FileSize       int64             `json:"file_size"`
	MimeType       string            `json:"mime_type"`
	FileHash       string            `json:"file_hash"`
	Status         string            `json:"status"`
	ErrorMessage   string            `json:"error_message,omitempty"`
	IsArchived     bool              `json:"is_archived"`
	ArchivedAt     string            `json:"archived_at,omitempty"`
	Metadata       map[string]string `json:"metadata,omitempty"`
	CreatedAt      string            `json:"created_at"`
	UpdatedAt      string            `json:"updated_at"`
}

type ContentSearchResultDTO struct {
	ChunkID       uuid.UUID `json:"chunk_id"`
	DocumentID    uuid.UUID `json:"document_id"`
	DocumentTitle string    `json:"document_title"`
	FileName      string    `json:"file_name"`
	MimeType      string    `json:"mime_type"`
	ChunkIndex    int32     `json:"chunk_index"`
	PageNumber    *int32    `json:"page_number,omitempty"`
	Snippet       string    `json:"snippet"`
}

type ProcessingStatusDTO struct {
	DocumentID   uuid.UUID `json:"document_id"`
	Status       string    `json:"status"`
	ErrorMessage string    `json:"error_message,omitempty"`
	JobType      string    `json:"job_type,omitempty"`
	Attempts     int32     `json:"attempts"`
	MaxAttempts  int32     `json:"max_attempts"`
	StartedAt    string    `json:"started_at,omitempty"`
	CompletedAt  string    `json:"completed_at,omitempty"`
	PageCount    int32     `json:"page_count"`
	WordCount    int32     `json:"word_count"`
	ChunkCount   int64     `json:"chunk_count"`
}

type UpdateDocumentInput struct {
	Title    *string           `json:"title"`
	Metadata map[string]string `json:"metadata"`
}

type DocumentProcessor interface {
	ProcessDocument(ctx context.Context, orgID, docID uuid.UUID) error
}

type Service struct {
	cfg       *config.Config
	queries   *db.Queries
	storage   storage.StorageService
	processor DocumentProcessor
}

func NewService(cfg *config.Config, queries *db.Queries, storage storage.StorageService) *Service {
	return &Service{
		cfg:     cfg,
		queries: queries,
		storage: storage,
	}
}

func (s *Service) SetProcessor(p DocumentProcessor) {
	s.processor = p
}

// IngestDocument handles uploading, hashing, disk storage, and metadata recording for a document
func (s *Service) IngestDocument(ctx context.Context, orgID, userID uuid.UUID, fileHeader *multipart.FileHeader, customTitle string) (*DocumentDTO, error) {
	// 1. Enforce max file size
	maxBytes := s.cfg.MaxUploadSizeMB * 1024 * 1024
	if fileHeader.Size > maxBytes {
		return nil, fmt.Errorf("%w (%d MB max)", ErrFileTooLarge, s.cfg.MaxUploadSizeMB)
	}

	// 2. Validate file extension
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	detectedMime, allowed := allowedExtensions[ext]
	if !allowed {
		return nil, ErrUnsupportedFileType
	}

	// 3. Open multipart file stream
	srcFile, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer srcFile.Close()

	// 4. Save file via StorageService (streams to disk & calculates SHA-256)
	storedFile, err := s.storage.Save(ctx, orgID, fileHeader.Filename, detectedMime, srcFile)
	if err != nil {
		return nil, fmt.Errorf("failed to store file: %w", err)
	}

	// 5. Determine title
	title := strings.TrimSpace(customTitle)
	if title == "" {
		title = strings.TrimSuffix(fileHeader.Filename, ext)
	}

	// 6. Insert document record in PostgreSQL
	doc, err := s.queries.CreateDocument(ctx, db.CreateDocumentParams{
		OrganisationID: orgID,
		UploadedBy:     pgtype.UUID{Bytes: userID, Valid: true},
		Title:          title,
		FileName:       fileHeader.Filename,
		FilePath:       storedFile.RelativePath,
		FileSize:       storedFile.FileSize,
		MimeType:       storedFile.MimeType,
		FileHash:       storedFile.FileHash,
		Status:         "queued",
	})
	if err != nil {
		_ = s.storage.Delete(ctx, storedFile.RelativePath)
		return nil, fmt.Errorf("failed to save document record: %w", err)
	}

	// 7. Create background processing job for extraction pipeline
	_, err = s.queries.CreateProcessingJob(ctx, db.CreateProcessingJobParams{
		DocumentID:     doc.ID,
		OrganisationID: orgID,
		JobType:        "text_extraction",
		Status:         "queued",
		MaxAttempts:    3,
	})
	if err != nil {
		fmt.Printf("Warning: failed to create processing job for doc %s: %v\n", doc.ID, err)
	}

	// 8. Log activity
	_, _ = s.queries.CreateActivityLog(ctx, db.CreateActivityLogParams{
		OrganisationID: orgID,
		UserID:         pgtype.UUID{Bytes: userID, Valid: true},
		Action:         "document.uploaded",
		EntityType:     "document",
		EntityID:       pgtype.UUID{Bytes: doc.ID, Valid: true},
		Metadata: []byte(fmt.Sprintf(
			`{"title":"%s","file_name":"%s","file_size":%d}`,
			escapeJSON(doc.Title),
			escapeJSON(doc.FileName),
			doc.FileSize,
		)),
	})

	// 9. Trigger async processing pipeline if processor is configured
	if s.processor != nil {
		go func(oID, dID uuid.UUID) {
			if procErr := s.processor.ProcessDocument(context.Background(), oID, dID); procErr != nil {
				fmt.Printf("[document] background processing failed for doc %s: %v\n", dID, procErr)
			}
		}(orgID, doc.ID)
	}

	dto := toDocumentDTO(doc)
	return dto, nil
}

// GetDocument retrieves a document by ID with multi-tenant organisation isolation and loads metadata
func (s *Service) GetDocument(ctx context.Context, orgID, docID uuid.UUID) (*DocumentDTO, error) {
	doc, err := s.queries.GetDocumentByID(ctx, db.GetDocumentByIDParams{
		ID:             docID,
		OrganisationID: orgID,
	})
	if err != nil {
		return nil, ErrDocumentNotFound
	}

	dto := toDocumentDTO(doc)

	// Fetch custom metadata
	metaRows, err := s.queries.GetDocumentMetadata(ctx, docID)
	if err == nil && len(metaRows) > 0 {
		dto.Metadata = make(map[string]string)
		for _, m := range metaRows {
			dto.Metadata[m.Key] = m.Value
		}
	}

	return dto, nil
}

// ListDocuments lists documents with optional filtering by status, MIME type, and archive state
func (s *Service) ListDocuments(ctx context.Context, orgID uuid.UUID, status, mimeType string, isArchived bool, page, pageSize int64) ([]DocumentDTO, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	var statusParam pgtype.Text
	if strings.TrimSpace(status) != "" {
		statusParam = pgtype.Text{String: strings.TrimSpace(status), Valid: true}
	}

	var mimeParam pgtype.Text
	if strings.TrimSpace(mimeType) != "" {
		mimeParam = pgtype.Text{String: strings.TrimSpace(mimeType), Valid: true}
	}

	docs, err := s.queries.ListDocumentsFiltered(ctx, db.ListDocumentsFilteredParams{
		OrganisationID: orgID,
		IsArchived:     isArchived,
		Status:         statusParam,
		MimeType:       mimeParam,
		Limit:          int32(pageSize),
		Offset:         int32(offset),
	})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query documents: %w", err)
	}

	total, err := s.queries.CountDocumentsFiltered(ctx, db.CountDocumentsFilteredParams{
		OrganisationID: orgID,
		IsArchived:     isArchived,
		Status:         statusParam,
		MimeType:       mimeParam,
	})
	if err != nil {
		total = int64(len(docs))
	}

	var dtos []DocumentDTO
	for _, d := range docs {
		dtos = append(dtos, *toDocumentDTO(d))
	}

	return dtos, total, nil
}

// SearchDocuments performs a keyword search on document titles and filenames
func (s *Service) SearchDocuments(ctx context.Context, orgID uuid.UUID, query string, page, pageSize int64) ([]DocumentDTO, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	docs, err := s.queries.SearchDocuments(ctx, db.SearchDocumentsParams{
		OrganisationID: orgID,
		Column2:        pgtype.Text{String: query, Valid: true},
		Limit:          int32(pageSize),
		Offset:         int32(offset),
	})
	if err != nil {
		return nil, 0, fmt.Errorf("search query failed: %w", err)
	}

	total, err := s.queries.CountSearchDocuments(ctx, db.CountSearchDocumentsParams{
		OrganisationID: orgID,
		Column2:        pgtype.Text{String: query, Valid: true},
	})
	if err != nil {
		total = int64(len(docs))
	}

	var dtos []DocumentDTO
	for _, d := range docs {
		dtos = append(dtos, *toDocumentDTO(d))
	}

	return dtos, total, nil
}

// SearchContent performs a full-text search across document chunks and returns matching passages
func (s *Service) SearchContent(ctx context.Context, orgID uuid.UUID, query string, page, pageSize int64) ([]ContentSearchResultDTO, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	cleanQuery := strings.TrimSpace(query)
	if cleanQuery == "" {
		return []ContentSearchResultDTO{}, 0, nil
	}

	chunks, err := s.queries.SearchChunksWithDocument(ctx, db.SearchChunksWithDocumentParams{
		OrganisationID: orgID,
		Column2:        pgtype.Text{String: cleanQuery, Valid: true},
		Limit:          int32(pageSize),
		Offset:         int32(offset),
	})
	if err != nil {
		return nil, 0, fmt.Errorf("content search failed: %w", err)
	}

	total, err := s.queries.CountSearchChunksWithDocument(ctx, db.CountSearchChunksWithDocumentParams{
		OrganisationID: orgID,
		Column2:        pgtype.Text{String: cleanQuery, Valid: true},
	})
	if err != nil {
		total = int64(len(chunks))
	}

	results := make([]ContentSearchResultDTO, 0, len(chunks))
	for _, c := range chunks {
		snippet := createHighlightedSnippet(c.Content, cleanQuery, 250)
		var pageNum *int32
		if c.PageNumber.Valid {
			v := c.PageNumber.Int32
			pageNum = &v
		}

		results = append(results, ContentSearchResultDTO{
			ChunkID:       c.ChunkID,
			DocumentID:    c.DocumentID,
			DocumentTitle: c.DocumentTitle,
			FileName:      c.DocumentFileName,
			MimeType:      c.DocumentMimeType,
			ChunkIndex:    c.ChunkIndex,
			PageNumber:    pageNum,
			Snippet:       snippet,
		})
	}

	return results, total, nil
}

// UpdateDocument updates title and/or custom key-value metadata
func (s *Service) UpdateDocument(ctx context.Context, orgID, docID uuid.UUID, input UpdateDocumentInput) (*DocumentDTO, error) {
	if _, err := s.queries.GetDocumentByID(ctx, db.GetDocumentByIDParams{
		ID:             docID,
		OrganisationID: orgID,
	}); err != nil {
		return nil, ErrDocumentNotFound
	}

	if input.Title != nil && strings.TrimSpace(*input.Title) != "" {
		_, err := s.queries.UpdateDocumentTitle(ctx, db.UpdateDocumentTitleParams{
			ID:             docID,
			OrganisationID: orgID,
			Title:          strings.TrimSpace(*input.Title),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to update title: %w", err)
		}
	}

	if len(input.Metadata) > 0 {
		for k, v := range input.Metadata {
			trimmedKey := strings.TrimSpace(k)
			if trimmedKey != "" {
				_, _ = s.queries.UpsertDocumentMetadata(ctx, db.UpsertDocumentMetadataParams{
					DocumentID: docID,
					Key:        trimmedKey,
					Value:      v,
				})
			}
		}
	}

	return s.GetDocument(ctx, orgID, docID)
}

// ArchiveDocument marks a document as archived
func (s *Service) ArchiveDocument(ctx context.Context, orgID, docID uuid.UUID) (*DocumentDTO, error) {
	doc, err := s.queries.ArchiveDocument(ctx, db.ArchiveDocumentParams{
		ID:             docID,
		OrganisationID: orgID,
	})
	if err != nil {
		return nil, ErrDocumentNotFound
	}

	_, _ = s.queries.CreateActivityLog(ctx, db.CreateActivityLogParams{
		OrganisationID: orgID,
		UserID:         doc.UploadedBy,
		Action:         "document.archived",
		EntityType:     "document",
		EntityID:       pgtype.UUID{Bytes: docID, Valid: true},
		Metadata:       []byte(fmt.Sprintf(`{"title":"%s"}`, escapeJSON(doc.Title))),
	})

	return toDocumentDTO(doc), nil
}

// UnarchiveDocument restores an archived document
func (s *Service) UnarchiveDocument(ctx context.Context, orgID, docID uuid.UUID) (*DocumentDTO, error) {
	doc, err := s.queries.UnarchiveDocument(ctx, db.UnarchiveDocumentParams{
		ID:             docID,
		OrganisationID: orgID,
	})
	if err != nil {
		return nil, ErrDocumentNotFound
	}

	_, _ = s.queries.CreateActivityLog(ctx, db.CreateActivityLogParams{
		OrganisationID: orgID,
		UserID:         doc.UploadedBy,
		Action:         "document.unarchived",
		EntityType:     "document",
		EntityID:       pgtype.UUID{Bytes: docID, Valid: true},
		Metadata:       []byte(fmt.Sprintf(`{"title":"%s"}`, escapeJSON(doc.Title))),
	})

	return toDocumentDTO(doc), nil
}

// GetProcessingStatus returns current job state, extraction stats, and chunk count
func (s *Service) GetProcessingStatus(ctx context.Context, orgID, docID uuid.UUID) (*ProcessingStatusDTO, error) {
	doc, err := s.queries.GetDocumentByID(ctx, db.GetDocumentByIDParams{
		ID:             docID,
		OrganisationID: orgID,
	})
	if err != nil {
		return nil, ErrDocumentNotFound
	}

	statusDTO := &ProcessingStatusDTO{
		DocumentID: docID,
		Status:     doc.Status,
	}
	if doc.ErrorMessage.Valid {
		statusDTO.ErrorMessage = doc.ErrorMessage.String
	}

	jobs, err := s.queries.ListProcessingJobsByDoc(ctx, db.ListProcessingJobsByDocParams{
		DocumentID:     docID,
		OrganisationID: orgID,
	})
	if err == nil && len(jobs) > 0 {
		latestJob := jobs[len(jobs)-1]
		statusDTO.JobType = latestJob.JobType
		statusDTO.Attempts = latestJob.Attempts
		statusDTO.MaxAttempts = latestJob.MaxAttempts
		if latestJob.StartedAt.Valid {
			statusDTO.StartedAt = latestJob.StartedAt.Time.Format(time.RFC3339)
		}
		if latestJob.CompletedAt.Valid {
			statusDTO.CompletedAt = latestJob.CompletedAt.Time.Format(time.RFC3339)
		}
		if latestJob.ErrorMessage.Valid && statusDTO.ErrorMessage == "" {
			statusDTO.ErrorMessage = latestJob.ErrorMessage.String
		}
	}

	extraction, err := s.queries.GetExtractionByDocumentID(ctx, docID)
	if err == nil {
		statusDTO.PageCount = extraction.PageCount
		statusDTO.WordCount = extraction.WordCount
	}

	chunkCount, err := s.queries.CountChunksByDocument(ctx, docID)
	if err == nil {
		statusDTO.ChunkCount = chunkCount
	}

	return statusDTO, nil
}

// RetryProcessing resets job state and re-executes text extraction and chunking asynchronously
func (s *Service) RetryProcessing(ctx context.Context, orgID, docID uuid.UUID) (*DocumentDTO, error) {
	doc, err := s.queries.GetDocumentByID(ctx, db.GetDocumentByIDParams{
		ID:             docID,
		OrganisationID: orgID,
	})
	if err != nil {
		return nil, ErrDocumentNotFound
	}

	if doc.Status == "processing" {
		return nil, ErrAlreadyProcessing
	}

	// 1. Reset document status
	updatedDoc, err := s.queries.UpdateDocumentStatus(ctx, db.UpdateDocumentStatusParams{
		ID:             docID,
		OrganisationID: orgID,
		Status:         "queued",
		ErrorMessage:   pgtype.Text{Valid: false},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to reset document status: %w", err)
	}

	// 2. Reset processing job if exists
	jobs, err := s.queries.ListProcessingJobsByDoc(ctx, db.ListProcessingJobsByDocParams{
		DocumentID:     docID,
		OrganisationID: orgID,
	})
	if err == nil && len(jobs) > 0 {
		_, _ = s.queries.ResetProcessingJob(ctx, jobs[0].ID)
	}

	// 3. Trigger processing in background
	if s.processor != nil {
		go func(oID, dID uuid.UUID) {
			if procErr := s.processor.ProcessDocument(context.Background(), oID, dID); procErr != nil {
				fmt.Printf("[document] retry processing failed for doc %s: %v\n", dID, procErr)
			}
		}(orgID, docID)
	}

	return toDocumentDTO(updatedDoc), nil
}

// DeleteDocument removes a document record and its file from storage
func (s *Service) DeleteDocument(ctx context.Context, orgID, docID uuid.UUID) error {
	doc, err := s.queries.GetDocumentByID(ctx, db.GetDocumentByIDParams{
		ID:             docID,
		OrganisationID: orgID,
	})
	if err != nil {
		return ErrDocumentNotFound
	}

	// Delete DB record first
	if err := s.queries.DeleteDocument(ctx, db.DeleteDocumentParams{
		ID:             docID,
		OrganisationID: orgID,
	}); err != nil {
		return fmt.Errorf("failed to delete document record: %w", err)
	}

	_ = s.storage.Delete(ctx, doc.FilePath)
	return nil
}

// OpenFile returns a reader for the raw file so it can be streamed to the client
func (s *Service) OpenFile(ctx context.Context, orgID, docID uuid.UUID) (interface{ Read([]byte) (int, error); Close() error }, string, string, error) {
	doc, err := s.queries.GetDocumentByID(ctx, db.GetDocumentByIDParams{
		ID:             docID,
		OrganisationID: orgID,
	})
	if err != nil {
		return nil, "", "", ErrDocumentNotFound
	}

	reader, err := s.storage.Open(ctx, doc.FilePath)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to open file: %w", err)
	}

	return reader, doc.FileName, doc.MimeType, nil
}

func toDocumentDTO(d db.Document) *DocumentDTO {
	var uploadedBy uuid.UUID
	if d.UploadedBy.Valid {
		uploadedBy = d.UploadedBy.Bytes
	}

	errMsg := ""
	if d.ErrorMessage.Valid {
		errMsg = d.ErrorMessage.String
	}

	createdAtStr := ""
	if d.CreatedAt.Valid {
		createdAtStr = d.CreatedAt.Time.Format(time.RFC3339)
	}

	updatedAtStr := ""
	if d.UpdatedAt.Valid {
		updatedAtStr = d.UpdatedAt.Time.Format(time.RFC3339)
	}

	archivedAtStr := ""
	if d.ArchivedAt.Valid {
		archivedAtStr = d.ArchivedAt.Time.Format(time.RFC3339)
	}

	return &DocumentDTO{
		ID:             d.ID,
		OrganisationID: d.OrganisationID,
		UploadedBy:     uploadedBy,
		Title:          d.Title,
		FileName:       d.FileName,
		FilePath:       d.FilePath,
		FileSize:       d.FileSize,
		MimeType:       d.MimeType,
		FileHash:       d.FileHash,
		Status:         d.Status,
		ErrorMessage:   errMsg,
		IsArchived:     d.IsArchived,
		ArchivedAt:     archivedAtStr,
		CreatedAt:      createdAtStr,
		UpdatedAt:      updatedAtStr,
	}
}

func createHighlightedSnippet(content, query string, maxLen int) string {
	lowerContent := strings.ToLower(content)
	lowerQuery := strings.ToLower(query)

	idx := strings.Index(lowerContent, lowerQuery)
	if idx < 0 {
		if len(content) > maxLen {
			return content[:maxLen] + "..."
		}
		return content
	}

	matchLen := len(query)
	padding := (maxLen - matchLen) / 2
	if padding < 0 {
		padding = 0
	}

	start := idx - padding
	if start < 0 {
		start = 0
	}
	end := start + maxLen
	if end > len(content) {
		end = len(content)
		if end-maxLen >= 0 {
			start = end - maxLen
		} else {
			start = 0
		}
	}

	snippet := content[start:end]
	if start > 0 {
		snippet = "..." + snippet
	}
	if end < len(content) {
		snippet = snippet + "..."
	}
	return snippet
}

func escapeJSON(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	return s
}
