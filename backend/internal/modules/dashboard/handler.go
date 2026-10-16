package dashboard

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lekeOmotayo/runtir/internal/middleware"
	"github.com/lekeOmotayo/runtir/internal/utils"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes mounts dashboard endpoints
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, authMW gin.HandlerFunc) {
	dash := rg.Group("/dashboard", authMW)
	{
		dash.GET("/stats", h.GetStats)
	}
}

// GetStats returns document counts and recent uploads for the dashboard
func (h *Handler) GetStats(c *gin.Context) {
	orgID, err := middleware.GetOrgID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Missing organisation context", err.Error())
		return
	}

	stats, err := h.service.GetStats(c.Request.Context(), orgID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve dashboard stats", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Dashboard stats retrieved", stats)
}
