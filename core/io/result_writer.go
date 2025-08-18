package io

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/ramonvermeulen/pre-commit-bump/core/types"

	"go.uber.org/zap"
)

// ResultWriter handles writing results of the pre-commit updates to files.
// It provides methods to write a summary of the updates and to update the pre-commit configuration file itself.
// It uses a FileSystem interface to abstract file operations, allowing for easier testing and mocking.
type ResultWriter struct {
	fs     FileSystem
	logger *zap.Logger
}

// NewResultWriter creates a new ResultWriter instance
func NewResultWriter(fs FileSystem, logger *zap.Logger) *ResultWriter {
	return &ResultWriter{
		fs:     fs,
		logger: logger,
	}
}

// WriteSummary generates a summary of the updates and writes it to a markdown file
func (s *ResultWriter) WriteSummary(results []types.UpdateResult, allowLevel string) error {
	summaryPath := "summary.md"

	var buf strings.Builder
	buf.WriteString("# Pre-commit Hook Update Summary\n\n")
	buf.WriteString(fmt.Sprintf("**Update Policy**: Only %s version updates are allowed\n\n", allowLevel))

	updatesApplied := 0
	upToDate := 0
	constrainedUpdates := 0

	for _, result := range results {
		if result.UpdateRequired {
			buf.WriteString(fmt.Sprintf("- 🔄 **%s**: %s → %s\n",
				result.Repo.Repo, result.Repo.Rev, result.LatestVersion.String()))
			updatesApplied++
		} else {
			if result.LatestVersion != nil && result.Repo.SemVer != nil {
				if result.LatestVersion.IsNewerVersionThan(result.Repo.SemVer) {
					buf.WriteString(fmt.Sprintf("- ⚠️ **%s**: %s (newer version %s available but not allowed by %s policy)\n",
						result.Repo.Repo, result.Repo.Rev, result.LatestVersion.String(), allowLevel))
					constrainedUpdates++
				} else {
					buf.WriteString(fmt.Sprintf("- ✅ **%s**: %s (up to date)\n",
						result.Repo.Repo, result.Repo.Rev))
					upToDate++
				}
			} else {
				buf.WriteString(fmt.Sprintf("- ✅ **%s**: %s (up to date)\n",
					result.Repo.Repo, result.Repo.Rev))
				upToDate++
			}
		}
	}

	buf.WriteString("---\n\n")
	buf.WriteString("## Summary\n\n")
	buf.WriteString(fmt.Sprintf("- 🔄 **%d** hooks updated\n", updatesApplied))
	buf.WriteString(fmt.Sprintf("- ✅ **%d** hooks up to date\n", upToDate))
	if constrainedUpdates > 0 {
		buf.WriteString(fmt.Sprintf("- ⚠️ **%d** hooks have newer versions available (blocked by %s policy)\n", constrainedUpdates, allowLevel))
	}

	return s.fs.WriteFile(summaryPath, []byte(buf.String()), 0644)
}

// WritePreCommitChanges updates the pre-commit configuration file with the latest versions
func (s *ResultWriter) WritePreCommitChanges(configPath string, results []types.UpdateResult) error {
	data, err := s.fs.ReadFile(configPath)
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

		pattern := fmt.Sprintf(`(repo:\s+%s\s+rev:\s+?[a-zA-Z]?)%s`, repoURL, currentRev)
		replacement := fmt.Sprintf("${1}%s", newRev)
		re := regexp.MustCompile(pattern)
		content = re.ReplaceAllString(content, replacement)

		s.logger.Sugar().Debugf("Updated %s from %s to %s", result.Repo.Repo, result.Repo.Rev, newRev)
	}

	return s.fs.WriteFile(configPath, []byte(content), 0644)
}
