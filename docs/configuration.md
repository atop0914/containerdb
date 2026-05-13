# Configuration

ContainerDB uses functional options for configuration. All settings have sensible defaults — you only need to configure what you want to change.

## MySQL Configuration

### Defaults

```go
config.MySQLConfig{
    Image:                "mysql:8.0",
    Username:             "root",
    Password:             "rootpassword",
    Database:             "testdb",
    HealthCheckTimeout:   30 * time.Second,
    HealthCheckRetries:   3,
    HealthCheckInterval:  500 * time.Millisecond,
    MaxOpenConns:         10,
    MaxIdleConns:         5,
    ConnMaxLifetime:      time.Hour,
}
```

### Available Options

```go
mysql.WithImage("mysql:8.4")                           // Custom image
mysql.WithUsername("myuser")                            // Custom username
mysql.WithPassword("mypass")                            // Custom password
mysql.WithDatabase("mydb")                              // Custom database name
mysql.WithHealthCheckTimeout(60 * time.Second)          // Longer startup timeout
mysql.WithHealthCheckRetry(5)                           // More retry attempts
mysql.WithHealthCheckInterval(2 * time.Second)          // Longer between retries
mysql.WithPoolSettings(20, 10, 30*time.Minute)         // Custom pool: maxOpen, maxIdle, maxLifetime
```

### Full Custom Example

```go
db, cleanup, err := mysql.NewWithOptions(ctx,
    mysql.WithImage("mysql:8.4"),
    mysql.WithUsername("app_user"),
    mysql.WithPassword("secure_password"),
    mysql.WithDatabase("myapp_test"),
    mysql.WithHealthCheckTimeout(60*time.Second),
    mysql.WithHealthCheckRetry(5),
    mysql.WithHealthCheckInterval(2*time.Second),
    mysql.WithPoolSettings(20, 10, 30*time.Minute),
)
```

---

## PostgreSQL Configuration

### Defaults

```go
config.PostgresConfig{
    Image:                "postgres:16-alpine",
    Username:             "postgres",
    Password:             "postgres",
    Database:             "testdb",
    HealthCheckTimeout:   30 * time.Second,
    HealthCheckRetries:   3,
    HealthCheckInterval:  500 * time.Millisecond,
    MaxOpenConns:         10,
    MaxIdleConns:         5,
    ConnMaxLifetime:      time.Hour,
}
```

### Available Options

```go
postgres.WithImage("postgres:17")                       // Custom image
postgres.WithUsername("admin")                           // Custom username
postgres.WithPassword("secret")                          // Custom password
postgres.WithDatabase("myapp")                           // Custom database
postgres.WithHealthCheckTimeout(45 * time.Second)        // Custom timeout
postgres.WithHealthCheckRetry(4)                         // Custom retries
postgres.WithHealthCheckInterval(time.Second)            // Custom interval
postgres.WithPoolSettings(15, 8, 2*time.Hour)           // Custom pool settings
```

---

## SQLite Configuration

### Defaults

```go
config.SQLiteConfig{
    Mode:  "memory",  // or "temp", "file"
    Path:  "",        // for "file" mode
    Cache: "",        // "shared", "private", "write"
}
```

### Available Options

```go
sqlite.WithMode("memory")     // In-memory database (fastest)
sqlite.WithMode("temp")       // Temporary file (auto-cleaned)
sqlite.WithMode("file")       // Persistent file
sqlite.WithPath("/tmp/test.db") // File path (for "file" mode)
sqlite.WithCache("shared")    // Shared cache mode
```

### SQLite Modes

| Mode | Description | Use Case |
|------|-------------|----------|
| `memory` | In-memory database | Unit tests, fastest |
| `temp` | Temporary file, auto-deleted | Need file-backed but auto-cleanup |
| `file` | Persistent file | Integration tests, debugging |

---

## Migration Configuration

```go
runner := migrate.NewRunner(db,
    migrate.WithDir("migrations"),          // Migration files directory
    migrate.WithTableName("schema_migrations"), // Tracking table name
    migrate.WithTimeout(30*time.Second),    // Per-migration timeout
)
```

---

## Connection Pool Configuration

All database types support connection pool tuning:

| Parameter | Description | Default | Recommendation |
|-----------|-------------|---------|---------------|
| `MaxOpenConns` | Max open connections | 10 | 20-50 for high concurrency |
| `MaxIdleConns` | Max idle connections | 5 | Half of MaxOpenConns |
| `ConnMaxLifetime` | Max connection age | 1h | 30m-2h |
| `ConnMaxIdleTime` | Max idle time | 30m | 5m-30m |

### Pool Monitoring

```go
import "github.com/atop0914/containerdb-bootcamp/internal/pool"

stats := pool.GetStats(db)
fmt.Printf("Open: %d, InUse: %d, Idle: %d\n",
    stats.OpenConnections, stats.InUse, stats.Idle)
```

---

## Health Check Configuration

Health checks verify the database is ready to accept connections.

| Parameter | Description | Default |
|-----------|-------------|---------|
| `Timeout` | Max total wait time | 30s |
| `Interval` | Time between retries | 500ms |
| `Retries` | Number of attempts | 3-5 |

### Custom Health Check

```go
import "github.com/atop0914/containerdb-bootcamp/internal/health"

result, err := health.Check(ctx, db,
    health.WithTimeout(60*time.Second),
    health.WithRetries(10),
    health.WithInterval(2*time.Second),
)
if result.Healthy {
    fmt.Printf("Database ready (latency: %v)\n", result.Latency)
}
```
