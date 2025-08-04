package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	shellpkg "github.com/iammatthew2/zoink/internal/shell"
	"github.com/spf13/cobra"
)

// setupCmd represents the setup command
var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Setup shell integration",
	Long: `Setup shell integration for zoink.

This command will detect your shell and install the necessary hooks
to enable the 'z' command for directory navigation.

Examples:
  zoink setup                    Interactive setup
  zoink setup --quiet            Non-interactive setup
  zoink setup --print-only       Show shell code without installing`,
	Run: handleSetupCommand,
}

func init() {
	rootCmd.AddCommand(setupCmd)

	setupCmd.Flags().Bool("quiet", false, "Non-interactive mode")
	setupCmd.Flags().Bool("print-only", false, "Print shell code without installing")
}

// ShellInfo holds information about a detected shell
type ShellInfo struct {
	Name       string
	ConfigFile string
	Binary     string
}

// handleSetupCommand manages shell integration
func handleSetupCommand(cmd *cobra.Command, args []string) {
	quiet, _ := cmd.Flags().GetBool("quiet")
	printOnly, _ := cmd.Flags().GetBool("print-only")

	// Detect available shells
	shells := detectShells()
	if len(shells) == 0 {
		fmt.Fprintf(os.Stderr, "Error: No supported shells found (bash, zsh, fish)\n")
		os.Exit(1)
	}

	if printOnly {
		handlePrintOnly(shells)
		return
	}

	if quiet {
		handleQuietSetup(shells)
		return
	}

	handleInteractiveSetup(shells)
}

// detectShells finds available shells and their config files
func detectShells() []ShellInfo {
	var shells []ShellInfo
	homeDir, _ := os.UserHomeDir()

	// Check for bash
	if bashPath, err := exec.LookPath("bash"); err == nil {
		configFile := filepath.Join(homeDir, ".bashrc")
		if runtime.GOOS == "darwin" {
			// big assumption here
			// macOS uses .bash_profile by default
			if _, err := os.Stat(filepath.Join(homeDir, ".bash_profile")); err == nil {
				configFile = filepath.Join(homeDir, ".bash_profile")
			}
		}
		shells = append(shells, ShellInfo{
			Name:       "bash",
			ConfigFile: configFile,
			Binary:     bashPath,
		})
	}

	// Check for zsh
	if zshPath, err := exec.LookPath("zsh"); err == nil {
		shells = append(shells, ShellInfo{
			Name:       "zsh",
			ConfigFile: filepath.Join(homeDir, ".zshrc"),
			Binary:     zshPath,
		})
	}

	// Check for fish
	if fishPath, err := exec.LookPath("fish"); err == nil {
		configDir := filepath.Join(homeDir, ".config", "fish")
		shells = append(shells, ShellInfo{
			Name:       "fish",
			ConfigFile: filepath.Join(configDir, "config.fish"),
			Binary:     fishPath,
		})
	}

	return shells
}

// handlePrintOnly prints shell integration code without installing
func handlePrintOnly(shells []ShellInfo) {
	fmt.Println("# Zoink Shell Integration")
	fmt.Println("# Run 'zoink setup' to install automatically, or add these lines manually:")
	fmt.Println()

	// Get config directory for display
	configDir, err := getZoinkConfigDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Could not determine config directory: %v\n", err)
		return
	}

	for _, shell := range shells {
		shellFile := filepath.Join(configDir, "shell", getShellFileName(shell.Name))
		sourceLine := generateSourceLine(shell.Name, shellFile)

		fmt.Printf("# For %s (%s):\n", shell.Name, shell.ConfigFile)
		fmt.Printf("# Add this line to your shell config:\n")
		fmt.Printf("%s\n", sourceLine)
		fmt.Printf("# (Shell integration file will be created at: %s)\n", shellFile)
		fmt.Println()
	}
}

// handleQuietSetup performs non-interactive setup for the current shell
func handleQuietSetup(shells []ShellInfo) {
	// Detect current shell from SHELL environment variable
	currentShell := os.Getenv("SHELL")
	if currentShell == "" {
		fmt.Fprintf(os.Stderr, "Error: Could not detect current shell from SHELL environment variable\n")
		os.Exit(1)
	}

	shellName := filepath.Base(currentShell)

	// Find the matching shell
	var targetShell *ShellInfo
	for _, shell := range shells {
		if shell.Name == shellName {
			targetShell = &shell
			break
		}
	}

	if targetShell == nil {
		fmt.Fprintf(os.Stderr, "Error: Current shell '%s' is not supported\n", shellName)
		os.Exit(1)
	}

	// Install for the detected shell
	success, showAdditionalInfo := installAndReportShellHook(*targetShell, true)
	if !success {
		os.Exit(1)
	}

	// Only show additional success details for completely new installations
	if showAdditionalInfo {
		fmt.Printf("Config file: %s\n", targetShell.ConfigFile)
		fmt.Printf("Please restart your shell or run: source %s\n", targetShell.ConfigFile)
	}
}

// handleInteractiveSetup provides interactive shell selection and installation
func handleInteractiveSetup(shells []ShellInfo) {
	fmt.Println("Zoink Shell Integration Setup")
	fmt.Println("=============================")
	fmt.Println()

	// Create shell options for selection
	var options []string
	for _, shell := range shells {
		options = append(options, fmt.Sprintf("%s (%s)", shell.Name, shell.ConfigFile))
	}

	// Let user select shells to configure
	var selected []int
	prompt := &survey.MultiSelect{
		Message: "Select shells to configure:",
		Options: options,
	}

	if err := survey.AskOne(prompt, &selected); err != nil {
		fmt.Fprintf(os.Stderr, "Setup cancelled: %v\n", err)
		os.Exit(1)
	}

	if len(selected) == 0 {
		fmt.Println("No shells selected. Setup cancelled.")
		return
	}

	// Install hooks for selected shells
	for _, idx := range selected {
		shell := shells[idx]
		fmt.Printf("\nInstalling Zoink integration for %s...\n", shell.Name)
		installAndReportShellHook(shell, false)
	}

	fmt.Println("\nSetup complete!")
	fmt.Println("Please restart your shell(s) or source the config files to activate Zoink.")
}

// Installs shell hook and handles all messaging
// Returns (success, shouldShowAdditionalInfo)
func installAndReportShellHook(shell ShellInfo, quiet bool) (bool, bool) {
	alreadyInstalled := isHookInstalled(shell.ConfigFile, "# Zoink shell integration")

	if err := installShellHook(shell); err != nil {
		if quiet {
			fmt.Fprintf(os.Stderr, "Error installing shell hook: %v\n", err)
		} else {
			fmt.Fprintf(os.Stderr, "Error installing %s hook: %v\n", shell.Name, err)
		}
		return false, false
	}

	// Success - simple message for all cases
	fmt.Printf("Installed (or updated) shell integration for %s\n", shell.Name)

	// Only show additional info for completely new installations
	return true, !alreadyInstalled
}

// Creates shell integration files and adds source line to user's config
func installShellHook(shell ShellInfo) error {
	// Get zoink config directory
	configDir, err := getZoinkConfigDir()
	if err != nil {
		return fmt.Errorf("failed to get zoink config directory: %v", err)
	}

	// Create shell integration file
	shellDir := filepath.Join(configDir, "shell")
	if err := os.MkdirAll(shellDir, 0755); err != nil {
		return fmt.Errorf("failed to create shell directory: %v", err)
	}

	shellFile := filepath.Join(shellDir, getShellFileName(shell.Name))
	hookCode := shellpkg.GenerateHook(shell.Name)

	// Write/overwrite the shell integration file
	if err := os.WriteFile(shellFile, []byte(hookCode), 0644); err != nil {
		return fmt.Errorf("failed to write shell integration file: %v", err)
	}

	// Add source line to user's shell config if not already present
	marker := "# Zoink shell integration"
	sourceLine := generateSourceLine(shell.Name, shellFile)

	// If already installed, we're done (file was already updated above)
	if isHookInstalled(shell.ConfigFile, marker) {
		return nil
	}

	// Create user config file directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(shell.ConfigFile), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %v", err)
	}

	// Append source line to config file
	file, err := os.OpenFile(shell.ConfigFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open config file: %v", err)
	}
	defer file.Close()

	_, err = fmt.Fprintf(file, "\n%s\n%s\n", marker, sourceLine)
	if err != nil {
		return fmt.Errorf("failed to write source line: %v", err)
	}

	return nil
}

// isHookInstalled checks if the Zoink hook is already present in the config file
func isHookInstalled(configFile, marker string) bool {
	file, err := os.Open(configFile)
	if err != nil {
		return false // File doesn't exist, so not installed
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), marker) {
			return true
		}
	}

	return false
}

// getZoinkConfigDir returns the zoink configuration directory
func getZoinkConfigDir() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "zoink"), nil
}

// getShellFileName returns the filename for shell integration scripts
func getShellFileName(shellName string) string {
	switch shellName {
	case "bash":
		return "bash.sh"
	case "zsh":
		return "zsh.sh"
	case "fish":
		return "fish.fish"
	default:
		return shellName + ".sh"
	}
}

// generateSourceLine creates the appropriate source line for each shell
func generateSourceLine(shellName, shellFile string) string {
	switch shellName {
	case "bash", "zsh":
		return fmt.Sprintf("source \"%s\"", shellFile)
	case "fish":
		return fmt.Sprintf("source \"%s\"", shellFile)
	default:
		return fmt.Sprintf("source \"%s\"", shellFile)
	}
}
