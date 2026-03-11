package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/pquerna/otp/totp"
)

var (
	Err2FANotEnabled  = errors.New("two-factor authentication is not enabled")
	Err2FAAlreadyOn   = errors.New("two-factor authentication is already enabled")
	ErrInvalid2FACode = errors.New("invalid two-factor authentication code")
)

func (s *Service) Enable2FA(ctx context.Context, userID uuid.UUID) (secret string, qrCodeURL string, err error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return "", "", err
	}

	if user.TwoFactorEnabled {
		return "", "", Err2FAAlreadyOn
	}

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "Deployer",
		AccountName: user.Email,
	})
	if err != nil {
		return "", "", fmt.Errorf("failed to generate TOTP key: %w", err)
	}

	secret = key.Secret()
	qrCodeURL = key.URL()

	// Store secret on user (not yet verified/enabled)
	user.TwoFactorSecret = &secret
	user.TwoFactorEnabled = false
	if err := s.userRepo.Update(ctx, user); err != nil {
		return "", "", fmt.Errorf("failed to save TOTP secret: %w", err)
	}

	return secret, qrCodeURL, nil
}

func (s *Service) Verify2FA(ctx context.Context, userID uuid.UUID, code string) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	if user.TwoFactorSecret == nil {
		return Err2FANotEnabled
	}

	if !totp.Validate(code, *user.TwoFactorSecret) {
		return ErrInvalid2FACode
	}

	// Mark 2FA as verified/enabled
	user.TwoFactorEnabled = true
	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to enable 2FA: %w", err)
	}

	return nil
}

func (s *Service) Validate2FA(ctx context.Context, userID uuid.UUID, code string) (bool, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return false, err
	}

	if !user.TwoFactorEnabled || user.TwoFactorSecret == nil {
		return false, Err2FANotEnabled
	}

	return totp.Validate(code, *user.TwoFactorSecret), nil
}

func (s *Service) Disable2FA(ctx context.Context, userID uuid.UUID, code string) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	if !user.TwoFactorEnabled || user.TwoFactorSecret == nil {
		return Err2FANotEnabled
	}

	// Verify code before disabling
	if !totp.Validate(code, *user.TwoFactorSecret) {
		return ErrInvalid2FACode
	}

	user.TwoFactorSecret = nil
	user.TwoFactorEnabled = false
	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to disable 2FA: %w", err)
	}

	return nil
}
