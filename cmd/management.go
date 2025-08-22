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
	Use:   "add [directory] [previous-directory]",
	Short: "Manually add directory to database",
	Long:  `Manually add a directory to the zoink database without visiting it.`,
	Args:  cobra.MaximumNArgs(2), // Allow 0-2 arguments
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			// No arguments - use current directory
			currentDir, err := os.Getwd()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error getting current directory: %v\n", err)
				os.Exit(1)
			}
			handleAdd(currentDir)
		} else if len(args) == 1 {
			handleAdd(args[0])
		} else {
			handleAddWithPrevious(args[0], args[1])
		}
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

// bookmarkCmd represents the bookmark command
var bookmarkCmd = &cobra.Command{
	Use:   "bookmark [name]",
	Short: "Add bookmark to current directory",
	Long:  `Add a bookmark to the current directory for quick navigation.`,
	Args:  cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		handleBookmarkArgs(args)
	},
}

func init() {
	rootCmd.AddCommand(statsCmd)
	rootCmd.AddCommand(cleanCmd)
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(removeCmd)
	rootCmd.AddCommand(bookmarkCmd)
}

// handleStats displays usage statistics
func handleStats() {
	// Get database config
	cfg := GetConfig()
	dbConfig := database.DatabaseConfig{Path: cfg.DatabasePath}

	// Check if database exists
	if _, err := os.Stat(cfg.DatabasePath); os.IsNotExist(err) {
		fmt.Println("Database does not exist yet")
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
		fmt.Println("Database is empty")
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
	fmt.Println("Database Statistics")
	fmt.Println("===================")
	fmt.Println()
	fmt.Printf("Database location: %s\n", cfg.DatabasePath)
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

	fmt.Println("\nTop 5 Most Visited:")
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
		fmt.Println("Database does not exist yet")
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
		fmt.Println("Database is empty - nothing to clean")
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
		fmt.Printf("All %d directories still exist - nothing to clean\n", len(entries))
		return
	}

	// Remove non-existent directories
	fmt.Printf("Cleaning %d non-existent directories:\n", len(toRemove))
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

	fmt.Printf("Cleaned %d entries. %d directories remain.\n",
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

// handleAddWithPrevious manually adds a directory to the database with previous directory
func handleAddWithPrevious(dir string, previousDir string) {
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

	// Add visit with previous directory
	if err := db.AddVisit(absDir, previousDir); err != nil {
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
		fmt.Printf("Added visit to: %s (from: %s)\n", absDir, previousDir)
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
		fmt.Println("Database does not exist yet")
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

	fmt.Printf("Removed: %s\n", absDir)
}

// handleBookmarkArgs processes bookmark command arguments
func handleBookmarkArgs(args []string) {
	// Filter out flags and get the bookmark name
	var bookmarkName string
	for _, arg := range args {
		if arg != "-b" && arg != "--bookmark" {
			bookmarkName = arg
			break
		}
	}

	if bookmarkName == "" {
		fmt.Fprintf(os.Stderr, "Error: bookmark requires a name\n")
		fmt.Fprintf(os.Stderr, "Usage: zoink bookmark <name>\n")
		os.Exit(1)
	}

	handleBookmark(bookmarkName)
}

// handleBookmark adds a bookmark to the current directory
func handleBookmark(bookmarkName string) {
	// Get current directory
	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting current directory: %v\n", err)
		os.Exit(1)
	}

	// Get database config
	cfg := GetConfig()
	dbConfig := database.DatabaseConfig{Path: cfg.DatabasePath}

	// Check if database exists
	if _, err := os.Stat(cfg.DatabasePath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Database does not exist yet. Visit some directories first.\n")
		os.Exit(1)
	}

	// Open database
	db, err := database.New(dbConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	// Add bookmark
	if err := db.AddBookmark(currentDir, bookmarkName); err != nil {
		fmt.Fprintf(os.Stderr, "Error adding bookmark: %v\n", err)
		os.Exit(1)
	}

	// Save database
	if err := db.Save(); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving database: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Bookmarked '%s' as '%s'\n", currentDir, bookmarkName)
}
