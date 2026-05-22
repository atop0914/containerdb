package benchmarks

import (
	"context"
	"testing"
	"time"

	"github.com/atop0914/containerdb/internal/health"
	"github.com/atop0914/containerdb/pkg/sqlite"
)

// BenchmarkHealth_DefaultConfig measures health config creation.
func BenchmarkHealth_DefaultConfig(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = health.DefaultConfig()
	}
}

// BenchmarkHealth_NewConfig measures health config with options.
func BenchmarkHealth_NewConfig(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = health.NewConfig(
			health.WithTimeout(10*time.Second),
			health.WithInterval(100*time.Millisecond),
			health.WithRetries(3),
		)
	}
}

// BenchmarkHealth_Check measures health check latency against SQLite.
func BenchmarkHealth_Check(b *testing.B) {
	db, cleanup, err := sqlite.InMemory()
	if err != nil {
		b.Fatal(err)
	}
	defer cleanup()

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := health.Check(ctx, db,
			health.WithTimeout(5*time.Second),
			health.WithRetries(1),
		)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkHealth_WaitForReady measures WaitForReady latency.
func BenchmarkHealth_WaitForReady(b *testing.B) {
	db, cleanup, err := sqlite.InMemory()
	if err != nil {
		b.Fatal(err)
	}
	defer cleanup()

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := health.WaitForReady(ctx, db,
			health.WithTimeout(5*time.Second),
			health.WithRetries(1),
		)
		if err != nil {
			b.Fatal(err)
		}
	}
}
