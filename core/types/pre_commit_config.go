package types

import (
	"fmt"
	"slices"
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
	if strings.Contains(r.Repo, config.VendorGitHubHost) {
		vendor = config.VendorGitHub
	} else if strings.Contains(r.Repo, config.VendorGitLabHost) {
		vendor = config.VendorGitLab
	}
	return vendor
}

// PreCommitConfig represents the entire pre-commit configuration file.
// It contains a slice of Repo structs, each representing a repository configuration.
type PreCommitConfig struct {
	Repos  []Repo `yaml:"repos"`
	Logger *zap.Logger
}

// Validate checks the PreCommitConfig for required fields and valid values.
// It returns an error if any validation fails.
func (c *PreCommitConfig) Validate() error {
	sentinelValues := []string{config.SentinelLocal, config.SentinelMeta}
	if len(c.Repos) == 0 {
		return fmt.Errorf("no repositories found in config")
	}

	for _, repo := range c.Repos {
		if repo.Repo == "" {
			return fmt.Errorf("repository URL is empty")
		}
		if !slices.Contains(sentinelValues, repo.Repo) {
			if repo.Rev == "" {
				return fmt.Errorf("revision is empty for repository: %s", repo.Repo)
			}
		}
	}

	return nil
}

// PopulateSemVer populates the SemVer field of each Repo in the PreCommitConfig.
// It parses the Rev field of each Repo and sets the SemVer field if the revision is a valid semantic version.
func (c *PreCommitConfig) PopulateSemVer() {
	for i := range c.Repos {
		if semVer, ok := GetSemanticVersion(c.Repos[i].Rev); ok {
			c.Repos[i].SemVer = semVer
		}
	}
}

// ValidRepos filters out sentinel values from the Repos slice and returns a slice of valid Repo structs.
// Sentinel values are "local" and "meta", which are not considered valid repositories.
// This function is useful for excluding certain repositories that are not meant to be processed.
func (c *PreCommitConfig) ValidRepos() []Repo {
	var validRepos []Repo

	sentinelValues := []string{config.SentinelMeta, config.SentinelLocal}
	for _, repo := range c.Repos {
		if slices.Contains(sentinelValues, repo.Repo) {
			c.Logger.Sugar().Debugf("Skipping sentinel repo: %s", repo.Repo)
			continue
		}
		if repo.SemVer == nil {
			c.Logger.Sugar().Debugf("Skipping repo with invalid semantic version: %s, rev: %s", repo.Repo, repo.Rev)
			continue
		}
		validRepos = append(validRepos, repo)
	}

	c.Logger.Sugar().Debugf("total_repos: %d, valid_repos: %d", len(c.Repos), len(validRepos))

	return validRepos
}
