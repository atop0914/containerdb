# ContainerDB Documentation

Welcome to the ContainerDB documentation. ContainerDB is a lightweight containerized database toolkit for Go development and testing.

## Table of Contents

- [Getting Started](getting-started.md) — Installation, quick setup, and first steps
- [API Reference](api-reference.md) — Complete Go package API documentation
- [CLI Reference](cli-reference.md) — Command-line tool usage
- [Configuration](configuration.md) — All configuration options and defaults
- [Examples](examples.md) — Real-world usage patterns and recipes
- [Docker Compose Integration](compose.md) — Using ContainerDB with Docker Compose

## Overview

ContainerDB lets you spin up real database containers with a single function call:

```go
db, cleanup, err := mysql.New(ctx)
if err != nil {
    t.Fatal(err)
}
defer cleanup()

// db is a *sql.DB — use it normally
db.ExecContext(ctx, "CREATE TABLE users (id INT, name TEXT)")
```

### Supported Databases

| Database     | Container Mode | In-Memory Mode | Notes                     |
|-------------|---------------|---------------|---------------------------|
| MySQL       | ✅            | ❌            | mysql:8.0 default         |
| PostgreSQL  | ✅            | ❌            | postgres:16-alpine default |
| SQLite      | ❌            | ✅            | No container needed       |

### Key Features

- **Zero configuration** — sensible defaults for everything
- **Functional options** — customize anything via `WithX()` options
- **Auto-cleanup** — containers destroyed when tests finish
- **Random ports** — no conflicts in parallel test runs
- **Health checks** — built-in readiness verification
- **Connection pooling** — pre-configured pool settings
- **Docker Compose** — generate and manage compose files
- **Migration support** — built-in schema migration helpers

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

- **`pkg/`** — Public API packages (import these in your code)
  - `pkg/mysql` — MySQL container management
  - `pkg/postgres` — PostgreSQL container management
  - `pkg/sqlite` — SQLite helpers (no container)
  - `pkg/migrate` — Database migration runner
  - `pkg/compose` — Docker Compose integration

- **`internal/`** — Private implementation (not importable)
  - `internal/config` — Configuration types and defaults
  - `internal/database` — Common database interfaces
  - `internal/container` — Base container utilities
  - `internal/health` — Health check logic
  - `internal/pool` — Connection pool management
  - `internal/compose` — Compose file generation
  - `internal/migrate` — Migration engine

- **`cmd/containerdb/`** — CLI tool
- **`examples/`** — Working example code
