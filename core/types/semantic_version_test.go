package types

import (
	"testing"
)

func TestSemanticVersionEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		version  string
		expected SemanticVersion
		valid    bool
	}{
		{
			name:    "zero version",
			version: "0.0.0",
			expected: SemanticVersion{
				Major: 0,
				Minor: 0,
				Patch: 0,
			},
			valid: true,
		},
		{
			name:    "large version numbers",
			version: "999.999.999",
			expected: SemanticVersion{
				Major: 999,
				Minor: 999,
				Patch: 999,
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
			version: "https://github.com/some-owner/some-repo.git?rev=v1.9.1&some-param=some-value",
			expected: SemanticVersion{
				Major: 1,
				Minor: 9,
				Patch: 1,
			},
			valid: true,
		},
		{
			name:    "version with leading zeros",
			version: "01.02.03",
			valid:   false,
		},
		{
			name:    "version with multiple hyphens in pre-release",
			version: "1.0.0-alpha-beta-1",
			expected: SemanticVersion{
				Major:      1,
				Minor:      0,
				Patch:      0,
				PreRelease: "alpha-beta-1",
			},
			valid: true,
		},
		{
			name:    "version with complex build metadata",
			version: "1.0.0+20130313144700",
			expected: SemanticVersion{
				Major:         1,
				Minor:         0,
				Patch:         0,
				BuildMetaData: "20130313144700",
			},
			valid: true,
		},
		{
			name:    "version with empty pre-release",
			version: "1.0.0-",
			expected: SemanticVersion{
				Major: 1,
				Minor: 0,
				Patch: 0,
			},
			valid: true,
		},
		{
			name:    "version with empty build metadata",
			version: "1.0.0+",
			expected: SemanticVersion{
				Major: 1,
				Minor: 0,
				Patch: 0,
			},
			valid: true,
		},
		{
			name:    "version with negative numbers",
			version: "-1.0.0",
			expected: SemanticVersion{
				Major: 1,
				Minor: 0,
				Patch: 0,
			},
			valid: true,
		},
		{
			name:    "version with extra dots",
			version: "1.0.0.1",
			expected: SemanticVersion{
				Major: 1,
				Minor: 0,
				Patch: 0,
			},
			valid: true,
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
				return
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

func TestSemanticVersionComparison(t *testing.T) {
	tests := []struct {
		name     string
		version1 string
		version2 string
		expected bool
	}{
		{
			name:     "equal versions",
			version1: "1.0.0",
			version2: "1.0.0",
			expected: false,
		},
		{
			name:     "major version difference",
			version1: "2.0.0",
			version2: "1.0.0",
			expected: true,
		},
		{
			name:     "minor version difference",
			version1: "1.1.0",
			version2: "1.0.0",
			expected: true,
		},
		{
			name:     "patch version difference",
			version1: "1.0.1",
			version2: "1.0.0",
			expected: true,
		},
		{
			name:     "pre-release vs release",
			version1: "1.0.0-alpha",
			version2: "1.0.0",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v1, ok1 := GetSemanticVersion(tt.version1)
			v2, ok2 := GetSemanticVersion(tt.version2)

			if !ok1 || !ok2 {
				t.Fatalf("Failed to parse versions: %q, %q", tt.version1, tt.version2)
			}

			result := v1.IsNewerVersionThan(v2)
			if result != tt.expected {
				t.Errorf("IsNewerVersionThan(%q, %q) = %v, want %v", tt.version1, tt.version2, result, tt.expected)
			}
		})
	}
}
