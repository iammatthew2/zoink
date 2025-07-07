package database

import (
	"fmt"
	"path/filepath"
	"testing"
)

func BenchmarkDatabaseOperations(b *testing.B) {
	tempDir := b.TempDir()
	dbPath := filepath.Join(tempDir, "bench.db")

	config := DatabaseConfig{Path: dbPath}
	db, err := New(config)
	if err != nil {
		b.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Pre-populate with realistic data
	for i := 0; i < 1000; i++ {
		path := fmt.Sprintf("/home/user/project%d", i)
		for j := 0; j < (i%10)+1; j++ { // Varying visit counts
			db.AddVisit(path)
		}
	}

	b.ResetTimer()

	b.Run("Query", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := db.Query("project", 10)
			if err != nil {
				b.Fatalf("Query failed: %v", err)
			}
		}
	})

	b.Run("AddVisit", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			path := fmt.Sprintf("/home/user/temp%d", i)
			err := db.AddVisit(path)
			if err != nil {
				b.Fatalf("AddVisit failed: %v", err)
			}
		}
	})

	b.Run("Save", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			err := db.Save()
			if err != nil {
				b.Fatalf("Save failed: %v", err)
			}
		}
	})
}

func BenchmarkDatabaseLoad(b *testing.B) {
	tempDir := b.TempDir()
	dbPath := filepath.Join(tempDir, "bench.db")

	// Create and populate database
	config := DatabaseConfig{Path: dbPath}
	db, err := New(config)
	if err != nil {
		b.Fatalf("Failed to create database: %v", err)
	}

	for i := 0; i < 1000; i++ {
		path := fmt.Sprintf("/home/user/project%d", i)
		db.AddVisit(path)
	}
	db.Save()
	db.Close()

	b.ResetTimer()

	// Benchmark loading
	for i := 0; i < b.N; i++ {
		db, err := New(config)
		if err != nil {
			b.Fatalf("Failed to load database: %v", err)
		}
		db.Close()
	}
}
