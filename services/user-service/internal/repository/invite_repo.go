package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/salemshafik/pote/services/user-service/internal/model"
)

// InviteRepository handles database operations for email invites.
type InviteRepository struct {
	db *pgxpool.Pool
}

// NewInviteRepository creates a new InviteRepository.
func NewInviteRepository(db *pgxpool.Pool) *InviteRepository {
	return &InviteRepository{db: db}
}

// Create inserts a new invite and returns the created record. The expires_at
// column defaults to NOW() + 7 days at the database level.
func (r *InviteRepository) Create(ctx context.Context, inv *model.Invite) (*model.Invite, error) {
	query := `
		INSERT INTO invites (inviter_id, email, status)
		VALUES ($1, $2, $3)
		RETURNING id, inviter_id, email, status, created_at, expires_at`

	var created model.Invite
	err := r.db.QueryRow(ctx, query, inv.InviterID, inv.Email, inv.Status).Scan(
		&created.ID,
		&created.InviterID,
		&created.Email,
		&created.Status,
		&created.CreatedAt,
		&created.ExpiresAt,
	)

	if err != nil {
		if isForeignKeyViolation(err) {
			return nil, ErrProfileNotFound
		}
		return nil, fmt.Errorf("inserting invite: %w", err)
	}

	return &created, nil
}

// ExistsPending reports whether the inviter already has a pending invite for
// the given email. Used to avoid sending duplicate invitations.
func (r *InviteRepository) ExistsPending(ctx context.Context, inviterID, email string) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT 1 FROM invites
			WHERE inviter_id = $1 AND email = $2 AND status = $3
		)`

	var exists bool
	err := r.db.QueryRow(ctx, query, inviterID, email, model.InviteStatusPending).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("checking pending invite: %w", err)
	}

	return exists, nil
}

// ListByInviter returns all invites sent by the given user, newest first.
func (r *InviteRepository) ListByInviter(ctx context.Context, inviterID string) ([]model.Invite, error) {
	query := `
		SELECT id, inviter_id, email, status, created_at, expires_at
		FROM invites
		WHERE inviter_id = $1
		ORDER BY created_at DESC`

	rows, err := r.db.Query(ctx, query, inviterID)
	if err != nil {
		return nil, fmt.Errorf("querying invites: %w", err)
	}
	defer rows.Close()

	invites := make([]model.Invite, 0)
	for rows.Next() {
		var inv model.Invite
		if err := rows.Scan(
			&inv.ID,
			&inv.InviterID,
			&inv.Email,
			&inv.Status,
			&inv.CreatedAt,
			&inv.ExpiresAt,
		); err != nil {
			return nil, fmt.Errorf("scanning invite row: %w", err)
		}
		invites = append(invites, inv)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating invite rows: %w", err)
	}

	return invites, nil
}

// GetByID retrieves a single invite by its ID.
func (r *InviteRepository) GetByID(ctx context.Context, id string) (*model.Invite, error) {
	query := `
		SELECT id, inviter_id, email, status, created_at, expires_at
		FROM invites
		WHERE id = $1`

	var inv model.Invite
	err := r.db.QueryRow(ctx, query, id).Scan(
		&inv.ID,
		&inv.InviterID,
		&inv.Email,
		&inv.Status,
		&inv.CreatedAt,
		&inv.ExpiresAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrInviteNotFound
		}
		return nil, fmt.Errorf("querying invite: %w", err)
	}

	return &inv, nil
}
