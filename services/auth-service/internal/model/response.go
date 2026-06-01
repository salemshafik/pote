package model

// AuthResponse is the response payload for successful authentication.
type AuthResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	User         *UserInfo `json:"user"`
}

// UserInfo is the public user info returned after authentication.
type UserInfo struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
	AvatarURL   string `json:"avatar_url,omitempty"`
	Role        string `json:"role"`
}

// APIResponse is the standard API response wrapper.
type APIResponse struct {
	Data  any       `json:"data,omitempty"`
	Error *APIError `json:"error,omitempty"`
}

// APIError represents a structured error in the API response.
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// UserInfoFromUser converts a User model to a UserInfo response.
func UserInfoFromUser(u *User) *UserInfo {
	return &UserInfo{
		ID:          u.ID,
		Email:       u.Email,
		DisplayName: u.DisplayName,
		AvatarURL:   u.AvatarURL,
		Role:        u.Role,
	}
}
