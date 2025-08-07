package bumper

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/ramonvermeulen/pre-commit-bump/core/parser"
)

// GithubBumper is a struct that implements the RepoBumper interface for GitHub repositories.
type GithubBumper struct {
	client *http.Client
}

// NewGithubBumper creates a new instance of GithubBumper with the provided HTTP client.
func NewGithubBumper(client *http.Client) *GithubBumper {
	return &GithubBumper{
		client: client,
	}
}

// GitHubTag represents a tag in a GitHub repository.
type GitHubTag struct {
	Ref string `json:"ref"`
}

// GetTagName returns the tag name by stripping the "refs/tags/" prefix from the Ref field.
func (gt GitHubTag) GetTagName() string {
	return strings.TrimPrefix(gt.Ref, "refs/tags/")
}

// GetLatestVersion retrieves the latest semantic version from a GitHub repository.
// It takes a pointer to a parser.Repo as input, fetches the tags using the GitHub API.
// And returns the latest semantic version found or an error if no valid semantic versions are present.
func (g *GithubBumper) GetLatestVersion(repo *parser.Repo) (*parser.SemanticVersion, error) {
	repoPath := extractGitHubRepo(repo.Repo)

	tags, err := g.fetchTags(repoPath)
	if err != nil {
		return nil, err
	}

	return findLatestVersion(tags, repo)
}

// fetchTags retrieves the tags from a GitHub repository using the GitHub API.
// It returns a slice of GitHubTag or an error if the API call fails.
func (g *GithubBumper) fetchTags(repoPath string) ([]GitHubTag, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/git/refs/tags", repoPath)

	resp, err := g.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to call GitHub API: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to close response body: %v\n", closeErr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var tags []GitHubTag
	if err := json.NewDecoder(resp.Body).Decode(&tags); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return tags, nil
}

// extractGitHubRepo extracts the owner and repository name from a GitHub repository URL.
// It handles both HTTPS and SSH formats, and removes the ".git" suffix if present.
func extractGitHubRepo(repoURL string) string {
	repoURL = strings.TrimSuffix(repoURL, ".git")
	repoURL = strings.TrimSuffix(repoURL, "/")

	var repoPath string
	if idx := strings.Index(repoURL, "github.com/"); idx != -1 {
		repoPath = repoURL[idx+len("github.com/"):]
	} else if idx := strings.Index(repoURL, "github.com:"); idx != -1 {
		repoPath = repoURL[idx+len("github.com:"):]
	}

	return repoPath
}
