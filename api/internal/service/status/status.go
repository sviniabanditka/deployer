package status

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/deployer/api/internal/config"
	"github.com/deployer/api/internal/model"
	"github.com/deployer/api/internal/repository"
	"github.com/docker/docker/client"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// StatusService provides system and application health information.
type StatusService struct {
	docker      client.APIClient
	appRepo     repository.AppRepository
	dbRepo      repository.DatabaseRepository
	db          *pgxpool.Pool
	redisClient *redis.Client
	cfg         *config.Config
}

// NewStatusService creates a new StatusService.
func NewStatusService(
	dockerClient client.APIClient,
	appRepo repository.AppRepository,
	dbRepo repository.DatabaseRepository,
	db *pgxpool.Pool,
	redisClient *redis.Client,
	cfg *config.Config,
) *StatusService {
	return &StatusService{
		docker:      dockerClient,
		appRepo:     appRepo,
		dbRepo:      dbRepo,
		db:          db,
		redisClient: redisClient,
		cfg:         cfg,
	}
}

// GetSystemStatus returns the health status of all system components.
func (s *StatusService) GetSystemStatus(ctx context.Context) (*model.SystemStatus, error) {
	var components []model.ComponentStatus

	// API health (always operational if we can respond).
	components = append(components, model.ComponentStatus{
		Name:   "api",
		Status: "operational",
	})

	// PostgreSQL connectivity.
	components = append(components, s.checkPostgres(ctx))

	// Redis connectivity.
	components = append(components, s.checkRedis(ctx))

	// Docker daemon connectivity.
	components = append(components, s.checkDocker(ctx))

	// Traefik health.
	components = append(components, s.checkTraefik(ctx))

	// Determine overall status.
	overall := "operational"
	for _, c := range components {
		if c.Status == "down" {
			overall = "down"
			break
		}
		if c.Status == "degraded" {
			overall = "degraded"
		}
	}

	return &model.SystemStatus{
		Overall:    overall,
		Components: components,
		UpdatedAt:  time.Now().UTC(),
	}, nil
}

// checkPostgres pings the PostgreSQL database and returns its status.
func (s *StatusService) checkPostgres(ctx context.Context) model.ComponentStatus {
	start := time.Now()
	err := s.db.Ping(ctx)
	latency := time.Since(start)

	if err != nil {
		return model.ComponentStatus{
			Name:    "postgresql",
			Status:  "down",
			Latency: latency,
			Message: fmt.Sprintf("ping failed: %v", err),
		}
	}

	return model.ComponentStatus{
		Name:    "postgresql",
		Status:  "operational",
		Latency: latency,
	}
}

// checkRedis pings the Redis server and returns its status.
func (s *StatusService) checkRedis(ctx context.Context) model.ComponentStatus {
	start := time.Now()
	err := s.redisClient.Ping(ctx).Err()
	latency := time.Since(start)

	if err != nil {
		return model.ComponentStatus{
			Name:    "redis",
			Status:  "down",
			Latency: latency,
			Message: fmt.Sprintf("ping failed: %v", err),
		}
	}

	return model.ComponentStatus{
		Name:    "redis",
		Status:  "operational",
		Latency: latency,
	}
}

// checkDocker pings the Docker daemon and returns its status.
func (s *StatusService) checkDocker(ctx context.Context) model.ComponentStatus {
	start := time.Now()
	_, err := s.docker.Ping(ctx)
	latency := time.Since(start)

	if err != nil {
		return model.ComponentStatus{
			Name:    "docker",
			Status:  "down",
			Latency: latency,
			Message: fmt.Sprintf("ping failed: %v", err),
		}
	}

	return model.ComponentStatus{
		Name:    "docker",
		Status:  "operational",
		Latency: latency,
	}
}

// checkTraefik checks Traefik health via its ping endpoint.
func (s *StatusService) checkTraefik(ctx context.Context) model.ComponentStatus {
	start := time.Now()

	httpClient := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://traefik:8080/ping", nil)
	if err != nil {
		return model.ComponentStatus{
			Name:    "traefik",
			Status:  "down",
			Latency: time.Since(start),
			Message: fmt.Sprintf("request creation failed: %v", err),
		}
	}

	resp, err := httpClient.Do(req)
	latency := time.Since(start)

	if err != nil {
		return model.ComponentStatus{
			Name:    "traefik",
			Status:  "degraded",
			Latency: latency,
			Message: fmt.Sprintf("health check failed: %v", err),
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return model.ComponentStatus{
			Name:    "traefik",
			Status:  "degraded",
			Latency: latency,
			Message: fmt.Sprintf("unexpected status code: %d", resp.StatusCode),
		}
	}

	return model.ComponentStatus{
		Name:    "traefik",
		Status:  "operational",
		Latency: latency,
	}
}

// GetAppStatus returns the runtime status of a deployed application.
func (s *StatusService) GetAppStatus(ctx context.Context, appID uuid.UUID) (*model.AppRuntimeStatus, error) {
	app, err := s.appRepo.GetByID(ctx, appID)
	if err != nil {
		return nil, fmt.Errorf("app not found: %w", err)
	}

	cName := "app-" + app.Slug

	inspect, err := s.docker.ContainerInspect(ctx, cName)
	if err != nil {
		return &model.AppRuntimeStatus{
			Running:      false,
			Healthy:      false,
			Uptime:       0,
			RestartCount: 0,
		}, nil
	}

	running := inspect.State != nil && inspect.State.Running
	healthy := false
	if inspect.State != nil && inspect.State.Health != nil {
		healthy = inspect.State.Health.Status == "healthy"
	}

	var uptime time.Duration
	if running && inspect.State != nil {
		startedAt, err := time.Parse(time.RFC3339Nano, inspect.State.StartedAt)
		if err == nil {
			uptime = time.Since(startedAt)
		}
	}

	restartCount := 0
	if inspect.RestartCount > 0 {
		restartCount = inspect.RestartCount
	}

	return &model.AppRuntimeStatus{
		Running:      running,
		Healthy:      healthy,
		Uptime:       uptime,
		RestartCount: restartCount,
	}, nil
}

// RecordIncident creates a new incident record.
func (s *StatusService) RecordIncident(ctx context.Context, component, description, severity string) error {
	id := uuid.New()
	_, err := s.db.Exec(ctx,
		`INSERT INTO incidents (id, component, description, severity, status, started_at)
		 VALUES ($1, $2, $3, $4, 'investigating', NOW())`,
		id, component, description, severity,
	)
	if err != nil {
		return fmt.Errorf("failed to record incident: %w", err)
	}

	log.Printf("incident recorded: [%s] %s - %s", severity, component, description)
	return nil
}

// GetIncidentHistory returns recent incidents.
func (s *StatusService) GetIncidentHistory(ctx context.Context, limit int) ([]model.Incident, error) {
	if limit <= 0 {
		limit = 50
	}

	rows, err := s.db.Query(ctx,
		`SELECT id, component, description, severity, status, started_at, resolved_at
		 FROM incidents ORDER BY started_at DESC LIMIT $1`, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query incidents: %w", err)
	}
	defer rows.Close()

	var incidents []model.Incident
	for rows.Next() {
		var inc model.Incident
		if err := rows.Scan(&inc.ID, &inc.Component, &inc.Description, &inc.Severity, &inc.Status, &inc.StartedAt, &inc.ResolvedAt); err != nil {
			return nil, fmt.Errorf("failed to scan incident: %w", err)
		}
		incidents = append(incidents, inc)
	}
	return incidents, rows.Err()
}

// ResolveIncident marks an incident as resolved.
func (s *StatusService) ResolveIncident(ctx context.Context, incidentID uuid.UUID) error {
	_, err := s.db.Exec(ctx,
		`UPDATE incidents SET status = 'resolved', resolved_at = NOW() WHERE id = $1`,
		incidentID,
	)
	return err
}

