// Package handler provides HTTP request handlers for the user-service.
package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	authutils "github.com/salemshafik/pote/packages/auth-utils"
	"github.com/salemshafik/pote/services/user-service/internal/model"
	"github.com/salemshafik/pote/services/user-service/internal/repository"
	"github.com/salemshafik/pote/services/user-service/internal/service"
)

// UserHandler handles HTTP requests for user profile, contact, and invite endpoints.
type UserHandler struct {
	userService *service.UserService
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// ---- Profiles ----

// CreateProfile handles POST /api/v1/users
// Intended for internal service-to-service provisioning (e.g., from auth-service).
func (h *UserHandler) CreateProfile(w http.ResponseWriter, r *http.Request) {
	var req model.CreateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	profile, err := h.userService.CreateProfile(r.Context(), &req)
	if err != nil {
		h.handleError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, profile)
}

// GetMe handles GET /api/v1/users/me
func (h *UserHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	userID := authutils.UserIDFromContext(r.Context())
	profile, err := h.userService.GetProfile(r.Context(), userID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, profile)
}

// GetProfile handles GET /api/v1/users/{id}
func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
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

// UpdateMe handles PUT /api/v1/users/me
func (h *UserHandler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	userID := authutils.UserIDFromContext(r.Context())

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

// UpdateStatus handles PUT /api/v1/users/me/status
func (h *UserHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	userID := authutils.UserIDFromContext(r.Context())

	var req model.UpdateStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	profile, err := h.userService.UpdateStatus(r.Context(), userID, &req)
	if err != nil {
		h.handleError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, profile)
}

// ---- Contacts ----

// ListContacts handles GET /api/v1/users/me/contacts
func (h *UserHandler) ListContacts(w http.ResponseWriter, r *http.Request) {
	userID := authutils.UserIDFromContext(r.Context())

	contacts, err := h.userService.ListContacts(r.Context(), userID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, contacts)
}

// AddContact handles POST /api/v1/users/me/contacts
func (h *UserHandler) AddContact(w http.ResponseWriter, r *http.Request) {
	userID := authutils.UserIDFromContext(r.Context())

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

// RemoveContact handles DELETE /api/v1/users/me/contacts/{contactID}
func (h *UserHandler) RemoveContact(w http.ResponseWriter, r *http.Request) {
	userID := authutils.UserIDFromContext(r.Context())
	contactID := chi.URLParam(r, "contactID")

	if err := h.userService.RemoveContact(r.Context(), userID, contactID); err != nil {
		h.handleError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "contact removed"})
}

// ---- Invites ----

// ListInvites handles GET /api/v1/users/me/invites
func (h *UserHandler) ListInvites(w http.ResponseWriter, r *http.Request) {
	userID := authutils.UserIDFromContext(r.Context())

	invites, err := h.userService.ListInvites(r.Context(), userID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, invites)
}

// CreateInvite handles POST /api/v1/users/me/invites
func (h *UserHandler) CreateInvite(w http.ResponseWriter, r *http.Request) {
	userID := authutils.UserIDFromContext(r.Context())

	var req model.CreateInviteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	invite, err := h.userService.CreateInvite(r.Context(), userID, &req)
	if err != nil {
		h.handleError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, invite)
}

// handleError maps service/repository errors to appropriate HTTP responses.
func (h *UserHandler) handleError(w http.ResponseWriter, err error) {
	switch {
	// Validation (400)
	case errors.Is(err, service.ErrInvalidEmail):
		writeError(w, http.StatusBadRequest, "INVALID_EMAIL", err.Error())
	case errors.Is(err, service.ErrEmptyDisplayName):
		writeError(w, http.StatusBadRequest, "MISSING_NAME", err.Error())
	case errors.Is(err, service.ErrDisplayNameTooLong):
		writeError(w, http.StatusBadRequest, "NAME_TOO_LONG", err.Error())
	case errors.Is(err, service.ErrInvalidStatus):
		writeError(w, http.StatusBadRequest, "INVALID_STATUS", err.Error())
	case errors.Is(err, service.ErrMissingContactID):
		writeError(w, http.StatusBadRequest, "MISSING_CONTACT_ID", err.Error())
	case errors.Is(err, repository.ErrSelfContact):
		writeError(w, http.StatusBadRequest, "SELF_CONTACT", err.Error())
	case errors.Is(err, service.ErrCannotInviteSelf):
		writeError(w, http.StatusBadRequest, "SELF_INVITE", err.Error())

	// Not found (404)
	case errors.Is(err, repository.ErrProfileNotFound):
		writeError(w, http.StatusNotFound, "PROFILE_NOT_FOUND", "User profile not found")
	case errors.Is(err, repository.ErrContactNotFound):
		writeError(w, http.StatusNotFound, "CONTACT_NOT_FOUND", "Contact not found")
	case errors.Is(err, repository.ErrInviteNotFound):
		writeError(w, http.StatusNotFound, "INVITE_NOT_FOUND", "Invite not found")

	// Conflict (409)
	case errors.Is(err, repository.ErrProfileExists):
		writeError(w, http.StatusConflict, "PROFILE_EXISTS", "A profile already exists for this user")
	case errors.Is(err, repository.ErrContactExists):
		writeError(w, http.StatusConflict, "CONTACT_EXISTS", "This contact has already been added")
	case errors.Is(err, repository.ErrInviteAlreadyExists):
		writeError(w, http.StatusConflict, "INVITE_EXISTS", "A pending invite already exists for this email")

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
