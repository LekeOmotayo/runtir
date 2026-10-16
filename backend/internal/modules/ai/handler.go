package ai

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
	return &Handler{service: service}
}

// RegisterRoutes mounts AI endpoints on the API group
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, authMW gin.HandlerFunc) {
	ai := rg.Group("/ai", authMW)
	{
		// Conversations
		ai.POST("/conversations", h.CreateConversation)
		ai.GET("/conversations", h.ListConversations)

		// Messages & Q&A
		ai.GET("/conversations/:conv_id/messages", h.GetMessages)
		ai.POST("/conversations/:conv_id/ask", h.Ask)
	}
}

// CreateConversation starts a new AI chat thread
func (h *Handler) CreateConversation(c *gin.Context) {
	orgID, err := middleware.GetOrgID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Missing organisation context", err.Error())
		return
	}
	userID, err := middleware.GetUserID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Missing user context", err.Error())
		return
	}

	var req CreateConversationRequest
	_ = c.ShouldBindJSON(&req) // title is optional

	conv, err := h.service.CreateConversation(c.Request.Context(), orgID, userID, req.Title)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create conversation", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Conversation created", conv)
}

// ListConversations returns paginated conversations for the current user
func (h *Handler) ListConversations(c *gin.Context) {
	orgID, _ := middleware.GetOrgID(c)
	userID, _ := middleware.GetUserID(c)
	page, _ := strconv.ParseInt(c.DefaultQuery("page", "1"), 10, 64)
	pageSize, _ := strconv.ParseInt(c.DefaultQuery("page_size", "20"), 10, 64)

	convs, err := h.service.ListConversations(c.Request.Context(), orgID, userID, page, pageSize)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to list conversations", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Conversations retrieved", convs)
}

// GetMessages returns all messages in a conversation
func (h *Handler) GetMessages(c *gin.Context) {
	orgID, _ := middleware.GetOrgID(c)

	convID, err := uuid.Parse(c.Param("conv_id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid conversation ID", err.Error())
		return
	}

	messages, err := h.service.GetMessages(c.Request.Context(), orgID, convID)
	if err != nil {
		if errors.Is(err, err) {
			utils.ErrorResponse(c, http.StatusNotFound, "Conversation not found", err.Error())
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve messages", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Messages retrieved", messages)
}

// Ask runs the full RAG pipeline and returns an AI-generated answer with citations
func (h *Handler) Ask(c *gin.Context) {
	orgID, _ := middleware.GetOrgID(c)
	userID, _ := middleware.GetUserID(c)

	convID, err := uuid.Parse(c.Param("conv_id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid conversation ID", err.Error())
		return
	}

	var req AskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Question is required", err.Error())
		return
	}

	msg, err := h.service.Ask(c.Request.Context(), orgID, userID, convID, req.Question)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to process question", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Answer generated", msg)
}
