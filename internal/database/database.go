package database

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
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

// MatchResult represents a search result with both fuzzy and frecency scores
type MatchResult struct {
	Entry         *DirectoryEntry
	FuzzyScore    int
	FrecencyScore float64
	CombinedScore float64
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

// Query searches for directories matching the given query using fuzzy matching combined with frecency
func (db *Database) Query(query string, maxResults int) ([]*DirectoryEntry, error) {
	db.mutex.RLock()
	defer db.mutex.RUnlock()

	if query == "" {
		// No query - return all entries sorted by frecency
		var entries []*DirectoryEntry
		for _, entry := range db.entries {
			entries = append(entries, entry)
		}

		// Sort by frecency score only
		sort.Slice(entries, func(i, j int) bool {
			return calculateFrecency(entries[i]) > calculateFrecency(entries[j])
		})

		// Limit results
		if len(entries) > maxResults {
			entries = entries[:maxResults]
		}

		return entries, nil
	}

	var matches []MatchResult

	// Fuzzy match against all entries
	for _, entry := range db.entries {
		fuzzyScore := fuzzyMatch(entry.Path, query)
		if fuzzyScore > 0 {
			frecencyScore := calculateFrecency(entry)

			// Combine fuzzy and frecency scores
			// Normalize fuzzy score to 0-1 range (assuming max score around 1000)
			normalizedFuzzy := float64(fuzzyScore) / 1000.0
			if normalizedFuzzy > 1.0 {
				normalizedFuzzy = 1.0
			}

			// Combine with weights: 60% fuzzy matching, 40% frecency
			combinedScore := (normalizedFuzzy * 0.6) + (frecencyScore * 0.4)

			matches = append(matches, MatchResult{
				Entry:         entry,
				FuzzyScore:    fuzzyScore,
				FrecencyScore: frecencyScore,
				CombinedScore: combinedScore,
			})
		}
	}

	// Sort by combined score
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].CombinedScore > matches[j].CombinedScore
	})

	// Convert to DirectoryEntry slice
	var entries []*DirectoryEntry
	for _, match := range matches {
		entries = append(entries, match.Entry)
	}

	// Limit results
	if len(entries) > maxResults {
		entries = entries[:maxResults]
	}

	return entries, nil
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
	for path := range db.entries {
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
		// Use proper exponential decay: e^(-ln(2) * age / halfLife)
		decayRate := math.Log(2) / halfLife
		recencyFactor = math.Exp(-decayRate * ageInDays)
		if recencyFactor < 0.01 {
			recencyFactor = 0.01 // Minimum factor
		}
	}

	return float64(entry.VisitCount) * recencyFactor
}

// fuzzyMatch implements an fzf-inspired fuzzy matching algorithm
func fuzzyMatch(text, pattern string) int {
	if len(pattern) == 0 {
		return 0
	}

	// Use only the basename for matching (like most directory jumpers)
	text = filepath.Base(text)

	// Convert to lowercase for case-insensitive matching
	textLower := strings.ToLower(text)
	patternLower := strings.ToLower(pattern)

	// Check if we can match all pattern characters
	if !canMatch(textLower, patternLower) {
		return 0
	}

	// Calculate detailed score
	return calculateFuzzyScore(text, textLower, pattern, patternLower)
}

// canMatch checks if all characters in pattern exist in text in order
func canMatch(text, pattern string) bool {
	textIdx := 0
	for _, patternChar := range pattern {
		found := false
		for textIdx < len(text) {
			if rune(text[textIdx]) == patternChar {
				found = true
				textIdx++
				break
			}
			textIdx++
		}
		if !found {
			return false
		}
	}
	return true
}

// calculateFuzzyScore computes a detailed fuzzy match score
func calculateFuzzyScore(text, textLower, pattern, patternLower string) int {
	score := 0
	patternIdx := 0
	textIdx := 0
	consecutiveCount := 0

	// Bonus constants (similar to fzf)
	const (
		scoreMatch            = 16
		scoreCaseMatch        = 1
		scoreConsecutive      = 32
		scoreWordBoundary     = 8
		scoreFirstCharBonus   = 32
		penaltyLeading        = -2
		penaltyMaxLeading     = -12
		penaltyNonConsecutive = -1
	)

	// Track leading penalty
	leadingPenalty := 0

	for patternIdx < len(pattern) && textIdx < len(text) {
		patternChar := rune(patternLower[patternIdx])
		textChar := rune(textLower[textIdx])

		if patternChar == textChar {
			// Base match score
			currentScore := scoreMatch

			// Case match bonus
			if rune(pattern[patternIdx]) == rune(text[textIdx]) {
				currentScore += scoreCaseMatch
			}

			// First character bonus
			if patternIdx == 0 {
				currentScore += scoreFirstCharBonus
			}

			// Consecutive character bonus
			if consecutiveCount > 0 {
				currentScore += scoreConsecutive
			}
			consecutiveCount++

			// Word boundary bonus (after slash, dash, underscore, space, or at start)
			if textIdx == 0 || isWordBoundary(rune(text[textIdx-1])) {
				currentScore += scoreWordBoundary
			}

			score += currentScore
			patternIdx++
			leadingPenalty = 0 // Reset leading penalty after first match
		} else {
			// Apply leading penalty only before first match
			if patternIdx == 0 && leadingPenalty > penaltyMaxLeading {
				leadingPenalty += penaltyLeading
			}

			// Non-consecutive penalty
			if consecutiveCount > 0 {
				score += penaltyNonConsecutive
			}
			consecutiveCount = 0
		}

		textIdx++
	}

	// Ensure all pattern characters were matched
	if patternIdx < len(pattern) {
		return 0
	}

	// Apply leading penalty
	score += leadingPenalty

	// Bonus for shorter matches (prefer more specific matches)
	lengthBonus := int(float64(len(pattern)) / float64(len(text)) * 50)
	score += lengthBonus

	return score
}

// isWordBoundary checks if a character is a word boundary
func isWordBoundary(char rune) bool {
	return char == '/' || char == '-' || char == '_' || char == ' ' || char == '.'
}
