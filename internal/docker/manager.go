package docker

import (
	"fmt"
	"io"
	"path/filepath"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/sinhaparth5/keynginx/internal/config"
	"github.com/sinhaparth5/keynginx/internal/utils"
)

type Manager struct {
	client *Client
}

func NewManager() (*Manager, error) {
	client, err := NewClient()
	if err != nil {
		return nil, err
	}

	return &Manager{
		client: client,
	}, nil
}

func (m *Manager) Close() error {
	return m.client.Close()
}

func (m *Manager) CheckDockerAvailability() error {
	if err := m.client.IsDockerAvailable(); err != nil {
		return fmt.Errorf(`Docker is not available: %w

Please ensure Docker is installed and running:
• macOS/Windows: Start Docker Desktop
• Linux: sudo systemctl start docker

Install Docker: https://docs.docker.com/get-docker/`, err)
	}
	return nil
}

func (m *Manager) CreateAndStartContainer(cfg *config.Config) (string, error) {
	if err := m.validateProjectFiles(cfg); err != nil {
		return "", err
	}

	projectDir, err := utils.GetAbsolutePath(cfg.Project.OutputDir)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}

	logsDir := filepath.Join(projectDir, "logs")
	if err := utils.EnsureDirectory(logsDir); err != nil {
		return "", fmt.Errorf("failed to create logs directory: %w", err)
	}

	containerName := fmt.Sprintf("keynginx-%s", cfg.Project.Domain)

	containerConfig := ContainerConfig{
		Name:        containerName,
		Domain:      cfg.Project.Domain,
		HTTPSPort:   fmt.Sprintf("%d", cfg.Nginx.HTTPSPort),
		HTTPPort:    fmt.Sprintf("%d", cfg.Nginx.HTTPPort),
		ProjectDir:  projectDir,
		NginxImage:  cfg.Docker.NginxImage,
		NetworkName: cfg.Docker.NetworkName,
	}

	containerID, err := m.client.CreateContainer(containerConfig)
	if err != nil {
		return "", err
	}

	if err := m.client.StartContainer(containerID); err != nil {
		return "", fmt.Errorf("container created but failed to start: %w", err)
	}

	return containerID, nil
}

func (m *Manager) StopAndRemoveContainer(cfg *config.Config) error {
	containerName := fmt.Sprintf("keynginx-%s", cfg.Project.Domain)

	container, err := m.client.GetContainerByName(containerName)
	if err != nil {
		return fmt.Errorf("container %s not found", containerName)
	}

	timeout := 30
	if err := m.client.StopContainer(container.ID, &timeout); err != nil {
		return fmt.Errorf("failed to stop container: %w", err)
	}

	if err := m.client.RemoveContainer(container.ID, false); err != nil {
		return fmt.Errorf("failed to remove container: %w", err)
	}

	return nil
}

func (m *Manager) GetProjectStatus(cfg *config.Config) (*ProjectStatus, error) {
	containerName := fmt.Sprintf("keynginx-%s", cfg.Project.Domain)

	container, err := m.client.GetContainerByName(containerName)
	if err != nil {
		return &ProjectStatus{
			ProjectName:   cfg.Project.Name,
			Domain:        cfg.Project.Domain,
			ContainerName: containerName,
			Status:        "not-found",
			Message:       "Container not found",
		}, nil
	}

	status, err := m.client.GetContainerStatus(container.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get container status: %w", err)
	}

	projectStatus := &ProjectStatus{
		ProjectName:   cfg.Project.Name,
		Domain:        cfg.Project.Domain,
		ContainerName: containerName,
		ContainerID:   status.ID,
		Status:        status.State,
		Message:       status.Status,
		Ports:         status.Ports,
		Created:       status.Created,
		Image:         status.Image,
	}

	return projectStatus, nil
}

func (m *Manager) RestartContainer(cfg *config.Config) error {
	containerName := fmt.Sprintf("keynginx-%s", cfg.Project.Domain)

	container, err := m.client.GetContainerByName(containerName)
	if err != nil {
		return fmt.Errorf("container %s not found", containerName)
	}

	timeout := 30
	return m.client.RestartContainer(container.ID, &timeout)
}

func (m *Manager) GetContainerByName(name string) (*types.Container, error) {
	return m.client.GetContainerByName(name)
}

func (m *Manager) StopContainer(containerID string, timeout *int) error {
	return m.client.StopContainer(containerID, timeout)
}

func (m *Manager) RemoveContainer(containerID string, force bool) error {
	return m.client.RemoveContainer(containerID, force)
}

func (m *Manager) GetContainerLogs(containerID string, follow bool, tail string) (io.ReadCloser, error) {
	return m.client.GetContainerLogs(containerID, follow, tail)
}

func (m *Manager) ListKeyNginxContainers() ([]types.Container, error) {
	return m.client.ListKeyNginxContainers()
}

func (m *Manager) StartContainer(containerID string) error {
	return m.client.StartContainer(containerID)
}

func (m *Manager) GetContainerStatus(containerID string) (*ContainerStatus, error) {
	return m.client.GetContainerStatus(containerID)
}

func (m *Manager) validateProjectFiles(cfg *config.Config) error {
	projectDir := cfg.Project.OutputDir

	requiredFiles := map[string]string{
		"nginx.conf":          "Nginx configuration file",
		"ssl/private.key":     "SSL private key",
		"ssl/certificate.crt": "SSL certificate",
	}

	for file, description := range requiredFiles {
		filePath := filepath.Join(projectDir, file)
		if !utils.FileExists(filePath) {
			return fmt.Errorf("%s not found: %s\n\nRun 'keynginx init' to generate required files", description, filePath)
		}
	}

	return nil
}

type ProjectStatus struct {
	ProjectName   string       `json:"project_name"`
	Domain        string       `json:"domain"`
	ContainerName string       `json:"container_name"`
	ContainerID   string       `json:"container_id"`
	Status        string       `json:"status"`
	Message       string       `json:"message"`
	Ports         []types.Port `json:"ports"`
	Created       time.Time    `json:"created"`
	Image         string       `json:"image"`
}

func (ps *ProjectStatus) IsRunning() bool {
	return ps.Status == "running"
}

func (ps *ProjectStatus) GetHTTPSURL() string {
	for _, port := range ps.Ports {
		if port.PrivatePort == 443 && port.PublicPort != 0 {
			return fmt.Sprintf("https://localhost:%d", port.PublicPort)
		}
	}
	return fmt.Sprintf("https://%s", ps.Domain)
}

func (ps *ProjectStatus) GetHTTPURL() string {
	for _, port := range ps.Ports {
		if port.PrivatePort == 80 && port.PublicPort != 0 {
			return fmt.Sprintf("http://localhost:%d", port.PublicPort)
		}
	}
	return fmt.Sprintf("http://%s", ps.Domain)
}
