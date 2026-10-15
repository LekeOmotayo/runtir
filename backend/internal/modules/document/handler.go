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
		docsGroup.GET("/:id", h.GetByID)
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

// List retrieves a paginated list of documents belonging to the authenticated tenant
func (h *Handler) List(c *gin.Context) {
	orgID, err := middleware.GetOrgID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Missing active organisation context", err.Error())
		return
	}

	page, _ := strconv.ParseInt(c.DefaultQuery("page", "1"), 10, 64)
	pageSize, _ := strconv.ParseInt(c.DefaultQuery("page_size", "20"), 10, 64)

	docs, total, err := h.service.ListDocuments(c.Request.Context(), orgID, page, pageSize)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve documents", err.Error())
		return
	}

	utils.PaginatedSuccessResponse(c, "Documents retrieved successfully", docs, page, pageSize, total)
}
