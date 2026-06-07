// Package model defines the data structures for the user-service domain.
package model

import "time"

// UserProfile represents a user's profile in the user-service database.
// This is a denormalized copy of core user data from auth-service,
// enriched with profile-specific fields (bio, status, etc.).
type UserProfile struct {
	ID          string    `json:"id"` // Same UUID as in auth-service
	Email       string    `json:"email"`
	DisplayName string    `json:"display_name"`
	AvatarURL   string    `json:"avatar_url,omitempty"`
	Bio         string    `json:"bio,omitempty"`
	Status      string    `json:"status,omitempty"` // "online", "away", "offline"
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Contact represents a user-to-user contact relationship.
type Contact struct {
	ID        string    `json:"id"`
	OwnerID   string    `json:"owner_id"`   // The user who added the contact
	ContactID string    `json:"contact_id"` // The user being added as a contact
	Nickname  string    `json:"nickname,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// ContactWithProfile is a contact enriched with the contact's profile info.
type ContactWithProfile struct {
	Contact
	Profile *UserProfile `json:"profile"`
}

// Invite represents an email invitation sent to a non-registered user.
type Invite struct {
	ID          string    `json:"id"`
	InviterID   string    `json:"inviter_id"` // The user who sent the invite
	Email       string    `json:"email"`       // Email of the invitee
	Status      string    `json:"status"`      // "pending", "accepted", "expired"
	CreatedAt   time.Time `json:"created_at"`
	ExpiresAt   time.Time `json:"expires_at"`
}
