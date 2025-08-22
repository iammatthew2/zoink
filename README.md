# Zoink

Zoink is a fast, cross-shell tool for directory navigation with directory bookmarking. It is inspired by [z.sh](https://github.com/rupa/z), [fasd](https://github.com/clvv/fasd), and [autojump](https://github.com/wting/autojump), leveraging a similar algorithm for fuzzy/frecency ranking for directory lookup.

The alias `z` is the daily driver with the `zoink` executable providing setup and configuration options.

## Quick start

Install via [Homebrew](https://brew.sh/)

`brew install iammatthew2/tap/zoink` 

Shell integration is automatically configured, giving you access to the `z` alias

Navigate around your directories (cd here and there) then use the zoink alias to navigate to locations you have already visited

```bash
z here
z there
```

## Install via Go
```bash
go install github.com/iammatthew2/zoink@latest
zoink setup --quiet  # Setup shell integration

# For completions, add one of these to your shell config:
source <(zoink completion zsh)   # for zsh
source <(zoink completion bash)  # for bash
zoink completion fish | source   # for fish
```

## Usage

### The basics
```bash
# Use the zoink alias (z) to navigate
z foo
```

### Advanced
```bash
# Setup and management
zoink setup [--quiet] [--print-only]  # Interactive setup
zoink stats                           # Show usage statistics and DB info
zoink clean                           # Remove non-existent directories
zoink add /path/to/dir                # Manually add directory
zoink remove /path/to/dir             # Remove directory

# Navigation
# After visiting directories, zoink remembers remembers where you went
cd ~/foo/my-app
cd ~/bar/someThing
cd ~/baz/somePlace

# Quick navigation with frecency ranking (using shell alias)
z foo                      # → ~/foo/my-app (most frequent/recent match)
z bar                      # → ~/bar/someThing
z foo --interactive        # Interactive selection with fzf (requires fzf)
z foo --list               # Lists all tracked directories with visit counts
z --echo foo               # Prints best match path only
z                          # Navigate to previous directory if no query provided
```

## Development

```bash
git clone github.com/iammatthew2/zoink
cd zoink
make build
export PATH="$(pwd)/bin:$PATH"
zoink setup

# Then update your shell for the current instance.
# Do one of:
1. source ~/.zshrc
2. source ~/.config/fish/config.fish
3. source ~/.bashrc 
```

### Architecture/Design Principles
- **Fast**: Minimal startup time, efficient database operations
- **Modern**: Clean UX, intuitive commands
- **Cross-shell**: Seamless integration with bash, zsh, fish
- **Modular**: Clean separation of concerns, testable components

### Roadmap
- [ ] **Package distribution** (Homebrew, etc.)
- [ ] **Bookmarking**
- [ ] **Advanced configuration** options
