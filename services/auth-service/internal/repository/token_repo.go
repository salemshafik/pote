package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/salemshafik/pote/services/auth-service/internal/model"
)

// Common token repository errors.
var (
	ErrTokenNotFound = errors.New("refresh token not found")
)

// TokenRepository handles database operations for refresh tokens.
type TokenRepository struct {
	db *pgxpool.Pool
}

// NewTokenRepository creates a new TokenRepository.
func NewTokenRepository(db *pgxpool.Pool) *TokenRepository {
	return &TokenRepository{db: db}
}

// Create inserts a new refresh token record.
func (r *TokenRepository) Create(ctx context.Context, token *model.RefreshToken) error {
	query := `
		INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)`

	_, err := r.db.Exec(ctx, query, token.UserID, token.TokenHash, token.ExpiresAt)
	if err != nil {
		return fmt.Errorf("inserting refresh token: %w", err)
	}

	return nil
}

// GetByTokenHash retrieves a refresh token by its hash.
func (r *TokenRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*model.RefreshToken, error) {
	query := `
		SELECT id, user_id, token_hash, expires_at, created_at, revoked
		FROM refresh_tokens
		WHERE token_hash = $1`

	var token model.RefreshToken
	err := r.db.QueryRow(ctx, query, tokenHash).Scan(
		&token.ID,
		&token.UserID,
		&token.TokenHash,
		&token.ExpiresAt,
		&token.CreatedAt,
		&token.Revoked,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrTokenNotFound
		}
		return nil, fmt.Errorf("querying refresh token: %w", err)
	}

	return &token, nil
}

// RevokeByTokenHash marks a refresh token as revoked.
func (r *TokenRepository) RevokeByTokenHash(ctx context.Context, tokenHash string) error {
	query := `UPDATE refresh_tokens SET revoked = TRUE WHERE token_hash = $1`

	result, err := r.db.Exec(ctx, query, tokenHash)
	if err != nil {
		return fmt.Errorf("revoking refresh token: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrTokenNotFound
	}

	return nil
}

// RevokeAllByUserID revokes all refresh tokens for a user.
func (r *TokenRepository) RevokeAllByUserID(ctx context.Context, userID string) error {
	query := `UPDATE refresh_tokens SET revoked = TRUE WHERE user_id = $1 AND revoked = FALSE`

	_, err := r.db.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("revoking all refresh tokens: %w", err)
	}

	return nil
}

// DeleteExpired removes expired refresh tokens (cleanup job).
func (r *TokenRepository) DeleteExpired(ctx context.Context) (int64, error) {
	query := `DELETE FROM refresh_tokens WHERE expires_at < NOW() OR revoked = TRUE`

	result, err := r.db.Exec(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("deleting expired tokens: %w", err)
	}

	return result.RowsAffected(), nil
}
