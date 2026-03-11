package handler

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/deployer/api/internal/config"
	"github.com/deployer/api/internal/model"
	"github.com/deployer/api/internal/repository"
	appService "github.com/deployer/api/internal/service/app"
	billingService "github.com/deployer/api/internal/service/billing"
	buildService "github.com/deployer/api/internal/service/build"
	deployService "github.com/deployer/api/internal/service/deploy"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

type AppHandler struct {
	appSvc      *appService.Service
	deploySvc   *deployService.DeployService
	buildSvc    *buildService.BuildService
	appRepo     repository.AppRepository
	cfg         *config.Config
	validate    *validator.Validate
	asynqClient *asynq.Client
	billingSvc  *billingService.BillingService
}

func NewAppHandler(
	appSvc *appService.Service,
	deploySvc *deployService.DeployService,
	buildSvc *buildService.BuildService,
	appRepo repository.AppRepository,
	cfg *config.Config,
	asynqClient *asynq.Client,
	billingSvc *billingService.BillingService,
) *AppHandler {
	return &AppHandler{
		appSvc:      appSvc,
		deploySvc:   deploySvc,
		buildSvc:    buildSvc,
		appRepo:     appRepo,
		cfg:         cfg,
		validate:    validator.New(),
		asynqClient: asynqClient,
		billingSvc:  billingSvc,
	}
}

type createAppRequest struct {
	Name string `json:"name" validate:"required,min=1,max=100"`
}

type updateEnvVarsRequest struct {
	Vars map[string]string `json:"vars" validate:"required"`
}

func getUserID(c *fiber.Ctx) uuid.UUID {
	return c.Locals("userID").(uuid.UUID)
}

func (h *AppHandler) Create(c *fiber.Ctx) error {
	var req createAppRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if err := h.validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	userID := getUserID(c)

	// Check billing quota before creating
	if h.billingSvc != nil {
		if err := h.billingSvc.EnforceAppQuota(c.Context(), userID); err != nil {
			if errors.Is(err, billingService.ErrQuotaExceeded) {
				return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
					"error": "app limit reached for your current plan, please upgrade",
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to check quota",
			})
		}
	}

	app, err := h.appSvc.Create(c.Context(), userID, req.Name)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to create app",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"app": app,
	})
}

func (h *AppHandler) List(c *fiber.Ctx) error {
	userID := getUserID(c)
	apps, err := h.appSvc.List(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to list apps",
		})
	}

	if apps == nil {
		apps = []model.App{}
	}

	return c.JSON(fiber.Map{
		"apps": apps,
	})
}

func (h *AppHandler) Get(c *fiber.Ctx) error {
	userID := getUserID(c)
	appID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid app id",
		})
	}

	app, err := h.appSvc.Get(c.Context(), userID, appID)
	if err != nil {
		if errors.Is(err, appService.ErrAppNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "app not found",
			})
		}
		if errors.Is(err, appService.ErrForbidden) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "forbidden",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get app",
		})
	}

	return c.JSON(fiber.Map{
		"app": app,
	})
}

func (h *AppHandler) Delete(c *fiber.Ctx) error {
	userID := getUserID(c)
	appID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid app id",
		})
	}

	if err := h.appSvc.Delete(c.Context(), userID, appID); err != nil {
		if errors.Is(err, appService.ErrAppNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "app not found",
			})
		}
		if errors.Is(err, appService.ErrForbidden) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "forbidden",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to delete app",
		})
	}

	return c.JSON(fiber.Map{
		"message": "app deleted",
	})
}

func (h *AppHandler) Deploy(c *fiber.Ctx) error {
	userID := getUserID(c)
	appID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid app id",
		})
	}

	// Verify ownership
	app, err := h.appSvc.Get(c.Context(), userID, appID)
	if err != nil {
		if errors.Is(err, appService.ErrAppNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "app not found"})
		}
		if errors.Is(err, appService.ErrForbidden) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "forbidden"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to get app"})
	}

	// Accept multipart file upload (ZIP)
	file, err := c.FormFile("archive")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "archive file is required (multipart field 'archive')",
		})
	}

	// Ensure upload directory exists
	if err := os.MkdirAll(h.cfg.UploadDir, 0o755); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to create upload directory",
		})
	}

	// Save ZIP to uploads directory
	archiveName := fmt.Sprintf("%s-%s.zip", app.Slug, uuid.New().String()[:8])
	archivePath := filepath.Join(h.cfg.UploadDir, archiveName)
	if err := c.SaveFile(file, archivePath); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to save uploaded file",
		})
	}

	// Determine version (previous + 1)
	version := 1
	latestDeployment, err := h.appRepo.GetLatestDeployment(c.Context(), appID)
	if err == nil && latestDeployment != nil {
		version = latestDeployment.Version + 1
	}

	// Create deployment record
	deployment := &model.Deployment{
		AppID:    appID,
		Version:  version,
		Status:   "pending",
		ImageTag: fmt.Sprintf("%s/%s:%d", h.cfg.RegistryURL, app.Slug, version),
	}
	if err := h.appRepo.CreateDeployment(c.Context(), deployment); err != nil {
		os.Remove(archivePath)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to create deployment record",
		})
	}

	// Enqueue build task
	if err := h.buildSvc.EnqueueBuild(appID, deployment.ID, archivePath); err != nil {
		os.Remove(archivePath)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to enqueue build task",
		})
	}

	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
		"message":    "deployment queued",
		"deployment": deployment,
	})
}

func (h *AppHandler) Logs(c *fiber.Ctx) error {
	userID := getUserID(c)
	appID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid app id",
		})
	}

	app, err := h.appSvc.Get(c.Context(), userID, appID)
	if err != nil {
		if errors.Is(err, appService.ErrAppNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "app not found"})
		}
		if errors.Is(err, appService.ErrForbidden) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "forbidden"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to get app"})
	}

	// Store app in locals for the websocket handler.
	c.Locals("app", app)
	return c.Next()
}

func (h *AppHandler) LogsWS(conn *websocket.Conn) {
	app, ok := conn.Locals("app").(*model.App)
	if !ok {
		log.Println("logs websocket: app not found in locals")
		return
	}

	ctx := context.Background()

	logReader, err := h.deploySvc.GetLogs(ctx, app, true)
	if err != nil {
		_ = conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("error: %v", err)))
		return
	}
	defer logReader.Close()

	scanner := bufio.NewScanner(logReader)
	for scanner.Scan() {
		if err := conn.WriteMessage(websocket.TextMessage, scanner.Bytes()); err != nil {
			break
		}
	}
}

func (h *AppHandler) UpdateEnvVars(c *fiber.Ctx) error {
	userID := getUserID(c)
	appID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid app id",
		})
	}

	// Verify ownership
	_, err = h.appSvc.Get(c.Context(), userID, appID)
	if err != nil {
		if errors.Is(err, appService.ErrAppNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "app not found",
			})
		}
		if errors.Is(err, appService.ErrForbidden) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "forbidden",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to verify app ownership",
		})
	}

	var req updateEnvVarsRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if err := h.validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	if err := h.appSvc.UpdateEnvVars(c.Context(), appID, req.Vars); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to update env vars",
		})
	}

	return c.JSON(fiber.Map{
		"message": "env vars updated",
	})
}

func (h *AppHandler) Stop(c *fiber.Ctx) error {
	userID := getUserID(c)
	appID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid app id",
		})
	}

	app, err := h.appSvc.Get(c.Context(), userID, appID)
	if err != nil {
		if errors.Is(err, appService.ErrAppNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "app not found"})
		}
		if errors.Is(err, appService.ErrForbidden) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "forbidden"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to get app"})
	}

	if err := h.deploySvc.Stop(c.Context(), app); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("failed to stop app: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"message": "app stopped",
		"app":     app,
	})
}

func (h *AppHandler) Start(c *fiber.Ctx) error {
	userID := getUserID(c)
	appID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid app id",
		})
	}

	app, err := h.appSvc.Get(c.Context(), userID, appID)
	if err != nil {
		if errors.Is(err, appService.ErrAppNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "app not found"})
		}
		if errors.Is(err, appService.ErrForbidden) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "forbidden"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to get app"})
	}

	// Fetch the latest deployment for this app.
	deployment, err := h.appRepo.GetLatestDeployment(c.Context(), appID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "no deployment found for this app",
		})
	}

	if err := h.deploySvc.Deploy(c.Context(), app, deployment); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("failed to start app: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"message": "app started",
		"app":     app,
	})
}

type rollbackRequest struct {
	DeploymentID uuid.UUID `json:"deployment_id" validate:"required"`
}

func (h *AppHandler) Rollback(c *fiber.Ctx) error {
	userID := getUserID(c)
	appID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid app id",
		})
	}

	app, err := h.appSvc.Get(c.Context(), userID, appID)
	if err != nil {
		if errors.Is(err, appService.ErrAppNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "app not found"})
		}
		if errors.Is(err, appService.ErrForbidden) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "forbidden"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to get app"})
	}

	var req rollbackRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if err := h.validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	if err := h.deploySvc.Rollback(c.Context(), app, req.DeploymentID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("rollback failed: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"message": "rollback completed",
		"app":     app,
	})
}

func (h *AppHandler) Stats(c *fiber.Ctx) error {
	userID := getUserID(c)
	appID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid app id",
		})
	}

	app, err := h.appSvc.Get(c.Context(), userID, appID)
	if err != nil {
		if errors.Is(err, appService.ErrAppNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "app not found"})
		}
		if errors.Is(err, appService.ErrForbidden) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "forbidden"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to get app"})
	}

	stats, err := h.deploySvc.GetContainerStats(c.Context(), app)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("failed to get stats: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"stats": stats,
	})
}
