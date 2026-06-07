package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/salemshafik/pote/services/user-service/internal/model"
)

// ProfileRepository handles database operations for user profiles.
type ProfileRepository struct {
	db *pgxpool.Pool
}

// NewProfileRepository creates a new ProfileRepository.
func NewProfileRepository(db *pgxpool.Pool) *ProfileRepository {
	return &ProfileRepository{db: db}
}

// Create inserts a new user profile and returns the created record.
// The ID is supplied by the caller (it mirrors the auth-service users.id).
func (r *ProfileRepository) Create(ctx context.Context, p *model.UserProfile) (*model.UserProfile, error) {
	query := `
		INSERT INTO user_profiles (id, email, display_name, avatar_url, bio, status)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, email, display_name, avatar_url, bio, status, created_at, updated_at`

	var created model.UserProfile
	err := r.db.QueryRow(ctx, query,
		p.ID,
		p.Email,
		p.DisplayName,
		p.AvatarURL,
		p.Bio,
		p.Status,
	).Scan(
		&created.ID,
		&created.Email,
		&created.DisplayName,
		&created.AvatarURL,
		&created.Bio,
		&created.Status,
		&created.CreatedAt,
		&created.UpdatedAt,
	)

	if err != nil {
		if isUniqueViolation(err) {
			return nil, ErrProfileExists
		}
		return nil, fmt.Errorf("inserting user profile: %w", err)
	}

	return &created, nil
}

// GetByID retrieves a user profile by its UUID.
func (r *ProfileRepository) GetByID(ctx context.Context, id string) (*model.UserProfile, error) {
	query := `
		SELECT id, email, display_name, avatar_url, bio, status, created_at, updated_at
		FROM user_profiles
		WHERE id = $1`

	return r.queryOne(ctx, query, id)
}

// GetByEmail retrieves a user profile by email address.
func (r *ProfileRepository) GetByEmail(ctx context.Context, email string) (*model.UserProfile, error) {
	query := `
		SELECT id, email, display_name, avatar_url, bio, status, created_at, updated_at
		FROM user_profiles
		WHERE email = $1`

	return r.queryOne(ctx, query, email)
}

// Update applies the provided field changes to a profile and returns the
// updated record. Only non-nil fields in the request are modified; COALESCE
// keeps existing values when a parameter is NULL.
func (r *ProfileRepository) Update(ctx context.Context, id string, req *model.UpdateProfileRequest) (*model.UserProfile, error) {
	query := `
		UPDATE user_profiles
		SET display_name = COALESCE($2, display_name),
		    avatar_url   = COALESCE($3, avatar_url),
		    bio          = COALESCE($4, bio),
		    updated_at   = NOW()
		WHERE id = $1
		RETURNING id, email, display_name, avatar_url, bio, status, created_at, updated_at`

	return r.queryOne(ctx, query, id, req.DisplayName, req.AvatarURL, req.Bio)
}

// UpdateStatus updates a user's presence status and returns the updated record.
func (r *ProfileRepository) UpdateStatus(ctx context.Context, id, status string) (*model.UserProfile, error) {
	query := `
		UPDATE user_profiles
		SET status = $2, updated_at = NOW()
		WHERE id = $1
		RETURNING id, email, display_name, avatar_url, bio, status, created_at, updated_at`

	return r.queryOne(ctx, query, id, status)
}

// queryOne runs a query expected to return a single profile row.
func (r *ProfileRepository) queryOne(ctx context.Context, query string, args ...any) (*model.UserProfile, error) {
	var p model.UserProfile
	err := r.db.QueryRow(ctx, query, args...).Scan(
		&p.ID,
		&p.Email,
		&p.DisplayName,
		&p.AvatarURL,
		&p.Bio,
		&p.Status,
		&p.CreatedAt,
		&p.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrProfileNotFound
		}
		return nil, fmt.Errorf("querying user profile: %w", err)
	}

	return &p, nil
}
