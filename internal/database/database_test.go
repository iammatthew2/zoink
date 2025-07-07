package database

import (
	"path/filepath"
	"testing"
	"time"
)

func TestDatabaseBasicOperations(t *testing.T) {
	// Create temporary database
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	config := DatabaseConfig{Path: dbPath}
	db, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Test adding visits
	testPaths := []string{
		"/home/user/projects",
		"/home/user/documents",
		"/home/user/projects/my-app",
	}

	for _, path := range testPaths {
		if err := db.AddVisit(path); err != nil {
			t.Errorf("Failed to add visit for %s: %v", path, err)
		}
	}

	// Add multiple visits to projects to increase frequency
	for i := 0; i < 5; i++ {
		if err := db.AddVisit("/home/user/projects"); err != nil {
			t.Errorf("Failed to add repeat visit: %v", err)
		}
	}

	// Test querying
	results, err := db.Query("proj", 10)
	if err != nil {
		t.Fatalf("Failed to query database: %v", err)
	}

	if len(results) == 0 {
		t.Error("Expected results for 'proj' query")
	}

	// Projects should be first due to higher frequency
	if len(results) > 0 && results[0].Path != "/home/user/projects" {
		t.Errorf("Expected /home/user/projects to be first result, got %s", results[0].Path)
	}

	// Test persistence
	if err := db.Save(); err != nil {
		t.Fatalf("Failed to save database: %v", err)
	}

	// Create new database instance and verify data persisted
	db2, err := New(config)
	if err != nil {
		t.Fatalf("Failed to reload database: %v", err)
	}
	defer db2.Close()

	results2, err := db2.Query("proj", 10)
	if err != nil {
		t.Fatalf("Failed to query reloaded database: %v", err)
	}

	if len(results2) != len(results) {
		t.Errorf("Expected %d results after reload, got %d", len(results), len(results2))
	}
}

func TestDatabaseCleanup(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	config := DatabaseConfig{Path: dbPath}
	db, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Add visits to existing and non-existing directories
	existingDir := tempDir // This exists
	nonExistingDir := "/this/path/does/not/exist"

	if err := db.AddVisit(existingDir); err != nil {
		t.Errorf("Failed to add visit for existing dir: %v", err)
	}
	if err := db.AddVisit(nonExistingDir); err != nil {
		t.Errorf("Failed to add visit for non-existing dir: %v", err)
	}

	// Test cleanup
	removed, err := db.CleanupMissing()
	if err != nil {
		t.Fatalf("Failed to cleanup missing directories: %v", err)
	}

	if removed != 1 {
		t.Errorf("Expected to remove 1 directory, removed %d", removed)
	}

	// Verify non-existing directory was removed
	results, err := db.Query("not/exist", 10)
	if err != nil {
		t.Fatalf("Failed to query after cleanup: %v", err)
	}

	if len(results) > 0 {
		t.Error("Expected no results for non-existing directory after cleanup")
	}
}

func TestFrecencyCalculation(t *testing.T) {
	now := time.Now().Unix()

	tests := []struct {
		name     string
		entry    *DirectoryEntry
		expected float64 // Approximate expected value
	}{
		{
			name: "High frequency, recent",
			entry: &DirectoryEntry{
				VisitCount:  10,
				LastVisited: now, // Now
			},
			expected: 10.0, // Full score
		},
		{
			name: "Low frequency, recent",
			entry: &DirectoryEntry{
				VisitCount:  2,
				LastVisited: now,
			},
			expected: 2.0,
		},
		{
			name: "High frequency, old",
			entry: &DirectoryEntry{
				VisitCount:  10,
				LastVisited: now - (60 * 24 * 60 * 60), // 60 days ago
			},
			expected: 2.5, // Should be much lower due to age
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := calculateFrecency(tt.entry)

			// Allow some tolerance in the comparison
			if score < tt.expected*0.8 || score > tt.expected*1.2 {
				t.Errorf("Frecency score %f not close to expected %f", score, tt.expected)
			}
		})
	}
}
