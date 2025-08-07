package bumper

import (
	"fmt"
	"net/http"
	"time"

	"github.com/ramonvermeulen/pre-commit-bump/config"
	"github.com/ramonvermeulen/pre-commit-bump/core/parser"
)

// RepoBumper defines the interface for updating repositories.
// To support different repository types, implement this interface (e.g., GitHub, GitLab).
type RepoBumper interface {
	GetLatestVersion(repo *parser.Repo) (*parser.SemanticVersion, error)
}

// TagProvider defines an interface for types that can provide a tag name.
// such as GitHubTag or GitLabTag.
type TagProvider interface {
	GetTagName() string
}

// UpdateResult holds the result of checking a repository for updates.
type UpdateResult struct {
	Repo           parser.Repo
	LatestVersion  *parser.SemanticVersion
	UpdateRequired bool
	Error          error
}

// Bumper coordinates the pre-commit hook bumping process.
type Bumper struct {
	parser *parser.Parser
	cfg    *config.Config
	client *http.Client
}

// NewBumper creates a new Bumper instance with the given parser and cfg
func NewBumper(parser *parser.Parser, cfg *config.Config) *Bumper {
	return &Bumper{
		parser: parser,
		cfg:    cfg,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

// parsePreCommitConfig parses the pre-commit configuration file and logs the action.
func (b *Bumper) parsePreCommitConfig() (*parser.PreCommitConfig, error) {
	b.cfg.Logger.Sugar().Debugf("Parsing configuration file: %s", b.cfg.PreCommitConfigPath)

	pCfg, err := b.parser.ParseConfig(b.cfg.PreCommitConfigPath)
	if err != nil {
		return nil, err
	}

	return pCfg, nil
}

// Check verifies if the pre-commit configuration file is valid and up-to-date.
// If the configuration is valid, it returns nil.
// If there are updates available, it returns an error.
func (b *Bumper) Check() error {
	pCfg, err := b.parsePreCommitConfig()
	if err != nil {
		return fmt.Errorf("failed to parse pre-commit configuration: %w", err)
	}

	results := b.checkReposForUpdates(pCfg.ValidRepos())

	return b.processCheckResults(results)
}

// Update checks for available updates and modifies the pre-commit configuration file.
func (b *Bumper) Update() error {
	pCfg, err := b.parsePreCommitConfig()
	if err != nil {
		return fmt.Errorf("failed to parse pre-commit configuration: %w", err)
	}

	results := b.checkReposForUpdates(pCfg.ValidRepos())

	return b.processUpdateResults(results)
}

func (b *Bumper) checkReposForUpdates(repos []parser.Repo) []UpdateResult {
	updaters := map[string]RepoBumper{
		config.VendorGitHub: NewGithubBumper(b.client),
		config.VendorGitLab: NewGitLabBumper(b.client),
	}
	results := make([]UpdateResult, 0, len(repos))

	for _, repo := range repos {
		vendor := repo.GetVendor()
		updater, exists := updaters[vendor]
		if !exists {
			b.cfg.Logger.Sugar().Warnf("No updater found for vendor: %s, skipping repo: %s", vendor, repo.Repo)
			continue
		}
		result := b.checkSingleRepo(repo, updater)
		results = append(results, result)
	}

	return results
}

func (b *Bumper) checkSingleRepo(repo parser.Repo, updater RepoBumper) UpdateResult {
	b.cfg.Logger.Sugar().Debugf("Checking repo: %s, current version: %s", repo.Repo, repo.Rev)

	latestVersion, err := updater.GetLatestVersion(&repo)
	if err != nil {
		return UpdateResult{
			Repo:  repo,
			Error: fmt.Errorf("failed to get latest version for %s: %w", repo.Repo, err),
		}
	}

	updateRequired := isNewerVersion(repo.SemVer, latestVersion)

	return UpdateResult{
		Repo:           repo,
		LatestVersion:  latestVersion,
		UpdateRequired: updateRequired,
	}
}

func isNewerVersion(current, latest *parser.SemanticVersion) bool {
	if current == nil || latest == nil {
		return false
	}

	if latest.Major > current.Major {
		return true
	}
	if latest.Major == current.Major && latest.Minor > current.Minor {
		return true
	}
	if latest.Major == current.Major && latest.Minor == current.Minor && latest.Patch > current.Patch {
		return true
	}

	return false
}

func (b *Bumper) processCheckResults(results []UpdateResult) error {
	var hasUpdates bool
	var errs []error

	for _, result := range results {
		if result.Error != nil {
			b.cfg.Logger.Sugar().Warnf("Error checking %s: %v", result.Repo.Repo, result.Error)
			errs = append(errs, result.Error)
			continue
		}

		if result.UpdateRequired {
			hasUpdates = true
			b.cfg.Logger.Sugar().Infof("Update available for %s: %s -> %s",
				result.Repo.Repo, result.Repo.Rev, formatVersion(result.LatestVersion))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors occurred while checking repositories: %v", errs)
	}

	if hasUpdates {
		return fmt.Errorf("updates are available")
	}

	return nil
}

func (b *Bumper) processUpdateResults(results []UpdateResult) error {
	// todo error handling and logging

	if b.cfg.DryRun {
		// log about dry-run
		b.cfg.Logger.Sugar().Debugf("not implemented yet")
	} else {
		// write config changes back to file
		b.cfg.Logger.Sugar().Debugf("not implemented yet")
	}

	if b.cfg.NoSummary {
		b.cfg.Logger.Sugar().Info("No summary generation requested, skipping summary file creation")
	} else {
		// generate summary file
		b.cfg.Logger.Sugar().Debugf("not implemented yet")
	}

	return nil
}

func formatVersion(v *parser.SemanticVersion) string {
	if v == nil {
		return "unknown"
	}
	version := fmt.Sprintf("v%d.%d.%d", v.Major, v.Minor, v.Patch)
	if v.PreRelease != "" {
		version += "-" + v.PreRelease
	}
	if v.BuildMetaData != "" {
		version += "+" + v.BuildMetaData
	}
	return version
}

// findLatestVersion iterating through the GitHub tags to find the latest semantic version.
// It returns the latest version found or an error if no valid semantic versions are present.
func findLatestVersionGeneric[T TagProvider](tags []T, repo *parser.Repo) (*parser.SemanticVersion, error) {
	var latest *parser.SemanticVersion

	for _, tag := range tags {
		semVer, ok := parser.GetSemanticVersion(tag.GetTagName())
		if !ok {
			continue
		}

		if latest == nil || isNewerVersion(latest, &semVer) {
			latest = &semVer
		}
	}

	if latest == nil {
		return nil, fmt.Errorf("no semantic version tags found for repo: %s with rev: %s", repo.Repo, repo.Rev)
	}

	return latest, nil
}
