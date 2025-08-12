package utils

import "regexp"

// GetGroup extracts a named group from regex match results.
// It returns an empty string if the group name doesn't exist or index is out of bounds.
func GetGroup(re *regexp.Regexp, match []string, name string) string {
	index := re.SubexpIndex(name)
	if index == -1 || index >= len(match) {
		return ""
	}
	return match[index]
}
