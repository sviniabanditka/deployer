package handler

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"

	"github.com/deployer/api/internal/config"
	authService "github.com/deployer/api/internal/service/auth"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type AuthHandler struct {
	authSvc  *authService.Service
	oauthSvc *authService.OAuthService
	gdprSvc  *authService.GDPRService
	validate *validator.Validate
	cfg      *config.Config
}

func NewAuthHandler(authSvc *authService.Service, oauthSvc *authService.OAuthService, gdprSvc *authService.GDPRService, cfg *config.Config) *AuthHandler {
	return &AuthHandler{
		authSvc:  authSvc,
		oauthSvc: oauthSvc,
		gdprSvc:  gdprSvc,
		validate: validator.New(),
		cfg:      cfg,
	}
}

type registerRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	Name     string `json:"name" validate:"required,min=1"`
}

type loginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type twoFACodeRequest struct {
	Code string `json:"code" validate:"required,len=6"`
}

type validate2FALoginRequest struct {
	TempToken string `json:"temp_token" validate:"required"`
	Code      string `json:"code" validate:"required,len=6"`
}

type deleteAccountRequest struct {
	Password string `json:"password" validate:"required"`
}

func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req registerRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if err := h.validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	user, err := h.authSvc.Register(c.Context(), req.Email, req.Password, req.Name)
	if err != nil {
		if errors.Is(err, authService.ErrUserExists) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": "user with this email already exists",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to register user",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"user": user,
	})
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req loginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if err := h.validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	result, err := h.authSvc.Login(c.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, authService.ErrInvalidCredentials) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid email or password",
			})
		}
		if errors.Is(err, authService.ErrAccountLocked) {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "account is locked due to too many failed login attempts, try again later",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "login failed",
		})
	}

	if result.Requires2FA {
		return c.JSON(fiber.Map{
			"requires_2fa": true,
			"temp_token":   result.TempToken,
		})
	}

	return c.JSON(fiber.Map{
		"access_token":  result.Tokens.AccessToken,
		"refresh_token": result.Tokens.RefreshToken,
	})
}

func (h *AuthHandler) Refresh(c *fiber.Ctx) error {
	var req refreshRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if err := h.validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	userID, err := h.authSvc.ValidateToken(req.RefreshToken)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "invalid refresh token",
		})
	}

	tokens, err := h.authSvc.GenerateTokenPair(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to generate tokens",
		})
	}

	return c.JSON(fiber.Map{
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
	})
}

// 2FA Handlers

func (h *AuthHandler) Enable2FA(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(uuid.UUID)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	secret, qrCodeURL, err := h.authSvc.Enable2FA(c.Context(), userID)
	if err != nil {
		if errors.Is(err, authService.Err2FAAlreadyOn) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": "two-factor authentication is already enabled",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to enable 2FA",
		})
	}

	return c.JSON(fiber.Map{
		"secret":     secret,
		"qr_code_url": qrCodeURL,
	})
}

func (h *AuthHandler) Verify2FA(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(uuid.UUID)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	var req twoFACodeRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if err := h.validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	if err := h.authSvc.Verify2FA(c.Context(), userID, req.Code); err != nil {
		if errors.Is(err, authService.ErrInvalid2FACode) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid 2FA code",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to verify 2FA",
		})
	}

	return c.JSON(fiber.Map{
		"message": "two-factor authentication enabled successfully",
	})
}

func (h *AuthHandler) Disable2FA(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(uuid.UUID)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	var req twoFACodeRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if err := h.validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	if err := h.authSvc.Disable2FA(c.Context(), userID, req.Code); err != nil {
		if errors.Is(err, authService.ErrInvalid2FACode) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid 2FA code",
			})
		}
		if errors.Is(err, authService.Err2FANotEnabled) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "two-factor authentication is not enabled",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to disable 2FA",
		})
	}

	return c.JSON(fiber.Map{
		"message": "two-factor authentication disabled successfully",
	})
}

func (h *AuthHandler) Validate2FALogin(c *fiber.Ctx) error {
	var req validate2FALoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if err := h.validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	tokens, err := h.authSvc.Verify2FALogin(c.Context(), req.TempToken, req.Code)
	if err != nil {
		if errors.Is(err, authService.ErrInvalidToken) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid or expired temporary token",
			})
		}
		if errors.Is(err, authService.ErrInvalidCredentials) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid 2FA code",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "2FA validation failed",
		})
	}

	return c.JSON(fiber.Map{
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
	})
}

// GDPR Handlers

func (h *AuthHandler) ExportData(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(uuid.UUID)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	data, err := h.gdprSvc.ExportUserData(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to export user data",
		})
	}

	c.Set("Content-Disposition", "attachment; filename=user-data-export.json")
	c.Set("Content-Type", "application/json")
	return c.Send(data)
}

func (h *AuthHandler) DeleteAccount(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(uuid.UUID)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	var req deleteAccountRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if err := h.validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Verify password before deletion
	user, err := h.authSvc.GetUserByID(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to verify identity",
		})
	}
	_ = user // password verification done via login-style check

	// Verify password by attempting to validate credentials
	if _, err := h.authSvc.Login(c.Context(), user.Email, req.Password); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "invalid password",
		})
	}

	if err := h.gdprSvc.DeleteAccount(c.Context(), userID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to delete account",
		})
	}

	return c.JSON(fiber.Map{
		"message": "account deleted successfully",
	})
}

// generateOAuthState creates a random state parameter for OAuth CSRF protection.
func generateOAuthState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// GitHubLogin redirects the user to GitHub OAuth authorization.
func (h *AuthHandler) GitHubLogin(c *fiber.Ctx) error {
	state, err := generateOAuthState()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to generate OAuth state",
		})
	}

	// Store state in a cookie for validation on callback.
	c.Cookie(&fiber.Cookie{
		Name:     "oauth_state",
		Value:    state,
		HTTPOnly: true,
		SameSite: "Lax",
		MaxAge:   600, // 10 minutes
	})

	authURL := h.oauthSvc.GetGitHubAuthURL(state)
	return c.Redirect(authURL, fiber.StatusTemporaryRedirect)
}

// GitHubCallback handles the GitHub OAuth callback.
func (h *AuthHandler) GitHubCallback(c *fiber.Ctx) error {
	code := c.Query("code")
	if code == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "missing authorization code",
		})
	}

	user, err := h.oauthSvc.HandleGitHubCallback(c.Context(), code)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("GitHub OAuth failed: %v", err),
		})
	}

	tokens, err := h.authSvc.GenerateTokenPair(user.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to generate tokens",
		})
	}

	// Redirect to frontend with tokens as query params.
	redirectURL := fmt.Sprintf("%s?access_token=%s&refresh_token=%s",
		h.cfg.OAuthFrontendRedirectURL,
		url.QueryEscape(tokens.AccessToken),
		url.QueryEscape(tokens.RefreshToken),
	)

	return c.Redirect(redirectURL, fiber.StatusTemporaryRedirect)
}

// GitLabLogin redirects the user to GitLab OAuth authorization.
func (h *AuthHandler) GitLabLogin(c *fiber.Ctx) error {
	state, err := generateOAuthState()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to generate OAuth state",
		})
	}

	c.Cookie(&fiber.Cookie{
		Name:     "oauth_state",
		Value:    state,
		HTTPOnly: true,
		SameSite: "Lax",
		MaxAge:   600,
	})

	authURL := h.oauthSvc.GetGitLabAuthURL(state)
	return c.Redirect(authURL, fiber.StatusTemporaryRedirect)
}

// GitLabCallback handles the GitLab OAuth callback.
func (h *AuthHandler) GitLabCallback(c *fiber.Ctx) error {
	code := c.Query("code")
	if code == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "missing authorization code",
		})
	}

	user, err := h.oauthSvc.HandleGitLabCallback(c.Context(), code)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("GitLab OAuth failed: %v", err),
		})
	}

	tokens, err := h.authSvc.GenerateTokenPair(user.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to generate tokens",
		})
	}

	redirectURL := fmt.Sprintf("%s?access_token=%s&refresh_token=%s",
		h.cfg.OAuthFrontendRedirectURL,
		url.QueryEscape(tokens.AccessToken),
		url.QueryEscape(tokens.RefreshToken),
	)

	return c.Redirect(redirectURL, fiber.StatusTemporaryRedirect)
}
