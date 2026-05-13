// Package commands provides CLI commands for containerdb.
package commands

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show status of running containers",
	Long:  `Show status of running database containers managed by containerdb.`,
	Run: func(cmd *cobra.Command, args []string) {
		all, _ := cmd.Flags().GetBool("all")

		fmt.Println("=== ContainerDB Status ===")
		fmt.Println("")

		// Check docker availability
		dockerArgs := []string{"ps"}
		if all {
			dockerArgs = append(dockerArgs, "-a")
		}

		dockerCmd := exec.Command("docker", dockerArgs...)
		output, err := dockerCmd.Output()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Unable to connect to Docker: %v\n", err)
			fmt.Fprintf(os.Stderr, "Is Docker running?\n")
			fmt.Println("")
		} else {
			fmt.Println("Running containers:")
			fmt.Println(string(output))
		}

		fmt.Println("---")
		fmt.Println("For full container list, run: docker ps -a")
		fmt.Println("For logs, run: docker logs <container-id>")
	},
}

func init() {
	statusCmd.Flags().BoolP("all", "a", false, "Show all containers (not just running)")
}
