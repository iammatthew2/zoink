package database

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

// DirectoryEntry represents a single directory with frecency data
type DirectoryEntry struct {
	Path         string
	VisitCount   uint32
	LastVisited  int64 // Unix timestamp
	FirstVisited int64 // Unix timestamp
}

// Database manages the binary database of directory entries
type Database struct {
	path    string
	entries map[string]*DirectoryEntry
	mutex   sync.RWMutex
}

// DatabaseConfig holds configuration for the database
type DatabaseConfig struct {
	Path string
}

// New creates a new database instance
func New(config DatabaseConfig) (*Database, error) {
	db := &Database{
		path:    config.Path,
		entries: make(map[string]*DirectoryEntry),
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(config.Path), 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	// Load existing data
	if err := db.load(); err != nil {
		return nil, fmt.Errorf("failed to load database: %w", err)
	}

	return db, nil
}

// AddVisit records a visit to a directory
func (db *Database) AddVisit(path string) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	// Clean and normalize path
	cleanPath := filepath.Clean(path)
	now := time.Now().Unix()

	entry, exists := db.entries[cleanPath]
	if exists {
		entry.VisitCount++
		entry.LastVisited = now
	} else {
		db.entries[cleanPath] = &DirectoryEntry{
			Path:         cleanPath,
			VisitCount:   1,
			LastVisited:  now,
			FirstVisited: now,
		}
	}

	return nil
}

// Query searches for directories matching the given query
func (db *Database) Query(query string, maxResults int) ([]*DirectoryEntry, error) {
	db.mutex.RLock()
	defer db.mutex.RUnlock()

	var matches []*DirectoryEntry

	// Simple substring matching for now (will add fuzzy matching later)
	for _, entry := range db.entries {
		if contains(entry.Path, query) {
			matches = append(matches, entry)
		}
	}

	// Sort by frecency score (frequency + recency)
	sort.Slice(matches, func(i, j int) bool {
		return calculateFrecency(matches[i]) > calculateFrecency(matches[j])
	})

	// Limit results
	if len(matches) > maxResults {
		matches = matches[:maxResults]
	}

	return matches, nil
}

// GetAll returns all directory entries
func (db *Database) GetAll() ([]*DirectoryEntry, error) {
	db.mutex.RLock()
	defer db.mutex.RUnlock()

	entries := make([]*DirectoryEntry, 0, len(db.entries))
	for _, entry := range db.entries {
		entries = append(entries, entry)
	}

	return entries, nil
}

// RemoveDirectory removes a directory from the database
func (db *Database) RemoveDirectory(path string) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	cleanPath := filepath.Clean(path)
	delete(db.entries, cleanPath)

	return nil
}

// CleanupMissing removes directories that no longer exist
func (db *Database) CleanupMissing() (int, error) {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	removed := 0
	for path, _ := range db.entries {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			delete(db.entries, path)
			removed++
		}
	}

	return removed, nil
}

// Save persists the database to disk
func (db *Database) Save() error {
	db.mutex.RLock()
	defer db.mutex.RUnlock()

	return db.save()
}

// Close saves the database and cleans up resources
func (db *Database) Close() error {
	return db.Save()
}

// save writes the database to disk (caller must hold lock)
func (db *Database) save() error {
	// Write to temporary file first for atomic operation
	tempPath := db.path + ".tmp"
	file, err := os.Create(tempPath)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer file.Close()

	// Write magic header and version
	if err := binary.Write(file, binary.LittleEndian, uint32(0x5A4F494E)); err != nil { // "ZOIN"
		return fmt.Errorf("failed to write magic: %w", err)
	}
	if err := binary.Write(file, binary.LittleEndian, uint32(1)); err != nil { // Version 1
		return fmt.Errorf("failed to write version: %w", err)
	}

	// Write number of entries
	if err := binary.Write(file, binary.LittleEndian, uint32(len(db.entries))); err != nil {
		return fmt.Errorf("failed to write entry count: %w", err)
	}

	// Write each entry
	for _, entry := range db.entries {
		if err := writeEntry(file, entry); err != nil {
			return fmt.Errorf("failed to write entry: %w", err)
		}
	}

	file.Close()

	// Atomic replace
	if err := os.Rename(tempPath, db.path); err != nil {
		os.Remove(tempPath) // Cleanup on failure
		return fmt.Errorf("failed to replace database file: %w", err)
	}

	return nil
}

// load reads the database from disk
func (db *Database) load() error {
	file, err := os.Open(db.path)
	if err != nil {
		if os.IsNotExist(err) {
			// New database, nothing to load
			return nil
		}
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer file.Close()

	// Read and verify magic header
	var magic uint32
	if err := binary.Read(file, binary.LittleEndian, &magic); err != nil {
		return fmt.Errorf("failed to read magic: %w", err)
	}
	if magic != 0x5A4F494E { // "ZOIN"
		return fmt.Errorf("invalid database format")
	}

	// Read version
	var version uint32
	if err := binary.Read(file, binary.LittleEndian, &version); err != nil {
		return fmt.Errorf("failed to read version: %w", err)
	}
	if version != 1 {
		return fmt.Errorf("unsupported database version: %d", version)
	}

	// Read number of entries
	var entryCount uint32
	if err := binary.Read(file, binary.LittleEndian, &entryCount); err != nil {
		return fmt.Errorf("failed to read entry count: %w", err)
	}

	// Read entries
	db.entries = make(map[string]*DirectoryEntry, entryCount)
	for i := uint32(0); i < entryCount; i++ {
		entry, err := readEntry(file)
		if err != nil {
			return fmt.Errorf("failed to read entry %d: %w", i, err)
		}
		db.entries[entry.Path] = entry
	}

	return nil
}

// writeEntry writes a single entry to the file
func writeEntry(w io.Writer, entry *DirectoryEntry) error {
	// Write path length and path
	pathBytes := []byte(entry.Path)
	if err := binary.Write(w, binary.LittleEndian, uint32(len(pathBytes))); err != nil {
		return err
	}
	if _, err := w.Write(pathBytes); err != nil {
		return err
	}

	// Write numeric fields
	if err := binary.Write(w, binary.LittleEndian, entry.VisitCount); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, entry.LastVisited); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, entry.FirstVisited); err != nil {
		return err
	}

	return nil
}

// readEntry reads a single entry from the file
func readEntry(r io.Reader) (*DirectoryEntry, error) {
	// Read path length
	var pathLen uint32
	if err := binary.Read(r, binary.LittleEndian, &pathLen); err != nil {
		return nil, err
	}

	// Read path
	pathBytes := make([]byte, pathLen)
	if _, err := io.ReadFull(r, pathBytes); err != nil {
		return nil, err
	}

	entry := &DirectoryEntry{
		Path: string(pathBytes),
	}

	// Read numeric fields
	if err := binary.Read(r, binary.LittleEndian, &entry.VisitCount); err != nil {
		return nil, err
	}
	if err := binary.Read(r, binary.LittleEndian, &entry.LastVisited); err != nil {
		return nil, err
	}
	if err := binary.Read(r, binary.LittleEndian, &entry.FirstVisited); err != nil {
		return nil, err
	}

	return entry, nil
}

// calculateFrecency computes the frecency score for an entry
func calculateFrecency(entry *DirectoryEntry) float64 {
	// Simple frecency algorithm:
	// Score = frequency * recency_factor
	// Recency factor decreases exponentially with age

	now := time.Now().Unix()
	age := float64(now - entry.LastVisited)

	// Convert age from seconds to days
	ageInDays := age / (24 * 60 * 60)

	// Exponential decay: score halves every 30 days
	recencyFactor := 1.0
	if ageInDays > 0 {
		halfLife := 30.0 // days
		// Correct exponential decay formula: factor = 0.5^(age/halfLife)
		recencyFactor = 1.0
		for i := 0.0; i < ageInDays/halfLife; i += 1.0 {
			recencyFactor *= 0.5
		}
		if recencyFactor < 0.01 {
			recencyFactor = 0.01 // Minimum factor
		}
	}

	return float64(entry.VisitCount) * recencyFactor
}

// contains performs case-insensitive substring search
func contains(path, query string) bool {
	// Simple implementation - will be replaced with fuzzy matching
	pathLower := filepath.Base(path) // Just check basename for now
	queryLower := query

	// Convert to lowercase for case-insensitive search
	for i := 0; i < len(pathLower); i++ {
		if pathLower[i] >= 'A' && pathLower[i] <= 'Z' {
			pathLower = pathLower[:i] + string(pathLower[i]+32) + pathLower[i+1:]
		}
	}
	for i := 0; i < len(queryLower); i++ {
		if queryLower[i] >= 'A' && queryLower[i] <= 'Z' {
			queryLower = queryLower[:i] + string(queryLower[i]+32) + queryLower[i+1:]
		}
	}

	return len(query) == 0 || stringContains(pathLower, queryLower)
}

// stringContains checks if s contains substr
func stringContains(s, substr string) bool {
	if len(substr) > len(s) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			if s[i+j] != substr[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}
