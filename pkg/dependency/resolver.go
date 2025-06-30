package dependency

import (
	"fmt"

	"github.com/ef-tech/github-issue-tool/pkg/github"
)

// Resolver handles dependency resolution and ordering of issues
type Resolver struct {
	issues   []*github.Issue
	idToIssue map[string]*github.Issue
}

// NewResolver creates a new dependency resolver
func NewResolver(issues []*github.Issue) *Resolver {
	idToIssue := make(map[string]*github.Issue)
	for _, issue := range issues {
		idToIssue[issue.ID] = issue
	}
	return &Resolver{
		issues:   issues,
		idToIssue: idToIssue,
	}
}

// Validate checks for circular dependencies and missing references
func (r *Resolver) Validate() error {
	// Check for missing dependencies
	for _, issue := range r.issues {
		for _, depID := range issue.DependsOn {
			if _, exists := r.idToIssue[depID]; !exists {
				return fmt.Errorf("issue %s depends on non-existent issue %s", issue.ID, depID)
			}
		}
		for _, blockID := range issue.Blocks {
			if _, exists := r.idToIssue[blockID]; !exists {
				return fmt.Errorf("issue %s blocks non-existent issue %s", issue.ID, blockID)
			}
		}
		for _, relID := range issue.Related {
			if _, exists := r.idToIssue[relID]; !exists {
				return fmt.Errorf("issue %s relates to non-existent issue %s", issue.ID, relID)
			}
		}
	}

	// Check for circular dependencies
	visited := make(map[string]bool)
	recStack := make(map[string]bool)
	
	for _, issue := range r.issues {
		if err := r.detectCycle(issue.ID, visited, recStack); err != nil {
			return err
		}
	}

	return nil
}

// detectCycle uses DFS to detect circular dependencies
func (r *Resolver) detectCycle(id string, visited, recStack map[string]bool) error {
	visited[id] = true
	recStack[id] = true

	issue := r.idToIssue[id]
	for _, depID := range issue.DependsOn {
		if !visited[depID] {
			if err := r.detectCycle(depID, visited, recStack); err != nil {
				return err
			}
		} else if recStack[depID] {
			return fmt.Errorf("circular dependency detected: %s -> %s", id, depID)
		}
	}

	// Also check blocks relationships (reverse dependency)
	for _, otherIssue := range r.issues {
		for _, blockID := range otherIssue.Blocks {
			if blockID == id {
				if !visited[otherIssue.ID] {
					if err := r.detectCycle(otherIssue.ID, visited, recStack); err != nil {
						return err
					}
				} else if recStack[otherIssue.ID] {
					return fmt.Errorf("circular dependency detected: %s blocks %s", otherIssue.ID, id)
				}
			}
		}
	}

	recStack[id] = false
	return nil
}

// GetCreationOrder returns issues in the order they should be created
func (r *Resolver) GetCreationOrder() ([]*github.Issue, error) {
	if err := r.Validate(); err != nil {
		return nil, err
	}

	// Topological sort using Kahn's algorithm
	inDegree := make(map[string]int)
	adjList := make(map[string][]string)

	// Initialize in-degree and adjacency list
	for _, issue := range r.issues {
		if _, exists := inDegree[issue.ID]; !exists {
			inDegree[issue.ID] = 0
		}
		if _, exists := adjList[issue.ID]; !exists {
			adjList[issue.ID] = []string{}
		}
	}

	// Build dependency graph
	for _, issue := range r.issues {
		// Handle DependsOn relationships
		for _, depID := range issue.DependsOn {
			adjList[depID] = append(adjList[depID], issue.ID)
			inDegree[issue.ID]++
		}

		// Handle Blocks relationships (reverse)
		for _, blockID := range issue.Blocks {
			adjList[issue.ID] = append(adjList[issue.ID], blockID)
			inDegree[blockID]++
		}
	}

	// Find all nodes with no incoming edges
	queue := []string{}
	for id, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, id)
		}
	}

	result := []*github.Issue{}
	processedCount := 0

	// Process queue
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		result = append(result, r.idToIssue[current])
		processedCount++

		// Reduce in-degree for neighbors
		for _, neighbor := range adjList[current] {
			inDegree[neighbor]--
			if inDegree[neighbor] == 0 {
				queue = append(queue, neighbor)
			}
		}
	}

	// Check if all issues were processed
	if processedCount != len(r.issues) {
		return nil, fmt.Errorf("dependency graph contains cycles or isolated nodes")
	}

	return result, nil
}

// GetDependencyInfo returns a string describing the dependencies for an issue
func (r *Resolver) GetDependencyInfo(issue *github.Issue) string {
	info := ""
	
	if len(issue.DependsOn) > 0 {
		info += fmt.Sprintf("Depends on: %v", issue.DependsOn)
	}
	
	if len(issue.Blocks) > 0 {
		if info != "" {
			info += ", "
		}
		info += fmt.Sprintf("Blocks: %v", issue.Blocks)
	}
	
	if len(issue.Related) > 0 {
		if info != "" {
			info += ", "
		}
		info += fmt.Sprintf("Related: %v", issue.Related)
	}
	
	return info
}