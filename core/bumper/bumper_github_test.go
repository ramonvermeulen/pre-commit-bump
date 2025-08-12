package bumper

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractGitHubRepo(t *testing.T) {
	tests := []struct {
		name     string
		repoURL  string
		expected string
	}{
		{
			name:     "https URL",
			repoURL:  "https://github.com/owner/repo",
			expected: "owner/repo",
		},
		{
			name:     "https URL with .git suffix",
			repoURL:  "https://github.com/owner/repo.git",
			expected: "owner/repo",
		},
		{
			name:     "ssh URL",
			repoURL:  "git@github.com:owner/repo.git",
			expected: "owner/repo",
		},
		{
			name:     "URL with trailing slash",
			repoURL:  "https://github.com/owner/repo/",
			expected: "owner/repo",
		},
		{
			name:     "URL with query parameters",
			repoURL:  "https://github.com/owner/repo?ref=main",
			expected: "owner/repo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractGitHubRepo(tt.repoURL)
			assert.Equal(t, tt.expected, result)
		})
	}
}
