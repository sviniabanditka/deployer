package handler

import (
	"errors"

	domainService "github.com/deployer/api/internal/service/domain"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type DomainHandler struct {
	domainSvc *domainService.DomainService
}

func NewDomainHandler(domainSvc *domainService.DomainService) *DomainHandler {
	return &DomainHandler{
		domainSvc: domainSvc,
	}
}

type addDomainRequest struct {
	Domain string `json:"domain"`
}

// AddDomain handles POST /api/v1/apps/:id/domains
func (h *DomainHandler) AddDomain(c *fiber.Ctx) error {
	userID := getUserID(c)

	appID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid app ID",
		})
	}

	var req addDomainRequest
	if err := c.BodyParser(&req); err != nil || req.Domain == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "domain is required",
		})
	}

	domain, err := h.domainSvc.AddDomain(c.Context(), userID, appID, req.Domain)
	if err != nil {
		if errors.Is(err, domainService.ErrInvalidDomain) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid domain format",
			})
		}
		if errors.Is(err, domainService.ErrDomainExists) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": "domain already exists",
			})
		}
		if errors.Is(err, domainService.ErrNotOwner) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "you do not own this app",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to add domain",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"domain": domain,
		"verification": fiber.Map{
			"type":   "TXT",
			"host":   "_deployer-verify." + domain.Domain,
			"value":  domain.VerificationToken,
		},
	})
}

// ListDomains handles GET /api/v1/apps/:id/domains
func (h *DomainHandler) ListDomains(c *fiber.Ctx) error {
	appID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid app ID",
		})
	}

	domains, err := h.domainSvc.ListDomains(c.Context(), appID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to list domains",
		})
	}

	return c.JSON(fiber.Map{
		"domains": domains,
	})
}

// VerifyDomain handles POST /api/v1/apps/:id/domains/:domainId/verify
func (h *DomainHandler) VerifyDomain(c *fiber.Ctx) error {
	userID := getUserID(c)

	appID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid app ID",
		})
	}

	domainID, err := uuid.Parse(c.Params("domainId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid domain ID",
		})
	}

	if err := h.domainSvc.VerifyDomain(c.Context(), userID, appID, domainID); err != nil {
		if errors.Is(err, domainService.ErrDomainNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "domain not found",
			})
		}
		if errors.Is(err, domainService.ErrNotOwner) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "you do not own this app",
			})
		}
		if errors.Is(err, domainService.ErrVerificationFail) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to verify domain",
		})
	}

	return c.JSON(fiber.Map{
		"message": "domain verified successfully",
	})
}

// RemoveDomain handles DELETE /api/v1/apps/:id/domains/:domainId
func (h *DomainHandler) RemoveDomain(c *fiber.Ctx) error {
	userID := getUserID(c)

	appID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid app ID",
		})
	}

	domainID, err := uuid.Parse(c.Params("domainId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid domain ID",
		})
	}

	if err := h.domainSvc.RemoveDomain(c.Context(), userID, appID, domainID); err != nil {
		if errors.Is(err, domainService.ErrDomainNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "domain not found",
			})
		}
		if errors.Is(err, domainService.ErrNotOwner) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "you do not own this app",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to remove domain",
		})
	}

	return c.JSON(fiber.Map{
		"message": "domain removed successfully",
	})
}
