package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/deployer/api/internal/config"
	"github.com/deployer/api/internal/repository"
	buildService "github.com/deployer/api/internal/service/build"
	databaseService "github.com/deployer/api/internal/service/database"
	dockerclient "github.com/docker/docker/client"
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

	// Parse Redis URL for Asynq
	redisOpts, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		log.Fatalf("failed to parse redis URL: %v", err)
	}

	asynqRedisOpt := asynq.RedisClientOpt{
		Addr:     redisOpts.Addr,
		Password: redisOpts.Password,
		DB:       redisOpts.DB,
	}

	// Docker client
	dockerClient, err := dockerclient.NewClientWithOpts(dockerclient.FromEnv, dockerclient.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatalf("failed to create docker client: %v", err)
	}
	defer dockerClient.Close()
	log.Println("connected to Docker")

	// Asynq client (used by build service for potential sub-task enqueueing)
	asynqClient := asynq.NewClient(asynqRedisOpt)
	defer asynqClient.Close()

	// Repositories
	appRepo := repository.NewAppRepository(dbPool)
	databaseRepo := repository.NewDatabaseRepository(dbPool)

	// Build service
	buildSvc := buildService.NewBuildService(appRepo, cfg, dockerClient, asynqClient)

	// Database service (for scheduled backups)
	dbSvc := databaseService.NewDatabaseService(dockerClient, databaseRepo, appRepo, dbPool, cfg)

	// Asynq server
	srv := asynq.NewServer(
		asynqRedisOpt,
		asynq.Config{
			Concurrency: 5,
			Queues: map[string]int{
				"default": 1,
			},
		},
	)

	// Register handlers
	mux := asynq.NewServeMux()
	mux.HandleFunc(buildService.TaskBuildImage, buildSvc.ProcessBuild)
	mux.HandleFunc(databaseService.TaskScheduledBackup, dbSvc.ProcessScheduledBackup)

	// Start backup scheduler in background
	backupScheduler := databaseService.NewBackupScheduler(databaseRepo, asynqRedisOpt)
	go func() {
		if err := backupScheduler.StartScheduler(ctx); err != nil {
			log.Printf("backup scheduler error: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := srv.Run(mux); err != nil {
			log.Fatalf("failed to start worker: %v", err)
		}
	}()

	log.Println("worker started, waiting for tasks...")

	<-quit
	log.Println("shutting down worker...")
	backupScheduler.Shutdown()
	srv.Shutdown()
	log.Println("worker stopped")
}
