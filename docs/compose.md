# Docker Compose Integration

ContainerDB can generate and manage Docker Compose files for database services.

## Overview

The compose integration provides:
- Generate `docker-compose.yml` from ContainerDB configurations
- Support for both Docker Compose V1 (`docker-compose`) and V2 (`docker compose`)
- Health check configuration
- Service dependency management
- Restart policies

## Generating Compose Files

### CLI

```bash
# Generate with all databases
containerdb compose init

# Generate for specific database
containerdb compose init -t postgres

# Include health checks
containerdb compose init --with-healthcheck
```

### Go API

```go
import "github.com/atop0914/containerdb-bootcamp/pkg/compose"

// Build a compose file programmatically
services := []compose.Service{
    {
        Name:  "mysql",
        Image: "mysql:8.0",
        Ports: []string{"3306:3306"},
        Environment: map[string]string{
            "MYSQL_ROOT_PASSWORD": "rootpassword",
            "MYSQL_DATABASE":      "testdb",
        },
        HealthCheck: &compose.HealthCheck{
            Test:        []string{"CMD", "mysqladmin", "ping", "-h", "localhost"},
            Interval:    "10s",
            Timeout:     "5s",
            Retries:     5,
            StartPeriod: "30s",
        },
    },
}

composeFile := compose.BuildComposeFile(services...)
```

## Running Services

### CLI

```bash
# Start services
containerdb compose up

# Start and wait for health checks
containerdb compose up --wait

# Stop services
containerdb compose down

# Stop and remove volumes
containerdb compose down -v

# View status
containerdb compose status

# View logs
containerdb compose logs -s mysql
```

### Go API

```go
runner := compose.NewRunner("myproject")
runner.SetVersion(compose.VersionV2)

ctx := context.Background()

// Start with health check wait
err := runner.Up(ctx, true)

// Stop without removing volumes
err := runner.Down(ctx, false)

// Get status
status, err := runner.Status(ctx)

// Get logs
logs, err := runner.Logs(ctx, "mysql")
```

## Compose File Example

A generated `docker-compose.yml` with health checks:

```yaml
version: "3.8"

services:
  mysql:
    image: mysql:8.0
    ports:
      - "3306:3306"
    environment:
      MYSQL_ROOT_PASSWORD: rootpassword
      MYSQL_DATABASE: testdb
    volumes:
      - mysql_data:/var/lib/mysql
    healthcheck:
      test: ["CMD", "mysqladmin", "ping", "-h", "localhost"]
      interval: 10s
      timeout: 5s
      retries: 5
      start_period: 30s
    restart: unless-stopped
    depends_on:
      - postgres

  postgres:
    image: postgres:16-alpine
    ports:
      - "5432:5432"
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
      POSTGRES_DB: testdb
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 10s
      timeout: 5s
      retries: 5
      start_period: 30s
    restart: unless-stopped

volumes:
  mysql_data:
  postgres_data:
```

## Version Detection

ContainerDB automatically detects the available Docker Compose version:

```go
version, err := compose.DetectComposeVersion()
switch version {
case compose.VersionV2:
    // Uses: docker compose
case compose.VersionV1:
    // Uses: docker-compose
}
```

## Parsing Existing Compose Files

```go
// From file
cf, err := compose.Parse("docker-compose.yml")

// From string
cf, err := compose.ParseFromString(yamlContent)

// Access services
for _, svc := range cf.Services {
    fmt.Printf("Service: %s, Image: %s\n", svc.Name, svc.Image)
}
```

## Best Practices

1. **Always use health checks** — ensures services are ready before tests run
2. **Use volumes for persistence** — avoid data loss during development
3. **Set restart policies** — `unless-stopped` for dev, `always` for production
4. **Use named networks** — for service discovery in multi-container setups
5. **Pin image versions** — avoid surprises from latest tags
