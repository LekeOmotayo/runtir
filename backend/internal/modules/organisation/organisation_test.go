package organisation

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

func TestAddMemberRoleValidation(t *testing.T) {
	service := &Service{}
	orgID := uuid.New()

	tests := []struct {
		name        string
		role        string
		expectedErr error
	}{
		{
			name:        "Invalid role returns ErrInvalidRole",
			role:        "superadmin",
			expectedErr: ErrInvalidRole,
		},
		{
			name:        "Empty role returns ErrInvalidRole",
			role:        "",
			expectedErr: ErrInvalidRole,
		},
		{
			name:        "Arbitrary string returns ErrInvalidRole",
			role:        "guest",
			expectedErr: ErrInvalidRole,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.AddMember(context.Background(), orgID, AddMemberRequest{
				Email: "test@example.com",
				Role:  tt.role,
			})
			if err != tt.expectedErr {
				t.Errorf("expected error %v, got %v", tt.expectedErr, err)
			}
		})
	}
}

func TestUpdateMemberRoleValidation(t *testing.T) {
	service := &Service{}
	orgID := uuid.New()
	targetUserID := uuid.New()

	_, err := service.UpdateMemberRole(context.Background(), orgID, targetUserID, "invalid_role")
	if err != ErrInvalidRole {
		t.Errorf("expected ErrInvalidRole, got %v", err)
	}
}
