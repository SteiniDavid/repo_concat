# Repo Concat

A terminal utility that clones a GitHub repository and concatenates all its text files into a single file with token estimation.

## Features

- Clone any public GitHub repository
- **Smart caching**: Repositories are cached for 5 minutes to speed up repeated runs
- Concatenate all text files with file headers showing paths
- Exclude files using regex patterns or path patterns
- Preview repository structure before processing (peek mode)
- Estimate token count for the resulting text
- Copy output to clipboard automatically
- Generate timestamped output filenames

## Usage

```bash
# Basic usage
./repo-concat -url https://github.com/user/repo

# With peek mode to preview structure and see dry run
./repo-concat -url https://github.com/user/repo -peek

# Peek mode with filters shows exactly what would be processed
./repo-concat -url https://github.com/user/repo -include ".*\.py$" -peek

# Exclude specific patterns
./repo-concat -url https://github.com/user/repo -exclude ".*\.test\..*" -exclude "vendor/"

# Include only specific patterns (strict mode)
./repo-concat -url https://github.com/user/repo -include ".*\.go$" -include ".*\.js$"

# Path-based filtering: include only utils directory and its children
./repo-concat -url https://github.com/user/repo -include "/utils"

# Path-based filtering: include multiple top-level directories
./repo-concat -url https://github.com/user/repo -include "/src" -include "/lib"

# Combine include and exclude patterns
./repo-concat -url https://github.com/user/repo -include ".*\.py$" -exclude ".*test.*"

# Custom output directory
./repo-concat -url https://github.com/user/repo -output /path/to/output

# Force fresh clone, ignore cache
./repo-concat -url https://github.com/user/repo -no-cache
```

## Flags

- `-url`: GitHub repository URL (required)
- `-peek`: Show folder structure and dry run of file filtering before processing
- `-exclude`: Regex patterns or path patterns to exclude files (can be used multiple times)
- `-include`: Regex patterns or path patterns to include files (if specified, only matching files are included)
- `-output`: Output directory for concatenated file (default: current directory)
- `-tokens`: Estimate token count (default: true)
- `-no-cache`: Force fresh clone, ignore cache

## Pattern Types

### Regex Patterns
Regular regex patterns for flexible file matching:
- `".*\.go$"` - All Go files
- `".*test.*"` - Files containing "test"
- `"src/.*\.js$"` - JavaScript files in src directory

### Path Patterns
Path-based patterns starting with `/` for directory filtering:
- `"/utils"` - Include only the top-level `utils` directory and all its children
- `"/src"` - Include only the top-level `src` directory and all its children
- `"/lib"` - Include only the top-level `lib` directory and all its children

Path patterns match directories by prefix, so `/util` will match `utils/`, `utilities/`, etc.

## Include vs Exclude Patterns

### Include Patterns (Strict Mode)
When `-include` patterns are specified, **only** files matching at least one include pattern will be processed. This creates a strict whitelist mode.

### Exclude Patterns
Files matching exclude patterns are always skipped, regardless of include patterns.

### Processing Order
1. Files are first checked against exclusion patterns (including defaults)
2. If include patterns are specified, files must match at least one include pattern
3. Files must pass the text file detection

### Peek Mode (Dry Run)
When using `-peek`, the utility shows:
- **Smart directory view**: If filters are applied, only shows directories containing matching files
- **Dry run results**: exactly which files would be included/excluded
- File counts and summary
- Confirmation prompt before proceeding

This is especially useful when testing include/exclude patterns to see their effects before processing. The filtered directory view helps you understand exactly which parts of the repository will be processed.

## Caching

The utility automatically caches cloned repositories for 5 minutes to improve performance when:
- Testing different include/exclude patterns on the same repository
- Running the utility multiple times on the same repository
- Using peek mode to preview before processing

**Cache location**: `~/.cache/repo-concat/`
**Cache duration**: 5 minutes from first clone
**Automatic cleanup**: Expired caches are automatically removed

The utility will show cache status with age information:
- `"Using cached repository (cached 2 minutes ago): https://github.com/user/repo"` when cache is found
- `"Cloning repository (cache ignored): https://github.com/user/repo"` when using `--no-cache`
- `"Cloning repository: https://github.com/user/repo"` when downloading fresh (no cache available)

## Default Exclusions

The utility automatically excludes:
- Git files (`.git/`, `.gitignore`)
- System files (`.DS_Store`)
- Dependencies (`node_modules/`)
- Environment files (`.env`)
- Binary files (images, videos, audio, archives, documents)

## Output Format

Files are concatenated with headers showing:
- File path relative to repository root
- Full file path
- File content wrapped in markdown code blocks

## Requirements

- Go 1.21 or later
- Git (for cloning repositories)
- Clipboard utility (`pbcopy` on macOS, `xclip` or `xsel` on Linux)

## Build

```bash
go build -o repo-concat .
```