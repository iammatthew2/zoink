package cmd

import (
	"fmt"
	"os"

	"github.com/iammatthew2/zoink/internal/config"
	"github.com/spf13/cobra"
)

var (
	cfg *config.Config
)

// base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "zoink [flags] [query]",
	Short: "Fast directory navigation with frecency",
	Long: `Zoink - Navigate directories quickly using frequency and recency.

Zoink tracks the directories you visit and helps you navigate to them quickly
using intelligent matching based on how often and how recently you've visited them.

Examples:
  zoink proj              Navigate to best project match
  zoink -i doc            Interactive selection for documents  
  zoink -l work           List work-related directories
  zoink --setup           Setup shell integration`,
	Args: cobra.ArbitraryArgs,
	Run:  executeZoink,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called in main
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().Bool("verbose", false, "verbose output")

	// Navigation flags
	rootCmd.Flags().BoolP("interactive", "i", false, "Force interactive selection")
	rootCmd.Flags().BoolP("list", "l", false, "List matches without navigating")
	rootCmd.Flags().BoolP("echo", "e", false, "Echo path only (for shell integration)")
	rootCmd.Flags().IntP("nth", "n", 0, "Select nth match directly (1-based)")

	// Matching mode flags
	rootCmd.Flags().BoolP("rank", "r", false, "Sort by frequency only")
	rootCmd.Flags().BoolP("recent", "t", false, "Sort by recency only")
	rootCmd.Flags().BoolP("exact", "x", false, "Exact matching only (no fuzzy)")
	rootCmd.Flags().BoolP("current", "c", false, "Search current directory children only")

	// Management flags
	rootCmd.Flags().Bool("setup", false, "Setup shell integration")
	rootCmd.Flags().Bool("clean", false, "Remove non-existent directories")
	rootCmd.Flags().Bool("stats", false, "Show usage statistics")
	rootCmd.Flags().String("add", "", "Manually add directory to database")
	rootCmd.Flags().String("remove", "", "Remove directory from database")

	// Output control flags (can override config defaults)
	rootCmd.Flags().Int("max-results", 10, "Maximum number of results to show")
	rootCmd.Flags().Float64("threshold", 0.8, "Auto-select confidence threshold (0.0-1.0)")

	// Setup modifiers
	rootCmd.Flags().Bool("quiet", false, "Non-interactive mode (use with --setup)")
	rootCmd.Flags().Bool("print-only", false, "Print shell code without installing (use with --setup)")

	// Version flag
	rootCmd.Flags().Bool("version", false, "Show version information")
}

// initConfig loads the configuration
func initConfig() {
	var err error
	cfg, err = config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Could not load config: %v\n", err)
		cfg = config.Default()
	}

	// Show config file location if verbose
	if verbose, _ := rootCmd.PersistentFlags().GetBool("verbose"); verbose {
		configDir, _ := config.GetConfigDir()
		fmt.Fprintf(os.Stderr, "Config directory: %s\n", configDir)
	}
}

// GetConfig returns the loaded configuration
func GetConfig() *config.Config {
	return cfg
}
