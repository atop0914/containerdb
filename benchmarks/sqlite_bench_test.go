// Package benchmarks provides performance benchmarks for ContainerDB.
package benchmarks

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/atop0914/containerdb-bootcamp/pkg/sqlite"
)

// BenchmarkSQLite_InMemory measures in-memory SQLite creation and teardown.
func BenchmarkSQLite_InMemory(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, cleanup, err := sqlite.InMemory()
		if err != nil {
			b.Fatal(err)
		}
		cleanup()
	}
}

// BenchmarkSQLite_TempDB measures temp file SQLite creation and teardown.
func BenchmarkSQLite_TempDB(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, cleanup, err := sqlite.TempDB()
		if err != nil {
			b.Fatal(err)
		}
		cleanup()
	}
}

// BenchmarkSQLite_InMemory_WithCache measures in-memory SQLite with shared cache.
func BenchmarkSQLite_InMemory_WithCache(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, cleanup, err := sqlite.NewWithOptions(
			sqlite.WithMode("memory"),
			sqlite.WithCache("shared"),
		)
		if err != nil {
			b.Fatal(err)
		}
		cleanup()
	}
}

// BenchmarkSQLite_SimpleInsert measures single row insert performance.
func BenchmarkSQLite_SimpleInsert(b *testing.B) {
	db, cleanup, err := sqlite.InMemory()
	if err != nil {
		b.Fatal(err)
	}
	defer cleanup()

	_, err = db.Exec("CREATE TABLE bench (id INTEGER PRIMARY KEY, value TEXT)")
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := db.Exec("INSERT INTO bench (value) VALUES (?)", fmt.Sprintf("value-%d", i))
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkSQLite_SimpleSelect measures single row select performance.
func BenchmarkSQLite_SimpleSelect(b *testing.B) {
	db, cleanup, err := sqlite.InMemory()
	if err != nil {
		b.Fatal(err)
	}
	defer cleanup()

	_, err = db.Exec("CREATE TABLE bench (id INTEGER PRIMARY KEY, value TEXT)")
	if err != nil {
		b.Fatal(err)
	}

	// Pre-populate data
	for i := 0; i < 1000; i++ {
		_, err = db.Exec("INSERT INTO bench (value) VALUES (?)", fmt.Sprintf("value-%d", i))
		if err != nil {
			b.Fatal(err)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var value string
		err := db.QueryRow("SELECT value FROM bench WHERE id = ?", (i%1000)+1).Scan(&value)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkSQLite_BatchInsert measures batch insert performance.
func BenchmarkSQLite_BatchInsert(b *testing.B) {
	db, cleanup, err := sqlite.InMemory()
	if err != nil {
		b.Fatal(err)
	}
	defer cleanup()

	_, err = db.Exec("CREATE TABLE bench (id INTEGER PRIMARY KEY, value TEXT)")
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tx, err := db.Begin()
		if err != nil {
			b.Fatal(err)
		}
		for j := 0; j < 100; j++ {
			_, err := tx.Exec("INSERT INTO bench (value) VALUES (?)", fmt.Sprintf("value-%d-%d", i, j))
			if err != nil {
				tx.Rollback()
				b.Fatal(err)
			}
		}
		if err := tx.Commit(); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkSQLite_RangeSelect measures range select performance.
func BenchmarkSQLite_RangeSelect(b *testing.B) {
	db, cleanup, err := sqlite.InMemory()
	if err != nil {
		b.Fatal(err)
	}
	defer cleanup()

	_, err = db.Exec("CREATE TABLE bench (id INTEGER PRIMARY KEY, value TEXT)")
	if err != nil {
		b.Fatal(err)
	}

	// Pre-populate with 10000 rows
	tx, _ := db.Begin()
	for i := 0; i < 10000; i++ {
		tx.Exec("INSERT INTO bench (value) VALUES (?)", fmt.Sprintf("value-%d", i))
	}
	tx.Commit()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rows, err := db.Query("SELECT id, value FROM bench WHERE id BETWEEN ? AND ?", 100, 200)
		if err != nil {
			b.Fatal(err)
		}
		for rows.Next() {
			var id int
			var value string
			rows.Scan(&id, &value)
		}
		rows.Close()
	}
}

// BenchmarkSQLite_Ping measures database ping latency.
func BenchmarkSQLite_Ping(b *testing.B) {
	db, cleanup, err := sqlite.InMemory()
	if err != nil {
		b.Fatal(err)
	}
	defer cleanup()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := db.Ping(); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkSQLite_ParallelReads measures concurrent read performance.
func BenchmarkSQLite_ParallelReads(b *testing.B) {
	db, cleanup, err := sqlite.TempDB()
	if err != nil {
		b.Fatal(err)
	}
	defer cleanup()

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS bench (id INTEGER PRIMARY KEY, value TEXT)")
	if err != nil {
		b.Fatal(err)
	}

	for i := 0; i < 1000; i++ {
		_, err = db.Exec("INSERT INTO bench (value) VALUES (?)", fmt.Sprintf("value-%d", i))
		if err != nil {
			b.Fatal(err)
		}
	}

	db.SetMaxOpenConns(10)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			var value string
			err := db.QueryRow("SELECT value FROM bench WHERE id = ?", (i%1000)+1).Scan(&value)
			if err != nil && err != sql.ErrNoRows {
				b.Fatal(err)
			}
			i++
		}
	})
}
