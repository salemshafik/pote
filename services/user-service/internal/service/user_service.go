// Package service contains the business logic for the user-service.
package service

import (
	"context"
	"errors"
	"fmt"
	"net/mail"

	"github.com/salemshafik/pote/packages/logger"
	"github.com/salemshafik/pote/services/user-service/internal/model"
	"github.com/salemshafik/pote/services/user-service/internal/repository"
)

// Service errors.
var (
	ErrInvalidEmail     = errors.New("invalid email address")
	ErrEmptyQuery       = errors.New("search query is required")
	ErrEmptyContactID   = errors.New("contact_id is required")
	ErrProfileNotFound  = errors.New("user profile not found")
	ErrContactNotFound  = errors.New("contact not found")
	ErrCannotAddSelf    = errors.New("cannot add yourself as a contact")
	ErrContactExists    = errors.New("contact already exists")
	ErrAlreadyRegistered = errors.New("user is already registered, add them as a contact instead")
)

// UserService handles user profile, contact, and invite business logic.
type UserService struct {
	profileRepo *repository.ProfileRepository
	contactRepo *repository.ContactRepository
	inviteRepo  *repository.InviteRepository
	log         *logger.Logger
}

// NewUserService creates a new UserService.
func NewUserService(
	profileRepo *repository.ProfileRepository,
	contactRepo *repository.ContactRepository,
	inviteRepo *repository.InviteRepository,
	log *logger.Logger,
) *UserService {
	return &UserService{
		profileRepo: profileRepo,
		contactRepo: contactRepo,
		inviteRepo:  inviteRepo,
		log:         log,
	}
}

// ---- Profile Operations ----

// GetProfile retrieves a user profile by ID.
func (s *UserService) GetProfile(ctx context.Context, userID string) (*model.UserProfile, error) {
	profile, err := s.profileRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrProfileNotFound) {
			return nil, ErrProfileNotFound
		}
		return nil, fmt.Errorf("fetching profile: %w", err)
	}
	return profile, nil
}

// UpdateProfile updates the current user's profile.
func (s *UserService) UpdateProfile(ctx context.Context, userID string, req *model.UpdateProfileRequest) (*model.UserProfile, error) {
	profile, err := s.profileRepo.Update(ctx, userID, req)
	if err != nil {
		if errors.Is(err, repository.ErrProfileNotFound) {
			return nil, ErrProfileNotFound
		}
		return nil, fmt.Errorf("updating profile: %w", err)
	}

	s.log.Info("profile updated", "user_id", userID)
	return profile, nil
}

// SyncProfile creates or updates a profile from auth-service data.
// Called internally when a user registers or logs in.
func (s *UserService) SyncProfile(ctx context.Context, req *model.CreateProfileRequest) (*model.UserProfile, error) {
	profile := &model.UserProfile{
		ID:          req.ID,
		Email:       req.Email,
		DisplayName: req.DisplayName,
		AvatarURL:   req.AvatarURL,
	}

	upserted, err := s.profileRepo.Upsert(ctx, profile)
	if err != nil {
		return nil, fmt.Errorf("syncing profile: %w", err)
	}

	s.log.Info("profile synced", "user_id", upserted.ID)
	return upserted, nil
}

// SearchUsers finds user profiles matching a query string.
func (s *UserService) SearchUsers(ctx context.Context, query string, limit, offset int) ([]*model.UserProfile, int, error) {
	if query == "" {
		return nil, 0, ErrEmptyQuery
	}

	if limit <= 0 || limit > 50 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	profiles, total, err := s.profileRepo.Search(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("searching users: %w", err)
	}

	return profiles, total, nil
}

// ---- Contact Operations ----

// ListContacts returns all contacts for a user.
func (s *UserService) ListContacts(ctx context.Context, ownerID string) ([]*model.ContactWithProfile, error) {
	contacts, err := s.contactRepo.ListByOwner(ctx, ownerID)
	if err != nil {
		return nil, fmt.Errorf("listing contacts: %w", err)
	}
	return contacts, nil
}

// AddContact adds a user as a contact.
func (s *UserService) AddContact(ctx context.Context, ownerID string, req *model.AddContactRequest) (*model.Contact, error) {
	if req.ContactID == "" {
		return nil, ErrEmptyContactID
	}

	if ownerID == req.ContactID {
		return nil, ErrCannotAddSelf
	}

	// Verify the contact user exists
	_, err := s.profileRepo.GetByID(ctx, req.ContactID)
	if err != nil {
		if errors.Is(err, repository.ErrProfileNotFound) {
			return nil, ErrProfileNotFound
		}
		return nil, fmt.Errorf("verifying contact user: %w", err)
	}

	contact, err := s.contactRepo.Create(ctx, ownerID, req.ContactID, req.Nickname)
	if err != nil {
		if errors.Is(err, repository.ErrContactAlreadyExists) {
			return nil, ErrContactExists
		}
		if errors.Is(err, repository.ErrCannotAddSelf) {
			return nil, ErrCannotAddSelf
		}
		return nil, fmt.Errorf("adding contact: %w", err)
	}

	s.log.Info("contact added", "owner_id", ownerID, "contact_id", req.ContactID)
	return contact, nil
}

// RemoveContact removes a contact by its record ID.
func (s *UserService) RemoveContact(ctx context.Context, ownerID, contactRecordID string) error {
	if err := s.contactRepo.Delete(ctx, ownerID, contactRecordID); err != nil {
		if errors.Is(err, repository.ErrContactNotFound) {
			return ErrContactNotFound
		}
		return fmt.Errorf("removing contact: %w", err)
	}

	s.log.Info("contact removed", "owner_id", ownerID, "contact_record_id", contactRecordID)
	return nil
}

// ---- Invite Operations ----

// SendInvite creates an email invite for a non-registered user.
func (s *UserService) SendInvite(ctx context.Context, inviterID string, req *model.SendInviteRequest) (*model.Invite, error) {
	if err := validateEmail(req.Email); err != nil {
		return nil, err
	}

	// Check if the email is already registered
	profiles, _, err := s.profileRepo.Search(ctx, req.Email, 1, 0)
	if err != nil {
		return nil, fmt.Errorf("checking existing user: %w", err)
	}
	if len(profiles) > 0 && profiles[0].Email == req.Email {
		return nil, ErrAlreadyRegistered
	}

	invite, err := s.inviteRepo.Create(ctx, inviterID, req.Email)
	if err != nil {
		return nil, fmt.Errorf("creating invite: %w", err)
	}

	// TODO: Publish event to notification-service to send the invite email
	s.log.Info("invite sent", "inviter_id", inviterID, "email", req.Email, "invite_id", invite.ID)

	return invite, nil
}

// validateEmail checks if an email address is valid.
func validateEmail(email string) error {
	if email == "" {
		return ErrInvalidEmail
	}
	_, err := mail.ParseAddress(email)
	if err != nil {
		return ErrInvalidEmail
	}
	return nil
}
