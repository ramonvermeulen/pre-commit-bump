package bumper

import (
	"encoding/json"
	"fmt"
	"net/http"
	url2 "net/url"
	"os"
	"strings"

	"github.com/ramonvermeulen/pre-commit-bump/config"

	"github.com/ramonvermeulen/pre-commit-bump/core/types"
)

// GitLabBumper is a struct that implements the RepoBumper interface for GitLab repositories.
type GitLabBumper struct {
	client *http.Client
}

// NewGitLabBumper creates a new instance of GitLabBumper with the provided HTTP client.
func NewGitLabBumper(client *http.Client) *GitLabBumper {
	return &GitLabBumper{
		client: client,
	}
}

// GitLabTag represents a tag in a GitLab repository.
type GitLabTag struct {
	Ref string `json:"name"`
}

// GetTagName returns the tag name from the GitLabTag struct.
func (gt GitLabTag) GetTagName() string {
	return gt.Ref
}

// GetLatestVersion retrieves the latest semantic version from a GitLab repository.
// It takes the repository URL as input, fetches the tags using the GitLab API,
// and returns the latest semantic version found or an error if no valid semantic versions are present.
func (g *GitLabBumper) GetLatestVersion(repo *types.Repo) (*types.SemanticVersion, error) {
	gitlabRepo := extractGitLabRepo(repo.Repo)
	url := fmt.Sprintf("https://%s/api/v4/projects/%s/repository/tags", config.VendorGitLabHost, url2.PathEscape(gitlabRepo))

	tags, err := g.fetchTags(url)
	if err != nil {
		return nil, err
	}

	return findLatestVersion(tags, repo)
}

// fetchTags retrieves the tags from a GitLab repository using the GitLab API.
// It returns a slice of GitLabTag or an error if the API call fails.
func (g *GitLabBumper) fetchTags(url string) ([]GitLabTag, error) {
	resp, err := g.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to call GitLab API: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to close response body: %v\n", closeErr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitLab API returned status %d", resp.StatusCode)
	}

	var tags []GitLabTag
	if err := json.NewDecoder(resp.Body).Decode(&tags); err != nil {
		return nil, fmt.Errorf("failed to decode GitLab API response: %w", err)
	}

	return tags, nil
}

// extractGitLabRepo extracts the owner and repository name from a GitLab repository URL.
func extractGitLabRepo(repoURL string) string {
	repoURL = strings.TrimSuffix(repoURL, ".git")
	repoURL = strings.TrimSuffix(repoURL, "/")

	var repoPath string

	if idx := strings.Index(repoURL, config.VendorGitLabHost); idx != -1 {
		// +1 works for both https and ssh because of the `:` or `/` after the host
		repoPath = repoURL[idx+len(config.VendorGitHubHost)+1:]
	}

	return repoPath
}
