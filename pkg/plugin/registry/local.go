package registry

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/chainlaunch/chainlaunch/pkg/plugin/types"
	"gopkg.in/yaml.v3"
)

// LocalSource implements PluginSource for local filesystem
type LocalSource struct {
	config RegistrySource
	path   string
}

// NewLocalSource creates a new local source
func NewLocalSource(config RegistrySource) (*LocalSource, error) {
	if config.URL == "" {
		return nil, fmt.Errorf("local source requires a path")
	}

	// Expand path if it contains ~
	if strings.HasPrefix(config.URL, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		config.URL = filepath.Join(home, config.URL[1:])
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(config.URL, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	return &LocalSource{
		config: config,
		path:   config.URL,
	}, nil
}

// List returns all plugins in the local directory
func (s *LocalSource) List() ([]PluginMetadata, error) {
	var plugins []PluginMetadata

	err := filepath.Walk(s.path, func(path string, info os.FileInfo, err error) error {
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
func (s *LocalSource) Get(name string) (*types.Plugin, error) {
	// Search for plugin file
	var pluginPath string
	err := filepath.Walk(s.path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if strings.HasPrefix(info.Name(), name+".") &&
			(strings.HasSuffix(info.Name(), ".yaml") || strings.HasSuffix(info.Name(), ".yml")) {
			pluginPath = path
			return io.EOF // Use EOF to stop walking
		}

		return nil
	})

	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("failed to find plugin: %w", err)
	}

	if pluginPath == "" {
		return nil, fmt.Errorf("plugin %s not found", name)
	}

	return s.loadPlugin(pluginPath)
}

// Search finds plugins matching the query
func (s *LocalSource) Search(query string) ([]PluginMetadata, error) {
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

// Verify checks if a plugin is valid
func (s *LocalSource) Verify(p *types.Plugin) error {
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
func (s *LocalSource) loadPlugin(path string) (*types.Plugin, error) {
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
func (s *LocalSource) getMetadata(p *types.Plugin, path string) (PluginMetadata, error) {
	// Calculate hash
	data, err := os.ReadFile(path)
	if err != nil {
		return PluginMetadata{}, fmt.Errorf("failed to read plugin file: %w", err)
	}

	hash := sha256.Sum256(data)

	info, err := os.Stat(path)
	if err != nil {
		return PluginMetadata{}, fmt.Errorf("failed to get file info: %w", err)
	}

	return PluginMetadata{
		Name:        p.Metadata.Name,
		Description: "", // Could be added to Plugin struct
		Source:      s.config.Name,
		Hash:        hex.EncodeToString(hash[:]),
		Created:     info.ModTime(),
		Updated:     info.ModTime(),
	}, nil
}
