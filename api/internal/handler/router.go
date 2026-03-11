package handler

import (
	"github.com/deployer/api/internal/config"
	"github.com/deployer/api/internal/middleware"
	"github.com/deployer/api/internal/repository"
	appService "github.com/deployer/api/internal/service/app"
	authService "github.com/deployer/api/internal/service/auth"
	billingService "github.com/deployer/api/internal/service/billing"
	buildService "github.com/deployer/api/internal/service/build"
	databaseService "github.com/deployer/api/internal/service/database"
	deployService "github.com/deployer/api/internal/service/deploy"
	domainService "github.com/deployer/api/internal/service/domain"
	gitService "github.com/deployer/api/internal/service/git"
	statusService "github.com/deployer/api/internal/service/status"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
)

func SetupRoutes(
	app *fiber.App,
	authSvc *authService.Service,
	oauthSvc *authService.OAuthService,
	gdprSvc *authService.GDPRService,
	appSvc *appService.Service,
	deploySvc *deployService.DeployService,
	buildSvc *buildService.BuildService,
	gitSvc *gitService.Service,
	dbSvc *databaseService.DatabaseService,
	dbWebUI *databaseService.DatabaseWebUI,
	statusSvc *statusService.StatusService,
	domainSvc *domainService.DomainService,
	billingSvc *billingService.BillingService,
	appRepo repository.AppRepository,
	cfg *config.Config,
	asynqClient *asynq.Client,
	redisClient *redis.Client,
) {
	authHandler := NewAuthHandler(authSvc, oauthSvc, gdprSvc, cfg)
	appHandler := NewAppHandler(appSvc, deploySvc, buildSvc, appRepo, cfg, asynqClient, billingSvc)
	gitHandler := NewGitHandler(gitSvc, appSvc)
	databaseHandler := NewDatabaseHandler(dbSvc, dbWebUI, billingSvc)
	statusHandler := NewStatusHandler(statusSvc)
	domainHandler := NewDomainHandler(domainSvc)
	healthHandler := NewHealthHandler()
	billingHandler := NewBillingHandler(billingSvc, cfg.StripeWebhookSecret)

	// Security headers middleware
	app.Use(middleware.SecurityHeaders())

	// Prometheus metrics middleware
	app.Use(middleware.PrometheusMiddleware())

	api := app.Group("/api/v1")

	// Metrics endpoint
	api.Get("/metrics", middleware.MetricsHandler())

	// Health
	api.Get("/health", healthHandler.Health)

	// Status page (public - no auth required)
	api.Get("/status", statusHandler.GetSystemStatus)
	api.Get("/status/history", statusHandler.GetIncidentHistory)

	// Auth routes (public) with rate limiting
	auth := api.Group("/auth", middleware.PublicRateLimiter(redisClient))
	auth.Post("/register", authHandler.Register)
	auth.Post("/login", authHandler.Login)
	auth.Post("/refresh", authHandler.Refresh)
	auth.Post("/2fa/validate", authHandler.Validate2FALogin)

	// OAuth routes (public)
	auth.Get("/github", authHandler.GitHubLogin)
	auth.Get("/github/callback", authHandler.GitHubCallback)
	auth.Get("/gitlab", authHandler.GitLabLogin)
	auth.Get("/gitlab/callback", authHandler.GitLabCallback)

	// Auth routes (protected) - 2FA management and GDPR
	authProtected := api.Group("/auth", middleware.JWTAuth(authSvc))
	authProtected.Post("/2fa/enable", authHandler.Enable2FA)
	authProtected.Post("/2fa/verify", authHandler.Verify2FA)
	authProtected.Post("/2fa/disable", authHandler.Disable2FA)
	authProtected.Get("/export-data", authHandler.ExportData)
	authProtected.Delete("/account", authHandler.DeleteAccount)

	// Webhook routes (public - verified by signature/token)
	webhooks := api.Group("/webhooks")
	webhooks.Post("/github", gitHandler.GitHubWebhook)
	webhooks.Post("/gitlab", gitHandler.GitLabWebhook)
	webhooks.Post("/stripe", billingHandler.StripeWebhook)

	// App routes (protected) with authenticated rate limiting
	apps := api.Group("/apps", middleware.JWTAuth(authSvc), middleware.AuthenticatedRateLimiter(redisClient))
	apps.Post("/", appHandler.Create)
	apps.Get("/", appHandler.List)
	apps.Get("/:id", appHandler.Get)
	apps.Delete("/:id", appHandler.Delete)
	apps.Post("/:id/deploy", appHandler.Deploy)
	apps.Post("/:id/stop", appHandler.Stop)
	apps.Post("/:id/start", appHandler.Start)
	apps.Get("/:id/logs", appHandler.Logs, websocket.New(appHandler.LogsWS))
	apps.Put("/:id/env", appHandler.UpdateEnvVars)
	apps.Get("/:id/stats", appHandler.Stats)
	apps.Post("/:id/rollback", appHandler.Rollback)

	// Custom domain routes (protected)
	apps.Post("/:id/domains", domainHandler.AddDomain)
	apps.Get("/:id/domains", domainHandler.ListDomains)
	apps.Post("/:id/domains/:domainId/verify", domainHandler.VerifyDomain)
	apps.Delete("/:id/domains/:domainId", domainHandler.RemoveDomain)

	// Git integration routes (protected)
	apps.Post("/:id/git/connect", gitHandler.ConnectRepo)
	apps.Delete("/:id/git/disconnect", gitHandler.DisconnectRepo)
	apps.Get("/:id/git", gitHandler.GetConnection)

	// Database routes (protected) with authenticated rate limiting
	databases := api.Group("/databases", middleware.JWTAuth(authSvc), middleware.AuthenticatedRateLimiter(redisClient))
	databases.Post("/", databaseHandler.Create)
	databases.Get("/", databaseHandler.List)
	databases.Get("/:id", databaseHandler.Get)
	databases.Delete("/:id", databaseHandler.Delete)
	databases.Post("/:id/stop", databaseHandler.Stop)
	databases.Post("/:id/start", databaseHandler.Start)
	databases.Post("/:id/link", databaseHandler.Link)
	databases.Post("/:id/unlink", databaseHandler.Unlink)
	databases.Post("/:id/backups", databaseHandler.CreateBackup)
	databases.Get("/:id/backups", databaseHandler.ListBackups)
	databases.Post("/:id/backups/:backup_id/restore", databaseHandler.RestoreBackup)
	databases.Put("/:id/backup-settings", databaseHandler.UpdateBackupSettings)
	databases.Post("/:id/webui", databaseHandler.StartWebUI)
	databases.Delete("/:id/webui", databaseHandler.StopWebUI)

	// Billing routes (protected) with authenticated rate limiting
	billing := api.Group("/billing", middleware.JWTAuth(authSvc), middleware.AuthenticatedRateLimiter(redisClient))
	billing.Get("/plans", billingHandler.ListPlans)
	billing.Get("/subscription", billingHandler.GetSubscription)
	billing.Post("/subscribe", billingHandler.Subscribe)
	billing.Post("/change-plan", billingHandler.ChangePlan)
	billing.Post("/cancel", billingHandler.Cancel)
	billing.Post("/resume", billingHandler.Resume)
	billing.Get("/portal", billingHandler.Portal)
	billing.Get("/invoices", billingHandler.Invoices)
	billing.Get("/usage", billingHandler.Usage)
	billing.Put("/address", billingHandler.UpdateBillingAddress)
}
