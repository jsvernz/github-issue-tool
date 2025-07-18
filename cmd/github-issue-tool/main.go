package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/ef-tech/github-issue-tool/pkg/cli"
	"github.com/ef-tech/github-issue-tool/pkg/creator"
	"github.com/ef-tech/github-issue-tool/pkg/github"
	"github.com/ef-tech/github-issue-tool/pkg/parser"
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

	// Parse issues from file
	fmt.Printf("\nLoading issues from file: %s\n", opts.File)
	issues, err := parser.ParseIssuesFile(opts.File)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing issues file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Loaded %d issues from file\n", len(issues))

	// Determine repository owner and name
	var owner, name string
	if opts.Repository != "" {
		// Use explicitly specified repository
		parts := strings.Split(opts.Repository, "/")
		if len(parts) != 2 {
			fmt.Fprintf(os.Stderr, "Error: Repository must be in 'owner/name' format\n")
			os.Exit(1)
		}
		owner, name = parts[0], parts[1]
	} else {
		// Use detected repository from environment
		owner, name = env.RepositoryOwner, env.RepositoryName
	}

	// Create GitHub client
	var client github.Client
	if env.PreferredMethod == "cli" {
		client = github.NewCLIClient(owner, name)
	} else {
		client = github.NewAPIClient(owner, name, "")
	}

	// Create issue creator
	issueCreator := creator.NewCreator(client, opts.DryRun, opts.NoSort)

	if opts.LabelOnly {
		// Label-only mode
		if opts.DryRun {
			fmt.Println("\n🔍 Running in dry-run mode (no labels will be created)")
		} else {
			fmt.Println("\n🏷️  Creating labels only...")
		}

		result, err := issueCreator.CreateLabelsOnly(issues)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating labels: %v\n", err)
			os.Exit(1)
		}

		if len(result.Errors) > 0 {
			fmt.Printf("\n❌ Errors:\n")
			for _, err := range result.Errors {
				fmt.Printf("  - %v\n", err)
			}
			os.Exit(1)
		}
	} else {
		// Normal issue creation mode
		if opts.DryRun {
			fmt.Println("\n🔍 Running in dry-run mode (no issues will be created)")
		} else {
			fmt.Println("\n🚀 Creating issues...")
		}
		
		if opts.NoSort {
			fmt.Println("📝 Creating issues in file order (dependency sorting disabled)")
		}

		// Create issues
		result, err := issueCreator.CreateIssues(issues)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating issues: %v\n", err)
			os.Exit(1)
		}

		// Print results
		fmt.Printf("\n✅ Summary:\n")
		fmt.Printf("  - Total issues processed: %d\n", len(issues))
		fmt.Printf("  - Successfully created: %d\n", len(result.CreatedIssues))
		fmt.Printf("  - Errors: %d\n", len(result.Errors))

		if len(result.CreatedIssues) > 0 {
			fmt.Printf("\n📋 Created Issues:\n")
			for _, issue := range result.CreatedIssues {
				if opts.DryRun {
					fmt.Printf("  - [%s] #%d: %s (DRY RUN)\n", issue.ID, issue.Number, issue.Title)
				} else {
					fmt.Printf("  - [%s] #%d: %s - %s\n", issue.ID, issue.Number, issue.Title, issue.URL)
				}
			}
		}

		if len(result.Errors) > 0 {
			fmt.Printf("\n❌ Errors:\n")
			for _, err := range result.Errors {
				fmt.Printf("  - %v\n", err)
			}
			os.Exit(1)
		}
	}
}