package parser

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestParseIssuesFile(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    int
		wantErr bool
	}{
		{
			name: "single issue",
			content: `## [ISSUE-001] First Issue
Labels: bug, enhancement
Assignees: user1, user2
Depends: ISSUE-002
Blocks: ISSUE-003
Related: ISSUE-004

This is the body of the first issue.
It can have multiple lines.`,
			want:    1,
			wantErr: false,
		},
		{
			name: "multiple issues with separator",
			content: `## [ISSUE-001] First Issue
Labels: bug

First issue body
---
## [ISSUE-002] Second Issue
Labels: feature

Second issue body`,
			want:    2,
			wantErr: false,
		},
		{
			name: "multiple issues without separator",
			content: `## [ISSUE-001] First Issue
Labels: bug

First issue body

## [ISSUE-002] Second Issue
Labels: feature

Second issue body`,
			want:    2,
			wantErr: false,
		},
		{
			name: "empty file",
			content: ``,
			want:    0,
			wantErr: true,
		},
		{
			name: "duplicate IDs",
			content: `## [ISSUE-001] First Issue

Body 1

## [ISSUE-001] Duplicate ID

Body 2`,
			want:    0,
			wantErr: true,
		},
		{
			name: "invalid header format",
			content: `## Invalid Header Format

Some body text`,
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary file
			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, "test_issues.txt")
			err := os.WriteFile(tmpFile, []byte(tt.content), 0644)
			if err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			// Parse the file
			issues, err := ParseIssuesFile(tmpFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseIssuesFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(issues) != tt.want {
				t.Errorf("ParseIssuesFile() got %d issues, want %d", len(issues), tt.want)
			}
		})
	}
}

func TestParseCommaSeparated(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "simple list",
			input: "a, b, c",
			want:  []string{"a", "b", "c"},
		},
		{
			name:  "with extra spaces",
			input: "  a  ,  b  ,  c  ",
			want:  []string{"a", "b", "c"},
		},
		{
			name:  "empty string",
			input: "",
			want:  nil,
		},
		{
			name:  "single item",
			input: "single",
			want:  []string{"single"},
		},
		{
			name:  "trailing comma",
			input: "a, b, c,",
			want:  []string{"a", "b", "c"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseCommaSeparated(tt.input)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseCommaSeparated() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseIssueDetails(t *testing.T) {
	content := `## [TEST-001] Test Issue
Labels: bug, high-priority
Assignees: alice, bob
Depends: TEST-002, TEST-003
Blocks: TEST-004
Related: TEST-005, TEST-006

This is the issue body.
It has multiple lines.

And even paragraphs.`

	// Create temporary file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test_issue.txt")
	err := os.WriteFile(tmpFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	issues, err := ParseIssuesFile(tmpFile)
	if err != nil {
		t.Fatalf("ParseIssuesFile() failed: %v", err)
	}

	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}

	issue := issues[0]

	// Check all fields
	if issue.ID != "TEST-001" {
		t.Errorf("ID = %s, want TEST-001", issue.ID)
	}
	if issue.Title != "Test Issue" {
		t.Errorf("Title = %s, want Test Issue", issue.Title)
	}
	if !reflect.DeepEqual(issue.Labels, []string{"bug", "high-priority"}) {
		t.Errorf("Labels = %v, want [bug high-priority]", issue.Labels)
	}
	if !reflect.DeepEqual(issue.Assignees, []string{"alice", "bob"}) {
		t.Errorf("Assignees = %v, want [alice bob]", issue.Assignees)
	}
	if !reflect.DeepEqual(issue.DependsOn, []string{"TEST-002", "TEST-003"}) {
		t.Errorf("DependsOn = %v, want [TEST-002 TEST-003]", issue.DependsOn)
	}
	if !reflect.DeepEqual(issue.Blocks, []string{"TEST-004"}) {
		t.Errorf("Blocks = %v, want [TEST-004]", issue.Blocks)
	}
	if !reflect.DeepEqual(issue.Related, []string{"TEST-005", "TEST-006"}) {
		t.Errorf("Related = %v, want [TEST-005 TEST-006]", issue.Related)
	}

	expectedBody := `This is the issue body.
It has multiple lines.

And even paragraphs.`
	if issue.Body != expectedBody {
		t.Errorf("Body = %q, want %q", issue.Body, expectedBody)
	}
}