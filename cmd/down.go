package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/sinhaparth5/keynginx/internal/docker"
)

var downCmd = &cobra.Command{
	Use:   "down",
	Short: "Stop and remove KeyNginx containers",
	Long: `Stop and remove KeyNginx containers for the current project.

This command will:
• Stop the running Nginx container
• Remove the container (keeping volumes and data)
• Clean up resources

Use 'keynginx up' to start the server again.`,
	RunE: runDown,
}

var (
	downProject string
	downRemove  bool
	downForce   bool
)

func init() {
	rootCmd.AddCommand(downCmd)

	downCmd.Flags().StringVarP(&downProject, "project", "p", ".", "Project directory path")
	downCmd.Flags().BoolVar(&downRemove, "remove", true, "Remove containers after stopping")
	downCmd.Flags().BoolVar(&downForce, "force", false, "Force stop containers")
}

func runDown(cmd *cobra.Command, args []string) error {
	fmt.Println("🛑 Stopping KeyNginx Server")
	fmt.Println("============================")

	cfg, err := loadProjectConfig(downProject)
	if err != nil {
		return fmt.Errorf("failed to load project configuration: %w", err)
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

	if status.Status == "not-found" {
		fmt.Printf("⚠️  No container found for domain '%s'\n", cfg.Project.Domain)
		return nil
	}

	fmt.Printf("📋 Found container: %s (%s)\n", status.ContainerName, status.Status)

	if downRemove {
		fmt.Print("🛑 Stopping and removing container... ")
		if err := manager.StopAndRemoveContainer(cfg); err != nil {
			fmt.Println("❌")
			return fmt.Errorf("failed to stop and remove container: %w", err)
		}
		fmt.Println("✅")
		fmt.Printf("✅ Container %s stopped and removed\n", status.ContainerName)
	} else {
		fmt.Print("⏸️  Stopping container... ")
		containerName := fmt.Sprintf("keynginx-%s", cfg.Project.Domain)
		container, err := manager.GetContainerByName(containerName)
		if err != nil {
			fmt.Println("❌")
			return fmt.Errorf("container not found: %w", err)
		}

		timeout := 30
		if downForce {
			timeout = 1
		}

		if err := manager.StopContainer(container.ID, &timeout); err != nil {
			fmt.Println("❌")
			return fmt.Errorf("failed to stop container: %w", err)
		}
		fmt.Println("✅")
		fmt.Printf("✅ Container %s stopped\n", status.ContainerName)
	}

	fmt.Println("\n💡 Server stopped successfully!")
	fmt.Printf("   • Start again: keynginx up\n")
	fmt.Printf("   • View project: cd %s\n", cfg.Project.OutputDir)

	return nil
}
