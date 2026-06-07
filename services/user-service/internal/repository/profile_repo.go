// Package repository provides data access for the user-service.
package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/salemshafik/pote/services/user-service/internal/model"
)

// Common repository errors.
var (
	ErrProfileNotFound    = errors.New("user profile not found")
	ErrEmailAlreadyExists = errors.New("email already exists")
)

// ProfileRepository handles database operations for user profiles.
type ProfileRepository struct {
	db *pgxpool.Pool
}

// NewProfileRepository creates a new ProfileRepository.
func NewProfileRepository(db *pgxpool.Pool) *ProfileRepository {
	return &ProfileRepository{db: db}
}

// Create inserts a new user profile.
func (r *ProfileRepository) Create(ctx context.Context, p *model.UserProfile) (*model.UserProfile, error) {
	query := `
		INSERT INTO user_profiles (id, email, display_name, avatar_url, bio)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, email, display_name, avatar_url, bio, status, created_at, updated_at`

	var created model.UserProfile
	err := r.db.QueryRow(ctx, query,
		p.ID, p.Email, p.DisplayName, p.AvatarURL, p.Bio,
	).Scan(
		&created.ID, &created.Email, &created.DisplayName,
		&created.AvatarURL, &created.Bio, &created.Status,
		&created.CreatedAt, &created.UpdatedAt,
	)
	if err != nil {
		if isDuplicateKeyError(err) {
			return nil, ErrEmailAlreadyExists
		}
		return nil, fmt.Errorf("inserting profile: %w", err)
	}

	return &created, nil
}

// GetByID retrieves a user profile by ID.
func (r *ProfileRepository) GetByID(ctx context.Context, id string) (*model.UserProfile, error) {
	query := `
		SELECT id, email, display_name, avatar_url, bio, status, created_at, updated_at
		FROM user_profiles
		WHERE id = $1`

	var p model.UserProfile
	err := r.db.QueryRow(ctx, query, id).Scan(
		&p.ID, &p.Email, &p.DisplayName,
		&p.AvatarURL, &p.Bio, &p.Status,
		&p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrProfileNotFound
		}
		return nil, fmt.Errorf("querying profile by id: %w", err)
	}

	return &p, nil
}

// Update updates a user profile's mutable fields.
func (r *ProfileRepository) Update(ctx context.Context, id string, req *model.UpdateProfileRequest) (*model.UserProfile, error) {
	query := `
		UPDATE user_profiles
		SET display_name = COALESCE($2, display_name),
			avatar_url   = COALESCE($3, avatar_url),
			bio          = COALESCE($4, bio),
			updated_at   = NOW()
		WHERE id = $1
		RETURNING id, email, display_name, avatar_url, bio, status, created_at, updated_at`

	var p model.UserProfile
	err := r.db.QueryRow(ctx, query,
		id, req.DisplayName, req.AvatarURL, req.Bio,
	).Scan(
		&p.ID, &p.Email, &p.DisplayName,
		&p.AvatarURL, &p.Bio, &p.Status,
		&p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrProfileNotFound
		}
		return nil, fmt.Errorf("updating profile: %w", err)
	}

	return &p, nil
}

// Search finds user profiles matching a query string by email or display name.
func (r *ProfileRepository) Search(ctx context.Context, query string, limit, offset int) ([]*model.UserProfile, int, error) {
	searchPattern := "%" + query + "%"

	countQuery := `
		SELECT COUNT(*)
		FROM user_profiles
		WHERE email ILIKE $1 OR display_name ILIKE $1`

	var total int
	if err := r.db.QueryRow(ctx, countQuery, searchPattern).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("counting search results: %w", err)
	}

	if total == 0 {
		return []*model.UserProfile{}, 0, nil
	}

	selectQuery := `
		SELECT id, email, display_name, avatar_url, bio, status, created_at, updated_at
		FROM user_profiles
		WHERE email ILIKE $1 OR display_name ILIKE $1
		ORDER BY display_name ASC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.Query(ctx, selectQuery, searchPattern, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("searching profiles: %w", err)
	}
	defer rows.Close()

	var profiles []*model.UserProfile
	for rows.Next() {
		var p model.UserProfile
		if err := rows.Scan(
			&p.ID, &p.Email, &p.DisplayName,
			&p.AvatarURL, &p.Bio, &p.Status,
			&p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scanning profile row: %w", err)
		}
		profiles = append(profiles, &p)
	}

	return profiles, total, nil
}

// Upsert creates or updates a user profile (used for syncing from auth-service).
func (r *ProfileRepository) Upsert(ctx context.Context, p *model.UserProfile) (*model.UserProfile, error) {
	query := `
		INSERT INTO user_profiles (id, email, display_name, avatar_url)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id) DO UPDATE SET
			email        = EXCLUDED.email,
			display_name = EXCLUDED.display_name,
			avatar_url   = EXCLUDED.avatar_url,
			updated_at   = NOW()
		RETURNING id, email, display_name, avatar_url, bio, status, created_at, updated_at`

	var upserted model.UserProfile
	err := r.db.QueryRow(ctx, query,
		p.ID, p.Email, p.DisplayName, p.AvatarURL,
	).Scan(
		&upserted.ID, &upserted.Email, &upserted.DisplayName,
		&upserted.AvatarURL, &upserted.Bio, &upserted.Status,
		&upserted.CreatedAt, &upserted.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("upserting profile: %w", err)
	}

	return &upserted, nil
}

// isDuplicateKeyError checks if a pgx error is a unique constraint violation (23505).
func isDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	for i := 0; i <= len(errStr)-5; i++ {
		if errStr[i:i+5] == "23505" {
			return true
		}
	}
	return false
}
