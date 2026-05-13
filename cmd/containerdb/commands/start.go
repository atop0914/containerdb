// Package commands provides CLI commands for containerdb.
package commands

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/atop0914/containerdb-bootcamp/pkg/mysql"
	"github.com/atop0914/containerdb-bootcamp/pkg/postgres"
	"github.com/spf13/cobra"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start a database container",
	Long:  `Start a MySQL, PostgreSQL, or SQLite container for development and testing.`,
	Run: func(cmd *cobra.Command, args []string) {
		dbType, _ := cmd.Flags().GetString("type")
		image, _ := cmd.Flags().GetString("image")
		username, _ := cmd.Flags().GetString("username")
		password, _ := cmd.Flags().GetString("password")
		database, _ := cmd.Flags().GetString("database")
		timeout, _ := cmd.Flags().GetDuration("timeout")

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		fmt.Printf("Starting %s container...\n", dbType)

		switch dbType {
		case "mysql":
			opts := []mysql.Option{
				mysql.WithHealthCheckTimeout(timeout),
			}
			if image != "" {
				opts = append(opts, mysql.WithImage(image))
			}
			if username != "" {
				opts = append(opts, mysql.WithUsername(username))
			}
			if password != "" {
				opts = append(opts, mysql.WithPassword(password))
			}
			if database != "" {
				opts = append(opts, mysql.WithDatabase(database))
			}

			db, cleanup, err := mysql.NewWithOptions(ctx, opts...)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to start MySQL: %v\n", err)
				os.Exit(1)
			}
			defer cleanup()

			if err := db.PingContext(ctx); err != nil {
				fmt.Fprintf(os.Stderr, "MySQL not ready: %v\n", err)
				os.Exit(1)
			}

			fmt.Println("MySQL container started successfully!")
			fmt.Println("Container is running and accepting connections.")

		case "postgres":
			opts := []postgres.Option{
				postgres.WithHealthCheckTimeout(timeout),
			}
			if image != "" {
				opts = append(opts, postgres.WithImage(image))
			}
			if username != "" {
				opts = append(opts, postgres.WithUsername(username))
			}
			if password != "" {
				opts = append(opts, postgres.WithPassword(password))
			}
			if database != "" {
				opts = append(opts, postgres.WithDatabase(database))
			}

			db, cleanup, err := postgres.NewWithOptions(ctx, opts...)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to start PostgreSQL: %v\n", err)
				os.Exit(1)
			}
			defer cleanup()

			if err := db.PingContext(ctx); err != nil {
				fmt.Fprintf(os.Stderr, "PostgreSQL not ready: %v\n", err)
				os.Exit(1)
			}

			fmt.Println("PostgreSQL container started successfully!")
			fmt.Println("Container is running and accepting connections.")

		default:
			fmt.Fprintf(os.Stderr, "Unsupported database type: %s\n", dbType)
			fmt.Fprintf(os.Stderr, "Supported types: mysql, postgres\n")
			os.Exit(1)
		}
	},
}

func init() {
	startCmd.Flags().StringP("type", "t", "mysql", "Database type (mysql, postgres)")
	startCmd.Flags().StringP("image", "i", "", "Docker image to use")
	startCmd.Flags().StringP("username", "u", "root", "Database username")
	startCmd.Flags().StringP("password", "p", "root", "Database password")
	startCmd.Flags().StringP("database", "d", "testdb", "Database name")
	startCmd.Flags().DurationP("timeout", "", 60*time.Second, "Timeout for starting container")
}
