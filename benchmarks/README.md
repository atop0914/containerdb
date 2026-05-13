# ContainerDB Benchmarks

Performance benchmarks for all ContainerDB components.

## Running Benchmarks

### Run all benchmarks
```bash
go test ./benchmarks/... -bench=. -benchtime=1s -count=3
```

### Run specific benchmark groups
```bash
# SQLite benchmarks
go test ./benchmarks/... -bench=BenchmarkSQLite -benchtime=100ms

# Config benchmarks
go test ./benchmarks/... -bench=BenchmarkDefault -benchtime=1s

# Pool benchmarks
go test ./benchmarks/... -bench=BenchmarkPool -benchtime=1s

# Health check benchmarks
go test ./benchmarks/... -bench=BenchmarkHealth -benchtime=1s

# Compose benchmarks
go test ./benchmarks/... -bench=BenchmarkCompose -benchtime=1s
```

### Generate benchmark comparison
```bash
# Save baseline
go test ./benchmarks/... -bench=. -benchtime=100ms -count=5 > old.txt

# After changes
go test ./benchmarks/... -bench=. -benchtime=100ms -count=5 > new.txt

# Compare
benchstat old.txt new.txt
```

### Memory allocation profiling
```bash
go test ./benchmarks/... -bench=BenchmarkSQLite_SimpleInsert -benchmem
```

## Benchmark Categories

### SQLite Benchmarks (`sqlite_bench_test.go`)
- **InMemory**: In-memory database creation/teardown
- **TempDB**: Temp file database creation/teardown
- **SimpleInsert**: Single row insert performance
- **SimpleSelect**: Single row select performance
- **BatchInsert**: Transaction-based batch insert (100 rows)
- **RangeSelect**: Range query with 100 rows
- **Ping**: Database ping latency
- **ParallelReads**: Concurrent read performance with temp file DB

### Config Benchmarks (`config_bench_test.go`)
- **DefaultMySQLConfig**: MySQL config creation
- **DefaultPostgresConfig**: PostgreSQL config creation
- **DefaultSQLiteConfig**: SQLite config creation

### Pool Benchmarks (`pool_bench_test.go`)
- **DefaultPoolConfig**: Pool config creation
- **PoolConfig_Validate**: Config validation
- **Pool_Configure**: Apply pool settings to DB
- **Pool_GetStats**: Retrieve pool statistics
- **Pool_TracedDB_ExecContext**: Traced exec overhead
- **Pool_TracedDB_QueryContext**: Traced query overhead

### Health Benchmarks (`health_bench_test.go`)
- **Health_DefaultConfig**: Health config creation
- **Health_NewConfig**: Health config with options
- **Health_Check**: Health check against SQLite
- **Health_WaitForReady**: WaitForReady latency

### Compose Benchmarks (`compose_bench_test.go`)
- **Compose_GenerateMySQLCompose**: Internal MySQL service generation
- **Compose_GeneratePostgresCompose**: Internal PostgreSQL service generation
- **Compose_GenerateMySQLService**: Public MySQL service generation
- **Compose_GeneratePostgresService**: Public PostgreSQL service generation
- **Compose_GenerateMySQLServiceWithHealthCheck**: MySQL with healthcheck
- **Compose_GeneratePostgresServiceWithHealthCheck**: PostgreSQL with healthcheck
- **Compose_BuildComposeFile**: Compose file building
- **Compose_NewRunner**: Runner creation
- **Compose_GenerateFile**: Full compose file generation
- **Compose_TemplateMySQL**: MySQL template generation
- **Compose_TemplatePostgres**: PostgreSQL template generation
- **Compose_TemplateMySQLPostgres**: Combined template generation

## Expected Performance (Intel Xeon Gold 6148)

| Benchmark | ns/op | allocs/op |
|-----------|-------|-----------|
| SQLite InMemory | ~80,000 | ~10 |
| SQLite TempDB | ~160,000 | ~15 |
| SQLite SimpleInsert | ~6,500 | ~5 |
| SQLite SimpleSelect | ~7,500 | ~5 |
| SQLite Ping | ~200 | ~0 |
| Config Default | ~0.4 | ~0 |
| Pool Configure | ~60 | ~0 |
| Health Check | ~1,500 | ~5 |
| Compose Generate | ~35,000 | ~50 |
