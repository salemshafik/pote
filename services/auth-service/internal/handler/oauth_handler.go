package handler

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/salemshafik/pote/services/auth-service/internal/service"
	"golang.org/x/oauth2"
)

// OAuthHandler handles Google OAuth authentication.
type OAuthHandler struct {
	authService *service.AuthService
	oauthCfg    *oauth2.Config
	frontendURL string
}

// NewOAuthHandler creates a new OAuthHandler.
func NewOAuthHandler(authService *service.AuthService, oauthCfg *oauth2.Config, frontendURL string) *OAuthHandler {
	return &OAuthHandler{
		authService: authService,
		oauthCfg:    oauthCfg,
		frontendURL: frontendURL,
	}
}

// GoogleLogin handles GET /api/v1/auth/google
// Redirects the user to Google's consent screen.
func (h *OAuthHandler) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	state := generateState()

	// Store state in a cookie for CSRF protection
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Path:     "/",
		MaxAge:   int(10 * time.Minute / time.Second),
		HttpOnly: true,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteLaxMode,
	})

	url := h.oauthCfg.AuthCodeURL(state, oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// GoogleCallback handles GET /api/v1/auth/google/callback
// Exchanges the auth code for tokens and creates/fetches the user.
func (h *OAuthHandler) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	// Verify state parameter for CSRF protection
	stateCookie, err := r.Cookie("oauth_state")
	if err != nil || stateCookie.Value != r.URL.Query().Get("state") {
		writeError(w, http.StatusBadRequest, "INVALID_STATE", "Invalid OAuth state parameter")
		return
	}

	// Clear state cookie
	http.SetCookie(w, &http.Cookie{
		Name:   "oauth_state",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})

	// Check for error from Google
	if errParam := r.URL.Query().Get("error"); errParam != "" {
		writeError(w, http.StatusBadRequest, "OAUTH_ERROR", fmt.Sprintf("Google OAuth error: %s", errParam))
		return
	}

	// Exchange authorization code for token
	code := r.URL.Query().Get("code")
	if code == "" {
		writeError(w, http.StatusBadRequest, "MISSING_CODE", "Authorization code is required")
		return
	}

	token, err := h.oauthCfg.Exchange(context.Background(), code)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "TOKEN_EXCHANGE", "Failed to exchange authorization code")
		return
	}

	// Fetch user info from Google
	googleUser, err := fetchGoogleUserInfo(r.Context(), h.oauthCfg, token)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "USERINFO_ERROR", "Failed to fetch Google user info")
		return
	}

	// Process the Google login
	resp, err := h.authService.GoogleCallback(r.Context(), googleUser)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL", "Failed to process Google login")
		return
	}

	// Redirect to frontend with tokens as query params
	// In production, consider using a short-lived code instead
	redirectURL := fmt.Sprintf("%s/auth/callback?access_token=%s&refresh_token=%s",
		h.frontendURL, resp.AccessToken, resp.RefreshToken)
	http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
}

// fetchGoogleUserInfo retrieves the user's profile from Google.
func fetchGoogleUserInfo(ctx context.Context, cfg *oauth2.Config, token *oauth2.Token) (*service.GoogleUserInfo, error) {
	client := cfg.Client(ctx, token)

	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, fmt.Errorf("fetching userinfo: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("google userinfo returned status %d", resp.StatusCode)
	}

	var userInfo service.GoogleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("decoding userinfo: %w", err)
	}

	return &userInfo, nil
}

// generateState creates a cryptographically random state string for CSRF protection.
func generateState() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}
