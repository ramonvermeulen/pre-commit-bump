package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/pflag"

	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Config holds all configuration values for the pre-commit bumper tool
type Config struct {
	// ConfigPath is the path to the pre-commit configuration file
	ConfigPath string

	// NoSummary disables summary generation (update command only)
	NoSummary bool

	// DryRun performs a dry run without modifying files (update command only)
	DryRun bool

	// LogLevel determines the logging verbosity
	LogLevel zapcore.Level

	// Logger is the configured logger instance
	Logger *zap.Logger
}

// getLogLevel determines the log level from verbose flag and environment variable
func getLogLevel() zapcore.Level {
	levelMap := map[string]zapcore.Level{
		"DEBUG":   zapcore.DebugLevel,
		"INFO":    zapcore.InfoLevel,
		"WARN":    zapcore.WarnLevel,
		"WARNING": zapcore.WarnLevel,
		"ERROR":   zapcore.ErrorLevel,
	}

	if envLevel := os.Getenv("PCB_LOG"); envLevel != "" {
		if lvl, ok := levelMap[strings.ToUpper(envLevel)]; ok {
			return lvl
		}
	}

	if viper.GetBool(FlagVerbose) {
		return zapcore.DebugLevel
	}

	return zapcore.InfoLevel
}

// newLogger creates a basic zap logger
func newLogger(level zapcore.Level) *zap.Logger {
	config := zap.NewDevelopmentConfig()
	config.Level = zap.NewAtomicLevelAt(level)
	config.DisableCaller = true
	logger, _ := config.Build()
	return logger
}

// FromViper creates a Config from viper values
func FromViper() (*Config, error) {
	configPath := viper.GetString(FlagConfig)
	noSummary := viper.GetBool(FlagNoSummary)
	dryRun := viper.GetBool(FlagDryRun)
	logLevel := getLogLevel()

	return &Config{
		ConfigPath: configPath,
		NoSummary:  noSummary,
		DryRun:     dryRun,
		LogLevel:   logLevel,
		Logger:     newLogger(logLevel),
	}, nil
}

// BindFlag binds a flag from a FlagSet to viper and handles errors during binding
func BindFlag(flagSet *pflag.FlagSet, flagName string) {
	if err := viper.BindPFlag(flagName, flagSet.Lookup(flagName)); err != nil {
		fmt.Fprintf(os.Stderr, "Error binding flag %s: %v\n", flagName, err)
		os.Exit(1)
	}
}
