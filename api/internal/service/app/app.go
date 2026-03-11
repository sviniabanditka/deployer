package app

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/deployer/api/internal/model"
	"github.com/deployer/api/internal/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrAppNotFound = errors.New("app not found")
	ErrForbidden   = errors.New("forbidden")
	ErrSlugTaken   = errors.New("slug already taken")
)

var slugRegex = regexp.MustCompile(`[^a-z0-9]+`)

type Service struct {
	appRepo repository.AppRepository
	db      *pgxpool.Pool
}

func NewService(appRepo repository.AppRepository, db *pgxpool.Pool) *Service {
	return &Service{
		appRepo: appRepo,
		db:      db,
	}
}

func (s *Service) DB() *pgxpool.Pool {
	return s.db
}

func generateSlug(name string) string {
	slug := strings.ToLower(strings.TrimSpace(name))
	slug = slugRegex.ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")
	if slug == "" {
		slug = "app"
	}
	return slug
}

func (s *Service) Create(ctx context.Context, userID uuid.UUID, name string) (*model.App, error) {
	slug := generateSlug(name)

	existing, _ := s.appRepo.GetBySlug(ctx, slug)
	if existing != nil {
		slug = slug + "-" + uuid.New().String()[:8]
	}

	app := &model.App{
		UserID:  userID,
		Name:    name,
		Slug:    slug,
		Status:  model.AppStatusCreated,
		Runtime: "auto",
	}

	if err := s.appRepo.Create(ctx, app); err != nil {
		return nil, err
	}

	return app, nil
}

func (s *Service) List(ctx context.Context, userID uuid.UUID) ([]model.App, error) {
	return s.appRepo.ListByUserID(ctx, userID)
}

func (s *Service) Get(ctx context.Context, userID, appID uuid.UUID) (*model.App, error) {
	app, err := s.appRepo.GetByID(ctx, appID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrAppNotFound
		}
		return nil, err
	}

	if app.UserID != userID {
		return nil, ErrForbidden
	}

	return app, nil
}

func (s *Service) Delete(ctx context.Context, userID, appID uuid.UUID) error {
	app, err := s.appRepo.GetByID(ctx, appID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrAppNotFound
		}
		return err
	}

	if app.UserID != userID {
		return ErrForbidden
	}

	// Delete env vars first
	_, _ = s.db.Exec(ctx, `DELETE FROM env_vars WHERE app_id = $1`, appID)

	return s.appRepo.Delete(ctx, appID)
}

func (s *Service) UpdateEnvVars(ctx context.Context, appID uuid.UUID, vars map[string]string) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Delete existing env vars for this app
	_, err = tx.Exec(ctx, `DELETE FROM env_vars WHERE app_id = $1`, appID)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	for key, value := range vars {
		_, err = tx.Exec(ctx,
			`INSERT INTO env_vars (id, app_id, key, value, created_at) VALUES ($1, $2, $3, $4, $5)`,
			uuid.New(), appID, key, value, now,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}
