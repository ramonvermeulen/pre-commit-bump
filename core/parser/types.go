package parser

import (
	"fmt"
	"slices"
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
	Repos []Repo `yaml:"repos"`
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
		if !slices.Contains(sentinelValues, repo.Repo) {
			validRepos = append(validRepos, repo)
		}
	}
	return validRepos
}
