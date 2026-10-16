package organisation

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/lekeOmotayo/runtir/internal/config"
	"github.com/lekeOmotayo/runtir/internal/db"
	"github.com/lekeOmotayo/runtir/internal/utils"
)

var (
	ErrOrgNotFound          = errors.New("organisation not found")
	ErrMemberNotFound       = errors.New("member not found in organisation")
	ErrMemberAlreadyExists  = errors.New("user is already a member of this organisation")
	ErrCannotRemoveLastOwner = errors.New("cannot remove or demote the last owner of the organisation")
	ErrInvalidRole          = errors.New("invalid role. Allowed roles: owner, admin, member")
)

type OrganisationDetailDTO struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	CreatedAt string    `json:"created_at"`
	UpdatedAt string    `json:"updated_at"`
}

type MemberDTO struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Email     string    `json:"email"`
	FullName  string    `json:"full_name"`
	AvatarURL string    `json:"avatar_url,omitempty"`
	Role      string    `json:"role"`
	CreatedAt string    `json:"created_at"`
}

type AddMemberRequest struct {
	Email    string `json:"email" binding:"required,email"`
	FullName string `json:"full_name"`
	Role     string `json:"role" binding:"required,oneof=owner admin member"`
}

type UpdateMemberRoleRequest struct {
	Role string `json:"role" binding:"required,oneof=owner admin member"`
}

type Service struct {
	cfg     *config.Config
	queries *db.Queries
}

func NewService(cfg *config.Config, queries *db.Queries) *Service {
	return &Service{
		cfg:     cfg,
		queries: queries,
	}
}

// GetCurrent returns the active organisation details
func (s *Service) GetCurrent(ctx context.Context, orgID uuid.UUID) (*OrganisationDetailDTO, error) {
	org, err := s.queries.GetOrganisationByID(ctx, orgID)
	if err != nil {
		return nil, ErrOrgNotFound
	}

	createdAtStr := ""
	if org.CreatedAt.Valid {
		createdAtStr = org.CreatedAt.Time.Format(time.RFC3339)
	}

	updatedAtStr := ""
	if org.UpdatedAt.Valid {
		updatedAtStr = org.UpdatedAt.Time.Format(time.RFC3339)
	}

	return &OrganisationDetailDTO{
		ID:        org.ID,
		Name:      org.Name,
		Slug:      org.Slug,
		CreatedAt: createdAtStr,
		UpdatedAt: updatedAtStr,
	}, nil
}

// ListMembers lists all users belonging to the organisation
func (s *Service) ListMembers(ctx context.Context, orgID uuid.UUID) ([]MemberDTO, error) {
	members, err := s.queries.ListMembershipsByOrg(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to list memberships: %w", err)
	}

	dtos := make([]MemberDTO, 0, len(members))
	for _, m := range members {
		avatar := ""
		if m.AvatarUrl.Valid {
			avatar = m.AvatarUrl.String
		}

		createdAtStr := ""
		if m.CreatedAt.Valid {
			createdAtStr = m.CreatedAt.Time.Format(time.RFC3339)
		}

		dtos = append(dtos, MemberDTO{
			ID:        m.ID,
			UserID:    m.UserID,
			Email:     m.Email,
			FullName:  m.FullName,
			AvatarURL: avatar,
			Role:      m.Role,
			CreatedAt: createdAtStr,
		})
	}

	return dtos, nil
}

// AddMember adds or invites a user by email to the organisation with a specified role
func (s *Service) AddMember(ctx context.Context, orgID uuid.UUID, req AddMemberRequest) (*MemberDTO, error) {
	role := strings.ToLower(strings.TrimSpace(req.Role))
	if role != "owner" && role != "admin" && role != "member" {
		return nil, ErrInvalidRole
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))

	// 1. Check if user already exists
	user, err := s.queries.GetUserByEmail(ctx, email)
	if err != nil {
		// User does not exist, provision an account with random hash
		fullName := strings.TrimSpace(req.FullName)
		if fullName == "" {
			parts := strings.Split(email, "@")
			fullName = parts[0]
		}

		tempPassword := uuid.New().String()
		hashedPassword, hashErr := utils.HashPassword(tempPassword)
		if hashErr != nil {
			return nil, fmt.Errorf("failed to hash password for provisioned user: %w", hashErr)
		}

		newUser, createErr := s.queries.CreateUser(ctx, db.CreateUserParams{
			Email:        email,
			PasswordHash: hashedPassword,
			FullName:     fullName,
			AvatarUrl:    pgtype.Text{Valid: false},
		})
		if createErr != nil {
			return nil, fmt.Errorf("failed to create user: %w", createErr)
		}
		user = newUser
	}

	// 2. Check if user is already a member of this organisation
	_, err = s.queries.GetMembership(ctx, db.GetMembershipParams{
		OrganisationID: orgID,
		UserID:         user.ID,
	})
	if err == nil {
		return nil, ErrMemberAlreadyExists
	}

	// 3. Create membership
	membership, err := s.queries.CreateMembership(ctx, db.CreateMembershipParams{
		OrganisationID: orgID,
		UserID:         user.ID,
		Role:           role,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create membership: %w", err)
	}

	// 4. Log activity
	_, _ = s.queries.CreateActivityLog(ctx, db.CreateActivityLogParams{
		OrganisationID: orgID,
		UserID:         pgtype.UUID{Bytes: user.ID, Valid: true},
		Action:         "member.added",
		EntityType:     "membership",
		EntityID:       pgtype.UUID{Bytes: membership.ID, Valid: true},
		Metadata:       []byte(fmt.Sprintf(`{"email":"%s","role":"%s"}`, user.Email, role)),
	})

	createdAtStr := ""
	if membership.CreatedAt.Valid {
		createdAtStr = membership.CreatedAt.Time.Format(time.RFC3339)
	}

	return &MemberDTO{
		ID:        membership.ID,
		UserID:    user.ID,
		Email:     user.Email,
		FullName:  user.FullName,
		Role:      membership.Role,
		CreatedAt: createdAtStr,
	}, nil
}

// UpdateMemberRole updates the role of an existing organisation member
func (s *Service) UpdateMemberRole(ctx context.Context, orgID, targetUserID uuid.UUID, newRole string) (*MemberDTO, error) {
	newRole = strings.ToLower(strings.TrimSpace(newRole))
	if newRole != "owner" && newRole != "admin" && newRole != "member" {
		return nil, ErrInvalidRole
	}

	currentMembership, err := s.queries.GetMembership(ctx, db.GetMembershipParams{
		OrganisationID: orgID,
		UserID:         targetUserID,
	})
	if err != nil {
		return nil, ErrMemberNotFound
	}

	// Prevent demoting the last owner
	if currentMembership.Role == "owner" && newRole != "owner" {
		ownerCount, err := s.queries.CountOwnersByOrg(ctx, orgID)
		if err == nil && ownerCount <= 1 {
			return nil, ErrCannotRemoveLastOwner
		}
	}

	updated, err := s.queries.UpdateMembershipRole(ctx, db.UpdateMembershipRoleParams{
		OrganisationID: orgID,
		UserID:         targetUserID,
		Role:           newRole,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update membership role: %w", err)
	}

	user, err := s.queries.GetUserByID(ctx, targetUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to load user details: %w", err)
	}

	avatar := ""
	if user.AvatarUrl.Valid {
		avatar = user.AvatarUrl.String
	}

	createdAtStr := ""
	if updated.CreatedAt.Valid {
		createdAtStr = updated.CreatedAt.Time.Format(time.RFC3339)
	}

	return &MemberDTO{
		ID:        updated.ID,
		UserID:    user.ID,
		Email:     user.Email,
		FullName:  user.FullName,
		AvatarURL: avatar,
		Role:      updated.Role,
		CreatedAt: createdAtStr,
	}, nil
}

// RemoveMember removes a user from the organisation
func (s *Service) RemoveMember(ctx context.Context, orgID, targetUserID uuid.UUID) error {
	membership, err := s.queries.GetMembership(ctx, db.GetMembershipParams{
		OrganisationID: orgID,
		UserID:         targetUserID,
	})
	if err != nil {
		return ErrMemberNotFound
	}

	// Prevent removing the last owner
	if membership.Role == "owner" {
		ownerCount, err := s.queries.CountOwnersByOrg(ctx, orgID)
		if err == nil && ownerCount <= 1 {
			return ErrCannotRemoveLastOwner
		}
	}

	if err := s.queries.DeleteMembership(ctx, db.DeleteMembershipParams{
		OrganisationID: orgID,
		UserID:         targetUserID,
	}); err != nil {
		return fmt.Errorf("failed to delete membership: %w", err)
	}

	return nil
}
