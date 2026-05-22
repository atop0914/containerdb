package benchmarks

import (
	"testing"

	"github.com/atop0914/containerdb/internal/config"
)

// BenchmarkDefaultMySQLConfig measures MySQL config creation.
func BenchmarkDefaultMySQLConfig(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = config.DefaultMySQLConfig()
	}
}

// BenchmarkDefaultPostgresConfig measures PostgreSQL config creation.
func BenchmarkDefaultPostgresConfig(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = config.DefaultPostgresConfig()
	}
}

// BenchmarkDefaultSQLiteConfig measures SQLite config creation.
func BenchmarkDefaultSQLiteConfig(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = config.DefaultSQLiteConfig()
	}
}
