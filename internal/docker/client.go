package docker

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/system"
	"github.com/docker/docker/client"
)

type Client struct {
	cli *client.Client
	ctx context.Context
}

func NewClient() (*Client, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %w", err)
	}

	return &Client{
		cli: cli,
		ctx: context.Background(),
	}, nil
}

func (c *Client) Close() error {
	return c.cli.Close()
}

func (c *Client) IsDockerAvailable() error {
	_, err := c.cli.Ping(c.ctx)
	if err != nil {
		return fmt.Errorf("Docker daemon is not running: %w", err)
	}
	return nil
}

func (c *Client) GetDockerInfo() (*system.Info, error) {
	info, err := c.cli.Info(c.ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get Docker info: %w", err)
	}
	return &info, nil
}

func (c *Client) GetDockerVersion() (types.Version, error) {
	version, err := c.cli.ServerVersion(c.ctx)
	if err != nil {
		return types.Version{}, fmt.Errorf("failed to get Docker version: %w", err)
	}
	return version, nil
}

func (c *Client) ListKeyNginxContainers() ([]types.Container, error) {
	containers, err := c.cli.ContainerList(c.ctx, container.ListOptions{ // ✅ Fixed
		All: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	var keynginxContainers []types.Container
	for _, container := range containers {
		if isKeyNginxContainer(container) {
			keynginxContainers = append(keynginxContainers, container)
		}
	}

	return keynginxContainers, nil
}

func isKeyNginxContainer(container types.Container) bool {
	for _, name := range container.Names {
		if len(name) > 0 && name[0] == '/' {
			name = name[1:] // Remove leading slash
		}
		if len(name) >= 8 && name[:8] == "keynginx" {
			return true
		}
	}

	if labels := container.Labels; labels != nil {
		if value, exists := labels["created-by"]; exists && value == "keynginx" {
			return true
		}
	}

	return false
}
