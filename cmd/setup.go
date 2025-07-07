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
	setupCmd.Flags().String("shell", "", "Target shell (bash, zsh, fish)")
	setupCmd.Flags().String("alias", "", "Alias name (default: z, or x for development)")
}

// ShellInfo holds information about a detected shell
type ShellInfo struct {
	Name       string
	ConfigFile string
	Binary     string
}

// handleSetupCommand manages shell integration setup
func handleSetupCommand(cmd *cobra.Command, args []string) {
	quiet, _ := cmd.Flags().GetBool("quiet")
	printOnly, _ := cmd.Flags().GetBool("print-only")

	// Detect available shells
	shells := detectShells()
	if len(shells) == 0 {
		fmt.Fprintf(os.Stderr, "Error: No supported shells found (bash, zsh, fish)\n")
		os.Exit(1)
	}

	// Handle print-only mode
	if printOnly {
		handlePrintOnly(shells)
		return
	}

	// Handle quiet (non-interactive) mode
	if quiet {
		handleQuietSetup(shells)
		return
	}

	// Interactive setup
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
	if err := installShellHook(*targetShell); err != nil {
		if strings.HasPrefix(err.Error(), "ALREADY_INSTALLED:") {
			configFile := strings.TrimPrefix(err.Error(), "ALREADY_INSTALLED:")
			fmt.Printf("Zoink integration already configured in %s\n", configFile)
			return
		}
		fmt.Fprintf(os.Stderr, "Error installing shell hook: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Zoink shell integration installed for %s\n", targetShell.Name)
	fmt.Printf("Config file: %s\n", targetShell.ConfigFile)
	fmt.Printf("Please restart your shell or run: source %s\n", targetShell.ConfigFile)
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

		if err := installShellHook(shell); err != nil {
			if strings.HasPrefix(err.Error(), "ALREADY_INSTALLED:") {
				configFile := strings.TrimPrefix(err.Error(), "ALREADY_INSTALLED:")
				fmt.Printf("Zoink integration already configured in %s\n", filepath.Base(configFile))
				continue
			}
			fmt.Fprintf(os.Stderr, "Error installing %s hook: %v\n", shell.Name, err)
			continue
		}

		fmt.Printf("Successfully configured %s\n", shell.Name)
	}

	fmt.Println("\nSetup complete!")
	fmt.Println("Please restart your shell(s) or source the config files to activate Zoink.")
}

// generateShellHook creates the shell-specific integration code
func generateShellHook(shellName string) string {
	switch shellName {
	case "bash", "zsh":
		return `# Zoink shell integration
zoink_track() {
    if command -v zoink >/dev/null 2>&1; then
        zoink add "$PWD" >/dev/null 2>&1
    fi
}

# Hook into cd command
cd() {
    builtin cd "$@" && zoink_track
}

# Hook into pushd/popd
pushd() {
    builtin pushd "$@" && zoink_track
}

popd() {
    builtin popd "$@" && zoink_track
}

# Main x command for navigation (dev alias - change to "z" once stable)
x() {
    if [ $# -eq 0 ]; then
        # No arguments: show interactive selection
        local result
        result=$(zoink --interactive)
        [ -n "$result" ] && cd "$result"
    else
        # Check if any argument starts with a dash (flag)
        local has_flags=false
        for arg in "$@"; do
            case "$arg" in
                -*) has_flags=true; break ;;
            esac
        done
        
        if [ "$has_flags" = true ]; then
            # Has flags: run directly without cd (let zoink handle it)
            zoink "$@"
        else
            # No flags: treat as navigation query
            local result
            result=$(zoink "$@")
            if [ $? -eq 0 ] && [ -n "$result" ] && [ -d "$result" ]; then
                cd "$result"
            else
                # If no valid directory returned, just show the output
                echo "$result"
            fi
        fi
    fi
}

# Initialize tracking for current directory
zoink_track`

	case "fish":
		return `# Zoink shell integration
function zoink_track
    if command -v zoink >/dev/null 2>&1
        zoink add $PWD >/dev/null 2>&1
    end
end

# Hook into cd command
function cd
    builtin cd $argv
    and zoink_track
end

# Main x command for navigation (dev alias - change to "z" once stable)
function x
    if test (count $argv) -eq 0
        # No arguments: show interactive selection
        set result (zoink --interactive)
        if test -n "$result"
            cd "$result"
        end
    else
        # Check if any argument starts with a dash (flag)
        set has_flags false
        for arg in $argv
            if string match -q -- '-*' $arg
                set has_flags true
                break
            end
        end
        
        if test "$has_flags" = "true"
            # Has flags: run directly without cd (let zoink handle it)
            zoink $argv
        else
            # No flags: treat as navigation query
            set result (zoink $argv)
            if test $status -eq 0 -a -n "$result" -a -d "$result"
                cd "$result"
            else
                # If no valid directory returned, just show the output
                echo "$result"
            end
        end
    end
end

# Initialize tracking for current directory
zoink_track`

	default:
		return "# Unsupported shell"
	}
}

// installShellHook creates shell integration files and adds source line to user's config
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
	hookCode := generateShellHook(shell.Name)

	if err := os.WriteFile(shellFile, []byte(hookCode), 0644); err != nil {
		return fmt.Errorf("failed to write shell integration file: %v", err)
	}

	// Add source line to user's shell config
	marker := "# Zoink shell integration"
	sourceLine := generateSourceLine(shell.Name, shellFile)

	// Check if already installed
	if isHookInstalled(shell.ConfigFile, marker) {
		// Return a special message indicating it's already configured (not an error)
		return fmt.Errorf("ALREADY_INSTALLED:%s", shell.ConfigFile)
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
