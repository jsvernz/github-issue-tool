package creator

import (
	"testing"

	"github.com/ef-tech/github-issue-tool/pkg/github"
)

// MockLabelClient implements both Client and LabelClient for testing
type MockLabelClient struct {
	owner, repo    string
	existingLabels map[string]bool
	createdLabels  []MockLabel
	issues         []*github.Issue
}

type MockLabel struct {
	Name        string
	Description string
	Color       string
}

func NewMockLabelClient(owner, repo string) *MockLabelClient {
	return &MockLabelClient{
		owner:          owner,
		repo:           repo,
		existingLabels: make(map[string]bool),
		createdLabels:  make([]MockLabel, 0),
		issues:         make([]*github.Issue, 0),
	}
}

func (m *MockLabelClient) CreateIssue(issue *github.Issue) error {
	issue.Number = len(m.issues) + 1
	m.issues = append(m.issues, issue)
	return nil
}

func (m *MockLabelClient) GetRepository() (string, string) {
	return m.owner, m.repo
}

func (m *MockLabelClient) CreateLabel(name, description, color string) error {
	m.createdLabels = append(m.createdLabels, MockLabel{
		Name:        name,
		Description: description,
		Color:       color,
	})
	m.existingLabels[name] = true
	return nil
}

func (m *MockLabelClient) LabelExists(name string) (bool, error) {
	return m.existingLabels[name], nil
}

func (m *MockLabelClient) SetExistingLabels(labels []string) {
	for _, label := range labels {
		m.existingLabels[label] = true
	}
}

func TestCreator_ensureLabelsExist(t *testing.T) {
	tests := []struct {
		name           string
		existingLabels []string
		requestedLabels []string
		expectedCreated []string
	}{
		{
			name:           "all labels exist",
			existingLabels: []string{"bug", "feature", "epic"},
			requestedLabels: []string{"bug", "feature"},
			expectedCreated: []string{},
		},
		{
			name:           "some labels missing",
			existingLabels: []string{"bug"},
			requestedLabels: []string{"bug", "feature", "epic"},
			expectedCreated: []string{"feature", "epic"},
		},
		{
			name:           "no existing labels",
			existingLabels: []string{},
			requestedLabels: []string{"bug", "feature", "unknown-label"},
			expectedCreated: []string{"bug", "feature", "unknown-label"},
		},
		{
			name:           "empty request",
			existingLabels: []string{"bug"},
			requestedLabels: []string{},
			expectedCreated: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := NewMockLabelClient("owner", "repo")
			mockClient.SetExistingLabels(tt.existingLabels)
			
			creator := NewCreator(mockClient, false)
			
			err := creator.ensureLabelsExist(mockClient, tt.requestedLabels)
			if err != nil {
				t.Errorf("ensureLabelsExist() error = %v", err)
				return
			}
			
			// Check that the expected labels were created
			if len(mockClient.createdLabels) != len(tt.expectedCreated) {
				t.Errorf("Expected %d labels to be created, but %d were created", 
					len(tt.expectedCreated), len(mockClient.createdLabels))
				return
			}
			
			createdNames := make(map[string]bool)
			for _, created := range mockClient.createdLabels {
				createdNames[created.Name] = true
			}
			
			for _, expected := range tt.expectedCreated {
				if !createdNames[expected] {
					t.Errorf("Expected label %s to be created, but it wasn't", expected)
				}
			}
		})
	}
}

func TestCreator_ensureLabelsExist_DefaultConfigs(t *testing.T) {
	mockClient := NewMockLabelClient("owner", "repo")
	creator := NewCreator(mockClient, false)
	
	// Test with a known default label
	err := creator.ensureLabelsExist(mockClient, []string{"epic"})
	if err != nil {
		t.Errorf("ensureLabelsExist() error = %v", err)
		return
	}
	
	if len(mockClient.createdLabels) != 1 {
		t.Errorf("Expected 1 label to be created, but %d were created", len(mockClient.createdLabels))
		return
	}
	
	created := mockClient.createdLabels[0]
	if created.Name != "epic" {
		t.Errorf("Expected label name 'epic', got '%s'", created.Name)
	}
	if created.Description != "Epic issue" {
		t.Errorf("Expected description 'Epic issue', got '%s'", created.Description)
	}
	if created.Color != "d73a4a" {
		t.Errorf("Expected color 'd73a4a', got '%s'", created.Color)
	}
}

func TestCreator_ensureLabelsExist_UnknownLabel(t *testing.T) {
	mockClient := NewMockLabelClient("owner", "repo")
	creator := NewCreator(mockClient, false)
	
	// Test with an unknown label
	err := creator.ensureLabelsExist(mockClient, []string{"unknown-custom-label"})
	if err != nil {
		t.Errorf("ensureLabelsExist() error = %v", err)
		return
	}
	
	if len(mockClient.createdLabels) != 1 {
		t.Errorf("Expected 1 label to be created, but %d were created", len(mockClient.createdLabels))
		return
	}
	
	created := mockClient.createdLabels[0]
	if created.Name != "unknown-custom-label" {
		t.Errorf("Expected label name 'unknown-custom-label', got '%s'", created.Name)
	}
	if created.Description != "Label: unknown-custom-label" {
		t.Errorf("Expected description 'Label: unknown-custom-label', got '%s'", created.Description)
	}
	if created.Color != "cccccc" {
		t.Errorf("Expected color 'cccccc', got '%s'", created.Color)
	}
}