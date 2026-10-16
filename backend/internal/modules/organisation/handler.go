package organisation

import (
	"errors"
	"net/http"

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

// RegisterRoutes registers organisation and membership endpoints
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, authMW gin.HandlerFunc) {
	orgGroup := rg.Group("/organisations", authMW)
	{
		orgGroup.GET("/current", h.GetCurrent)
		orgGroup.GET("/members", h.ListMembers)

		// Member management endpoints restricted to owner and admin roles
		privileged := orgGroup.Group("", middleware.RequireRole("owner", "admin"))
		{
			privileged.POST("/members", h.AddMember)
			privileged.PATCH("/members/:user_id", h.UpdateMemberRole)
			privileged.DELETE("/members/:user_id", h.RemoveMember)
		}
	}
}

// GetCurrent returns the active organisation profile
func (h *Handler) GetCurrent(c *gin.Context) {
	orgID, err := middleware.GetOrgID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Missing active organisation context", err.Error())
		return
	}

	org, err := h.service.GetCurrent(c.Request.Context(), orgID)
	if err != nil {
		if errors.Is(err, ErrOrgNotFound) {
			utils.ErrorResponse(c, http.StatusNotFound, "Organisation not found", nil)
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve organisation", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Organisation retrieved successfully", org)
}

// ListMembers returns all members belonging to the active organisation
func (h *Handler) ListMembers(c *gin.Context) {
	orgID, err := middleware.GetOrgID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Missing active organisation context", err.Error())
		return
	}

	members, err := h.service.ListMembers(c.Request.Context(), orgID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to list members", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Members retrieved successfully", members)
}

// AddMember adds or invites a user to the organisation
func (h *Handler) AddMember(c *gin.Context) {
	orgID, err := middleware.GetOrgID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Missing active organisation context", err.Error())
		return
	}

	var req AddMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	member, err := h.service.AddMember(c.Request.Context(), orgID, req)
	if err != nil {
		if errors.Is(err, ErrMemberAlreadyExists) {
			utils.ErrorResponse(c, http.StatusConflict, err.Error(), nil)
			return
		}
		if errors.Is(err, ErrInvalidRole) {
			utils.ErrorResponse(c, http.StatusBadRequest, err.Error(), nil)
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to add member", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Member added successfully", member)
}

// UpdateMemberRole modifies an existing member's role
func (h *Handler) UpdateMemberRole(c *gin.Context) {
	orgID, err := middleware.GetOrgID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Missing active organisation context", err.Error())
		return
	}

	userIDParam := c.Param("user_id")
	targetUserID, err := uuid.Parse(userIDParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID format", err.Error())
		return
	}

	var req UpdateMemberRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	member, err := h.service.UpdateMemberRole(c.Request.Context(), orgID, targetUserID, req.Role)
	if err != nil {
		if errors.Is(err, ErrMemberNotFound) {
			utils.ErrorResponse(c, http.StatusNotFound, "Member not found", nil)
			return
		}
		if errors.Is(err, ErrCannotRemoveLastOwner) {
			utils.ErrorResponse(c, http.StatusForbidden, err.Error(), nil)
			return
		}
		if errors.Is(err, ErrInvalidRole) {
			utils.ErrorResponse(c, http.StatusBadRequest, err.Error(), nil)
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update member role", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Member role updated successfully", member)
}

// RemoveMember deletes a user's membership from the organisation
func (h *Handler) RemoveMember(c *gin.Context) {
	orgID, err := middleware.GetOrgID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Missing active organisation context", err.Error())
		return
	}

	userIDParam := c.Param("user_id")
	targetUserID, err := uuid.Parse(userIDParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID format", err.Error())
		return
	}

	if err := h.service.RemoveMember(c.Request.Context(), orgID, targetUserID); err != nil {
		if errors.Is(err, ErrMemberNotFound) {
			utils.ErrorResponse(c, http.StatusNotFound, "Member not found", nil)
			return
		}
		if errors.Is(err, ErrCannotRemoveLastOwner) {
			utils.ErrorResponse(c, http.StatusForbidden, err.Error(), nil)
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to remove member", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Member removed from organisation", nil)
}
