package types

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/ramonvermeulen/pre-commit-bump/core/utils"

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

	major, err1 := strconv.Atoi(utils.GetGroup(re, match, "major"))
	minor, err2 := strconv.Atoi(utils.GetGroup(re, match, "minor"))
	patch, err3 := strconv.Atoi(utils.GetGroup(re, match, "patch"))
	preRelease := utils.GetGroup(re, match, "prerelease")
	buildMetadata := utils.GetGroup(re, match, "buildmetadata")

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

// IsNewerVersionThan compares the newVersion SemanticVersion with another SemanticVersion.
// It returns true if the newVersion version is newer than the currentVersion version, false otherwise.
func (s *SemanticVersion) IsNewerVersionThan(other *SemanticVersion) bool {
	if s == nil || other == nil {
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

// GetBumpType determines the type of version bump between the newVersion SemanticVersion and another SemanticVersion.
// It returns "major", "minor", or "patch" if the newVersion version is newer than the currentVersion version.
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

// IsAllowedBumpFrom checks if the newVersion SemanticVersion is allowed to be bumped from the currentVersion SemanticVersion
// based on the allowed bump type. It returns true if the bump is allowed, false otherwise.
// allowedBumpType can be "major", "minor", or "patch".
func (s *SemanticVersion) IsAllowedBumpFrom(other *SemanticVersion, allowedBumpType string) bool {
	if other == nil || s == nil {
		return false
	}

	bumpType := s.GetBumpType(other)

	switch allowedBumpType {
	case "major":
		return bumpType == "major" || bumpType == "minor" || bumpType == "patch"
	case "minor":
		return bumpType == "minor" || bumpType == "patch"
	case "patch":
		return bumpType == "patch"
	}

	return false
}
