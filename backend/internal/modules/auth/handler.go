package auth

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lekeOmotayo/runtir/internal/config"
	"github.com/lekeOmotayo/runtir/internal/middleware"
	"github.com/lekeOmotayo/runtir/internal/utils"
)

type Handler struct {
	cfg     *config.Config
	service *Service
}

func NewHandler(cfg *config.Config, service *Service) *Handler {
	return &Handler{
		cfg:     cfg,
		service: service,
	}
}

// RegisterRoutes registers the authentication routes on the given router group
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, authMW gin.HandlerFunc) {
	authGroup := rg.Group("/auth")
	{
		authGroup.POST("/register", h.Register)
		authGroup.POST("/login", h.Login)
		authGroup.POST("/logout", h.Logout)
		authGroup.POST("/refresh", h.RefreshToken)

		// Protected endpoints
		authGroup.GET("/me", authMW, h.GetMe)
		authGroup.POST("/switch-org", authMW, h.SwitchOrg)
	}
}

// Register handles new user and organisation registration
func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid registration input", err.Error())
		return
	}

	resp, err := h.service.Register(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, ErrUserAlreadyExists) {
			utils.ErrorResponse(c, http.StatusConflict, err.Error(), nil)
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Registration failed", err.Error())
		return
	}

	// Set HTTP-only secure cookie
	h.setAuthCookie(c, resp.Token)

	utils.SuccessResponse(c, http.StatusCreated, "User registered successfully", resp)
}

// Login handles user authentication and sets an HTTP-only cookie
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid login credentials format", err.Error())
		return
	}

	resp, err := h.service.Login(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) || errors.Is(err, ErrNotMemberOfOrg) {
			utils.ErrorResponse(c, http.StatusUnauthorized, err.Error(), nil)
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Login failed", err.Error())
		return
	}

	// Set HTTP-only secure cookie
	h.setAuthCookie(c, resp.Token)

	utils.SuccessResponse(c, http.StatusOK, "Login successful", resp)
}

// Logout clears the HTTP-only auth cookie
func (h *Handler) Logout(c *gin.Context) {
	h.clearAuthCookie(c)
	utils.SuccessResponse(c, http.StatusOK, "Logged out successfully", nil)
}

// RefreshToken re-issues an access token for an active session and updates the cookie
func (h *Handler) RefreshToken(c *gin.Context) {
	tokenStr := ""

	// 1. Try reading from cookie first
	if cookieToken, err := c.Cookie(h.cfg.CookieName); err == nil && cookieToken != "" {
		tokenStr = cookieToken
	}

	// 2. Fall back to Authorization: Bearer <token>
	if tokenStr == "" {
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
			tokenStr = strings.TrimPrefix(authHeader, "Bearer ")
		}
	}

	if tokenStr == "" {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Missing session token", "no token found in cookie or authorization header")
		return
	}

	resp, err := h.service.RefreshToken(c.Request.Context(), tokenStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Session refresh failed", err.Error())
		return
	}

	// Update HTTP-only cookie with the fresh token
	h.setAuthCookie(c, resp.Token)

	utils.SuccessResponse(c, http.StatusOK, "Session refreshed successfully", resp)
}

// GetMe returns current authenticated user and organisation profile
func (h *Handler) GetMe(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", err.Error())
		return
	}

	orgID, _ := middleware.GetOrgID(c)

	resp, err := h.service.GetMe(c.Request.Context(), userID, orgID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to load user profile", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Profile retrieved successfully", resp)
}

// SwitchOrg switches active organisation and updates the auth cookie
func (h *Handler) SwitchOrg(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", err.Error())
		return
	}

	var req SwitchOrgRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	resp, err := h.service.SwitchOrg(c.Request.Context(), userID, req.OrgID)
	if err != nil {
		if errors.Is(err, ErrNotMemberOfOrg) {
			utils.ErrorResponse(c, http.StatusForbidden, err.Error(), nil)
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to switch organisation", err.Error())
		return
	}

	// Update HTTP-only cookie with the new token scoped to the target organisation
	h.setAuthCookie(c, resp.Token)

	utils.SuccessResponse(c, http.StatusOK, "Switched organisation successfully", resp)
}

func (h *Handler) setAuthCookie(c *gin.Context, token string) {
	maxAge := h.cfg.JWTExpiryHours * 3600
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(
		h.cfg.CookieName,
		token,
		maxAge,
		"/",
		h.cfg.CookieDomain,
		h.cfg.CookieSecure,
		true, // HttpOnly
	)
}

func (h *Handler) clearAuthCookie(c *gin.Context) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(
		h.cfg.CookieName,
		"",
		-1, // MaxAge < 0 deletes cookie
		"/",
		h.cfg.CookieDomain,
		h.cfg.CookieSecure,
		true, // HttpOnly
	)
}
