package benchmarks

import (
	"testing"
	"time"

	"github.com/atop0914/containerdb-bootcamp/internal/pool"
	"github.com/atop0914/containerdb-bootcamp/pkg/sqlite"
)

// BenchmarkDefaultPoolConfig measures default pool config creation.
func BenchmarkDefaultPoolConfig(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = pool.DefaultPoolConfig()
	}
}

// BenchmarkPoolConfig_Validate measures pool config validation.
func BenchmarkPoolConfig_Validate(b *testing.B) {
	cfg := pool.DefaultPoolConfig()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cfg.Validate()
	}
}

// BenchmarkPool_Configure measures pool configuration application.
func BenchmarkPool_Configure(b *testing.B) {
	db, cleanup, err := sqlite.InMemory()
	if err != nil {
		b.Fatal(err)
	}
	defer cleanup()

	cfg := &pool.PoolConfig{
		MaxOpenConns:    20,
		MaxIdleConns:    10,
		ConnMaxLifetime: time.Hour,
		ConnMaxIdleTime: 30 * time.Minute,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = pool.Configure(db, cfg)
	}
}

// BenchmarkPool_GetStats measures pool statistics retrieval.
func BenchmarkPool_GetStats(b *testing.B) {
	db, cleanup, err := sqlite.InMemory()
	if err != nil {
		b.Fatal(err)
	}
	defer cleanup()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = pool.GetStats(db)
	}
}

// BenchmarkPool_TracedDB_ExecContext measures traced exec overhead.
func BenchmarkPool_TracedDB_ExecContext(b *testing.B) {
	db, cleanup, err := sqlite.InMemory()
	if err != nil {
		b.Fatal(err)
	}
	defer cleanup()

	_, err = db.Exec("CREATE TABLE bench (id INTEGER PRIMARY KEY, value TEXT)")
	if err != nil {
		b.Fatal(err)
	}

	traced := pool.NewTracedDB(db)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := traced.ExecContext(nil, "INSERT INTO bench (value) VALUES (?)", "test")
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkPool_TracedDB_QueryContext measures traced query overhead.
func BenchmarkPool_TracedDB_QueryContext(b *testing.B) {
	db, cleanup, err := sqlite.InMemory()
	if err != nil {
		b.Fatal(err)
	}
	defer cleanup()

	_, err = db.Exec("CREATE TABLE bench (id INTEGER PRIMARY KEY, value TEXT)")
	if err != nil {
		b.Fatal(err)
	}

	for i := 0; i < 100; i++ {
		db.Exec("INSERT INTO bench (value) VALUES (?)", "test")
	}

	traced := pool.NewTracedDB(db)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rows, err := traced.QueryContext(nil, "SELECT id, value FROM bench LIMIT 10")
		if err != nil {
			b.Fatal(err)
		}
		rows.Close()
	}
}
