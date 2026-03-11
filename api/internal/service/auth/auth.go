package auth

import (
	"context"
	"errors"
	"time"

	"github.com/deployer/api/internal/config"
	"github.com/deployer/api/internal/model"
	"github.com/deployer/api/internal/repository"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserExists         = errors.New("user already exists")
	ErrInvalidToken       = errors.New("invalid token")
	ErrAccountLocked      = errors.New("account is locked due to too many failed login attempts")
	ErrRequires2FA        = errors.New("two-factor authentication required")
)

const (
	maxLoginAttempts = 5
	lockDuration     = 15 * time.Minute
)

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type LoginResult struct {
	Tokens      *TokenPair `json:"tokens,omitempty"`
	Requires2FA bool       `json:"requires_2fa,omitempty"`
	TempToken   string     `json:"temp_token,omitempty"`
}

type Service struct {
	userRepo repository.UserRepository
	cfg      *config.Config
}

func NewService(userRepo repository.UserRepository, cfg *config.Config) *Service {
	return &Service{
		userRepo: userRepo,
		cfg:      cfg,
	}
}

func (s *Service) Register(ctx context.Context, email, password, name string) (*model.User, error) {
	existing, _ := s.userRepo.GetByEmail(ctx, email)
	if existing != nil {
		return nil, ErrUserExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Email:        email,
		PasswordHash: string(hash),
		Name:         name,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *Service) Login(ctx context.Context, email, password string) (*LoginResult, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	// Check if account is locked
	if user.LockedUntil != nil && user.LockedUntil.After(time.Now()) {
		return nil, ErrAccountLocked
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		// Increment login attempts on failure
		_ = s.userRepo.IncrementLoginAttempts(ctx, user.ID)
		if user.LoginAttempts+1 >= maxLoginAttempts {
			_ = s.userRepo.LockUser(ctx, user.ID, time.Now().Add(lockDuration))
		}
		return nil, ErrInvalidCredentials
	}

	// Reset login attempts on success
	_ = s.userRepo.ResetLoginAttempts(ctx, user.ID)

	// Update last login time
	now := time.Now().UTC()
	user.LastLoginAt = &now
	user.LoginAttempts = 0
	user.LockedUntil = nil
	_ = s.userRepo.Update(ctx, user)

	// If 2FA is enabled, return a temporary token that requires 2FA verification
	if user.TwoFactorEnabled {
		tempToken, err := s.generateTempToken(user.ID)
		if err != nil {
			return nil, err
		}
		return &LoginResult{
			Requires2FA: true,
			TempToken:   tempToken,
		}, nil
	}

	tokens, err := s.GenerateTokenPair(user.ID)
	if err != nil {
		return nil, err
	}

	return &LoginResult{Tokens: tokens}, nil
}

func (s *Service) generateTempToken(userID uuid.UUID) (string, error) {
	claims := jwt.MapClaims{
		"sub":  userID.String(),
		"type": "2fa_temp",
		"exp":  time.Now().Add(5 * time.Minute).Unix(),
		"iat":  time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.JWTSecret))
}

func (s *Service) ValidateTempToken(tokenStr string) (uuid.UUID, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(s.cfg.JWTSecret), nil
	})
	if err != nil {
		return uuid.Nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return uuid.Nil, ErrInvalidToken
	}

	tokenType, _ := claims["type"].(string)
	if tokenType != "2fa_temp" {
		return uuid.Nil, ErrInvalidToken
	}

	sub, ok := claims["sub"].(string)
	if !ok {
		return uuid.Nil, ErrInvalidToken
	}

	return uuid.Parse(sub)
}

func (s *Service) Verify2FALogin(ctx context.Context, tempToken, code string) (*TokenPair, error) {
	userID, err := s.ValidateTempToken(tempToken)
	if err != nil {
		return nil, ErrInvalidToken
	}

	valid, err := s.Validate2FA(ctx, userID, code)
	if err != nil {
		return nil, err
	}
	if !valid {
		return nil, ErrInvalidCredentials
	}

	return s.GenerateTokenPair(userID)
}

func (s *Service) GetUserByID(ctx context.Context, userID uuid.UUID) (*model.User, error) {
	return s.userRepo.GetByID(ctx, userID)
}

func (s *Service) ValidateToken(tokenStr string) (uuid.UUID, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(s.cfg.JWTSecret), nil
	})
	if err != nil {
		return uuid.Nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return uuid.Nil, ErrInvalidToken
	}

	sub, ok := claims["sub"].(string)
	if !ok {
		return uuid.Nil, ErrInvalidToken
	}

	userID, err := uuid.Parse(sub)
	if err != nil {
		return uuid.Nil, ErrInvalidToken
	}

	return userID, nil
}

func (s *Service) GenerateTokenPair(userID uuid.UUID) (*TokenPair, error) {
	accessClaims := jwt.MapClaims{
		"sub":  userID.String(),
		"type": "access",
		"exp":  time.Now().Add(15 * time.Minute).Unix(),
		"iat":  time.Now().Unix(),
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessStr, err := accessToken.SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		return nil, err
	}

	refreshClaims := jwt.MapClaims{
		"sub":  userID.String(),
		"type": "refresh",
		"exp":  time.Now().Add(7 * 24 * time.Hour).Unix(),
		"iat":  time.Now().Unix(),
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshStr, err := refreshToken.SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessStr,
		RefreshToken: refreshStr,
	}, nil
}
