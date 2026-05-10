# Examples

Real-world usage patterns for ContainerDB.

## Table of Contents

- [Unit Testing a Repository](#unit-testing-a-repository)
- [Integration Testing an API](#integration-testing-an-api)
- [Testing Migrations](#testing-migrations)
- [Testing with Docker Compose](#testing-with-docker-compose)
- [Benchmark Setup](#benchmark-setup)
- [CI/CD Integration](#cicd-integration)

---

## Unit Testing a Repository

Test a user repository against a real MySQL database:

```go
package repo_test

import (
    "context"
    "testing"

    "github.com/atop0914/containerdb-bootcamp/pkg/mysql"
    "github.com/atop0914/containerdb-bootcamp/pkg/migrate"
)

type User struct {
    ID    int
    Name  string
    Email string
}

type UserRepo struct {
    db interface{ ExecContext, QueryRowContext }
}

func TestUserRepo_Create(t *testing.T) {
    ctx := context.Background()
    db, cleanup, err := mysql.New(ctx)
    if err != nil {
        t.Fatal(err)
    }
    defer cleanup()

    // Setup schema
    _, err = db.ExecContext(ctx, `
        CREATE TABLE users (
            id INT AUTO_INCREMENT PRIMARY KEY,
            name VARCHAR(255) NOT NULL,
            email VARCHAR(255) UNIQUE NOT NULL
        )
    `)
    if err != nil {
        t.Fatal(err)
    }

    // Test insert
    _, err = db.ExecContext(ctx, "INSERT INTO users (name, email) VALUES (?, ?)", "Alice", "alice@test.com")
    if err != nil {
        t.Fatal(err)
    }

    // Test query
    var user User
    err = db.QueryRowContext(ctx, "SELECT id, name, email FROM users WHERE email = ?", "alice@test.com").
        Scan(&user.ID, &user.Name, &user.Email)
    if err != nil {
        t.Fatal(err)
    }

    if user.Name != "Alice" {
        t.Errorf("expected Alice, got %s", user.Name)
    }
}
```

---

## Integration Testing an API

Test an HTTP handler with a real PostgreSQL database:

```go
func TestAPIHandler(t *testing.T) {
    ctx := context.Background()
    db, cleanup, err := postgres.New(ctx)
    if err != nil {
        t.Fatal(err)
    }
    defer cleanup()

    // Setup
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

    // Create handler with real DB
    handler := NewProductHandler(db)

    // Test HTTP
    req := httptest.NewRequest("POST", "/products", strings.NewReader(`{"name":"Widget","price":9.99}`))
    rec := httptest.NewRecorder()
    handler.Create(rec, req)

    if rec.Code != 201 {
        t.Errorf("expected 201, got %d", rec.Code)
    }
}
```

---

## Testing Migrations

```go
func TestMigrations(t *testing.T) {
    ctx := context.Background()
    db, cleanup, err := postgres.New(ctx)
    if err != nil {
        t.Fatal(err)
    }
    defer cleanup()

    runner := migrate.NewRunner(db,
        migrate.WithDir("testdata/migrations"),
    )

    // Run all migrations
    if err := runner.Up(ctx); err != nil {
        t.Fatal(err)
    }

    // Check status
    statuses, err := runner.Status(ctx)
    if err != nil {
        t.Fatal(err)
    }

    for _, s := range statuses {
        if !s.Applied {
            t.Errorf("migration %s not applied", s.Name)
        }
    }

    // Rollback last migration
    if err := runner.Down(ctx, 1); err != nil {
        t.Fatal(err)
    }
}
```

---

## Testing with Docker Compose

```go
func TestWithCompose(t *testing.T) {
    runner := compose.NewRunner("mytest")
    runner.SetVersion(compose.VersionV2)

    ctx := context.Background()

    // Start services
    if err := runner.Up(ctx, true); err != nil {
        t.Fatal(err)
    }
    defer runner.Down(ctx, false)

    // Services are ready — run tests
    // ...
}
```

---

## Benchmark Setup

```go
func BenchmarkQuery(b *testing.B) {
    ctx := context.Background()
    db, cleanup, err := mysql.New(ctx)
    if err != nil {
        b.Fatal(err)
    }
    defer cleanup()

    // Seed data
    for i := 0; i < 1000; i++ {
        db.ExecContext(ctx, "INSERT INTO users (name) VALUES (?)", fmt.Sprintf("user_%d", i))
    }

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        var name string
        db.QueryRowContext(ctx, "SELECT name FROM users WHERE id = ?", i%1000).Scan(&name)
    }
}
```

---

## CI/CD Integration

### GitHub Actions

```yaml
name: Tests
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'

      - name: Run tests
        run: go test -v -race ./...

      - name: Run benchmarks
        run: go test -bench=. -benchmem ./...
```

### GitLab CI

```yaml
test:
  image: golang:1.22
  services:
    - docker:dind
  variables:
    DOCKER_HOST: tcp://docker:2375
  script:
    - go test -v -race ./...
```

### Docker-in-Docker Considerations

When running in CI with Docker-in-Docker:
- Ensure the Docker socket is available
- Set `TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE` if needed
- Use `ryuk.disabled=true` for some CI environments

```bash
export TESTCONTAINERS_RYUK_DISABLED=true
go test ./...
```
