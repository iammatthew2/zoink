package cmd

import (
	"github.com/spf13/cobra"
)

// findCmd represents the find command for navigation
var findCmd = &cobra.Command{
	Use:   "find [query]",
	Short: "Find and navigate to directories (used by shell alias)",
	Long: `Find and navigate to directories based on frequency and recency.

This command is primarily used by the Zoink shell alias (z) for directory navigation.
It provides the core directory matching and selection functionality.

For general help and usage information, run: zoink --help

Examples (via shell alias):
  x foo                  Navigate to best project match
  x -i foo               Interactive selection for foo-related documents
  x -l foo               List foo-related directories`,
	Args: cobra.ArbitraryArgs,
	Run:  executeZoink, // Reuse the same navigation logic
}

func init() {
	rootCmd.AddCommand(findCmd)

	// Navigation flags (same as root command)
	findCmd.Flags().BoolP("interactive", "i", false, "Interactive selection when multiple matches")
	findCmd.Flags().BoolP("list", "l", false, "List matches without navigating")
	findCmd.Flags().BoolP("echo", "e", false, "Echo path only (for shell integration)")
	findCmd.Flags().BoolP("recent", "t", false, "Prefer recent directories")
	findCmd.Flags().BoolP("frequent", "f", false, "Prefer frequently used directories")
}
