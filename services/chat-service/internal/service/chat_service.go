// Package service contains the business logic for the chat-service.
package service

import (
	"context"
	"errors"
	"strings"
	"unicode/utf8"

	"github.com/salemshafik/pote/packages/logger"
	"github.com/salemshafik/pote/services/chat-service/internal/model"
	"github.com/salemshafik/pote/services/chat-service/internal/repository"
)

// Service errors.
var (
	ErrInvalidChatType   = errors.New("chat type must be 'direct' or 'group'")
	ErrEmptyGroupName    = errors.New("group chats require a name")
	ErrGroupNameTooLong  = errors.New("chat name must be at most 150 characters")
	ErrInvalidDirectChat = errors.New("a direct chat must have exactly one other member")
	ErrInvalidRole       = errors.New("role must be 'admin' or 'member'")
	ErrMissingUserID     = errors.New("user_id is required")
	ErrForbidden         = errors.New("insufficient permissions for this action")
	ErrNotMember         = errors.New("you are not a member of this chat")
	ErrCannotModifyOwner = errors.New("the chat owner cannot be modified or removed")
	ErrCannotTargetSelf  = errors.New("use the leave endpoint to remove yourself")
)

// ChatService handles chat and membership business logic, including per-chat RBAC.
type ChatService struct {
	chats   *repository.ChatRepository
	members *repository.MemberRepository
	log     *logger.Logger
}

// NewChatService creates a new ChatService.
func NewChatService(
	chats *repository.ChatRepository,
	members *repository.MemberRepository,
	log *logger.Logger,
) *ChatService {
	return &ChatService{chats: chats, members: members, log: log}
}

// ---- Chats ----

// CreateChat creates a direct or group chat with the creator as owner.
func (s *ChatService) CreateChat(ctx context.Context, creatorID string, req *model.CreateChatRequest) (*model.ChatWithMembers, error) {
	chat := &model.Chat{CreatedBy: creatorID}

	switch req.Type {
	case model.ChatTypeDirect:
		others := dedupeExcluding(req.MemberIDs, creatorID)
		if len(others) != 1 {
			return nil, ErrInvalidDirectChat
		}
		chat.Type = model.ChatTypeDirect
		chat.Name = "" // direct chats are named client-side from participants

	case model.ChatTypeGroup:
		name := strings.TrimSpace(req.Name)
		if name == "" {
			return nil, ErrEmptyGroupName
		}
		if utf8.RuneCountInString(name) > 150 {
			return nil, ErrGroupNameTooLong
		}
		chat.Type = model.ChatTypeGroup
		chat.Name = name

	default:
		return nil, ErrInvalidChatType
	}

	// Build the initial member set: creator as owner, others as members.
	members := []model.ChatMember{{UserID: creatorID, Role: model.RoleOwner}}
	for _, uid := range dedupeExcluding(req.MemberIDs, creatorID) {
		members = append(members, model.ChatMember{UserID: uid, Role: model.RoleMember})
	}

	created, err := s.chats.CreateWithMembers(ctx, chat, members)
	if err != nil {
		return nil, err
	}

	s.log.Info("chat created", "chat_id", created.ID, "type", created.Type, "creator", creatorID)

	return s.withMembers(ctx, created)
}

// GetChat returns a chat with its members. The caller must be a member.
func (s *ChatService) GetChat(ctx context.Context, userID, chatID string) (*model.ChatWithMembers, error) {
	if _, err := s.requireMembership(ctx, chatID, userID); err != nil {
		return nil, err
	}

	chat, err := s.chats.GetByID(ctx, chatID)
	if err != nil {
		return nil, err
	}

	return s.withMembers(ctx, chat)
}

// ListChats returns all chats the caller is a member of.
func (s *ChatService) ListChats(ctx context.Context, userID string) ([]model.Chat, error) {
	return s.chats.ListByUser(ctx, userID)
}

// RenameChat updates a chat's name. Requires owner or admin.
func (s *ChatService) RenameChat(ctx context.Context, userID, chatID string, req *model.UpdateChatRequest) (*model.Chat, error) {
	if _, err := s.requireRole(ctx, chatID, userID, model.RoleOwner, model.RoleAdmin); err != nil {
		return nil, err
	}

	if req.Name == nil {
		return s.chats.GetByID(ctx, chatID)
	}

	name := strings.TrimSpace(*req.Name)
	if name == "" {
		return nil, ErrEmptyGroupName
	}
	if utf8.RuneCountInString(name) > 150 {
		return nil, ErrGroupNameTooLong
	}

	return s.chats.UpdateName(ctx, chatID, name)
}

// DeleteChat deletes a chat. Requires owner.
func (s *ChatService) DeleteChat(ctx context.Context, userID, chatID string) error {
	if _, err := s.requireRole(ctx, chatID, userID, model.RoleOwner); err != nil {
		return err
	}

	if err := s.chats.Delete(ctx, chatID); err != nil {
		return err
	}

	s.log.Info("chat deleted", "chat_id", chatID, "by", userID)
	return nil
}

// ---- Members ----

// ListMembers returns the members of a chat. The caller must be a member.
func (s *ChatService) ListMembers(ctx context.Context, userID, chatID string) ([]model.ChatMember, error) {
	if _, err := s.requireMembership(ctx, chatID, userID); err != nil {
		return nil, err
	}
	return s.members.ListByChat(ctx, chatID)
}

// AddMember adds a member to a chat. Requires owner or admin. New members may
// only be added as 'admin' or 'member' (ownership is not transferable here).
func (s *ChatService) AddMember(ctx context.Context, actorID, chatID string, req *model.AddMemberRequest) (*model.ChatMember, error) {
	if _, err := s.requireRole(ctx, chatID, actorID, model.RoleOwner, model.RoleAdmin); err != nil {
		return nil, err
	}

	if strings.TrimSpace(req.UserID) == "" {
		return nil, ErrMissingUserID
	}

	role := req.Role
	if role == "" {
		role = model.RoleMember
	}
	if role != model.RoleAdmin && role != model.RoleMember {
		return nil, ErrInvalidRole
	}

	member := &model.ChatMember{ChatID: chatID, UserID: req.UserID, Role: role}
	created, err := s.members.Add(ctx, member)
	if err != nil {
		return nil, err
	}

	s.log.Info("member added", "chat_id", chatID, "user_id", req.UserID, "role", role, "by", actorID)
	return created, nil
}

// RemoveMember removes a member from a chat. Owners/admins may remove others
// (but never the owner); the owner cannot be removed.
func (s *ChatService) RemoveMember(ctx context.Context, actorID, chatID, targetID string) error {
	if strings.TrimSpace(targetID) == "" {
		return ErrMissingUserID
	}
	if targetID == actorID {
		return ErrCannotTargetSelf
	}

	if _, err := s.requireRole(ctx, chatID, actorID, model.RoleOwner, model.RoleAdmin); err != nil {
		return err
	}

	targetRole, err := s.members.GetRole(ctx, chatID, targetID)
	if err != nil {
		return err
	}
	if targetRole == model.RoleOwner {
		return ErrCannotModifyOwner
	}

	if err := s.members.Remove(ctx, chatID, targetID); err != nil {
		return err
	}

	s.log.Info("member removed", "chat_id", chatID, "user_id", targetID, "by", actorID)
	return nil
}

// Leave removes the caller from a chat. The owner cannot leave (they must
// delete the chat or transfer ownership in a future iteration).
func (s *ChatService) Leave(ctx context.Context, userID, chatID string) error {
	role, err := s.members.GetRole(ctx, chatID, userID)
	if err != nil {
		if errors.Is(err, repository.ErrMemberNotFound) {
			return ErrNotMember
		}
		return err
	}
	if role == model.RoleOwner {
		return ErrCannotModifyOwner
	}

	if err := s.members.Remove(ctx, chatID, userID); err != nil {
		return err
	}

	s.log.Info("member left chat", "chat_id", chatID, "user_id", userID)
	return nil
}

// UpdateMemberRole changes a member's role. Requires owner. The owner's role
// cannot be changed through this endpoint.
func (s *ChatService) UpdateMemberRole(ctx context.Context, actorID, chatID, targetID string, req *model.UpdateMemberRoleRequest) (*model.ChatMember, error) {
	if strings.TrimSpace(targetID) == "" {
		return nil, ErrMissingUserID
	}
	if req.Role != model.RoleAdmin && req.Role != model.RoleMember {
		return nil, ErrInvalidRole
	}

	if _, err := s.requireRole(ctx, chatID, actorID, model.RoleOwner); err != nil {
		return nil, err
	}

	targetRole, err := s.members.GetRole(ctx, chatID, targetID)
	if err != nil {
		return nil, err
	}
	if targetRole == model.RoleOwner {
		return nil, ErrCannotModifyOwner
	}

	updated, err := s.members.UpdateRole(ctx, chatID, targetID, req.Role)
	if err != nil {
		return nil, err
	}

	s.log.Info("member role updated", "chat_id", chatID, "user_id", targetID, "role", req.Role, "by", actorID)
	return updated, nil
}

// ---- Authorization helpers ----

// requireMembership ensures the user belongs to the chat and returns their role.
func (s *ChatService) requireMembership(ctx context.Context, chatID, userID string) (string, error) {
	role, err := s.members.GetRole(ctx, chatID, userID)
	if err != nil {
		if errors.Is(err, repository.ErrMemberNotFound) {
			return "", ErrNotMember
		}
		return "", err
	}
	return role, nil
}

// requireRole ensures the user is a member with one of the allowed roles.
func (s *ChatService) requireRole(ctx context.Context, chatID, userID string, allowed ...string) (string, error) {
	role, err := s.requireMembership(ctx, chatID, userID)
	if err != nil {
		return "", err
	}
	for _, a := range allowed {
		if role == a {
			return role, nil
		}
	}
	return "", ErrForbidden
}

// withMembers enriches a chat with its member list.
func (s *ChatService) withMembers(ctx context.Context, chat *model.Chat) (*model.ChatWithMembers, error) {
	members, err := s.members.ListByChat(ctx, chat.ID)
	if err != nil {
		return nil, err
	}
	return &model.ChatWithMembers{Chat: *chat, Members: members}, nil
}

// ---- small helpers ----

// dedupeExcluding returns the unique, non-empty IDs from ids, excluding `exclude`.
func dedupeExcluding(ids []string, exclude string) []string {
	seen := make(map[string]struct{}, len(ids))
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id == "" || id == exclude {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}
