# Getting Started

## Prerequisites

- **Go 1.22+** (or 1.25+ for latest features)
- **Docker** running on your machine
- **Git** for cloning the repository

## Installation

```bash
go get github.com/atop0914/containerdb-bootcamp
```

Or add it to your `go.mod`:

```bash
cd your-project
go get github.com/atop0914/containerdb-bootcamp/pkg/mysql
go get github.com/atop0914/containerdb-bootcamp/pkg/postgres
go get github.com/atop0914/containerdb-bootcamp/pkg/sqlite
```

## Quick Start: MySQL

```go
package mytest

import (
    "context"
    "testing"

    "github.com/atop0914/containerdb-bootcamp/pkg/mysql"
)

func TestWithMySQL(t *testing.T) {
    ctx := context.Background()

    // Start a MySQL container (takes ~5-10 seconds on first run)
    db, cleanup, err := mysql.New(ctx)
    if err != nil {
        t.Fatal(err)
    }
    defer cleanup()

    // Use db as a normal *sql.DB
    _, err = db.ExecContext(ctx, `
        CREATE TABLE users (
            id INT AUTO_INCREMENT PRIMARY KEY,
            name VARCHAR(255) NOT NULL,
            email VARCHAR(255) UNIQUE
        )
    `)
    if err != nil {
        t.Fatal(err)
    }

    // Insert and query
    _, err = db.ExecContext(ctx, "INSERT INTO users (name, email) VALUES (?, ?)", "Alice", "alice@example.com")
    if err != nil {
        t.Fatal(err)
    }

    var name string
    err = db.QueryRowContext(ctx, "SELECT name FROM users WHERE email = ?", "alice@example.com").Scan(&name)
    if err != nil {
        t.Fatal(err)
    }

    if name != "Alice" {
        t.Errorf("expected Alice, got %s", name)
    }
}
```

## Quick Start: PostgreSQL

```go
func TestWithPostgres(t *testing.T) {
    ctx := context.Background()

    db, cleanup, err := postgres.New(ctx)
    if err != nil {
        t.Fatal(err)
    }
    defer cleanup()

    _, err = db.ExecContext(ctx, `
        CREATE TABLE products (
            id SERIAL PRIMARY KEY,
            name TEXT NOT NULL,
            price DECIMAL(10,2)
        )
    `)
    if err != nil {
        t.Fatal(err)
    }

    _, err = db.ExecContext(ctx, "INSERT INTO products (name, price) VALUES ($1, $2)", "Widget", 9.99)
    if err != nil {
        t.Fatal(err)
    }
}
```

## Quick Start: SQLite

```go
func TestWithSQLite(t *testing.T) {
    // In-memory — fastest, no cleanup needed
    db, cleanup, err := sqlite.InMemory()
    if err != nil {
        t.Fatal(err)
    }
    defer cleanup()

    _, err = db.ExecContext(context.Background(), `
        CREATE TABLE items (id INTEGER PRIMARY KEY, name TEXT)
    `)
    if err != nil {
        t.Fatal(err)
    }
}
```

## Custom Configuration

```go
// Custom MySQL with specific image and credentials
db, cleanup, err := mysql.NewWithOptions(ctx,
    mysql.WithImage("mysql:8.4"),
    mysql.WithUsername("testuser"),
    mysql.WithPassword("testpass"),
    mysql.WithDatabase("myapp_test"),
    mysql.WithHealthCheckTimeout(60*time.Second),
    mysql.WithPoolSettings(20, 10, 30*time.Minute),
)
if err != nil {
    t.Fatal(err)
}
defer cleanup()
```

## Parallel Testing

Each call to `mysql.New()` or `postgres.New()` creates a fresh container with a random port, so tests can run in parallel safely:

```go
func TestParallelA(t *testing.T) {
    t.Parallel()
    db, cleanup, err := mysql.New(context.Background())
    if err != nil {
        t.Fatal(err)
    }
    defer cleanup()
    // This gets its own isolated MySQL instance
}

func TestParallelB(t *testing.T) {
    t.Parallel()
    db, cleanup, err := mysql.New(context.Background())
    if err != nil {
        t.Fatal(err)
    }
    defer cleanup()
    // This gets a different MySQL instance on a different port
}
```

## Using with Test Suites (testify)

```go
type UserRepoTestSuite struct {
    suite.Suite
    db      *sql.DB
    cleanup func()
}

func (s *UserRepoTestSuite) SetupSuite() {
    ctx := context.Background()
    db, cleanup, err := mysql.New(ctx)
    s.Require().NoError(err)
    s.db = db
    s.cleanup = cleanup

    // Run migrations
    runner := migrate.NewRunner(s.db, migrate.WithDir("migrations"))
    s.Require().NoError(runner.Up(ctx))
}

func (s *UserRepoTestSuite) TearDownSuite() {
    s.cleanup()
}

func TestUserRepo(t *testing.T) {
    suite.Run(t, new(UserRepoTestSuite))
}
```

## Troubleshooting

### Docker not running
```
Error: Cannot connect to the Docker daemon
```
**Fix**: Start Docker Desktop or run `sudo systemctl start docker`

### Image pull timeout
```
Error: context deadline exceeded
```
**Fix**: Increase timeout: `mysql.WithHealthCheckTimeout(120*time.Second)`
Or pre-pull: `docker pull mysql:8.0`

### Port conflicts
ContainerDB uses random ports, so conflicts are rare. If they occur, ensure no other services are binding random ports.
