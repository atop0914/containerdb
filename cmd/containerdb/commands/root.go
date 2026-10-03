// Package commands provides CLI commands for containerdb.
package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/atop0914/containerdb-bootcamp/internal/version"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(version.Info())
	},
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "containerdb",
	Short: "ContainerDB - Database containers for development and testing",
	Long: `ContainerDB is a lightweight toolkit for spinning up real databases in containers
with a single function call. Perfect for local development and testing.

Supported databases:
  - MySQL
  - PostgreSQL
  - SQLite (in-memory or temp file)

Examples:
  containerdb start                    # Start a MySQL container with defaults
  containerdb start -t postgres        # Start a PostgreSQL container
  containerdb start -t mysql -i mysql:8.0  # Start specific MySQL version
  containerdb status                    # Show running containers
  containerdb stop                      # Stop containers`,
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

// Execute adds all child commands to the root command
func Execute() error {
	return rootCmd.Execute()
}

// AddCommands adds all subcommands to the root command
func AddCommands() {
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(stopCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(composeCmd)
	rootCmd.AddCommand(versionCmd)
}
