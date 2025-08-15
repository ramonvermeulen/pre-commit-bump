package parser

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/ramonvermeulen/pre-commit-bump/core/types"
)

func TestParser_ParseConfig(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		filename    string
		expectError bool
		errorMsg    string
		validate    func(t *testing.T, config *types.PreCommitConfig)
	}{
		{
			name:     "valid config with GitHub repo",
			filename: "valid-config.yaml",
			content: `repos:
  - repo: https://github.com/psf/black
    rev: 22.3.0
    hooks:
      - id: black`,
			expectError: false,
			validate: func(t *testing.T, config *types.PreCommitConfig) {
				assert.Len(t, config.Repos, 1)
				assert.Equal(t, "https://github.com/psf/black", config.Repos[0].Repo)
				assert.Equal(t, "22.3.0", config.Repos[0].Rev)
				assert.NotNil(t, config.Repos[0].SemVer)
			},
		},
		{
			name:     "valid config with GitLab repo",
			filename: "gitlab-config.yaml",
			content: `repos:
  - repo: https://gitlab.com/owner/repo
    rev: v1.2.3
    hooks:
      - id: test`,
			expectError: false,
			validate: func(t *testing.T, config *types.PreCommitConfig) {
				assert.Len(t, config.Repos, 1)
				assert.Equal(t, "https://gitlab.com/owner/repo", config.Repos[0].Repo)
				assert.Equal(t, "v1.2.3", config.Repos[0].Rev)
			},
		},
		{
			name:     "valid config with GitLab repo and random newlines",
			filename: "gitlab-config.yaml",
			content: `
repos:
  - repo: https://gitlab.com/owner/repo


    rev: v1.2.3

    hooks:

      - id: test



`,
			expectError: false,
			validate: func(t *testing.T, config *types.PreCommitConfig) {
				assert.Len(t, config.Repos, 1)
				assert.Equal(t, "https://gitlab.com/owner/repo", config.Repos[0].Repo)
				assert.Equal(t, "v1.2.3", config.Repos[0].Rev)
			},
		},
		{
			name:     "config with local and meta repos",
			filename: "sentinel-config.yaml",
			content: `repos:
  - repo: local
    hooks:
      - id: test-local
  - repo: meta
    hooks:
      - id: test-meta
  - repo: https://github.com/owner/repo
    rev: 1.0.0
    hooks:
      - id: test`,
			expectError: false,
			validate: func(t *testing.T, config *types.PreCommitConfig) {
				assert.Len(t, config.Repos, 3)
				assert.Equal(t, "local", config.Repos[0].Repo)
				assert.Equal(t, "meta", config.Repos[1].Repo)
				assert.Equal(t, "https://github.com/owner/repo", config.Repos[2].Repo)
			},
		},
		{
			name:     "config with invalid semantic version",
			filename: "invalid-semver.yaml",
			content: `repos:
  - repo: https://github.com/owner/repo
    rev: invalid-version
    hooks:
      - id: test`,
			expectError: false,
			validate: func(t *testing.T, config *types.PreCommitConfig) {
				assert.Len(t, config.Repos, 1)
				assert.Nil(t, config.Repos[0].SemVer)
			},
		},
		{
			name:        "empty config file",
			filename:    "empty.yaml",
			content:     "",
			expectError: true,
			errorMsg:    "no repositories found in config",
		},
		{
			name:     "config with empty repo URL",
			filename: "empty-repo.yaml",
			content: `repos:
  - repo: ""
    rev: 1.0.0`,
			expectError: true,
			errorMsg:    "repository URL is empty",
		},
		{
			name:     "config with missing revision",
			filename: "missing-rev.yaml",
			content: `repos:
  - repo: https://github.com/owner/repo
    hooks:
      - id: test`,
			expectError: true,
			errorMsg:    "revision is empty for repository",
		},
		{
			name:        "invalid YAML syntax",
			filename:    "invalid.yaml",
			content:     "repos:\n  - repo: https://github.com/owner/repo\n    rev: 1.0.0\n  invalid yaml",
			expectError: true,
			errorMsg:    "failed to parse",
		},
		{
			name:     "config with multiple repos",
			filename: "multiple-repos.yaml",
			content: `repos:
  - repo: https://github.com/psf/black
    rev: 22.3.0
    hooks:
      - id: black
  - repo: https://gitlab.com/owner/repo
    rev: v2.1.0
    hooks:
      - id: test
  - repo: local
    hooks:
      - id: local-hook`,
			expectError: false,
			validate: func(t *testing.T, config *types.PreCommitConfig) {
				assert.Len(t, config.Repos, 3)
				assert.NotNil(t, config.Repos[0].SemVer)
				assert.NotNil(t, config.Repos[1].SemVer)
				assert.Nil(t, config.Repos[2].SemVer)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, tt.filename)

			err := os.WriteFile(configPath, []byte(tt.content), 0644)
			require.NoError(t, err, "Failed to create test file")

			parser := NewParser(zap.NewNop())
			config, err := parser.ParseConfig(configPath)

			if tt.expectError {
				assert.Error(t, err, "Expected error but got none")
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
				assert.Nil(t, config, "Config should be nil when error expected")
			} else {
				assert.NoError(t, err, "Unexpected error: %v", err)
				assert.NotNil(t, config, "Config should not be nil")
				assert.NotNil(t, config.Logger, "Logger should be set")

				if tt.validate != nil {
					tt.validate(t, config)
				}
			}
		})
	}
}

func TestParser_ParseConfig_FileErrors(t *testing.T) {
	tests := []struct {
		name        string
		setupFile   func(t *testing.T) string
		expectError bool
		errorMsg    string
	}{
		{
			name: "non-existent file",
			setupFile: func(t *testing.T) string {
				return "/non/existent/file.yaml"
			},
			expectError: true,
			errorMsg:    "path does not exist",
		},
		{
			name: "relative path",
			setupFile: func(t *testing.T) string {
				tmpDir := t.TempDir()
				configPath := filepath.Join(tmpDir, "config.yaml")
				validContent := `repos:
  - repo: https://github.com/test/repo
    rev: v1.0.0
    hooks:
      - id: test`
				err := os.WriteFile(configPath, []byte(validContent), 0644)
				require.NoError(t, err)

				oldWd, err := os.Getwd()
				require.NoError(t, err)

				err = os.Chdir(tmpDir)
				require.NoError(t, err)

				t.Cleanup(func() {
					_ = os.Chdir(oldWd)
				})

				return "config.yaml"
			},
			expectError: false,
		},
		{
			name: "directory instead of file",
			setupFile: func(t *testing.T) string {
				tmpDir := t.TempDir()
				dirPath := filepath.Join(tmpDir, "notafile")
				err := os.Mkdir(dirPath, 0755)
				require.NoError(t, err)
				return dirPath
			},
			expectError: true,
			errorMsg:    "failed to read",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser(zap.NewNop())
			filePath := tt.setupFile(t)

			config, err := parser.ParseConfig(filePath)

			if tt.expectError {
				assert.Error(t, err, "Expected error but got none")
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
				assert.Nil(t, config, "Config should be nil when error expected")
			} else {
				assert.NoError(t, err, "Unexpected error: %v", err)
				assert.NotNil(t, config, "Config should not be nil")
			}
		})
	}
}

func TestParser_validatePath(t *testing.T) {
	tests := []struct {
		name        string
		setupPath   func(t *testing.T) string
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid absolute path",
			setupPath: func(t *testing.T) string {
				tmpDir := t.TempDir()
				configPath := filepath.Join(tmpDir, "test.yaml")
				err := os.WriteFile(configPath, []byte("test"), 0644)
				require.NoError(t, err)
				return configPath
			},
			expectError: false,
		},
		{
			name: "valid relative path",
			setupPath: func(t *testing.T) string {
				tmpDir := t.TempDir()
				configPath := filepath.Join(tmpDir, "test.yaml")
				err := os.WriteFile(configPath, []byte("test"), 0644)
				require.NoError(t, err)

				oldWd, err := os.Getwd()
				require.NoError(t, err)

				err = os.Chdir(tmpDir)
				require.NoError(t, err)

				t.Cleanup(func() {
					_ = os.Chdir(oldWd)
				})

				return "test.yaml"
			},
			expectError: false,
		},
		{
			name: "non-existent file",
			setupPath: func(t *testing.T) string {
				return "/path/that/does/not/exist.yaml"
			},
			expectError: true,
			errorMsg:    "path does not exist",
		},
		{
			name: "empty path resulting in directory error",
			setupPath: func(t *testing.T) string {
				return ""
			},
			expectError: true,
			errorMsg:    "is a directory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser(zap.NewNop())
			testPath := tt.setupPath(t)

			_, err := parser.ParseConfig(testPath)

			if tt.expectError {
				assert.Error(t, err, "Expected error but got none")
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				if err != nil {
					assert.NotContains(t, err.Error(), "path does not exist")
				}
			}
		})
	}
}

func TestNewParser(t *testing.T) {
	logger := zap.NewNop()
	parser := NewParser(logger)

	assert.NotNil(t, parser, "Parser should not be nil")
}
