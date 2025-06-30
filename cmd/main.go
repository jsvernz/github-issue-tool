package main

import (
	"fmt"
	"os"

	"github.com/ef-tech/github-issue-tool/pkg/cli"
)

func main() {
	opts, err := cli.ParseFlags()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		fmt.Fprintf(os.Stderr, "Use --help for usage information\n")
		os.Exit(1)
	}

	fmt.Printf("Loading issues from file: %s\n", opts.File)
	if opts.DryRun {
		fmt.Println("Running in dry-run mode (no issues will be created)")
	}

	// TODO: Implement the main logic
	fmt.Println("Tool implementation is in progress...")
}