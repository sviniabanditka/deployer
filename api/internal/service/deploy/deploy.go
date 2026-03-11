package deploy

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/deployer/api/internal/config"
	"github.com/deployer/api/internal/model"
	"github.com/deployer/api/internal/repository"
	dockertypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DeployService manages container lifecycle for deployed applications.
type DeployService struct {
	docker      client.APIClient
	appRepo     repository.AppRepository
	domainRepo  repository.DomainRepository
	db          *pgxpool.Pool
	cfg         *config.Config
	networkName string
	netMgr      *NetworkManager
}

// NewDeployService creates a new DeployService.
func NewDeployService(
	dockerClient client.APIClient,
	appRepo repository.AppRepository,
	domainRepo repository.DomainRepository,
	db *pgxpool.Pool,
	cfg *config.Config,
) *DeployService {
	return &DeployService{
		docker:      dockerClient,
		appRepo:     appRepo,
		domainRepo:  domainRepo,
		db:          db,
		cfg:         cfg,
		networkName: cfg.DockerNetwork,
		netMgr:      NewNetworkManager(dockerClient),
	}
}

// containerName returns the canonical container name for an app.
func containerName(slug string) string {
	return "app-" + slug
}

// blueGreenContainerName returns a colored container name for blue-green deploys.
func blueGreenContainerName(slug, color string) string {
	return fmt.Sprintf("app-%s-%s", slug, color)
}

// BlueGreenDeploy performs a zero-downtime deployment using blue-green strategy.
func (s *DeployService) BlueGreenDeploy(ctx context.Context, app *model.App, deployment *model.Deployment) error {
	imageRef := fmt.Sprintf("%s/%s:%d", s.cfg.RegistryURL, app.Slug, deployment.Version)

	// 1. Pull image from registry.
	pullOut, err := s.docker.ImagePull(ctx, imageRef, image.PullOptions{})
	if err != nil {
		return fmt.Errorf("image pull failed: %w", err)
	}
	_, _ = io.Copy(io.Discard, pullOut)
	pullOut.Close()

	// 2. Determine current color.
	currentColor := s.detectCurrentColor(ctx, app.Slug)
	newColor := "blue"
	if currentColor == "blue" {
		newColor = "green"
	}

	newContainerName := blueGreenContainerName(app.Slug, newColor)
	oldContainerName := blueGreenContainerName(app.Slug, currentColor)

	// Also clean up the non-colored legacy container if it exists.
	legacyName := containerName(app.Slug)
	_ = s.stopAndRemoveContainer(ctx, legacyName)

	// 3. Stop and remove the new color container if it already exists (stale).
	_ = s.stopAndRemoveContainer(ctx, newContainerName)

	// 4. Fetch environment variables from DB.
	envVars, err := s.getEnvVars(ctx, app.ID)
	if err != nil {
		return fmt.Errorf("failed to fetch env vars: %w", err)
	}

	var envSlice []string
	for k, v := range envVars {
		envSlice = append(envSlice, k+"="+v)
	}

	// 5. Create the new container.
	slug := app.Slug

	// Build Traefik Host rule including verified custom domains.
	hostRules := []string{fmt.Sprintf("Host(`%s.%s`)", slug, s.cfg.AppDomain)}
	if s.domainRepo != nil {
		verifiedDomains, domErr := s.domainRepo.ListVerifiedByAppID(ctx, app.ID)
		if domErr == nil {
			for _, d := range verifiedDomains {
				hostRules = append(hostRules, fmt.Sprintf("Host(`%s`)", d.Domain))
			}
		}
	}
	hostRule := strings.Join(hostRules, " || ")

	labels := map[string]string{
		"traefik.enable": "true",
		fmt.Sprintf("traefik.http.routers.%s.rule", slug):                     hostRule,
		fmt.Sprintf("traefik.http.routers.%s.entrypoints", slug):              "web",
		fmt.Sprintf("traefik.http.services.%s.loadbalancer.server.port", slug): "8080",
		"deployer.color": newColor,
	}

	// If custom domains exist, add HTTPS router with Let's Encrypt.
	if len(hostRules) > 1 {
		labels[fmt.Sprintf("traefik.http.routers.%s-secure.rule", slug)] = hostRule
		labels[fmt.Sprintf("traefik.http.routers.%s-secure.entrypoints", slug)] = "websecure"
		labels[fmt.Sprintf("traefik.http.routers.%s-secure.tls.certresolver", slug)] = "letsencrypt"
		labels[fmt.Sprintf("traefik.http.services.%s-secure.loadbalancer.server.port", slug)] = "8080"
	}

	containerCfg := &container.Config{
		Image:  imageRef,
		Env:    envSlice,
		Labels: labels,
		Healthcheck: &container.HealthConfig{
			Test:     []string{"CMD-SHELL", "curl -f http://localhost:8080/ || exit 1"},
			Interval: 5 * time.Second,
			Timeout:  3 * time.Second,
			Retries:  3,
		},
	}

	hostCfg := &container.HostConfig{
		Resources: container.Resources{
			Memory:   s.cfg.DefaultMemoryLimit,
			NanoCPUs: int64(s.cfg.DefaultCPULimit * 1e9),
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

	resp, err := s.docker.ContainerCreate(ctx, containerCfg, hostCfg, networkCfg, nil, newContainerName)
	if err != nil {
		return fmt.Errorf("container create failed: %w", err)
	}

	// 6. Connect to user-specific network for isolation.
	userNetworkID, err := s.netMgr.GetOrCreateUserNetwork(ctx, app.UserID)
	if err == nil && userNetworkID != "" {
		_ = s.netMgr.ConnectContainerToNetwork(ctx, resp.ID, userNetworkID)

		// Also connect linked database containers to the user network.
		s.connectLinkedDatabases(ctx, app.ID, userNetworkID)
	}

	// 7. Start the new container.
	if err := s.docker.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		_ = s.stopAndRemoveContainer(ctx, newContainerName)
		return fmt.Errorf("container start failed: %w", err)
	}

	// 8. Wait for health check.
	if err := s.healthCheck(resp.ID, 60*time.Second); err != nil {
		// Rollback: stop and remove the unhealthy new container.
		_ = s.stopAndRemoveContainer(ctx, newContainerName)
		return fmt.Errorf("health check failed, rolled back: %w", err)
	}

	// 9. New container is healthy. Stop and remove the old container.
	if currentColor != "" {
		_ = s.stopAndRemoveContainer(ctx, oldContainerName)
	}

	// 10. Update deployment status.
	_, err = s.db.Exec(ctx,
		`UPDATE deployments SET status = $1 WHERE id = $2`,
		"running", deployment.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update deployment status: %w", err)
	}

	// 11. Update app status.
	app.Status = model.AppStatusRunning
	if err := s.appRepo.Update(ctx, app); err != nil {
		return fmt.Errorf("failed to update app status: %w", err)
	}

	return nil
}

// Deploy pulls the image, recreates the container, and starts it.
// This is the simpler fallback deploy method without blue-green logic.
func (s *DeployService) Deploy(ctx context.Context, app *model.App, deployment *model.Deployment) error {
	// Use blue-green deployment as the default strategy.
	return s.BlueGreenDeploy(ctx, app, deployment)
}

// Rollback deploys a previous deployment version.
func (s *DeployService) Rollback(ctx context.Context, app *model.App, targetDeploymentID uuid.UUID) error {
	deployment, err := s.appRepo.GetDeployment(ctx, targetDeploymentID)
	if err != nil {
		return fmt.Errorf("failed to get target deployment: %w", err)
	}

	if deployment.AppID != app.ID {
		return fmt.Errorf("deployment does not belong to this app")
	}

	return s.BlueGreenDeploy(ctx, app, deployment)
}

// healthCheck polls the container health status via the Docker API until it reports
// healthy or the timeout expires.
func (s *DeployService) healthCheck(containerID string, timeout time.Duration) error {
	ctx := context.Background()
	deadline := time.Now().Add(timeout)
	interval := 2 * time.Second

	for time.Now().Before(deadline) {
		inspect, err := s.docker.ContainerInspect(ctx, containerID)
		if err != nil {
			return fmt.Errorf("failed to inspect container: %w", err)
		}

		// If the container has no healthcheck configured, try an HTTP probe.
		if inspect.State.Health == nil {
			if s.httpHealthProbe(inspect.NetworkSettings, 8080) {
				return nil
			}
			time.Sleep(interval)
			continue
		}

		switch inspect.State.Health.Status {
		case "healthy":
			return nil
		case "unhealthy":
			return fmt.Errorf("container reported unhealthy")
		}

		// Still starting, wait and retry.
		time.Sleep(interval)
	}

	return fmt.Errorf("health check timed out after %v", timeout)
}

// httpHealthProbe attempts an HTTP GET to the container's IP on the given port.
func (s *DeployService) httpHealthProbe(netSettings *dockertypes.NetworkSettings, port int) bool {
	if netSettings == nil {
		return false
	}
	for _, ep := range netSettings.Networks {
		if ep.IPAddress == "" {
			continue
		}
		url := fmt.Sprintf("http://%s:%d/", ep.IPAddress, port)
		client := &http.Client{Timeout: 3 * time.Second}
		resp, err := client.Get(url)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode < 500 {
				return true
			}
		}
	}
	return false
}

// detectCurrentColor inspects which blue/green container is currently running.
func (s *DeployService) detectCurrentColor(ctx context.Context, slug string) string {
	for _, color := range []string{"blue", "green"} {
		name := blueGreenContainerName(slug, color)
		inspect, err := s.docker.ContainerInspect(ctx, name)
		if err == nil && inspect.State != nil && inspect.State.Running {
			return color
		}
	}
	return ""
}

// connectLinkedDatabases connects database containers linked to an app to the user network.
func (s *DeployService) connectLinkedDatabases(ctx context.Context, appID uuid.UUID, userNetworkID string) {
	rows, err := s.db.Query(ctx,
		`SELECT container_id FROM managed_databases WHERE app_id = $1 AND status = 'running' AND container_id != ''`,
		appID,
	)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var containerID string
		if err := rows.Scan(&containerID); err != nil {
			continue
		}
		_ = s.netMgr.ConnectContainerToNetwork(ctx, containerID, userNetworkID)
	}
}

// Stop stops a running container for the given app.
func (s *DeployService) Stop(ctx context.Context, app *model.App) error {
	// Try to stop blue-green containers first.
	color := s.detectCurrentColor(ctx, app.Slug)
	var cName string
	if color != "" {
		cName = blueGreenContainerName(app.Slug, color)
	} else {
		cName = containerName(app.Slug)
	}

	timeout := 30 // seconds
	stopOpts := container.StopOptions{Timeout: &timeout}
	if err := s.docker.ContainerStop(ctx, cName, stopOpts); err != nil {
		return fmt.Errorf("container stop failed: %w", err)
	}

	app.Status = model.AppStatusStopped
	if err := s.appRepo.Update(ctx, app); err != nil {
		return fmt.Errorf("failed to update app status: %w", err)
	}

	return nil
}

// Remove stops and removes the container, resetting the app status.
func (s *DeployService) Remove(ctx context.Context, app *model.App) error {
	// Remove both blue and green containers.
	for _, color := range []string{"blue", "green"} {
		_ = s.stopAndRemoveContainer(ctx, blueGreenContainerName(app.Slug, color))
	}
	// Also remove legacy container name.
	_ = s.stopAndRemoveContainer(ctx, containerName(app.Slug))

	app.Status = model.AppStatusCreated
	if err := s.appRepo.Update(ctx, app); err != nil {
		return fmt.Errorf("failed to update app status: %w", err)
	}

	return nil
}

// GetLogs returns a log stream from the container.
func (s *DeployService) GetLogs(ctx context.Context, app *model.App, follow bool) (io.ReadCloser, error) {
	// Try blue-green containers first.
	color := s.detectCurrentColor(ctx, app.Slug)
	var cName string
	if color != "" {
		cName = blueGreenContainerName(app.Slug, color)
	} else {
		cName = containerName(app.Slug)
	}

	opts := container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     follow,
		Tail:       "200",
	}

	return s.docker.ContainerLogs(ctx, cName, opts)
}

// GetContainerStats returns current CPU/memory/network stats for the container.
func (s *DeployService) GetContainerStats(ctx context.Context, app *model.App) (*ContainerStats, error) {
	// Try blue-green containers first.
	color := s.detectCurrentColor(ctx, app.Slug)
	var cName string
	if color != "" {
		cName = blueGreenContainerName(app.Slug, color)
	} else {
		cName = containerName(app.Slug)
	}

	resp, err := s.docker.ContainerStats(ctx, cName, false)
	if err != nil {
		return nil, fmt.Errorf("container stats failed: %w", err)
	}
	defer resp.Body.Close()

	return parseStats(resp.Body)
}

// stopAndRemoveContainer attempts to stop and remove a container by name. Errors are
// silently ignored when the container does not exist.
func (s *DeployService) stopAndRemoveContainer(ctx context.Context, name string) error {
	timeout := 10
	stopOpts := container.StopOptions{Timeout: &timeout}
	err := s.docker.ContainerStop(ctx, name, stopOpts)
	if err != nil && !isNotFoundError(err) {
		return err
	}

	removeOpts := container.RemoveOptions{Force: true}
	err = s.docker.ContainerRemove(ctx, name, removeOpts)
	if err != nil && !isNotFoundError(err) {
		return err
	}

	return nil
}

// getEnvVars fetches all environment variables for the given app from the database.
func (s *DeployService) getEnvVars(ctx context.Context, appID interface{}) (map[string]string, error) {
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

// isNotFoundError checks if a Docker API error is a "not found" error.
func isNotFoundError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "No such container")
}
