package deploy

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/google/uuid"
)

// NetworkManager manages per-user Docker networks for container isolation.
type NetworkManager struct {
	docker client.APIClient
}

// NewNetworkManager creates a new NetworkManager.
func NewNetworkManager(dockerClient client.APIClient) *NetworkManager {
	return &NetworkManager{docker: dockerClient}
}

// shortUserID returns the first 12 characters of a UUID string for use in network names.
func shortUserID(userID uuid.UUID) string {
	s := userID.String()
	if len(s) > 12 {
		return s[:12]
	}
	return s
}

// networkName returns the canonical Docker network name for a user.
func userNetworkName(userID uuid.UUID) string {
	return "deployer-user-" + shortUserID(userID)
}

// CreateUserNetwork creates an isolated Docker bridge network for a user.
func (m *NetworkManager) CreateUserNetwork(ctx context.Context, userID uuid.UUID) (string, error) {
	name := userNetworkName(userID)

	resp, err := m.docker.NetworkCreate(ctx, name, network.CreateOptions{
		Driver:     "bridge",
		Internal:   false, // needs internet access
		EnableIPv6: nil,
		Options: map[string]string{
			"com.docker.network.bridge.enable_icc": "true",
		},
		Labels: map[string]string{
			"managed-by": "deployer",
			"user-id":    userID.String(),
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to create user network: %w", err)
	}

	return resp.ID, nil
}

// DeleteUserNetwork removes the Docker network for a user.
func (m *NetworkManager) DeleteUserNetwork(ctx context.Context, userID uuid.UUID) error {
	name := userNetworkName(userID)
	if err := m.docker.NetworkRemove(ctx, name); err != nil {
		return fmt.Errorf("failed to remove user network: %w", err)
	}
	return nil
}

// GetOrCreateUserNetwork returns the network ID for the user, creating it if it does not exist.
func (m *NetworkManager) GetOrCreateUserNetwork(ctx context.Context, userID uuid.UUID) (string, error) {
	name := userNetworkName(userID)

	// Try to inspect existing network first.
	inspect, err := m.docker.NetworkInspect(ctx, name, network.InspectOptions{})
	if err == nil {
		return inspect.ID, nil
	}

	// Network does not exist, create it.
	return m.CreateUserNetwork(ctx, userID)
}

// ConnectContainerToNetwork connects a container to a Docker network.
func (m *NetworkManager) ConnectContainerToNetwork(ctx context.Context, containerID, networkID string) error {
	if err := m.docker.NetworkConnect(ctx, networkID, containerID, nil); err != nil {
		return fmt.Errorf("failed to connect container to network: %w", err)
	}
	return nil
}

// DisconnectContainerFromNetwork disconnects a container from a Docker network.
func (m *NetworkManager) DisconnectContainerFromNetwork(ctx context.Context, containerID, networkID string) error {
	if err := m.docker.NetworkDisconnect(ctx, networkID, containerID, false); err != nil {
		return fmt.Errorf("failed to disconnect container from network: %w", err)
	}
	return nil
}
