package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	version = "0.1.0" // Will be set by build process
)

// rootCmd represents the base command when called without any subcommands
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
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.config/zoink/config.toml)")
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

	// Output control flags
	rootCmd.Flags().Int("max-results", 10, "Maximum number of results to show")
	rootCmd.Flags().Float64("threshold", 0.8, "Auto-select confidence threshold (0.0-1.0)")

	// Setup modifiers
	rootCmd.Flags().Bool("quiet", false, "Non-interactive mode (use with --setup)")
	rootCmd.Flags().Bool("print-only", false, "Print shell code without installing (use with --setup)")

	// Version flag
	rootCmd.Flags().Bool("version", false, "Show version information")

	// Bind flags to viper for config file and environment variable support
	viper.BindPFlag("interactive", rootCmd.Flags().Lookup("interactive"))
	viper.BindPFlag("max-results", rootCmd.Flags().Lookup("max-results"))
	viper.BindPFlag("threshold", rootCmd.Flags().Lookup("threshold"))
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))

	// Environment variable support
	viper.SetEnvPrefix("ZOINK")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		// Search config in home directory and config directories
		viper.AddConfigPath(home + "/.config/zoink")
		viper.AddConfigPath(home)
		viper.SetConfigType("toml")
		viper.SetConfigName("config")
	}

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil && viper.GetBool("verbose") {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}

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

func buildConfigFromFlags(cmd *cobra.Command) *NavigationConfig {
	interactive, _ := cmd.Flags().GetBool("interactive")
	listOnly, _ := cmd.Flags().GetBool("list")
	echoOnly, _ := cmd.Flags().GetBool("echo")
	rankOnly, _ := cmd.Flags().GetBool("rank")
	recentOnly, _ := cmd.Flags().GetBool("recent")
	exactMatch, _ := cmd.Flags().GetBool("exact")
	currentOnly, _ := cmd.Flags().GetBool("current")
	maxResults, _ := cmd.Flags().GetInt("max-results")
	threshold, _ := cmd.Flags().GetFloat64("threshold")
	nthMatch, _ := cmd.Flags().GetInt("nth")

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

// Placeholder handlers - we'll implement these in separate files
func handleVersion() {
	fmt.Printf("zoink version %s\n", version)
}

func handleSetup(cmd *cobra.Command) {
	quiet, _ := cmd.Flags().GetBool("quiet")
	printOnly, _ := cmd.Flags().GetBool("print-only")

	if printOnly {
		fmt.Print("# Zoink shell integration code will go here")
		return
	}

	if quiet {
		fmt.Println("Non-interactive setup not yet implemented")
	} else {
		fmt.Println("Interactive setup not yet implemented")
	}
}

func handleStats() {
	fmt.Println("Statistics not yet implemented")
}

func handleClean() {
	fmt.Println("Database cleanup not yet implemented")
}

func handleAdd(dir string) {
	fmt.Printf("Adding directory '%s' not yet implemented\n", dir)
}

func handleRemove(dir string) {
	fmt.Printf("Removing directory '%s' not yet implemented\n", dir)
}

func handleNavigation(query string, config *NavigationConfig) {
	fmt.Printf("Navigation for query '%s' not yet implemented\n", query)
	fmt.Printf("Config: %+v\n", config)
}
