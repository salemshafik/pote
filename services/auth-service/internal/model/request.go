package model

// RegisterRequest is the payload for email/password registration.
type RegisterRequest struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
}

// LoginRequest is the payload for email/password login.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RefreshRequest is the payload for refreshing an access token.
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// LogoutRequest is the payload for logging out (revoking refresh token).
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}
