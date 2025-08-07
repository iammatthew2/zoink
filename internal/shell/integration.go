package shell

// GenerateHook creates the shell-specific integration code
func GenerateHook(shellName string) string {
	switch shellName {
	case "bash", "zsh":
		return bashZshHook
	case "fish":
		return fishHook
	default:
		return "# Unsupported shell"
	}
}

const bashZshHook = `# Zoink shell integration
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

# Main z command for navigation
z() {
    if [ $# -eq 0 ]; then
        # No arguments: let zoink handle the empty case
        local result
        result=$(zoink find)
        [ -n "$result" ] && [ -d "$result" ] && cd "$result"
    else
        # All arguments go to zoink find - let it handle everything
        local result
        result=$(zoink find "$@")
        if [ $? -eq 0 ] && [ -n "$result" ] && [ -d "$result" ]; then
            cd "$result"
        else
            # If no valid directory returned, just show the output
            echo "$result"
        fi
    fi
}

# Initialize tracking for current directory
zoink_track`

const fishHook = `# Zoink shell integration
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

# Main z command for navigation
function z
    if test (count $argv) -eq 0
        # No arguments: let zoink handle the empty case
        set result (zoink find)
        if test -n "$result" -a -d "$result"
            cd "$result"
        end
    else
        # All arguments go to zoink find - let it handle everything
        set result (zoink find $argv)
        if test $status -eq 0 -a -n "$result" -a -d "$result"
            cd "$result"
        else
            # If no valid directory returned, just show the output
            echo "$result"
        end
    end
end

# Initialize tracking for current directory
zoink_track`
