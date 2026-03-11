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

type DomainRepository interface {
	Create(ctx context.Context, domain *model.CustomDomain) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.CustomDomain, error)
	GetByDomain(ctx context.Context, domain string) (*model.CustomDomain, error)
	ListByAppID(ctx context.Context, appID uuid.UUID) ([]model.CustomDomain, error)
	Update(ctx context.Context, domain *model.CustomDomain) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListVerifiedByAppID(ctx context.Context, appID uuid.UUID) ([]model.CustomDomain, error)
}

type domainRepo struct {
	db *pgxpool.Pool
}

func NewDomainRepository(db *pgxpool.Pool) DomainRepository {
	return &domainRepo{db: db}
}

func (r *domainRepo) Create(ctx context.Context, domain *model.CustomDomain) error {
	domain.ID = uuid.New()
	now := time.Now().UTC()
	domain.CreatedAt = now
	domain.UpdatedAt = now

	_, err := r.db.Exec(ctx,
		`INSERT INTO custom_domains (id, app_id, domain, verification_token, status, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		domain.ID, domain.AppID, domain.Domain, domain.VerificationToken,
		domain.Status, domain.CreatedAt, domain.UpdatedAt,
	)
	return err
}

func (r *domainRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.CustomDomain, error) {
	var d model.CustomDomain
	err := r.db.QueryRow(ctx,
		`SELECT id, app_id, domain, verification_token, status, created_at, updated_at
		 FROM custom_domains WHERE id = $1`, id,
	).Scan(&d.ID, &d.AppID, &d.Domain, &d.VerificationToken, &d.Status, &d.CreatedAt, &d.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return &d, err
}

func (r *domainRepo) GetByDomain(ctx context.Context, domain string) (*model.CustomDomain, error) {
	var d model.CustomDomain
	err := r.db.QueryRow(ctx,
		`SELECT id, app_id, domain, verification_token, status, created_at, updated_at
		 FROM custom_domains WHERE domain = $1`, domain,
	).Scan(&d.ID, &d.AppID, &d.Domain, &d.VerificationToken, &d.Status, &d.CreatedAt, &d.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return &d, err
}

func (r *domainRepo) ListByAppID(ctx context.Context, appID uuid.UUID) ([]model.CustomDomain, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, app_id, domain, verification_token, status, created_at, updated_at
		 FROM custom_domains WHERE app_id = $1 ORDER BY created_at DESC`, appID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var domains []model.CustomDomain
	for rows.Next() {
		var d model.CustomDomain
		if err := rows.Scan(&d.ID, &d.AppID, &d.Domain, &d.VerificationToken,
			&d.Status, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, err
		}
		domains = append(domains, d)
	}
	return domains, rows.Err()
}

func (r *domainRepo) ListVerifiedByAppID(ctx context.Context, appID uuid.UUID) ([]model.CustomDomain, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, app_id, domain, verification_token, status, created_at, updated_at
		 FROM custom_domains WHERE app_id = $1 AND status = 'verified' ORDER BY created_at ASC`, appID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var domains []model.CustomDomain
	for rows.Next() {
		var d model.CustomDomain
		if err := rows.Scan(&d.ID, &d.AppID, &d.Domain, &d.VerificationToken,
			&d.Status, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, err
		}
		domains = append(domains, d)
	}
	return domains, rows.Err()
}

func (r *domainRepo) Update(ctx context.Context, domain *model.CustomDomain) error {
	domain.UpdatedAt = time.Now().UTC()
	_, err := r.db.Exec(ctx,
		`UPDATE custom_domains SET domain=$1, verification_token=$2, status=$3, updated_at=$4
		 WHERE id=$5`,
		domain.Domain, domain.VerificationToken, domain.Status, domain.UpdatedAt, domain.ID,
	)
	return err
}

func (r *domainRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM custom_domains WHERE id = $1`, id)
	return err
}
