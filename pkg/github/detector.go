package github

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type Environment struct {
	HasGitHubCLI     bool
	IsAuthenticated  bool
	IsGitRepository  bool
	RepositoryOwner  string
	RepositoryName   string
	PreferredMethod  string // "cli" or "api"
}

func DetectEnvironment() (*Environment, error) {
	env := &Environment{}

	// Check if GitHub CLI is available
	env.HasGitHubCLI = checkGitHubCLI()

	// Check if GitHub CLI is authenticated
	if env.HasGitHubCLI {
		env.IsAuthenticated = checkGitHubCLIAuth()
	}

	// Check if we're in a Git repository
	env.IsGitRepository = checkGitRepository()

	// Get repository information if in a Git repository
	if env.IsGitRepository {
		owner, name, err := getRepositoryInfo()
		if err == nil {
			env.RepositoryOwner = owner
			env.RepositoryName = name
		}
	}

	// Determine preferred method
	if env.HasGitHubCLI && env.IsAuthenticated {
		env.PreferredMethod = "cli"
	} else if os.Getenv("GITHUB_TOKEN") != "" {
		env.PreferredMethod = "api"
	} else {
		env.PreferredMethod = ""
	}

	return env, nil
}

func checkGitHubCLI() bool {
	cmd := exec.Command("gh", "--version")
	err := cmd.Run()
	return err == nil
}

func checkGitHubCLIAuth() bool {
	cmd := exec.Command("gh", "auth", "status")
	err := cmd.Run()
	return err == nil
}

func checkGitRepository() bool {
	_, err := os.Stat(".git")
	return err == nil
}

func getRepositoryInfo() (string, string, error) {
	cmd := exec.Command("git", "remote", "get-url", "origin")
	output, err := cmd.Output()
	if err != nil {
		return "", "", err
	}

	url := strings.TrimSpace(string(output))
	owner, name := parseGitHubURL(url)
	if owner == "" || name == "" {
		return "", "", fmt.Errorf("could not parse repository information from URL: %s", url)
	}

	return owner, name, nil
}

func parseGitHubURL(url string) (string, string) {
	// Handle SSH URLs: git@github.com:owner/repo.git
	if strings.HasPrefix(url, "git@github.com:") {
		url = strings.TrimPrefix(url, "git@github.com:")
		url = strings.TrimSuffix(url, ".git")
		parts := strings.Split(url, "/")
		if len(parts) == 2 {
			return parts[0], parts[1]
		}
	}

	// Handle HTTPS URLs: https://github.com/owner/repo.git
	if strings.HasPrefix(url, "https://github.com/") {
		url = strings.TrimPrefix(url, "https://github.com/")
		url = strings.TrimSuffix(url, ".git")
		parts := strings.Split(url, "/")
		if len(parts) == 2 {
			return parts[0], parts[1]
		}
	}

	return "", ""
}

func (e *Environment) String() string {
	var sb strings.Builder
	sb.WriteString("GitHub Environment Detection:\n")
	sb.WriteString(fmt.Sprintf("  GitHub CLI available: %v\n", e.HasGitHubCLI))
	if e.HasGitHubCLI {
		sb.WriteString(fmt.Sprintf("  GitHub CLI authenticated: %v\n", e.IsAuthenticated))
	}
	sb.WriteString(fmt.Sprintf("  In Git repository: %v\n", e.IsGitRepository))
	if e.IsGitRepository && e.RepositoryOwner != "" {
		sb.WriteString(fmt.Sprintf("  Repository: %s/%s\n", e.RepositoryOwner, e.RepositoryName))
	}
	if e.PreferredMethod != "" {
		sb.WriteString(fmt.Sprintf("  Preferred method: %s\n", e.PreferredMethod))
	} else {
		sb.WriteString("  No authentication method available\n")
	}
	return sb.String()
}