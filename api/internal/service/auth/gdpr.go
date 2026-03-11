package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type GDPRService struct {
	db       *pgxpool.Pool
	authSvc  *Service
}

func NewGDPRService(db *pgxpool.Pool, authSvc *Service) *GDPRService {
	return &GDPRService{
		db:      db,
		authSvc: authSvc,
	}
}

type exportedData struct {
	ExportedAt  time.Time       `json:"exported_at"`
	User        json.RawMessage `json:"user"`
	Apps        json.RawMessage `json:"apps"`
	Databases   json.RawMessage `json:"databases"`
	Deployments json.RawMessage `json:"deployments"`
	Billing     json.RawMessage `json:"billing"`
}

func (g *GDPRService) ExportUserData(ctx context.Context, userID uuid.UUID) ([]byte, error) {
	data := exportedData{
		ExportedAt: time.Now().UTC(),
	}

	// Export user profile
	row := g.db.QueryRow(ctx,
		`SELECT json_build_object(
			'id', id, 'email', email, 'name', name,
			'email_verified', email_verified, 'two_factor_enabled', two_factor_enabled,
			'created_at', created_at, 'updated_at', updated_at, 'last_login_at', last_login_at
		) FROM users WHERE id = $1`, userID)
	if err := row.Scan(&data.User); err != nil {
		return nil, fmt.Errorf("failed to export user data: %w", err)
	}

	// Export apps
	row = g.db.QueryRow(ctx,
		`SELECT COALESCE(json_agg(json_build_object(
			'id', id, 'name', name, 'slug', slug, 'status', status,
			'runtime', runtime, 'created_at', created_at
		)), '[]'::json) FROM apps WHERE user_id = $1`, userID)
	if err := row.Scan(&data.Apps); err != nil {
		return nil, fmt.Errorf("failed to export apps data: %w", err)
	}

	// Export databases
	row = g.db.QueryRow(ctx,
		`SELECT COALESCE(json_agg(json_build_object(
			'id', id, 'name', name, 'engine', engine, 'version', version,
			'status', status, 'created_at', created_at
		)), '[]'::json) FROM managed_databases WHERE user_id = $1`, userID)
	if err := row.Scan(&data.Databases); err != nil {
		return nil, fmt.Errorf("failed to export databases data: %w", err)
	}

	// Export deployments
	row = g.db.QueryRow(ctx,
		`SELECT COALESCE(json_agg(json_build_object(
			'id', d.id, 'app_id', d.app_id, 'version', d.version,
			'status', d.status, 'created_at', d.created_at
		)), '[]'::json) FROM deployments d
		 JOIN apps a ON a.id = d.app_id WHERE a.user_id = $1`, userID)
	if err := row.Scan(&data.Deployments); err != nil {
		return nil, fmt.Errorf("failed to export deployments data: %w", err)
	}

	// Export billing data
	row = g.db.QueryRow(ctx,
		`SELECT COALESCE(json_agg(json_build_object(
			'id', id, 'amount_cents', amount_cents, 'currency', currency,
			'status', status, 'description', description, 'created_at', created_at
		)), '[]'::json) FROM invoices WHERE user_id = $1`, userID)
	if err := row.Scan(&data.Billing); err != nil {
		return nil, fmt.Errorf("failed to export billing data: %w", err)
	}

	return json.MarshalIndent(data, "", "  ")
}

func (g *GDPRService) DeleteAccount(ctx context.Context, userID uuid.UUID) error {
	// Delete user record — cascading will handle apps, deployments, env_vars, etc.
	_, err := g.db.Exec(ctx, `DELETE FROM users WHERE id = $1`, userID)
	if err != nil {
		return fmt.Errorf("failed to delete user account: %w", err)
	}
	return nil
}

func (g *GDPRService) AnonymizeUser(ctx context.Context, userID uuid.UUID) error {
	anonymized := fmt.Sprintf("deleted-%s@anonymized.local", userID.String()[:8])
	_, err := g.db.Exec(ctx,
		`UPDATE users SET email = $1, name = 'Deleted User', password_hash = '',
		        two_factor_secret = NULL, two_factor_enabled = false, updated_at = NOW()
		 WHERE id = $2`,
		anonymized, userID,
	)
	if err != nil {
		return fmt.Errorf("failed to anonymize user: %w", err)
	}
	return nil
}
