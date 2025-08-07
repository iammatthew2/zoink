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
	Use:   "zoink [query]",
	Short: "Fast directory navigation with frecency",
	Long: `Zoink - Navigate directories quickly using frequency and recency.

Zoink tracks the directories you visit and helps you navigate to them quickly
using intelligent matching based on how often and how recently you've visited them.

Primary usage (via shell alias):
  z foo                  Navigate to best project match
  z -i doc               Interactive selection for documents  
  z -l work              List work-related directories

Direct usage:
  zoink find foo         Navigate to best project match
  zoink setup            Setup shell integration
  zoink stats            Show usage statistics`,
	Args: cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		// Handle version flag
		if version, _ := cmd.Flags().GetBool("version"); version {
			handleVersion()
			return
		}
		// For root command without subcommands, just show help
		cmd.Help()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called in main
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")
	rootCmd.Flags().BoolP("version", "V", false, "Show version information")
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
