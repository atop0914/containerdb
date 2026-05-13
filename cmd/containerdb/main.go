// ContainerDB CLI - Manage database containers from command line
package main

import (
	"fmt"
	"os"

	"github.com/atop0914/containerdb-bootcamp/cmd/containerdb/commands"
)

func main() {
	// Register all commands
	commands.AddCommands()

	// Execute the root command
	if err := commands.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
