package deploy

import (
	"context"
	"fmt"

	"github.com/deployer/api/internal/model"
)

// PostBuildDeploy is a hook that can be called after a successful build completes.
// It triggers a deployment of the built image to a running container.
func (s *DeployService) PostBuildDeploy(ctx context.Context, app *model.App, deployment *model.Deployment) error {
	if err := s.Deploy(ctx, app, deployment); err != nil {
		return fmt.Errorf("post-build deploy failed: %w", err)
	}
	return nil
}
