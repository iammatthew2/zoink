package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/iammatthew2/zoink/internal/database"
	"github.com/spf13/cobra"
)

// statsCmd represents the stats command
var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show usage statistics",
	Long:  `Display statistics about your directory usage and the zoink database.`,
	Run: func(cmd *cobra.Command, args []string) {
		handleStats()
	},
}

// cleanCmd represents the clean command
var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Remove non-existent directories",
	Long:  `Clean up the database by removing directories that no longer exist.`,
	Run: func(cmd *cobra.Command, args []string) {
		handleClean()
	},
}

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add [directory]",
	Short: "Manually add directory to database",
	Long:  `Manually add a directory to the zoink database without visiting it.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		handleAdd(args[0])
	},
}

// removeCmd represents the remove command
var removeCmd = &cobra.Command{
	Use:   "remove [directory]",
	Short: "Remove directory from database",
	Long:  `Remove a directory from the zoink database.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		handleRemove(args[0])
	},
}

func init() {
	rootCmd.AddCommand(statsCmd)
	rootCmd.AddCommand(cleanCmd)
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(removeCmd)
}

// handleStats displays usage statistics
func handleStats() {
	// Get database config
	cfg := GetConfig()
	dbConfig := database.DatabaseConfig{Path: cfg.DatabasePath}

	// Check if database exists
	if _, err := os.Stat(cfg.DatabasePath); os.IsNotExist(err) {
		fmt.Println("⚠️  Database does not exist yet")
		fmt.Println("Visit some directories or use 'zoink add /path' to create it")
		return
	}

	// Open database
	db, err := database.New(dbConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	// Get all entries
	entries, err := db.GetAll()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting entries: %v\n", err)
		os.Exit(1)
	}

	if len(entries) == 0 {
		fmt.Println("📊 Database is empty")
		return
	}

	// Calculate statistics
	var totalVisits uint32 = 0
	var maxVisits uint32 = 0
	var oldestEntry, newestEntry *database.DirectoryEntry

	for i, entry := range entries {
		totalVisits += entry.VisitCount
		if entry.VisitCount > maxVisits {
			maxVisits = entry.VisitCount
		}

		firstVisited := time.Unix(entry.FirstVisited, 0)
		lastVisited := time.Unix(entry.LastVisited, 0)

		if oldestEntry == nil || firstVisited.Before(time.Unix(oldestEntry.FirstVisited, 0)) {
			oldestEntry = entries[i]
		}
		if newestEntry == nil || lastVisited.After(time.Unix(newestEntry.LastVisited, 0)) {
			newestEntry = entries[i]
		}
	}

	avgVisits := float64(totalVisits) / float64(len(entries))

	// Display statistics
	fmt.Println("📊 Database Statistics")
	fmt.Println("===================")
	fmt.Println()
	fmt.Printf("Total entries: %d\n", len(entries))
	fmt.Printf("Total visits: %d\n", totalVisits)
	fmt.Printf("Average visits per directory: %.1f\n", avgVisits)
	fmt.Printf("Most visited directory: %d visits\n", maxVisits)

	if oldestEntry != nil {
		fmt.Printf("Oldest entry: %s (%s)\n",
			filepath.Base(oldestEntry.Path),
			time.Unix(oldestEntry.FirstVisited, 0).Format("2006-01-02"))
	}
	if newestEntry != nil {
		fmt.Printf("Most recent visit: %s (%s)\n",
			filepath.Base(newestEntry.Path),
			time.Unix(newestEntry.LastVisited, 0).Format("2006-01-02 15:04"))
	}

	// Show top 5 most visited
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].VisitCount > entries[j].VisitCount
	})

	fmt.Println("\n🏆 Top 5 Most Visited:")
	limit := 5
	if len(entries) < limit {
		limit = len(entries)
	}

	for i := 0; i < limit; i++ {
		entry := entries[i]
		lastVisited := time.Unix(entry.LastVisited, 0)
		lastVisit := "just now"
		if time.Since(lastVisited) > time.Minute {
			lastVisit = lastVisited.Format("Jan 2")
		}
		fmt.Printf("  %d. %s (%d visits, last: %s)\n",
			i+1, entry.Path, entry.VisitCount, lastVisit)
	}
}

// handleClean removes non-existent directories from database
func handleClean() {
	// Get database config
	cfg := GetConfig()
	dbConfig := database.DatabaseConfig{Path: cfg.DatabasePath}

	// Check if database exists
	if _, err := os.Stat(cfg.DatabasePath); os.IsNotExist(err) {
		fmt.Println("⚠️  Database does not exist yet")
		return
	}

	// Open database
	db, err := database.New(dbConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	// Get all entries
	entries, err := db.GetAll()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting entries: %v\n", err)
		os.Exit(1)
	}

	if len(entries) == 0 {
		fmt.Println("📊 Database is empty - nothing to clean")
		return
	}

	// Check which directories no longer exist
	var toRemove []string
	for _, entry := range entries {
		if _, err := os.Stat(entry.Path); os.IsNotExist(err) {
			toRemove = append(toRemove, entry.Path)
		}
	}

	if len(toRemove) == 0 {
		fmt.Printf("✅ All %d directories still exist - nothing to clean\n", len(entries))
		return
	}

	// Remove non-existent directories
	fmt.Printf("🧹 Cleaning %d non-existent directories:\n", len(toRemove))
	for _, path := range toRemove {
		fmt.Printf("  - %s\n", path)
		if err := db.RemoveDirectory(path); err != nil {
			fmt.Fprintf(os.Stderr, "Error removing %s: %v\n", path, err)
			continue
		}
	}

	// Save database
	if err := db.Save(); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving database: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Cleaned %d entries. %d directories remain.\n",
		len(toRemove), len(entries)-len(toRemove))
}

// handleAdd manually adds a directory to the database
func handleAdd(dir string) {
	// Convert to absolute path
	absDir, err := filepath.Abs(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error resolving path '%s': %v\n", dir, err)
		os.Exit(1)
	}

	// Check if directory exists
	if _, err := os.Stat(absDir); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: Directory '%s' does not exist\n", absDir)
		os.Exit(1)
	}

	// Get database config
	cfg := GetConfig()
	dbConfig := database.DatabaseConfig{Path: cfg.DatabasePath}

	// Open database
	db, err := database.New(dbConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	// Add visit
	if err := db.AddVisit(absDir); err != nil {
		fmt.Fprintf(os.Stderr, "Error adding visit: %v\n", err)
		os.Exit(1)
	}

	// Save database
	if err := db.Save(); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving database: %v\n", err)
		os.Exit(1)
	}

	// Only print success in verbose mode to avoid cluttering shell output
	if verbose, _ := rootCmd.PersistentFlags().GetBool("verbose"); verbose {
		fmt.Printf("Added visit to: %s\n", absDir)
	}
}

// handleRemove removes a directory from the database
func handleRemove(dir string) {
	// Convert to absolute path
	absDir, err := filepath.Abs(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error resolving path '%s': %v\n", dir, err)
		os.Exit(1)
	}

	// Get database config
	cfg := GetConfig()
	dbConfig := database.DatabaseConfig{Path: cfg.DatabasePath}

	// Check if database exists
	if _, err := os.Stat(cfg.DatabasePath); os.IsNotExist(err) {
		fmt.Println("⚠️  Database does not exist yet")
		return
	}

	// Open database
	db, err := database.New(dbConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	// Check if directory exists in database
	entries, err := db.GetAll()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting entries: %v\n", err)
		os.Exit(1)
	}

	found := false
	for _, entry := range entries {
		if entry.Path == absDir {
			found = true
			break
		}
	}

	if !found {
		fmt.Printf("Directory '%s' is not in the database\n", absDir)
		return
	}

	// Remove directory
	if err := db.RemoveDirectory(absDir); err != nil {
		fmt.Fprintf(os.Stderr, "Error removing directory: %v\n", err)
		os.Exit(1)
	}

	// Save database
	if err := db.Save(); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving database: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Removed: %s\n", absDir)
}
