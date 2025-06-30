package github

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

type CLIClient struct {
	owner string
	repo  string
}

func NewCLIClient(owner, repo string) *CLIClient {
	return &CLIClient{
		owner: owner,
		repo:  repo,
	}
}

func (c *CLIClient) CreateIssue(issue *Issue) error {
	args := []string{"issue", "create"}
	
	// Add repository if specified
	if c.owner != "" && c.repo != "" {
		args = append(args, "--repo", fmt.Sprintf("%s/%s", c.owner, c.repo))
	}
	
	// Add title
	args = append(args, "--title", issue.Title)
	
	// Add body if present
	if issue.Body != "" {
		args = append(args, "--body", issue.Body)
	}
	
	// Add labels if present
	if len(issue.Labels) > 0 {
		for _, label := range issue.Labels {
			args = append(args, "--label", label)
		}
	}
	
	// Add assignees if present
	if len(issue.Assignees) > 0 {
		for _, assignee := range issue.Assignees {
			args = append(args, "--assignee", assignee)
		}
	}
	
	// Execute the command
	cmd := exec.Command("gh", args...)
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to create issue %s: %w", issue.ID, err)
	}
	
	// Parse the issue number from the output
	issueNumber, err := c.parseIssueNumber(string(output))
	if err != nil {
		return fmt.Errorf("failed to parse issue number for %s: %w", issue.ID, err)
	}
	
	issue.Number = issueNumber
	return nil
}

func (c *CLIClient) GetRepository() (owner, name string) {
	return c.owner, c.repo
}

func (c *CLIClient) parseIssueNumber(output string) (int, error) {
	// Try to parse as JSON first (gh cli can output JSON)
	var issueData struct {
		Number int `json:"number"`
	}
	
	if err := json.Unmarshal([]byte(output), &issueData); err == nil {
		return issueData.Number, nil
	}
	
	// Fallback: parse from URL format
	// Output typically looks like: https://github.com/owner/repo/issues/123
	lines := strings.Split(strings.TrimSpace(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "/issues/") {
			parts := strings.Split(line, "/issues/")
			if len(parts) == 2 {
				numberStr := strings.TrimSpace(parts[1])
				var number int
				if _, err := fmt.Sscanf(numberStr, "%d", &number); err == nil {
					return number, nil
				}
			}
		}
	}
	
	return 0, fmt.Errorf("could not parse issue number from output: %s", output)
}

// AddComment adds a comment to an existing issue
func (c *CLIClient) AddComment(issueNumber int, comment string) error {
	args := []string{"issue", "comment", fmt.Sprintf("%d", issueNumber)}
	
	// Add repository if specified
	if c.owner != "" && c.repo != "" {
		args = append(args, "--repo", fmt.Sprintf("%s/%s", c.owner, c.repo))
	}
	
	// Add comment body
	args = append(args, "--body", comment)
	
	cmd := exec.Command("gh", args...)
	_, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to add comment to issue %d: %w", issueNumber, err)
	}
	
	return nil
}

// UpdateIssueWithDependencies updates an issue with dependency information
func (c *CLIClient) UpdateIssueWithDependencies(issue *Issue, issueNumberMap map[string]int) error {
	if len(issue.DependsOn) == 0 && len(issue.Blocks) == 0 && len(issue.Related) == 0 {
		return nil
	}
	
	var dependencyInfo []string
	
	// Add depends on information
	if len(issue.DependsOn) > 0 {
		var deps []string
		for _, depID := range issue.DependsOn {
			if depNumber, exists := issueNumberMap[depID]; exists {
				deps = append(deps, fmt.Sprintf("#%d", depNumber))
			}
		}
		if len(deps) > 0 {
			dependencyInfo = append(dependencyInfo, fmt.Sprintf("**Depends on:** %s", strings.Join(deps, ", ")))
		}
	}
	
	// Add blocks information
	if len(issue.Blocks) > 0 {
		var blocks []string
		for _, blockID := range issue.Blocks {
			if blockNumber, exists := issueNumberMap[blockID]; exists {
				blocks = append(blocks, fmt.Sprintf("#%d", blockNumber))
			}
		}
		if len(blocks) > 0 {
			dependencyInfo = append(dependencyInfo, fmt.Sprintf("**Blocks:** %s", strings.Join(blocks, ", ")))
		}
	}
	
	// Add related information
	if len(issue.Related) > 0 {
		var related []string
		for _, relID := range issue.Related {
			if relNumber, exists := issueNumberMap[relID]; exists {
				related = append(related, fmt.Sprintf("#%d", relNumber))
			}
		}
		if len(related) > 0 {
			dependencyInfo = append(dependencyInfo, fmt.Sprintf("**Related:** %s", strings.Join(related, ", ")))
		}
	}
	
	if len(dependencyInfo) > 0 {
		comment := "## Dependencies\n\n" + strings.Join(dependencyInfo, "\n")
		return c.AddComment(issue.Number, comment)
	}
	
	return nil
}