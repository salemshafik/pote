package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/salemshafik/pote/services/chat-service/internal/model"
)

// MemberRepository handles database operations for chat members.
type MemberRepository struct {
	db *pgxpool.Pool
}

// NewMemberRepository creates a new MemberRepository.
func NewMemberRepository(db *pgxpool.Pool) *MemberRepository {
	return &MemberRepository{db: db}
}

// Add inserts a new member into a chat and returns the created record.
func (r *MemberRepository) Add(ctx context.Context, m *model.ChatMember) (*model.ChatMember, error) {
	query := `
		INSERT INTO chat_members (chat_id, user_id, role)
		VALUES ($1, $2, $3)
		RETURNING id, chat_id, user_id, role, joined_at`

	var created model.ChatMember
	err := r.db.QueryRow(ctx, query, m.ChatID, m.UserID, m.Role).Scan(
		&created.ID,
		&created.ChatID,
		&created.UserID,
		&created.Role,
		&created.JoinedAt,
	)

	if err != nil {
		switch {
		case isUniqueViolation(err):
			return nil, ErrMemberExists
		case isForeignKeyViolation(err):
			return nil, ErrChatNotFound
		default:
			return nil, fmt.Errorf("inserting chat member: %w", err)
		}
	}

	return &created, nil
}

// ListByChat returns all members of a chat, owners and admins first.
func (r *MemberRepository) ListByChat(ctx context.Context, chatID string) ([]model.ChatMember, error) {
	query := `
		SELECT id, chat_id, user_id, role, joined_at
		FROM chat_members
		WHERE chat_id = $1
		ORDER BY
			CASE role
				WHEN 'owner' THEN 0
				WHEN 'admin' THEN 1
				ELSE 2
			END,
			joined_at ASC`

	rows, err := r.db.Query(ctx, query, chatID)
	if err != nil {
		return nil, fmt.Errorf("querying chat members: %w", err)
	}
	defer rows.Close()

	members := make([]model.ChatMember, 0)
	for rows.Next() {
		var m model.ChatMember
		if err := rows.Scan(&m.ID, &m.ChatID, &m.UserID, &m.Role, &m.JoinedAt); err != nil {
			return nil, fmt.Errorf("scanning member row: %w", err)
		}
		members = append(members, m)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating member rows: %w", err)
	}

	return members, nil
}

// GetRole returns the role of a user within a chat, or ErrMemberNotFound if the
// user is not a member.
func (r *MemberRepository) GetRole(ctx context.Context, chatID, userID string) (string, error) {
	query := `SELECT role FROM chat_members WHERE chat_id = $1 AND user_id = $2`

	var role string
	err := r.db.QueryRow(ctx, query, chatID, userID).Scan(&role)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrMemberNotFound
		}
		return "", fmt.Errorf("querying member role: %w", err)
	}

	return role, nil
}

// UpdateRole changes a member's role within a chat and returns the updated record.
func (r *MemberRepository) UpdateRole(ctx context.Context, chatID, userID, role string) (*model.ChatMember, error) {
	query := `
		UPDATE chat_members
		SET role = $3
		WHERE chat_id = $1 AND user_id = $2
		RETURNING id, chat_id, user_id, role, joined_at`

	var m model.ChatMember
	err := r.db.QueryRow(ctx, query, chatID, userID, role).Scan(
		&m.ID,
		&m.ChatID,
		&m.UserID,
		&m.Role,
		&m.JoinedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrMemberNotFound
		}
		if isCheckViolation(err) {
			return nil, fmt.Errorf("invalid role %q: %w", role, err)
		}
		return nil, fmt.Errorf("updating member role: %w", err)
	}

	return &m, nil
}

// Remove deletes a member from a chat. Returns ErrMemberNotFound if the user is
// not a member.
func (r *MemberRepository) Remove(ctx context.Context, chatID, userID string) error {
	query := `DELETE FROM chat_members WHERE chat_id = $1 AND user_id = $2`

	result, err := r.db.Exec(ctx, query, chatID, userID)
	if err != nil {
		return fmt.Errorf("removing chat member: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrMemberNotFound
	}

	return nil
}

// CountByRole returns how many members in a chat hold the given role.
func (r *MemberRepository) CountByRole(ctx context.Context, chatID, role string) (int, error) {
	query := `SELECT COUNT(*) FROM chat_members WHERE chat_id = $1 AND role = $2`

	var count int
	if err := r.db.QueryRow(ctx, query, chatID, role).Scan(&count); err != nil {
		return 0, fmt.Errorf("counting members by role: %w", err)
	}

	return count, nil
}
