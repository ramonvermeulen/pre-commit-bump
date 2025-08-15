package bumper

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"

	"github.com/ramonvermeulen/pre-commit-bump/config"
	"github.com/ramonvermeulen/pre-commit-bump/core/types"
)

// MockRepoBumper is a testify mock for the RepoBumper interface
type MockRepoBumper struct {
	mock.Mock
}

func (m *MockRepoBumper) GetLatestVersion(repo *types.Repo) (*types.SemanticVersion, error) {
	args := m.Called(repo)
	return args.Get(0).(*types.SemanticVersion), args.Error(1)
}

func TestBumper_checkSingleRepo(t *testing.T) {
	tests := []struct {
		name           string
		repo           types.Repo
		latestVersion  *types.SemanticVersion
		updaterError   error
		allowedBump    string
		expectedUpdate bool
		expectedError  bool
	}{
		{
			name: "update available and allowed",
			repo: types.Repo{
				Repo:   "https://github.com/owner/repo",
				Rev:    "1.0.0",
				SemVer: &types.SemanticVersion{Major: 1, Minor: 0, Patch: 0},
			},
			latestVersion:  &types.SemanticVersion{Major: 1, Minor: 1, Patch: 0},
			allowedBump:    "minor",
			expectedUpdate: true,
			expectedError:  false,
		},
		{
			name: "update available but not allowed",
			repo: types.Repo{
				Repo:   "https://github.com/owner/repo",
				Rev:    "1.0.0",
				SemVer: &types.SemanticVersion{Major: 1, Minor: 0, Patch: 0},
			},
			latestVersion:  &types.SemanticVersion{Major: 2, Minor: 0, Patch: 0},
			allowedBump:    "minor",
			expectedUpdate: false,
			expectedError:  false,
		},
		{
			name: "updater returns error",
			repo: types.Repo{
				Repo:   "https://github.com/owner/repo",
				Rev:    "1.0.0",
				SemVer: &types.SemanticVersion{Major: 1, Minor: 0, Patch: 0},
			},
			updaterError:   fmt.Errorf("API error"),
			expectedUpdate: false,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUpdater := new(MockRepoBumper)

			if tt.updaterError != nil {
				mockUpdater.On("GetLatestVersion", &tt.repo).Return((*types.SemanticVersion)(nil), tt.updaterError)
			} else {
				mockUpdater.On("GetLatestVersion", &tt.repo).Return(tt.latestVersion, nil)
			}

			cfg := &config.Config{
				Allow:  tt.allowedBump,
				Logger: zap.NewNop(),
			}
			bumper := &Bumper{cfg: cfg}

			result := bumper.checkSingleRepo(tt.repo, mockUpdater)

			if tt.expectedError {
				assert.Error(t, result.Error, "Expected error but got none")
			} else {
				assert.NoError(t, result.Error, "Unexpected error: %v", result.Error)
			}

			assert.Equal(t, tt.expectedUpdate, result.UpdateRequired, "UpdateRequired mismatch")
			assert.Equal(t, tt.repo, result.Repo, "Repo should match")

			if !tt.expectedError && tt.latestVersion != nil {
				assert.Equal(t, tt.latestVersion, result.LatestVersion, "LatestVersion should match")
			}

			mockUpdater.AssertExpectations(t)
		})
	}
}

func TestFindLatestVersionGitHub(t *testing.T) {
	tests := []struct {
		name        string
		tags        []GitHubTag
		repo        *types.Repo
		expectedVer *types.SemanticVersion
		expectError bool
	}{
		{
			name:        "empty tag list",
			tags:        []GitHubTag{},
			repo:        &types.Repo{Repo: "test/repo", Rev: "1.0.0"},
			expectError: true,
		},
		{
			name: "no valid semantic version tags",
			tags: []GitHubTag{
				{Ref: "refs/tags/invalid-tag"},
				{Ref: "refs/tags/not-semver"},
			},
			repo:        &types.Repo{Repo: "test/repo", Rev: "1.0.0"},
			expectError: true,
		},
		{
			name: "finds latest semantic version",
			tags: []GitHubTag{
				{Ref: "refs/tags/v1.0.0"},
				{Ref: "refs/tags/v2.1.0"},
				{Ref: "refs/tags/v1.5.0"},
			},
			expectedVer: &types.SemanticVersion{Major: 2, Minor: 1, Patch: 0},
			expectError: false,
		},
		{
			name: "mixed valid and invalid tags",
			tags: []GitHubTag{
				{Ref: "refs/tags/invalid"},
				{Ref: "refs/tags/v1.2.3"},
				{Ref: "refs/tags/not-semver"},
				{Ref: "refs/tags/v0.9.0"},
			},
			expectedVer: &types.SemanticVersion{Major: 1, Minor: 2, Patch: 3},
			expectError: false,
		},
		{
			name: "pre-release versions",
			tags: []GitHubTag{
				{Ref: "refs/tags/v1.0.0"},
				{Ref: "refs/tags/v1.1.0-alpha.1"},
				{Ref: "refs/tags/v1.0.5"},
			},
			expectedVer: &types.SemanticVersion{Major: 1, Minor: 1, Patch: 0, PreRelease: "alpha.1"},
			expectError: false,
		},
		{
			name: "tags without refs/tags prefix",
			tags: []GitHubTag{
				{Ref: "v1.0.0"},
				{Ref: "v2.0.0"},
				{Ref: "v1.5.0"},
			},
			expectedVer: &types.SemanticVersion{Major: 2, Minor: 0, Patch: 0},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := findLatestVersion(tt.tags, tt.repo)

			assertFindLatestVersionResult(t, result, err, tt.expectedVer, tt.expectError)
		})
	}
}

func TestFindLatestVersionGitLab(t *testing.T) {
	tests := []struct {
		name        string
		tags        []GitLabTag
		repo        *types.Repo
		expectedVer *types.SemanticVersion
		expectError bool
	}{
		{
			name:        "empty tag list",
			tags:        []GitLabTag{},
			repo:        &types.Repo{Repo: "test/repo", Rev: "1.0.0"},
			expectError: true,
		},
		{
			name: "no valid semantic version tags",
			tags: []GitLabTag{
				{Ref: "invalid-tag"},
				{Ref: "not-semver"},
			},
			repo:        &types.Repo{Repo: "test/repo", Rev: "1.0.0"},
			expectError: true,
		},
		{
			name: "finds latest semantic version",
			tags: []GitLabTag{
				{Ref: "v1.0.0"},
				{Ref: "v2.1.0"},
				{Ref: "v1.5.0"},
			},
			expectedVer: &types.SemanticVersion{Major: 2, Minor: 1, Patch: 0},
			expectError: false,
		},
		{
			name: "mixed valid and invalid tags",
			tags: []GitLabTag{
				{Ref: "invalid"},
				{Ref: "v1.2.3"},
				{Ref: "not-semver"},
				{Ref: "v0.9.0"},
			},
			expectedVer: &types.SemanticVersion{Major: 1, Minor: 2, Patch: 3},
			expectError: false,
		},
		{
			name: "pre-release versions",
			tags: []GitLabTag{
				{Ref: "v1.0.0"},
				{Ref: "v1.1.0-alpha.1"},
				{Ref: "v1.0.5"},
			},
			expectedVer: &types.SemanticVersion{Major: 1, Minor: 1, Patch: 0, PreRelease: "alpha.1"},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := findLatestVersion(tt.tags, tt.repo)

			assertFindLatestVersionResult(t, result, err, tt.expectedVer, tt.expectError)
		})
	}
}

func assertFindLatestVersionResult(t *testing.T, result *types.SemanticVersion, err error, expectedVer *types.SemanticVersion, expectError bool) {
	if expectError {
		assert.Error(t, err, "Expected error but got none")
		assert.Nil(t, result, "Result should be nil when error expected")
	} else {
		assert.NoError(t, err, "Unexpected error: %v", err)
		assert.NotNil(t, result, "Result should not be nil")
		assert.Equal(t, expectedVer.Major, result.Major, "Major version mismatch")
		assert.Equal(t, expectedVer.Minor, result.Minor, "Minor version mismatch")
		assert.Equal(t, expectedVer.Patch, result.Patch, "Patch version mismatch")
		assert.Equal(t, expectedVer.PreRelease, result.PreRelease, "PreRelease mismatch")
	}
}
