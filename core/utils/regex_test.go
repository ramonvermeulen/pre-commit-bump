package utils

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetGroup(t *testing.T) {
	tests := []struct {
		name      string
		pattern   string
		input     string
		groupName string
		expected  string
	}{
		{
			name:      "valid named group exists",
			pattern:   `(?P<version>\d+\.\d+\.\d+)`,
			input:     "v1.2.3",
			groupName: "version",
			expected:  "1.2.3",
		},
		{
			name:      "multiple named groups - get first",
			pattern:   `(?P<major>\d+)\.(?P<minor>\d+)\.(?P<patch>\d+)`,
			input:     "1.2.3",
			groupName: "major",
			expected:  "1",
		},
		{
			name:      "multiple named groups - get middle",
			pattern:   `(?P<major>\d+)\.(?P<minor>\d+)\.(?P<patch>\d+)`,
			input:     "1.2.3",
			groupName: "minor",
			expected:  "2",
		},
		{
			name:      "multiple named groups - get last",
			pattern:   `(?P<major>\d+)\.(?P<minor>\d+)\.(?P<patch>\d+)`,
			input:     "1.2.3",
			groupName: "patch",
			expected:  "3",
		},
		{
			name:      "named group with empty match",
			pattern:   `(?P<optional>\d*)`,
			input:     "",
			groupName: "optional",
			expected:  "",
		},
		{
			name:      "non-existent group name",
			pattern:   `(?P<version>\d+\.\d+\.\d+)`,
			input:     "1.2.3",
			groupName: "nonexistent",
			expected:  "",
		},
		{
			name:      "group name exists but match is empty",
			pattern:   `(?P<version>\d+\.\d+\.\d+)`,
			input:     "no-match",
			groupName: "version",
			expected:  "",
		},
		{
			name:      "complex pattern with multiple groups",
			pattern:   `https://github\.com/(?P<owner>[^/]+)/(?P<repo>[^/]+)`,
			input:     "https://github.com/owner/repository",
			groupName: "owner",
			expected:  "owner",
		},
		{
			name:      "complex pattern with multiple groups - get repo",
			pattern:   `https://github\.com/(?P<owner>[^/]+)/(?P<repo>[^/]+)`,
			input:     "https://github.com/owner/repository",
			groupName: "repo",
			expected:  "repository",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			re := regexp.MustCompile(tt.pattern)
			match := re.FindStringSubmatch(tt.input)

			result := GetGroup(re, match, tt.groupName)

			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetGroup_EdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		setupTest func(t *testing.T) (*regexp.Regexp, []string, string)
		expected  string
	}{
		{
			name: "nil match slice",
			setupTest: func(t *testing.T) (*regexp.Regexp, []string, string) {
				re := regexp.MustCompile(`(?P<test>\d+)`)
				return re, nil, "test"
			},
			expected: "",
		},
		{
			name: "empty match slice",
			setupTest: func(t *testing.T) (*regexp.Regexp, []string, string) {
				re := regexp.MustCompile(`(?P<test>\d+)`)
				return re, []string{}, "test"
			},
			expected: "",
		},
		{
			name: "match slice shorter than expected index",
			setupTest: func(t *testing.T) (*regexp.Regexp, []string, string) {
				re := regexp.MustCompile(`(?P<test>\d+)`)
				return re, []string{"full"}, "test"
			},
			expected: "",
		},
		{
			name: "empty group name",
			setupTest: func(t *testing.T) (*regexp.Regexp, []string, string) {
				re := regexp.MustCompile(`(?P<test>\d+)`)
				match := re.FindStringSubmatch("123")
				return re, match, ""
			},
			expected: "",
		},
		{
			name: "regex without named groups",
			setupTest: func(t *testing.T) (*regexp.Regexp, []string, string) {
				re := regexp.MustCompile(`(\d+)\.(\d+)\.(\d+)`)
				match := re.FindStringSubmatch("1.2.3")
				return re, match, "version"
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			re, match, groupName := tt.setupTest(t)

			result := GetGroup(re, match, groupName)

			assert.Equal(t, tt.expected, result)
		})
	}
}
