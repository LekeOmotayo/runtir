package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lekeOmotayo/runtir/internal/config"
	"github.com/lekeOmotayo/runtir/internal/db"
	"github.com/lekeOmotayo/runtir/internal/utils"
)

const (
	CtxUserIDKey    = "user_id"
	CtxOrgIDKey     = "org_id"
	CtxUserEmailKey = "user_email"
	CtxUserRoleKey  = "user_role"
)

// AuthMiddleware validates JWT from HTTP-only Cookie or Bearer header, and verifies tenant membership
func AuthMiddleware(cfg *config.Config, queries *db.Queries) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := ""

		// 1. Try reading from HTTP-only Cookie first
		if cookieToken, err := c.Cookie(cfg.CookieName); err == nil && cookieToken != "" {
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
			utils.ErrorResponse(c, http.StatusUnauthorized, "Authentication required", "missing authorization cookie or header")
			c.Abort()
			return
		}

		// 3. Validate JWT signature and expiration
		claims, err := utils.ValidateJWT(tokenStr, cfg.JWTSecret)
		if err != nil {
			utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid or expired session", err.Error())
			c.Abort()
			return
		}

		// 4. Verify tenant membership in database
		membership, err := queries.GetMembership(c.Request.Context(), db.GetMembershipParams{
			OrganisationID: claims.OrgID,
			UserID:         claims.UserID,
		})
		if err != nil {
			utils.ErrorResponse(c, http.StatusForbidden, "Access denied: not a member of this organisation", "invalid organisation membership")
			c.Abort()
			return
		}

		// 5. Store validated identity and tenant context in request context
		c.Set(CtxUserIDKey, claims.UserID)
		c.Set(CtxOrgIDKey, claims.OrgID)
		c.Set(CtxUserEmailKey, claims.Email)
		c.Set(CtxUserRoleKey, membership.Role)

		c.Next()
	}
}

// RequireRole ensures that the authenticated user possesses one of the required roles
func RequireRole(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleVal, exists := c.Get(CtxUserRoleKey)
		if !exists {
			utils.ErrorResponse(c, http.StatusUnauthorized, "Authentication required", "missing user role")
			c.Abort()
			return
		}

		userRole, ok := roleVal.(string)
		if !ok {
			utils.ErrorResponse(c, http.StatusInternalServerError, "Internal context error", "invalid role type")
			c.Abort()
			return
		}

		for _, r := range allowedRoles {
			if strings.EqualFold(r, userRole) {
				c.Next()
				return
			}
		}

		utils.ErrorResponse(c, http.StatusForbidden, "Forbidden: insufficient permissions", "role unauthorized")
		c.Abort()
	}
}

// GetUserID retrieves the authenticated UserID from the gin context
func GetUserID(c *gin.Context) (uuid.UUID, error) {
	val, exists := c.Get(CtxUserIDKey)
	if !exists {
		return uuid.Nil, errors.New("user ID not found in context")
	}
	id, ok := val.(uuid.UUID)
	if !ok {
		return uuid.Nil, errors.New("invalid user ID type in context")
	}
	return id, nil
}

// GetOrgID retrieves the active OrganisationID from the gin context
func GetOrgID(c *gin.Context) (uuid.UUID, error) {
	val, exists := c.Get(CtxOrgIDKey)
	if !exists {
		return uuid.Nil, errors.New("organisation ID not found in context")
	}
	id, ok := val.(uuid.UUID)
	if !ok {
		return uuid.Nil, errors.New("invalid organisation ID type in context")
	}
	return id, nil
}

// GetUserRole retrieves the active user's role in the organisation
func GetUserRole(c *gin.Context) string {
	if val, exists := c.Get(CtxUserRoleKey); exists {
		if role, ok := val.(string); ok {
			return role
		}
	}
	return ""
}
