# ContainerDB
[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go)](https://go.dev/)

A lightweight containerized database toolkit for Go development and testing. Spin up real databases in containers with a single function call — no Docker Compose required.

[![Go Reference](https://pkg.go.dev/badge/github.com/atop0914/containerdb.svg)](https://pkg.go.dev/github.com/atop0914/containerdb)
[![Go Report Card](https://goreportcard.com/badge/github.com/atop0914/containerdb)](https://goreportcard.com/report/github.com/atop0914/containerdb)

## Features

- **MySQL, PostgreSQL, SQLite** support out of the box
- **One-line setup**: `db, cleanup, err := mysql.New(ctx)` — that's it
- **Auto-cleanup**: Containers are automatically cleaned up when the test ends
- **Random port allocation**: No port conflicts between parallel tests
- **Configurable**: Custom image tags, credentials, volumes, health checks
- **Connection pooling**: Pre-configured pool settings with easy tuning
- **Built-in migrations**: Schema migration runner with rollback support
- **Docker Compose**: Generate and manage compose files programmatically
- **CLI tool**: Start/stop/status from the command line

## Quick Start

```go
package mytest

import (
    "context"
    "testing"

    "github.com/atop0914/containerdb/pkg/mysql"
)

func TestDatabase(t *testing.T) {
    ctx := context.Background()

    // Spin up a MySQL container
    db, cleanup, err := mysql.New(ctx)
    if err != nil {
        t.Fatal(err)
    }
    defer cleanup()

    // Use db as a normal *sql.DB
    _, err = db.ExecContext(ctx, "CREATE TABLE users (id INT AUTO_INCREMENT PRIMARY KEY, name VARCHAR(255))")
    if err != nil {
        t.Fatal(err)
    }

    _, err = db.ExecContext(ctx, "INSERT INTO users (name) VALUES (?)", "Alice")
    if err != nil {
        t.Fatal(err)
    }

    var name string
    err = db.QueryRowContext(ctx, "SELECT name FROM users WHERE id = 1").Scan(&name)
    if err != nil {
        t.Fatal(err)
    }

    if name != "Alice" {
        t.Errorf("expected Alice, got %s", name)
    }
}
```

## Supported Databases

| Database     | Container | In-Memory | Default Image |
|-------------|-----------|-----------|---------------|
| MySQL       | ✅        | ❌        | `mysql:8.0` |
| PostgreSQL  | ✅        | ❌        | `postgres:16-alpine` |
| SQLite      | ❌        | ✅        | N/A (no container) |

## Installation

```bash
go get github.com/atop0914/containerdb
```

### CLI Tool

```bash
go install github.com/atop0914/containerdb/cmd/containerdb@latest
```

## Usage

### MySQL

```go
import "github.com/atop0914/containerdb/pkg/mysql"

// Default settings
db, cleanup, err := mysql.New(ctx)
defer cleanup()

// Custom settings
db, cleanup, err := mysql.NewWithOptions(ctx,
    mysql.WithImage("mysql:8.4"),
    mysql.WithUsername("testuser"),
    mysql.WithPassword("testpass"),
    mysql.WithDatabase("myapp_test"),
    mysql.WithPoolSettings(20, 10, 30*time.Minute),
)
defer cleanup()
```

### PostgreSQL

```go
import "github.com/atop0914/containerdb/pkg/postgres"

// Default settings
db, cleanup, err := postgres.New(ctx)
defer cleanup()

// Custom settings
db, cleanup, err := postgres.NewWithOptions(ctx,
    postgres.WithImage("postgres:17"),
    postgres.WithDatabase("myapp_test"),
)
defer cleanup()
```

### SQLite

```go
import "github.com/atop0914/containerdb/pkg/sqlite"

// In-memory (fastest)
db, cleanup, err := sqlite.InMemory()
defer cleanup()

// Temporary file (auto-cleaned)
db, cleanup, err := sqlite.TempDB()
defer cleanup()
```

### Migrations

```go
import "github.com/atop0914/containerdb/pkg/migrate"

runner := migrate.NewRunner(db,
    migrate.WithDir("migrations"),
    migrate.WithTableName("schema_migrations"),
)

// Run all pending migrations
err := runner.Up(ctx)

// Rollback last 2 migrations
err := runner.Down(ctx, 2)

// Check migration status
statuses, err := runner.Status(ctx)

// Create new migration files
path, err := migrate.CreateMigration("migrations", "add_users_table")
```

### Docker Compose

```go
import "github.com/atop0914/containerdb/pkg/compose"

runner := compose.NewRunner("myproject")

// Start services
err := runner.Up(ctx, true) // true = wait for health checks

// Stop services
err := runner.Down(ctx, false) // false = keep volumes

// Generate compose file
services := []compose.Service{...}
cf := compose.BuildComposeFile(services...)
```

### CLI

```bash
# Start MySQL container
containerdb start

# Start PostgreSQL
containerdb start -t postgres

# Check status
containerdb status

# Generate docker-compose.yml
containerdb compose init --with-healthcheck

# Start compose services
containerdb compose up --wait
```

## Architecture

```
┌─────────────────────────────────────────────────┐
│                  Your Application                │
├────────────┬────────────┬───────────┬───────────┤
│ pkg/mysql  │pkg/postgres│pkg/sqlite │pkg/migrate│
├────────────┴────────────┴───────────┴───────────┤
│              internal/ packages                  │
│  config/  database/  container/  health/  pool/  │
├─────────────────────────────────────────────────┤
│          testcontainers-go / Docker              │
└─────────────────────────────────────────────────┘
```

### Package Layout

| Package | Description |
|---------|-------------|
| `pkg/mysql` | MySQL container management |
| `pkg/postgres` | PostgreSQL container management |
| `pkg/sqlite` | SQLite helpers (no container) |
| `pkg/migrate` | Database migration runner |
| `pkg/compose` | Docker Compose integration |
| `internal/config` | Configuration types and defaults |
| `internal/database` | Common database interfaces |
| `internal/container` | Base container utilities |
| `internal/health` | Health check logic |
| `internal/pool` | Connection pool management |
| `internal/compose` | Compose file generation engine |
| `internal/migrate` | Migration engine |
| `cmd/containerdb` | CLI tool |

## Documentation

- **[Getting Started](docs/getting-started.md)** — Installation and first steps
- **[API Reference](docs/api-reference.md)** — Complete Go package API
- **[CLI Reference](docs/cli-reference.md)** — Command-line tool usage
- **[Configuration](docs/configuration.md)** — All options and defaults
- **[Examples](docs/examples.md)** — Real-world usage patterns
- **[Docker Compose](docs/compose.md)** — Compose integration guide

## Examples

See the [`examples/`](examples/) directory for working code:

- [`examples/mysql_example_test.go`](examples/mysql_example_test.go) — MySQL usage
- [`examples/postgres_example_test.go`](examples/postgres_example_test.go) — PostgreSQL usage
- [`examples/sqlite_example_test.go`](examples/sqlite_example_test.go) — SQLite usage
- [`examples/migrate/`](examples/migrate/) — Migration workflow
- [`examples/compose/`](examples/compose/) — Docker Compose integration

## Motivation

Testcontainers (JVM) is too heavy for Go projects. This library provides the same experience with a native Go API, zero external dependencies beyond `testcontainers-go`, and Go-idiomatic functional options.

## Requirements

- Go 1.22+
- Docker running locally
- (Optional) Docker Compose V1 or V2

## License

MIT
