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

type GitRepository interface {
	Create(ctx context.Context, conn *model.GitConnection) error
	GetByAppID(ctx context.Context, appID uuid.UUID) (*model.GitConnection, error)
	GetByWebhookSecret(ctx context.Context, secret string) (*model.GitConnection, error)
	Update(ctx context.Context, conn *model.GitConnection) error
	Delete(ctx context.Context, appID uuid.UUID) error
}

type gitRepo struct {
	db *pgxpool.Pool
}

func NewGitRepository(db *pgxpool.Pool) GitRepository {
	return &gitRepo{db: db}
}

func (r *gitRepo) Create(ctx context.Context, conn *model.GitConnection) error {
	conn.ID = uuid.New()
	now := time.Now().UTC()
	conn.CreatedAt = now
	conn.UpdatedAt = now

	_, err := r.db.Exec(ctx,
		`INSERT INTO git_connections (id, app_id, provider, repo_url, repo_owner, repo_name, branch, auto_deploy, webhook_secret, access_token, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
		conn.ID, conn.AppID, conn.Provider, conn.RepoURL, conn.RepoOwner, conn.RepoName,
		conn.Branch, conn.AutoDeploy, conn.WebhookSecret, conn.AccessToken, conn.CreatedAt, conn.UpdatedAt,
	)
	return err
}

func (r *gitRepo) GetByAppID(ctx context.Context, appID uuid.UUID) (*model.GitConnection, error) {
	var c model.GitConnection
	err := r.db.QueryRow(ctx,
		`SELECT id, app_id, provider, repo_url, repo_owner, repo_name, branch, auto_deploy, webhook_secret, access_token, created_at, updated_at
		 FROM git_connections WHERE app_id = $1`, appID,
	).Scan(&c.ID, &c.AppID, &c.Provider, &c.RepoURL, &c.RepoOwner, &c.RepoName,
		&c.Branch, &c.AutoDeploy, &c.WebhookSecret, &c.AccessToken, &c.CreatedAt, &c.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return &c, err
}

func (r *gitRepo) GetByWebhookSecret(ctx context.Context, secret string) (*model.GitConnection, error) {
	var c model.GitConnection
	err := r.db.QueryRow(ctx,
		`SELECT id, app_id, provider, repo_url, repo_owner, repo_name, branch, auto_deploy, webhook_secret, access_token, created_at, updated_at
		 FROM git_connections WHERE webhook_secret = $1`, secret,
	).Scan(&c.ID, &c.AppID, &c.Provider, &c.RepoURL, &c.RepoOwner, &c.RepoName,
		&c.Branch, &c.AutoDeploy, &c.WebhookSecret, &c.AccessToken, &c.CreatedAt, &c.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return &c, err
}

func (r *gitRepo) Update(ctx context.Context, conn *model.GitConnection) error {
	conn.UpdatedAt = time.Now().UTC()
	_, err := r.db.Exec(ctx,
		`UPDATE git_connections SET provider=$1, repo_url=$2, repo_owner=$3, repo_name=$4, branch=$5, auto_deploy=$6, webhook_secret=$7, access_token=$8, updated_at=$9 WHERE app_id=$10`,
		conn.Provider, conn.RepoURL, conn.RepoOwner, conn.RepoName, conn.Branch,
		conn.AutoDeploy, conn.WebhookSecret, conn.AccessToken, conn.UpdatedAt, conn.AppID,
	)
	return err
}

func (r *gitRepo) Delete(ctx context.Context, appID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM git_connections WHERE app_id = $1`, appID)
	return err
}
