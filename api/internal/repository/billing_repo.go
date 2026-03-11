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

type BillingRepository interface {
	GetPlans(ctx context.Context) ([]model.Plan, error)
	GetPlanByName(ctx context.Context, name string) (*model.Plan, error)
	GetPlanByID(ctx context.Context, id uuid.UUID) (*model.Plan, error)
	CreateSubscription(ctx context.Context, sub *model.Subscription) error
	GetSubscriptionByUserID(ctx context.Context, userID uuid.UUID) (*model.Subscription, error)
	UpdateSubscription(ctx context.Context, sub *model.Subscription) error
	GetSubscriptionByStripeID(ctx context.Context, stripeSubID string) (*model.Subscription, error)
	CreateInvoice(ctx context.Context, inv *model.Invoice) error
	ListInvoices(ctx context.Context, userID uuid.UUID) ([]model.Invoice, error)
	RecordUsage(ctx context.Context, record *model.UsageRecord) error
	CountAppsByUserID(ctx context.Context, userID uuid.UUID) (int, error)
	CountDatabasesByUserID(ctx context.Context, userID uuid.UUID) (int, error)
	SumDatabaseStorageByUserID(ctx context.Context, userID uuid.UUID) (int64, error)
	UpsertBillingAddress(ctx context.Context, addr *model.BillingAddress) error
	GetBillingAddress(ctx context.Context, userID uuid.UUID) (*model.BillingAddress, error)
}

type billingRepo struct {
	db *pgxpool.Pool
}

func NewBillingRepository(db *pgxpool.Pool) BillingRepository {
	return &billingRepo{db: db}
}

func (r *billingRepo) GetPlans(ctx context.Context) ([]model.Plan, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, name, display_name, price_cents, app_limit, db_limit,
		        memory_limit, cpu_limit, storage_limit, custom_domains,
		        priority_support, stripe_price_id, created_at
		 FROM plans ORDER BY price_cents ASC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var plans []model.Plan
	for rows.Next() {
		var p model.Plan
		if err := rows.Scan(
			&p.ID, &p.Name, &p.DisplayName, &p.PriceCents,
			&p.AppLimit, &p.DBLimit, &p.MemoryLimit, &p.CPULimit,
			&p.StorageLimit, &p.CustomDomains, &p.PrioritySupport,
			&p.StripePriceID, &p.CreatedAt,
		); err != nil {
			return nil, err
		}
		plans = append(plans, p)
	}
	return plans, rows.Err()
}

func (r *billingRepo) GetPlanByName(ctx context.Context, name string) (*model.Plan, error) {
	var p model.Plan
	err := r.db.QueryRow(ctx,
		`SELECT id, name, display_name, price_cents, app_limit, db_limit,
		        memory_limit, cpu_limit, storage_limit, custom_domains,
		        priority_support, stripe_price_id, created_at
		 FROM plans WHERE name = $1`, name,
	).Scan(
		&p.ID, &p.Name, &p.DisplayName, &p.PriceCents,
		&p.AppLimit, &p.DBLimit, &p.MemoryLimit, &p.CPULimit,
		&p.StorageLimit, &p.CustomDomains, &p.PrioritySupport,
		&p.StripePriceID, &p.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return &p, err
}

func (r *billingRepo) GetPlanByID(ctx context.Context, id uuid.UUID) (*model.Plan, error) {
	var p model.Plan
	err := r.db.QueryRow(ctx,
		`SELECT id, name, display_name, price_cents, app_limit, db_limit,
		        memory_limit, cpu_limit, storage_limit, custom_domains,
		        priority_support, stripe_price_id, created_at
		 FROM plans WHERE id = $1`, id,
	).Scan(
		&p.ID, &p.Name, &p.DisplayName, &p.PriceCents,
		&p.AppLimit, &p.DBLimit, &p.MemoryLimit, &p.CPULimit,
		&p.StorageLimit, &p.CustomDomains, &p.PrioritySupport,
		&p.StripePriceID, &p.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return &p, err
}

func (r *billingRepo) CreateSubscription(ctx context.Context, sub *model.Subscription) error {
	sub.ID = uuid.New()
	now := time.Now().UTC()
	sub.CreatedAt = now
	sub.UpdatedAt = now

	_, err := r.db.Exec(ctx,
		`INSERT INTO subscriptions (id, user_id, plan_id, status, stripe_customer_id,
		 stripe_subscription_id, current_period_start, current_period_end,
		 cancel_at_period_end, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		sub.ID, sub.UserID, sub.PlanID, sub.Status,
		sub.StripeCustomerID, sub.StripeSubscriptionID,
		sub.CurrentPeriodStart, sub.CurrentPeriodEnd,
		sub.CancelAtPeriodEnd, sub.CreatedAt, sub.UpdatedAt,
	)
	return err
}

func (r *billingRepo) GetSubscriptionByUserID(ctx context.Context, userID uuid.UUID) (*model.Subscription, error) {
	var s model.Subscription
	err := r.db.QueryRow(ctx,
		`SELECT id, user_id, plan_id, status, stripe_customer_id,
		        stripe_subscription_id, current_period_start, current_period_end,
		        cancel_at_period_end, created_at, updated_at
		 FROM subscriptions WHERE user_id = $1`, userID,
	).Scan(
		&s.ID, &s.UserID, &s.PlanID, &s.Status,
		&s.StripeCustomerID, &s.StripeSubscriptionID,
		&s.CurrentPeriodStart, &s.CurrentPeriodEnd,
		&s.CancelAtPeriodEnd, &s.CreatedAt, &s.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return &s, err
}

func (r *billingRepo) UpdateSubscription(ctx context.Context, sub *model.Subscription) error {
	sub.UpdatedAt = time.Now().UTC()
	_, err := r.db.Exec(ctx,
		`UPDATE subscriptions SET plan_id=$1, status=$2, stripe_customer_id=$3,
		 stripe_subscription_id=$4, current_period_start=$5, current_period_end=$6,
		 cancel_at_period_end=$7, updated_at=$8
		 WHERE id=$9`,
		sub.PlanID, sub.Status, sub.StripeCustomerID,
		sub.StripeSubscriptionID, sub.CurrentPeriodStart, sub.CurrentPeriodEnd,
		sub.CancelAtPeriodEnd, sub.UpdatedAt, sub.ID,
	)
	return err
}

func (r *billingRepo) GetSubscriptionByStripeID(ctx context.Context, stripeSubID string) (*model.Subscription, error) {
	var s model.Subscription
	err := r.db.QueryRow(ctx,
		`SELECT id, user_id, plan_id, status, stripe_customer_id,
		        stripe_subscription_id, current_period_start, current_period_end,
		        cancel_at_period_end, created_at, updated_at
		 FROM subscriptions WHERE stripe_subscription_id = $1`, stripeSubID,
	).Scan(
		&s.ID, &s.UserID, &s.PlanID, &s.Status,
		&s.StripeCustomerID, &s.StripeSubscriptionID,
		&s.CurrentPeriodStart, &s.CurrentPeriodEnd,
		&s.CancelAtPeriodEnd, &s.CreatedAt, &s.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return &s, err
}

func (r *billingRepo) CreateInvoice(ctx context.Context, inv *model.Invoice) error {
	inv.ID = uuid.New()
	inv.CreatedAt = time.Now().UTC()

	_, err := r.db.Exec(ctx,
		`INSERT INTO invoices (id, user_id, stripe_invoice_id, amount_cents, currency,
		 status, description, invoice_url, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		inv.ID, inv.UserID, inv.StripeInvoiceID, inv.AmountCents,
		inv.Currency, inv.Status, inv.Description, inv.InvoiceURL, inv.CreatedAt,
	)
	return err
}

func (r *billingRepo) ListInvoices(ctx context.Context, userID uuid.UUID) ([]model.Invoice, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, user_id, stripe_invoice_id, amount_cents, currency,
		        status, description, invoice_url, created_at
		 FROM invoices WHERE user_id = $1 ORDER BY created_at DESC`, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var invoices []model.Invoice
	for rows.Next() {
		var inv model.Invoice
		if err := rows.Scan(
			&inv.ID, &inv.UserID, &inv.StripeInvoiceID, &inv.AmountCents,
			&inv.Currency, &inv.Status, &inv.Description, &inv.InvoiceURL,
			&inv.CreatedAt,
		); err != nil {
			return nil, err
		}
		invoices = append(invoices, inv)
	}
	return invoices, rows.Err()
}

func (r *billingRepo) RecordUsage(ctx context.Context, record *model.UsageRecord) error {
	record.ID = uuid.New()
	record.RecordedAt = time.Now().UTC()

	_, err := r.db.Exec(ctx,
		`INSERT INTO usage_records (id, user_id, resource_type, resource_id, quantity, unit, recorded_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		record.ID, record.UserID, record.ResourceType, record.ResourceID,
		record.Quantity, record.Unit, record.RecordedAt,
	)
	return err
}

func (r *billingRepo) CountAppsByUserID(ctx context.Context, userID uuid.UUID) (int, error) {
	var count int
	err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM apps WHERE user_id = $1`, userID,
	).Scan(&count)
	return count, err
}

func (r *billingRepo) CountDatabasesByUserID(ctx context.Context, userID uuid.UUID) (int, error) {
	var count int
	err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM managed_databases WHERE user_id = $1 AND status != 'deleted'`, userID,
	).Scan(&count)
	return count, err
}

func (r *billingRepo) SumDatabaseStorageByUserID(ctx context.Context, userID uuid.UUID) (int64, error) {
	var total int64
	err := r.db.QueryRow(ctx,
		`SELECT COALESCE(SUM(storage_used), 0) FROM managed_databases WHERE user_id = $1 AND status != 'deleted'`, userID,
	).Scan(&total)
	return total, err
}

func (r *billingRepo) UpsertBillingAddress(ctx context.Context, addr *model.BillingAddress) error {
	now := time.Now().UTC()
	addr.UpdatedAt = now

	_, err := r.db.Exec(ctx,
		`INSERT INTO billing_addresses (id, user_id, country, postal_code, vat_number, tax_exempt, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		 ON CONFLICT (user_id) DO UPDATE SET
		   country = EXCLUDED.country,
		   postal_code = EXCLUDED.postal_code,
		   vat_number = EXCLUDED.vat_number,
		   tax_exempt = EXCLUDED.tax_exempt,
		   updated_at = EXCLUDED.updated_at`,
		uuid.New(), addr.UserID, addr.Country, addr.PostalCode,
		addr.VATNumber, addr.TaxExempt, now, now,
	)
	return err
}

func (r *billingRepo) GetBillingAddress(ctx context.Context, userID uuid.UUID) (*model.BillingAddress, error) {
	var a model.BillingAddress
	err := r.db.QueryRow(ctx,
		`SELECT id, user_id, country, postal_code, vat_number, tax_exempt, created_at, updated_at
		 FROM billing_addresses WHERE user_id = $1`, userID,
	).Scan(&a.ID, &a.UserID, &a.Country, &a.PostalCode, &a.VATNumber, &a.TaxExempt, &a.CreatedAt, &a.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return &a, err
}
