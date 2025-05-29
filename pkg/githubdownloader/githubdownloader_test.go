package githubdownloader

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestNewDownloader(t *testing.T) {
	d := NewDownloader(os.TempDir())
	if d == nil {
		t.Fatal("expected non-nil Downloader instance")
	}
}

func TestDownloadRepo(t *testing.T) {
	d := NewDownloader(os.TempDir())
	// Use a small public repo for testing
	url := "https://github.com/octocat/Hello-World"
	zipPath, meta, err := d.DownloadRepo(url)
	if err != nil {
		t.Fatalf("DownloadRepo failed: %v", err)
	}
	if zipPath == "" {
		t.Error("expected non-empty zipPath")
	}
	if meta == nil || meta.SourceURL != url {
		t.Error("unexpected metadata")
	}
	// Clean up
	if zipPath != "" {
		_ = os.Remove(zipPath)
	}
}

func TestDownloadRepo_Caching(t *testing.T) {
	cacheDir, err := ioutil.TempDir("", "ghcache")
	if err != nil {
		t.Fatalf("failed to create temp cache dir: %v", err)
	}
	defer os.RemoveAll(cacheDir)
	d := NewDownloader(cacheDir)
	url := "https://github.com/octocat/Hello-World"
	// First call: should download
	zipPath1, _, err := d.DownloadRepo(url)
	if err != nil {
		t.Fatalf("DownloadRepo failed: %v", err)
	}
	if zipPath1 == "" {
		t.Error("expected non-empty zipPath1")
	}
	// Second call: should use cache
	zipPath2, _, err := d.DownloadRepo(url)
	if err != nil {
		t.Fatalf("DownloadRepo (cached) failed: %v", err)
	}
	if zipPath2 != zipPath1 {
		t.Errorf("expected cached path %q, got %q", zipPath1, zipPath2)
	}
}
