package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/sinhaparth5/keynginx/internal/docker"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show KeyNginx project status",
	Long: `Show detailed status information for KeyNginx projects.

This command displays:
• Container status and health
• Port mappings and URLs
• Resource usage
• Configuration summary`,
	RunE: runStatus,
}

var (
	statusProject string
	statusJSON    bool
	statusAll     bool
)

func init() {
	rootCmd.AddCommand(statusCmd)

	statusCmd.Flags().StringVarP(&statusProject, "project", "p", ".", "Project directory path")
	statusCmd.Flags().BoolVar(&statusJSON, "json", false, "Output status in JSON format")
	statusCmd.Flags().BoolVar(&statusAll, "all", false, "Show all KeyNginx containers")
}

func runStatus(cmd *cobra.Command, args []string) error {
	manager, err := docker.NewManager()
	if err != nil {
		return fmt.Errorf("failed to initialize Docker manager: %w", err)
	}
	defer manager.Close()

	if err := manager.CheckDockerAvailability(); err != nil {
		if statusJSON {
			errorOutput := map[string]string{"error": "Docker not available", "details": err.Error()}
			jsonData, _ := json.MarshalIndent(errorOutput, "", "  ")
			fmt.Println(string(jsonData))
		} else {
			fmt.Printf("❌ Docker not available: %v\n", err)
		}
		return err
	}

	if statusAll {
		return showAllContainers(manager)
	}

	cfg, err := loadProjectConfig(statusProject)
	if err != nil {
		return fmt.Errorf("failed to load project configuration: %w", err)
	}

	status, err := manager.GetProjectStatus(cfg)
	if err != nil {
		return fmt.Errorf("failed to get project status: %w", err)
	}

	if statusJSON {
		return outputStatusJSON(status)
	}

	return displayStatusTable(status)
}

func showAllContainers(manager *docker.Manager) error {
	containers, err := manager.ListKeyNginxContainers()
	if err != nil {
		return fmt.Errorf("failed to list containers: %w", err)
	}

	if len(containers) == 0 {
		fmt.Println("📭 No KeyNginx containers found")
		return nil
	}

	fmt.Printf("📦 Found %d KeyNginx container(s)\n", len(containers))
	fmt.Println("=" + fmt.Sprintf("%40s", "="))

	for _, container := range containers {
		status, err := manager.GetContainerStatus(container.ID)
		if err != nil {
			fmt.Printf("❌ Error getting status for %s: %v\n", container.Names[0], err)
			continue
		}

		displayContainerSummary(status)
		fmt.Println()
	}

	return nil
}

func displayStatusTable(status *docker.ProjectStatus) error {
	fmt.Println("📊 KeyNginx Project Status")
	fmt.Println("==========================")

	statusIcon := "❌"
	statusColor := "stopped"
	if status.IsRunning() {
		statusIcon = "✅"
		statusColor = "running"
	} else if status.Status == "not-found" {
		statusIcon = "⚠️"
		statusColor = "not found"
	}

	fmt.Printf("🏷️  Project: %s\n", status.ProjectName)
	fmt.Printf("🌐 Domain: %s\n", status.Domain)
	fmt.Printf("📦 Container: %s\n", status.ContainerName)
	fmt.Printf("🔄 Status: %s %s\n", statusIcon, statusColor)

	if status.Status != "not-found" {
		fmt.Printf("🆔 ID: %s\n", status.ContainerID[:12])
		fmt.Printf("🏷️  Image: %s\n", status.Image)
		fmt.Printf("📅 Created: %s\n", status.Created.Format("2006-01-02 15:04:05"))

		if len(status.Ports) > 0 {
			fmt.Println("🔗 URLs:")
			if status.IsRunning() {
				fmt.Printf("   • HTTPS: %s\n", status.GetHTTPSURL())
				fmt.Printf("   • HTTP:  %s (redirects to HTTPS)\n", status.GetHTTPURL())
			}

			fmt.Println("🚪 Port Mappings:")
			for _, port := range status.Ports {
				if port.PublicPort != 0 {
					fmt.Printf("   • %d:%d (%s)\n", port.PublicPort, port.PrivatePort, port.Type)
				}
			}
		}
	}

	// Actions
	fmt.Println("\n💡 Available Actions:")
	if status.IsRunning() {
		fmt.Printf("   • keynginx down    (stop server)\n")
		fmt.Printf("   • keynginx logs    (view logs)\n")
		fmt.Printf("   • keynginx logs -f (follow logs)\n")
	} else if status.Status == "exited" {
		fmt.Printf("   • keynginx up      (start server)\n")
		fmt.Printf("   • keynginx down    (remove container)\n")
	} else if status.Status == "not-found" {
		fmt.Printf("   • keynginx up      (create and start)\n")
		fmt.Printf("   • keynginx init    (regenerate project)\n")
	}

	return nil
}

func displayContainerSummary(status *docker.ContainerStatus) {
	statusIcon := "❌"
	if status.State == "running" {
		statusIcon = "✅"
	}

	fmt.Printf("%s %s (%s)\n", statusIcon, status.Name, status.State)
	fmt.Printf("   ID: %s | Image: %s\n", status.ID[:12], status.Image)
	fmt.Printf("   Created: %s\n", status.Created.Format("2006-01-02 15:04:05"))
}

func outputStatusJSON(status *docker.ProjectStatus) error {
	jsonData, err := json.MarshalIndent(status, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal status to JSON: %w", err)
	}

	fmt.Println(string(jsonData))
	return nil
}
