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

func TestSemanticVersionIsAllowedBump(t *testing.T) {
	tests := []struct {
		name           string
		newVersion     string
		currentVersion string
		allowedType    string
		expected       bool
		description    string
	}{
		{
			name:           "major allowed - major bump",
			newVersion:     "2.0.0",
			currentVersion: "1.0.0",
			allowedType:    "major",
			expected:       true,
			description:    "major bump should be allowed when major is allowed",
		},
		{
			name:           "major allowed - minor bump",
			newVersion:     "1.1.0",
			currentVersion: "1.0.0",
			allowedType:    "major",
			expected:       true,
			description:    "minor bump should be allowed when major is allowed",
		},
		{
			name:           "major allowed - patch bump",
			newVersion:     "1.0.1",
			currentVersion: "1.0.0",
			allowedType:    "major",
			expected:       true,
			description:    "patch bump should be allowed when major is allowed",
		},
		{
			name:           "minor allowed - major bump",
			newVersion:     "2.0.0",
			currentVersion: "1.0.0",
			allowedType:    "minor",
			expected:       false,
			description:    "major bump should not be allowed when only minor is allowed",
		},
		{
			name:           "minor allowed - minor bump",
			newVersion:     "1.1.0",
			currentVersion: "1.0.0",
			allowedType:    "minor",
			expected:       true,
			description:    "minor bump should be allowed when minor is allowed",
		},
		{
			name:           "minor allowed - patch bump",
			newVersion:     "1.0.1",
			currentVersion: "1.0.0",
			allowedType:    "minor",
			expected:       true,
			description:    "patch bump should be allowed when minor is allowed",
		},
		{
			name:           "patch allowed - major bump",
			newVersion:     "2.0.0",
			currentVersion: "1.0.0",
			allowedType:    "patch",
			expected:       false,
			description:    "major bump should not be allowed when only patch is allowed",
		},
		{
			name:           "patch allowed - minor bump",
			newVersion:     "1.1.0",
			currentVersion: "1.0.0",
			allowedType:    "patch",
			expected:       false,
			description:    "minor bump should not be allowed when only patch is allowed",
		},
		{
			name:           "patch allowed - patch bump",
			newVersion:     "1.0.1",
			currentVersion: "1.0.0",
			allowedType:    "patch",
			expected:       true,
			description:    "patch bump should be allowed when patch is allowed",
		},
		{
			name:           "nil currentVersion version",
			newVersion:     "1.0.0",
			currentVersion: "",
			allowedType:    "major",
			expected:       false,
			description:    "should return false when currentVersion version is nil",
		},
		{
			name:           "same versions",
			newVersion:     "1.0.0",
			currentVersion: "1.0.0",
			allowedType:    "major",
			expected:       false,
			description:    "should return false when versions are the same",
		},
		{
			name:           "downgrade version",
			newVersion:     "1.0.0",
			currentVersion: "2.0.0",
			allowedType:    "major",
			expected:       false,
			description:    "should return false when newVersion is older than currentVersion",
		},
		{
			name:           "invalid allowed type",
			newVersion:     "1.1.0",
			currentVersion: "1.0.0",
			allowedType:    "invalid",
			expected:       false,
			description:    "should return false for invalid allowed bump type",
		},
		{
			name:           "pre-release to release patch",
			newVersion:     "1.0.1",
			currentVersion: "1.0.0-alpha",
			allowedType:    "patch",
			expected:       true,
			description:    "patch bump from pre-release should be allowed",
		},
		{
			name:           "pre-release to release minor not allowed",
			newVersion:     "1.1.0",
			currentVersion: "1.0.0-beta",
			allowedType:    "patch",
			expected:       false,
			description:    "minor bump from pre-release should not be allowed when only patch allowed",
		},
		{
			name:           "large version major bump",
			newVersion:     "100.0.0",
			currentVersion: "99.999.999",
			allowedType:    "major",
			expected:       true,
			description:    "major bump with large version numbers should work",
		},
		{
			name:           "large version minor bump restricted",
			newVersion:     "99.1000.0",
			currentVersion: "99.999.999",
			allowedType:    "patch",
			expected:       false,
			description:    "minor bump with large numbers should be restricted when only patch allowed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			current, okCurrent := GetSemanticVersion(tt.newVersion)
			if !okCurrent {
				t.Fatalf("Failed to parse newVersion version: %q", tt.newVersion)
			}

			var other *SemanticVersion
			if tt.currentVersion != "" {
				var okOther bool
				other, okOther = GetSemanticVersion(tt.currentVersion)
				if !okOther {
					t.Fatalf("Failed to parse currentVersion version: %q", tt.currentVersion)
				}
			}

			result := current.IsAllowedBumpFrom(other, tt.allowedType)
			if result != tt.expected {
				t.Errorf("IsAllowedBumpFrom(%q, %q, %q) = %v, want %v - %s",
					tt.newVersion, tt.currentVersion, tt.allowedType, result, tt.expected, tt.description)
			}
		})
	}
}
