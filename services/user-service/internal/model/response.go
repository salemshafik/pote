package model

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

// PaginatedResponse wraps a list of items with pagination metadata.
type PaginatedResponse struct {
	Items   any    `json:"items"`
	Total   int    `json:"total"`
	HasMore bool   `json:"has_more"`
}
