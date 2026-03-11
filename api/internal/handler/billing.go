package handler

import (
	"errors"
	"io"

	billingService "github.com/deployer/api/internal/service/billing"
	"github.com/gofiber/fiber/v2"
)

type BillingHandler struct {
	billingSvc     *billingService.BillingService
	webhookSecret  string
}

func NewBillingHandler(billingSvc *billingService.BillingService, webhookSecret string) *BillingHandler {
	return &BillingHandler{
		billingSvc:    billingSvc,
		webhookSecret: webhookSecret,
	}
}

// ListPlans handles GET /api/v1/billing/plans
func (h *BillingHandler) ListPlans(c *fiber.Ctx) error {
	plans, err := h.billingSvc.GetPlans(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get plans",
		})
	}

	return c.JSON(fiber.Map{
		"plans": plans,
	})
}

// GetSubscription handles GET /api/v1/billing/subscription
func (h *BillingHandler) GetSubscription(c *fiber.Ctx) error {
	userID := getUserID(c)

	plan, sub, err := h.billingSvc.GetCurrentPlan(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get subscription",
		})
	}

	return c.JSON(fiber.Map{
		"plan":         plan,
		"subscription": sub,
	})
}

type subscribeRequest struct {
	Plan string `json:"plan"`
}

// Subscribe handles POST /api/v1/billing/subscribe
func (h *BillingHandler) Subscribe(c *fiber.Ctx) error {
	userID := getUserID(c)

	var req subscribeRequest
	if err := c.BodyParser(&req); err != nil || req.Plan == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "plan is required",
		})
	}

	sub, clientSecret, err := h.billingSvc.Subscribe(c.Context(), userID, req.Plan)
	if err != nil {
		if errors.Is(err, billingService.ErrNoPlan) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "plan not found",
			})
		}
		if errors.Is(err, billingService.ErrFreePlan) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "cannot subscribe to free plan via stripe",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to subscribe: " + err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"subscription":  sub,
		"client_secret": clientSecret,
	})
}

type changePlanRequest struct {
	Plan string `json:"plan"`
}

// ChangePlan handles POST /api/v1/billing/change-plan
func (h *BillingHandler) ChangePlan(c *fiber.Ctx) error {
	userID := getUserID(c)

	var req changePlanRequest
	if err := c.BodyParser(&req); err != nil || req.Plan == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "plan is required",
		})
	}

	if err := h.billingSvc.ChangePlan(c.Context(), userID, req.Plan); err != nil {
		if errors.Is(err, billingService.ErrNoPlan) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "plan not found",
			})
		}
		if errors.Is(err, billingService.ErrNoSubscription) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "no active subscription",
			})
		}
		if errors.Is(err, billingService.ErrAlreadyOnPlan) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "already on this plan",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to change plan: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "plan changed",
	})
}

// Cancel handles POST /api/v1/billing/cancel
func (h *BillingHandler) Cancel(c *fiber.Ctx) error {
	userID := getUserID(c)

	if err := h.billingSvc.CancelSubscription(c.Context(), userID); err != nil {
		if errors.Is(err, billingService.ErrNoSubscription) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "no active subscription",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to cancel subscription",
		})
	}

	return c.JSON(fiber.Map{
		"message": "subscription will be canceled at end of billing period",
	})
}

// Resume handles POST /api/v1/billing/resume
func (h *BillingHandler) Resume(c *fiber.Ctx) error {
	userID := getUserID(c)

	if err := h.billingSvc.ResumeSubscription(c.Context(), userID); err != nil {
		if errors.Is(err, billingService.ErrNoSubscription) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "no active subscription",
			})
		}
		if errors.Is(err, billingService.ErrNotCanceled) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "subscription is not canceled",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to resume subscription",
		})
	}

	return c.JSON(fiber.Map{
		"message": "subscription resumed",
	})
}

// Portal handles GET /api/v1/billing/portal
func (h *BillingHandler) Portal(c *fiber.Ctx) error {
	userID := getUserID(c)

	url, err := h.billingSvc.GetBillingPortalURL(c.Context(), userID)
	if err != nil {
		if errors.Is(err, billingService.ErrNoSubscription) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "no active subscription",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to create billing portal session",
		})
	}

	return c.JSON(fiber.Map{
		"url": url,
	})
}

// Invoices handles GET /api/v1/billing/invoices
func (h *BillingHandler) Invoices(c *fiber.Ctx) error {
	userID := getUserID(c)

	invoices, err := h.billingSvc.GetInvoices(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get invoices",
		})
	}

	return c.JSON(fiber.Map{
		"invoices": invoices,
	})
}

// Usage handles GET /api/v1/billing/usage
func (h *BillingHandler) Usage(c *fiber.Ctx) error {
	userID := getUserID(c)

	summary, err := h.billingSvc.GetUsageSummary(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get usage summary",
		})
	}

	return c.JSON(fiber.Map{
		"usage": summary,
	})
}

type updateBillingAddressRequest struct {
	Country    string `json:"country"`
	PostalCode string `json:"postal_code"`
	VATNumber  string `json:"vat_number"`
}

// UpdateBillingAddress handles PUT /api/v1/billing/address
func (h *BillingHandler) UpdateBillingAddress(c *fiber.Ctx) error {
	userID := getUserID(c)

	var req updateBillingAddressRequest
	if err := c.BodyParser(&req); err != nil || req.Country == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "country is required",
		})
	}

	if err := h.billingSvc.UpdateBillingAddress(c.Context(), userID, req.Country, req.PostalCode, req.VATNumber); err != nil {
		if errors.Is(err, billingService.ErrNoSubscription) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "no active subscription",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to update billing address: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "billing address updated",
	})
}

// StripeWebhook handles POST /api/v1/webhooks/stripe
func (h *BillingHandler) StripeWebhook(c *fiber.Ctx) error {
	payload, err := io.ReadAll(c.Request().BodyStream())
	if err != nil {
		// If BodyStream is nil or empty, fall back to Body()
		payload = c.Body()
	}
	if len(payload) == 0 {
		payload = c.Body()
	}

	signature := c.Get("Stripe-Signature")
	if signature == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "missing stripe signature",
		})
	}

	if err := h.billingSvc.HandleWebhook(payload, signature, h.webhookSecret); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "webhook processing failed: " + err.Error(),
		})
	}

	return c.SendStatus(fiber.StatusOK)
}
