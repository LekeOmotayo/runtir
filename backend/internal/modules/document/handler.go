package document

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lekeOmotayo/runtir/internal/middleware"
	"github.com/lekeOmotayo/runtir/internal/utils"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

// RegisterRoutes registers the document endpoints onto the router group
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, authMW gin.HandlerFunc) {
	docsGroup := rg.Group("/documents", authMW)
	{
		docsGroup.POST("/upload", h.Upload)
		docsGroup.POST("", h.Upload)
		docsGroup.GET("", h.List)
		docsGroup.GET("/search", h.SearchContent)
		docsGroup.GET("/:id", h.GetByID)
		docsGroup.PATCH("/:id", h.Update)
		docsGroup.POST("/:id/archive", h.Archive)
		docsGroup.POST("/:id/unarchive", h.Unarchive)
		docsGroup.GET("/:id/processing", h.GetProcessingStatus)
		docsGroup.POST("/:id/retry", h.RetryProcessing)
		docsGroup.DELETE("/:id", h.Delete)
		docsGroup.GET("/:id/download", h.Download)
	}
}

// Upload handles multipart document ingestion
func (h *Handler) Upload(c *gin.Context) {
	orgID, err := middleware.GetOrgID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Missing active organisation context", err.Error())
		return
	}

	userID, err := middleware.GetUserID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Missing authenticated user context", err.Error())
		return
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "No file provided in 'file' form field", err.Error())
		return
	}

	title := c.PostForm("title")

	doc, err := h.service.IngestDocument(c.Request.Context(), orgID, userID, fileHeader, title)
	if err != nil {
		if errors.Is(err, ErrUnsupportedFileType) || errors.Is(err, ErrFileTooLarge) {
			utils.ErrorResponse(c, http.StatusBadRequest, err.Error(), nil)
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Document ingestion failed", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Document uploaded and queued for processing", doc)
}

// GetByID retrieves a single document by its UUID within the tenant organisation
func (h *Handler) GetByID(c *gin.Context) {
	orgID, err := middleware.GetOrgID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Missing active organisation context", err.Error())
		return
	}

	idParam := c.Param("id")
	docID, err := uuid.Parse(idParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid document ID format", err.Error())
		return
	}

	doc, err := h.service.GetDocument(c.Request.Context(), orgID, docID)
	if err != nil {
		if errors.Is(err, ErrDocumentNotFound) {
			utils.ErrorResponse(c, http.StatusNotFound, "Document not found", nil)
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve document", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Document retrieved successfully", doc)
}

// List retrieves a paginated list of documents with filtering (q, status, mime_type, is_archived)
func (h *Handler) List(c *gin.Context) {
	orgID, err := middleware.GetOrgID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Missing active organisation context", err.Error())
		return
	}

	query := c.Query("q")
	status := c.Query("status")
	mimeType := c.Query("mime_type")
	archivedParam := c.Query("is_archived")
	isArchived := archivedParam == "true" || archivedParam == "1"

	page, _ := strconv.ParseInt(c.DefaultQuery("page", "1"), 10, 64)
	pageSize, _ := strconv.ParseInt(c.DefaultQuery("page_size", "20"), 10, 64)

	var docs []DocumentDTO
	var total int64

	if query != "" {
		docs, total, err = h.service.SearchDocuments(c.Request.Context(), orgID, query, page, pageSize)
	} else {
		docs, total, err = h.service.ListDocuments(c.Request.Context(), orgID, status, mimeType, isArchived, page, pageSize)
	}

	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve documents", err.Error())
		return
	}

	utils.PaginatedSuccessResponse(c, "Documents retrieved successfully", docs, page, pageSize, total)
}

// SearchContent performs full-text search across document chunks, returning matching text passages with snippets
func (h *Handler) SearchContent(c *gin.Context) {
	orgID, err := middleware.GetOrgID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Missing active organisation context", err.Error())
		return
	}

	query := c.Query("q")
	if query == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Missing search query parameter 'q'", nil)
		return
	}

	page, _ := strconv.ParseInt(c.DefaultQuery("page", "1"), 10, 64)
	pageSize, _ := strconv.ParseInt(c.DefaultQuery("page_size", "20"), 10, 64)

	results, total, err := h.service.SearchContent(c.Request.Context(), orgID, query, page, pageSize)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to search document contents", err.Error())
		return
	}

	utils.PaginatedSuccessResponse(c, "Document content search completed", results, page, pageSize, total)
}

// Update updates document title and/or custom metadata key-values
func (h *Handler) Update(c *gin.Context) {
	orgID, err := middleware.GetOrgID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Missing active organisation context", err.Error())
		return
	}

	idParam := c.Param("id")
	docID, err := uuid.Parse(idParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid document ID format", err.Error())
		return
	}

	var req UpdateDocumentInput
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	doc, err := h.service.UpdateDocument(c.Request.Context(), orgID, docID, req)
	if err != nil {
		if errors.Is(err, ErrDocumentNotFound) {
			utils.ErrorResponse(c, http.StatusNotFound, "Document not found", nil)
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update document", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Document updated successfully", doc)
}

// Archive marks a document as archived
func (h *Handler) Archive(c *gin.Context) {
	orgID, err := middleware.GetOrgID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Missing active organisation context", err.Error())
		return
	}

	idParam := c.Param("id")
	docID, err := uuid.Parse(idParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid document ID format", err.Error())
		return
	}

	doc, err := h.service.ArchiveDocument(c.Request.Context(), orgID, docID)
	if err != nil {
		if errors.Is(err, ErrDocumentNotFound) {
			utils.ErrorResponse(c, http.StatusNotFound, "Document not found", nil)
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to archive document", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Document archived successfully", doc)
}

// Unarchive restores an archived document
func (h *Handler) Unarchive(c *gin.Context) {
	orgID, err := middleware.GetOrgID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Missing active organisation context", err.Error())
		return
	}

	idParam := c.Param("id")
	docID, err := uuid.Parse(idParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid document ID format", err.Error())
		return
	}

	doc, err := h.service.UnarchiveDocument(c.Request.Context(), orgID, docID)
	if err != nil {
		if errors.Is(err, ErrDocumentNotFound) {
			utils.ErrorResponse(c, http.StatusNotFound, "Document not found", nil)
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to unarchive document", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Document unarchived successfully", doc)
}

// GetProcessingStatus returns extraction and job status details for a document
func (h *Handler) GetProcessingStatus(c *gin.Context) {
	orgID, err := middleware.GetOrgID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Missing active organisation context", err.Error())
		return
	}

	idParam := c.Param("id")
	docID, err := uuid.Parse(idParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid document ID format", err.Error())
		return
	}

	statusDTO, err := h.service.GetProcessingStatus(c.Request.Context(), orgID, docID)
	if err != nil {
		if errors.Is(err, ErrDocumentNotFound) {
			utils.ErrorResponse(c, http.StatusNotFound, "Document not found", nil)
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve processing status", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Document processing status retrieved", statusDTO)
}

// RetryProcessing re-triggers text extraction and chunking for a failed document
func (h *Handler) RetryProcessing(c *gin.Context) {
	orgID, err := middleware.GetOrgID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Missing active organisation context", err.Error())
		return
	}

	idParam := c.Param("id")
	docID, err := uuid.Parse(idParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid document ID format", err.Error())
		return
	}

	doc, err := h.service.RetryProcessing(c.Request.Context(), orgID, docID)
	if err != nil {
		if errors.Is(err, ErrDocumentNotFound) {
			utils.ErrorResponse(c, http.StatusNotFound, "Document not found", nil)
			return
		}
		if errors.Is(err, ErrAlreadyProcessing) {
			utils.ErrorResponse(c, http.StatusConflict, err.Error(), nil)
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retry document processing", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Document queued for re-processing", doc)
}

// Delete deletes a document and its storage file
func (h *Handler) Delete(c *gin.Context) {
	orgID, err := middleware.GetOrgID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Missing active organisation context", err.Error())
		return
	}

	idParam := c.Param("id")
	docID, err := uuid.Parse(idParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid document ID format", err.Error())
		return
	}

	if err := h.service.DeleteDocument(c.Request.Context(), orgID, docID); err != nil {
		if errors.Is(err, ErrDocumentNotFound) {
			utils.ErrorResponse(c, http.StatusNotFound, "Document not found", nil)
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete document", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Document deleted successfully", nil)
}

// Download streams the raw file directly to the client
func (h *Handler) Download(c *gin.Context) {
	orgID, err := middleware.GetOrgID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Missing active organisation context", err.Error())
		return
	}

	idParam := c.Param("id")
	docID, err := uuid.Parse(idParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid document ID format", err.Error())
		return
	}

	reader, fileName, mimeType, err := h.service.OpenFile(c.Request.Context(), orgID, docID)
	if err != nil {
		if errors.Is(err, ErrDocumentNotFound) {
			utils.ErrorResponse(c, http.StatusNotFound, "Document not found", nil)
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to open document file", err.Error())
		return
	}
	defer reader.Close()

	c.Header("Content-Disposition", "attachment; filename=\""+fileName+"\"")
	c.Header("Content-Type", mimeType)
	c.DataFromReader(http.StatusOK, -1, mimeType, reader, nil)
}
