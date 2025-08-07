package types

// UpdateResult holds the result of checking a repository for updates.
type UpdateResult struct {
	Repo           Repo
	LatestVersion  *SemanticVersion
	UpdateRequired bool
	Error          error
}
