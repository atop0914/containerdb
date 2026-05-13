// Package commands provides CLI commands for containerdb.
package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

// stopCmd represents the stop command
var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop a running container",
	Long:  `Stop a running database container. Note: containers started via this CLI are ephemeral and should be stopped using this command.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Containers started via CLI are managed by docker directly
		// This is a placeholder for future container tracking functionality
		fmt.Println("Note: Containers started via containerdb CLI are ephemeral.")
		fmt.Println("To stop a container, use: docker ps | grep <container-name>")
		fmt.Println("Then: docker stop <container-id>")
		fmt.Println("")
		fmt.Println("For persistent container management, consider using docker-compose or the library directly.")
	},
}

func init() {
	stopCmd.Flags().StringP("name", "n", "", "Container name to stop")
}
