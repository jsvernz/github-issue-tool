package github

type Issue struct {
	ID        string   // Unique ID from the input file
	Title     string
	Body      string
	Labels    []string
	Assignees []string
	DependsOn []string // IDs of issues this depends on
	Blocks    []string // IDs of issues this blocks
	Related   []string // IDs of related issues
	Number    int      // GitHub issue number after creation
}

type Client interface {
	CreateIssue(issue *Issue) error
	GetRepository() (owner, name string)
}

type DependencyClient interface {
	Client
	UpdateIssueWithDependencies(issue *Issue, issueNumberMap map[string]int) error
}