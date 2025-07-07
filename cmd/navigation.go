package cmd

import (
	"fmt"
	"strings"

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
	fmt.Printf("Navigation for query '%s' not yet implemented\n", query)
	fmt.Printf("Config: %+v\n", config)
}

// executeZoink is the main command handler for the root command
func executeZoink(cmd *cobra.Command, args []string) {
	// Handle version flag (for backwards compatibility)
	if version, _ := cmd.Flags().GetBool("version"); version {
		handleVersion()
		return
	}

	// Main navigation logic
	query := strings.Join(args, " ")
	config := buildConfigFromFlags(cmd)

	// If no query and not interactive, show help
	if query == "" && !config.Interactive {
		cmd.Help()
		return
	}

	handleNavigation(query, config)
}
