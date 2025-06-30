package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
)

type APIClient struct {
	token string
	owner string
	repo  string
}

type APIIssueRequest struct {
	Title     string   `json:"title"`
	Body      string   `json:"body,omitempty"`
	Labels    []string `json:"labels,omitempty"`
	Assignees []string `json:"assignees,omitempty"`
}

type APIIssueResponse struct {
	Number int    `json:"number"`
	Title  string `json:"title"`
	URL    string `json:"html_url"`
}

type APICommentRequest struct {
	Body string `json:"body"`
}

func NewAPIClient(owner, repo, token string) *APIClient {
	if token == "" {
		token = os.Getenv("GITHUB_TOKEN")
	}
	return &APIClient{
		token: token,
		owner: owner,
		repo:  repo,
	}
}

func (c *APIClient) CreateIssue(issue *Issue) error {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues", c.owner, c.repo)
	
	reqBody := APIIssueRequest{
		Title:     issue.Title,
		Body:      issue.Body,
		Labels:    issue.Labels,
		Assignees: issue.Assignees,
	}
	
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal issue data for %s: %w", issue.ID, err)
	}
	
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request for issue %s: %w", issue.ID, err)
	}
	
	req.Header.Set("Authorization", "token "+c.token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request for issue %s: %w", issue.ID, err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("failed to create issue %s: HTTP %d", issue.ID, resp.StatusCode)
	}
	
	var respData APIIssueResponse
	if err := json.NewDecoder(resp.Body).Decode(&respData); err != nil {
		return fmt.Errorf("failed to decode response for issue %s: %w", issue.ID, err)
	}
	
	issue.Number = respData.Number
	return nil
}

func (c *APIClient) GetRepository() (owner, name string) {
	return c.owner, c.repo
}

func (c *APIClient) AddComment(issueNumber int, comment string) error {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues/%d/comments", c.owner, c.repo, issueNumber)
	
	reqBody := APICommentRequest{
		Body: comment,
	}
	
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal comment data: %w", err)
	}
	
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create comment request: %w", err)
	}
	
	req.Header.Set("Authorization", "token "+c.token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send comment request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("failed to create comment: HTTP %d", resp.StatusCode)
	}
	
	return nil
}

func (c *APIClient) UpdateIssueWithDependencies(issue *Issue, issueNumberMap map[string]int) error {
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