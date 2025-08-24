package cmd

import (
	"bufio"
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/sinhaparth5/keynginx/internal/docker"
)

var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "View KeyNginx container logs",
	Long: `View logs from the KeyNginx container.

This command displays:
• Nginx access logs
• Nginx error logs  
• Container startup logs
• Real-time log streaming (with --follow)`,
	RunE: runLogs,
}

var (
	logsProject string
	logsFollow  bool
	logsTail    string
)

func init() {
	rootCmd.AddCommand(logsCmd)

	logsCmd.Flags().StringVarP(&logsProject, "project", "p", ".", "Project directory path")
	logsCmd.Flags().BoolVarP(&logsFollow, "follow", "f", false, "Follow log output")
	logsCmd.Flags().StringVar(&logsTail, "tail", "100", "Number of lines to show from end of logs")
}

func runLogs(cmd *cobra.Command, args []string) error {
	cfg, err := loadProjectConfig(logsProject)
	if err != nil {
		return fmt.Errorf("failed to load project configuration: %w", err)
	}

	manager, err := docker.NewManager()
	if err != nil {
		return fmt.Errorf("failed to initialize Docker manager: %w", err)
	}
	defer manager.Close()

	if err := manager.CheckDockerAvailability(); err != nil {
		return err
	}

	containerName := fmt.Sprintf("keynginx-%s", cfg.Project.Domain)
	container, err := manager.GetContainerByName(containerName)
	if err != nil {
		return fmt.Errorf("container not found: %w\n\nRun 'keynginx up' to start the container", err)
	}

	fmt.Printf("📋 Viewing logs for %s\n", containerName)
	if logsFollow {
		fmt.Println("🔄 Following logs (press Ctrl+C to stop)...")
	}
	fmt.Println("=" + fmt.Sprintf("%50s", "="))

	logs, err := manager.GetContainerLogs(container.ID, logsFollow, logsTail)
	if err != nil {
		return fmt.Errorf("failed to get container logs: %w", err)
	}
	defer logs.Close()

	return streamLogs(logs)
}

func streamLogs(logs io.ReadCloser) error {
	scanner := bufio.NewScanner(logs)

	for scanner.Scan() {
		line := scanner.Text()

		if len(line) > 8 {
			cleanLine := line[8:]
			fmt.Println(cleanLine)
		} else {
			fmt.Println(line)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading logs: %w", err)
	}

	return nil
}
