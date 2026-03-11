package repository

import (
	"context"
	"errors"
	"time"

	"github.com/deployer/api/internal/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("record not found")

type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	GetByOAuthID(ctx context.Context, provider, oauthID string) (*model.User, error)
	Update(ctx context.Context, user *model.User) error
	IncrementLoginAttempts(ctx context.Context, userID uuid.UUID) error
	ResetLoginAttempts(ctx context.Context, userID uuid.UUID) error
	LockUser(ctx context.Context, userID uuid.UUID, until time.Time) error
}

type userRepo struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) UserRepository {
	return &userRepo{db: db}
}

func (r *userRepo) Create(ctx context.Context, user *model.User) error {
	user.ID = uuid.New()
	now := time.Now().UTC()
	user.CreatedAt = now
	user.UpdatedAt = now

	_, err := r.db.Exec(ctx,
		`INSERT INTO users (id, email, password_hash, name, oauth_provider, oauth_id, avatar_url, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		user.ID, user.Email, user.PasswordHash, user.Name,
		user.OAuthProvider, user.OAuthID, user.AvatarURL,
		user.CreatedAt, user.UpdatedAt,
	)
	return err
}

func (r *userRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	var u model.User
	err := r.db.QueryRow(ctx,
		`SELECT id, email, password_hash, name, two_factor_secret, two_factor_enabled,
		        email_verified, last_login_at, login_attempts, locked_until,
		        oauth_provider, oauth_id, avatar_url, created_at, updated_at
		 FROM users WHERE id = $1`, id,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Name, &u.TwoFactorSecret, &u.TwoFactorEnabled,
		&u.EmailVerified, &u.LastLoginAt, &u.LoginAttempts, &u.LockedUntil,
		&u.OAuthProvider, &u.OAuthID, &u.AvatarURL, &u.CreatedAt, &u.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return &u, err
}

func (r *userRepo) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	var u model.User
	err := r.db.QueryRow(ctx,
		`SELECT id, email, password_hash, name, two_factor_secret, two_factor_enabled,
		        email_verified, last_login_at, login_attempts, locked_until,
		        oauth_provider, oauth_id, avatar_url, created_at, updated_at
		 FROM users WHERE email = $1`, email,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Name, &u.TwoFactorSecret, &u.TwoFactorEnabled,
		&u.EmailVerified, &u.LastLoginAt, &u.LoginAttempts, &u.LockedUntil,
		&u.OAuthProvider, &u.OAuthID, &u.AvatarURL, &u.CreatedAt, &u.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return &u, err
}

func (r *userRepo) GetByOAuthID(ctx context.Context, provider, oauthID string) (*model.User, error) {
	var u model.User
	err := r.db.QueryRow(ctx,
		`SELECT id, email, password_hash, name, two_factor_secret, two_factor_enabled,
		        email_verified, last_login_at, login_attempts, locked_until,
		        oauth_provider, oauth_id, avatar_url, created_at, updated_at
		 FROM users WHERE oauth_provider = $1 AND oauth_id = $2`, provider, oauthID,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Name, &u.TwoFactorSecret, &u.TwoFactorEnabled,
		&u.EmailVerified, &u.LastLoginAt, &u.LoginAttempts, &u.LockedUntil,
		&u.OAuthProvider, &u.OAuthID, &u.AvatarURL, &u.CreatedAt, &u.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return &u, err
}

func (r *userRepo) Update(ctx context.Context, user *model.User) error {
	user.UpdatedAt = time.Now().UTC()
	_, err := r.db.Exec(ctx,
		`UPDATE users SET email = $1, password_hash = $2, name = $3,
		        two_factor_secret = $4, two_factor_enabled = $5, email_verified = $6,
		        last_login_at = $7, login_attempts = $8, locked_until = $9,
		        oauth_provider = $10, oauth_id = $11, avatar_url = $12, updated_at = $13
		 WHERE id = $14`,
		user.Email, user.PasswordHash, user.Name, user.TwoFactorSecret, user.TwoFactorEnabled,
		user.EmailVerified, user.LastLoginAt, user.LoginAttempts, user.LockedUntil,
		user.OAuthProvider, user.OAuthID, user.AvatarURL, user.UpdatedAt, user.ID,
	)
	return err
}

func (r *userRepo) IncrementLoginAttempts(ctx context.Context, userID uuid.UUID) error {
	_, err := r.db.Exec(ctx,
		`UPDATE users SET login_attempts = login_attempts + 1, updated_at = NOW() WHERE id = $1`,
		userID,
	)
	return err
}

func (r *userRepo) ResetLoginAttempts(ctx context.Context, userID uuid.UUID) error {
	_, err := r.db.Exec(ctx,
		`UPDATE users SET login_attempts = 0, updated_at = NOW() WHERE id = $1`,
		userID,
	)
	return err
}

func (r *userRepo) LockUser(ctx context.Context, userID uuid.UUID, until time.Time) error {
	_, err := r.db.Exec(ctx,
		`UPDATE users SET locked_until = $1, updated_at = NOW() WHERE id = $2`,
		until, userID,
	)
	return err
}
