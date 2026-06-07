package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/salemshafik/pote/services/chat-service/internal/model"
)

// ChatRepository handles database operations for chats.
type ChatRepository struct {
	db *pgxpool.Pool
}

// NewChatRepository creates a new ChatRepository.
func NewChatRepository(db *pgxpool.Pool) *ChatRepository {
	return &ChatRepository{db: db}
}

// CreateWithMembers inserts a chat together with its initial members inside a
// single transaction, so a chat is never persisted without its creator/members.
// The provided members slice must already include the creator (typically as
// owner). The created chat is returned.
func (r *ChatRepository) CreateWithMembers(ctx context.Context, chat *model.Chat, members []model.ChatMember) (*model.Chat, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("beginning transaction: %w", err)
	}
	// Rollback is a no-op if the transaction has already been committed.
	defer tx.Rollback(ctx)

	chatQuery := `
		INSERT INTO chats (type, name, created_by)
		VALUES ($1, $2, $3)
		RETURNING id, type, name, created_by, created_at, updated_at`

	var created model.Chat
	err = tx.QueryRow(ctx, chatQuery, chat.Type, chat.Name, chat.CreatedBy).Scan(
		&created.ID,
		&created.Type,
		&created.Name,
		&created.CreatedBy,
		&created.CreatedAt,
		&created.UpdatedAt,
	)
	if err != nil {
		if isCheckViolation(err) {
			return nil, fmt.Errorf("invalid chat type %q: %w", chat.Type, err)
		}
		return nil, fmt.Errorf("inserting chat: %w", err)
	}

	memberQuery := `
		INSERT INTO chat_members (chat_id, user_id, role)
		VALUES ($1, $2, $3)`

	for _, m := range members {
		if _, err := tx.Exec(ctx, memberQuery, created.ID, m.UserID, m.Role); err != nil {
			if isUniqueViolation(err) {
				return nil, ErrMemberExists
			}
			return nil, fmt.Errorf("inserting chat member %s: %w", m.UserID, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("committing transaction: %w", err)
	}

	return &created, nil
}

// GetByID retrieves a chat by its ID.
func (r *ChatRepository) GetByID(ctx context.Context, id string) (*model.Chat, error) {
	query := `
		SELECT id, type, name, created_by, created_at, updated_at
		FROM chats
		WHERE id = $1`

	var c model.Chat
	err := r.db.QueryRow(ctx, query, id).Scan(
		&c.ID,
		&c.Type,
		&c.Name,
		&c.CreatedBy,
		&c.CreatedAt,
		&c.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrChatNotFound
		}
		return nil, fmt.Errorf("querying chat: %w", err)
	}

	return &c, nil
}

// ListByUser returns all chats the given user is a member of, most recently
// updated first.
func (r *ChatRepository) ListByUser(ctx context.Context, userID string) ([]model.Chat, error) {
	query := `
		SELECT c.id, c.type, c.name, c.created_by, c.created_at, c.updated_at
		FROM chats c
		INNER JOIN chat_members m ON m.chat_id = c.id
		WHERE m.user_id = $1
		ORDER BY c.updated_at DESC`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("querying chats: %w", err)
	}
	defer rows.Close()

	chats := make([]model.Chat, 0)
	for rows.Next() {
		var c model.Chat
		if err := rows.Scan(&c.ID, &c.Type, &c.Name, &c.CreatedBy, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning chat row: %w", err)
		}
		chats = append(chats, c)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating chat rows: %w", err)
	}

	return chats, nil
}

// UpdateName renames a chat and returns the updated record.
func (r *ChatRepository) UpdateName(ctx context.Context, id, name string) (*model.Chat, error) {
	query := `
		UPDATE chats
		SET name = $2, updated_at = NOW()
		WHERE id = $1
		RETURNING id, type, name, created_by, created_at, updated_at`

	var c model.Chat
	err := r.db.QueryRow(ctx, query, id, name).Scan(
		&c.ID,
		&c.Type,
		&c.Name,
		&c.CreatedBy,
		&c.CreatedAt,
		&c.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrChatNotFound
		}
		return nil, fmt.Errorf("updating chat name: %w", err)
	}

	return &c, nil
}

// Delete removes a chat. Membership rows are removed via ON DELETE CASCADE.
func (r *ChatRepository) Delete(ctx context.Context, id string) error {
	result, err := r.db.Exec(ctx, `DELETE FROM chats WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("deleting chat: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrChatNotFound
	}

	return nil
}
