// Package handler provides HTTP request handlers for the auth-service.
package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/salemshafik/pote/services/auth-service/internal/model"
	"github.com/salemshafik/pote/services/auth-service/internal/repository"
	"github.com/salemshafik/pote/services/auth-service/internal/service"
)

// AuthHandler handles HTTP requests for authentication endpoints.
type AuthHandler struct {
	authService *service.AuthService
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Register handles POST /api/v1/auth/register
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req model.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	resp, err := h.authService.Register(r.Context(), &req)
	if err != nil {
		h.handleAuthError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, resp)
}

// Login handles POST /api/v1/auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req model.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	resp, err := h.authService.Login(r.Context(), &req)
	if err != nil {
		h.handleAuthError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

// Refresh handles POST /api/v1/auth/refresh
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req model.RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if req.RefreshToken == "" {
		writeError(w, http.StatusBadRequest, "MISSING_TOKEN", "Refresh token is required")
		return
	}

	resp, err := h.authService.RefreshToken(r.Context(), req.RefreshToken)
	if err != nil {
		h.handleAuthError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

// Logout handles POST /api/v1/auth/logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var req model.LogoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if req.RefreshToken == "" {
		writeError(w, http.StatusBadRequest, "MISSING_TOKEN", "Refresh token is required")
		return
	}

	if err := h.authService.Logout(r.Context(), req.RefreshToken); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL", "Failed to logout")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "logged out successfully"})
}

// handleAuthError maps service errors to appropriate HTTP responses.
func (h *AuthHandler) handleAuthError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidEmail):
		writeError(w, http.StatusBadRequest, "INVALID_EMAIL", err.Error())
	case errors.Is(err, service.ErrWeakPassword):
		writeError(w, http.StatusBadRequest, "WEAK_PASSWORD", err.Error())
	case errors.Is(err, service.ErrEmptyDisplayName):
		writeError(w, http.StatusBadRequest, "MISSING_NAME", err.Error())
	case errors.Is(err, repository.ErrEmailAlreadyExists):
		writeError(w, http.StatusConflict, "EMAIL_EXISTS", "An account with this email already exists")
	case errors.Is(err, service.ErrInvalidCredentials):
		writeError(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Invalid email or password")
	case errors.Is(err, service.ErrTokenRevoked):
		writeError(w, http.StatusUnauthorized, "TOKEN_REVOKED", "Refresh token has been revoked")
	case errors.Is(err, service.ErrTokenExpired):
		writeError(w, http.StatusUnauthorized, "TOKEN_EXPIRED", "Refresh token has expired")
	default:
		writeError(w, http.StatusInternalServerError, "INTERNAL", "An unexpected error occurred")
	}
}

// writeJSON writes a JSON response with the given status code.
func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	resp := model.APIResponse{Data: data}
	json.NewEncoder(w).Encode(resp)
}

// writeError writes a JSON error response.
func writeError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	resp := model.APIResponse{
		Error: &model.APIError{
			Code:    code,
			Message: message,
		},
	}
	json.NewEncoder(w).Encode(resp)
}
