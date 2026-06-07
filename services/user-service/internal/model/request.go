package model

// CreateProfileRequest is used to create or sync a user profile
// (typically called by auth-service after registration).
type CreateProfileRequest struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
	AvatarURL   string `json:"avatar_url,omitempty"`
}

// UpdateProfileRequest is the payload for updating the current user's profile.
type UpdateProfileRequest struct {
	DisplayName *string `json:"display_name,omitempty"`
	AvatarURL   *string `json:"avatar_url,omitempty"`
	Bio         *string `json:"bio,omitempty"`
}

// AddContactRequest is the payload for adding a user as a contact.
type AddContactRequest struct {
	ContactID string `json:"contact_id"`
	Nickname  string `json:"nickname,omitempty"`
}

// SendInviteRequest is the payload for sending an email invite.
type SendInviteRequest struct {
	Email string `json:"email"`
}
