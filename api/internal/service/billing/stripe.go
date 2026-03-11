package billing

import (
	"fmt"

	stripe "github.com/stripe/stripe-go/v82"
	billingportalsession "github.com/stripe/stripe-go/v82/billingportal/session"
	"github.com/stripe/stripe-go/v82/customer"
	"github.com/stripe/stripe-go/v82/subscription"
	"github.com/stripe/stripe-go/v82/taxid"
)

func createStripeCustomer(email, name string) (string, error) {
	params := &stripe.CustomerParams{
		Email: stripe.String(email),
		Name:  stripe.String(name),
	}
	c, err := customer.New(params)
	if err != nil {
		return "", fmt.Errorf("failed to create stripe customer: %w", err)
	}
	return c.ID, nil
}

func createStripeSubscription(customerID, priceID string) (subscriptionID, clientSecret string, err error) {
	params := &stripe.SubscriptionParams{
		Customer: stripe.String(customerID),
		Items: []*stripe.SubscriptionItemsParams{
			{
				Price: stripe.String(priceID),
			},
		},
		PaymentBehavior: stripe.String("default_incomplete"),
		AutomaticTax:    &stripe.SubscriptionAutomaticTaxParams{Enabled: stripe.Bool(true)},
	}
	params.AddExpand("latest_invoice.payment_intent")

	sub, err := subscription.New(params)
	if err != nil {
		return "", "", fmt.Errorf("failed to create stripe subscription: %w", err)
	}

	subscriptionID = sub.ID

	if sub.LatestInvoice != nil &&
		sub.LatestInvoice.ConfirmationSecret != nil {
		clientSecret = sub.LatestInvoice.ConfirmationSecret.ClientSecret
	}

	return subscriptionID, clientSecret, nil
}

func updateStripeSubscription(subscriptionID, newPriceID string) error {
	sub, err := subscription.Get(subscriptionID, nil)
	if err != nil {
		return fmt.Errorf("failed to get stripe subscription: %w", err)
	}

	if len(sub.Items.Data) == 0 {
		return fmt.Errorf("subscription has no items")
	}

	params := &stripe.SubscriptionParams{
		Items: []*stripe.SubscriptionItemsParams{
			{
				ID:    stripe.String(sub.Items.Data[0].ID),
				Price: stripe.String(newPriceID),
			},
		},
		ProrationBehavior: stripe.String("create_prorations"),
	}

	_, err = subscription.Update(subscriptionID, params)
	if err != nil {
		return fmt.Errorf("failed to update stripe subscription: %w", err)
	}
	return nil
}

func cancelStripeSubscription(subscriptionID string) error {
	params := &stripe.SubscriptionParams{
		CancelAtPeriodEnd: stripe.Bool(true),
	}
	_, err := subscription.Update(subscriptionID, params)
	if err != nil {
		return fmt.Errorf("failed to cancel stripe subscription: %w", err)
	}
	return nil
}

func resumeStripeSubscription(subscriptionID string) error {
	params := &stripe.SubscriptionParams{
		CancelAtPeriodEnd: stripe.Bool(false),
	}
	_, err := subscription.Update(subscriptionID, params)
	if err != nil {
		return fmt.Errorf("failed to resume stripe subscription: %w", err)
	}
	return nil
}

func createBillingPortalSession(customerID, returnURL string) (string, error) {
	params := &stripe.BillingPortalSessionParams{
		Customer:  stripe.String(customerID),
		ReturnURL: stripe.String(returnURL),
	}
	s, err := billingportalsession.New(params)
	if err != nil {
		return "", fmt.Errorf("failed to create billing portal session: %w", err)
	}
	return s.URL, nil
}

// enableStripeTax enables automatic tax calculation on a Stripe subscription.
func enableStripeTax(subscriptionID string) error {
	params := &stripe.SubscriptionParams{
		AutomaticTax: &stripe.SubscriptionAutomaticTaxParams{Enabled: stripe.Bool(true)},
	}
	_, err := subscription.Update(subscriptionID, params)
	if err != nil {
		return fmt.Errorf("failed to enable automatic tax: %w", err)
	}
	return nil
}

// updateCustomerAddress updates the billing address on a Stripe customer.
func updateCustomerAddress(customerID, countryCode, postalCode string) error {
	params := &stripe.CustomerParams{
		Address: &stripe.AddressParams{
			Country:    stripe.String(countryCode),
			PostalCode: stripe.String(postalCode),
		},
	}
	_, err := customer.Update(customerID, params)
	if err != nil {
		return fmt.Errorf("failed to update customer address: %w", err)
	}
	return nil
}

// addCustomerTaxID adds a VAT tax ID to a Stripe customer.
func addCustomerTaxID(customerID, vatNumber string) error {
	params := &stripe.TaxIDParams{
		Customer: stripe.String(customerID),
		Type:     stripe.String("eu_vat"),
		Value:    stripe.String(vatNumber),
	}
	_, err := taxid.New(params)
	if err != nil {
		return fmt.Errorf("failed to add tax ID: %w", err)
	}
	return nil
}
