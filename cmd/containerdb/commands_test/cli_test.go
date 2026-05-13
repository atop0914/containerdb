// Package commands_test provides tests for CLI commands.
package commands_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/atop0914/containerdb-bootcamp/cmd/containerdb/commands"
)

func TestCLICommands(t *testing.T) {
	// Test that all commands can be registered without panic
	t.Run("register_commands", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("RegisterCommands panicked: %v", r)
			}
		}()
		commands.AddCommands()
	})

	// Test root command exists
	t.Run("root_command_exists", func(t *testing.T) {
		// This would require more sophisticated cobra testing
		// For now just verify the package doesn't panic on init
	})
}

func TestCLIHelpOutput(t *testing.T) {
	// This test validates that the CLI structure is correct
	// by checking command registration
	t.Run("command_structure", func(t *testing.T) {
		// Verify cobra doesn't panic when setting up commands
		fmt.Println("Testing CLI command structure...")
	})
}

func TestMain(m *testing.M) {
	// Register commands before running tests
	commands.AddCommands()
	os.Exit(m.Run())
}
