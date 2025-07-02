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
	RankOnly    bool
	RecentOnly  bool
	ExactMatch  bool
	CurrentOnly bool
	MaxResults  int
	Threshold   float64
	NthMatch    int
}

// buildConfigFromFlags extracts navigation configuration from command flags with optional config overrides
func buildConfigFromFlags(cmd *cobra.Command) *NavigationConfig {
	cfg := GetConfig()

	// Start with flag values (which include Cobra's defaults)
	interactive, _ := cmd.Flags().GetBool("interactive")
	listOnly, _ := cmd.Flags().GetBool("list")
	echoOnly, _ := cmd.Flags().GetBool("echo")
	rankOnly, _ := cmd.Flags().GetBool("rank")
	recentOnly, _ := cmd.Flags().GetBool("recent")
	exactMatch, _ := cmd.Flags().GetBool("exact")
	currentOnly, _ := cmd.Flags().GetBool("current")
	nthMatch, _ := cmd.Flags().GetInt("nth")
	maxResults, _ := cmd.Flags().GetInt("max-results")
	threshold, _ := cmd.Flags().GetFloat64("threshold")

	// Override with config ONLY if user customized it AND flag wasn't explicitly set
	if !cmd.Flags().Changed("max-results") && cfg.MaxResults > 0 {
		maxResults = cfg.MaxResults
	}

	if !cmd.Flags().Changed("threshold") && cfg.Threshold > 0 {
		threshold = cfg.Threshold
	}

	return &NavigationConfig{
		Interactive: interactive,
		ListOnly:    listOnly,
		EchoOnly:    echoOnly,
		RankOnly:    rankOnly,
		RecentOnly:  recentOnly,
		ExactMatch:  exactMatch,
		CurrentOnly: currentOnly,
		MaxResults:  maxResults,
		Threshold:   threshold,
		NthMatch:    nthMatch,
	}
}

// handleNavigation processes directory navigation requests
func handleNavigation(query string, config *NavigationConfig) {
	fmt.Printf("Navigation for query '%s' not yet implemented\n", query)
	fmt.Printf("Config: %+v\n", config)
}

// executeZoink is the main command handler
func executeZoink(cmd *cobra.Command, args []string) {
	// Handle version flag
	if version, _ := cmd.Flags().GetBool("version"); version {
		handleVersion()
		return
	}

	// Handle setup command
	if setup, _ := cmd.Flags().GetBool("setup"); setup {
		handleSetup(cmd)
		return
	}

	// Handle management commands
	if stats, _ := cmd.Flags().GetBool("stats"); stats {
		handleStats()
		return
	}

	if clean, _ := cmd.Flags().GetBool("clean"); clean {
		handleClean()
		return
	}

	// Handle manual directory operations
	if addDir, _ := cmd.Flags().GetString("add"); addDir != "" {
		handleAdd(addDir)
		return
	}

	if removeDir, _ := cmd.Flags().GetString("remove"); removeDir != "" {
		handleRemove(removeDir)
		return
	}

	// Main navigation logic
	query := strings.Join(args, " ")
	config := buildConfigFromFlags(cmd)

	if query == "" && !config.Interactive {
		cmd.Help()
		return
	}

	handleNavigation(query, config)
}
