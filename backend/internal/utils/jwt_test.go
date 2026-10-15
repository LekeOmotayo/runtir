package utils

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestJWTGenerationAndValidation(t *testing.T) {
	userID := uuid.New()
	orgID := uuid.New()
	email := "alex@afriops.com"
	role := "owner"
	secret := "test-super-secret-key-12345"

	// 1. Valid token
	token, err := GenerateJWT(userID, orgID, email, role, secret, 1*time.Hour)
	if err != nil {
		t.Fatalf("expected no error generating JWT, got: %v", err)
	}

	claims, err := ValidateJWT(token, secret)
	if err != nil {
		t.Fatalf("expected valid token, got error: %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("expected userID %s, got %s", userID, claims.UserID)
	}
	if claims.OrgID != orgID {
		t.Errorf("expected orgID %s, got %s", orgID, claims.OrgID)
	}
	if claims.Email != email {
		t.Errorf("expected email %s, got %s", email, claims.Email)
	}
	if claims.Role != role {
		t.Errorf("expected role %s, got %s", role, claims.Role)
	}

	// 2. Invalid secret
	_, err = ValidateJWT(token, "wrong-secret-key")
	if err == nil {
		t.Fatalf("expected error validating token with wrong secret, got nil")
	}

	// 3. Expired token
	expiredToken, err := GenerateJWT(userID, orgID, email, role, secret, -1*time.Minute)
	if err != nil {
		t.Fatalf("expected no error generating expired JWT, got: %v", err)
	}

	_, err = ValidateJWT(expiredToken, secret)
	if err == nil {
		t.Fatalf("expected error validating expired token, got nil")
	}
}
