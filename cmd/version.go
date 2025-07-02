package cmd

import (
	"fmt"
)

var (
	version   = "0.1.0" // Will be set by build process
	buildTime = "dev"   // Will be set by build process
)

// handleVersion displays version information
func handleVersion() {
	fmt.Printf("zoink version %s\n", version)
	if buildTime != "dev" {
		fmt.Printf("built at %s\n", buildTime)
	}
}
