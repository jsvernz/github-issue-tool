package creator

import (
	"fmt"
	"sort"
	"time"

	"github.com/ef-tech/github-issue-tool/pkg/dependency"
	"github.com/ef-tech/github-issue-tool/pkg/github"
)

type Creator struct {
	client github.Client
	dryRun bool
	noSort bool
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

func NewCreator(client github.Client, dryRun bool, noSort bool) *Creator {
	return &Creator{
		client: client,
		dryRun: dryRun,
		noSort: noSort,
	}
}

func (c *Creator) CreateIssues(issues []*github.Issue) (*Result, error) {
	// Resolve dependencies and get creation order
	resolver := dependency.NewResolver(issues)
	var orderedIssues []*github.Issue
	var err error
	
	if c.noSort {
		orderedIssues, err = resolver.GetOriginalOrder()
		if err != nil {
			return nil, fmt.Errorf("failed to validate issue references: %w", err)
		}
	} else {
		orderedIssues, err = resolver.GetCreationOrder()
		if err != nil {
			return nil, fmt.Errorf("failed to resolve dependencies: %w", err)
		}
	}

	result := &Result{
		CreatedIssues: make([]IssueResult, 0, len(orderedIssues)),
		Errors:        make([]error, 0),
	}

	// Track issue ID to number mapping for dependency comments
	issueNumberMap := make(map[string]int)

	if c.noSort {
		fmt.Printf("Creating %d issues in file order:\n\n", len(orderedIssues))
	} else {
		fmt.Printf("Creating %d issues in dependency order:\n\n", len(orderedIssues))
	}

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
			// Check and create missing labels if supported
			if labelClient, ok := c.client.(github.LabelClient); ok {
				if err := c.ensureLabelsExist(labelClient, issue.Labels); err != nil {
					fmt.Printf("         WARNING: Failed to ensure labels exist: %v\n", err)
				}
			}

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

// ensureLabelsExist checks if labels exist and creates them if they don't
func (c *Creator) ensureLabelsExist(labelClient github.LabelClient, labels []string) error {
	for _, label := range labels {
		exists, err := labelClient.LabelExists(label)
		if err != nil {
			return fmt.Errorf("failed to check if label %s exists: %w", label, err)
		}

		if !exists {
			config, hasDefault := getDefaultLabelConfig(label)
			var description, color string
			
			if hasDefault {
				description = config.description
				color = config.color
			} else {
				// Use generic defaults for unknown labels
				description = fmt.Sprintf("Label: %s", label)
				color = "cccccc"
			}

			fmt.Printf("         Creating missing label: %s\n", label)
			if err := labelClient.CreateLabel(label, description, color); err != nil {
				return fmt.Errorf("failed to create label %s: %w", label, err)
			}
		}
	}

	return nil
}

// CreateLabelsOnly extracts all unique labels from issues and creates them
func (c *Creator) CreateLabelsOnly(issues []*github.Issue) (*Result, error) {
	// Check if the client supports label operations
	labelClient, ok := c.client.(github.LabelClient)
	if !ok {
		return nil, fmt.Errorf("the current GitHub client does not support label operations")
	}

	// Extract all unique labels
	labelSet := make(map[string]bool)
	for _, issue := range issues {
		for _, label := range issue.Labels {
			labelSet[label] = true
		}
	}

	// Convert to slice and sort
	var labels []string
	for label := range labelSet {
		labels = append(labels, label)
	}
	sort.Strings(labels)

	fmt.Printf("Found %d unique labels to process\n\n", len(labels))

	result := &Result{
		CreatedIssues: make([]IssueResult, 0),
		Errors:        make([]error, 0),
	}

	// Process labels
	createdCount := 0
	existingCount := 0

	for i, label := range labels {
		fmt.Printf("[%d/%d] Processing label: %s\n", i+1, len(labels), label)

		if c.dryRun {
			fmt.Printf("         [DRY RUN] Would check/create label: %s\n", label)
			createdCount++
		} else {
			exists, err := labelClient.LabelExists(label)
			if err != nil {
				result.Errors = append(result.Errors, fmt.Errorf("failed to check if label %s exists: %w", label, err))
				fmt.Printf("         ERROR: %v\n", err)
				continue
			}

			if exists {
				fmt.Printf("         Label already exists\n")
				existingCount++
			} else {
				// Get label configuration
				config, hasDefault := getDefaultLabelConfig(label)
				var description, color string
				
				if hasDefault {
					description = config.description
					color = config.color
				} else {
					description = fmt.Sprintf("Label: %s", label)
					color = "cccccc"
				}

				if err := labelClient.CreateLabel(label, description, color); err != nil {
					result.Errors = append(result.Errors, fmt.Errorf("failed to create label %s: %w", label, err))
					fmt.Printf("         ERROR: %v\n", err)
					continue
				}

				fmt.Printf("         Created label with color #%s\n", color)
				createdCount++
			}
		}

		fmt.Println()
	}

	// Summary
	if c.dryRun {
		fmt.Printf("\n✅ Dry Run Summary:\n")
		fmt.Printf("  - Total labels found: %d\n", len(labels))
		fmt.Printf("  - Would create: %d\n", createdCount)
	} else {
		fmt.Printf("\n✅ Summary:\n")
		fmt.Printf("  - Total labels found: %d\n", len(labels))
		fmt.Printf("  - Already existed: %d\n", existingCount)
		fmt.Printf("  - Created: %d\n", createdCount)
		fmt.Printf("  - Errors: %d\n", len(result.Errors))
	}

	return result, nil
}

// getDefaultLabelConfig returns the default configuration for known labels
func getDefaultLabelConfig(label string) (struct{ description, color string }, bool) {
	defaultLabels := map[string]struct {
		description string
		color       string
	}{
		"epic":         {"Epic issue", "d73a4a"},
		"priority-high": {"High priority issue", "b60205"},
		"priority-medium": {"Medium priority issue", "fbca04"},
		"priority-low":  {"Low priority issue", "0e8a16"},
		"setup":        {"Project setup", "7057ff"},
		"foundation":   {"Foundation implementation", "006b75"},
		"config":       {"Configuration implementation", "fef2c0"},
		"provider":     {"Service provider implementation", "c2e0c6"},
		"template":     {"Template implementation", "e99695"},
		"engine":       {"Engine implementation", "f7c6c7"},
		"command":      {"Command implementation", "c5def5"},
		"init":         {"Initialization command", "bfd4f2"},
		"entity":       {"Entity related", "d4c5f9"},
		"generator":    {"Code generator", "fbca04"},
		"feature":      {"New feature", "a2eeef"},
		"bug":          {"Bug fix", "d73a4a"},
		"enhancement":  {"Enhancement", "84b6eb"},
		"documentation": {"Documentation", "0075ca"},
		"testing":      {"Testing related", "d4edda"},
	}

	config, exists := defaultLabels[label]
	return config, exists
}