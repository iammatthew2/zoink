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

# Main x command for navigation (dev alias - change to "z" once stable)
x() {
    if [ $# -eq 0 ]; then
        # No arguments: show interactive selection
        local result
        result=$(zoink find --interactive)
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
            # Has flags: run find command directly without cd (let zoink handle it)
            zoink find "$@"
        else
            # No flags: treat as navigation query
            local result
            result=$(zoink find "$@")
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

# Main x command for navigation (dev alias - change to "z" once stable)
function x
    if test (count $argv) -eq 0
        # No arguments: show interactive selection
        set result (zoink find --interactive)
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
            # Has flags: run find command directly without cd (let zoink handle it)
            zoink find $argv
        else
            # No flags: treat as navigation query
            set result (zoink find $argv)
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
