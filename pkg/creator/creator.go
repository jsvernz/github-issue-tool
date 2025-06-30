package creator

import (
	"fmt"
	"time"

	"github.com/ef-tech/github-issue-tool/pkg/dependency"
	"github.com/ef-tech/github-issue-tool/pkg/github"
)

type Creator struct {
	client github.Client
	dryRun bool
}

type Result struct {
	CreatedIssues []IssueResult
	Errors        []error
}

type IssueResult struct {
	ID     string
	Title  string
	Number int
	URL    string
}

func NewCreator(client github.Client, dryRun bool) *Creator {
	return &Creator{
		client: client,
		dryRun: dryRun,
	}
}

func (c *Creator) CreateIssues(issues []*github.Issue) (*Result, error) {
	// Resolve dependencies and get creation order
	resolver := dependency.NewResolver(issues)
	orderedIssues, err := resolver.GetCreationOrder()
	if err != nil {
		return nil, fmt.Errorf("failed to resolve dependencies: %w", err)
	}

	result := &Result{
		CreatedIssues: make([]IssueResult, 0, len(orderedIssues)),
		Errors:        make([]error, 0),
	}

	// Track issue ID to number mapping for dependency comments
	issueNumberMap := make(map[string]int)

	fmt.Printf("Creating %d issues in dependency order:\n\n", len(orderedIssues))

	for i, issue := range orderedIssues {
		fmt.Printf("[%d/%d] Creating issue: %s - %s\n", i+1, len(orderedIssues), issue.ID, issue.Title)

		// Show dependency info
		depInfo := resolver.GetDependencyInfo(issue)
		if depInfo != "" {
			fmt.Printf("         Dependencies: %s\n", depInfo)
		}

		if c.dryRun {
			fmt.Printf("         [DRY RUN] Would create issue with:\n")
			fmt.Printf("         - Title: %s\n", issue.Title)
			if issue.Body != "" {
				fmt.Printf("         - Body: %s\n", truncateString(issue.Body, 100))
			}
			if len(issue.Labels) > 0 {
				fmt.Printf("         - Labels: %v\n", issue.Labels)
			}
			if len(issue.Assignees) > 0 {
				fmt.Printf("         - Assignees: %v\n", issue.Assignees)
			}
			
			// Simulate issue number for dry run
			issue.Number = i + 1000
			issueNumberMap[issue.ID] = issue.Number
			
			result.CreatedIssues = append(result.CreatedIssues, IssueResult{
				ID:     issue.ID,
				Title:  issue.Title,
				Number: issue.Number,
				URL:    fmt.Sprintf("https://github.com/owner/repo/issues/%d", issue.Number),
			})
		} else {
			// Create the actual issue
			if err := c.client.CreateIssue(issue); err != nil {
				result.Errors = append(result.Errors, fmt.Errorf("failed to create issue %s: %w", issue.ID, err))
				fmt.Printf("         ERROR: %v\n", err)
				continue
			}

			issueNumberMap[issue.ID] = issue.Number
			owner, repo := c.client.GetRepository()
			url := fmt.Sprintf("https://github.com/%s/%s/issues/%d", owner, repo, issue.Number)

			fmt.Printf("         Created: #%d - %s\n", issue.Number, url)

			result.CreatedIssues = append(result.CreatedIssues, IssueResult{
				ID:     issue.ID,
				Title:  issue.Title,
				Number: issue.Number,
				URL:    url,
			})

			// Add a small delay to avoid rate limiting
			time.Sleep(100 * time.Millisecond)
		}

		fmt.Println()
	}

	// Add dependency comments in a second pass
	if !c.dryRun {
		fmt.Println("Adding dependency information to issues...")
		for _, issue := range orderedIssues {
			if hasClientWithDependencyUpdate(c.client) {
				if err := updateIssueWithDependencies(c.client, issue, issueNumberMap); err != nil {
					result.Errors = append(result.Errors, fmt.Errorf("failed to update dependencies for issue %s: %w", issue.ID, err))
				}
			}
		}
	}

	return result, nil
}

func hasClientWithDependencyUpdate(client github.Client) bool {
	_, ok := client.(github.DependencyClient)
	return ok
}

func updateIssueWithDependencies(client github.Client, issue *github.Issue, issueNumberMap map[string]int) error {
	if depClient, ok := client.(github.DependencyClient); ok {
		return depClient.UpdateIssueWithDependencies(issue, issueNumberMap)
	}
	return fmt.Errorf("client does not support dependency updates")
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}