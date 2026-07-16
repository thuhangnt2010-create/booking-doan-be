package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"time"

	"github.com/redis/go-redis/v9"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/thuhangnt2010-create/booking-doan-be/internal/repository"
)

const gdriveOAuthStateTTL = 10 * time.Minute

// BackupOAuthService drives the Google Drive OAuth Web flow: build the
// consent URL, validate the CSRF state, exchange the code for tokens, and
// hand the resulting refresh token to rclone via "rclone config create".
type BackupOAuthService struct {
	Redis *redis.Client
	Repo  *repository.BackupRepository
}

func (s *BackupOAuthService) BuildAuthURL(ctx context.Context, clientID, clientSecret, redirectURL string) (string, error) {
	stateBytes := make([]byte, 16)
	if _, err := rand.Read(stateBytes); err != nil {
		return "", err
	}
	state := hex.EncodeToString(stateBytes)
	if err := s.Redis.Set(ctx, "backup_oauth_state:"+state, "1", gdriveOAuthStateTTL).Err(); err != nil {
		return "", err
	}

	cfg := s.oauthConfig(clientID, clientSecret, redirectURL)
	return cfg.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.SetAuthURLParam("prompt", "consent")), nil
}

// ValidateState is one-time: a state value can only be redeemed once, which
// blocks replay of an intercepted callback URL.
func (s *BackupOAuthService) ValidateState(ctx context.Context, state string) bool {
	n, err := s.Redis.Del(ctx, "backup_oauth_state:"+state).Result()
	return err == nil && n > 0
}

func (s *BackupOAuthService) oauthConfig(clientID, clientSecret, redirectURL string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes: []string{
			"https://www.googleapis.com/auth/drive",
			"https://www.googleapis.com/auth/userinfo.email",
		},
		Endpoint: google.Endpoint,
	}
}

func (s *BackupOAuthService) ExchangeAndStore(ctx context.Context, code, clientID, clientSecret, redirectURL, remoteName string) (string, error) {
	cfg := s.oauthConfig(clientID, clientSecret, redirectURL)
	tok, err := cfg.Exchange(ctx, code)
	if err != nil {
		return "", fmt.Errorf("exchange failed: %w", err)
	}

	email := fetchGoogleEmail(ctx, tok.AccessToken)

	tokJSON, err := json.Marshal(map[string]any{
		"access_token":  tok.AccessToken,
		"token_type":    "Bearer",
		"refresh_token": tok.RefreshToken,
		"expiry":        tok.Expiry.Format(time.RFC3339),
	})
	if err != nil {
		return "", err
	}

	cmd := exec.CommandContext(ctx, "rclone", "config", "create", remoteName, "drive",
		"client_id", clientID,
		"client_secret", clientSecret,
		"scope", "drive",
		"token", string(tokJSON),
		"--non-interactive",
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("rclone config create failed: %s: %w", out, err)
	}

	return email, nil
}

func fetchGoogleEmail(ctx context.Context, accessToken string) string {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	if err != nil {
		return ""
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return ""
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ""
	}
	var parsed struct {
		Email string `json:"email"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return ""
	}
	return parsed.Email
}

func (s *BackupOAuthService) Disconnect(ctx context.Context, remoteName string) error {
	cmd := exec.CommandContext(ctx, "rclone", "config", "delete", remoteName)
	if out, err := cmd.CombinedOutput(); err != nil {
		return errors.New(string(out))
	}
	return s.Repo.ClearGdriveAccount(ctx)
}
