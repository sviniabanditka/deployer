package handler

import (
	"errors"

	"github.com/deployer/api/internal/model"
	billingService "github.com/deployer/api/internal/service/billing"
	databaseService "github.com/deployer/api/internal/service/database"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type DatabaseHandler struct {
	dbSvc      *databaseService.DatabaseService
	webUI      *databaseService.DatabaseWebUI
	validate   *validator.Validate
	billingSvc *billingService.BillingService
}

func NewDatabaseHandler(dbSvc *databaseService.DatabaseService, webUI *databaseService.DatabaseWebUI, billingSvc *billingService.BillingService) *DatabaseHandler {
	return &DatabaseHandler{
		dbSvc:      dbSvc,
		webUI:      webUI,
		validate:   validator.New(),
		billingSvc: billingSvc,
	}
}

type createDatabaseRequest struct {
	Name    string     `json:"name" validate:"required,min=1,max=100"`
	Engine  string     `json:"engine" validate:"required"`
	Version string     `json:"version"`
	AppID   *uuid.UUID `json:"app_id"`
}

type linkDatabaseRequest struct {
	AppID uuid.UUID `json:"app_id" validate:"required"`
}

// Create handles POST /api/v1/databases
func (h *DatabaseHandler) Create(c *fiber.Ctx) error {
	userID := getUserID(c)

	var req createDatabaseRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	if err := h.validate.Struct(req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	if !model.ValidEngine(req.Engine) {
		return fiber.NewError(fiber.StatusBadRequest, "unsupported engine: "+req.Engine)
	}

	// Check billing quota before creating
	if h.billingSvc != nil {
		if err := h.billingSvc.EnforceDBQuota(c.Context(), userID); err != nil {
			if errors.Is(err, billingService.ErrQuotaExceeded) {
				return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
					"error": "database limit reached for your current plan, please upgrade",
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to check quota",
			})
		}
	}

	mdb, err := h.dbSvc.Create(c.Context(), userID, req.Name, req.Engine, req.Version, req.AppID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.Status(fiber.StatusCreated).JSON(mdb.ConnectionInfo())
}

// List handles GET /api/v1/databases
func (h *DatabaseHandler) List(c *fiber.Ctx) error {
	userID := getUserID(c)

	dbs, err := h.dbSvc.List(c.Context(), userID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(fiber.Map{"databases": dbs})
}

// Get handles GET /api/v1/databases/:id
func (h *DatabaseHandler) Get(c *fiber.Ctx) error {
	userID := getUserID(c)
	dbID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid database id")
	}

	mdb, err := h.dbSvc.GetConnectionInfo(c.Context(), userID, dbID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, err.Error())
	}

	return c.JSON(mdb.ConnectionInfo())
}

// Delete handles DELETE /api/v1/databases/:id
func (h *DatabaseHandler) Delete(c *fiber.Ctx) error {
	userID := getUserID(c)
	dbID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid database id")
	}

	if err := h.dbSvc.Delete(c.Context(), userID, dbID); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(fiber.Map{"message": "database deleted"})
}

// Stop handles POST /api/v1/databases/:id/stop
func (h *DatabaseHandler) Stop(c *fiber.Ctx) error {
	userID := getUserID(c)
	dbID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid database id")
	}

	if err := h.dbSvc.Stop(c.Context(), userID, dbID); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(fiber.Map{"message": "database stopped"})
}

// Start handles POST /api/v1/databases/:id/start
func (h *DatabaseHandler) Start(c *fiber.Ctx) error {
	userID := getUserID(c)
	dbID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid database id")
	}

	if err := h.dbSvc.Start(c.Context(), userID, dbID); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(fiber.Map{"message": "database started"})
}

// Link handles POST /api/v1/databases/:id/link
func (h *DatabaseHandler) Link(c *fiber.Ctx) error {
	userID := getUserID(c)
	dbID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid database id")
	}

	var req linkDatabaseRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	if err := h.validate.Struct(req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	if err := h.dbSvc.LinkToApp(c.Context(), userID, dbID, req.AppID); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(fiber.Map{"message": "database linked to app"})
}

// Unlink handles POST /api/v1/databases/:id/unlink
func (h *DatabaseHandler) Unlink(c *fiber.Ctx) error {
	userID := getUserID(c)
	dbID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid database id")
	}

	if err := h.dbSvc.UnlinkFromApp(c.Context(), userID, dbID); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(fiber.Map{"message": "database unlinked from app"})
}

// CreateBackup handles POST /api/v1/databases/:id/backups
func (h *DatabaseHandler) CreateBackup(c *fiber.Ctx) error {
	userID := getUserID(c)
	dbID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid database id")
	}

	backup, err := h.dbSvc.CreateBackup(c.Context(), userID, dbID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.Status(fiber.StatusCreated).JSON(backup)
}

// ListBackups handles GET /api/v1/databases/:id/backups
func (h *DatabaseHandler) ListBackups(c *fiber.Ctx) error {
	userID := getUserID(c)
	dbID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid database id")
	}

	backups, err := h.dbSvc.ListBackups(c.Context(), userID, dbID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(fiber.Map{"backups": backups})
}

// RestoreBackup handles POST /api/v1/databases/:id/backups/:backup_id/restore
func (h *DatabaseHandler) RestoreBackup(c *fiber.Ctx) error {
	userID := getUserID(c)
	dbID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid database id")
	}
	backupID, err := uuid.Parse(c.Params("backup_id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid backup id")
	}

	if err := h.dbSvc.RestoreBackup(c.Context(), userID, dbID, backupID); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(fiber.Map{"message": "backup restored"})
}

type updateBackupSettingsRequest struct {
	AutoBackup *bool   `json:"auto_backup"`
	Retention  *int    `json:"retention"`
	Schedule   *string `json:"schedule"`
}

// UpdateBackupSettings handles PUT /api/v1/databases/:id/backup-settings
func (h *DatabaseHandler) UpdateBackupSettings(c *fiber.Ctx) error {
	userID := getUserID(c)
	dbID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid database id")
	}

	var req updateBackupSettingsRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	if err := h.dbSvc.UpdateBackupSettings(c.Context(), userID, dbID, req.AutoBackup, req.Retention, req.Schedule); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(fiber.Map{"message": "backup settings updated"})
}

// StartWebUI handles POST /api/v1/databases/:id/webui — starts Adminer/web UI for the database.
func (h *DatabaseHandler) StartWebUI(c *fiber.Ctx) error {
	userID := getUserID(c)
	dbID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid database id")
	}

	if h.webUI == nil {
		return fiber.NewError(fiber.StatusServiceUnavailable, "web UI service not available")
	}

	proxyURL, err := h.webUI.StartAdminer(c.Context(), userID, dbID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"url":     proxyURL,
		"message": "web UI started (auto-stops after 30 minutes)",
	})
}

// StopWebUI handles DELETE /api/v1/databases/:id/webui — stops Adminer/web UI for the database.
func (h *DatabaseHandler) StopWebUI(c *fiber.Ctx) error {
	userID := getUserID(c)
	dbID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid database id")
	}

	if h.webUI == nil {
		return fiber.NewError(fiber.StatusServiceUnavailable, "web UI service not available")
	}

	if err := h.webUI.StopAdminer(c.Context(), userID, dbID); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(fiber.Map{"message": "web UI stopped"})
}
