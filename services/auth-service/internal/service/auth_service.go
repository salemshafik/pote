// Package service contains the business logic for the auth-service.
package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/mail"
	"time"
	"unicode/utf8"

	authutils "github.com/salemshafik/pote/packages/auth-utils"
	"github.com/salemshafik/pote/packages/logger"
	"github.com/salemshafik/pote/services/auth-service/internal/model"
	"github.com/salemshafik/pote/services/auth-service/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

// Service errors.
var (
	ErrInvalidEmail       = errors.New("invalid email address")
	ErrWeakPassword       = errors.New("password must be at least 8 characters")
	ErrEmptyDisplayName   = errors.New("display name is required")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrTokenRevoked       = errors.New("refresh token has been revoked")
	ErrTokenExpired       = errors.New("refresh token has expired")
)

// AuthService handles authentication business logic.
type AuthService struct {
	userRepo  *repository.UserRepository
	tokenRepo *repository.TokenRepository
	tokenCfg  authutils.TokenConfig
	log       *logger.Logger
}

// NewAuthService creates a new AuthService.
func NewAuthService(
	userRepo *repository.UserRepository,
	tokenRepo *repository.TokenRepository,
	tokenCfg authutils.TokenConfig,
	log *logger.Logger,
) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		tokenRepo: tokenRepo,
		tokenCfg:  tokenCfg,
		log:       log,
	}
}

// Register creates a new user with email and password.
func (s *AuthService) Register(ctx context.Context, req *model.RegisterRequest) (*model.AuthResponse, error) {
	// Validate input
	if err := validateEmail(req.Email); err != nil {
		return nil, err
	}
	if utf8.RuneCountInString(req.Password) < 8 {
		return nil, ErrWeakPassword
	}
	if req.DisplayName == "" {
		return nil, ErrEmptyDisplayName
	}

	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hashing password: %w", err)
	}

	// Create user
	user := &model.User{
		Email:        req.Email,
		DisplayName:  req.DisplayName,
		PasswordHash: string(hash),
		Provider:     "email",
		Role:         "user",
	}

	created, err := s.userRepo.Create(ctx, user)
	if err != nil {
		if errors.Is(err, repository.ErrEmailAlreadyExists) {
			return nil, repository.ErrEmailAlreadyExists
		}
		return nil, fmt.Errorf("creating user: %w", err)
	}

	s.log.Info("user registered", "user_id", created.ID, "email", created.Email)

	// Generate tokens
	return s.generateAuthResponse(ctx, created)
}

// Login authenticates a user with email and password.
func (s *AuthService) Login(ctx context.Context, req *model.LoginRequest) (*model.AuthResponse, error) {
	if err := validateEmail(req.Email); err != nil {
		return nil, ErrInvalidCredentials
	}

	// Fetch user
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("fetching user: %w", err)
	}

	// Verify password
	if user.Provider != "email" || user.PasswordHash == "" {
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	s.log.Info("user logged in", "user_id", user.ID, "email", user.Email)

	return s.generateAuthResponse(ctx, user)
}

// GoogleCallback handles the Google OAuth callback by creating or fetching
// the user and returning an auth response with tokens.
func (s *AuthService) GoogleCallback(ctx context.Context, googleUser *GoogleUserInfo) (*model.AuthResponse, error) {
	// Try to find existing user by provider ID
	user, err := s.userRepo.GetByProviderID(ctx, "google", googleUser.ID)
	if err != nil && !errors.Is(err, repository.ErrUserNotFound) {
		return nil, fmt.Errorf("looking up google user: %w", err)
	}

	if user == nil {
		// Check if email already exists (link accounts)
		user, err = s.userRepo.GetByEmail(ctx, googleUser.Email)
		if err != nil && !errors.Is(err, repository.ErrUserNotFound) {
			return nil, fmt.Errorf("looking up user by email: %w", err)
		}

		if user == nil {
			// Create new user from Google profile
			user = &model.User{
				Email:       googleUser.Email,
				DisplayName: googleUser.Name,
				AvatarURL:   googleUser.Picture,
				Provider:    "google",
				ProviderID:  googleUser.ID,
				Role:        "user",
			}

			user, err = s.userRepo.Create(ctx, user)
			if err != nil {
				return nil, fmt.Errorf("creating google user: %w", err)
			}

			s.log.Info("google user registered", "user_id", user.ID, "email", user.Email)
		}
	}

	s.log.Info("google user logged in", "user_id", user.ID, "email", user.Email)

	return s.generateAuthResponse(ctx, user)
}

// RefreshToken validates a refresh token and issues new tokens.
func (s *AuthService) RefreshToken(ctx context.Context, refreshTokenStr string) (*model.AuthResponse, error) {
	// Hash the token to look it up
	tokenHash := hashToken(refreshTokenStr)

	// Look up the refresh token
	storedToken, err := s.tokenRepo.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, repository.ErrTokenNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("looking up refresh token: %w", err)
	}

	// Check if revoked
	if storedToken.Revoked {
		return nil, ErrTokenRevoked
	}

	// Check if expired
	if storedToken.ExpiresAt.Before(time.Now()) {
		return nil, ErrTokenExpired
	}

	// Revoke the old refresh token (rotation)
	if err := s.tokenRepo.RevokeByTokenHash(ctx, tokenHash); err != nil {
		return nil, fmt.Errorf("revoking old refresh token: %w", err)
	}

	// Fetch the user
	user, err := s.userRepo.GetByID(ctx, storedToken.UserID)
	if err != nil {
		return nil, fmt.Errorf("fetching user for refresh: %w", err)
	}

	s.log.Info("token refreshed", "user_id", user.ID)

	return s.generateAuthResponse(ctx, user)
}

// Logout revokes a refresh token.
func (s *AuthService) Logout(ctx context.Context, refreshTokenStr string) error {
	tokenHash := hashToken(refreshTokenStr)

	if err := s.tokenRepo.RevokeByTokenHash(ctx, tokenHash); err != nil {
		if errors.Is(err, repository.ErrTokenNotFound) {
			return nil // Already logged out, idempotent
		}
		return fmt.Errorf("revoking refresh token: %w", err)
	}

	s.log.Info("user logged out")
	return nil
}

// generateAuthResponse creates access and refresh tokens for a user.
func (s *AuthService) generateAuthResponse(ctx context.Context, user *model.User) (*model.AuthResponse, error) {
	role := authutils.Role(user.Role)

	// Generate access token
	accessToken, err := authutils.GenerateAccessToken(s.tokenCfg, user.ID, user.Email, role)
	if err != nil {
		return nil, fmt.Errorf("generating access token: %w", err)
	}

	// Generate refresh token
	refreshToken, err := authutils.GenerateRefreshToken(s.tokenCfg, user.ID)
	if err != nil {
		return nil, fmt.Errorf("generating refresh token: %w", err)
	}

	// Store refresh token hash in database
	tokenHash := hashToken(refreshToken)
	storedToken := &model.RefreshToken{
		UserID:    user.ID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(s.tokenCfg.RefreshTokenExpiry),
	}

	if err := s.tokenRepo.Create(ctx, storedToken); err != nil {
		return nil, fmt.Errorf("storing refresh token: %w", err)
	}

	return &model.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         model.UserInfoFromUser(user),
	}, nil
}

// GoogleUserInfo represents user information from Google's OAuth API.
type GoogleUserInfo struct {
	ID      string `json:"id"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
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

// hashToken creates a SHA-256 hash of a token string.
// We store hashes instead of raw tokens for security.
func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}
