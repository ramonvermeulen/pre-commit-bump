package bumper

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractGitLabRepo(t *testing.T) {
	tests := []struct {
		name     string
		repoURL  string
		expected string
	}{
		{
			name:     "https URL",
			repoURL:  "https://gitlab.com/owner/repo",
			expected: "owner/repo",
		},
		{
			name:     "nested group URL",
			repoURL:  "https://gitlab.com/group/subgroup/repo/very/deeply/nested",
			expected: "group/subgroup/repo/very/deeply/nested",
		},
		{
			name:     "https URL with .git suffix",
			repoURL:  "https://gitlab.com/owner/repo.git",
			expected: "owner/repo",
		},
		{
			name:     "ssh URL",
			repoURL:  "git@gitlab.com:owner/repo.git",
			expected: "owner/repo",
		},
		{
			name:     "URL with trailing slash",
			repoURL:  "https://gitlab.com/owner/repo/",
			expected: "owner/repo",
		},
		{
			name:     "URL with query parameters",
			repoURL:  "https://gitlab.com/owner/repo?ref=main",
			expected: "owner/repo",
		},
		{
			name:     "Wrong vendor URL",
			repoURL:  "https://bitbucket.org/owner/repo",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractGitLabRepo(tt.repoURL)
			assert.Equal(t, tt.expected, result)
		})
	}
}
