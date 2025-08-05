package parser

import (
	"fmt"
	"os"
	"path/filepath"

	"go.uber.org/zap"

	"github.com/goccy/go-yaml"
)

// Parser is responsible for parsing the pre-commit configuration file.
// It provides methods to read and validate the configuration file.
type Parser struct {
	logger *zap.Logger
}

// NewParser creates a new instance of Parser.
// It initializes the parser and returns a pointer to it.
func NewParser(logger *zap.Logger) *Parser {
	return &Parser{logger: logger}
}

// ParseConfig reads and parses the pre-commit configuration file from the given path.
// It returns a PreCommitConfig struct or an error if the parsing fails.
func (p *Parser) ParseConfig(configPath string) (*PreCommitConfig, error) {
	absPath, err := p.validatePath(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to validate config path: %w", err)
	}

	data, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config PreCommitConfig
	config.logger = p.logger
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse yaml: %w", err)
	}

	err = config.Validate()
	if err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	config.PopulateSemVer()

	return &config, nil
}

// validatePath checks if the provided configPath is valid and exists.
// It returns the absolute path if valid, or an error if not.
func (p *Parser) validatePath(configPath string) (string, error) {
	absPath, err := filepath.Abs(configPath)
	if err != nil {
		return "", fmt.Errorf("invalid path: %w", err)
	}

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return "", fmt.Errorf("path does not exist: %s", absPath)
	}

	return absPath, nil
}
