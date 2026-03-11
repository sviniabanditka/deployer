package billing

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/deployer/api/internal/model"
	"github.com/deployer/api/internal/repository"
	"github.com/google/uuid"
	stripe "github.com/stripe/stripe-go/v82"
)

var (
	ErrNoPlan          = errors.New("plan not found")
	ErrNoSubscription  = errors.New("no active subscription")
	ErrAlreadyOnPlan   = errors.New("already on this plan")
	ErrFreePlan        = errors.New("cannot subscribe to free plan via stripe")
	ErrQuotaExceeded   = errors.New("quota exceeded")
	ErrNotCanceled     = errors.New("subscription is not canceled")
)

type BillingService struct {
	billingRepo    repository.BillingRepository
	userRepo       repository.UserRepository
	portalReturnURL string
}

func NewBillingService(
	billingRepo repository.BillingRepository,
	userRepo repository.UserRepository,
	stripeKey string,
	portalReturnURL string,
) *BillingService {
	stripe.Key = stripeKey
	return &BillingService{
		billingRepo:    billingRepo,
		userRepo:       userRepo,
		portalReturnURL: portalReturnURL,
	}
}

func (s *BillingService) GetPlans(ctx context.Context) ([]model.Plan, error) {
	plans, err := s.billingRepo.GetPlans(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get plans: %w", err)
	}
	if plans == nil {
		plans = []model.Plan{}
	}
	return plans, nil
}

func (s *BillingService) GetCurrentPlan(ctx context.Context, userID uuid.UUID) (*model.Plan, *model.Subscription, error) {
	sub, err := s.billingRepo.GetSubscriptionByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			// No subscription, return the free plan
			plan, planErr := s.billingRepo.GetPlanByName(ctx, "free")
			if planErr != nil {
				return nil, nil, fmt.Errorf("failed to get free plan: %w", planErr)
			}
			return plan, nil, nil
		}
		return nil, nil, fmt.Errorf("failed to get subscription: %w", err)
	}

	plan, err := s.billingRepo.GetPlanByID(ctx, sub.PlanID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get plan: %w", err)
	}

	return plan, sub, nil
}

func (s *BillingService) Subscribe(ctx context.Context, userID uuid.UUID, planName string) (*model.Subscription, string, error) {
	plan, err := s.billingRepo.GetPlanByName(ctx, planName)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, "", ErrNoPlan
		}
		return nil, "", fmt.Errorf("failed to get plan: %w", err)
	}

	if plan.Name == "free" {
		return nil, "", ErrFreePlan
	}

	if plan.StripePriceID == nil || *plan.StripePriceID == "" {
		return nil, "", fmt.Errorf("plan %s has no stripe price configured", planName)
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get user: %w", err)
	}

	// Check if already subscribed
	existing, err := s.billingRepo.GetSubscriptionByUserID(ctx, userID)
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return nil, "", fmt.Errorf("failed to check existing subscription: %w", err)
	}

	var customerID string
	if existing != nil && existing.StripeCustomerID != nil {
		customerID = *existing.StripeCustomerID
	} else {
		customerID, err = createStripeCustomer(user.Email, user.Name)
		if err != nil {
			return nil, "", err
		}
	}

	stripeSubID, clientSecret, err := createStripeSubscription(customerID, *plan.StripePriceID)
	if err != nil {
		return nil, "", err
	}

	now := time.Now().UTC()
	periodEnd := now.AddDate(0, 1, 0)

	sub := &model.Subscription{
		UserID:               userID,
		PlanID:               plan.ID,
		Status:               "active",
		StripeCustomerID:     &customerID,
		StripeSubscriptionID: &stripeSubID,
		CurrentPeriodStart:   &now,
		CurrentPeriodEnd:     &periodEnd,
		CancelAtPeriodEnd:    false,
	}

	if existing != nil {
		sub.ID = existing.ID
		sub.CreatedAt = existing.CreatedAt
		if err := s.billingRepo.UpdateSubscription(ctx, sub); err != nil {
			return nil, "", fmt.Errorf("failed to update subscription: %w", err)
		}
	} else {
		if err := s.billingRepo.CreateSubscription(ctx, sub); err != nil {
			return nil, "", fmt.Errorf("failed to create subscription: %w", err)
		}
	}

	return sub, clientSecret, nil
}

func (s *BillingService) ChangePlan(ctx context.Context, userID uuid.UUID, newPlanName string) error {
	newPlan, err := s.billingRepo.GetPlanByName(ctx, newPlanName)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrNoPlan
		}
		return fmt.Errorf("failed to get plan: %w", err)
	}

	sub, err := s.billingRepo.GetSubscriptionByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrNoSubscription
		}
		return fmt.Errorf("failed to get subscription: %w", err)
	}

	if sub.PlanID == newPlan.ID {
		return ErrAlreadyOnPlan
	}

	if newPlan.StripePriceID == nil || *newPlan.StripePriceID == "" {
		return fmt.Errorf("plan %s has no stripe price configured", newPlanName)
	}

	if sub.StripeSubscriptionID != nil && *sub.StripeSubscriptionID != "" {
		if err := updateStripeSubscription(*sub.StripeSubscriptionID, *newPlan.StripePriceID); err != nil {
			return err
		}
	}

	sub.PlanID = newPlan.ID
	sub.CancelAtPeriodEnd = false
	sub.Status = "active"

	if err := s.billingRepo.UpdateSubscription(ctx, sub); err != nil {
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	return nil
}

func (s *BillingService) CancelSubscription(ctx context.Context, userID uuid.UUID) error {
	sub, err := s.billingRepo.GetSubscriptionByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrNoSubscription
		}
		return fmt.Errorf("failed to get subscription: %w", err)
	}

	if sub.StripeSubscriptionID != nil && *sub.StripeSubscriptionID != "" {
		if err := cancelStripeSubscription(*sub.StripeSubscriptionID); err != nil {
			return err
		}
	}

	sub.CancelAtPeriodEnd = true
	sub.Status = "canceled"

	if err := s.billingRepo.UpdateSubscription(ctx, sub); err != nil {
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	return nil
}

func (s *BillingService) ResumeSubscription(ctx context.Context, userID uuid.UUID) error {
	sub, err := s.billingRepo.GetSubscriptionByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrNoSubscription
		}
		return fmt.Errorf("failed to get subscription: %w", err)
	}

	if !sub.CancelAtPeriodEnd && sub.Status != "canceled" {
		return ErrNotCanceled
	}

	if sub.StripeSubscriptionID != nil && *sub.StripeSubscriptionID != "" {
		if err := resumeStripeSubscription(*sub.StripeSubscriptionID); err != nil {
			return err
		}
	}

	sub.CancelAtPeriodEnd = false
	sub.Status = "active"

	if err := s.billingRepo.UpdateSubscription(ctx, sub); err != nil {
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	return nil
}

func (s *BillingService) GetBillingPortalURL(ctx context.Context, userID uuid.UUID) (string, error) {
	sub, err := s.billingRepo.GetSubscriptionByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return "", ErrNoSubscription
		}
		return "", fmt.Errorf("failed to get subscription: %w", err)
	}

	if sub.StripeCustomerID == nil || *sub.StripeCustomerID == "" {
		return "", fmt.Errorf("no stripe customer associated with subscription")
	}

	url, err := createBillingPortalSession(*sub.StripeCustomerID, s.portalReturnURL)
	if err != nil {
		return "", err
	}

	return url, nil
}

func (s *BillingService) CheckQuota(ctx context.Context, userID uuid.UUID, resourceType string) (bool, error) {
	plan, _, err := s.GetCurrentPlan(ctx, userID)
	if err != nil {
		return false, err
	}

	switch resourceType {
	case "app":
		count, err := s.billingRepo.CountAppsByUserID(ctx, userID)
		if err != nil {
			return false, fmt.Errorf("failed to count apps: %w", err)
		}
		return plan.CanCreateApp(count), nil
	case "database":
		count, err := s.billingRepo.CountDatabasesByUserID(ctx, userID)
		if err != nil {
			return false, fmt.Errorf("failed to count databases: %w", err)
		}
		return plan.CanCreateDB(count), nil
	default:
		return false, fmt.Errorf("unknown resource type: %s", resourceType)
	}
}

func (s *BillingService) UpdateBillingAddress(ctx context.Context, userID uuid.UUID, country, postalCode, vatNumber string) error {
	sub, err := s.billingRepo.GetSubscriptionByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrNoSubscription
		}
		return fmt.Errorf("failed to get subscription: %w", err)
	}

	if sub.StripeCustomerID == nil || *sub.StripeCustomerID == "" {
		return fmt.Errorf("no stripe customer associated with subscription")
	}

	// Update the customer address on Stripe.
	if err := updateCustomerAddress(*sub.StripeCustomerID, country, postalCode); err != nil {
		return fmt.Errorf("failed to update billing address: %w", err)
	}

	// If a VAT number is provided, add the tax ID to the Stripe customer.
	if vatNumber != "" {
		if err := addCustomerTaxID(*sub.StripeCustomerID, vatNumber); err != nil {
			return fmt.Errorf("failed to add VAT number: %w", err)
		}
	}

	// Enable automatic tax on the subscription if not already enabled.
	if sub.StripeSubscriptionID != nil && *sub.StripeSubscriptionID != "" {
		if err := enableStripeTax(*sub.StripeSubscriptionID); err != nil {
			return fmt.Errorf("failed to enable automatic tax: %w", err)
		}
	}

	// Save the billing address to local DB.
	addr := &model.BillingAddress{
		UserID:     userID,
		Country:    country,
		PostalCode: postalCode,
		VATNumber:  vatNumber,
		TaxExempt:  vatNumber != "",
	}

	if err := s.billingRepo.UpsertBillingAddress(ctx, addr); err != nil {
		return fmt.Errorf("failed to save billing address: %w", err)
	}

	return nil
}

func (s *BillingService) GetInvoices(ctx context.Context, userID uuid.UUID) ([]model.Invoice, error) {
	invoices, err := s.billingRepo.ListInvoices(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list invoices: %w", err)
	}
	if invoices == nil {
		invoices = []model.Invoice{}
	}
	return invoices, nil
}
