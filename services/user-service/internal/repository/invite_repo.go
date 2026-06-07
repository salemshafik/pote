package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/salemshafik/pote/services/user-service/internal/model"
)

// Invite repository errors.
var (
	ErrInviteNotFound = errors.New("invite not found")
)

// InviteRepository handles database operations for email invites.
type InviteRepository struct {
	db *pgxpool.Pool
}

// NewInviteRepository creates a new InviteRepository.
func NewInviteRepository(db *pgxpool.Pool) *InviteRepository {
	return &InviteRepository{db: db}
}

// Create inserts a new invite record.
func (r *InviteRepository) Create(ctx context.Context, inviterID, email string) (*model.Invite, error) {
	query := `
		INSERT INTO invites (inviter_id, email)
		VALUES ($1, $2)
		RETURNING id, inviter_id, email, status, created_at, expires_at`

	var inv model.Invite
	err := r.db.QueryRow(ctx, query, inviterID, email).Scan(
		&inv.ID, &inv.InviterID, &inv.Email,
		&inv.Status, &inv.CreatedAt, &inv.ExpiresAt,
	)
	if err != nil {
		return nil, fmt.Errorf("inserting invite: %w", err)
	}

	return &inv, nil
}

// ListByInviter returns all invites sent by a user.
func (r *InviteRepository) ListByInviter(ctx context.Context, inviterID string) ([]*model.Invite, error) {
	query := `
		SELECT id, inviter_id, email, status, created_at, expires_at
		FROM invites
		WHERE inviter_id = $1
		ORDER BY created_at DESC`

	rows, err := r.db.Query(ctx, query, inviterID)
	if err != nil {
		return nil, fmt.Errorf("listing invites: %w", err)
	}
	defer rows.Close()

	var invites []*model.Invite
	for rows.Next() {
		var inv model.Invite
		if err := rows.Scan(
			&inv.ID, &inv.InviterID, &inv.Email,
			&inv.Status, &inv.CreatedAt, &inv.ExpiresAt,
		); err != nil {
			return nil, fmt.Errorf("scanning invite row: %w", err)
		}
		invites = append(invites, &inv)
	}

	if invites == nil {
		invites = []*model.Invite{}
	}

	return invites, nil
}

// GetPendingByEmail retrieves a pending invite by email.
func (r *InviteRepository) GetPendingByEmail(ctx context.Context, email string) (*model.Invite, error) {
	query := `
		SELECT id, inviter_id, email, status, created_at, expires_at
		FROM invites
		WHERE email = $1 AND status = 'pending' AND expires_at > NOW()
		ORDER BY created_at DESC
		LIMIT 1`

	var inv model.Invite
	err := r.db.QueryRow(ctx, query, email).Scan(
		&inv.ID, &inv.InviterID, &inv.Email,
		&inv.Status, &inv.CreatedAt, &inv.ExpiresAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrInviteNotFound
		}
		return nil, fmt.Errorf("querying pending invite: %w", err)
	}

	return &inv, nil
}

// MarkAccepted marks an invite as accepted.
func (r *InviteRepository) MarkAccepted(ctx context.Context, id string) error {
	query := `UPDATE invites SET status = 'accepted' WHERE id = $1`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("marking invite accepted: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrInviteNotFound
	}

	return nil
}
