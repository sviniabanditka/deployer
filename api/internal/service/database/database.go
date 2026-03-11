package database

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/deployer/api/internal/config"
	"github.com/deployer/api/internal/model"
	"github.com/deployer/api/internal/repository"
	deployService "github.com/deployer/api/internal/service/deploy"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DatabaseService manages the lifecycle of managed database containers.
type DatabaseService struct {
	docker   client.APIClient
	dbRepo   repository.DatabaseRepository
	appRepo  repository.AppRepository
	db       *pgxpool.Pool
	cfg      *config.Config
	netMgr   *deployService.NetworkManager
}

// NewDatabaseService creates a new DatabaseService.
func NewDatabaseService(
	dockerClient client.APIClient,
	dbRepo repository.DatabaseRepository,
	appRepo repository.AppRepository,
	db *pgxpool.Pool,
	cfg *config.Config,
) *DatabaseService {
	return &DatabaseService{
		docker:  dockerClient,
		dbRepo:  dbRepo,
		appRepo: appRepo,
		db:      db,
		cfg:     cfg,
		netMgr:  deployService.NewNetworkManager(dockerClient),
	}
}

// Create provisions a new managed database container.
func (s *DatabaseService) Create(
	ctx context.Context,
	userID uuid.UUID,
	name, engine, version string,
	appID *uuid.UUID,
) (*model.ManagedDatabase, error) {
	eng := model.DatabaseEngine(engine)
	if !model.ValidEngine(engine) {
		return nil, fmt.Errorf("unsupported database engine: %s", engine)
	}

	if version == "" {
		version = model.DefaultVersions[eng]
	}

	username := generateUsername(eng)
	password := generatePassword(24)
	dbName := generateDBName(name)
	port := model.DefaultPort(eng)
	shortID := uuid.New().String()[:8]
	containerName := fmt.Sprintf("db-%s-%s", name, shortID)
	volumeName := fmt.Sprintf("dbdata-%s-%s", name, shortID)
	host := containerName // internal Docker DNS name

	// Pull image
	imgRef := model.DockerImage(eng, version)
	pullOut, err := s.docker.ImagePull(ctx, imgRef, image.PullOptions{})
	if err != nil {
		return nil, fmt.Errorf("image pull failed: %w", err)
	}
	_, _ = io.Copy(io.Discard, pullOut)
	pullOut.Close()

	// Create volume
	_, err = s.docker.VolumeCreate(ctx, volume.CreateOptions{
		Name: volumeName,
	})
	if err != nil {
		return nil, fmt.Errorf("volume create failed: %w", err)
	}

	// Build env vars and data path based on engine
	var envSlice []string
	var dataPath string
	var cmd []string

	switch eng {
	case model.EnginePostgres:
		envSlice = []string{
			"POSTGRES_USER=" + username,
			"POSTGRES_PASSWORD=" + password,
			"POSTGRES_DB=" + dbName,
		}
		dataPath = "/var/lib/postgresql/data"
	case model.EngineMySQL:
		envSlice = []string{
			"MYSQL_ROOT_PASSWORD=" + password,
			"MYSQL_DATABASE=" + dbName,
			"MYSQL_USER=" + username,
			"MYSQL_PASSWORD=" + password,
		}
		dataPath = "/var/lib/mysql"
	case model.EngineMongoDB:
		envSlice = []string{
			"MONGO_INITDB_ROOT_USERNAME=" + username,
			"MONGO_INITDB_ROOT_PASSWORD=" + password,
			"MONGO_INITDB_DATABASE=" + dbName,
		}
		dataPath = "/data/db"
	case model.EngineRedis:
		cmd = []string{"redis-server", "--requirepass", password}
		dataPath = "/data"
	}

	// Create container
	containerCfg := &container.Config{
		Image: imgRef,
		Env:   envSlice,
		Labels: map[string]string{
			"managed-by": "deployer",
			"db-engine":  string(eng),
		},
	}
	if len(cmd) > 0 {
		containerCfg.Cmd = cmd
	}

	hostCfg := &container.HostConfig{
		Resources: container.Resources{
			Memory: s.cfg.DefaultDBMemoryLimit,
		},
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeVolume,
				Source: volumeName,
				Target: dataPath,
			},
		},
		RestartPolicy: container.RestartPolicy{
			Name: container.RestartPolicyUnlessStopped,
		},
	}

	networkCfg := &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			s.cfg.DockerNetwork: {},
		},
	}

	resp, err := s.docker.ContainerCreate(ctx, containerCfg, hostCfg, networkCfg, nil, containerName)
	if err != nil {
		return nil, fmt.Errorf("container create failed: %w", err)
	}

	// Start container
	if err := s.docker.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return nil, fmt.Errorf("container start failed: %w", err)
	}

	// Connect to user-specific network for isolation.
	userNetworkID, err := s.netMgr.GetOrCreateUserNetwork(ctx, userID)
	if err == nil && userNetworkID != "" {
		_ = s.netMgr.ConnectContainerToNetwork(ctx, resp.ID, userNetworkID)
	}

	connURL := buildConnectionURL(eng, host, port, username, password, dbName)

	mdb := &model.ManagedDatabase{
		UserID:        userID,
		AppID:         appID,
		Name:          name,
		Engine:        eng,
		Version:       version,
		Status:        model.DatabaseStatusRunning,
		Host:          host,
		Port:          port,
		DBName:        dbName,
		Username:      username,
		Password:      password,
		ConnectionURL: connURL,
		ContainerID:   resp.ID,
		MemoryLimit:   s.cfg.DefaultDBMemoryLimit,
		StorageLimit:  s.cfg.DefaultDBStorageLimit,
	}

	if err := s.dbRepo.Create(ctx, mdb); err != nil {
		return nil, fmt.Errorf("failed to save database record: %w", err)
	}

	// If appID provided, auto-set DATABASE_URL env var on the app
	if appID != nil {
		_ = s.setAppDatabaseURL(ctx, *appID, connURL)
	}

	return mdb, nil
}

// Delete stops and removes the database container and volume.
func (s *DatabaseService) Delete(ctx context.Context, userID, dbID uuid.UUID) error {
	mdb, err := s.dbRepo.GetByID(ctx, dbID)
	if err != nil {
		return err
	}
	if mdb.UserID != userID {
		return fmt.Errorf("unauthorized")
	}

	// Stop and remove container
	if mdb.ContainerID != "" {
		timeout := 10
		stopOpts := container.StopOptions{Timeout: &timeout}
		_ = s.docker.ContainerStop(ctx, mdb.ContainerID, stopOpts)
		_ = s.docker.ContainerRemove(ctx, mdb.ContainerID, container.RemoveOptions{Force: true})
	}

	// Remove volume
	volumeName := s.volumeNameFromContainer(mdb)
	if volumeName != "" {
		_ = s.docker.VolumeRemove(ctx, volumeName, true)
	}

	mdb.Status = model.DatabaseStatusDeleted
	return s.dbRepo.Update(ctx, mdb)
}

// Stop stops a running database container.
func (s *DatabaseService) Stop(ctx context.Context, userID, dbID uuid.UUID) error {
	mdb, err := s.dbRepo.GetByID(ctx, dbID)
	if err != nil {
		return err
	}
	if mdb.UserID != userID {
		return fmt.Errorf("unauthorized")
	}

	if mdb.ContainerID != "" {
		timeout := 30
		stopOpts := container.StopOptions{Timeout: &timeout}
		if err := s.docker.ContainerStop(ctx, mdb.ContainerID, stopOpts); err != nil {
			return fmt.Errorf("container stop failed: %w", err)
		}
	}

	mdb.Status = model.DatabaseStatusStopped
	return s.dbRepo.Update(ctx, mdb)
}

// Start starts a stopped database container.
func (s *DatabaseService) Start(ctx context.Context, userID, dbID uuid.UUID) error {
	mdb, err := s.dbRepo.GetByID(ctx, dbID)
	if err != nil {
		return err
	}
	if mdb.UserID != userID {
		return fmt.Errorf("unauthorized")
	}

	if mdb.ContainerID != "" {
		if err := s.docker.ContainerStart(ctx, mdb.ContainerID, container.StartOptions{}); err != nil {
			return fmt.Errorf("container start failed: %w", err)
		}
	}

	mdb.Status = model.DatabaseStatusRunning
	return s.dbRepo.Update(ctx, mdb)
}

// GetConnectionInfo returns the full database record including sensitive connection details.
func (s *DatabaseService) GetConnectionInfo(ctx context.Context, userID, dbID uuid.UUID) (*model.ManagedDatabase, error) {
	mdb, err := s.dbRepo.GetByID(ctx, dbID)
	if err != nil {
		return nil, err
	}
	if mdb.UserID != userID {
		return nil, fmt.Errorf("unauthorized")
	}
	return mdb, nil
}

// List returns all databases for a user.
func (s *DatabaseService) List(ctx context.Context, userID uuid.UUID) ([]model.ManagedDatabase, error) {
	return s.dbRepo.ListByUserID(ctx, userID)
}

// LinkToApp sets DATABASE_URL on the app and updates the managed_databases.app_id.
func (s *DatabaseService) LinkToApp(ctx context.Context, userID, dbID, appID uuid.UUID) error {
	mdb, err := s.dbRepo.GetByID(ctx, dbID)
	if err != nil {
		return err
	}
	if mdb.UserID != userID {
		return fmt.Errorf("unauthorized")
	}

	// Verify app ownership
	app, err := s.appRepo.GetByID(ctx, appID)
	if err != nil {
		return fmt.Errorf("app not found: %w", err)
	}
	if app.UserID != userID {
		return fmt.Errorf("unauthorized")
	}

	if err := s.setAppDatabaseURL(ctx, appID, mdb.ConnectionURL); err != nil {
		return fmt.Errorf("failed to set DATABASE_URL: %w", err)
	}

	mdb.AppID = &appID
	return s.dbRepo.Update(ctx, mdb)
}

// UnlinkFromApp removes the app association from the database.
func (s *DatabaseService) UnlinkFromApp(ctx context.Context, userID, dbID uuid.UUID) error {
	mdb, err := s.dbRepo.GetByID(ctx, dbID)
	if err != nil {
		return err
	}
	if mdb.UserID != userID {
		return fmt.Errorf("unauthorized")
	}

	// Remove DATABASE_URL env var from the app if linked
	if mdb.AppID != nil {
		_, _ = s.db.Exec(ctx,
			`DELETE FROM env_vars WHERE app_id = $1 AND key = 'DATABASE_URL'`,
			*mdb.AppID,
		)
	}

	mdb.AppID = nil
	return s.dbRepo.Update(ctx, mdb)
}

// setAppDatabaseURL upserts the DATABASE_URL env var for an app.
func (s *DatabaseService) setAppDatabaseURL(ctx context.Context, appID uuid.UUID, connURL string) error {
	_, err := s.db.Exec(ctx,
		`INSERT INTO env_vars (id, app_id, key, value)
		 VALUES (uuid_generate_v4(), $1, 'DATABASE_URL', $2)
		 ON CONFLICT (app_id, key) DO UPDATE SET value = $2`,
		appID, connURL,
	)
	return err
}

// volumeNameFromContainer extracts the volume name from container mounts.
func (s *DatabaseService) volumeNameFromContainer(mdb *model.ManagedDatabase) string {
	if mdb.ContainerID == "" {
		return ""
	}
	ctx := context.Background()
	inspect, err := s.docker.ContainerInspect(ctx, mdb.ContainerID)
	if err != nil {
		return ""
	}
	for _, m := range inspect.Mounts {
		if m.Type == mount.TypeVolume {
			return m.Name
		}
	}
	return ""
}

// UpdateBackupSettings updates the auto-backup configuration for a managed database.
func (s *DatabaseService) UpdateBackupSettings(ctx context.Context, userID, dbID uuid.UUID, autoBackup *bool, retention *int, schedule *string) error {
	mdb, err := s.dbRepo.GetByID(ctx, dbID)
	if err != nil {
		return err
	}
	if mdb.UserID != userID {
		return fmt.Errorf("unauthorized")
	}

	if autoBackup != nil {
		mdb.AutoBackup = *autoBackup
	}
	if retention != nil {
		if *retention < 1 {
			return fmt.Errorf("retention must be at least 1")
		}
		mdb.BackupRetention = *retention
	}
	if schedule != nil {
		mdb.BackupSchedule = *schedule
	}

	return s.dbRepo.Update(ctx, mdb)
}

// isNotFoundError checks if a Docker API error is a "not found" error.
func isNotFoundError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "No such container")
}
