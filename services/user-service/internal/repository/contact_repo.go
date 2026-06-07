package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/salemshafik/pote/services/user-service/internal/model"
)

// Contact repository errors.
var (
	ErrContactNotFound      = errors.New("contact not found")
	ErrContactAlreadyExists = errors.New("contact already exists")
	ErrCannotAddSelf        = errors.New("cannot add yourself as a contact")
)

// ContactRepository handles database operations for contacts.
type ContactRepository struct {
	db *pgxpool.Pool
}

// NewContactRepository creates a new ContactRepository.
func NewContactRepository(db *pgxpool.Pool) *ContactRepository {
	return &ContactRepository{db: db}
}

// Create inserts a new contact relationship.
func (r *ContactRepository) Create(ctx context.Context, ownerID, contactID, nickname string) (*model.Contact, error) {
	query := `
		INSERT INTO contacts (owner_id, contact_id, nickname)
		VALUES ($1, $2, $3)
		RETURNING id, owner_id, contact_id, nickname, created_at`

	var c model.Contact
	err := r.db.QueryRow(ctx, query, ownerID, contactID, nickname).Scan(
		&c.ID, &c.OwnerID, &c.ContactID, &c.Nickname, &c.CreatedAt,
	)
	if err != nil {
		if isDuplicateKeyError(err) {
			return nil, ErrContactAlreadyExists
		}
		errStr := err.Error()
		for i := 0; i <= len(errStr)-5; i++ {
			if errStr[i:i+5] == "23514" { // check constraint violation
				return nil, ErrCannotAddSelf
			}
		}
		return nil, fmt.Errorf("inserting contact: %w", err)
	}

	return &c, nil
}

// ListByOwner returns all contacts for a given user, joined with profile info.
func (r *ContactRepository) ListByOwner(ctx context.Context, ownerID string) ([]*model.ContactWithProfile, error) {
	query := `
		SELECT c.id, c.owner_id, c.contact_id, c.nickname, c.created_at,
			   p.id, p.email, p.display_name, p.avatar_url, p.bio, p.status, p.created_at, p.updated_at
		FROM contacts c
		JOIN user_profiles p ON p.id = c.contact_id
		WHERE c.owner_id = $1
		ORDER BY p.display_name ASC`

	rows, err := r.db.Query(ctx, query, ownerID)
	if err != nil {
		return nil, fmt.Errorf("listing contacts: %w", err)
	}
	defer rows.Close()

	var contacts []*model.ContactWithProfile
	for rows.Next() {
		var cwp model.ContactWithProfile
		var p model.UserProfile
		if err := rows.Scan(
			&cwp.ID, &cwp.OwnerID, &cwp.ContactID, &cwp.Nickname, &cwp.CreatedAt,
			&p.ID, &p.Email, &p.DisplayName, &p.AvatarURL, &p.Bio, &p.Status, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning contact row: %w", err)
		}
		cwp.Profile = &p
		contacts = append(contacts, &cwp)
	}

	if contacts == nil {
		contacts = []*model.ContactWithProfile{}
	}

	return contacts, nil
}

// Delete removes a contact relationship by its ID, scoped to the owner.
func (r *ContactRepository) Delete(ctx context.Context, ownerID, contactID string) error {
	query := `DELETE FROM contacts WHERE id = $1 AND owner_id = $2`

	result, err := r.db.Exec(ctx, query, contactID, ownerID)
	if err != nil {
		return fmt.Errorf("deleting contact: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrContactNotFound
	}

	return nil
}

// GetByOwnerAndContact retrieves a specific contact by owner and contact user IDs.
func (r *ContactRepository) GetByOwnerAndContact(ctx context.Context, ownerID, contactUserID string) (*model.Contact, error) {
	query := `
		SELECT id, owner_id, contact_id, nickname, created_at
		FROM contacts
		WHERE owner_id = $1 AND contact_id = $2`

	var c model.Contact
	err := r.db.QueryRow(ctx, query, ownerID, contactUserID).Scan(
		&c.ID, &c.OwnerID, &c.ContactID, &c.Nickname, &c.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrContactNotFound
		}
		return nil, fmt.Errorf("querying contact: %w", err)
	}

	return &c, nil
}
