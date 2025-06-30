package github

import (
	"testing"
)

func TestParseIssueNumber(t *testing.T) {
	tests := []struct {
		name   string
		output string
		want   int
		hasErr bool
	}{
		{
			name:   "JSON format",
			output: `{"number": 123, "title": "Test Issue"}`,
			want:   123,
			hasErr: false,
		},
		{
			name:   "URL format",
			output: "https://github.com/owner/repo/issues/456",
			want:   456,
			hasErr: false,
		},
		{
			name:   "URL format with newlines",
			output: "Issue created successfully\nhttps://github.com/owner/repo/issues/789\nDone",
			want:   789,
			hasErr: false,
		},
		{
			name:   "invalid format",
			output: "Some random output without issue number",
			want:   0,
			hasErr: true,
		},
		{
			name:   "empty output",
			output: "",
			want:   0,
			hasErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewCLIClient("owner", "repo")
			got, err := client.parseIssueNumber(tt.output)
			if (err != nil) != tt.hasErr {
				t.Errorf("parseIssueNumber() error = %v, hasErr %v", err, tt.hasErr)
				return
			}
			if got != tt.want {
				t.Errorf("parseIssueNumber() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCLIClient_GetRepository(t *testing.T) {
	client := NewCLIClient("test-owner", "test-repo")
	owner, repo := client.GetRepository()
	
	if owner != "test-owner" {
		t.Errorf("GetRepository() owner = %v, want %v", owner, "test-owner")
	}
	if repo != "test-repo" {
		t.Errorf("GetRepository() repo = %v, want %v", repo, "test-repo")
	}
}