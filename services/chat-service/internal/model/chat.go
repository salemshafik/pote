// Package model defines the data structures for the chat-service domain.
package model

import "time"

// Chat types.
const (
	ChatTypeDirect = "direct" // 1-on-1 conversation
	ChatTypeGroup  = "group"  // multi-user group conversation
)

// Per-chat member roles for RBAC. These are distinct from the global JWT roles
// in packages/auth-utils; they govern permissions within a single chat.
const (
	RoleOwner  = "owner"  // full control: rename, delete, manage members & roles
	RoleAdmin  = "admin"  // can add/remove members
	RoleMember = "member" // can participate and leave
)

// Chat represents a conversation's metadata.
type Chat struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Name      string    `json:"name"`
	CreatedBy string    `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ChatMember represents a user's membership and role within a chat.
type ChatMember struct {
	ID       string    `json:"id"`
	ChatID   string    `json:"chat_id"`
	UserID   string    `json:"user_id"`
	Role     string    `json:"role"`
	JoinedAt time.Time `json:"joined_at"`
}

// ChatWithMembers is a chat enriched with its full member list.
type ChatWithMembers struct {
	Chat
	Members []ChatMember `json:"members"`
}
