package billing

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/deployer/api/internal/model"
	"github.com/deployer/api/internal/repository"
	stripe "github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/webhook"
)

func (s *BillingService) HandleWebhook(payload []byte, signature string, webhookSecret string) error {
	event, err := webhook.ConstructEvent(payload, signature, webhookSecret)
	if err != nil {
		return fmt.Errorf("failed to verify webhook signature: %w", err)
	}

	ctx := context.Background()

	switch event.Type {
	case "invoice.paid":
		return s.handleInvoicePaid(ctx, event)
	case "invoice.payment_failed":
		return s.handleInvoicePaymentFailed(ctx, event)
	case "customer.subscription.updated":
		return s.handleSubscriptionUpdated(ctx, event)
	case "customer.subscription.deleted":
		return s.handleSubscriptionDeleted(ctx, event)
	default:
		log.Printf("unhandled stripe event type: %s", event.Type)
	}

	return nil
}

// getSubscriptionIDFromInvoice extracts the subscription ID from a stripe Invoice.
func getSubscriptionIDFromInvoice(invoice *stripe.Invoice) string {
	if invoice.Parent != nil &&
		invoice.Parent.SubscriptionDetails != nil &&
		invoice.Parent.SubscriptionDetails.Subscription != nil {
		return invoice.Parent.SubscriptionDetails.Subscription.ID
	}
	return ""
}

func (s *BillingService) handleInvoicePaid(ctx context.Context, event stripe.Event) error {
	var invoice stripe.Invoice
	if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
		return fmt.Errorf("failed to unmarshal invoice: %w", err)
	}

	subID := getSubscriptionIDFromInvoice(&invoice)
	if subID == "" {
		return nil
	}

	sub, err := s.billingRepo.GetSubscriptionByStripeID(ctx, subID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			log.Printf("subscription not found for stripe sub %s", subID)
			return nil
		}
		return err
	}

	// Ensure subscription is active
	if sub.Status != "active" {
		sub.Status = "active"
		if err := s.billingRepo.UpdateSubscription(ctx, sub); err != nil {
			return fmt.Errorf("failed to update subscription status: %w", err)
		}
	}

	// Record the invoice
	stripeInvoiceID := invoice.ID
	amountCents := int(invoice.AmountPaid)
	status := "paid"
	description := "Payment for subscription"
	var invoiceURL *string
	if invoice.HostedInvoiceURL != "" {
		invoiceURL = &invoice.HostedInvoiceURL
	}

	inv := &model.Invoice{
		UserID:          sub.UserID,
		StripeInvoiceID: &stripeInvoiceID,
		AmountCents:     amountCents,
		Currency:        string(invoice.Currency),
		Status:          status,
		Description:     &description,
		InvoiceURL:      invoiceURL,
	}

	if err := s.billingRepo.CreateInvoice(ctx, inv); err != nil {
		return fmt.Errorf("failed to create invoice record: %w", err)
	}

	return nil
}

func (s *BillingService) handleInvoicePaymentFailed(ctx context.Context, event stripe.Event) error {
	var invoice stripe.Invoice
	if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
		return fmt.Errorf("failed to unmarshal invoice: %w", err)
	}

	subID := getSubscriptionIDFromInvoice(&invoice)
	if subID == "" {
		return nil
	}

	sub, err := s.billingRepo.GetSubscriptionByStripeID(ctx, subID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			log.Printf("subscription not found for stripe sub %s", subID)
			return nil
		}
		return err
	}

	// Mark subscription as past_due
	sub.Status = "past_due"
	if err := s.billingRepo.UpdateSubscription(ctx, sub); err != nil {
		return fmt.Errorf("failed to update subscription status: %w", err)
	}

	// Record the failed invoice
	stripeInvoiceID := invoice.ID
	amountCents := int(invoice.AmountDue)
	status := "failed"
	description := "Failed payment for subscription"

	inv := &model.Invoice{
		UserID:          sub.UserID,
		StripeInvoiceID: &stripeInvoiceID,
		AmountCents:     amountCents,
		Currency:        string(invoice.Currency),
		Status:          status,
		Description:     &description,
	}

	if err := s.billingRepo.CreateInvoice(ctx, inv); err != nil {
		return fmt.Errorf("failed to create invoice record: %w", err)
	}

	return nil
}

func (s *BillingService) handleSubscriptionUpdated(ctx context.Context, event stripe.Event) error {
	var stripeSub stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &stripeSub); err != nil {
		return fmt.Errorf("failed to unmarshal subscription: %w", err)
	}

	sub, err := s.billingRepo.GetSubscriptionByStripeID(ctx, stripeSub.ID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			log.Printf("subscription not found for stripe sub %s", stripeSub.ID)
			return nil
		}
		return err
	}

	// Sync status
	sub.Status = string(stripeSub.Status)
	sub.CancelAtPeriodEnd = stripeSub.CancelAtPeriodEnd

	// Use the latest invoice period if available
	if stripeSub.LatestInvoice != nil {
		periodStart := time.Unix(stripeSub.LatestInvoice.PeriodStart, 0).UTC()
		periodEnd := time.Unix(stripeSub.LatestInvoice.PeriodEnd, 0).UTC()
		sub.CurrentPeriodStart = &periodStart
		sub.CurrentPeriodEnd = &periodEnd
	}

	if err := s.billingRepo.UpdateSubscription(ctx, sub); err != nil {
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	return nil
}

func (s *BillingService) handleSubscriptionDeleted(ctx context.Context, event stripe.Event) error {
	var stripeSub stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &stripeSub); err != nil {
		return fmt.Errorf("failed to unmarshal subscription: %w", err)
	}

	sub, err := s.billingRepo.GetSubscriptionByStripeID(ctx, stripeSub.ID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			log.Printf("subscription not found for stripe sub %s", stripeSub.ID)
			return nil
		}
		return err
	}

	// Downgrade to free plan
	freePlan, err := s.billingRepo.GetPlanByName(ctx, "free")
	if err != nil {
		return fmt.Errorf("failed to get free plan: %w", err)
	}

	sub.PlanID = freePlan.ID
	sub.Status = "canceled"
	sub.CancelAtPeriodEnd = false
	emptyStr := ""
	sub.StripeSubscriptionID = &emptyStr

	if err := s.billingRepo.UpdateSubscription(ctx, sub); err != nil {
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	return nil
}
