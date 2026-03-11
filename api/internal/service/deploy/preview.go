package deploy

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
	"github.com/jackc/pgx/v5/pgxpool"
)

// PreviewService manages preview deployments for pull requests.
type PreviewService struct {
	docker      client.APIClient
	appRepo     repository.AppRepository
	db          *pgxpool.Pool
	cfg         *config.Config
	networkName string
}

// NewPreviewService creates a new PreviewService.
func NewPreviewService(
	dockerClient client.APIClient,
	appRepo repository.AppRepository,
	db *pgxpool.Pool,
	cfg *config.Config,
) *PreviewService {
	return &PreviewService{
		docker:      dockerClient,
		appRepo:     appRepo,
		db:          db,
		cfg:         cfg,
		networkName: cfg.DockerNetwork,
	}
}

// previewContainerName returns the container name for a preview deployment.
func previewContainerName(slug string, prNumber int) string {
	return fmt.Sprintf("preview-%s-pr-%d", slug, prNumber)
}

// CreatePreview creates a preview deployment for a pull request.
func (s *PreviewService) CreatePreview(ctx context.Context, app *model.App, prNumber int, prURL string, imageTag string) (*model.Deployment, error) {
	// Pull image from registry.
	pullOut, err := s.docker.ImagePull(ctx, imageTag, image.PullOptions{})
	if err != nil {
		return nil, fmt.Errorf("image pull failed: %w", err)
	}
	_, _ = io.Copy(io.Discard, pullOut)
	pullOut.Close()

	cName := previewContainerName(app.Slug, prNumber)

	// Stop and remove existing preview container (if any).
	_ = s.stopAndRemoveContainer(ctx, cName)

	// Fetch environment variables from DB.
	envVars, err := s.getEnvVars(ctx, app.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch env vars: %w", err)
	}

	var envSlice []string
	for k, v := range envVars {
		envSlice = append(envSlice, k+"="+v)
	}

	// Build preview subdomain: pr-{number}-{slug}.{appDomain}
	previewHost := fmt.Sprintf("pr-%d-%s.%s", prNumber, app.Slug, s.cfg.AppDomain)
	previewURL := fmt.Sprintf("https://%s", previewHost)

	routerName := fmt.Sprintf("preview-%s-pr-%d", app.Slug, prNumber)

	containerCfg := &container.Config{
		Image: imageTag,
		Env:   envSlice,
		Labels: map[string]string{
			"traefik.enable": "true",
			fmt.Sprintf("traefik.http.routers.%s.rule", routerName):                     fmt.Sprintf("Host(`%s`)", previewHost),
			fmt.Sprintf("traefik.http.routers.%s.entrypoints", routerName):               "web",
			fmt.Sprintf("traefik.http.services.%s.loadbalancer.server.port", routerName): "8080",
			"deployer.preview":    "true",
			"deployer.app-slug":   app.Slug,
			"deployer.pr-number":  fmt.Sprintf("%d", prNumber),
		},
	}

	// Preview deployments get lower resource limits: 256MB, 0.25 CPU.
	hostCfg := &container.HostConfig{
		Resources: container.Resources{
			Memory:   268435456, // 256MB
			NanoCPUs: int64(0.25 * 1e9),
		},
		RestartPolicy: container.RestartPolicy{
			Name: container.RestartPolicyUnlessStopped,
		},
	}

	networkCfg := &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			s.networkName: {},
		},
	}

	resp, err := s.docker.ContainerCreate(ctx, containerCfg, hostCfg, networkCfg, nil, cName)
	if err != nil {
		return nil, fmt.Errorf("container create failed: %w", err)
	}

	if err := s.docker.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return nil, fmt.Errorf("container start failed: %w", err)
	}

	// Determine version.
	version := 1
	latestDeployment, err := s.appRepo.GetLatestDeployment(ctx, app.ID)
	if err == nil && latestDeployment != nil {
		version = latestDeployment.Version + 1
	}

	// Create deployment record with preview fields.
	deployment := &model.Deployment{
		AppID:             app.ID,
		Version:           version,
		Status:            "running",
		ImageTag:          imageTag,
		IsPreview:         true,
		PreviewURL:        previewURL,
		PullRequestNumber: prNumber,
		PullRequestURL:    prURL,
	}
	if err := s.appRepo.CreateDeployment(ctx, deployment); err != nil {
		return nil, fmt.Errorf("failed to create deployment record: %w", err)
	}

	log.Printf("preview deployment created for app %s PR #%d at %s", app.Slug, prNumber, previewURL)
	return deployment, nil
}

// DestroyPreview stops and removes a preview deployment.
func (s *PreviewService) DestroyPreview(ctx context.Context, app *model.App, prNumber int) error {
	cName := previewContainerName(app.Slug, prNumber)

	if err := s.stopAndRemoveContainer(ctx, cName); err != nil {
		return fmt.Errorf("failed to remove preview container: %w", err)
	}

	// Update deployment status to destroyed.
	deployment, err := s.appRepo.GetPreviewDeployment(ctx, app.ID, prNumber)
	if err != nil {
		log.Printf("warning: could not find preview deployment for app %s PR #%d: %v", app.Slug, prNumber, err)
		return nil
	}

	_, err = s.db.Exec(ctx,
		`UPDATE deployments SET status = 'destroyed' WHERE id = $1`,
		deployment.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update deployment status: %w", err)
	}

	log.Printf("preview deployment destroyed for app %s PR #%d", app.Slug, prNumber)
	return nil
}

// CleanupStalePreviews finds and destroys preview deployments older than 7 days.
func (s *PreviewService) CleanupStalePreviews(ctx context.Context) error {
	staleAge := 7 * 24 * time.Hour
	deployments, err := s.appRepo.ListStalePreviewDeployments(ctx, staleAge)
	if err != nil {
		return fmt.Errorf("failed to list stale preview deployments: %w", err)
	}

	for _, d := range deployments {
		app, err := s.appRepo.GetByID(ctx, d.AppID)
		if err != nil {
			log.Printf("warning: could not find app %s for stale preview cleanup: %v", d.AppID, err)
			continue
		}

		if err := s.DestroyPreview(ctx, app, d.PullRequestNumber); err != nil {
			log.Printf("warning: failed to cleanup stale preview for app %s PR #%d: %v", app.Slug, d.PullRequestNumber, err)
		}
	}

	return nil
}

// stopAndRemoveContainer stops and removes a container by name.
func (s *PreviewService) stopAndRemoveContainer(ctx context.Context, name string) error {
	timeout := 10
	stopOpts := container.StopOptions{Timeout: &timeout}
	err := s.docker.ContainerStop(ctx, name, stopOpts)
	if err != nil && !isPreviewNotFoundError(err) {
		return err
	}

	removeOpts := container.RemoveOptions{Force: true}
	err = s.docker.ContainerRemove(ctx, name, removeOpts)
	if err != nil && !isPreviewNotFoundError(err) {
		return err
	}

	return nil
}

// getEnvVars fetches all environment variables for the given app from the database.
func (s *PreviewService) getEnvVars(ctx context.Context, appID interface{}) (map[string]string, error) {
	rows, err := s.db.Query(ctx,
		`SELECT key, value FROM env_vars WHERE app_id = $1`, appID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	vars := make(map[string]string)
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			return nil, err
		}
		vars[k] = v
	}
	return vars, rows.Err()
}

// isPreviewNotFoundError checks if a Docker API error is a "not found" error.
func isPreviewNotFoundError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "No such container")
}
