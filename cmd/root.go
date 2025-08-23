package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "keynginx",
	Short: "KeyNginx - SSL Certificate Generator",
	Long: `KeyNginx is a CLI tool for generatoring SSL certificates.
	This is Phase 1 - focused on core certificates generation functionality`,
	Version: getVersion(),
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")
}

func getVersion() string {
	return fmt.Sprintf("%s (commit: %s, built: %s)", Version, GitCommit, BuildTime)
}
