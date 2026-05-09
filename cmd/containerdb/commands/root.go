// Package commands provides CLI commands for containerdb.
package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

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
		cmd.Help()
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
}

func printVersion() {
	fmt.Println("containerdb version 0.1.0")
}
