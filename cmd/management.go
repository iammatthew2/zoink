package cmd

import (
	"fmt"
)

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
