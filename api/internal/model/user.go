package model

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID               uuid.UUID  `json:"id"`
	Email            string     `json:"email"`
	PasswordHash     string     `json:"-"`
	Name             string     `json:"name"`
	TwoFactorSecret  *string    `json:"-"`
	TwoFactorEnabled bool       `json:"two_factor_enabled"`
	EmailVerified    bool       `json:"email_verified"`
	LastLoginAt      *time.Time `json:"last_login_at,omitempty"`
	LoginAttempts    int        `json:"-"`
	LockedUntil      *time.Time `json:"-"`
	OAuthProvider    *string    `json:"oauth_provider,omitempty"`
	OAuthID          *string    `json:"oauth_id,omitempty"`
	AvatarURL        *string    `json:"avatar_url,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}
