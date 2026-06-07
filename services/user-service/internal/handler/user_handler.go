// Package handler provides HTTP request handlers for the user-service.
package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	authutils "github.com/salemshafik/pote/packages/auth-utils"
	"github.com/salemshafik/pote/services/user-service/internal/model"
	"github.com/salemshafik/pote/services/user-service/internal/service"
)

// UserHandler handles HTTP requests for user-related endpoints.
type UserHandler struct {
	userService *service.UserService
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// GetMe handles GET /api/v1/users/me
func (h *UserHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	userID := authutils.UserIDFromContext(r.Context())
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	profile, err := h.userService.GetProfile(r.Context(), userID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, profile)
}

// UpdateMe handles PUT /api/v1/users/me
func (h *UserHandler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	userID := authutils.UserIDFromContext(r.Context())
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	var req model.UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	profile, err := h.userService.UpdateProfile(r.Context(), userID, &req)
	if err != nil {
		h.handleError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, profile)
}

// GetUser handles GET /api/v1/users/{id}
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "MISSING_ID", "User ID is required")
		return
	}

	profile, err := h.userService.GetProfile(r.Context(), id)
	if err != nil {
		h.handleError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, profile)
}

// SearchUsers handles GET /api/v1/users/search?q=&limit=&offset=
func (h *UserHandler) SearchUsers(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	profiles, total, err := h.userService.SearchUsers(r.Context(), query, limit, offset)
	if err != nil {
		h.handleError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, model.PaginatedResponse{
		Items:   profiles,
		Total:   total,
		HasMore: offset+len(profiles) < total,
	})
}

// SyncProfile handles POST /internal/v1/users/sync
// Internal endpoint called by auth-service to create/update profiles.
func (h *UserHandler) SyncProfile(w http.ResponseWriter, r *http.Request) {
	var req model.CreateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if req.ID == "" || req.Email == "" || req.DisplayName == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "id, email, and display_name are required")
		return
	}

	profile, err := h.userService.SyncProfile(r.Context(), &req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL", "Failed to sync profile")
		return
	}

	writeJSON(w, http.StatusOK, profile)
}

// ListContacts handles GET /api/v1/contacts
func (h *UserHandler) ListContacts(w http.ResponseWriter, r *http.Request) {
	userID := authutils.UserIDFromContext(r.Context())
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	contacts, err := h.userService.ListContacts(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL", "Failed to list contacts")
		return
	}

	writeJSON(w, http.StatusOK, contacts)
}

// AddContact handles POST /api/v1/contacts
func (h *UserHandler) AddContact(w http.ResponseWriter, r *http.Request) {
	userID := authutils.UserIDFromContext(r.Context())
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	var req model.AddContactRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	contact, err := h.userService.AddContact(r.Context(), userID, &req)
	if err != nil {
		h.handleError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, contact)
}

// RemoveContact handles DELETE /api/v1/contacts/{id}
func (h *UserHandler) RemoveContact(w http.ResponseWriter, r *http.Request) {
	userID := authutils.UserIDFromContext(r.Context())
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	contactID := chi.URLParam(r, "id")
	if contactID == "" {
		writeError(w, http.StatusBadRequest, "MISSING_ID", "Contact ID is required")
		return
	}

	if err := h.userService.RemoveContact(r.Context(), userID, contactID); err != nil {
		h.handleError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "contact removed"})
}

// SendInvite handles POST /api/v1/invites
func (h *UserHandler) SendInvite(w http.ResponseWriter, r *http.Request) {
	userID := authutils.UserIDFromContext(r.Context())
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	var req model.SendInviteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	invite, err := h.userService.SendInvite(r.Context(), userID, &req)
	if err != nil {
		h.handleError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, invite)
}

// handleError maps service errors to HTTP responses.
func (h *UserHandler) handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrProfileNotFound):
		writeError(w, http.StatusNotFound, "NOT_FOUND", "User not found")
	case errors.Is(err, service.ErrContactNotFound):
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Contact not found")
	case errors.Is(err, service.ErrCannotAddSelf):
		writeError(w, http.StatusBadRequest, "CANNOT_ADD_SELF", err.Error())
	case errors.Is(err, service.ErrContactExists):
		writeError(w, http.StatusConflict, "CONTACT_EXISTS", err.Error())
	case errors.Is(err, service.ErrEmptyContactID):
		writeError(w, http.StatusBadRequest, "MISSING_CONTACT_ID", err.Error())
	case errors.Is(err, service.ErrEmptyQuery):
		writeError(w, http.StatusBadRequest, "EMPTY_QUERY", err.Error())
	case errors.Is(err, service.ErrInvalidEmail):
		writeError(w, http.StatusBadRequest, "INVALID_EMAIL", err.Error())
	case errors.Is(err, service.ErrAlreadyRegistered):
		writeError(w, http.StatusConflict, "ALREADY_REGISTERED", err.Error())
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
