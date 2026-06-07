package model

// CreateProfileRequest is the payload for provisioning a profile. This is an
// internal endpoint, typically called by the auth-service after a user
// successfully registers, so the ID must match the auth users.id UUID.
type CreateProfileRequest struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
	AvatarURL   string `json:"avatar_url"`
}

// UpdateProfileRequest is the payload for updating the authenticated user's
// own profile. All fields are optional; only non-nil fields are applied.
type UpdateProfileRequest struct {
	DisplayName *string `json:"display_name"`
	AvatarURL   *string `json:"avatar_url"`
	Bio         *string `json:"bio"`
}

// UpdateStatusRequest is the payload for updating presence status.
type UpdateStatusRequest struct {
	Status string `json:"status"`
}

// AddContactRequest is the payload for adding a contact.
type AddContactRequest struct {
	ContactID string `json:"contact_id"`
	Nickname  string `json:"nickname"`
}

// CreateInviteRequest is the payload for sending an email invite.
type CreateInviteRequest struct {
	Email string `json:"email"`
}
