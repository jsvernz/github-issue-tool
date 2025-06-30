package main

import (
	"fmt"
	"os"

	"github.com/ef-tech/github-issue-tool/pkg/cli"
	"github.com/ef-tech/github-issue-tool/pkg/github"
)

func main() {
	opts, err := cli.ParseFlags()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		fmt.Fprintf(os.Stderr, "Use --help for usage information\n")
		os.Exit(1)
	}

	// Detect GitHub environment
	env, err := github.DetectEnvironment()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error detecting environment: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(env)

	// Check if we can proceed
	if env.PreferredMethod == "" {
		fmt.Fprintf(os.Stderr, "\nError: No authentication method available.\n")
		fmt.Fprintf(os.Stderr, "Please do one of the following:\n")
		fmt.Fprintf(os.Stderr, "  1. Install and authenticate GitHub CLI: gh auth login\n")
		fmt.Fprintf(os.Stderr, "  2. Set GITHUB_TOKEN environment variable\n")
		os.Exit(1)
	}

	fmt.Printf("\nLoading issues from file: %s\n", opts.File)
	if opts.DryRun {
		fmt.Println("Running in dry-run mode (no issues will be created)")
	}

	// TODO: Implement the main logic
	fmt.Println("\nTool implementation is in progress...")
}