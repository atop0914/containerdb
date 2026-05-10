# CLI Reference

The `containerdb` CLI tool provides a command-line interface for managing database containers.

## Installation

```bash
go install github.com/atop0914/containerdb-bootcamp/cmd/containerdb@latest
```

Or build from source:

```bash
git clone https://github.com/atop0914/containerdb-bootcamp.git
cd containerdb-bootcamp
go build -o containerdb ./cmd/containerdb/
```

## Commands

### `containerdb start`

Start a database container.

```bash
containerdb start [flags]
```

**Flags:**

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--type` | `-t` | Database type: `mysql`, `postgres` | `mysql` |
| `--image` | `-i` | Docker image to use | (type default) |
| `--username` | `-u` | Database username | (type default) |
| `--password` | `-p` | Database password | (type default) |
| `--database` | `-d` | Database name | `testdb` |
| `--timeout` | | Container startup timeout | `30s` |

**Examples:**

```bash
# Start MySQL with defaults
containerdb start

# Start PostgreSQL
containerdb start -t postgres

# Start specific MySQL version with custom credentials
containerdb start -t mysql -i mysql:8.4 -u myuser -p mypass -d mydb
```

### `containerdb stop`

Stop a running container.

```bash
containerdb stop [flags]
```

**Flags:**

| Flag | Short | Description |
|------|-------|-------------|
| `--name` | `-n` | Container name to stop |

> **Note**: Containers started via the CLI are ephemeral. For persistent management, use Docker Compose integration or the library directly.

### `containerdb status`

Show status of running containers.

```bash
containerdb status [flags]
```

**Flags:**

| Flag | Short | Description |
|------|-------|-------------|
| `--all` | `-a` | Show all containers (including stopped) |

### `containerdb compose`

Docker Compose integration commands.

```bash
containerdb compose [command]
```

#### `containerdb compose init`

Generate a `docker-compose.yml` file.

```bash
containerdb compose init [flags]
```

**Flags:**

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--type` | `-t` | Database type: `mysql`, `postgres`, `all` | `all` |
| `--output` | `-o` | Output file path | `docker-compose.yml` |
| `--with-healthcheck` | | Include health check config | `false` |

**Examples:**

```bash
# Generate compose file with all databases
containerdb compose init

# Generate for PostgreSQL only
containerdb compose init -t postgres

# Generate with health checks
containerdb compose init --with-healthcheck

# Custom output path
containerdb compose init -o infra/docker-compose.yml
```

#### `containerdb compose up`

Start services defined in docker-compose.yml.

```bash
containerdb compose up [flags]
```

**Flags:**

| Flag | Description |
|------|-------------|
| `--wait` | Wait for health checks to pass before returning |
| `--detach` | Run in detached mode (default: true) |

**Examples:**

```bash
# Start services
containerdb compose up

# Start and wait for readiness
containerdb compose up --wait
```

#### `containerdb compose down`

Stop and remove services.

```bash
containerdb compose down [flags]
```

**Flags:**

| Flag | Short | Description |
|------|-------|-------------|
| `--volumes` | `-v` | Remove named volumes (WARNING: deletes data) |

**Examples:**

```bash
# Stop services
containerdb compose down

# Stop and remove volumes
containerdb compose down -v
```

#### `containerdb compose status`

Show service status.

```bash
containerdb compose status
```

#### `containerdb compose logs`

Show service logs.

```bash
containerdb compose logs [flags]
```

**Flags:**

| Flag | Short | Description |
|------|-------|-------------|
| `--service` | `-s` | Show logs for specific service |
| `--follow` | `-f` | Follow log output |

## Global Flags

| Flag | Description |
|------|-------------|
| `--help` | Show help for any command |
| `--version` | Show version information |

## Environment Variables

| Variable | Description |
|----------|-------------|
| `CONTAINERDB_DOCKER_HOST` | Custom Docker host |
| `CONTAINERDB_LOG_LEVEL` | Log level: debug, info, warn, error |
