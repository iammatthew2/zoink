package cmd

import (
	"fmt"

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
	fmt.Println("Statistics not yet implemented")
}

// handleClean removes non-existent directories from database
func handleClean() {
	fmt.Println("Database cleanup not yet implemented")
}

// handleAdd manually adds a directory to the database
func handleAdd(dir string) {
	fmt.Printf("Adding directory '%s' not yet implemented\n", dir)
}

// handleRemove removes a directory from the database
func handleRemove(dir string) {
	fmt.Printf("Removing directory '%s' not yet implemented\n", dir)
}
