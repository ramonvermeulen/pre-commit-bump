package parser

import (
	"fmt"
	"regexp"
	"slices"
	"strconv"

	"github.com/ramonvermeulen/pre-commit-bump/config"
	"go.uber.org/zap"
)

// Repo represents a single repository configuration in the pre-commit config file.
// It contains the repository URL and the revision (branch, tag, or commit) to use
type Repo struct {
	Repo string `yaml:"repo"`
	Rev  string `yaml:"rev"`
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
		if _, ok := GetSemanticVersion(repo.Rev); !ok {
			config.logger.Sugar().Debugf("Skipping repo with invalid semantic version: %s, rev: %s", repo.Repo, repo.Rev)
			continue
		}
		validRepos = append(validRepos, repo)
	}
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
func GetSemanticVersion(version string) (SemanticVersion, bool) {
	re := regexp.MustCompile(config.ReSemanticVersion)
	match := re.FindStringSubmatch(version)
	if match == nil {
		return SemanticVersion{}, false
	}

	major, err1 := strconv.Atoi(getGroup(re, match, "major"))
	minor, err2 := strconv.Atoi(getGroup(re, match, "minor"))
	patch, err3 := strconv.Atoi(getGroup(re, match, "patch"))
	preRelease := getGroup(re, match, "prerelease")
	buildMetadata := getGroup(re, match, "buildmetadata")

	if err1 != nil || err2 != nil || err3 != nil {
		return SemanticVersion{}, false
	}

	return SemanticVersion{
		Major:         major,
		Minor:         minor,
		Patch:         patch,
		PreRelease:    preRelease,
		BuildMetaData: buildMetadata,
	}, true
}

func getGroup(re *regexp.Regexp, match []string, name string) string {
	index := re.SubexpIndex(name)
	if index == -1 || index >= len(match) {
		return ""
	}
	return match[index]
}
