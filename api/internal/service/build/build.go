package build

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/deployer/api/internal/config"
	"github.com/deployer/api/internal/repository"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/registry"
	dockerclient "github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

// BuildService handles container image builds for deployments.
type BuildService struct {
	appRepo      repository.AppRepository
	cfg          *config.Config
	dockerClient *dockerclient.Client
	asynqClient  *asynq.Client
}

// NewBuildService creates a new BuildService with the given dependencies.
func NewBuildService(
	appRepo repository.AppRepository,
	cfg *config.Config,
	dockerClient *dockerclient.Client,
	asynqClient *asynq.Client,
) *BuildService {
	return &BuildService{
		appRepo:      appRepo,
		cfg:          cfg,
		dockerClient: dockerClient,
		asynqClient:  asynqClient,
	}
}

// EnqueueBuild creates and enqueues an Asynq task to build a container image.
func (s *BuildService) EnqueueBuild(appID, deploymentID uuid.UUID, archivePath string) error {
	task, err := NewBuildImageTask(appID, deploymentID, archivePath)
	if err != nil {
		return fmt.Errorf("failed to create build task: %w", err)
	}
	_, err = s.asynqClient.Enqueue(task)
	if err != nil {
		return fmt.Errorf("failed to enqueue build task: %w", err)
	}
	return nil
}

// ProcessBuild is the Asynq task handler that performs the actual image build.
func (s *BuildService) ProcessBuild(ctx context.Context, task *asynq.Task) error {
	payload, err := ParseBuildImagePayload(task)
	if err != nil {
		return fmt.Errorf("failed to parse payload: %w", err)
	}

	appID, err := uuid.Parse(payload.AppID)
	if err != nil {
		return fmt.Errorf("invalid app ID: %w", err)
	}
	deploymentID, err := uuid.Parse(payload.DeploymentID)
	if err != nil {
		return fmt.Errorf("invalid deployment ID: %w", err)
	}

	var buildLog strings.Builder

	// Update status to building
	if err := s.appRepo.UpdateDeploymentStatus(ctx, deploymentID, "building", ""); err != nil {
		return fmt.Errorf("failed to update deployment status: %w", err)
	}

	// Extract ZIP archive to temp directory
	tmpDir, err := os.MkdirTemp("", "build-*")
	if err != nil {
		s.failDeployment(ctx, deploymentID, &buildLog, "failed to create temp dir: "+err.Error())
		return fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	buildLog.WriteString("Extracting archive...\n")
	if err := extractZip(payload.ArchivePath, tmpDir); err != nil {
		s.failDeployment(ctx, deploymentID, &buildLog, "failed to extract archive: "+err.Error())
		return fmt.Errorf("failed to extract archive: %w", err)
	}
	buildLog.WriteString("Archive extracted successfully.\n")

	// Detect runtime
	runtime, err := DetectRuntime(tmpDir)
	if err != nil {
		s.failDeployment(ctx, deploymentID, &buildLog, "failed to detect runtime: "+err.Error())
		return fmt.Errorf("failed to detect runtime: %w", err)
	}
	buildLog.WriteString(fmt.Sprintf("Detected runtime: %s\n", runtime))

	// Get app for slug and version info
	app, err := s.appRepo.GetByID(ctx, appID)
	if err != nil {
		s.failDeployment(ctx, deploymentID, &buildLog, "failed to get app: "+err.Error())
		return fmt.Errorf("failed to get app: %w", err)
	}

	deployment, err := s.appRepo.GetDeployment(ctx, deploymentID)
	if err != nil {
		s.failDeployment(ctx, deploymentID, &buildLog, "failed to get deployment: "+err.Error())
		return fmt.Errorf("failed to get deployment: %w", err)
	}

	imageTag := fmt.Sprintf("%s/%s:%d", s.cfg.RegistryURL, app.Slug, deployment.Version)
	buildLog.WriteString(fmt.Sprintf("Building image: %s\n", imageTag))

	// Build the image
	hasDockerfile := runtime == "docker"
	if hasDockerfile {
		if err := s.buildWithDocker(ctx, tmpDir, imageTag, &buildLog); err != nil {
			s.failDeployment(ctx, deploymentID, &buildLog, "docker build failed: "+err.Error())
			return fmt.Errorf("docker build failed: %w", err)
		}
	} else {
		if err := s.buildWithNixpacks(ctx, tmpDir, imageTag, &buildLog); err != nil {
			s.failDeployment(ctx, deploymentID, &buildLog, "nixpacks build failed: "+err.Error())
			return fmt.Errorf("nixpacks build failed: %w", err)
		}
	}

	buildLog.WriteString("Build completed successfully.\n")

	// Push image to registry
	buildLog.WriteString(fmt.Sprintf("Pushing image to %s...\n", s.cfg.RegistryURL))
	if err := s.pushImage(ctx, imageTag, &buildLog); err != nil {
		s.failDeployment(ctx, deploymentID, &buildLog, "failed to push image: "+err.Error())
		return fmt.Errorf("failed to push image: %w", err)
	}
	buildLog.WriteString("Image pushed successfully.\n")

	// Update deployment status to built
	if err := s.appRepo.UpdateDeploymentStatus(ctx, deploymentID, "built", buildLog.String()); err != nil {
		return fmt.Errorf("failed to update deployment status: %w", err)
	}

	// Update app runtime
	app.Runtime = runtime
	if err := s.appRepo.Update(ctx, app); err != nil {
		log.Printf("warning: failed to update app runtime: %v", err)
	}

	// Clean up the uploaded archive
	os.Remove(payload.ArchivePath)

	log.Printf("build completed for app %s deployment %s", appID, deploymentID)
	return nil
}

func (s *BuildService) failDeployment(ctx context.Context, deploymentID uuid.UUID, buildLog *strings.Builder, msg string) {
	buildLog.WriteString("ERROR: " + msg + "\n")
	if err := s.appRepo.UpdateDeploymentStatus(ctx, deploymentID, "failed", buildLog.String()); err != nil {
		log.Printf("failed to update deployment status to failed: %v", err)
	}
}

func (s *BuildService) buildWithDocker(ctx context.Context, projectDir, imageTag string, buildLog *strings.Builder) error {
	buildLog.WriteString("Building with Docker...\n")

	tar, err := archive.TarWithOptions(projectDir, &archive.TarOptions{})
	if err != nil {
		return fmt.Errorf("failed to create tar archive: %w", err)
	}
	defer tar.Close()

	resp, err := s.dockerClient.ImageBuild(ctx, tar, types.ImageBuildOptions{
		Tags:       []string{imageTag},
		Dockerfile: "Dockerfile",
		Remove:     true,
	})
	if err != nil {
		return fmt.Errorf("failed to start docker build: %w", err)
	}
	defer resp.Body.Close()

	// Read build output
	output, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read build output: %w", err)
	}
	buildLog.WriteString(string(output))

	return nil
}

func (s *BuildService) buildWithNixpacks(ctx context.Context, projectDir, imageTag string, buildLog *strings.Builder) error {
	buildLog.WriteString("Building with Nixpacks...\n")

	cmd := exec.CommandContext(ctx, s.cfg.NixpacksPath, "build", projectDir, "--name", imageTag)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		buildLog.WriteString(stderr.String())
		return fmt.Errorf("nixpacks build failed: %w: %s", err, stderr.String())
	}

	buildLog.WriteString(stdout.String())
	return nil
}

func (s *BuildService) pushImage(ctx context.Context, imageTag string, buildLog *strings.Builder) error {
	authConfig := registry.AuthConfig{
		ServerAddress: s.cfg.RegistryURL,
	}
	authJSON, err := json.Marshal(authConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal auth config: %w", err)
	}
	authStr := base64.URLEncoding.EncodeToString(authJSON)

	resp, err := s.dockerClient.ImagePush(ctx, imageTag, image.PushOptions{
		RegistryAuth: authStr,
	})
	if err != nil {
		return fmt.Errorf("failed to push image: %w", err)
	}
	defer resp.Close()

	output, err := io.ReadAll(resp)
	if err != nil {
		return fmt.Errorf("failed to read push output: %w", err)
	}
	buildLog.WriteString(string(output))

	return nil
}

func extractZip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return fmt.Errorf("failed to open zip: %w", err)
	}
	defer r.Close()

	for _, f := range r.File {
		target := filepath.Join(dest, f.Name)

		// Prevent zip slip
		if !strings.HasPrefix(filepath.Clean(target), filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path in zip: %s", f.Name)
		}

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(target, 0o755); err != nil {
				return err
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}

		outFile, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		rc.Close()
		outFile.Close()
		if err != nil {
			return err
		}
	}

	return nil
}
