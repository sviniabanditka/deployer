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

type AppRepository interface {
	Create(ctx context.Context, app *model.App) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.App, error)
	ListByUserID(ctx context.Context, userID uuid.UUID) ([]model.App, error)
	Update(ctx context.Context, app *model.App) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetBySlug(ctx context.Context, slug string) (*model.App, error)
	CreateDeployment(ctx context.Context, deployment *model.Deployment) error
	GetDeployment(ctx context.Context, deploymentID uuid.UUID) (*model.Deployment, error)
	UpdateDeploymentStatus(ctx context.Context, deploymentID uuid.UUID, status, buildLog string) error
	GetLatestDeployment(ctx context.Context, appID uuid.UUID) (*model.Deployment, error)
	ListDeployments(ctx context.Context, appID uuid.UUID) ([]model.Deployment, error)
	GetPreviewDeployment(ctx context.Context, appID uuid.UUID, prNumber int) (*model.Deployment, error)
	ListStalePreviewDeployments(ctx context.Context, olderThan time.Duration) ([]model.Deployment, error)
}

type appRepo struct {
	db *pgxpool.Pool
}

func NewAppRepository(db *pgxpool.Pool) AppRepository {
	return &appRepo{db: db}
}

func (r *appRepo) Create(ctx context.Context, app *model.App) error {
	app.ID = uuid.New()
	now := time.Now().UTC()
	app.CreatedAt = now
	app.UpdatedAt = now

	_, err := r.db.Exec(ctx,
		`INSERT INTO apps (id, user_id, name, slug, status, runtime, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		app.ID, app.UserID, app.Name, app.Slug, app.Status, app.Runtime, app.CreatedAt, app.UpdatedAt,
	)
	return err
}

func (r *appRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.App, error) {
	var a model.App
	err := r.db.QueryRow(ctx,
		`SELECT id, user_id, name, slug, status, runtime, created_at, updated_at
		 FROM apps WHERE id = $1`, id,
	).Scan(&a.ID, &a.UserID, &a.Name, &a.Slug, &a.Status, &a.Runtime, &a.CreatedAt, &a.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return &a, err
}

func (r *appRepo) ListByUserID(ctx context.Context, userID uuid.UUID) ([]model.App, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, user_id, name, slug, status, runtime, created_at, updated_at
		 FROM apps WHERE user_id = $1 ORDER BY created_at DESC`, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var apps []model.App
	for rows.Next() {
		var a model.App
		if err := rows.Scan(&a.ID, &a.UserID, &a.Name, &a.Slug, &a.Status, &a.Runtime, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, err
		}
		apps = append(apps, a)
	}
	return apps, rows.Err()
}

func (r *appRepo) Update(ctx context.Context, app *model.App) error {
	app.UpdatedAt = time.Now().UTC()
	_, err := r.db.Exec(ctx,
		`UPDATE apps SET name=$1, slug=$2, status=$3, runtime=$4, updated_at=$5 WHERE id=$6`,
		app.Name, app.Slug, app.Status, app.Runtime, app.UpdatedAt, app.ID,
	)
	return err
}

func (r *appRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM apps WHERE id = $1`, id)
	return err
}

func (r *appRepo) GetBySlug(ctx context.Context, slug string) (*model.App, error) {
	var a model.App
	err := r.db.QueryRow(ctx,
		`SELECT id, user_id, name, slug, status, runtime, created_at, updated_at
		 FROM apps WHERE slug = $1`, slug,
	).Scan(&a.ID, &a.UserID, &a.Name, &a.Slug, &a.Status, &a.Runtime, &a.CreatedAt, &a.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return &a, err
}

func (r *appRepo) CreateDeployment(ctx context.Context, d *model.Deployment) error {
	d.ID = uuid.New()
	d.CreatedAt = time.Now().UTC()

	_, err := r.db.Exec(ctx,
		`INSERT INTO deployments (id, app_id, version, status, image_tag, build_log, is_preview, preview_url, pull_request_number, pull_request_url, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		d.ID, d.AppID, d.Version, d.Status, d.ImageTag, d.BuildLog,
		d.IsPreview, d.PreviewURL, d.PullRequestNumber, d.PullRequestURL, d.CreatedAt,
	)
	return err
}

func (r *appRepo) GetDeployment(ctx context.Context, deploymentID uuid.UUID) (*model.Deployment, error) {
	var d model.Deployment
	err := r.db.QueryRow(ctx,
		`SELECT id, app_id, version, status, image_tag, build_log, is_preview, preview_url, pull_request_number, pull_request_url, created_at
		 FROM deployments WHERE id = $1`, deploymentID,
	).Scan(&d.ID, &d.AppID, &d.Version, &d.Status, &d.ImageTag, &d.BuildLog,
		&d.IsPreview, &d.PreviewURL, &d.PullRequestNumber, &d.PullRequestURL, &d.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return &d, err
}

func (r *appRepo) UpdateDeploymentStatus(ctx context.Context, deploymentID uuid.UUID, status, buildLog string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE deployments SET status = $1, build_log = $2 WHERE id = $3`,
		status, buildLog, deploymentID,
	)
	return err
}

func (r *appRepo) GetLatestDeployment(ctx context.Context, appID uuid.UUID) (*model.Deployment, error) {
	var d model.Deployment
	err := r.db.QueryRow(ctx,
		`SELECT id, app_id, version, status, image_tag, build_log, is_preview, preview_url, pull_request_number, pull_request_url, created_at
		 FROM deployments WHERE app_id = $1 ORDER BY version DESC LIMIT 1`, appID,
	).Scan(&d.ID, &d.AppID, &d.Version, &d.Status, &d.ImageTag, &d.BuildLog,
		&d.IsPreview, &d.PreviewURL, &d.PullRequestNumber, &d.PullRequestURL, &d.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return &d, err
}

func (r *appRepo) ListDeployments(ctx context.Context, appID uuid.UUID) ([]model.Deployment, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, app_id, version, status, image_tag, build_log, is_preview, preview_url, pull_request_number, pull_request_url, created_at
		 FROM deployments WHERE app_id = $1 ORDER BY version DESC`, appID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deployments []model.Deployment
	for rows.Next() {
		var d model.Deployment
		if err := rows.Scan(&d.ID, &d.AppID, &d.Version, &d.Status, &d.ImageTag, &d.BuildLog,
			&d.IsPreview, &d.PreviewURL, &d.PullRequestNumber, &d.PullRequestURL, &d.CreatedAt); err != nil {
			return nil, err
		}
		deployments = append(deployments, d)
	}
	return deployments, rows.Err()
}

// GetPreviewDeployment returns the active preview deployment for a given app and PR number.
func (r *appRepo) GetPreviewDeployment(ctx context.Context, appID uuid.UUID, prNumber int) (*model.Deployment, error) {
	var d model.Deployment
	err := r.db.QueryRow(ctx,
		`SELECT id, app_id, version, status, image_tag, build_log, is_preview, preview_url, pull_request_number, pull_request_url, created_at
		 FROM deployments WHERE app_id = $1 AND is_preview = true AND pull_request_number = $2 AND status != 'destroyed'
		 ORDER BY version DESC LIMIT 1`, appID, prNumber,
	).Scan(&d.ID, &d.AppID, &d.Version, &d.Status, &d.ImageTag, &d.BuildLog,
		&d.IsPreview, &d.PreviewURL, &d.PullRequestNumber, &d.PullRequestURL, &d.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return &d, err
}

// ListStalePreviewDeployments returns preview deployments older than the given duration.
func (r *appRepo) ListStalePreviewDeployments(ctx context.Context, olderThan time.Duration) ([]model.Deployment, error) {
	cutoff := time.Now().UTC().Add(-olderThan)
	rows, err := r.db.Query(ctx,
		`SELECT id, app_id, version, status, image_tag, build_log, is_preview, preview_url, pull_request_number, pull_request_url, created_at
		 FROM deployments WHERE is_preview = true AND status != 'destroyed' AND created_at < $1
		 ORDER BY created_at ASC`, cutoff,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deployments []model.Deployment
	for rows.Next() {
		var d model.Deployment
		if err := rows.Scan(&d.ID, &d.AppID, &d.Version, &d.Status, &d.ImageTag, &d.BuildLog,
			&d.IsPreview, &d.PreviewURL, &d.PullRequestNumber, &d.PullRequestURL, &d.CreatedAt); err != nil {
			return nil, err
		}
		deployments = append(deployments, d)
	}
	return deployments, rows.Err()
}
