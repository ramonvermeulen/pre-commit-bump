package bumper

import (
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
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

// checkReposForUpdates iterates through the repositories in the pre-commit configuration
// and checks for updates using the appropriate RepoBumper based on the vendor.
// it uses a goroutine for each repository to perform the check concurrently.
func (b *Bumper) checkReposForUpdates(repos []parser.Repo) []UpdateResult {
	repositoryUpdaters := map[string]RepoBumper{
		config.VendorGitHub: NewGithubBumper(b.client),
		config.VendorGitLab: NewGitLabBumper(b.client),
	}

	updateResults := make([]UpdateResult, len(repos))
	var waitGroup sync.WaitGroup

	for repoIndex, currentRepo := range repos {
		vendor := currentRepo.GetVendor()
		updater, vendorSupported := repositoryUpdaters[vendor]

		if !vendorSupported {
			b.cfg.Logger.Sugar().Warnf("No updater found for vendor: %s, skipping repo: %s", vendor, currentRepo.Repo)
			updateResults[repoIndex] = UpdateResult{
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
func (b *Bumper) checkRepoAsync(waitGroup *sync.WaitGroup, results []UpdateResult, index int, repo parser.Repo, updater RepoBumper) {
	defer waitGroup.Done()
	results[index] = b.checkSingleRepo(repo, updater)
}

// checkSingleRepo checks a single repository for updates.
// It retrieves the latest version using the provided RepoBumper and compares it with the current version.
func (b *Bumper) checkSingleRepo(repo parser.Repo, updater RepoBumper) UpdateResult {
	b.cfg.Logger.Sugar().Debugf("Checking repo: %s, current version: %s", repo.Repo, repo.Rev)

	latestVersion, err := updater.GetLatestVersion(&repo)
	if err != nil {
		return UpdateResult{
			Repo:  repo,
			Error: fmt.Errorf("failed to get latest version for %s: %w", repo.Repo, err),
		}
	}

	updateRequired := latestVersion.IsNewerVersionThan(repo.SemVer)

	return UpdateResult{
		Repo:           repo,
		LatestVersion:  latestVersion,
		UpdateRequired: updateRequired,
	}
}

// writeSummary generates a summary of the updates and writes it to a markdown file.
// It includes the repository name, current version, latest version, and a link to the changelog.
// If no updates are available, it indicates that in the summary.
func (b *Bumper) writeSummary(results []UpdateResult) error {
	summaryPath := "summary.md"

	var buf strings.Builder
	buf.WriteString("# Pre-commit Hook Update Summary\n\n")

	for _, result := range results {
		if result.UpdateRequired {
			buf.WriteString(fmt.Sprintf("- ✅ **%s**: %s → %s\n",
				result.Repo.Repo, result.Repo.Rev, result.LatestVersion.String()))
			buf.WriteString(fmt.Sprintf("See changelog at: %s/releases/tag/%s\n\n", result.Repo.Repo, result.LatestVersion.String()))
		} else {
			buf.WriteString(fmt.Sprintf("- ⏸️ **%s**: %s (up to date)\n",
				result.Repo.Repo, result.Repo.Rev))
		}
	}

	return os.WriteFile(summaryPath, []byte(buf.String()), 0644)
}

// writePreCommitChanges updates the pre-commit configuration file with the latest versions.
// It reads the file, replaces the old versions with the new ones, and writes the changes back to the file.
func (b *Bumper) writePreCommitChanges(results []UpdateResult) error {
	data, err := os.ReadFile(b.cfg.PreCommitConfigPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	content := string(data)

	for _, result := range results {
		if !result.UpdateRequired || result.Error != nil {
			continue
		}

		repoURL := regexp.QuoteMeta(result.Repo.Repo)
		currentRev := regexp.QuoteMeta(result.Repo.SemVer.String())
		newRev := result.LatestVersion.String()

		pattern := fmt.Sprintf(`(repo:\s+%s\s+rev:\s+)%s`, repoURL, currentRev)
		replacement := fmt.Sprintf("${1}%s", newRev)

		re := regexp.MustCompile(pattern)
		content = re.ReplaceAllString(content, replacement)

		b.cfg.Logger.Sugar().Debugf("Updated %s from %s to %s", result.Repo.Repo, result.Repo.Rev, newRev)
	}

	return os.WriteFile(b.cfg.PreCommitConfigPath, []byte(content), 0644)
}

// processResults handles common error checking and logging
// returns a boolean indicating if updates are available in any of the hooks or an error if any occurred.
func (b *Bumper) processResults(results []UpdateResult) (bool, error) {
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
func (b *Bumper) processCheckResults(results []UpdateResult) error {
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
func (b *Bumper) processUpdateResults(results []UpdateResult) error {
	hasUpdates, err := b.processResults(results)
	if err != nil {
		return err
	}

	if hasUpdates && !b.cfg.DryRun {
		err := b.writePreCommitChanges(results)
		if err != nil {
			return fmt.Errorf("failed to write pre-commit changes: %w", err)
		}
		b.cfg.Logger.Sugar().Info("Pre-commit configuration file updated successfully")

		if !b.cfg.NoSummary {
			err = b.writeSummary(results)
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

// findLatestVersion iterating through the GitHub tags to find the latest semantic version.
// It returns the latest version found or an error if no valid semantic versions are present.
func findLatestVersion[T TagProvider](tags []T, repo *parser.Repo) (*parser.SemanticVersion, error) {
	var latest *parser.SemanticVersion

	for _, tag := range tags {
		semVer, ok := parser.GetSemanticVersion(tag.GetTagName())
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
