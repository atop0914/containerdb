// Package commands provides CLI commands for containerdb.
package commands

import (
	"context"
	"fmt"
	"os"

	"github.com/atop0914/containerdb/internal/config"
	"github.com/atop0914/containerdb/pkg/compose"
	"github.com/spf13/cobra"
)

var composeCmd = &cobra.Command{
	Use:   "compose",
	Short: "Docker Compose integration",
	Long: `Manage database containers using Docker Compose.

ContainerDB can generate docker-compose.yml files and manage services
through the Docker Compose CLI. Supports both v1 (docker-compose) and
v2 (docker compose) commands.

Examples:
  containerdb compose init                    # Generate docker-compose.yml with defaults
  containerdb compose init -t postgres        # Generate for PostgreSQL only
  containerdb compose init --with-healthcheck # Include healthcheck configuration
  containerdb compose up                      # Start services
  containerdb compose up --wait               # Start and wait for healthchecks
  containerdb compose down                    # Stop services
  containerdb compose down -v                 # Stop and remove volumes
  containerdb compose status                  # Show service status
  containerdb compose logs                    # Show service logs`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var composeInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Generate a docker-compose.yml file",
	Long: `Generate a docker-compose.yml file with database service configurations.

The generated file includes proper healthcheck configurations, restart
policies, and volume definitions for persistent data.`,
	RunE: runComposeInit,
}

var composeUpCmd = &cobra.Command{
	Use:   "up",
	Short: "Start database services via docker compose",
	Long: `Start the database services defined in docker-compose.yml.

Use --wait to start services and wait for all healthchecks to pass
before returning. This ensures databases are ready to accept connections.`,
	RunE: runComposeUp,
}

var composeDownCmd = &cobra.Command{
	Use:   "down",
	Short: "Stop database services",
	Long: `Stop and remove the database services started by compose up.

Use -v to also remove named volumes (WARNING: this deletes all data).`,
	RunE: runComposeDown,
}

var composeStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show service status",
	Long:  `Show the current status of database services managed by docker compose.`,
	RunE:  runComposeStatus,
}

var composeLogsCmd = &cobra.Command{
	Use:   "logs [service]",
	Short: "Show service logs",
	Long:  `Show logs from database services. Optionally specify a service name.`,
	Args:  cobra.MaximumNArgs(1),
	RunE:  runComposeLogs,
}

// Compose flags
var (
	composeType        string
	composeImage       string
	composePort        string
	composeUsername    string
	composePassword    string
	composeDatabase    string
	composeOutput      string
	composeProject     string
	composeWithHealth  bool
	composeWait        bool
	composeWithVolumes bool
	composeVersion     string
)

func init() {
	// Init flags
	composeInitCmd.Flags().StringVarP(&composeType, "type", "t", "all", "Database type: mysql, postgres, or all")
	composeInitCmd.Flags().StringVar(&composeImage, "image", "", "Docker image (e.g., mysql:8.0)")
	composeInitCmd.Flags().StringVar(&composePort, "port", "", "Host port mapping")
	composeInitCmd.Flags().StringVarP(&composeUsername, "username", "u", "", "Database username")
	composeInitCmd.Flags().StringVarP(&composePassword, "password", "p", "", "Database password")
	composeInitCmd.Flags().StringVarP(&composeDatabase, "database", "d", "", "Database name")
	composeInitCmd.Flags().StringVarP(&composeOutput, "output", "o", "docker-compose.yml", "Output file path")
	composeInitCmd.Flags().StringVar(&composeProject, "project", "containerdb", "Docker Compose project name")
	composeInitCmd.Flags().BoolVar(&composeWithHealth, "with-healthcheck", true, "Include healthcheck configuration")

	// Up flags
	composeUpCmd.Flags().BoolVar(&composeWait, "wait", false, "Wait for healthchecks to pass")
	composeUpCmd.Flags().StringVar(&composeProject, "project", "containerdb", "Docker Compose project name")
	composeUpCmd.Flags().StringVarP(&composeOutput, "file", "f", "docker-compose.yml", "Compose file path")
	composeUpCmd.Flags().StringVar(&composeVersion, "version", "", "Force compose version: v1 or v2")

	// Down flags
	composeDownCmd.Flags().BoolVarP(&composeWithVolumes, "volumes", "v", false, "Remove named volumes")
	composeDownCmd.Flags().StringVar(&composeProject, "project", "containerdb", "Docker Compose project name")
	composeDownCmd.Flags().StringVarP(&composeOutput, "file", "f", "docker-compose.yml", "Compose file path")
	composeDownCmd.Flags().StringVar(&composeVersion, "version", "", "Force compose version: v1 or v2")

	// Status flags
	composeStatusCmd.Flags().StringVar(&composeProject, "project", "containerdb", "Docker Compose project name")
	composeStatusCmd.Flags().StringVarP(&composeOutput, "file", "f", "docker-compose.yml", "Compose file path")
	composeStatusCmd.Flags().StringVar(&composeVersion, "version", "", "Force compose version: v1 or v2")

	// Logs flags
	composeLogsCmd.Flags().StringVar(&composeProject, "project", "containerdb", "Docker Compose project name")
	composeLogsCmd.Flags().StringVarP(&composeOutput, "file", "f", "docker-compose.yml", "Compose file path")
	composeLogsCmd.Flags().StringVar(&composeVersion, "version", "", "Force compose version: v1 or v2")

	// Add subcommands
	composeCmd.AddCommand(composeInitCmd)
	composeCmd.AddCommand(composeUpCmd)
	composeCmd.AddCommand(composeDownCmd)
	composeCmd.AddCommand(composeStatusCmd)
	composeCmd.AddCommand(composeLogsCmd)
}

func runComposeInit(cmd *cobra.Command, args []string) error {
	services := make(map[string]compose.Service)

	switch composeType {
	case "mysql":
		cfg := config.DefaultMySQLConfig()
		if composeUsername != "" {
			cfg.Username = composeUsername
		}
		if composePassword != "" {
			cfg.Password = composePassword
		}
		if composeDatabase != "" {
			cfg.Database = composeDatabase
		}
		if composeImage != "" {
			cfg.Image = composeImage
		}

		svc := compose.GenerateMySQLService("mysql", cfg)
		if composeWithHealth {
			svc.AddHealthCheck("10s", "5s", 5, "CMD", "mysqladmin", "ping", "-h", "localhost")
			svc.SetRestart("unless-stopped")
		}
		if composePort != "" {
			svc.Ports = []string{composePort + ":3306"}
		}
		svc.Volumes = []string{"mysql_data:/var/lib/mysql"}
		services["mysql"] = svc

	case "postgres":
		cfg := config.DefaultPostgresConfig()
		if composeUsername != "" {
			cfg.Username = composeUsername
		}
		if composePassword != "" {
			cfg.Password = composePassword
		}
		if composeDatabase != "" {
			cfg.Database = composeDatabase
		}
		if composeImage != "" {
			cfg.Image = composeImage
		}

		svc := compose.GeneratePostgresService("postgres", cfg)
		if composeWithHealth {
			svc.AddHealthCheck("10s", "5s", 5, "CMD", "pg_isready", "-U", cfg.Username)
			svc.SetRestart("unless-stopped")
		}
		if composePort != "" {
			svc.Ports = []string{composePort + ":5432"}
		}
		svc.Volumes = []string{"pg_data:/var/lib/postgresql/data"}
		services["postgres"] = svc

	case "all":
		mysqlCfg := config.DefaultMySQLConfig()
		pgCfg := config.DefaultPostgresConfig()

		if composeUsername != "" {
			mysqlCfg.Username = composeUsername
		}
		if composePassword != "" {
			mysqlCfg.Password = composePassword
		}
		if composeDatabase != "" {
			mysqlCfg.Database = composeDatabase
			pgCfg.Database = composeDatabase
		}

		mysqlSvc := compose.GenerateMySQLService("mysql", mysqlCfg)
		if composeWithHealth {
			mysqlSvc.AddHealthCheck("10s", "5s", 5, "CMD", "mysqladmin", "ping", "-h", "localhost")
			mysqlSvc.SetRestart("unless-stopped")
		}
		mysqlSvc.Volumes = []string{"mysql_data:/var/lib/mysql"}
		services["mysql"] = mysqlSvc

		pgSvc := compose.GeneratePostgresService("postgres", pgCfg)
		if composeWithHealth {
			pgSvc.AddHealthCheck("10s", "5s", 5, "CMD", "pg_isready", "-U", pgCfg.Username)
			pgSvc.SetRestart("unless-stopped")
		}
		pgSvc.Volumes = []string{"pg_data:/var/lib/postgresql/data"}
		services["postgres"] = pgSvc

	default:
		return fmt.Errorf("unsupported type: %s (use mysql, postgres, or all)", composeType)
	}

	cf := compose.BuildComposeFile(services)
	if err := cf.WriteToFile(composeOutput); err != nil {
		return fmt.Errorf("failed to write compose file: %w", err)
	}

	fmt.Printf("✅ Generated %s with %d service(s)\n", composeOutput, len(services))
	fmt.Printf("   Type: %s\n", composeType)
	if composeWithHealth {
		fmt.Println("   Healthcheck: enabled")
	}
	fmt.Printf("\nTo start: containerdb compose up\n")
	return nil
}

func runComposeUp(cmd *cobra.Command, args []string) error {
	r := compose.NewRunnerWithFile(composeProject, composeOutput)

	if composeVersion != "" {
		r.SetVersion(compose.ComposeVersion(composeVersion))
	} else {
		if err := r.DetectVersion(); err != nil {
			return fmt.Errorf("docker compose not available: %w", err)
		}
	}

	ctx := context.Background()

	if composeWait {
		fmt.Println("🚀 Starting services and waiting for healthchecks...")
		if err := r.UpWithWait(ctx, "."); err != nil {
			return err
		}
		fmt.Println("✅ All services are healthy and ready!")
	} else {
		fmt.Println("🚀 Starting services...")
		if err := r.Up(ctx, "."); err != nil {
			return err
		}
		fmt.Println("✅ Services started!")
		fmt.Println("   Use 'containerdb compose status' to check status")
		fmt.Println("   Use 'containerdb compose logs' to view logs")
	}
	return nil
}

func runComposeDown(cmd *cobra.Command, args []string) error {
	r := compose.NewRunnerWithFile(composeProject, composeOutput)

	if composeVersion != "" {
		r.SetVersion(compose.ComposeVersion(composeVersion))
	} else {
		if err := r.DetectVersion(); err != nil {
			return fmt.Errorf("docker compose not available: %w", err)
		}
	}

	ctx := context.Background()

	if composeWithVolumes {
		fmt.Println("🛑 Stopping services and removing volumes...")
		if err := r.DownWithVolumes(ctx, "."); err != nil {
			return err
		}
		fmt.Println("✅ Services stopped and volumes removed")
	} else {
		fmt.Println("🛑 Stopping services...")
		if err := r.Down(ctx, "."); err != nil {
			return err
		}
		fmt.Println("✅ Services stopped")
	}
	return nil
}

func runComposeStatus(cmd *cobra.Command, args []string) error {
	r := compose.NewRunnerWithFile(composeProject, composeOutput)

	if composeVersion != "" {
		r.SetVersion(compose.ComposeVersion(composeVersion))
	} else {
		if err := r.DetectVersion(); err != nil {
			return fmt.Errorf("docker compose not available: %w", err)
		}
	}

	ctx := context.Background()
	output, err := r.Ps(ctx, ".")
	if err != nil {
		return err
	}

	if output == "" {
		fmt.Println("No services running")
		return nil
	}

	fmt.Print(output)
	return nil
}

func runComposeLogs(cmd *cobra.Command, args []string) error {
	r := compose.NewRunnerWithFile(composeProject, composeOutput)

	if composeVersion != "" {
		r.SetVersion(compose.ComposeVersion(composeVersion))
	} else {
		if err := r.DetectVersion(); err != nil {
			return fmt.Errorf("docker compose not available: %w", err)
		}
	}

	ctx := context.Background()
	service := ""
	if len(args) > 0 {
		service = args[0]
	}

	output, err := r.Logs(ctx, ".", service)
	if err != nil {
		return err
	}

	if output == "" {
		fmt.Println("No logs available")
		return nil
	}

	fmt.Fprint(os.Stdout, output)
	return nil
}
