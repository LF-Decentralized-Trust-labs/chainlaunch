package registry

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"archive/zip"
	"io/ioutil"
	"strings"

	"github.com/chainlaunch/chainlaunch/pkg/githubdownloader"
	"github.com/chainlaunch/chainlaunch/pkg/plugin/types"
	"gopkg.in/yaml.v3"
)

// GitHubSource implements PluginSource for GitHub repositories
// config.URL should be a GitHub repo URL
// Only public repos are supported for now
// Plugins are expected to be YAML files in the root or a plugins/ directory

type GitHubSource struct {
	config     RegistrySource
	downloader *githubdownloader.Downloader
	cacheDir   string
	repoURL    string
}

func NewGitHubSource(config RegistrySource) (*GitHubSource, error) {
	if config.URL == "" {
		return nil, fmt.Errorf("github source requires a URL")
	}
	cacheDir := filepath.Join(os.TempDir(), "plugin-github-cache")
	downloader := githubdownloader.NewDownloader(cacheDir)
	return &GitHubSource{
		config:     config,
		downloader: downloader,
		cacheDir:   cacheDir,
		repoURL:    config.URL,
	}, nil
}

func (s *GitHubSource) List() ([]PluginMetadata, error) {
	zipPath, _, err := s.downloader.DownloadRepo(s.repoURL)
	if err != nil {
		return nil, fmt.Errorf("failed to download repo: %w", err)
	}
	zipReader, err := zip.OpenReader(zipPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open zip: %w", err)
	}
	defer zipReader.Close()

	var plugins []PluginMetadata
	for _, f := range zipReader.File {
		if f.FileInfo().IsDir() {
			continue
		}
		filename := filepath.Base(f.Name)
		if !(filename == "plugin.yaml" || filename == "plugin.yml") {
			continue
		}
		file, err := f.Open()
		if err != nil {
			continue
		}
		data, err := ioutil.ReadAll(file)
		file.Close()
		if err != nil {
			continue
		}
		var p types.Plugin
		if err := yaml.Unmarshal(data, &p); err != nil {
			continue
		}
		plugins = append(plugins, PluginMetadata{
			Name:        p.Metadata.Name,
			Version:     p.Metadata.Version,
			Description: p.Metadata.Description,
			Tags:        p.Metadata.Tags,
			Author:      p.Metadata.Author,
			License:     p.Metadata.License,
			Source:      s.repoURL,
			Hash:        "",
			Rating:      0,
			Downloads:   0,
			Created:     time.Time{},
			Updated:     time.Time{},
			Labels:      map[string]string{},
			RawYAML:     string(data),
		})
		break
	}
	return plugins, nil
}

func (s *GitHubSource) Get(name string) (*types.Plugin, error) {
	zipPath, _, err := s.downloader.DownloadRepo(s.repoURL)
	if err != nil {
		return nil, fmt.Errorf("failed to download repo: %w", err)
	}
	zipReader, err := zip.OpenReader(zipPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open zip: %w", err)
	}
	defer zipReader.Close()

	candidates := []string{
		name + ".yaml",
		name + ".yml",
		"plugins/" + name + ".yaml",
		"plugins/" + name + ".yml",
	}
	for _, f := range zipReader.File {
		if f.FileInfo().IsDir() {
			continue
		}
		for _, candidate := range candidates {
			if strings.HasSuffix(f.Name, candidate) {
				file, err := f.Open()
				if err != nil {
					continue
				}
				data, err := ioutil.ReadAll(file)
				file.Close()
				if err != nil {
					continue
				}
				var p types.Plugin
				if err := yaml.Unmarshal(data, &p); err != nil {
					continue
				}
				if p.Metadata.Name == name {
					return &p, nil
				}
			}
		}
	}
	return nil, fmt.Errorf("plugin %s not found in GitHub repo", name)
}

func (s *GitHubSource) Search(query string) ([]PluginMetadata, error) {
	return nil, fmt.Errorf("not implemented yet")
}

func (s *GitHubSource) Verify(p *types.Plugin) error {
	// Basic validation
	if p.APIVersion == "" {
		return fmt.Errorf("apiVersion is required")
	}
	if p.Kind == "" {
		return fmt.Errorf("kind is required")
	}
	if p.Metadata.Name == "" {
		return fmt.Errorf("metadata.name is required")
	}
	return nil
}
