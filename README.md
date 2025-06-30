# GitHub Issue Tool

A smart CLI tool for bulk creation of GitHub issues with dependency management.

## Features

- **Environment Auto-Detection**: Automatically detects GitHub CLI or API availability
- **Git Repository Auto-Detection**: Automatically detects Git repository and GitHub information
- **Bulk Issue Creation**: Load multiple issues from a text file
- **Dependency Management**: Supports `Depends`, `Blocks`, and `Related` relationships between issues
- **Smart Creation Order**: Automatically optimizes creation order based on dependencies
- **Dry Run Mode**: Preview what will be created before actually creating issues
- **Progress Display**: Shows detailed progress and error handling

## Installation

### From Source

```bash
git clone https://github.com/ef-tech/github-issue-tool.git
cd github-issue-tool
go build -o github-issue-tool cmd/main.go
```

### Using Go Install

```bash
go install github.com/ef-tech/github-issue-tool/cmd@latest
```

## Usage

### Basic Usage

```bash
# Create issues from a file
github-issue-tool --file issues.txt

# Dry run to preview what will be created
github-issue-tool --file issues.txt --dry-run

# Using short form options
github-issue-tool -f issues.txt -n
```

### Command Line Options

- `--file, -f`: Path to the issues file (required)
- `--dry-run, -n`: Perform a dry run without creating issues
- `--help, -h`: Show help message
- `--version, -v`: Show version information

## Issue File Format

The tool supports a simple text format for defining issues:

```
## [UNIQUE-ID] Issue Title
Labels: label1, label2, label3
Assignees: username1, username2
Depends: OTHER-ID1, OTHER-ID2
Blocks: FUTURE-ID1, FUTURE-ID2
Related: RELATED-ID1, RELATED-ID2

Issue body content.
Multiple lines are supported.

You can include markdown and other formatting.
---
## [ANOTHER-ID] Another Issue Title
Labels: bug

Another issue body.
```

### Format Specifications

- **Header**: `## [UNIQUE-ID] Issue Title` - Each issue starts with this format
- **Separator**: `---` - Optional separator between issues
- **Metadata Fields**:
  - `Labels:` - Comma-separated list of labels
  - `Assignees:` - Comma-separated list of GitHub usernames
  - `Depends:` - IDs of issues this depends on
  - `Blocks:` - IDs of issues this blocks
  - `Related:` - IDs of related issues
- **Body**: Everything after the metadata fields is treated as the issue body

## Environment Setup

### GitHub CLI (Recommended)

1. Install GitHub CLI: https://cli.github.com/
2. Authenticate with GitHub:
   ```bash
   gh auth login
   ```

### GitHub API Token

1. Create a personal access token: https://github.com/settings/tokens
2. Set the environment variable:
   ```bash
   export GITHUB_TOKEN=your_token_here
   ```

## Execution Patterns

### Inside Git Repository + GitHub CLI

Complete automation - the tool will:
- Detect the current repository
- Use GitHub CLI for authentication
- Create issues automatically

### Inside Git Repository + API Token

Semi-automatic - the tool will:
- Detect the current repository
- Use the API token for authentication
- Create issues automatically

### Outside Git Repository

Interactive mode - you'll need to specify the repository manually in the future (currently not implemented).

## Examples

### Sample Issue File

See `examples/sample_issues.txt` for a complete example with dependencies.

### Dependency Resolution

The tool automatically resolves dependencies and creates issues in the correct order:

```
SETUP-001 (no dependencies) → Created first
AUTH-001 (depends on SETUP-001) → Created second
API-001 (depends on AUTH-001) → Created third
FRONTEND-001 (depends on API-001) → Created fourth
```

### Dry Run Output

```bash
$ github-issue-tool --file examples/sample_issues.txt --dry-run

GitHub Environment Detection:
  GitHub CLI available: true
  GitHub CLI authenticated: true
  In Git repository: true
  Repository: ef-tech/github-issue-tool
  Preferred method: cli

Loading issues from file: examples/sample_issues.txt
Loaded 5 issues from file

🔍 Running in dry-run mode (no issues will be created)
Creating 5 issues in dependency order:

[1/5] Creating issue: SETUP-001 - Project Setup and Configuration
         [DRY RUN] Would create issue with:
         - Title: Project Setup and Configuration
         - Labels: [setup documentation]
         - Assignees: [developer]

✅ Summary:
  - Total issues processed: 5
  - Successfully created: 5
  - Errors: 0
```

## Error Handling

The tool provides comprehensive error handling for:

- **Circular Dependencies**: Detects and reports circular dependency loops
- **Missing References**: Validates that all referenced issue IDs exist
- **Duplicate IDs**: Ensures all issue IDs are unique
- **Authentication Issues**: Clear error messages for authentication problems
- **API Rate Limits**: Includes delays to avoid rate limiting

## Development

### Project Structure

```
github-issue-tool/
├── cmd/
│   └── main.go          # Entry point
├── pkg/
│   ├── parser/          # Issue format parser
│   ├── github/          # GitHub API/CLI wrapper
│   ├── dependency/      # Dependency resolution logic
│   ├── creator/         # Issue creation orchestration
│   ├── config/          # Configuration and version
│   └── cli/             # CLI flag parsing
├── examples/            # Usage examples
└── docs/               # Additional documentation
```

### Building

```bash
go build -o github-issue-tool cmd/main.go
```

### Testing

```bash
go test ./...
go test -v ./...  # Verbose output
go test ./pkg/... # Test specific packages
```

### Linting

```bash
golangci-lint run
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Built with Go and the GitHub CLI/API
- Inspired by the need for efficient bulk issue management
- Designed for developer productivity and workflow automation