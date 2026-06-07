package model

// APIResponse is the standard API response wrapper, matching the convention
// established across all Pote services.
type APIResponse struct {
	Data  any       `json:"data,omitempty"`
	Error *APIError `json:"error,omitempty"`
}

// APIError represents a structured error in the API response.
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
