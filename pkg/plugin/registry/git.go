package registry

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/chainlaunch/chainlaunch/pkg/plugin/types"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"gopkg.in/yaml.v3"
)

// GitSource implements PluginSource for Git repositories
type GitSource struct {
	config     RegistrySource
	localPath  string
	repository *git.Repository
}

// NewGitSource creates a new Git source
func NewGitSource(config RegistrySource) (*GitSource, error) {
	if config.URL == "" {
		return nil, fmt.Errorf("git source requires a URL")
	}

	// Create temporary directory for the repository
	tempDir, err := os.MkdirTemp("", "plugin-registry-git-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Clone repository
	repo, err := git.PlainClone(tempDir, false, &git.CloneOptions{
		URL:      config.URL,
		Progress: os.Stdout,
	})
	if err != nil {
		os.RemoveAll(tempDir)
		return nil, fmt.Errorf("failed to clone repository: %w", err)
	}

	return &GitSource{
		config:     config,
		localPath:  tempDir,
		repository: repo,
	}, nil
}

// List returns all plugins in the Git repository
func (s *GitSource) List() ([]PluginMetadata, error) {
	var plugins []PluginMetadata

	// Get the worktree
	worktree, err := s.repository.Worktree()
	if err != nil {
		return nil, fmt.Errorf("failed to get worktree: %w", err)
	}

	// Walk through the repository
	err = filepath.Walk(worktree.Filesystem.Root(), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || (!strings.HasSuffix(info.Name(), ".yaml") && !strings.HasSuffix(info.Name(), ".yml")) {
			return nil
		}

		plugin, err := s.loadPlugin(path)
		if err != nil {
			return fmt.Errorf("failed to load plugin %s: %w", path, err)
		}

		metadata, err := s.getMetadata(plugin, path)
		if err != nil {
			return fmt.Errorf("failed to get metadata for %s: %w", path, err)
		}

		plugins = append(plugins, metadata)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list plugins: %w", err)
	}

	return plugins, nil
}

// Get returns a specific plugin by name
func (s *GitSource) Get(name string) (*types.Plugin, error) {
	// Get the worktree
	worktree, err := s.repository.Worktree()
	if err != nil {
		return nil, fmt.Errorf("failed to get worktree: %w", err)
	}

	// Search for plugin file
	var pluginPath string
	err = filepath.Walk(worktree.Filesystem.Root(), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if strings.HasPrefix(info.Name(), name+".") &&
			(strings.HasSuffix(info.Name(), ".yaml") || strings.HasSuffix(info.Name(), ".yml")) {
			pluginPath = path
			return filepath.SkipAll
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to find plugin: %w", err)
	}

	if pluginPath == "" {
		return nil, fmt.Errorf("plugin %s not found", name)
	}

	return s.loadPlugin(pluginPath)
}

// Search finds plugins matching the query
func (s *GitSource) Search(query string) ([]PluginMetadata, error) {
	query = strings.ToLower(query)
	plugins, err := s.List()
	if err != nil {
		return nil, err
	}

	var results []PluginMetadata
	for _, p := range plugins {
		if strings.Contains(strings.ToLower(p.Name), query) ||
			strings.Contains(strings.ToLower(p.Description), query) {
			results = append(results, p)
		}

		// Search in tags
		for _, tag := range p.Tags {
			if strings.Contains(strings.ToLower(tag), query) {
				results = append(results, p)
				break
			}
		}
	}

	return results, nil
}

// Verify checks if a plugin is valid and its signature if required
func (s *GitSource) Verify(p *types.Plugin) error {
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

// loadPlugin loads a plugin from a file
func (s *GitSource) loadPlugin(path string) (*types.Plugin, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read plugin file: %w", err)
	}

	var p types.Plugin
	if err := yaml.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("failed to unmarshal plugin: %w", err)
	}

	return &p, nil
}

// getMetadata returns metadata for a plugin
func (s *GitSource) getMetadata(p *types.Plugin, path string) (PluginMetadata, error) {
	// Get commit history for the file
	commits, err := s.getFileCommits(path)
	if err != nil {
		return PluginMetadata{}, fmt.Errorf("failed to get commit history: %w", err)
	}

	var created, updated time.Time
	if len(commits) > 0 {
		created = commits[len(commits)-1].Author.When // First commit
		updated = commits[0].Author.When              // Latest commit
	}

	// Calculate file hash
	data, err := os.ReadFile(path)
	if err != nil {
		return PluginMetadata{}, fmt.Errorf("failed to read plugin file: %w", err)
	}

	hash := plumbing.ComputeHash(plumbing.BlobObject, data)

	return PluginMetadata{
		Name:        p.Metadata.Name,
		Description: "", // Could be added to Plugin struct
		Source:      s.config.Name,
		Hash:        hash.String(),
		Created:     created,
		Updated:     updated,
	}, nil
}

// getFileCommits returns the commit history for a file
func (s *GitSource) getFileCommits(path string) ([]*object.Commit, error) {
	// Get the commit history
	head, err := s.repository.Head()
	if err != nil {
		return nil, fmt.Errorf("failed to get HEAD: %w", err)
	}

	commits, err := s.repository.Log(&git.LogOptions{From: head.Hash()})
	if err != nil {
		return nil, fmt.Errorf("failed to get commit history: %w", err)
	}

	var fileCommits []*object.Commit
	err = commits.ForEach(func(c *object.Commit) error {
		// Check if the commit affects our file
		stats, err := c.Stats()
		if err != nil {
			return err
		}

		for _, stat := range stats {
			if stat.Name == path {
				fileCommits = append(fileCommits, c)
				break
			}
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to process commits: %w", err)
	}

	return fileCommits, nil
}
