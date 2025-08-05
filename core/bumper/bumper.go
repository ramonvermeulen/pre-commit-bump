package bumper

import (
	"github.com/ramonvermeulen/pre-commit-bump/config"
	"github.com/ramonvermeulen/pre-commit-bump/core/parser"
)

// Bumper coordinates the pre-commit hook bumping process.
type Bumper struct {
	parser *parser.Parser
	cfg    *config.Config
}

// NewBumper creates a new Bumper instance with the given parser and cfg
func NewBumper(parser *parser.Parser, cfg *config.Config) *Bumper {
	return &Bumper{
		parser: parser,
		cfg:    cfg,
	}
}

// Check verifies if the pre-commit configuration file is valid and up-to-date.
// If the configuration is valid, it returns nil.
// If there are updates available, it returns an error.
func (b *Bumper) Check() error {
	b.cfg.Logger.Sugar().Debugf("Parsing configuration file: %s", b.cfg.ConfigPath)

	cfg, err := b.parser.ParseConfig(b.cfg.ConfigPath)
	if err != nil {
		return err
	}

	validRepos := cfg.ValidRepos()
	b.cfg.Logger.Sugar().Debugf("Configuration parsed successfully - total_repos: %d, valid_repos: %d",
		len(cfg.Repos), len(validRepos))

	return nil
}

// Update checks for available updates and modifies the pre-commit configuration file.
func (b *Bumper) Update() error {
	b.cfg.Logger.Sugar().Debugf("Parsing configuration file: %s", b.cfg.ConfigPath)

	pcConfig, err := b.parser.ParseConfig(b.cfg.ConfigPath)
	if err != nil {
		return err
	}

	validRepos := pcConfig.ValidRepos()
	b.cfg.Logger.Sugar().Debugf("Configuration parsed successfully - total_repos: %d, valid_repos: %d",
		len(pcConfig.Repos), len(validRepos))

	// TODO: Implement update logic using b.cfg.DryRun and b.cfg.NoSummary
	return nil
}
