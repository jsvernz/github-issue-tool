package parser

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/ef-tech/github-issue-tool/pkg/github"
)

func ParseIssuesFile(filePath string) ([]*github.Issue, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	var issues []*github.Issue
	scanner := bufio.NewScanner(file)
	
	var currentIssue *github.Issue
	var bodyLines []string
	inBody := false
	lineNumber := 0

	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)

		// Check for separator
		if trimmedLine == "---" {
			if currentIssue != nil {
				if len(bodyLines) > 0 {
					currentIssue.Body = strings.TrimSpace(strings.Join(bodyLines, "\n"))
				}
				issues = append(issues, currentIssue)
			}
			currentIssue = nil
			bodyLines = nil
			inBody = false
			continue
		}

		// Parse header line
		if strings.HasPrefix(line, "## ") {
			if currentIssue != nil {
				if len(bodyLines) > 0 {
					currentIssue.Body = strings.TrimSpace(strings.Join(bodyLines, "\n"))
				}
				issues = append(issues, currentIssue)
				bodyLines = nil
			}

			// Extract ID and title
			header := strings.TrimPrefix(line, "## ")
			parts := strings.SplitN(header, "]", 2)
			if len(parts) != 2 {
				return nil, fmt.Errorf("line %d: invalid header format, expected '## [ID] Title'", lineNumber)
			}

			id := strings.TrimSpace(strings.TrimPrefix(parts[0], "["))
			title := strings.TrimSpace(parts[1])

			currentIssue = &github.Issue{
				ID:    id,
				Title: title,
			}
			inBody = false
			continue
		}

		// Parse metadata fields
		if currentIssue != nil && !inBody {
			if strings.HasPrefix(trimmedLine, "Labels:") {
				labelsStr := strings.TrimPrefix(trimmedLine, "Labels:")
				currentIssue.Labels = parseCommaSeparated(labelsStr)
				continue
			}
			if strings.HasPrefix(trimmedLine, "Assignees:") {
				assigneesStr := strings.TrimPrefix(trimmedLine, "Assignees:")
				currentIssue.Assignees = parseCommaSeparated(assigneesStr)
				continue
			}
			if strings.HasPrefix(trimmedLine, "Depends:") {
				dependsStr := strings.TrimPrefix(trimmedLine, "Depends:")
				currentIssue.DependsOn = parseCommaSeparated(dependsStr)
				continue
			}
			if strings.HasPrefix(trimmedLine, "Blocks:") {
				blocksStr := strings.TrimPrefix(trimmedLine, "Blocks:")
				currentIssue.Blocks = parseCommaSeparated(blocksStr)
				continue
			}
			if strings.HasPrefix(trimmedLine, "Related:") {
				relatedStr := strings.TrimPrefix(trimmedLine, "Related:")
				currentIssue.Related = parseCommaSeparated(relatedStr)
				continue
			}

			// Empty line marks the start of body
			if trimmedLine == "" && !inBody {
				inBody = true
				continue
			}
		}

		// Collect body lines
		if currentIssue != nil && inBody {
			bodyLines = append(bodyLines, line)
		}
	}

	// Handle the last issue
	if currentIssue != nil {
		if len(bodyLines) > 0 {
			currentIssue.Body = strings.TrimSpace(strings.Join(bodyLines, "\n"))
		}
		issues = append(issues, currentIssue)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	if len(issues) == 0 {
		return nil, fmt.Errorf("no issues found in file")
	}

	// Validate all IDs are unique
	idMap := make(map[string]bool)
	for _, issue := range issues {
		if idMap[issue.ID] {
			return nil, fmt.Errorf("duplicate issue ID found: %s", issue.ID)
		}
		idMap[issue.ID] = true
	}

	return issues, nil
}

func parseCommaSeparated(input string) []string {
	var result []string
	parts := strings.Split(input, ",")
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}