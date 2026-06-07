// Package handler provides the HTTP layer for the chat-service.
package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	authutils "github.com/salemshafik/pote/packages/auth-utils"
	"github.com/salemshafik/pote/packages/logger"
	"github.com/salemshafik/pote/services/chat-service/internal/model"
	"github.com/salemshafik/pote/services/chat-service/internal/repository"
	"github.com/salemshafik/pote/services/chat-service/internal/service"
)

// ChatHandler handles HTTP requests for chats and members.
type ChatHandler struct {
	svc *service.ChatService
	log *logger.Logger
}

// NewChatHandler creates a new ChatHandler.
func NewChatHandler(svc *service.ChatService, log *logger.Logger) *ChatHandler {
	return &ChatHandler{svc: svc, log: log}
}

// ---- Chat endpoints ----

// CreateChat handles POST /api/v1/chats.
func (h *ChatHandler) CreateChat(w http.ResponseWriter, r *http.Request) {
	userID := authutils.UserIDFromContext(r.Context())

	var req model.CreateChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_body", "invalid request body")
		return
	}

	chat, err := h.svc.CreateChat(r.Context(), userID, &req)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, chat)
}

// ListChats handles GET /api/v1/chats.
func (h *ChatHandler) ListChats(w http.ResponseWriter, r *http.Request) {
	userID := authutils.UserIDFromContext(r.Context())

	chats, err := h.svc.ListChats(r.Context(), userID)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, chats)
}

// GetChat handles GET /api/v1/chats/{chatID}.
func (h *ChatHandler) GetChat(w http.ResponseWriter, r *http.Request) {
	userID := authutils.UserIDFromContext(r.Context())
	chatID := chi.URLParam(r, "chatID")

	chat, err := h.svc.GetChat(r.Context(), userID, chatID)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, chat)
}

// UpdateChat handles PUT /api/v1/chats/{chatID}.
func (h *ChatHandler) UpdateChat(w http.ResponseWriter, r *http.Request) {
	userID := authutils.UserIDFromContext(r.Context())
	chatID := chi.URLParam(r, "chatID")

	var req model.UpdateChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_body", "invalid request body")
		return
	}

	chat, err := h.svc.RenameChat(r.Context(), userID, chatID, &req)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, chat)
}

// DeleteChat handles DELETE /api/v1/chats/{chatID}.
func (h *ChatHandler) DeleteChat(w http.ResponseWriter, r *http.Request) {
	userID := authutils.UserIDFromContext(r.Context())
	chatID := chi.URLParam(r, "chatID")

	if err := h.svc.DeleteChat(r.Context(), userID, chatID); err != nil {
		h.writeServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ---- Member endpoints ----

// ListMembers handles GET /api/v1/chats/{chatID}/members.
func (h *ChatHandler) ListMembers(w http.ResponseWriter, r *http.Request) {
	userID := authutils.UserIDFromContext(r.Context())
	chatID := chi.URLParam(r, "chatID")

	members, err := h.svc.ListMembers(r.Context(), userID, chatID)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, members)
}

// AddMember handles POST /api/v1/chats/{chatID}/members.
func (h *ChatHandler) AddMember(w http.ResponseWriter, r *http.Request) {
	userID := authutils.UserIDFromContext(r.Context())
	chatID := chi.URLParam(r, "chatID")

	var req model.AddMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_body", "invalid request body")
		return
	}

	member, err := h.svc.AddMember(r.Context(), userID, chatID, &req)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, member)
}

// RemoveMember handles DELETE /api/v1/chats/{chatID}/members/{userID}.
func (h *ChatHandler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	actorID := authutils.UserIDFromContext(r.Context())
	chatID := chi.URLParam(r, "chatID")
	targetID := chi.URLParam(r, "userID")

	if err := h.svc.RemoveMember(r.Context(), actorID, chatID, targetID); err != nil {
		h.writeServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// UpdateMemberRole handles PUT /api/v1/chats/{chatID}/members/{userID}/role.
func (h *ChatHandler) UpdateMemberRole(w http.ResponseWriter, r *http.Request) {
	actorID := authutils.UserIDFromContext(r.Context())
	chatID := chi.URLParam(r, "chatID")
	targetID := chi.URLParam(r, "userID")

	var req model.UpdateMemberRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_body", "invalid request body")
		return
	}

	member, err := h.svc.UpdateMemberRole(r.Context(), actorID, chatID, targetID, &req)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, member)
}

// Leave handles DELETE /api/v1/chats/{chatID}/members/me.
func (h *ChatHandler) Leave(w http.ResponseWriter, r *http.Request) {
	userID := authutils.UserIDFromContext(r.Context())
	chatID := chi.URLParam(r, "chatID")

	if err := h.svc.Leave(r.Context(), userID, chatID); err != nil {
		h.writeServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ---- Error mapping & response helpers ----

// writeServiceError maps domain/repository errors to HTTP status codes.
func (h *ChatHandler) writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, repository.ErrChatNotFound):
		writeError(w, http.StatusNotFound, "chat_not_found", err.Error())
	case errors.Is(err, repository.ErrMemberNotFound):
		writeError(w, http.StatusNotFound, "member_not_found", err.Error())
	case errors.Is(err, repository.ErrMemberExists):
		writeError(w, http.StatusConflict, "member_exists", err.Error())
	case errors.Is(err, service.ErrNotMember):
		writeError(w, http.StatusForbidden, "not_a_member", err.Error())
	case errors.Is(err, service.ErrForbidden):
		writeError(w, http.StatusForbidden, "forbidden", err.Error())
	case errors.Is(err, service.ErrCannotModifyOwner):
		writeError(w, http.StatusForbidden, "owner_protected", err.Error())
	case errors.Is(err, service.ErrInvalidChatType),
		errors.Is(err, service.ErrEmptyGroupName),
		errors.Is(err, service.ErrGroupNameTooLong),
		errors.Is(err, service.ErrInvalidDirectChat),
		errors.Is(err, service.ErrInvalidRole),
		errors.Is(err, service.ErrMissingUserID),
		errors.Is(err, service.ErrCannotTargetSelf):
		writeError(w, http.StatusBadRequest, "invalid_request", err.Error())
	default:
		h.log.Error("unhandled service error", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "an unexpected error occurred")
	}
}

// writeJSON writes a successful JSON response wrapped in the standard envelope.
func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(model.APIResponse{Data: data})
}

// writeError writes a structured error response.
func writeError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(model.APIResponse{
		Error: &model.APIError{Code: code, Message: message},
	})
}
