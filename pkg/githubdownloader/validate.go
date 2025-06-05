package githubdownloader

import (
	"errors"
	"regexp"
)

var (
	ErrInvalidGitHubURL = errors.New("invalid GitHub repository URL")
)

// ValidateGitHubURL checks if the provided URL is a valid GitHub repository URL.
func ValidateGitHubURL(url string) error {
	// Basic regex for GitHub repo URLs (https only)
	re := regexp.MustCompile(`^https://github\.com/[^/]+/[^/]+/?$`)
	if !re.MatchString(url) {
		return ErrInvalidGitHubURL
	}
	return nil
}
