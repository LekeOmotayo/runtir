package document

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"

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
)

var allowedExtensions = map[string]string{
	".pdf":  "application/pdf",
	".txt":  "text/plain",
	".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
	".doc":  "application/msword",
	".md":   "text/markdown",
}

type DocumentDTO struct {
	ID             uuid.UUID `json:"id"`
	OrganisationID uuid.UUID `json:"organisation_id"`
	UploadedBy     uuid.UUID `json:"uploaded_by,omitempty"`
	Title          string    `json:"title"`
	FileName       string    `json:"file_name"`
	FilePath       string    `json:"file_path"`
	FileSize       int64     `json:"file_size"`
	MimeType       string    `json:"mime_type"`
	FileHash       string    `json:"file_hash"`
	Status         string    `json:"status"`
	ErrorMessage   string    `json:"error_message,omitempty"`
	CreatedAt      string    `json:"created_at"`
	UpdatedAt      string    `json:"updated_at"`
}

type Service struct {
	cfg     *config.Config
	queries *db.Queries
	storage storage.StorageService
}

func NewService(cfg *config.Config, queries *db.Queries, storage storage.StorageService) *Service {
	return &Service{
		cfg:     cfg,
		queries: queries,
		storage: storage,
	}
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
		// Clean up stored file if db insertion fails
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
		// Non-fatal if job record fails, but log
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

	return toDocumentDTO(doc), nil
}

// GetDocument retrieves a document by ID with multi-tenant organisation isolation
func (s *Service) GetDocument(ctx context.Context, orgID, docID uuid.UUID) (*DocumentDTO, error) {
	doc, err := s.queries.GetDocumentByID(ctx, db.GetDocumentByIDParams{
		ID:             docID,
		OrganisationID: orgID,
	})
	if err != nil {
		return nil, ErrDocumentNotFound
	}

	return toDocumentDTO(doc), nil
}

// ListDocuments lists documents belonging to an organisation with pagination
func (s *Service) ListDocuments(ctx context.Context, orgID uuid.UUID, page, pageSize int64) ([]DocumentDTO, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	docs, err := s.queries.ListDocumentsByOrg(ctx, db.ListDocumentsByOrgParams{
		OrganisationID: orgID,
		Limit:          int32(pageSize),
		Offset:         int32(offset),
	})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query documents: %w", err)
	}

	total, err := s.queries.CountDocumentsByOrg(ctx, orgID)
	if err != nil {
		total = int64(len(docs))
	}

	var dtos []DocumentDTO
	for _, d := range docs {
		dtos = append(dtos, *toDocumentDTO(d))
	}

	return dtos, total, nil
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
		createdAtStr = d.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00")
	}

	updatedAtStr := ""
	if d.UpdatedAt.Valid {
		updatedAtStr = d.UpdatedAt.Time.Format("2006-01-02T15:04:05Z07:00")
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
		CreatedAt:      createdAtStr,
		UpdatedAt:      updatedAtStr,
	}
}

func escapeJSON(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	return s
}
