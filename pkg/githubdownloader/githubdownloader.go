// Package githubdownloader provides functionality to download and cache GitHub repositories for plugin import and integration.
package githubdownloader

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	retryCount = 3
	retryDelay = 500 * time.Millisecond
)

// Downloader is the main struct for handling GitHub repository downloads.
type Downloader struct {
	CacheDir string
}

// NewDownloader creates a new Downloader instance with the given cache directory.
func NewDownloader(cacheDir string) *Downloader {
	return &Downloader{CacheDir: cacheDir}
}

// DownloadRepo downloads a GitHub repository as a zip archive and returns the path to the archive and metadata.
// It retries HTTP downloads up to retryCount times on transient errors.
func (d *Downloader) DownloadRepo(url string) (string, *RepoMetadata, error) {
	if err := ValidateGitHubURL(url); err != nil {
		return "", nil, err
	}
	parts := strings.Split(strings.TrimPrefix(url, "https://github.com/"), "/")
	if len(parts) < 2 {
		return "", nil, ErrInvalidGitHubURL
	}
	owner, repo := parts[0], parts[1]
	repo = strings.TrimSuffix(repo, "/")

	branches := []string{"main", "master"}
	for _, branch := range branches {
		// Deterministic cache filename
		cacheFile := filepath.Join(d.CacheDir, owner+"-"+repo+"-"+branch+".zip")
		if _, err := os.Stat(cacheFile); err == nil {
			meta := &RepoMetadata{
				SourceURL:    url,
				CommitHash:   "", // Not available from zip download
				DownloadedAt: 0,  // Not tracked for cached files
			}
			return cacheFile, meta, nil
		}
		// Not cached, try to download
		zipURL := fmt.Sprintf("https://github.com/%s/%s/archive/refs/heads/%s.zip", owner, repo, branch)
		var resp *http.Response
		var err error
		for attempt := 1; attempt <= retryCount; attempt++ {
			resp, err = http.Get(zipURL)
			if err == nil && resp.StatusCode == http.StatusOK {
				break
			}
			if resp != nil {
				resp.Body.Close()
			}
			if attempt < retryCount {
				time.Sleep(retryDelay)
			}
		}
		if err == nil && resp != nil && resp.StatusCode == http.StatusOK {
			defer resp.Body.Close()
			// Ensure cache dir exists
			if err := os.MkdirAll(d.CacheDir, 0o755); err != nil {
				return "", nil, fmt.Errorf("failed to create cache dir: %w", err)
			}
			out, err := os.Create(cacheFile)
			if err != nil {
				return "", nil, fmt.Errorf("failed to create cache file: %w", err)
			}
			defer out.Close()
			if _, err := io.Copy(out, resp.Body); err != nil {
				return "", nil, fmt.Errorf("failed to save zip file: %w", err)
			}
			meta := &RepoMetadata{
				SourceURL:    url,
				CommitHash:   "", // Not available from zip download
				DownloadedAt: time.Now().Unix(),
			}
			return cacheFile, meta, nil
		}
	}
	return "", nil, fmt.Errorf("failed to download repo: not found on main or master branch after %d retries", retryCount)
}
