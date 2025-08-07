package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/iammatthew2/zoink/internal/database"
	"github.com/spf13/cobra"
)

// NavigationConfig holds the configuration for navigation operations
type NavigationConfig struct {
	Interactive bool
	ListOnly    bool
	EchoOnly    bool
	Recent      bool
	Frequent    bool
	MaxResults  int
	Threshold   float64
}

// buildConfigFromFlags extracts navigation configuration from command flags with optional config overrides
func buildConfigFromFlags(cmd *cobra.Command) *NavigationConfig {
	cfg := GetConfig()

	// Get flag values
	interactive, _ := cmd.Flags().GetBool("interactive")
	listOnly, _ := cmd.Flags().GetBool("list")
	echoOnly, _ := cmd.Flags().GetBool("echo")
	recent, _ := cmd.Flags().GetBool("recent")
	frequent, _ := cmd.Flags().GetBool("frequent")

	// Use config defaults for advanced settings
	maxResults := cfg.MaxResults
	if maxResults <= 0 {
		maxResults = 10 // sensible default
	}

	threshold := cfg.Threshold
	if threshold <= 0 {
		threshold = 0.8 // sensible default
	}

	return &NavigationConfig{
		Interactive: interactive,
		ListOnly:    listOnly,
		EchoOnly:    echoOnly,
		Recent:      recent,
		Frequent:    frequent,
		MaxResults:  maxResults,
		Threshold:   threshold,
	}
}

// handleNavigation processes directory navigation requests
func handleNavigation(query string, config *NavigationConfig) {
	// Get database config
	cfg := GetConfig()
	dbConfig := database.DatabaseConfig{Path: cfg.DatabasePath}

	// Check if database exists
	if _, err := os.Stat(cfg.DatabasePath); os.IsNotExist(err) {
		if config.ListOnly {
			fmt.Println("Database does not exist yet")
			return
		}
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

	// Query database
	var entries []*database.DirectoryEntry
	if query == "" {
		// No query - get all entries for interactive selection
		entries, err = db.GetAll()
	} else {
		// Query with search term
		entries, err = db.Query(query, config.MaxResults)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error querying database: %v\n", err)
		os.Exit(1)
	}

	// Handle no results
	if len(entries) == 0 {
		if config.ListOnly {
			if query == "" {
				fmt.Println("Database is empty")
			} else {
				fmt.Printf("No directories found matching '%s'\n", query)
			}
			return
		}
		fmt.Fprintf(os.Stderr, "No directories found matching '%s'\n", query)
		os.Exit(1)
	}

	// Handle list-only mode
	if config.ListOnly {
		printDirectoryList(entries, config.EchoOnly)
		return
	}

	// Select directory
	selectedPath := selectDirectory(entries, config)
	if selectedPath == "" {
		os.Exit(1)
	}

	// Output the selected path
	if config.EchoOnly {
		// No newline for shell integration
		fmt.Print(selectedPath)
	} else {
		fmt.Println(selectedPath)
	}
}

// selectDirectory handles directory selection logic
func selectDirectory(entries []*database.DirectoryEntry, config *NavigationConfig) string {
	// Single result - return it directly
	if len(entries) == 1 {
		return entries[0].Path
	}

	// Multiple results - handle based on config
	if config.Interactive {
		return selectInteractively(entries)
	}

	// Non-interactive with multiple results - return best match
	return entries[0].Path
}

// selectInteractively shows an interactive selection menu
func selectInteractively(entries []*database.DirectoryEntry) string {
	if len(entries) == 0 {
		return ""
	}

	// Create options for selection
	var options []string
	for _, entry := range entries {
		options = append(options, entry.Path)
	}

	var selected string
	prompt := &survey.Select{
		Message: "Select directory:",
		Options: options,
	}

	if err := survey.AskOne(prompt, &selected); err != nil {
		return "" // User cancelled
	}

	return selected
}

// printDirectoryList prints a formatted list of directories
func printDirectoryList(entries []*database.DirectoryEntry, simpleFormat bool) {
	if simpleFormat {
		// just paths, one per line
		for _, entry := range entries {
			fmt.Println(entry.Path)
		}
		return
	}

	fmt.Printf("Found %d director", len(entries))
	if len(entries) == 1 {
		fmt.Println("y:")
	} else {
		fmt.Println("ies:")
	}
	fmt.Println()

	for i, entry := range entries {
		fmt.Printf("  %d. %s\n", i+1, entry.Path)
		fmt.Printf("     Visits: %d | Last: %s\n",
			entry.VisitCount,
			formatLastVisit(entry.LastVisited))
		if i < len(entries)-1 {
			fmt.Println()
		}
	}
}

// formatLastVisit formats the last visit timestamp
func formatLastVisit(timestamp int64) string {
	lastVisited := time.Unix(timestamp, 0)
	elapsed := time.Since(lastVisited)

	if elapsed < time.Minute {
		return "just now"
	} else if elapsed < time.Hour {
		mins := int(elapsed.Minutes())
		return fmt.Sprintf("%dm ago", mins)
	} else if elapsed < 24*time.Hour {
		hours := int(elapsed.Hours())
		return fmt.Sprintf("%dh ago", hours)
	} else {
		days := int(elapsed.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	}
}

// handleEmptyQuery handles the case when no query is provided from shell integration
func handleEmptyQuery() {
	// Get database config
	cfg := GetConfig()
	dbConfig := database.DatabaseConfig{Path: cfg.DatabasePath}

	// Check if database exists
	if _, err := os.Stat(cfg.DatabasePath); os.IsNotExist(err) {
		// No database yet - nothing to do
		return
	}

	// Open database
	db, err := database.New(dbConfig)
	if err != nil {
		// Silently fail for shell integration
		return
	}
	defer db.Close()

	// Get all entries and return the best one (most frecent)
	entries, err := db.Query("", 1) // Get top 1 result with empty query
	if err != nil || len(entries) == 0 {
		// Silently fail for shell integration
		return
	}

	// Output the best directory
	fmt.Print(entries[0].Path)
}
