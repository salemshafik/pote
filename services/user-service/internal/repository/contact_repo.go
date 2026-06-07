package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/salemshafik/pote/services/user-service/internal/model"
)

// ContactRepository handles database operations for contacts.
type ContactRepository struct {
	db *pgxpool.Pool
}

// NewContactRepository creates a new ContactRepository.
func NewContactRepository(db *pgxpool.Pool) *ContactRepository {
	return &ContactRepository{db: db}
}

// Create inserts a new contact relationship and returns the created record.
// It maps PostgreSQL constraint violations to domain errors:
//   - unique_violation      -> ErrContactExists (owner already has this contact)
//   - foreign_key_violation -> ErrProfileNotFound (contact_id has no profile)
//   - check_violation       -> ErrSelfContact (owner_id == contact_id)
func (r *ContactRepository) Create(ctx context.Context, c *model.Contact) (*model.Contact, error) {
	query := `
		INSERT INTO contacts (owner_id, contact_id, nickname)
		VALUES ($1, $2, $3)
		RETURNING id, owner_id, contact_id, nickname, created_at`

	var created model.Contact
	err := r.db.QueryRow(ctx, query, c.OwnerID, c.ContactID, c.Nickname).Scan(
		&created.ID,
		&created.OwnerID,
		&created.ContactID,
		&created.Nickname,
		&created.CreatedAt,
	)

	if err != nil {
		switch {
		case isUniqueViolation(err):
			return nil, ErrContactExists
		case isForeignKeyViolation(err):
			return nil, ErrProfileNotFound
		case isCheckViolation(err):
			return nil, ErrSelfContact
		default:
			return nil, fmt.Errorf("inserting contact: %w", err)
		}
	}

	return &created, nil
}

// ListByOwner returns all contacts owned by the given user, newest first.
func (r *ContactRepository) ListByOwner(ctx context.Context, ownerID string) ([]model.Contact, error) {
	query := `
		SELECT id, owner_id, contact_id, nickname, created_at
		FROM contacts
		WHERE owner_id = $1
		ORDER BY created_at DESC`

	rows, err := r.db.Query(ctx, query, ownerID)
	if err != nil {
		return nil, fmt.Errorf("querying contacts: %w", err)
	}
	defer rows.Close()

	contacts := make([]model.Contact, 0)
	for rows.Next() {
		var c model.Contact
		if err := rows.Scan(&c.ID, &c.OwnerID, &c.ContactID, &c.Nickname, &c.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning contact row: %w", err)
		}
		contacts = append(contacts, c)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating contact rows: %w", err)
	}

	return contacts, nil
}

// Delete removes a contact owned by ownerID. Returns ErrContactNotFound if no
// matching row exists.
func (r *ContactRepository) Delete(ctx context.Context, ownerID, contactID string) error {
	query := `DELETE FROM contacts WHERE owner_id = $1 AND contact_id = $2`

	result, err := r.db.Exec(ctx, query, ownerID, contactID)
	if err != nil {
		return fmt.Errorf("deleting contact: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrContactNotFound
	}

	return nil
}
