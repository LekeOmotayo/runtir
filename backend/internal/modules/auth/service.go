package auth

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/lekeOmotayo/runtir/internal/config"
	"github.com/lekeOmotayo/runtir/internal/db"
	"github.com/lekeOmotayo/runtir/internal/utils"
)

var (
	ErrUserAlreadyExists   = errors.New("a user with this email already exists")
	ErrInvalidCredentials   = errors.New("invalid email or password")
	ErrUserNotFound        = errors.New("user not found")
	ErrOrgNotFound         = errors.New("organisation not found")
	ErrNotMemberOfOrg      = errors.New("user is not a member of the specified organisation")
	ErrNoOrganisations     = errors.New("user does not belong to any organisation")
)

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	FullName string `json:"full_name" binding:"required"`
	OrgName  string `json:"org_name"`
}

type LoginRequest struct {
	Email    string     `json:"email" binding:"required,email"`
	Password string     `json:"password" binding:"required"`
	OrgID    *uuid.UUID `json:"org_id"`
}

type SwitchOrgRequest struct {
	OrgID uuid.UUID `json:"org_id" binding:"required"`
}

type UserDTO struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	FullName  string    `json:"full_name"`
	AvatarURL string    `json:"avatar_url,omitempty"`
}

type OrganisationDTO struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
	Slug string    `json:"slug"`
	Role string    `json:"role"`
}

type AuthResponse struct {
	Token         string            `json:"token"`
	User          UserDTO           `json:"user"`
	ActiveOrg     OrganisationDTO   `json:"active_organisation"`
	Organisations []OrganisationDTO `json:"organisations"`
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

// Register creates a new user, their default organization, and an owner membership
func (s *Service) Register(ctx context.Context, req RegisterRequest) (*AuthResponse, error) {
	// 1. Check if user already exists
	if _, err := s.queries.GetUserByEmail(ctx, req.Email); err == nil {
		return nil, ErrUserAlreadyExists
	}

	// 2. Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// 3. Create user record
	user, err := s.queries.CreateUser(ctx, db.CreateUserParams{
		Email:        strings.ToLower(strings.TrimSpace(req.Email)),
		PasswordHash: hashedPassword,
		FullName:     strings.TrimSpace(req.FullName),
		AvatarUrl:    pgtype.Text{Valid: false},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// 4. Determine organisation name & slug
	orgName := strings.TrimSpace(req.OrgName)
	if orgName == "" {
		orgName = fmt.Sprintf("%s's Org", req.FullName)
	}

	baseSlug := generateSlug(orgName)
	uniqueSlug := fmt.Sprintf("%s-%s", baseSlug, uuid.New().String()[:8])

	// 5. Create organisation
	org, err := s.queries.CreateOrganisation(ctx, db.CreateOrganisationParams{
		Name: orgName,
		Slug: uniqueSlug,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create organisation: %w", err)
	}

	// 6. Create owner membership
	membership, err := s.queries.CreateMembership(ctx, db.CreateMembershipParams{
		OrganisationID: org.ID,
		UserID:         user.ID,
		Role:           "owner",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create organisation membership: %w", err)
	}

	// 7. Generate JWT
	duration := time.Duration(s.cfg.JWTExpiryHours) * time.Hour
	token, err := utils.GenerateJWT(user.ID, org.ID, user.Email, membership.Role, s.cfg.JWTSecret, duration)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// 8. Log registration activity
	_, _ = s.queries.CreateActivityLog(ctx, db.CreateActivityLogParams{
		OrganisationID: org.ID,
		UserID:         pgtype.UUID{Bytes: user.ID, Valid: true},
		Action:         "user.registered",
		EntityType:     "user",
		EntityID:       pgtype.UUID{Bytes: user.ID, Valid: true},
		Metadata:       []byte(fmt.Sprintf(`{"email":"%s"}`, user.Email)),
	})

	activeOrg := OrganisationDTO{
		ID:   org.ID,
		Name: org.Name,
		Slug: org.Slug,
		Role: membership.Role,
	}

	return &AuthResponse{
		Token: token,
		User: UserDTO{
			ID:       user.ID,
			Email:    user.Email,
			FullName: user.FullName,
		},
		ActiveOrg:     activeOrg,
		Organisations: []OrganisationDTO{activeOrg},
	}, nil
}

// Login authenticates a user and returns a token scoped to their active organisation
func (s *Service) Login(ctx context.Context, req LoginRequest) (*AuthResponse, error) {
	// 1. Fetch user by email
	user, err := s.queries.GetUserByEmail(ctx, strings.ToLower(strings.TrimSpace(req.Email)))
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	// 2. Verify password hash
	if !utils.CheckPasswordHash(req.Password, user.PasswordHash) {
		return nil, ErrInvalidCredentials
	}

	// 3. Fetch user's organisations
	userOrgs, err := s.queries.ListUserOrganisations(ctx, user.ID)
	if err != nil || len(userOrgs) == 0 {
		return nil, ErrNoOrganisations
	}

	// 4. Select active organisation
	var selectedOrg *db.ListUserOrganisationsRow
	if req.OrgID != nil {
		for _, org := range userOrgs {
			if org.ID == *req.OrgID {
				selectedOrg = &org
				break
			}
		}
		if selectedOrg == nil {
			return nil, ErrNotMemberOfOrg
		}
	} else {
		selectedOrg = &userOrgs[0]
	}

	// 5. Generate JWT token
	duration := time.Duration(s.cfg.JWTExpiryHours) * time.Hour
	token, err := utils.GenerateJWT(user.ID, selectedOrg.ID, user.Email, selectedOrg.Role, s.cfg.JWTSecret, duration)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// 6. Log activity
	_, _ = s.queries.CreateActivityLog(ctx, db.CreateActivityLogParams{
		OrganisationID: selectedOrg.ID,
		UserID:         pgtype.UUID{Bytes: user.ID, Valid: true},
		Action:         "user.logged_in",
		EntityType:     "user",
		EntityID:       pgtype.UUID{Bytes: user.ID, Valid: true},
		Metadata:       []byte(`{"method":"password"}`),
	})

	var orgList []OrganisationDTO
	for _, o := range userOrgs {
		orgList = append(orgList, OrganisationDTO{
			ID:   o.ID,
			Name: o.Name,
			Slug: o.Slug,
			Role: o.Role,
		})
	}

	return &AuthResponse{
		Token: token,
		User: UserDTO{
			ID:       user.ID,
			Email:    user.Email,
			FullName: user.FullName,
		},
		ActiveOrg: OrganisationDTO{
			ID:   selectedOrg.ID,
			Name: selectedOrg.Name,
			Slug: selectedOrg.Slug,
			Role: selectedOrg.Role,
		},
		Organisations: orgList,
	}, nil
}

// SwitchOrg switches the active organisation context for an authenticated user and issues a fresh token
func (s *Service) SwitchOrg(ctx context.Context, userID, targetOrgID uuid.UUID) (*AuthResponse, error) {
	// 1. Verify membership
	membership, err := s.queries.GetMembership(ctx, db.GetMembershipParams{
		OrganisationID: targetOrgID,
		UserID:         userID,
	})
	if err != nil {
		return nil, ErrNotMemberOfOrg
	}

	// 2. Fetch user details
	user, err := s.queries.GetUserByID(ctx, userID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	// 3. Fetch organisation details
	org, err := s.queries.GetOrganisationByID(ctx, targetOrgID)
	if err != nil {
		return nil, ErrOrgNotFound
	}

	// 4. Fetch all user organisations
	userOrgs, err := s.queries.ListUserOrganisations(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list organisations: %w", err)
	}

	// 5. Generate new JWT
	duration := time.Duration(s.cfg.JWTExpiryHours) * time.Hour
	token, err := utils.GenerateJWT(user.ID, org.ID, user.Email, membership.Role, s.cfg.JWTSecret, duration)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	var orgList []OrganisationDTO
	for _, o := range userOrgs {
		orgList = append(orgList, OrganisationDTO{
			ID:   o.ID,
			Name: o.Name,
			Slug: o.Slug,
			Role: o.Role,
		})
	}

	return &AuthResponse{
		Token: token,
		User: UserDTO{
			ID:       user.ID,
			Email:    user.Email,
			FullName: user.FullName,
		},
		ActiveOrg: OrganisationDTO{
			ID:   org.ID,
			Name: org.Name,
			Slug: org.Slug,
			Role: membership.Role,
		},
		Organisations: orgList,
	}, nil
}

// GetMe returns current authenticated user details and memberships
func (s *Service) GetMe(ctx context.Context, userID, currentOrgID uuid.UUID) (*AuthResponse, error) {
	user, err := s.queries.GetUserByID(ctx, userID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	userOrgs, err := s.queries.ListUserOrganisations(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list user organisations: %w", err)
	}

	var activeOrg OrganisationDTO
	var orgList []OrganisationDTO
	for _, o := range userOrgs {
		dto := OrganisationDTO{
			ID:   o.ID,
			Name: o.Name,
			Slug: o.Slug,
			Role: o.Role,
		}
		orgList = append(orgList, dto)
		if o.ID == currentOrgID {
			activeOrg = dto
		}
	}

	if activeOrg.ID == uuid.Nil && len(orgList) > 0 {
		activeOrg = orgList[0]
	}

	return &AuthResponse{
		Token: "", // Token not re-issued for GetMe
		User: UserDTO{
			ID:       user.ID,
			Email:    user.Email,
			FullName: user.FullName,
		},
		ActiveOrg:     activeOrg,
		Organisations: orgList,
	}, nil
}

// RefreshToken validates an active session token, checks DB validity, and returns a fresh JWT
func (s *Service) RefreshToken(ctx context.Context, tokenStr string) (*AuthResponse, error) {
	claims, err := utils.ValidateJWT(tokenStr, s.cfg.JWTSecret)
	if err != nil {
		return nil, utils.ErrInvalidToken
	}

	user, err := s.queries.GetUserByID(ctx, claims.UserID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	membership, err := s.queries.GetMembership(ctx, db.GetMembershipParams{
		OrganisationID: claims.OrgID,
		UserID:         claims.UserID,
	})
	if err != nil {
		return nil, ErrNotMemberOfOrg
	}

	org, err := s.queries.GetOrganisationByID(ctx, claims.OrgID)
	if err != nil {
		return nil, ErrOrgNotFound
	}

	userOrgs, err := s.queries.ListUserOrganisations(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to list user organisations: %w", err)
	}

	duration := time.Duration(s.cfg.JWTExpiryHours) * time.Hour
	newToken, err := utils.GenerateJWT(user.ID, org.ID, user.Email, membership.Role, s.cfg.JWTSecret, duration)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refreshed token: %w", err)
	}

	var orgList []OrganisationDTO
	for _, o := range userOrgs {
		orgList = append(orgList, OrganisationDTO{
			ID:   o.ID,
			Name: o.Name,
			Slug: o.Slug,
			Role: o.Role,
		})
	}

	return &AuthResponse{
		Token: newToken,
		User: UserDTO{
			ID:       user.ID,
			Email:    user.Email,
			FullName: user.FullName,
		},
		ActiveOrg: OrganisationDTO{
			ID:   org.ID,
			Name: org.Name,
			Slug: org.Slug,
			Role: membership.Role,
		},
		Organisations: orgList,
	}, nil
}

var nonAlphaNumRegex = regexp.MustCompile(`[^a-z0-9]+`)

func generateSlug(name string) string {
	lower := strings.ToLower(strings.TrimSpace(name))
	slug := nonAlphaNumRegex.ReplaceAllString(lower, "-")
	return strings.Trim(slug, "-")
}
