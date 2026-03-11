package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/deployer/api/internal/config"
	"github.com/deployer/api/internal/handler"
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
	"github.com/docker/docker/client"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

func main() {
	cfg := config.Load()

	// Connect to PostgreSQL
	ctx := context.Background()
	dbPool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer dbPool.Close()

	if err := dbPool.Ping(ctx); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}
	log.Println("connected to PostgreSQL")

	// Connect to Redis
	redisOpts, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		log.Fatalf("failed to parse redis URL: %v", err)
	}
	redisClient := redis.NewClient(redisOpts)
	defer redisClient.Close()

	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatalf("failed to ping redis: %v", err)
	}
	log.Println("connected to Redis")

	// Asynq client for task queue
	asynqClient := asynq.NewClient(asynq.RedisClientOpt{Addr: redisOpts.Addr, Password: redisOpts.Password, DB: redisOpts.DB})
	defer asynqClient.Close()
	log.Println("asynq client initialized")

	// Repositories
	userRepo := repository.NewUserRepository(dbPool)
	appRepo := repository.NewAppRepository(dbPool)
	gitRepo := repository.NewGitRepository(dbPool)
	databaseRepo := repository.NewDatabaseRepository(dbPool)
	billingRepo := repository.NewBillingRepository(dbPool)
	domainRepo := repository.NewDomainRepository(dbPool)

	// Docker client
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatalf("failed to create docker client: %v", err)
	}
	defer dockerClient.Close()
	log.Println("docker client initialized")

	// Services
	authSvc := authService.NewService(userRepo, cfg)
	oauthSvc := authService.NewOAuthService(userRepo, cfg)
	gdprSvc := authService.NewGDPRService(dbPool, authSvc)
	appSvc := appService.NewService(appRepo, dbPool)
	deploySvc := deployService.NewDeployService(dockerClient, appRepo, domainRepo, dbPool, cfg)
	previewSvc := deployService.NewPreviewService(dockerClient, appRepo, dbPool, cfg)
	buildSvc := buildService.NewBuildService(appRepo, cfg, dockerClient, asynqClient)
	gitSvc := gitService.NewService(gitRepo, appRepo, buildSvc, previewSvc, cfg)
	dbSvc := databaseService.NewDatabaseService(dockerClient, databaseRepo, appRepo, dbPool, cfg)
	dbWebUI := databaseService.NewDatabaseWebUI(dockerClient, databaseRepo, cfg)
	statusSvc := statusService.NewStatusService(dockerClient, appRepo, databaseRepo, dbPool, redisClient, cfg)
	domainSvc := domainService.NewDomainService(domainRepo, appRepo, dockerClient, cfg)
	billingSvc := billingService.NewBillingService(billingRepo, userRepo, cfg.StripeSecretKey, cfg.BillingPortalReturnURL)

	// Register Prometheus metrics
	middleware.RegisterMetrics()

	// Fiber app
	app := fiber.New(fiber.Config{
		AppName:      "Deployer API",
		ErrorHandler: defaultErrorHandler,
	})

	// Middleware
	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Origin,Content-Type,Accept,Authorization",
	}))

	// Routes
	handler.SetupRoutes(app, authSvc, oauthSvc, gdprSvc, appSvc, deploySvc, buildSvc, gitSvc, dbSvc, dbWebUI, statusSvc, domainSvc, billingSvc, appRepo, cfg, asynqClient, redisClient)

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := app.Listen(":" + cfg.ServerPort); err != nil {
			log.Fatalf("failed to start server: %v", err)
		}
	}()

	log.Printf("server started on port %s", cfg.ServerPort)

	<-quit
	log.Println("shutting down server...")

	if err := app.Shutdown(); err != nil {
		log.Fatalf("server shutdown error: %v", err)
	}

	log.Println("server stopped")
}

func defaultErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}
	return c.Status(code).JSON(fiber.Map{
		"error": err.Error(),
	})
}
