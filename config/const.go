package config

import "time"

// Flags for the pre-commit bumper tool
const (
	FlagConfig    = "config"
	FlagVerbose   = "verbose"
	FlagAllow     = "allow"
	FlagNoSummary = "no-summary"
	FlagDryRun    = "dry-run"
)

// Sentinel values for hooks
const (
	SentinelLocal = "local"
	SentinelMeta  = "meta"
)

// Supported vendors for pre-commit hooks
const (
	VendorGitHub = "github"
	VendorGitLab = "gitlab"
)

// Regex patterns and other constants used within the pre-commit bumper tool
const (
	// ReSemanticVersion is a regex pattern for validating semantic versioning
	// Regex is used from https://semver.org/, added support for leading or trailing characters like 'v' or 'V'
	ReSemanticVersion  = `(?P<major>0|[1-9]\d*)\.(?P<minor>0|[1-9]\d*)\.(?P<patch>0|[1-9]\d*)(?:-(?P<prerelease>(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+(?P<buildmetadata>[0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?`
	DefaultHTTPTimeout = 30 * time.Second
)
