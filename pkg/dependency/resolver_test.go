package dependency

import (
	"testing"

	"github.com/ef-tech/github-issue-tool/pkg/github"
)

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		issues  []*github.Issue
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid dependencies",
			issues: []*github.Issue{
				{ID: "A", DependsOn: []string{"B"}},
				{ID: "B", DependsOn: []string{}},
			},
			wantErr: false,
		},
		{
			name: "missing dependency",
			issues: []*github.Issue{
				{ID: "A", DependsOn: []string{"B"}},
			},
			wantErr: true,
			errMsg:  "non-existent issue B",
		},
		{
			name: "circular dependency",
			issues: []*github.Issue{
				{ID: "A", DependsOn: []string{"B"}},
				{ID: "B", DependsOn: []string{"A"}},
			},
			wantErr: true,
			errMsg:  "circular dependency",
		},
		{
			name: "circular via blocks",
			issues: []*github.Issue{
				{ID: "A", Blocks: []string{"B"}},
				{ID: "B", Blocks: []string{"A"}},
			},
			wantErr: false, // Blocks relationships don't create circular dependencies in v0.3.1+
		},
		{
			name: "complex valid graph",
			issues: []*github.Issue{
				{ID: "A", DependsOn: []string{"B", "C"}},
				{ID: "B", DependsOn: []string{"D"}},
				{ID: "C", DependsOn: []string{"D"}},
				{ID: "D", DependsOn: []string{}},
			},
			wantErr: false,
		},
		{
			name: "missing related issue",
			issues: []*github.Issue{
				{ID: "A", Related: []string{"B"}},
			},
			wantErr: true,
			errMsg:  "non-existent issue B",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewResolver(tt.issues)
			err := r.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.errMsg != "" {
				if !contains(err.Error(), tt.errMsg) {
					t.Errorf("Validate() error = %v, want error containing %v", err, tt.errMsg)
				}
			}
		})
	}
}

func TestGetCreationOrder(t *testing.T) {
	tests := []struct {
		name    string
		issues  []*github.Issue
		wantIDs []string
		wantErr bool
	}{
		{
			name: "simple chain",
			issues: []*github.Issue{
				{ID: "A", DependsOn: []string{"B"}},
				{ID: "B", DependsOn: []string{"C"}},
				{ID: "C", DependsOn: []string{}},
			},
			wantIDs: []string{"C", "B", "A"},
			wantErr: false,
		},
		{
			name: "parallel dependencies",
			issues: []*github.Issue{
				{ID: "A", DependsOn: []string{"B", "C"}},
				{ID: "B", DependsOn: []string{}},
				{ID: "C", DependsOn: []string{}},
			},
			wantIDs: []string{"B", "C", "A"}, // B and C can be in any order
			wantErr: false,
		},
		{
			name: "blocks relationship",
			issues: []*github.Issue{
				{ID: "A", Blocks: []string{"B"}},
				{ID: "B", DependsOn: []string{}},
			},
			wantIDs: []string{"A", "B"},
			wantErr: false,
		},
		{
			name: "mixed dependencies",
			issues: []*github.Issue{
				{ID: "A", DependsOn: []string{"B"}, Blocks: []string{"C"}},
				{ID: "B", DependsOn: []string{}},
				{ID: "C", DependsOn: []string{}},
			},
			wantIDs: []string{"B", "A", "C"},
			wantErr: false,
		},
		{
			name: "circular dependency",
			issues: []*github.Issue{
				{ID: "A", DependsOn: []string{"B"}},
				{ID: "B", DependsOn: []string{"A"}},
			},
			wantIDs: nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewResolver(tt.issues)
			order, err := r.GetCreationOrder()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetCreationOrder() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if len(order) != len(tt.wantIDs) {
					t.Errorf("GetCreationOrder() returned %d issues, want %d", len(order), len(tt.wantIDs))
					return
				}

				// For cases where order matters strictly
				if tt.name == "simple chain" || tt.name == "blocks relationship" || tt.name == "mixed dependencies" {
					for i, issue := range order {
						if issue.ID != tt.wantIDs[i] {
							t.Errorf("GetCreationOrder()[%d] = %s, want %s", i, issue.ID, tt.wantIDs[i])
						}
					}
				} else {
					// For cases where some orders are flexible, just check all IDs are present
					gotIDs := make(map[string]bool)
					for _, issue := range order {
						gotIDs[issue.ID] = true
					}
					for _, wantID := range tt.wantIDs {
						if !gotIDs[wantID] {
							t.Errorf("GetCreationOrder() missing ID %s", wantID)
						}
					}
				}
			}
		})
	}
}

func TestGetDependencyInfo(t *testing.T) {
	tests := []struct {
		name  string
		issue *github.Issue
		want  string
	}{
		{
			name:  "no dependencies",
			issue: &github.Issue{ID: "A"},
			want:  "",
		},
		{
			name:  "only depends on",
			issue: &github.Issue{ID: "A", DependsOn: []string{"B", "C"}},
			want:  "Depends on: [B C]",
		},
		{
			name:  "only blocks",
			issue: &github.Issue{ID: "A", Blocks: []string{"B"}},
			want:  "Blocks: [B]",
		},
		{
			name:  "all types",
			issue: &github.Issue{
				ID:        "A",
				DependsOn: []string{"B"},
				Blocks:    []string{"C"},
				Related:   []string{"D"},
			},
			want: "Depends on: [B], Blocks: [C], Related: [D]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewResolver([]*github.Issue{tt.issue})
			got := r.GetDependencyInfo(tt.issue)
			if got != tt.want {
				t.Errorf("GetDependencyInfo() = %v, want %v", got, tt.want)
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}