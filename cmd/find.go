package cmd

import (
	"strings"

	"github.com/spf13/cobra"
)

// findCmd represents the find command for navigation
var findCmd = &cobra.Command{
	Use:   "find [query]",
	Short: "Find and navigate to directories (used by shell alias)",
	Long: `Find and navigate to directories based on frequency and recency.

This command is used by the shell alias (z) for directory navigation.
It provides the core directory matching and selection functionality.

Examples (via shell alias):
  z foo                  Navigate to best project match
  z -i foo               Interactive selection for foo-related documents
  z -l foo               List foo-related directories`,
	Args: cobra.ArbitraryArgs,
	Run:  executeFind,
}

func init() {
	rootCmd.AddCommand(findCmd)

	// Navigation flags
	findCmd.Flags().BoolP("interactive", "i", false, "Interactive selection when multiple matches")
	findCmd.Flags().BoolP("list", "l", false, "List matches without navigating")
	findCmd.Flags().BoolP("echo", "e", false, "Echo path only (for shell integration)")
	findCmd.Flags().BoolP("recent", "t", false, "Prefer recent directories")
	findCmd.Flags().BoolP("frequent", "f", false, "Prefer frequently used directories")
}

// executeFind is the main command handler for the find command
func executeFind(cmd *cobra.Command, args []string) {
	// Handle version flag (for backwards compatibility)
	if version, _ := cmd.Flags().GetBool("version"); version {
		handleVersion()
		return
	}

	// Main navigation logic
	query := strings.Join(args, " ")
	config := buildConfigFromFlags(cmd)

	// Handle empty query - return most frecent directory for shell integration
	if query == "" && !config.Interactive && !config.ListOnly {
		handleEmptyQuery()
		return
	}

	handleNavigation(query, config)
}
