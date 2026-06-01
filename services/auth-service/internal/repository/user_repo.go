// Package repository provides data access for the auth-service.
package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/salemshafik/pote/services/auth-service/internal/model"
)

// Common repository errors.
var (
	ErrUserNotFound      = errors.New("user not found")
	ErrEmailAlreadyExists = errors.New("email already exists")
)

// UserRepository handles database operations for users.
type UserRepository struct {
	db *pgxpool.Pool
}

// NewUserRepository creates a new UserRepository.
func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

// Create inserts a new user into the database and returns the created user.
func (r *UserRepository) Create(ctx context.Context, user *model.User) (*model.User, error) {
	query := `
		INSERT INTO users (email, display_name, password_hash, avatar_url, provider, provider_id, role)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, email, display_name, password_hash, avatar_url, provider, provider_id, role, created_at, updated_at`

	var created model.User
	err := r.db.QueryRow(ctx, query,
		user.Email,
		user.DisplayName,
		user.PasswordHash,
		user.AvatarURL,
		user.Provider,
		user.ProviderID,
		user.Role,
	).Scan(
		&created.ID,
		&created.Email,
		&created.DisplayName,
		&created.PasswordHash,
		&created.AvatarURL,
		&created.Provider,
		&created.ProviderID,
		&created.Role,
		&created.CreatedAt,
		&created.UpdatedAt,
	)

	if err != nil {
		if isDuplicateKeyError(err) {
			return nil, ErrEmailAlreadyExists
		}
		return nil, fmt.Errorf("inserting user: %w", err)
	}

	return &created, nil
}

// GetByEmail retrieves a user by email address.
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	query := `
		SELECT id, email, display_name, password_hash, avatar_url, provider, provider_id, role, created_at, updated_at
		FROM users
		WHERE email = $1`

	var user model.User
	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.DisplayName,
		&user.PasswordHash,
		&user.AvatarURL,
		&user.Provider,
		&user.ProviderID,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("querying user by email: %w", err)
	}

	return &user, nil
}

// GetByID retrieves a user by their UUID.
func (r *UserRepository) GetByID(ctx context.Context, id string) (*model.User, error) {
	query := `
		SELECT id, email, display_name, password_hash, avatar_url, provider, provider_id, role, created_at, updated_at
		FROM users
		WHERE id = $1`

	var user model.User
	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.DisplayName,
		&user.PasswordHash,
		&user.AvatarURL,
		&user.Provider,
		&user.ProviderID,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("querying user by id: %w", err)
	}

	return &user, nil
}

// GetByProviderID retrieves a user by OAuth provider and provider-specific ID.
func (r *UserRepository) GetByProviderID(ctx context.Context, provider, providerID string) (*model.User, error) {
	query := `
		SELECT id, email, display_name, password_hash, avatar_url, provider, provider_id, role, created_at, updated_at
		FROM users
		WHERE provider = $1 AND provider_id = $2`

	var user model.User
	err := r.db.QueryRow(ctx, query, provider, providerID).Scan(
		&user.ID,
		&user.Email,
		&user.DisplayName,
		&user.PasswordHash,
		&user.AvatarURL,
		&user.Provider,
		&user.ProviderID,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("querying user by provider: %w", err)
	}

	return &user, nil
}

// isDuplicateKeyError checks if a pgx error is a unique constraint violation.
func isDuplicateKeyError(err error) bool {
	// pgx wraps PostgreSQL error codes; unique_violation = 23505
	return err != nil && contains(err.Error(), "23505")
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
