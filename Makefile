.PHONY: test bench bench-sqlite bench-pool bench-health bench-compose bench-config bench-mem clean

# Run all tests (skip Docker-dependent tests)
test:
	go test ./internal/... ./pkg/... -short -timeout=120s

# Run all benchmarks
bench:
	go test ./benchmarks/... -bench=. -benchtime=100ms -count=1 -timeout=300s

# Run specific benchmark groups
bench-sqlite:
	go test ./benchmarks/... -bench=BenchmarkSQLite -benchtime=100ms -count=1

bench-pool:
	go test ./benchmarks/... -bench=BenchmarkPool -benchtime=1s -count=1

bench-health:
	go test ./benchmarks/... -bench=BenchmarkHealth -benchtime=1s -count=1

bench-compose:
	go test ./benchmarks/... -bench=BenchmarkCompose -benchtime=1s -count=1

bench-config:
	go test ./benchmarks/... -bench=BenchmarkDefault -benchtime=1s -count=1

# Run benchmarks with memory allocation stats
bench-mem:
	go test ./benchmarks/... -bench=. -benchtime=100ms -benchmem -count=1

# Run benchmarks multiple times for statistical significance
bench-stat:
	go test ./benchmarks/... -bench=. -benchtime=100ms -count=5 > bench-results.txt
	@echo "Results saved to bench-results.txt"
	@echo "Use 'benchstat' to compare with previous runs"

# Build binaries for release
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT  := $(shell git rev-parse --short HEAD)
DATE    := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -s -w \
	-X github.com/atop0914/containerdb-bootcamp/internal/version.Version=$(VERSION) \
	-X github.com/atop0914/containerdb-bootcamp/internal/version.GitCommit=$(COMMIT) \
	-X github.com/atop0914/containerdb-bootcamp/internal/version.BuildDate=$(DATE)

build:
	go build -ldflags="$(LDFLAGS)" -o containerdb ./cmd/containerdb/

build-all:
	@mkdir -p dist
	@for pair in linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64; do \
		GOOS=$${pair%/*} GOARCH=$${pair#*/} EXT=""; \
		if [ "$$GOOS" = "windows" ]; then EXT=".exe"; fi; \
		OUT="dist/containerdb-$$GOOS-$$GOARCH$$EXT"; \
		echo "Building $$OUT..."; \
		GOOS=$$GOOS GOARCH=$$GOARCH go build -ldflags="$(LDFLAGS)" -o "$$OUT" ./cmd/containerdb/; \
	done
	@echo "Build complete: dist/"

# Run golangci-lint
lint:
	golangci-lint run ./...

# Clean generated files
clean:
	rm -f bench-results.txt
	rm -f /tmp/testdb-*.sqlite
	rm -rf dist/
