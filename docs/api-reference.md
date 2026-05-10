# API Reference

## pkg/mysql

MySQL container management for testing.

### Functions

#### `New(ctx context.Context) (*sql.DB, func(), error)`
Starts a MySQL container with default settings and returns a connected `*sql.DB`.
The cleanup function stops and removes the container.

**Defaults**: mysql:8.0, root/rootpassword, testdb, port random

```go
db, cleanup, err := mysql.New(ctx)
defer cleanup()
```

#### `NewWithOptions(ctx context.Context, opts ...Option) (*sql.DB, func(), error)`
Starts a MySQL container with custom options.

```go
db, cleanup, err := mysql.NewWithOptions(ctx,
    mysql.WithImage("mysql:8.4"),
    mysql.WithDatabase("mydb"),
)
```

### Options

| Option | Description | Default |
|--------|-------------|---------|
| `WithImage(image string)` | Docker image | `mysql:8.0` |
| `WithUsername(username string)` | MySQL username | `root` |
| `WithPassword(password string)` | MySQL password | `rootpassword` |
| `WithDatabase(database string)` | Database name | `testdb` |
| `WithHealthCheckTimeout(timeout time.Duration)` | Max wait for readiness | `30s` |
| `WithHealthCheckRetry(retries int)` | Connection retry count | `3` |
| `WithHealthCheckInterval(interval time.Duration)` | Time between retries | `500ms` |
| `WithPoolSettings(maxOpen, maxIdle int, maxLifetime time.Duration)` | Connection pool config | `10, 5, 1h` |

---

## pkg/postgres

PostgreSQL container management for testing.

### Functions

#### `New(ctx context.Context) (*sql.DB, func(), error)`
Starts a PostgreSQL container with default settings.

**Defaults**: postgres:16-alpine, postgres/postgres, testdb

```go
db, cleanup, err := postgres.New(ctx)
defer cleanup()
```

#### `NewWithOptions(ctx context.Context, opts ...Option) (*sql.DB, func(), error)`
Starts a PostgreSQL container with custom options.

### Options

| Option | Description | Default |
|--------|-------------|---------|
| `WithImage(image string)` | Docker image | `postgres:16-alpine` |
| `WithUsername(username string)` | PostgreSQL username | `postgres` |
| `WithPassword(password string)` | PostgreSQL password | `postgres` |
| `WithDatabase(database string)` | Database name | `testdb` |
| `WithHealthCheckTimeout(timeout time.Duration)` | Max wait for readiness | `30s` |
| `WithHealthCheckRetry(retries int)` | Connection retry count | `3` |
| `WithHealthCheckInterval(interval time.Duration)` | Time between retries | `500ms` |
| `WithPoolSettings(maxOpen, maxIdle int, maxLifetime time.Duration)` | Connection pool config | `10, 5, 1h` |

---

## pkg/sqlite

SQLite helpers — no containers needed.

### Functions

#### `InMemory() (*sql.DB, func(), error)`
Creates an in-memory SQLite database. Fastest option for unit tests.

```go
db, cleanup, err := sqlite.InMemory()
defer cleanup()
```

#### `TempDB() (*sql.DB, func(), error)`
Creates a temporary file-based SQLite database. File is auto-removed on cleanup.

```go
db, cleanup, err := sqlite.TempDB()
defer cleanup()
```

#### `NewWithOptions(opts ...Option) (*sql.DB, func(), error)`
Creates a SQLite database with custom options.

```go
db, cleanup, err := sqlite.NewWithOptions(
    sqlite.WithMode("memory"),
    sqlite.WithCache("shared"),
)
```

### Options

| Option | Description | Default |
|--------|-------------|---------|
| `WithMode(mode string)` | Database mode: `memory`, `temp`, `file` | `memory` |
| `WithPath(path string)` | File path (for `file` mode) | — |
| `WithCache(cache string)` | Cache mode: `shared`, `private` | — |

---

## pkg/migrate

Database migration runner.

### Types

#### `Runner`
Handles database migrations.

```go
runner := migrate.NewRunner(db,
    migrate.WithDir("migrations"),
    migrate.WithTableName("schema_migrations"),
    migrate.WithTimeout(30*time.Second),
)
```

### Functions

#### `NewRunner(db *sql.DB, opts ...Option) *Runner`
Creates a new migration runner.

#### `(*Runner) Up(ctx context.Context) error`
Runs all pending up migrations.

#### `(*Runner) Down(ctx context.Context, steps int) error`
Rolls back the last N migrations. Pass `0` to roll back all.

#### `(*Runner) Status(ctx context.Context) ([]Status, error)`
Returns the status of all migrations (applied/pending).

#### `CreateMigration(dir, name string) (string, error)`
Creates a new migration file pair (up and down) with timestamp prefix.

#### `ListMigrations(dir string) ([]Migration, error)`
Returns all available migrations from the directory.

#### `ForceVersion(ctx context.Context, db *sql.DB, version string, opts ...Option) error`
Sets the schema version without running migrations (for recovery).

### Migration File Format

Migration files follow the pattern: `{timestamp}_{name}.up.sql` and `{timestamp}_{name}.down.sql`

```
migrations/
├── 20260501120000_create_users.up.sql
├── 20260501120000_create_users.down.sql
├── 20260502150000_add_email_index.up.sql
└── 20260502150000_add_email_index.down.sql
```

---

## pkg/compose

Docker Compose integration.

### Types

#### `Runner`
Manages docker-compose operations.

```go
runner := compose.NewRunner("myproject")
runner.SetVersion(compose.VersionV2)
```

### Functions

#### `NewRunner(projectName string) *Runner`
Creates a runner with default compose file (`docker-compose.yml`).

#### `NewRunnerWithFile(projectName, composeFile string) *Runner`
Creates a runner with a custom compose file path.

#### `(*Runner) Up(ctx context.Context, wait bool) error`
Starts services. If `wait` is true, waits for health checks.

#### `(*Runner) Down(ctx context.Context, volumes bool) error`
Stops services. If `volumes` is true, removes volumes too.

#### `(*Runner) Status(ctx context.Context) (string, error)`
Returns service status output.

#### `(*Runner) Logs(ctx context.Context, service string) (string, error)`
Returns service logs.

#### `BuildComposeFile(services ...Service) *ComposeFile`
Creates a compose file from services.

#### `Parse(path string) (*ComposeFile, error)`
Reads and parses a compose file.

#### `ParseFromString(content string) (*ComposeFile, error)`
Parses compose content from a string.

#### `DetectComposeVersion() (ComposeVersion, error)`
Detects whether docker compose v2 or v1 is available.

### Re-exported Types

- `Service` — Represents a compose service
- `HealthCheck` — Docker health check configuration
- `ComposeFile` — Full compose file structure

---

## internal/config

Configuration types and defaults (not directly importable).

### Default Configs

| Type | Function | Default Values |
|------|----------|---------------|
| `MySQLConfig` | `DefaultMySQLConfig()` | Image: mysql:8.0, User: root, Pass: rootpassword, DB: testdb |
| `PostgresConfig` | `DefaultPostgresConfig()` | Image: postgres:16-alpine, User: postgres, Pass: postgres, DB: testdb |
| `SQLiteConfig` | `DefaultSQLiteConfig()` | Mode: memory |

### Config Fields

All database configs share common fields:
- `HealthCheckTimeout` (30s) — Max time to wait for container readiness
- `HealthCheckRetries` (3) — Number of connection retry attempts
- `HealthCheckInterval` (500ms) — Time between retries
- `MaxOpenConns` (10) — Maximum open connections
- `MaxIdleConns` (5) — Maximum idle connections
- `ConnMaxLifetime` (1h) — Max connection lifetime
