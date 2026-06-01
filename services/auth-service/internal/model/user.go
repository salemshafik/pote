// Package model defines the data structures for the auth-service domain.
package model

import "time"

// User represents a registered user in the auth database.
type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	DisplayName  string    `json:"display_name"`
	PasswordHash string    `json:"-"` // Never serialized to JSON
	AvatarURL    string    `json:"avatar_url,omitempty"`
	Provider     string    `json:"provider"` // "email" or "google"
	ProviderID   string    `json:"-"`        // External OAuth provider ID
	Role         string    `json:"role"`     // "user" or "admin"
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// RefreshToken represents a stored refresh token for session management.
type RefreshToken struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	TokenHash string    `json:"-"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
	Revoked   bool      `json:"revoked"`
}
