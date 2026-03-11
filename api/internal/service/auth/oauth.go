package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/deployer/api/internal/config"
	"github.com/deployer/api/internal/model"
	"github.com/deployer/api/internal/repository"
)

var (
	ErrOAuthCodeExchange = errors.New("failed to exchange OAuth code for token")
	ErrOAuthUserInfo     = errors.New("failed to get user info from OAuth provider")
)

// OAuthProviderConfig holds the OAuth configuration for a single provider.
type OAuthProviderConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

// OAuthService handles OAuth authentication flows for GitHub and GitLab.
type OAuthService struct {
	github   OAuthProviderConfig
	gitlab   OAuthProviderConfig
	userRepo repository.UserRepository
	cfg      *config.Config
}

// NewOAuthService creates a new OAuthService.
func NewOAuthService(userRepo repository.UserRepository, cfg *config.Config) *OAuthService {
	return &OAuthService{
		github: OAuthProviderConfig{
			ClientID:     cfg.GitHubClientID,
			ClientSecret: cfg.GitHubClientSecret,
			RedirectURL:  cfg.GitHubRedirectURL,
		},
		gitlab: OAuthProviderConfig{
			ClientID:     cfg.GitLabClientID,
			ClientSecret: cfg.GitLabClientSecret,
			RedirectURL:  cfg.GitLabRedirectURL,
		},
		userRepo: userRepo,
		cfg:      cfg,
	}
}

// GetGitHubAuthURL returns the GitHub OAuth authorization URL.
func (s *OAuthService) GetGitHubAuthURL(state string) string {
	params := url.Values{
		"client_id":    {s.github.ClientID},
		"redirect_uri": {s.github.RedirectURL},
		"scope":        {"user:email"},
		"state":        {state},
	}
	return "https://github.com/login/oauth/authorize?" + params.Encode()
}

// GetGitLabAuthURL returns the GitLab OAuth authorization URL.
func (s *OAuthService) GetGitLabAuthURL(state string) string {
	params := url.Values{
		"client_id":     {s.gitlab.ClientID},
		"redirect_uri":  {s.gitlab.RedirectURL},
		"response_type": {"code"},
		"scope":         {"read_user"},
		"state":         {state},
	}
	return "https://gitlab.com/oauth/authorize?" + params.Encode()
}

// gitHubTokenResponse represents the GitHub OAuth token exchange response.
type gitHubTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
}

// gitHubUser represents the GitHub user API response.
type gitHubUser struct {
	ID        int64   `json:"id"`
	Login     string  `json:"login"`
	Name      string  `json:"name"`
	Email     string  `json:"email"`
	AvatarURL string  `json:"avatar_url"`
}

// gitHubEmail represents a GitHub user email.
type gitHubEmail struct {
	Email    string `json:"email"`
	Primary  bool   `json:"primary"`
	Verified bool   `json:"verified"`
}

// HandleGitHubCallback exchanges the OAuth code for a token, fetches user info,
// and finds or creates the user.
func (s *OAuthService) HandleGitHubCallback(ctx context.Context, code string) (*model.User, error) {
	// 1. Exchange code for access token.
	tokenReqData := url.Values{
		"client_id":     {s.github.ClientID},
		"client_secret": {s.github.ClientSecret},
		"code":          {code},
		"redirect_uri":  {s.github.RedirectURL},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://github.com/login/oauth/access_token",
		strings.NewReader(tokenReqData.Encode()))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrOAuthCodeExchange, err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrOAuthCodeExchange, err)
	}
	defer resp.Body.Close()

	var tokenResp gitHubTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrOAuthCodeExchange, err)
	}
	if tokenResp.AccessToken == "" {
		return nil, ErrOAuthCodeExchange
	}

	// 2. Get user info.
	ghUser, err := s.getGitHubUser(ctx, tokenResp.AccessToken)
	if err != nil {
		return nil, err
	}

	// 3. Get user email if not public.
	email := ghUser.Email
	if email == "" {
		email, err = s.getGitHubPrimaryEmail(ctx, tokenResp.AccessToken)
		if err != nil {
			return nil, err
		}
	}

	if email == "" {
		return nil, fmt.Errorf("%w: no email found on GitHub account", ErrOAuthUserInfo)
	}

	// 4. Find or create user.
	oauthID := strconv.FormatInt(ghUser.ID, 10)
	provider := "github"
	name := ghUser.Name
	if name == "" {
		name = ghUser.Login
	}

	return s.findOrCreateOAuthUser(ctx, provider, oauthID, email, name, ghUser.AvatarURL)
}

func (s *OAuthService) getGitHubUser(ctx context.Context, accessToken string) (*gitHubUser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/user", nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrOAuthUserInfo, err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrOAuthUserInfo, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("%w: status %d: %s", ErrOAuthUserInfo, resp.StatusCode, string(body))
	}

	var user gitHubUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrOAuthUserInfo, err)
	}
	return &user, nil
}

func (s *OAuthService) getGitHubPrimaryEmail(ctx context.Context, accessToken string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/user/emails", nil)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrOAuthUserInfo, err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrOAuthUserInfo, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%w: failed to fetch emails, status %d", ErrOAuthUserInfo, resp.StatusCode)
	}

	var emails []gitHubEmail
	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		return "", fmt.Errorf("%w: %v", ErrOAuthUserInfo, err)
	}

	for _, e := range emails {
		if e.Primary && e.Verified {
			return e.Email, nil
		}
	}
	// Fallback to first verified email.
	for _, e := range emails {
		if e.Verified {
			return e.Email, nil
		}
	}
	return "", nil
}

// gitLabTokenResponse represents the GitLab OAuth token exchange response.
type gitLabTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
}

// gitLabUser represents the GitLab user API response.
type gitLabUser struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
}

// HandleGitLabCallback exchanges the OAuth code for a token, fetches user info,
// and finds or creates the user.
func (s *OAuthService) HandleGitLabCallback(ctx context.Context, code string) (*model.User, error) {
	// 1. Exchange code for access token.
	tokenReqData := url.Values{
		"client_id":     {s.gitlab.ClientID},
		"client_secret": {s.gitlab.ClientSecret},
		"code":          {code},
		"redirect_uri":  {s.gitlab.RedirectURL},
		"grant_type":    {"authorization_code"},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://gitlab.com/oauth/token",
		strings.NewReader(tokenReqData.Encode()))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrOAuthCodeExchange, err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrOAuthCodeExchange, err)
	}
	defer resp.Body.Close()

	var tokenResp gitLabTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrOAuthCodeExchange, err)
	}
	if tokenResp.AccessToken == "" {
		return nil, ErrOAuthCodeExchange
	}

	// 2. Get user info.
	glUser, err := s.getGitLabUser(ctx, tokenResp.AccessToken)
	if err != nil {
		return nil, err
	}

	if glUser.Email == "" {
		return nil, fmt.Errorf("%w: no email found on GitLab account", ErrOAuthUserInfo)
	}

	// 3. Find or create user.
	oauthID := strconv.FormatInt(glUser.ID, 10)
	provider := "gitlab"
	name := glUser.Name
	if name == "" {
		name = glUser.Username
	}

	return s.findOrCreateOAuthUser(ctx, provider, oauthID, glUser.Email, name, glUser.AvatarURL)
}

func (s *OAuthService) getGitLabUser(ctx context.Context, accessToken string) (*gitLabUser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://gitlab.com/api/v4/user", nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrOAuthUserInfo, err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrOAuthUserInfo, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("%w: status %d: %s", ErrOAuthUserInfo, resp.StatusCode, string(body))
	}

	var user gitLabUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrOAuthUserInfo, err)
	}
	return &user, nil
}

// findOrCreateOAuthUser looks up an existing user by OAuth ID or email, or creates a new one.
func (s *OAuthService) findOrCreateOAuthUser(ctx context.Context, provider, oauthID, email, name, avatarURL string) (*model.User, error) {
	// First try to find by OAuth ID.
	user, err := s.userRepo.GetByOAuthID(ctx, provider, oauthID)
	if err == nil {
		// Update avatar if changed.
		if avatarURL != "" && (user.AvatarURL == nil || *user.AvatarURL != avatarURL) {
			user.AvatarURL = &avatarURL
			_ = s.userRepo.Update(ctx, user)
		}
		return user, nil
	}
	if !errors.Is(err, repository.ErrNotFound) {
		return nil, fmt.Errorf("failed to look up OAuth user: %w", err)
	}

	// Try to find by email and link the OAuth account.
	user, err = s.userRepo.GetByEmail(ctx, email)
	if err == nil {
		user.OAuthProvider = &provider
		user.OAuthID = &oauthID
		if avatarURL != "" {
			user.AvatarURL = &avatarURL
		}
		if updateErr := s.userRepo.Update(ctx, user); updateErr != nil {
			return nil, fmt.Errorf("failed to link OAuth account: %w", updateErr)
		}
		return user, nil
	}
	if !errors.Is(err, repository.ErrNotFound) {
		return nil, fmt.Errorf("failed to look up user by email: %w", err)
	}

	// Create a new user.
	var avatarPtr *string
	if avatarURL != "" {
		avatarPtr = &avatarURL
	}

	newUser := &model.User{
		Email:         email,
		PasswordHash:  "", // OAuth users have no password.
		Name:          name,
		OAuthProvider: &provider,
		OAuthID:       &oauthID,
		AvatarURL:     avatarPtr,
		EmailVerified: true, // OAuth emails are considered verified.
	}

	if err := s.userRepo.Create(ctx, newUser); err != nil {
		return nil, fmt.Errorf("failed to create OAuth user: %w", err)
	}

	return newUser, nil
}
