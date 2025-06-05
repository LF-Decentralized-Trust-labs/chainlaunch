package githubdownloader

import "testing"

func TestValidateGitHubURL(t *testing.T) {
	valid := []string{
		"https://github.com/user/repo",
		"https://github.com/user/repo/",
	}
	invalid := []string{
		"http://github.com/user/repo",
		"https://github.com/user",
		"https://github.com/user/repo/extra",
		"https://gitlab.com/user/repo",
		"not a url",
		"",
	}
	for _, url := range valid {
		if err := ValidateGitHubURL(url); err != nil {
			t.Errorf("expected valid, got error for %q: %v", url, err)
		}
	}
	for _, url := range invalid {
		if err := ValidateGitHubURL(url); err == nil {
			t.Errorf("expected error, got valid for %q", url)
		}
	}
}
