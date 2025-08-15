package bumper

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/ramonvermeulen/pre-commit-bump/core/types"

	"github.com/ramonvermeulen/pre-commit-bump/config"
	"github.com/ramonvermeulen/pre-commit-bump/core/io"
	"github.com/ramonvermeulen/pre-commit-bump/core/parser"
)

// RepoBumper defines the interface for updating repositories.
// To support different repository types, implement this interface (e.g., GitHub, GitLab).
type RepoBumper interface {
	GetLatestVersion(repo *types.Repo) (*types.SemanticVersion, error)
}

// TagProvider defines an interface for types that can provide a tag name.
// such as GitHubTag or GitLabTag.
type TagProvider interface {
	GetTagName() string
}

// Bumper coordinates the pre-commit hook bumping process.
type Bumper struct {
	parser     *parser.Parser
	cfg        *config.Config
	fileWriter *io.ResultWriter
	httpClient *http.Client
}

// NewBumper creates a new Bumper instance with dependency injection
func NewBumper(parser *parser.Parser, cfg *config.Config, fileWriter *io.ResultWriter, httpClient *http.Client) *Bumper {
	return &Bumper{
		parser:     parser,
		cfg:        cfg,
		fileWriter: fileWriter,
		httpClient: httpClient,
	}
}

// parsePreCommitConfig parses the pre-commit configuration file and logs the action.
func (b *Bumper) parsePreCommitConfig() (*types.PreCommitConfig, error) {
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

// checkReposForUpdates iterates through the repositories in the pre-commit configuration
// and checks for updates using the appropriate RepoBumper based on the vendor.
// it uses a goroutine for each repository to perform the check concurrently.
func (b *Bumper) checkReposForUpdates(repos []types.Repo) []types.UpdateResult {
	repositoryUpdaters := map[string]RepoBumper{
		config.VendorGitHub: NewGithubBumper(b.httpClient),
		config.VendorGitLab: NewGitLabBumper(b.httpClient),
	}

	updateResults := make([]types.UpdateResult, len(repos))
	var waitGroup sync.WaitGroup

	for repoIndex, currentRepo := range repos {
		vendor := currentRepo.GetVendor()
		updater, vendorSupported := repositoryUpdaters[vendor]

		if !vendorSupported {
			b.cfg.Logger.Sugar().Warnf("No updater found for vendor: %s, skipping repo: %s", vendor, currentRepo.Repo)
			updateResults[repoIndex] = types.UpdateResult{
				Repo:  currentRepo,
				Error: fmt.Errorf("no updater found for vendor: %s", vendor),
			}
			continue
		}

		waitGroup.Add(1)
		go b.checkRepoAsync(&waitGroup, updateResults, repoIndex, currentRepo, updater)
	}

	waitGroup.Wait()

	return updateResults
}

// checkRepoAsync checks a single repository for updates and is intended to be called concurrently as a goroutine.
func (b *Bumper) checkRepoAsync(waitGroup *sync.WaitGroup, results []types.UpdateResult, index int, repo types.Repo, updater RepoBumper) {
	defer waitGroup.Done()
	results[index] = b.checkSingleRepo(repo, updater)
}

// checkSingleRepo checks a single repository for updates.
// It retrieves the latest version using the provided RepoBumper and compares it with the current version.
func (b *Bumper) checkSingleRepo(repo types.Repo, updater RepoBumper) types.UpdateResult {
	b.cfg.Logger.Sugar().Debugf("Checking repo: %s, current version: %s", repo.Repo, repo.Rev)

	latestVersion, err := updater.GetLatestVersion(&repo)
	if err != nil {
		return types.UpdateResult{
			Repo:  repo,
			Error: fmt.Errorf("failed to get latest version for %s: %w", repo.Repo, err),
		}
	}

	updateRequired := latestVersion.IsAllowedBumpFrom(repo.SemVer, b.cfg.Allow)

	if latestVersion.IsNewerVersionThan(repo.SemVer) && !updateRequired {
		bumpType := latestVersion.GetBumpType(repo.SemVer)
		b.cfg.Logger.Sugar().Debugf("Update available for %s (%s -> %s) but %s bump not allowed (only %s allowed)",
			repo.Repo, repo.Rev, latestVersion.String(), bumpType, b.cfg.Allow)
	}

	return types.UpdateResult{
		Repo:           repo,
		LatestVersion:  latestVersion,
		UpdateRequired: updateRequired,
	}
}

// processResults handles common error checking and logging
// returns a boolean indicating if updates are available in any of the hooks or an error if any occurred.
func (b *Bumper) processResults(results []types.UpdateResult) (bool, error) {
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
				result.Repo.Repo, result.Repo.Rev, result.LatestVersion.String())
		}
	}

	if len(errs) > 0 {
		return false, fmt.Errorf("errors occurred while checking repositories: %v", errs)
	}

	return hasUpdates, nil
}

// processCheckResults processes the results of the check for updates.
// It checks if any updates are available and returns an error if so.
func (b *Bumper) processCheckResults(results []types.UpdateResult) error {
	hasUpdates, err := b.processResults(results)
	if err != nil {
		return err
	}

	if hasUpdates {
		return fmt.Errorf("updates are available")
	}
	return nil
}

// processUpdateResults processes the results of the update check.
// It writes the changes to the pre-commit configuration file and generates a summary if requested.
func (b *Bumper) processUpdateResults(results []types.UpdateResult) error {
	hasUpdates, err := b.processResults(results)
	if err != nil {
		return err
	}

	if hasUpdates && !b.cfg.DryRun {
		err := b.fileWriter.WritePreCommitChanges(b.cfg.PreCommitConfigPath, results)
		if err != nil {
			return fmt.Errorf("failed to write pre-commit changes: %w", err)
		}
		b.cfg.Logger.Sugar().Info("Pre-commit configuration file updated successfully")

		if !b.cfg.NoSummary {
			err = b.fileWriter.WriteSummary(results)
			if err != nil {
				return fmt.Errorf("failed to write summary: %w", err)
			}
			b.cfg.Logger.Sugar().Info("Summary file created successfully")
		} else {
			b.cfg.Logger.Sugar().Info("No summary generation requested, skipping summary file creation")
		}
	} else if b.cfg.DryRun {
		b.cfg.Logger.Sugar().Info("Dry run mode enabled, will not modify the pre-commit-config.yaml file or create a summary")
	}

	return nil
}

// findLatestVersion iterating through the Vendor tags to find the latest semantic version.
// It returns the latest version found or an error if no valid semantic versions are present.
func findLatestVersion[T TagProvider](tags []T, repo *types.Repo) (*types.SemanticVersion, error) {
	var latest *types.SemanticVersion

	for _, tag := range tags {
		semVer, ok := types.GetSemanticVersion(tag.GetTagName())
		if !ok {
			continue
		}

		if latest == nil || semVer.IsNewerVersionThan(latest) {
			latest = semVer
		}
	}

	if latest == nil {
		return nil, fmt.Errorf("no semantic version tags found for repo: %s with rev: %s", repo.Repo, repo.Rev)
	}

	return latest, nil
}
