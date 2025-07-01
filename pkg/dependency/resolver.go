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

// detectCycle uses DFS to detect circular dependencies (DependsOn only)
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

	// Blocks and Related relationships are not checked for cycles

	recStack[id] = false
	return nil
}

// GetCreationOrder returns issues in the order they should be created
// Uses file order as primary sort, DependsOn relationships as secondary constraint
func (r *Resolver) GetCreationOrder() ([]*github.Issue, error) {
	if err := r.Validate(); err != nil {
		return nil, err
	}

	// Create position map to preserve original order
	positionMap := make(map[string]int)
	for i, issue := range r.issues {
		positionMap[issue.ID] = i
	}

	// Topological sort using Kahn's algorithm with stable sorting
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

	// Build dependency graph - only consider DependsOn relationships
	for _, issue := range r.issues {
		// Handle DependsOn relationships only
		for _, depID := range issue.DependsOn {
			adjList[depID] = append(adjList[depID], issue.ID)
			inDegree[issue.ID]++
		}
		
		// Blocks and Related relationships are ignored for ordering
	}

	result := []*github.Issue{}
	processedCount := 0

	// Process all issues with stable topological sorting
	for {
		// Find all nodes with no incoming edges for this level
		currentLevel := []string{}
		for _, issue := range r.issues {
			if inDegree[issue.ID] == 0 {
				currentLevel = append(currentLevel, issue.ID)
			}
		}

		if len(currentLevel) == 0 {
			break
		}

		// Process current level one by one to maintain file order
		// Select the earliest one in file order first
		minPosition := len(r.issues)
		selectedID := ""
		for _, id := range currentLevel {
			if pos := positionMap[id]; pos < minPosition {
				minPosition = pos
				selectedID = id
			}
		}

		if selectedID != "" {
			result = append(result, r.idToIssue[selectedID])
			processedCount++

			// Remove this node from consideration
			inDegree[selectedID] = -1

			// Reduce in-degree for neighbors
			for _, neighbor := range adjList[selectedID] {
				if inDegree[neighbor] > 0 {
					inDegree[neighbor]--
				}
			}
		} else {
			break
		}
	}

	// Check if all issues were processed
	if processedCount != len(r.issues) {
		return nil, fmt.Errorf("dependency graph contains cycles or isolated nodes")
	}

	return result, nil
}

// GetOriginalOrder returns issues in their original file order without dependency sorting
func (r *Resolver) GetOriginalOrder() ([]*github.Issue, error) {
	// Still validate to check for missing references, but don't enforce dependency order
	if err := r.validateReferences(); err != nil {
		return nil, err
	}

	// Return issues in their original order
	return r.issues, nil
}

// validateReferences checks for missing references without dependency cycle validation
func (r *Resolver) validateReferences() error {
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

	return nil
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