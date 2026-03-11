package database

import (
	"context"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"github.com/deployer/api/internal/config"
	"github.com/deployer/api/internal/model"
	"github.com/deployer/api/internal/repository"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/google/uuid"
)

// DatabaseWebUI manages web-based database administration containers.
type DatabaseWebUI struct {
	docker client.APIClient
	dbRepo repository.DatabaseRepository
	cfg    *config.Config
}

// NewDatabaseWebUI creates a new DatabaseWebUI.
func NewDatabaseWebUI(
	dockerClient client.APIClient,
	dbRepo repository.DatabaseRepository,
	cfg *config.Config,
) *DatabaseWebUI {
	return &DatabaseWebUI{
		docker: dockerClient,
		dbRepo: dbRepo,
		cfg:    cfg,
	}
}

// adminerContainerName returns the container name for a web UI instance.
func adminerContainerName(dbID uuid.UUID) string {
	return fmt.Sprintf("adminer-%s", dbID.String()[:8])
}

// StartAdminer starts a web UI container for the specified database.
func (w *DatabaseWebUI) StartAdminer(ctx context.Context, userID, dbID uuid.UUID) (string, error) {
	// Verify ownership.
	mdb, err := w.dbRepo.GetByID(ctx, dbID)
	if err != nil {
		return "", fmt.Errorf("database not found: %w", err)
	}
	if mdb.UserID != userID {
		return "", fmt.Errorf("unauthorized")
	}

	if mdb.Status != model.DatabaseStatusRunning {
		return "", fmt.Errorf("database is not running")
	}

	cName := adminerContainerName(dbID)
	shortID := dbID.String()[:8]

	// Determine image and environment variables based on engine.
	imgRef, envSlice := w.getWebUIConfig(mdb)

	// Pull image.
	pullOut, err := w.docker.ImagePull(ctx, imgRef, image.PullOptions{})
	if err != nil {
		return "", fmt.Errorf("image pull failed: %w", err)
	}
	_, _ = io.Copy(io.Discard, pullOut)
	pullOut.Close()

	// Stop existing container if any.
	_ = w.stopContainer(ctx, cName)

	// Build subdomain and Traefik labels.
	adminerHost := fmt.Sprintf("adminer-%s.%s", shortID, w.cfg.AppDomain)
	proxyURL := fmt.Sprintf("https://%s", adminerHost)
	routerName := fmt.Sprintf("adminer-%s", shortID)

	// Determine the port for the web UI.
	webUIPort := w.getWebUIPort(mdb.Engine)

	containerCfg := &container.Config{
		Image: imgRef,
		Env:   envSlice,
		Labels: map[string]string{
			"traefik.enable": "true",
			fmt.Sprintf("traefik.http.routers.%s.rule", routerName):                     fmt.Sprintf("Host(`%s`)", adminerHost),
			fmt.Sprintf("traefik.http.routers.%s.entrypoints", routerName):               "web",
			fmt.Sprintf("traefik.http.services.%s.loadbalancer.server.port", routerName): webUIPort,
			"deployer.adminer":   "true",
			"deployer.db-id":     dbID.String(),
		},
	}

	hostCfg := &container.HostConfig{
		Resources: container.Resources{
			Memory: 134217728, // 128MB
		},
		RestartPolicy: container.RestartPolicy{
			Name: "no",
		},
	}

	networkCfg := &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			w.cfg.DockerNetwork: {},
		},
	}

	resp, err := w.docker.ContainerCreate(ctx, containerCfg, hostCfg, networkCfg, nil, cName)
	if err != nil {
		return "", fmt.Errorf("container create failed: %w", err)
	}

	if err := w.docker.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return "", fmt.Errorf("container start failed: %w", err)
	}

	// Schedule auto-stop after 30 minutes.
	go func() {
		time.Sleep(30 * time.Minute)
		stopCtx := context.Background()
		if stopErr := w.stopContainer(stopCtx, cName); stopErr != nil {
			log.Printf("warning: failed to auto-stop adminer container %s: %v", cName, stopErr)
		} else {
			log.Printf("adminer container %s auto-stopped after 30 minutes", cName)
		}
	}()

	log.Printf("adminer started for database %s at %s", dbID, proxyURL)
	return proxyURL, nil
}

// StopAdminer stops and removes the web UI container for the specified database.
func (w *DatabaseWebUI) StopAdminer(ctx context.Context, userID, dbID uuid.UUID) error {
	// Verify ownership.
	mdb, err := w.dbRepo.GetByID(ctx, dbID)
	if err != nil {
		return fmt.Errorf("database not found: %w", err)
	}
	if mdb.UserID != userID {
		return fmt.Errorf("unauthorized")
	}

	cName := adminerContainerName(dbID)
	return w.stopContainer(ctx, cName)
}

// getWebUIConfig returns the Docker image and environment variables for the web UI.
func (w *DatabaseWebUI) getWebUIConfig(mdb *model.ManagedDatabase) (string, []string) {
	switch mdb.Engine {
	case model.EnginePostgres, model.EngineMySQL:
		return "adminer:latest", []string{
			fmt.Sprintf("ADMINER_DEFAULT_SERVER=%s", mdb.Host),
		}
	case model.EngineMongoDB:
		connURL := fmt.Sprintf("mongodb://%s:%s@%s:%d/%s?authSource=admin",
			mdb.Username, mdb.Password, mdb.Host, mdb.Port, mdb.DBName)
		return "mongo-express:latest", []string{
			fmt.Sprintf("ME_CONFIG_MONGODB_URL=%s", connURL),
			"ME_CONFIG_BASICAUTH=false",
		}
	case model.EngineRedis:
		return "rediscommander/redis-commander:latest", []string{
			fmt.Sprintf("REDIS_HOSTS=local:%s:%d:0:%s", mdb.Host, mdb.Port, mdb.Password),
		}
	default:
		return "adminer:latest", nil
	}
}

// getWebUIPort returns the exposed port for the web UI container.
func (w *DatabaseWebUI) getWebUIPort(engine model.DatabaseEngine) string {
	switch engine {
	case model.EngineMongoDB:
		return "8081"
	case model.EngineRedis:
		return "8081"
	default:
		return "8080"
	}
}

// stopContainer stops and removes a container by name.
func (w *DatabaseWebUI) stopContainer(ctx context.Context, name string) error {
	timeout := 10
	stopOpts := container.StopOptions{Timeout: &timeout}
	err := w.docker.ContainerStop(ctx, name, stopOpts)
	if err != nil && !strings.Contains(err.Error(), "No such container") {
		return err
	}

	removeOpts := container.RemoveOptions{Force: true}
	err = w.docker.ContainerRemove(ctx, name, removeOpts)
	if err != nil && !strings.Contains(err.Error(), "No such container") {
		return err
	}

	return nil
}
