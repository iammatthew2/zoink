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
        zoink add "$PWD" "$OLDPWD" >/dev/null 2>&1
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
        # Check if interactive flag is present
        case "$*" in
            *-i*|*--interactive*)
                # Interactive mode
                if ! command -v fzf >/dev/null 2>&1; then
                    echo "fzf is required for interactive mode. Please install fzf." >&2
                    return 1
                fi
                # Strip interactive flags and use remaining args for search
                local search_args=$(echo "$@" | sed 's/-i//g; s/--interactive//g' | xargs)
                local dir
                if [ -n "$search_args" ]; then
                    dir=$(zoink find --list --echo "$search_args" | fzf --height 40% --reverse --header "Select directory:")
                else
                    dir=$(zoink find --list --echo | fzf --height 40% --reverse --header "Select directory:")
                fi
                [ -n "$dir" ] && [ -d "$dir" ] && cd "$dir"
                ;;
            *)
                # Non-interactive mode
                local result
                result=$(zoink find "$@")
                if [ $? -eq 0 ] && [ -n "$result" ] && [ -d "$result" ]; then
                    cd "$result"
                else
                    # If no valid directory returned, just show the output
                    echo "$result"
                fi
                ;;
        esac
    fi
}

# Initialize tracking for current directory
zoink_track`

const fishHook = `# Zoink shell integration
function zoink_track
    if command -v zoink >/dev/null 2>&1
        zoink add $PWD $OLDPWD >/dev/null 2>&1
    end
end

# Hook into cd command
function cd
    builtin cd $argv
    and zoink_track
end

# Hook into pushd/popd
pushd() {
    builtin pushd "$@" && zoink_track
}

popd() {
    builtin popd "$@" && zoink_track
}

# Main z command for navigation
function z
    if test (count $argv) -eq 0
        # No arguments: let zoink handle the empty case
        set result (zoink find)
        if test -n "$result" -a -d "$result"
            cd "$result"
        end
    else
        # Check if interactive flag is present
        set interactive_mode 0
        for arg in $argv
            if test "$arg" = "-i" -o "$arg" = "--interactive"
                set interactive_mode 1
                break
            end
        end
        
        if test $interactive_mode -eq 1
            # Interactive mode with fzf
            if not command -v fzf >/dev/null 2>&1
                echo "fzf is required for interactive mode. Please install fzf." >&2
                return 1
            end
            # Strip interactive flags and use remaining args for search
            set search_args
            for arg in $argv
                if test "$arg" != "-i" -a "$arg" != "--interactive"
                    set search_args $search_args $arg
                end
            end
            set dir
            if test (count $search_args) -gt 0
                set dir (zoink find --list --echo $search_args | fzf --height 40% --reverse --header "Select directory:")
            else
                set dir (zoink find --list --echo | fzf --height 40% --reverse --header "Select directory:")
            end
            test -n "$dir" -a -d "$dir"; and cd "$dir"
        else
            # Non-interactive mode
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
