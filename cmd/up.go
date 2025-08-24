package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/sinhaparth5/keynginx/internal/config"
	"github.com/sinhaparth5/keynginx/internal/docker"
	"github.com/sinhaparth5/keynginx/internal/utils"
)

var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Start KeyNginx containers",
	Long: `Start KeyNginx containers for the current project.

This command will:
• Check if Docker is running
• Validate project configuration
• Create and start the Nginx container
• Display connection information

Use this after running 'keynginx init' to start your server.`,
	RunE: runUp,
}

var (
	upDetach   bool
	upRecreate bool
	upProject  string
)

func init() {
	rootCmd.AddCommand(upCmd)

	upCmd.Flags().BoolVarP(&upDetach, "detach", "d", true, "Run containers in background")
	upCmd.Flags().BoolVar(&upRecreate, "recreate", false, "Recreate containers even if they exist")
	upCmd.Flags().StringVarP(&upProject, "project", "p", ".", "Project directory path")
}

func runUp(cmd *cobra.Command, args []string) error {
	fmt.Println("🚀 Starting KeyNginx Server")
	fmt.Println("============================")
	cfg, err := loadProjectConfig(upProject)
	if err != nil {
		return fmt.Errorf("failed to load project configuration: %w\n\nRun 'keynginx init' to create a new project", err)
	}

	manager, err := docker.NewManager()
	if err != nil {
		return fmt.Errorf("failed to initialize Docker manager: %w", err)
	}
	defer manager.Close()

	fmt.Print("🐳 Checking Docker availability... ")
	if err := manager.CheckDockerAvailability(); err != nil {
		fmt.Println("❌")
		return err
	}
	fmt.Println("✅")

	status, err := manager.GetProjectStatus(cfg)
	if err != nil {
		return fmt.Errorf("failed to get project status: %w", err)
	}

	if status.Status != "not-found" {
		if status.IsRunning() && !upRecreate {
			fmt.Printf("⚠️  Container is already running!\n\n")
			printServerInfo(status)
			return nil
		}

		if upRecreate {
			fmt.Print("🔄 Recreating container... ")
			if err := manager.StopAndRemoveContainer(cfg); err != nil {
				fmt.Println("❌")
				return fmt.Errorf("failed to remove existing container: %w", err)
			}
			fmt.Println("✅")
		} else if status.Status == "exited" {
			fmt.Print("▶️  Starting existing container... ")
			if err := manager.RestartContainer(cfg); err != nil {
				fmt.Println("❌")
				return fmt.Errorf("failed to start existing container: %w", err)
			}
			fmt.Println("✅")

			status, err = manager.GetProjectStatus(cfg)
			if err != nil {
				return fmt.Errorf("failed to get updated status: %w", err)
			}

			printServerInfo(status)
			return nil
		}
	}

	fmt.Print("📦 Creating and starting container... ")
	containerID, err := manager.CreateAndStartContainer(cfg)
	if err != nil {
		fmt.Println("❌")
		return fmt.Errorf("failed to create and start container: %w", err)
	}
	fmt.Println("✅")

	fmt.Print("⏳ Waiting for server to be ready... ")
	if err := waitForContainer(manager, cfg, 30*time.Second); err != nil {
		fmt.Println("⚠️")
		fmt.Printf("Container started but may not be fully ready: %v\n", err)
	} else {
		fmt.Println("✅")
	}

	status, err = manager.GetProjectStatus(cfg)
	if err != nil {
		return fmt.Errorf("failed to get final status: %w", err)
	}

	fmt.Printf("🎉 KeyNginx server started successfully!\n")
	fmt.Printf("📋 Container ID: %s\n\n", containerID[:12])

	printServerInfo(status)

	if upDetach {
		fmt.Println("\n💡 Server is running in the background")
		fmt.Printf("   • View logs: keynginx logs\n")
		fmt.Printf("   • Check status: keynginx status\n")
		fmt.Printf("   • Stop server: keynginx down\n")
	}

	return nil
}

func waitForContainer(manager *docker.Manager, cfg *config.Config, timeout time.Duration) error {
	start := time.Now()
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			status, err := manager.GetProjectStatus(cfg)
			if err != nil {
				return err
			}

			if status.IsRunning() {
				return nil
			}

			if time.Since(start) > timeout {
				return fmt.Errorf("timeout waiting for container to start")
			}

		case <-time.After(timeout):
			return fmt.Errorf("timeout waiting for container to start")
		}
	}
}

func printServerInfo(status *docker.ProjectStatus) {
	fmt.Println("🌐 Server Information:")
	fmt.Printf("   • Domain: %s\n", status.Domain)
	fmt.Printf("   • HTTPS: %s\n", status.GetHTTPSURL())
	fmt.Printf("   • HTTP: %s (redirects to HTTPS)\n", status.GetHTTPURL())
	fmt.Printf("   • Status: %s\n", status.Status)
	fmt.Printf("   • Container: %s\n", status.ContainerName)

	if len(status.Ports) > 0 {
		fmt.Println("   • Ports:")
		for _, port := range status.Ports {
			if port.PublicPort != 0 {
				fmt.Printf("     - %d:%d (%s)\n", port.PublicPort, port.PrivatePort, port.Type)
			}
		}
	}
}

func loadProjectConfig(projectPath string) (*config.Config, error) {
	configPaths := []string{
		fmt.Sprintf("%s/keynginx.yaml", projectPath),
		fmt.Sprintf("%s/keynginx.yml", projectPath),
		"./keynginx.yaml",
		"./keynginx.yml",
	}

	for _, path := range configPaths {
		if utils.FileExists(path) {
			return config.LoadConfig(path)
		}
	}

	return nil, fmt.Errorf("KeyNginx configuration file not found in %s", projectPath)
}
