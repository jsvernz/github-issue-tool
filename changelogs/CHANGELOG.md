# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.4.0] - 2025-01-01

### Added

- **Label-Only Mode**
  - Added `--label-only` option to create only labels without creating issues
  - Useful for preparing labels before issue creation or standardizing labels across repositories
  - Extracts all unique labels from the issue file and creates them in alphabetical order
  - Supports dry-run mode to preview what labels would be created
  - Shows summary with created, existing, and error counts

### Fixed

- **Test Suite Stability**
  - Fixed failing test in dependency resolver that was checking for circular dependencies via `Blocks` relationships
  - Updated test expectations to match v0.3.1+ behavior where `Blocks` relationships don't create circular dependencies
  - All tests now pass successfully

### Technical Details

- Modified `pkg/cli/flags.go` to add `LabelOnly` boolean field to Options struct
- Implemented `CreateLabelsOnly` method in `pkg/creator/creator.go`
- Enhanced main.go to handle label-only mode separately from issue creation
- Added comprehensive test coverage for the new functionality
- Refactored label configuration to avoid code duplication with `getDefaultLabelConfig` function
- Fixed `pkg/dependency/resolver_test.go` to align with current dependency resolution behavior

## [0.3.1] - 2025-01-01

### Added

- **No-Sort Option**
  - Added `--no-sort` option to create issues in file order without dependency sorting
  - Allows bypassing dependency resolution when file order is preferred
  - Useful for workflows that prioritize input file sequence over dependency constraints
  - Still validates issue references to prevent creation of issues with non-existent dependencies

### Improved

- **Enhanced Dependency Sorting**
  - Simplified dependency resolution to only consider `DependsOn` relationships
  - Removed `Blocks` and `Related` relationships from ordering calculations for cleaner results
  - Improved file order preservation within dependency levels
  - Stable topological sorting that prioritizes original file order as primary criterion
  - Better predictability and intuitive ordering of issues

### Fixed

- **File Order Stability**
  - Fixed issue where INFRA-001~004 and similar sequential issues were separated
  - Enhanced stable sorting algorithm to maintain file order when dependencies allow
  - Reduced unexpected reordering of logically grouped issues
  - Improved user experience with more predictable issue creation order

### Technical Details








- Modified dependency resolver to use position-based stable sorting
- Implemented level-by-level processing to maintain file order precedence
- Added GetOriginalOrder() method for no-sort functionality
- Separated reference validation from dependency cycle detection for no-sort mode

## [0.3.0] - 2025-01-01

### Fixed

- **Issue Creation Order Stability**
  - Fixed non-deterministic issue creation order caused by Go map iteration
  - Issues with no dependencies are now created in the order they appear in the input file
  - Maintains original file order when resolving dependencies, only reordering when necessary
  - Ensures consistent and predictable issue creation order across multiple runs

### Technical Details

- Modified the dependency resolver to preserve original issue order when building the initial queue
- Changed from iterating over maps to iterating over the original issue slice
- When new nodes become available for processing after dependency resolution, they are added in file order
- This fix addresses the issue where independent issues (e.g., INFRA-001) were being created in random positions

## [0.2.0] - 2025-01-01

### Added

- **Automatic Label Creation**
  - Automatically creates missing labels when they don't exist in the target repository
  - Predefined configurations for common labels (epic, priority-high, setup, config, etc.)
  - Assigns appropriate colors and descriptions to known labels
  - Generic fallback for unknown labels with gray color and descriptive names
  - Comprehensive test coverage for label creation functionality

- **Repository Targeting**
  - Added `--repo` option to specify target repository in `owner/name` format
  - Allows creating issues in repositories different from the current working directory
  - Enhanced CLI to support explicit repository specification

- **Enhanced Error Handling**
  - Improved error reporting with detailed output from GitHub CLI
  - Better debugging information when issue creation fails
  - Clear indication when labels are being automatically created

### Fixed

- **Issue Creation Failures**
  - Fixed issue creation failures caused by non-existent labels
  - Enhanced error output to include both error messages and command output
  - Improved debugging capabilities for troubleshooting creation issues

### Technical Improvements

- **Extended GitHub Client Interface**
  - Added `LabelClient` interface with `CreateLabel` and `LabelExists` methods
  - Enhanced CLI client with label management capabilities
  - JSON parsing for label existence verification

- **Creator Enhancement**
  - Enhanced issue creator with automatic label checking and creation
  - Pre-creation label validation and automatic creation workflow
  - Support for label clients in the creation pipeline

- **Test Coverage**
  - Added comprehensive tests for label creation functionality
  - Mock implementations for testing label client behavior
  - JSON parsing tests for label existence verification

## [0.1.2] - 2025-01-01

### Fixed

- **Version Display Issue**
  - Fixed `--version` flag to display the correct version number
  - Updated version constant from "0.1.0" to "0.1.2" in `pkg/config/version.go`
  - Ensures version consistency between code and releases

## [0.1.1] - 2025-01-01

### Fixed

- **Go Install Command Compatibility**
  - Fixed directory structure to support `go install` command properly
  - Moved `cmd/main.go` to `cmd/github-issue-tool/main.go` following Go conventions
  - Updated build paths in Makefile and documentation
  - Ensured `.gitignore` does not exclude the required `cmd/github-issue-tool/` directory
  - Users can now install the tool directly using `go install github.com/ef-tech/github-issue-tool/cmd/github-issue-tool@latest`

### Technical Details

- The issue was caused by the non-standard command structure where `main.go` was placed directly in the `cmd/` directory
- Go's `install` command expects the main package to be in a subdirectory matching the binary name
- This fix ensures compatibility with standard Go tooling and package management

## [0.1.0] - 2025-01-01

### Added

- **Core Functionality**
  - Smart CLI tool for bulk creation of GitHub issues with dependency management
  - Environment auto-detection for GitHub CLI and API availability
  - Git repository auto-detection for seamless integration
  - Bulk issue creation from formatted text files
  - Dependency management supporting `Depends`, `Blocks`, and `Related` relationships
  - Smart creation order optimization based on dependency resolution
  - Dry run mode for previewing actions before execution
  - Progress display with detailed error handling

- **Parser Implementation**
  - Custom issue format parser supporting unique IDs, metadata fields, and markdown content
  - Support for labels, assignees, and dependency relationships
  - Validation for circular dependencies and missing references

- **GitHub Integration**
  - Automatic detection and use of GitHub CLI when available
  - Fallback to GitHub API with token authentication
  - Repository information extraction from Git configuration

- **CLI Features**
  - Command-line flags for file input, dry run mode, and help
  - Version information display
  - Clear error messages and status reporting

- **Documentation**
  - Comprehensive README with usage examples
  - CLAUDE.md for AI-assisted development guidance
  - Example issue files demonstrating proper format

- **Development Tools**
  - Makefile for common development tasks
  - Project structure with clear package separation
  - Configuration management for version tracking

### Technical Details

- Built with Go for performance and reliability
- Modular architecture with separate packages for parsing, GitHub interaction, dependency resolution, and orchestration
- Comprehensive error handling including circular dependency detection
- Support for both GitHub CLI and API authentication methods

[0.4.0]: https://github.com/ef-tech/github-issue-tool/releases/tag/v0.4.0
[0.3.1]: https://github.com/ef-tech/github-issue-tool/releases/tag/v0.3.1
[0.3.0]: https://github.com/ef-tech/github-issue-tool/releases/tag/v0.3.0
[0.2.0]: https://github.com/ef-tech/github-issue-tool/releases/tag/v0.2.0
[0.1.2]: https://github.com/ef-tech/github-issue-tool/releases/tag/v0.1.2
[0.1.1]: https://github.com/ef-tech/github-issue-tool/releases/tag/v0.1.1
[0.1.0]: https://github.com/ef-tech/github-issue-tool/releases/tag/v0.1.0