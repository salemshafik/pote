// Package service contains the business logic for the user-service.
package service

import (
	"context"
	"errors"
	"net/mail"
	"strings"
	"unicode/utf8"

	"github.com/salemshafik/pote/packages/logger"
	"github.com/salemshafik/pote/services/user-service/internal/model"
	"github.com/salemshafik/pote/services/user-service/internal/repository"
)

// Service errors.
var (
	ErrInvalidEmail       = errors.New("invalid email address")
	ErrEmptyDisplayName   = errors.New("display name is required")
	ErrDisplayNameTooLong = errors.New("display name must be at most 100 characters")
	ErrInvalidStatus      = errors.New("invalid status")
	ErrMissingContactID   = errors.New("contact_id is required")
	ErrCannotInviteSelf   = errors.New("cannot invite your own email address")
)

// UserService handles user profile, contact, and invite business logic.
type UserService struct {
	profiles *repository.ProfileRepository
	contacts *repository.ContactRepository
	invites  *repository.InviteRepository
	log      *logger.Logger
}

// NewUserService creates a new UserService.
func NewUserService(
	profiles *repository.ProfileRepository,
	contacts *repository.ContactRepository,
	invites *repository.InviteRepository,
	log *logger.Logger,
) *UserService {
	return &UserService{
		profiles: profiles,
		contacts: contacts,
		invites:  invites,
		log:      log,
	}
}

// ---- Profiles ----

// CreateProfile provisions a new profile. The ID must match the auth-service
// users.id UUID so identities stay aligned across services.
func (s *UserService) CreateProfile(ctx context.Context, req *model.CreateProfileRequest) (*model.UserProfile, error) {
	if req.ID == "" {
		return nil, repository.ErrProfileNotFound // caller must supply the auth UUID
	}
	if err := validateEmail(req.Email); err != nil {
		return nil, err
	}
	if err := validateDisplayName(req.DisplayName); err != nil {
		return nil, err
	}

	profile := &model.UserProfile{
		ID:          req.ID,
		Email:       strings.TrimSpace(req.Email),
		DisplayName: strings.TrimSpace(req.DisplayName),
		AvatarURL:   req.AvatarURL,
		Status:      model.StatusOffline,
	}

	created, err := s.profiles.Create(ctx, profile)
	if err != nil {
		return nil, err
	}

	s.log.Info("profile created", "user_id", created.ID, "email", created.Email)
	return created, nil
}

// GetProfile fetches a profile by ID.
func (s *UserService) GetProfile(ctx context.Context, id string) (*model.UserProfile, error) {
	return s.profiles.GetByID(ctx, id)
}

// UpdateProfile updates the authenticated user's own profile.
func (s *UserService) UpdateProfile(ctx context.Context, userID string, req *model.UpdateProfileRequest) (*model.UserProfile, error) {
	if req.DisplayName != nil {
		if err := validateDisplayName(*req.DisplayName); err != nil {
			return nil, err
		}
		trimmed := strings.TrimSpace(*req.DisplayName)
		req.DisplayName = &trimmed
	}

	updated, err := s.profiles.Update(ctx, userID, req)
	if err != nil {
		return nil, err
	}

	s.log.Info("profile updated", "user_id", userID)
	return updated, nil
}

// UpdateStatus updates the authenticated user's presence status.
func (s *UserService) UpdateStatus(ctx context.Context, userID string, req *model.UpdateStatusRequest) (*model.UserProfile, error) {
	if !isValidStatus(req.Status) {
		return nil, ErrInvalidStatus
	}

	updated, err := s.profiles.UpdateStatus(ctx, userID, req.Status)
	if err != nil {
		return nil, err
	}

	return updated, nil
}

// ---- Contacts ----

// AddContact adds a contact for the authenticated user.
func (s *UserService) AddContact(ctx context.Context, ownerID string, req *model.AddContactRequest) (*model.Contact, error) {
	if req.ContactID == "" {
		return nil, ErrMissingContactID
	}
	if req.ContactID == ownerID {
		return nil, repository.ErrSelfContact
	}

	contact := &model.Contact{
		OwnerID:   ownerID,
		ContactID: req.ContactID,
		Nickname:  strings.TrimSpace(req.Nickname),
	}

	created, err := s.contacts.Create(ctx, contact)
	if err != nil {
		return nil, err
	}

	s.log.Info("contact added", "owner_id", ownerID, "contact_id", req.ContactID)
	return created, nil
}

// ListContacts returns all contacts owned by the authenticated user.
func (s *UserService) ListContacts(ctx context.Context, ownerID string) ([]model.Contact, error) {
	return s.contacts.ListByOwner(ctx, ownerID)
}

// RemoveContact deletes a contact owned by the authenticated user.
func (s *UserService) RemoveContact(ctx context.Context, ownerID, contactID string) error {
	if contactID == "" {
		return ErrMissingContactID
	}
	return s.contacts.Delete(ctx, ownerID, contactID)
}

// ---- Invites ----

// CreateInvite sends an email invitation on behalf of the authenticated user.
func (s *UserService) CreateInvite(ctx context.Context, inviterID string, req *model.CreateInviteRequest) (*model.Invite, error) {
	if err := validateEmail(req.Email); err != nil {
		return nil, err
	}
	email := strings.ToLower(strings.TrimSpace(req.Email))

	// Prevent inviting your own registered email.
	if profile, err := s.profiles.GetByID(ctx, inviterID); err == nil {
		if strings.EqualFold(profile.Email, email) {
			return nil, ErrCannotInviteSelf
		}
	}

	// Avoid duplicate pending invites for the same email.
	exists, err := s.invites.ExistsPending(ctx, inviterID, email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, repository.ErrInviteAlreadyExists
	}

	invite := &model.Invite{
		InviterID: inviterID,
		Email:     email,
		Status:    model.InviteStatusPending,
	}

	created, err := s.invites.Create(ctx, invite)
	if err != nil {
		return nil, err
	}

	s.log.Info("invite created", "inviter_id", inviterID, "email", email)
	return created, nil
}

// ListInvites returns all invites sent by the authenticated user.
func (s *UserService) ListInvites(ctx context.Context, inviterID string) ([]model.Invite, error) {
	return s.invites.ListByInviter(ctx, inviterID)
}

// ---- Validation helpers ----

func validateEmail(email string) error {
	if strings.TrimSpace(email) == "" {
		return ErrInvalidEmail
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return ErrInvalidEmail
	}
	return nil
}

func validateDisplayName(name string) error {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return ErrEmptyDisplayName
	}
	if utf8.RuneCountInString(trimmed) > 100 {
		return ErrDisplayNameTooLong
	}
	return nil
}

func isValidStatus(status string) bool {
	switch status {
	case model.StatusOnline, model.StatusOffline, model.StatusAway, model.StatusBusy:
		return true
	default:
		return false
	}
}
