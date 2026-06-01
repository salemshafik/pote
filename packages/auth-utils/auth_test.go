package authutils_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	authutils "github.com/salemshafik/pote/packages/auth-utils"
)

var testCfg = authutils.TokenConfig{
	Secret:             "test-secret-key-for-pote",
	Issuer:             "pote",
	AccessTokenExpiry:  15 * time.Minute,
	RefreshTokenExpiry: 7 * 24 * time.Hour,
}

func TestGenerateAndValidateAccessToken(t *testing.T) {
	token, err := authutils.GenerateAccessToken(testCfg, "user-123", "test@pote.app", authutils.RoleUser)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	claims, err := authutils.ValidateToken(token, testCfg.Secret)
	if err != nil {
		t.Fatalf("failed to validate token: %v", err)
	}

	if claims.UserID != "user-123" {
		t.Errorf("expected userID=user-123, got %s", claims.UserID)
	}
	if claims.Email != "test@pote.app" {
		t.Errorf("expected email=test@pote.app, got %s", claims.Email)
	}
	if claims.Role != authutils.RoleUser {
		t.Errorf("expected role=user, got %s", claims.Role)
	}
	if claims.Issuer != "pote" {
		t.Errorf("expected issuer=pote, got %s", claims.Issuer)
	}
}

func TestValidateTokenInvalidSecret(t *testing.T) {
	token, _ := authutils.GenerateAccessToken(testCfg, "user-123", "test@pote.app", authutils.RoleUser)

	_, err := authutils.ValidateToken(token, "wrong-secret")
	if err == nil {
		t.Error("expected error for invalid secret")
	}
}

func TestHasRole(t *testing.T) {
	token, _ := authutils.GenerateAccessToken(testCfg, "user-123", "test@pote.app", authutils.RoleAdmin)
	claims, _ := authutils.ValidateToken(token, testCfg.Secret)

	if !claims.HasRole(authutils.RoleAdmin) {
		t.Error("expected HasRole(admin) to be true")
	}
	if claims.HasRole(authutils.RoleUser) {
		t.Error("expected HasRole(user) to be false")
	}
}

func TestHasAnyRole(t *testing.T) {
	token, _ := authutils.GenerateAccessToken(testCfg, "user-123", "test@pote.app", authutils.RoleOwner)
	claims, _ := authutils.ValidateToken(token, testCfg.Secret)

	if !claims.HasAnyRole(authutils.RoleAdmin, authutils.RoleOwner) {
		t.Error("expected HasAnyRole(admin, owner) to be true")
	}
	if claims.HasAnyRole(authutils.RoleUser) {
		t.Error("expected HasAnyRole(user) to be false")
	}
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	token, _ := authutils.GenerateAccessToken(testCfg, "user-123", "test@pote.app", authutils.RoleUser)

	handler := authutils.AuthMiddleware(testCfg.Secret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uid := authutils.UserIDFromContext(r.Context())
		if uid != "user-123" {
			t.Errorf("expected user-123, got %s", uid)
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestAuthMiddleware_MissingToken(t *testing.T) {
	handler := authutils.AuthMiddleware(testCfg.Secret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestRequireRole_Forbidden(t *testing.T) {
	token, _ := authutils.GenerateAccessToken(testCfg, "user-123", "test@pote.app", authutils.RoleUser)

	handler := authutils.AuthMiddleware(testCfg.Secret)(
		authutils.RequireRole(authutils.RoleAdmin)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Error("handler should not be called")
		})),
	)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}
}
