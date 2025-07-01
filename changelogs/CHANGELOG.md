# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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

[0.1.0]: https://github.com/ef-tech/github-issue-tool/releases/tag/v0.1.0