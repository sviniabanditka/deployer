package handler

import (
	"errors"
	"log"

	"github.com/deployer/api/internal/repository"
	appService "github.com/deployer/api/internal/service/app"
	gitService "github.com/deployer/api/internal/service/git"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type GitHandler struct {
	gitSvc   *gitService.Service
	appSvc   *appService.Service
	validate *validator.Validate
}

func NewGitHandler(gitSvc *gitService.Service, appSvc *appService.Service) *GitHandler {
	return &GitHandler{
		gitSvc:   gitSvc,
		appSvc:   appSvc,
		validate: validator.New(),
	}
}

type connectRepoRequest struct {
	Provider    string `json:"provider" validate:"required,oneof=github gitlab"`
	RepoURL     string `json:"repo_url" validate:"required,url"`
	Branch      string `json:"branch"`
	AccessToken string `json:"access_token" validate:"required"`
}

func (h *GitHandler) ConnectRepo(c *fiber.Ctx) error {
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
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "app not found"})
		}
		if errors.Is(err, appService.ErrForbidden) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "forbidden"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to get app"})
	}

	var req connectRepoRequest
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

	conn, err := h.gitSvc.ConnectRepo(c.Context(), appID, req.Provider, req.RepoURL, req.Branch, req.AccessToken)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to connect repository: " + err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"git_connection": conn,
	})
}

func (h *GitHandler) DisconnectRepo(c *fiber.Ctx) error {
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
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "app not found"})
		}
		if errors.Is(err, appService.ErrForbidden) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "forbidden"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to get app"})
	}

	if err := h.gitSvc.DisconnectRepo(c.Context(), appID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "no git connection found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to disconnect repository: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "repository disconnected",
	})
}

func (h *GitHandler) GetConnection(c *fiber.Ctx) error {
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
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "app not found"})
		}
		if errors.Is(err, appService.ErrForbidden) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "forbidden"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to get app"})
	}

	conn, err := h.gitSvc.GetConnection(c.Context(), appID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "no git connection found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get git connection",
		})
	}

	return c.JSON(fiber.Map{
		"git_connection": conn,
	})
}

func (h *GitHandler) GitHubWebhook(c *fiber.Ctx) error {
	signature := c.Get("X-Hub-Signature-256")
	if signature == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "missing signature header",
		})
	}

	eventType := c.Get("X-GitHub-Event")

	body := c.Body()
	// Make a copy of the body since Fiber may reuse the buffer
	payload := make([]byte, len(body))
	copy(payload, body)

	if err := h.gitSvc.HandleWebhook(c.Context(), "github", eventType, signature, payload); err != nil {
		log.Printf("github webhook error: %v", err)
		// Return 200 to avoid GitHub retrying
		return c.JSON(fiber.Map{
			"message": "webhook received but processing failed",
		})
	}

	return c.JSON(fiber.Map{
		"message": "webhook processed",
	})
}

func (h *GitHandler) GitLabWebhook(c *fiber.Ctx) error {
	token := c.Get("X-Gitlab-Token")
	if token == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "missing token header",
		})
	}

	eventType := c.Get("X-Gitlab-Event")

	body := c.Body()
	payload := make([]byte, len(body))
	copy(payload, body)

	if err := h.gitSvc.HandleWebhook(c.Context(), "gitlab", eventType, token, payload); err != nil {
		log.Printf("gitlab webhook error: %v", err)
		return c.JSON(fiber.Map{
			"message": "webhook received but processing failed",
		})
	}

	return c.JSON(fiber.Map{
		"message": "webhook processed",
	})
}

