package parser

import (
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/ramonvermeulen/pre-commit-bump/config"
	"go.uber.org/zap"
)

// Repo represents a single repository configuration in the pre-commit config file.
// It contains the repository URL and the revision (branch, tag, or commit) to use
type Repo struct {
	Repo   string `yaml:"repo"`
	Rev    string `yaml:"rev"`
	SemVer *SemanticVersion
}

// GetVendor determines the vendor of the repository based on its URL.
func (r *Repo) GetVendor() string {
	vendor := ""
	if strings.Contains(r.Repo, "github.com") {
		vendor = config.VendorGitHub
	} else if strings.Contains(r.Repo, "gitlab.com") {
		vendor = config.VendorGitLab
	}
	return vendor
}

// PreCommitConfig represents the entire pre-commit configuration file.
// It contains a slice of Repo structs, each representing a repository configuration.
type PreCommitConfig struct {
	Repos  []Repo `yaml:"repos"`
	logger *zap.Logger
}

// Validate checks the PreCommitConfig for required fields and valid values.
// It returns an error if any validation fails.
func (config *PreCommitConfig) Validate() error {
	if len(config.Repos) == 0 {
		return fmt.Errorf("no repositories found in config")
	}

	for _, repo := range config.Repos {
		if repo.Repo == "" {
			return fmt.Errorf("repository URL is empty")
		}
		if repo.Rev == "" {
			return fmt.Errorf("revision is empty for repository %s", repo.Repo)
		}
	}

	return nil
}

// PopulateSemVer populates the SemVer field of each Repo in the PreCommitConfig.
// It parses the Rev field of each Repo and sets the SemVer field if the revision is a valid semantic version.
func (config *PreCommitConfig) PopulateSemVer() {
	for i := range config.Repos {
		if semVer, ok := GetSemanticVersion(config.Repos[i].Rev); ok {
			config.Repos[i].SemVer = semVer
		}
	}
}

// ValidRepos filters out sentinel values from the Repos slice and returns a slice of valid Repo structs.
// Sentinel values are "local" and "meta", which are not considered valid repositories.
// This function is useful for excluding certain repositories that are not meant to be processed.
func (config *PreCommitConfig) ValidRepos() []Repo {
	var validRepos []Repo

	sentinelValues := []string{"local", "meta"}
	for _, repo := range config.Repos {
		if slices.Contains(sentinelValues, repo.Repo) {
			config.logger.Sugar().Debugf("Skipping sentinel repo: %s", repo.Repo)
			continue
		}
		if repo.SemVer == nil {
			config.logger.Sugar().Debugf("Skipping repo with invalid semantic version: %s, rev: %s", repo.Repo, repo.Rev)
			continue
		}
		validRepos = append(validRepos, repo)
	}

	config.logger.Sugar().Debugf("total_repos: %d, valid_repos: %d", len(config.Repos), len(validRepos))

	return validRepos
}

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

func getGroup(re *regexp.Regexp, match []string, name string) string {
	index := re.SubexpIndex(name)
	if index == -1 || index >= len(match) {
		return ""
	}
	return match[index]
}
