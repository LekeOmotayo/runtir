package utils

import (
	"testing"
)

func TestPasswordHashingAndVerification(t *testing.T) {
	password := "SecretP@ssw0rd123!"

	hashed, err := HashPassword(password)
	if err != nil {
		t.Fatalf("expected no error hashing password, got: %v", err)
	}

	if hashed == password {
		t.Fatalf("expected hashed password to be different from plaintext")
	}

	if !CheckPasswordHash(password, hashed) {
		t.Fatalf("expected password verification to succeed with correct password")
	}

	if CheckPasswordHash("WrongPassword", hashed) {
		t.Fatalf("expected password verification to fail with incorrect password")
	}
}
