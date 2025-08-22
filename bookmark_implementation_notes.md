# Bookmark Implementation Project Notes

## Project Goal
Implement bookmarking functionality in Zoink using the interface:
```bash
z --bookmark myproject  # Add bookmark to current directory
z -b myproject         # Short form for bookmarking
```

## Core Challenge
Make the `z` shell alias work seamlessly with Cobra command parsing, specifically handling the `--bookmark` flag.

## Current State Analysis

### Shell Integration Constraints
- `z` is a shell function that processes arguments before calling `zoink`
- Current `z()` function handles `-i` and `--interactive` flags specially
- Arguments are processed in shell before being passed to `zoink find`
- Need to extend this to handle `--bookmark` flag

### Cobra Integration Points
- `zoink` executable uses Cobra for command parsing
- `find` command is the main navigation command
- Need to add bookmark functionality to the `find` command
- Must maintain backward compatibility with existing `z` usage

## Implementation Plan

### Phase 1: Extend Cobra find Command
- Add `--bookmark` flag to the `findCmd` in `cmd/find.go`
- Handle bookmark creation when flag is present
- Maintain existing navigation behavior when flag is not present

### Phase 2: Update Shell Integration
- Modify `z()` shell function to detect `--bookmark` flag
- Pass appropriate arguments to `zoink find --bookmark <name>`
- Handle edge cases and error conditions

### Phase 3: Database Integration
- Extend `DirectoryEntry` struct with `BookmarkName` field
- Add bookmark database methods
- Update database save/load for backward compatibility

### Phase 4: Additional Bookmark Commands
- `z --bookmarks` - List all bookmarks
- `z --goto <bookmark>` - Navigate to bookmark
- `z --unbookmark <name>` - Remove bookmark

## Technical Challenges

### 1. Shell Function Complexity
Current `z()` function pattern:
```bash
case "$*" in
    *-i*|*--interactive*)
        # Interactive handling
        ;;
    *)
        # Regular navigation
        ;;
esac
```

Need to add:
```bash
case "$*" in
    *--bookmark*)
        # Bookmark handling
        ;;
    *-i*|*--interactive*)
        # Interactive handling
        ;;
    *)
        # Regular navigation
        ;;
esac
```

### 2. Cobra Flag Integration
- Add `--bookmark` flag to `findCmd`
- Handle flag parsing in `executeFind` function
- Distinguish between navigation and bookmark operations

### 3. Database Schema Changes
- Add `BookmarkName string` to `DirectoryEntry`
- Update binary database format
- Maintain backward compatibility

### 4. Shell Function Edge Cases
- **Multiple flags**: `z --bookmark myproject --interactive` (should error)
- **No bookmark name**: `z --bookmark` (should show usage)
- **Bookmark name with spaces**: `z --bookmark "my project"` (should work)
- **Special characters**: `z --bookmark my-project_123` (should work)

### 5. Cobra Integration Challenges
- **Flag parsing**: Cobra needs to parse `--bookmark myproject` correctly
- **Argument handling**: Distinguish between bookmark name and navigation query
- **Help integration**: Update help text to show bookmark functionality
- **Validation**: Validate bookmark names in Cobra layer

## Key Design Decisions

### 1. Flag Processing Order
**Decision**: `--bookmark` should take precedence over navigation
**Rationale**: When user explicitly wants to bookmark, that's the primary action
**Implementation**: Check for `--bookmark` flag first in `executeFind()`

### 2. Bookmark Name Extraction
**Challenge**: How to extract bookmark name from `z --bookmark myproject`
**Options**:
- Option A: `--bookmark` takes exactly one argument (bookmark name)
- Option B: `--bookmark` takes remaining arguments as bookmark name
- Option C: Use `--bookmark-name` for explicit naming

**Decision**: Option A - `--bookmark` takes exactly one argument
**Rationale**: Clear and explicit, follows standard flag patterns

### 3. Error Handling Strategy
**Bookmark Name Conflicts**: Return error if bookmark name already exists
**Missing Directory**: Return error if current directory not in database
**Invalid Names**: Validate bookmark names (no spaces, special chars)

### 4. User Feedback
**Success**: `Bookmarked '/path/to/dir' as 'myproject'`
**Error**: `Error: bookmark name 'myproject' is already used by '/other/path'`
**Usage**: Show help when `--bookmark` used incorrectly

### 5. Shell Function Complexity
**Current Pattern**: `case "$*" in *--bookmark*) ;; esac`
**Challenge**: Need to extract bookmark name from arguments
**Solution**: Use `sed` or similar to extract bookmark name after `--bookmark`

## Implementation Details

### Cobra Integration Plan (Simplified)
```go
// In cmd/management.go
var bookmarkCmd = &cobra.Command{
    Use:   "bookmark [name]",
    Short: "Add bookmark to current directory",
    Long:  `Add a bookmark to the current directory for quick navigation.`,
    Args:  cobra.ArbitraryArgs,
    Run: func(cmd *cobra.Command, args []string) {
        handleBookmarkArgs(args)  // Handles flag filtering and validation
    },
}

// handleBookmarkArgs processes bookmark command arguments
func handleBookmarkArgs(args []string) {
    // Filter out flags and get the bookmark name
    var bookmarkName string
    for _, arg := range args {
        if arg != "-b" && arg != "--bookmark" {
            bookmarkName = arg
            break
        }
    }

    if bookmarkName == "" {
        fmt.Fprintf(os.Stderr, "Error: bookmark requires a name\n")
        fmt.Fprintf(os.Stderr, "Usage: zoink bookmark <name>\n")
        os.Exit(1)
    }

    handleBookmark(bookmarkName)
}
```

### Shell Function Plan (Simplified)
```bash
case "$*" in
    *-b*|*--bookmark*)
        # Bookmark mode - let Cobra handle validation
        zoink bookmark "$@"
        ;;
    # ... existing cases
esac
```

### Database Integration Plan
- Add `BookmarkName string` to `DirectoryEntry`
- Update `writeEntry()` and `readEntry()` to handle bookmark data
- Add `AddBookmark()`, `RemoveBookmark()`, `GetBookmark()` methods
- Maintain backward compatibility for existing databases

## User Experience Flow

### Bookmark Creation Flow
```bash
# User navigates to directory
cd ~/projects/my-complex-project-name

# User creates bookmark (long form)
z --bookmark myproject

# Or short form
z -b myproject

# Expected output
Bookmarked '/Users/user/projects/my-complex-project-name' as 'myproject'
```

### Error Handling Flow
```bash
# Bookmark name conflict
z --bookmark myproject
Error: bookmark name 'myproject' is already used by '/other/path'

# No bookmark name provided
z --bookmark
Error: -b/--bookmark requires a name

# Directory not in database
z --bookmark myproject
Error: current directory not found in database. Visit some directories first.
```

### Navigation Flow (Future)
```bash
# Navigate to bookmark (Phase 2)
z --goto myproject
# Changes to bookmarked directory

# List bookmarks (Phase 2)
z --bookmarks
# Shows all bookmarks with details
```

## Success Criteria

- [ ] `z --bookmark myproject` works seamlessly
- [ ] Existing `z` navigation continues to work
- [ ] Shell function remains maintainable
- [ ] Database changes are backward compatible
- [ ] User experience is intuitive
- [ ] Error messages are clear and helpful
- [ ] Bookmark names are validated appropriately

## Current Structure Analysis

### findCmd Structure (`cmd/find.go`)
- Uses `cobra.ArbitraryArgs` to accept any number of arguments
- Has flags: `--interactive`, `--list`, `--echo`, `--recent`, `--frequent`
- Main logic in `executeFind()` function
- Calls `handleNavigation()` for core functionality

### Shell Integration Pattern (`internal/shell/integration.go`)
Current `z()` function structure:
```bash
z() {
    if [ $# -eq 0 ]; then
        # No arguments: let zoink handle the empty case
        result=$(zoink find)
        [ -n "$result" ] && [ -d "$result" ] && cd "$result"
    else
        # Check if interactive flag is present
        case "$*" in
            *-i*|*--interactive*)
                # Interactive mode handling
                ;;
            *)
                # Non-interactive mode
                result=$(zoink find "$@")
                if [ $? -eq 0 ] && [ -n "$result" ] && [ -d "$result" ]; then
                    cd "$result"
                else
                    echo "$result"
                fi
                ;;
        esac
    fi
}
```

### Key Insights
1. **Shell Function Pattern**: Uses `case "$*"` to detect flags in the entire argument string
2. **Flag Processing**: Strips flags and passes remaining args to `zoink find`
3. **Error Handling**: Checks exit code and directory existence before `cd`
4. **Fish Compatibility**: Similar pattern but with fish-specific syntax

## Implementation Strategy

### Option A: Extend findCmd with Bookmark Flag
- Add `--bookmark` flag to `findCmd.Flags()`
- Modify `executeFind()` to detect bookmark mode
- Create `handleBookmark()` function for bookmark logic

### Option B: Create Separate Bookmark Command
- Create new `bookmarkCmd` in `cmd/management.go`
- Add to `rootCmd.AddCommand()` in management.go init()
- Keep `findCmd` focused on navigation only

### Option C: Create Bookmark Subcommands
- Create `bookmarkCmd` with subcommands: `add`, `list`, `goto`, `remove`
- More explicit and extensible

## Analysis: Separate Command vs findCmd Extension

### Problems with Extending findCmd
1. **Mixed Responsibilities**: `findCmd` becomes both navigation AND bookmark management
2. **Complex Flag Logic**: Need to distinguish between navigation and bookmark operations
3. **Help Text Confusion**: Help becomes cluttered with both navigation and bookmark options
4. **Testing Complexity**: Harder to test navigation and bookmark logic separately
5. **Future Extensibility**: Adding more bookmark operations makes `findCmd` even more complex

### Benefits of Separate Command
1. **Single Responsibility**: Each command has one clear purpose
2. **Cleaner Help**: `zoink bookmark --help` shows only bookmark operations
3. **Easier Testing**: Can test bookmark logic independently
4. **Better Organization**: Follows existing pattern (stats, clean, add, remove)
5. **Future Extensibility**: Easy to add more bookmark operations

### Shell Integration Implications
**Current Shell Pattern**: `z` calls `zoink find`
**With Separate Command**: `z --bookmark` would call `zoink bookmark`

**Shell Function Changes Needed**:
```bash
case "$*" in
    *--bookmark*)
        # Extract bookmark name and call separate command
        local bookmark_name=$(echo "$@" | sed 's/.*--bookmark //')
        if [ -n "$bookmark_name" ]; then
            zoink bookmark "$bookmark_name"
        else
            echo "Error: --bookmark requires a name" >&2
            return 1
        fi
        ;;
    # ... existing cases
esac
```

### Recommendation: Use Separate Command
**Decision**: Create `bookmarkCmd` in `cmd/management.go`
**Rationale**: 
- Follows existing command organization pattern
- Keeps `findCmd` focused on navigation
- Easier to maintain and extend
- Better separation of concerns

### Phase 3: Database Integration
- Extend `DirectoryEntry` with `BookmarkName` field
- Add bookmark database methods
- Update database save/load

## Next Steps

1. **Create Bookmark Command**: Add `bookmarkCmd` to `cmd/management.go`
2. **Plan Shell Function Changes**: Design the `*--bookmark*` case logic to call `zoink bookmark`
3. **Database Schema Design**: Plan the `DirectoryEntry` extension
4. **Error Handling Strategy**: Plan how to handle conflicts and errors
5. **Future Commands**: Plan `bookmarks`, `goto`, `unbookmark` commands
