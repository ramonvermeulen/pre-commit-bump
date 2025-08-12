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
