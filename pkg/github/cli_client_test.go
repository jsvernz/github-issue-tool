package github

import (
	"encoding/json"
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

func TestCLIClient_LabelExists_JSONParsing(t *testing.T) {
	tests := []struct {
		name       string
		jsonOutput string
		labelName  string
		want       bool
		hasErr     bool
	}{
		{
			name:       "label exists",
			jsonOutput: `[{"name": "bug"}, {"name": "feature"}, {"name": "enhancement"}]`,
			labelName:  "bug",
			want:       true,
			hasErr:     false,
		},
		{
			name:       "label does not exist",
			jsonOutput: `[{"name": "bug"}, {"name": "feature"}, {"name": "enhancement"}]`,
			labelName:  "non-existent",
			want:       false,
			hasErr:     false,
		},
		{
			name:       "empty labels",
			jsonOutput: `[]`,
			labelName:  "bug",
			want:       false,
			hasErr:     false,
		},
		{
			name:       "invalid JSON",
			jsonOutput: `invalid json`,
			labelName:  "bug",
			want:       false,
			hasErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the JSON parsing logic manually
			var labels []struct {
				Name string `json:"name"`
			}
			
			err := json.Unmarshal([]byte(tt.jsonOutput), &labels)
			if (err != nil) != tt.hasErr {
				if !tt.hasErr {
					t.Errorf("Unexpected JSON parsing error: %v", err)
				}
				return
			}
			
			if tt.hasErr {
				return // Expected error, test passed
			}
			
			// Check if label exists
			exists := false
			for _, label := range labels {
				if label.Name == tt.labelName {
					exists = true
					break
				}
			}
			
			if exists != tt.want {
				t.Errorf("Label existence check = %v, want %v", exists, tt.want)
			}
		})
	}
}