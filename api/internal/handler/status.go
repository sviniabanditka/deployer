package handler

import (
	"strconv"

	statusService "github.com/deployer/api/internal/service/status"
	"github.com/gofiber/fiber/v2"
)

// StatusHandler handles status page endpoints.
type StatusHandler struct {
	statusSvc *statusService.StatusService
}

// NewStatusHandler creates a new StatusHandler.
func NewStatusHandler(statusSvc *statusService.StatusService) *StatusHandler {
	return &StatusHandler{
		statusSvc: statusSvc,
	}
}

// GetSystemStatus handles GET /api/v1/status — public, returns system status.
func (h *StatusHandler) GetSystemStatus(c *fiber.Ctx) error {
	status, err := h.statusSvc.GetSystemStatus(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get system status",
		})
	}

	return c.JSON(status)
}

// GetIncidentHistory handles GET /api/v1/status/history — public, returns incident history.
func (h *StatusHandler) GetIncidentHistory(c *fiber.Ctx) error {
	limit := 50
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	incidents, err := h.statusSvc.GetIncidentHistory(c.Context(), limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get incident history",
		})
	}

	return c.JSON(fiber.Map{
		"incidents": incidents,
	})
}
