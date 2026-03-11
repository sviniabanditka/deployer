package model

import (
	"time"

	"github.com/google/uuid"
)

type Plan struct {
	ID              uuid.UUID `json:"id"`
	Name            string    `json:"name"`
	DisplayName     string    `json:"display_name"`
	PriceCents      int       `json:"price_cents"`
	AppLimit        int       `json:"app_limit"`
	DBLimit         int       `json:"db_limit"`
	MemoryLimit     int64     `json:"memory_limit"`
	CPULimit        float64   `json:"cpu_limit"`
	StorageLimit    int64     `json:"storage_limit"`
	CustomDomains   bool      `json:"custom_domains"`
	PrioritySupport bool      `json:"priority_support"`
	StripePriceID   *string   `json:"stripe_price_id,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
}

// IsUnlimited returns true if the given limit value represents unlimited (-1).
func (p *Plan) IsUnlimited(limit int) bool {
	return limit < 0
}

// CanCreateApp returns true if the user has not exceeded the app limit.
func (p *Plan) CanCreateApp(currentCount int) bool {
	if p.IsUnlimited(p.AppLimit) {
		return true
	}
	return currentCount < p.AppLimit
}

// CanCreateDB returns true if the user has not exceeded the database limit.
func (p *Plan) CanCreateDB(currentCount int) bool {
	if p.IsUnlimited(p.DBLimit) {
		return true
	}
	return currentCount < p.DBLimit
}

type Subscription struct {
	ID                   uuid.UUID  `json:"id"`
	UserID               uuid.UUID  `json:"user_id"`
	PlanID               uuid.UUID  `json:"plan_id"`
	Status               string     `json:"status"`
	StripeCustomerID     *string    `json:"stripe_customer_id,omitempty"`
	StripeSubscriptionID *string    `json:"stripe_subscription_id,omitempty"`
	CurrentPeriodStart   *time.Time `json:"current_period_start,omitempty"`
	CurrentPeriodEnd     *time.Time `json:"current_period_end,omitempty"`
	CancelAtPeriodEnd    bool       `json:"cancel_at_period_end"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

type Invoice struct {
	ID              uuid.UUID `json:"id"`
	UserID          uuid.UUID `json:"user_id"`
	StripeInvoiceID *string   `json:"stripe_invoice_id,omitempty"`
	AmountCents     int       `json:"amount_cents"`
	Currency        string    `json:"currency"`
	Status          string    `json:"status"`
	Description     *string   `json:"description,omitempty"`
	InvoiceURL      *string   `json:"invoice_url,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
}

type UsageRecord struct {
	ID           uuid.UUID  `json:"id"`
	UserID       uuid.UUID  `json:"user_id"`
	ResourceType string     `json:"resource_type"`
	ResourceID   *uuid.UUID `json:"resource_id,omitempty"`
	Quantity     int64      `json:"quantity"`
	Unit         string     `json:"unit"`
	RecordedAt   time.Time  `json:"recorded_at"`
}

type UsageSummary struct {
	AppCount    int   `json:"app_count"`
	AppLimit    int   `json:"app_limit"`
	DBCount     int   `json:"db_count"`
	DBLimit     int   `json:"db_limit"`
	StorageUsed int64 `json:"storage_used"`
	StorageMax  int64 `json:"storage_max"`
}

type BillingAddress struct {
	ID         uuid.UUID `json:"id"`
	UserID     uuid.UUID `json:"user_id"`
	Country    string    `json:"country"`
	PostalCode string    `json:"postal_code"`
	VATNumber  string    `json:"vat_number,omitempty"`
	TaxExempt  bool      `json:"tax_exempt"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
