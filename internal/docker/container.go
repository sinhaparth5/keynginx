package docker

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/go-connections/nat"
)

type ContainerConfig struct {
	Name        string
	Domain      string
	HTTPSPort   string
	HTTPPort    string
	ProjectDir  string
	NginxImage  string
	NetworkName string
}

type ContainerStatus struct {
	ID      string
	Name    string
	State   string
	Status  string
	Ports   []types.Port
	Created time.Time
	Image   string
}

func (c *Client) CreateContainer(config ContainerConfig) (string, error) {
	existing, err := c.GetContainerByName(config.Name)
	if err == nil && existing != nil {
		return "", fmt.Errorf("container %s already exists", config.Name)
	}

	portBindings := nat.PortMap{
		"80/tcp":  []nat.PortBinding{{HostIP: "0.0.0.0", HostPort: config.HTTPPort}},
		"443/tcp": []nat.PortBinding{{HostIP: "0.0.0.0", HostPort: config.HTTPSPort}},
	}

	mounts := []mount.Mount{
		{
			Type:     mount.TypeBind,
			Source:   fmt.Sprintf("%s/nginx.conf", config.ProjectDir),
			Target:   "/etc/nginx/nginx.conf",
			ReadOnly: true,
		},
		{
			Type:     mount.TypeBind,
			Source:   fmt.Sprintf("%s/ssl", config.ProjectDir),
			Target:   "/etc/nginx/ssl",
			ReadOnly: true,
		},
		{
			Type:   mount.TypeBind,
			Source: fmt.Sprintf("%s/logs", config.ProjectDir),
			Target: "/var/log/nginx",
		},
	}

	containerConfig := &container.Config{
		Image: config.NginxImage,
		ExposedPorts: nat.PortSet{
			"80/tcp":  {},
			"443/tcp": {},
		},
		Labels: map[string]string{
			"created-by":       "keynginx",
			"keynginx.domain":  config.Domain,
			"keynginx.version": "1.0.0",
		},
	}

	hostConfig := &container.HostConfig{
		PortBindings: portBindings,
		Mounts:       mounts,
		RestartPolicy: container.RestartPolicy{
			Name: "unless-stopped",
		},
	}

	resp, err := c.cli.ContainerCreate(
		c.ctx,
		containerConfig,
		hostConfig,
		nil,
		nil,
		config.Name,
	)
	if err != nil {
		return "", fmt.Errorf("failed to create container: %w", err)
	}

	return resp.ID, nil
}

func (c *Client) StartContainer(containerID string) error {
	err := c.cli.ContainerStart(c.ctx, containerID, container.StartOptions{})
	if err != nil {
		return fmt.Errorf("failed to start container: %w", err)
	}
	return nil
}

func (c *Client) StopContainer(containerID string, timeout *int) error {
	var timeoutPtr *int
	if timeout != nil {
		timeoutPtr = timeout
	}

	err := c.cli.ContainerStop(c.ctx, containerID, container.StopOptions{
		Timeout: timeoutPtr,
	})
	if err != nil {
		return fmt.Errorf("failed to stop container: %w", err)
	}
	return nil
}

func (c *Client) RemoveContainer(containerID string, force bool) error {
	err := c.cli.ContainerRemove(c.ctx, containerID, container.RemoveOptions{
		Force: force,
	})
	if err != nil {
		return fmt.Errorf("failed to remove container: %w", err)
	}
	return nil
}

func (c *Client) GetContainerByName(name string) (*types.Container, error) {
	containers, err := c.cli.ContainerList(c.ctx, container.ListOptions{
		All: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	for _, container := range containers {
		for _, containerName := range container.Names {
			if strings.TrimPrefix(containerName, "/") == name {
				return &container, nil
			}
		}
	}

	return nil, fmt.Errorf("container %s not found", name)
}

func (c *Client) GetContainerStatus(containerID string) (*ContainerStatus, error) {
	containerInfo, err := c.cli.ContainerInspect(c.ctx, containerID)
	if err != nil {
		return nil, fmt.Errorf("failed to inspect container: %w", err)
	}

	containers, err := c.cli.ContainerList(c.ctx, container.ListOptions{All: true})
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	var ports []types.Port
	for _, c := range containers {
		if c.ID == containerInfo.ID {
			ports = c.Ports
			break
		}
	}

	createdTime, err := time.Parse(time.RFC3339Nano, containerInfo.Created)
	if err != nil {
		createdTime = time.Now()
	}

	healthStatus := "unknown"
	if containerInfo.State.Health != nil {
		healthStatus = containerInfo.State.Health.Status
	}

	status := &ContainerStatus{
		ID:      containerInfo.ID,
		Name:    strings.TrimPrefix(containerInfo.Name, "/"),
		State:   containerInfo.State.Status,
		Status:  fmt.Sprintf("%s (%s)", containerInfo.State.Status, healthStatus),
		Ports:   ports,
		Created: createdTime, // ✅ Fixed
		Image:   containerInfo.Config.Image,
	}

	return status, nil
}

func (c *Client) GetContainerLogs(containerID string, follow bool, tail string) (io.ReadCloser, error) {
	options := container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     follow,
		Tail:       tail,
		Timestamps: true,
	}

	logs, err := c.cli.ContainerLogs(c.ctx, containerID, options)
	if err != nil {
		return nil, fmt.Errorf("failed to get container logs: %w", err)
	}

	return logs, nil
}

func (c *Client) RestartContainer(containerID string, timeout *int) error {
	var timeoutPtr *int
	if timeout != nil {
		timeoutPtr = timeout
	}

	err := c.cli.ContainerRestart(c.ctx, containerID, container.StopOptions{
		Timeout: timeoutPtr,
	})
	if err != nil {
		return fmt.Errorf("failed to restart container: %w", err)
	}
	return nil
}

func (c *Client) IsContainerRunning(containerID string) (bool, error) {
	containerInfo, err := c.cli.ContainerInspect(c.ctx, containerID)
	if err != nil {
		return false, err
	}
	return containerInfo.State.Running, nil
}
