# Zoink

**Zoink** is a fast, cross-shell tool for directory navigation, inspired by z.sh/fasd with fuzzy/frecency ranking, directory bookmarking and shell snippet management/execution

## Usage
```bash
# Setup and management
zoink setup [--quiet] [--print-only]  # Interactive setup
zoink stats                           # Show usage statistics
zoink clean                           # Remove non-existent directories
zoink add /path/to/dir                # Manually add directory
zoink remove /path/to/dir             # Remove directory

# Navigation (working)
zoink foo                             # Navigate to best match for "foo"
zoink -i docs                         # Interactive selection (coming soon)
zoink -l work                         # List all matches without navigating
zoink --echo project                  # Print best match path only
```

## Shell Integration
- **Automatic shell detection** (bash, zsh, fish)
- **XDG-compliant configuration** placement  
- **Safe shell config editing** with backup protection
- **Development-friendly** alias switching (`z` vs `x`)
- **Automatic directory tracking** via shell hooks

## Development Workflow
```bash
make build                     # Build the binary
make dev-setup                 # Setup shell integration with local binary
make completions               # Generate shell completions
make clean                     # Clean build artifacts
```

## Data storage info

Zoink uses a binary database stored in platform-specific. The location may vary, but should look like:
- **macOS**: `~/Library/Application Support/zoink/zoink.db`
- **Linux**: `~/.config/zoink/zoink.db` 
- **Windows**: `%APPDATA%\zoink\zoink.db`

The database:
- **Persists across rebuilds** and only changes when directories are visited
- **Thread-safe** with atomic saves to prevent corruption
- **Frecency scoring** that balances frequency and recency of visits
- **Manual cleanup** of non-existent directories via `zoink clean`

## Roadmap

### Completed 
- [x] **Binary database** for fast directory storage
- [x] **Frecency algorithm** (frequency + recency scoring)
- [x] **Directory tracking** via shell hooks
- [x] **Basic navigation** with automatic path output
- [x] **Management commands** (stats, clean, add, remove)
- [x] **Cross-platform config** directories
- [x] **Shell integration** setup

### Next Steps
- [ ] **Fuzzy matching** for directory queries (currently substring matching)
- [ ] **Interactive selection** using survey/v2
- [ ] **Import from existing tools** (z.sh, fasd, autojump)
- [ ] **Advanced ranking** options (recent/frequent flags)

### Future Enhancements
- [ ] **Fuzzy completion** for directory shortcuts
- [ ] **Performance optimization** for large databases
- [ ] **Advanced configuration** options
- [ ] **Package distribution** (Homebrew, etc.)
- [ ] **Shell completion** for directory paths

## Architecture

### Design Principles
- **Fast**: Minimal startup time, efficient database operations
- **Modern**: Clean UX, intuitive commands, modern Go practices
- **Cross-shell**: Seamless integration with bash, zsh, fish
- **Modular**: Clean separation of concerns, testable components

### Stack
- **Go**: Core application with Cobra CLI framework
- **Survey/v2**: Interactive UI components
- **JSON**: Minimal configuration format
- **Binary DB**: Fast directory storage (planned)
- **Shell Scripts**: Integration hooks for each shell

## Usage Examples

```bash
# After visiting directories, zoink learns your patterns
cd ~/projects/my-app
cd ~/Documents/work  
cd ~/dev/zoink

# Quick navigation with frecency ranking
zoink proj          # → ~/projects/my-app (most frequent/recent match)
zoink doc           # → ~/Documents/work
zoink zoink         # → ~/dev/zoink

# List and inspect without navigating
zoink -l            # Lists all tracked directories with visit counts
zoink --echo proj   # Prints best match path only

# Management
zoink stats         # Show usage statistics
zoink clean         # Remove deleted directories  
zoink add ~/new/dir # Manually add a directory
zoink remove ~/old  # Remove a directory from tracking
```

## Development

### Building
```bash
git clone <repository>
cd zoink
make deps       # Install dependencies
make build      # Build binary
```

### Testing in Development
```bash
make dev-setup                        # Setup with local binary
export PATH="$(pwd)/bin:$PATH"        # Use local binary  
source ~/.zshrc                       # Activate shell integration

# Test functionality
zoink stats                           # View database statistics
zoink --list                         # List all tracked directories
zoink --echo some_dir                # Test path resolution
```

## Inspiration

Zoink is inspired by the following tools:
- **[z.sh](https://github.com/rupa/z)** - Simple frecency-based navigation
- **[fasd](https://github.com/clvv/fasd)** - Fast access to files and directories
- **[autojump](https://github.com/wting/autojump)** - Directory jumping with learning
