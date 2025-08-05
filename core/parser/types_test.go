package parser

import (
	"testing"
)

func TestGetSemanticVersion(t *testing.T) {
	tests := []struct {
		name     string
		version  string
		expected SemanticVersion
		valid    bool
	}{
		{
			name:    "valid basic version",
			version: "1.2.3",
			expected: SemanticVersion{
				Major: 1,
				Minor: 2,
				Patch: 3,
			},
			valid: true,
		},
		{
			name:    "valid version with v prefix",
			version: "v1.2.3",
			expected: SemanticVersion{
				Major: 1,
				Minor: 2,
				Patch: 3,
			},
			valid: true,
		},
		{
			name:    "valid version with url prefix",
			version: "https://github.com/some-owner/some-repo.git?rev=v1.9.1",
			expected: SemanticVersion{
				Major: 1,
				Minor: 9,
				Patch: 1,
			},
			valid: true,
		},
		{
			name:    "valid version with pre-release",
			version: "1.2.3-beta",
			expected: SemanticVersion{
				Major:      1,
				Minor:      2,
				Patch:      3,
				PreRelease: "beta",
			},
			valid: true,
		},
		{
			name:    "valid version with build metadata",
			version: "1.2.3+build.1",
			expected: SemanticVersion{
				Major:         1,
				Minor:         2,
				Patch:         3,
				BuildMetaData: "build.1",
			},
			valid: true,
		},
		{
			name:    "valid version with pre-release and build metadata",
			version: "1.2.3-alpha.1+build.123",
			expected: SemanticVersion{
				Major:         1,
				Minor:         2,
				Patch:         3,
				PreRelease:    "alpha.1",
				BuildMetaData: "build.123",
			},
			valid: true,
		},
		{
			name:    "invalid version missing patch",
			version: "1.2",
			valid:   false,
		},
		{
			name:    "invalid version missing minor",
			version: "1",
			valid:   false,
		},
		{
			name:    "invalid version with non-numeric major",
			version: "a.2.3",
			valid:   false,
		},
		{
			name:    "invalid version with non-numeric minor",
			version: "1.b.3",
			valid:   false,
		},
		{
			name:    "invalid version with non-numeric patch",
			version: "1.2.c",
			valid:   false,
		},
		{
			name:    "invalid empty version",
			version: "",
			valid:   false,
		},
		{
			name:    "invalid random string",
			version: "not-a-version",
			valid:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, ok := GetSemanticVersion(tt.version)

			if ok != tt.valid {
				t.Errorf("GetSemanticVersion(%q) valid = %v, want %v", tt.version, ok, tt.valid)
				return
			}

			if !tt.valid {
				return // Skip checking result for invalid cases
			}

			if result.Major != tt.expected.Major {
				t.Errorf("GetSemanticVersion(%q) Major = %d, want %d", tt.version, result.Major, tt.expected.Major)
			}
			if result.Minor != tt.expected.Minor {
				t.Errorf("GetSemanticVersion(%q) Minor = %d, want %d", tt.version, result.Minor, tt.expected.Minor)
			}
			if result.Patch != tt.expected.Patch {
				t.Errorf("GetSemanticVersion(%q) Patch = %d, want %d", tt.version, result.Patch, tt.expected.Patch)
			}
			if result.PreRelease != tt.expected.PreRelease {
				t.Errorf("GetSemanticVersion(%q) PreRelease = %q, want %q", tt.version, result.PreRelease, tt.expected.PreRelease)
			}
			if result.BuildMetaData != tt.expected.BuildMetaData {
				t.Errorf("GetSemanticVersion(%q) BuildMetaData = %q, want %q", tt.version, result.BuildMetaData, tt.expected.BuildMetaData)
			}
		})
	}
}
