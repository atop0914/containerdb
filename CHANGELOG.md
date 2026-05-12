# Changelog

All notable changes to ContainerDB will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2026-05-12

### Added
- MySQL container management with testcontainers-go (`pkg/mysql`)
- PostgreSQL container management with testcontainers-go (`pkg/postgres`)
- SQLite helpers for in-memory and temp file databases (`pkg/sqlite`)
- Functional options pattern for all database configurations
- Connection pool configuration and monitoring (`internal/pool`)
- TracedDB wrapper for slow query logging
- Health check utilities with retry logic (`internal/health`)
- Database migration runner with rollback support (`internal/migrate`, `pkg/migrate`)
- Docker Compose file generation and management (`internal/compose`, `pkg/compose`)
- CLI tool with start/stop/status/compose/version commands (`cmd/containerdb`)
- Comprehensive documentation suite (`docs/`)
- Performance benchmarks for SQLite, config, pool, health, and compose packages
- Version information package with build-time ldflags support

### Fixed
- QueryRowContext in TracedDB now properly logs slow queries instead of spawning leaked goroutines
- Configure() in pool package now unconditionally sets ConnMaxIdleTime (removed dead code path)
- Consistent panic error format across MySQL and PostgreSQL MustNew functions

### Removed
- No-op `sqlite.Pool()` function (was identity function, provided no value)
