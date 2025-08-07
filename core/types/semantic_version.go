package types

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/ramonvermeulen/pre-commit-bump/config"
)

// SemanticVersion represents a semantic version with major, minor, patch, and optional pre-release and build metadata components.
type SemanticVersion struct {
	Major         int
	Minor         int
	Patch         int
	PreRelease    string
	BuildMetaData string
}

// GetSemanticVersion parses a version string and return a SemanticVersion struct if it matches the semantic versioning format.
func GetSemanticVersion(version string) (*SemanticVersion, bool) {
	re := regexp.MustCompile(config.ReSemanticVersion)
	match := re.FindStringSubmatch(version)
	if match == nil {
		return &SemanticVersion{}, false
	}

	major, err1 := strconv.Atoi(getGroup(re, match, "major"))
	minor, err2 := strconv.Atoi(getGroup(re, match, "minor"))
	patch, err3 := strconv.Atoi(getGroup(re, match, "patch"))
	preRelease := getGroup(re, match, "prerelease")
	buildMetadata := getGroup(re, match, "buildmetadata")

	if err1 != nil || err2 != nil || err3 != nil {
		return &SemanticVersion{}, false
	}

	return &SemanticVersion{
		Major:         major,
		Minor:         minor,
		Patch:         patch,
		PreRelease:    preRelease,
		BuildMetaData: buildMetadata,
	}, true
}

// String returns the string representation of the SemanticVersion in the format "major.minor.patch-preRelease+buildMetaData".
func (s *SemanticVersion) String() string {
	version := fmt.Sprintf("%d.%d.%d", s.Major, s.Minor, s.Patch)
	if s.PreRelease != "" {
		version += "-" + s.PreRelease
	}
	if s.BuildMetaData != "" {
		version += "+" + s.BuildMetaData
	}
	return version
}

// IsNewerVersionThan compares the current SemanticVersion with another SemanticVersion.
// It returns true if the current version is newer than the other version, false otherwise.
func (s *SemanticVersion) IsNewerVersionThan(other *SemanticVersion) bool {
	if other == nil {
		return false
	}

	if s.Major > other.Major {
		return true
	}
	if s.Major == other.Major && s.Minor > other.Minor {
		return true
	}
	if s.Major == other.Major && s.Minor == other.Minor && s.Patch > other.Patch {
		return true
	}

	return false
}

// GetBumpType determines the type of version bump between the current SemanticVersion and another SemanticVersion.
// It returns "major", "minor", or "patch" if the current version is newer than the other version.
func (s *SemanticVersion) GetBumpType(other *SemanticVersion) string {
	if other == nil {
		return ""
	}

	if s.Major > other.Major {
		return "major"
	}
	if s.Major == other.Major && s.Minor > other.Minor {
		return "minor"
	}
	if s.Major == other.Major && s.Minor == other.Minor && s.Patch > other.Patch {
		return "patch"
	}

	return ""
}

// IsAllowedBump checks if the current SemanticVersion is allowed to be bumped to the other SemanticVersion
// based on the allowed bump type. It returns true if the bump is allowed, false otherwise.
func (s *SemanticVersion) IsAllowedBump(other *SemanticVersion, allowedBumpType string) bool {
	if allowedBumpType == "major" {
		return true
	}

	bumpType := s.GetBumpType(other)

	switch allowedBumpType {
	case "minor":
		return bumpType == "minor" || bumpType == "patch"
	case "patch":
		return bumpType == "patch"
	}

	return false
}

func getGroup(re *regexp.Regexp, match []string, name string) string {
	index := re.SubexpIndex(name)
	if index == -1 || index >= len(match) {
		return ""
	}
	return match[index]
}
