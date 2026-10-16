package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestRequireRole(t *testing.T) {
	tests := []struct {
		name           string
		userRole       string
		setRole        bool
		allowedRoles   []string
		expectedStatus int
	}{
		{
			name:           "Owner allowed for owner or admin",
			userRole:       "owner",
			setRole:        true,
			allowedRoles:   []string{"owner", "admin"},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Admin allowed for owner or admin",
			userRole:       "admin",
			setRole:        true,
			allowedRoles:   []string{"owner", "admin"},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Member forbidden for owner or admin",
			userRole:       "member",
			setRole:        true,
			allowedRoles:   []string{"owner", "admin"},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Missing role context returns unauthorized",
			setRole:        false,
			allowedRoles:   []string{"owner"},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			r.Use(func(c *gin.Context) {
				if tt.setRole {
					c.Set(CtxUserRoleKey, tt.userRole)
				}
				c.Next()
			})
			r.GET("/protected", RequireRole(tt.allowedRoles...), func(c *gin.Context) {
				c.Status(http.StatusOK)
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/protected", nil)
			r.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestContextHelpers(t *testing.T) {
	testUserID := uuid.New()
	testOrgID := uuid.New()
	testRole := "owner"

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Set(CtxUserIDKey, testUserID)
	c.Set(CtxOrgIDKey, testOrgID)
	c.Set(CtxUserRoleKey, testRole)

	uid, err := GetUserID(c)
	if err != nil || uid != testUserID {
		t.Fatalf("GetUserID failed: expected %v, got %v (err: %v)", testUserID, uid, err)
	}

	oid, err := GetOrgID(c)
	if err != nil || oid != testOrgID {
		t.Fatalf("GetOrgID failed: expected %v, got %v (err: %v)", testOrgID, oid, err)
	}

	role := GetUserRole(c)
	if role != testRole {
		t.Fatalf("GetUserRole failed: expected %s, got %s", testRole, role)
	}
}
