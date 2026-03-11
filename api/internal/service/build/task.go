package build

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

const TaskBuildImage = "build:image"

type BuildImagePayload struct {
	AppID        string `json:"app_id"`
	DeploymentID string `json:"deployment_id"`
	ArchivePath  string `json:"archive_path"`
}

func NewBuildImageTask(appID, deploymentID uuid.UUID, archivePath string) (*asynq.Task, error) {
	payload, err := json.Marshal(BuildImagePayload{
		AppID:        appID.String(),
		DeploymentID: deploymentID.String(),
		ArchivePath:  archivePath,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal build image payload: %w", err)
	}
	return asynq.NewTask(TaskBuildImage, payload), nil
}

func ParseBuildImagePayload(task *asynq.Task) (*BuildImagePayload, error) {
	var p BuildImagePayload
	if err := json.Unmarshal(task.Payload(), &p); err != nil {
		return nil, fmt.Errorf("failed to unmarshal build image payload: %w", err)
	}
	return &p, nil
}
