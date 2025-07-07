# Zoink

**Zoink** is a fast, modern, cross-shell CLI tool for directory navigation, inspired by z.sh/fasd with modern UX, fuzzy/frecency ranking, and robust shell integration

## NOTE

This project is under active development. The core CLI structure and shell integration are complete, but the navigation database and frecency algorithm are still being implemented.


## Available Commands
```bash
# Setup and management
zoink setup [--quiet] [--print-only]  # Interactive setup
zoink stats                           # Show usage statistics
zoink clean                           # Remove non-existent directories
zoink add /path/to/dir                # Manually add directory
zoink remove /path/to/dir             # Remove directory

# Navigation (work in progress)
zoink foo                             # Navigate to best match
zoink -i docs                         # Interactive selection
zoink -l work                         # List matches without navigating
zoink -t recent_dir                   # Prefer recent directories
zoink -f frequent_dir                 # Prefer frequent directories
```

## Shell Integration
- **Automatic shell detection** (bash, zsh, fish)
- **XDG-compliant configuration** placement
- **Safe shell config editing** with backup protection
- **Development-friendly** alias switching (`z` vs `x`)

## Development Workflow
```bash
make build                     # Build the binary
make dev-setup                 # Setup shell integration with local binary
make completions               # Generate shell completions
make clean                     # Clean build artifacts
```

## Planned Features 🎯

### Phase 1: Core Navigation
- [ ] **Binary database** for fast directory storage
- [ ] **Frecency algorithm** (frequency + recency scoring)
- [ ] **Directory tracking** via shell hooks
- [ ] **Fuzzy matching** for directory queries
- [ ] **Basic navigation** with automatic path output

### Phase 2: Advanced Features
- [ ] **Interactive selection** using survey/v2
- [ ] **Fuzzy completion** for directory shortcuts
- [ ] **Import from existing tools** (z.sh, fasd, autojump)
- [ ] **Smart ranking** with usage analytics
- [ ] **Cross-platform testing** and CI

### Phase 3: Polish
- [ ] **Performance optimization** for large databases
- [ ] **Advanced configuration** options
- [ ] **Documentation** and user guides
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

## Usage Examples (Planned)

```bash
# After visiting directories, zoink learns your patterns
cd ~/projects/my-app
cd ~/Documents/work
cd ~/dev/zoink

# Quick navigation with frecency ranking
z proj          # → ~/projects/my-app (most frequent/recent match)
z doc           # → ~/Documents/work
z zoink         # → ~/dev/zoink

# Interactive selection when multiple matches
z -i pro        # Shows menu: ~/projects/my-app, ~/projects/other-app
z -l work       # Lists all work-related directories

# Management
z --stats       # Show usage statistics
z --clean       # Remove deleted directories
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
```

## Inspiration

Zoink is inspired by the following tools:
- **[z.sh](https://github.com/rupa/z)** - Simple frecency-based navigation
- **[fasd](https://github.com/clvv/fasd)** - Fast access to files and directories
- **[autojump](https://github.com/wting/autojump)** - Directory jumping with learning
