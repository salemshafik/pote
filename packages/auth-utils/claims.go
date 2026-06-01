// Package authutils provides JWT validation middleware and helpers shared
// across all Pote Go microservices. It handles token parsing, claims
// extraction, and role-based access control.
package authutils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Role represents a user's role for RBAC.
type Role string

const (
	RoleUser  Role = "user"
	RoleAdmin Role = "admin"
	RoleOwner Role = "owner" // Group chat owner
)

// Claims holds the custom JWT claims for Pote.
type Claims struct {
	jwt.RegisteredClaims
	UserID string `json:"uid"`
	Email  string `json:"email"`
	Role   Role   `json:"role"`
}

// Common errors returned during token operations.
var (
	ErrTokenExpired  = errors.New("token has expired")
	ErrTokenInvalid  = errors.New("token is invalid")
	ErrTokenMissing  = errors.New("authorization token is missing")
	ErrInvalidClaims = errors.New("token claims are invalid")
	ErrAccessDenied  = errors.New("access denied: insufficient permissions")
)

// TokenConfig holds JWT signing configuration.
type TokenConfig struct {
	// Secret is the HMAC signing key.
	Secret string
	// Issuer is the token issuer (e.g., "pote").
	Issuer string
	// AccessTokenExpiry is the lifetime of an access token.
	AccessTokenExpiry time.Duration
	// RefreshTokenExpiry is the lifetime of a refresh token.
	RefreshTokenExpiry time.Duration
}

// GenerateAccessToken creates a signed JWT access token for the given user.
func GenerateAccessToken(cfg TokenConfig, userID, email string, role Role) (string, error) {
	now := time.Now()
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    cfg.Issuer,
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(cfg.AccessTokenExpiry)),
		},
		UserID: userID,
		Email:  email,
		Role:   role,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.Secret))
}

// GenerateRefreshToken creates a signed JWT refresh token for the given user.
func GenerateRefreshToken(cfg TokenConfig, userID string) (string, error) {
	now := time.Now()
	claims := jwt.RegisteredClaims{
		Issuer:    cfg.Issuer,
		Subject:   userID,
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(cfg.RefreshTokenExpiry)),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.Secret))
}

// ValidateToken parses and validates a JWT token string, returning the claims.
func ValidateToken(tokenStr, secret string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrTokenInvalid
		}
		return []byte(secret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrTokenInvalid
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidClaims
	}

	return claims, nil
}

// HasRole checks whether the claims include the specified role.
func (c *Claims) HasRole(role Role) bool {
	return c.Role == role
}

// HasAnyRole checks whether the claims include any of the specified roles.
func (c *Claims) HasAnyRole(roles ...Role) bool {
	for _, r := range roles {
		if c.Role == r {
			return true
		}
	}
	return false
}
