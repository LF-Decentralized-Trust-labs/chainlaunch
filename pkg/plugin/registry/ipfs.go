package registry

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/chainlaunch/chainlaunch/pkg/plugin/types"
	shell "github.com/ipfs/go-ipfs-api"
	"gopkg.in/yaml.v3"
)

// IPFSSource implements PluginSource for IPFS
type IPFSSource struct {
	config RegistrySource
	sh     *shell.Shell
}

// NewIPFSSource creates a new IPFS source
func NewIPFSSource(config RegistrySource) (*IPFSSource, error) {
	if config.URL == "" {
		return nil, fmt.Errorf("IPFS source requires a gateway URL")
	}

	return &IPFSSource{
		config: config,
		sh:     shell.NewShell(config.URL),
	}, nil
}

// List returns all plugins from IPFS
func (s *IPFSSource) List() ([]PluginMetadata, error) {
	// Get the root directory CID from the config
	rootCID := s.config.Credentials["rootCID"]
	if rootCID == "" {
		return nil, fmt.Errorf("rootCID is required in credentials")
	}

	// List files in the directory
	files, err := s.sh.List(rootCID)
	if err != nil {
		return nil, fmt.Errorf("failed to list IPFS directory: %w", err)
	}

	var plugins []PluginMetadata
	for _, f := range files {
		if !strings.HasSuffix(f.Name, ".yaml") && !strings.HasSuffix(f.Name, ".yml") {
			continue
		}

		plugin, err := s.loadPlugin(f.Hash)
		if err != nil {
			return nil, fmt.Errorf("failed to load plugin %s: %w", f.Name, err)
		}

		metadata, err := s.getMetadata(plugin, f)
		if err != nil {
			return nil, fmt.Errorf("failed to get metadata for %s: %w", f.Name, err)
		}

		plugins = append(plugins, metadata)
	}

	return plugins, nil
}

// Get returns a specific plugin by name
func (s *IPFSSource) Get(name string) (*types.Plugin, error) {
	// Get the root directory CID from the config
	rootCID := s.config.Credentials["rootCID"]
	if rootCID == "" {
		return nil, fmt.Errorf("rootCID is required in credentials")
	}

	// List files in the directory to find the plugin
	files, err := s.sh.List(rootCID)
	if err != nil {
		return nil, fmt.Errorf("failed to list IPFS directory: %w", err)
	}

	for _, f := range files {
		if strings.HasPrefix(f.Name, name+".") &&
			(strings.HasSuffix(f.Name, ".yaml") || strings.HasSuffix(f.Name, ".yml")) {
			return s.loadPlugin(f.Hash)
		}
	}

	return nil, fmt.Errorf("plugin %s not found", name)
}

// Search finds plugins matching the query
func (s *IPFSSource) Search(query string) ([]PluginMetadata, error) {
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
func (s *IPFSSource) Verify(p *types.Plugin) error {
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

// loadPlugin loads a plugin from IPFS
func (s *IPFSSource) loadPlugin(cid string) (*types.Plugin, error) {
	reader, err := s.sh.Cat(cid)
	if err != nil {
		return nil, fmt.Errorf("failed to read from IPFS: %w", err)
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read data: %w", err)
	}

	var p types.Plugin
	if err := yaml.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("failed to unmarshal plugin: %w", err)
	}

	return &p, nil
}

// getMetadata returns metadata for a plugin
func (s *IPFSSource) getMetadata(p *types.Plugin, file *shell.LsLink) (PluginMetadata, error) {
	// Get object stats from IPFS
	stats, err := s.sh.ObjectStat(file.Hash)
	if err != nil {
		return PluginMetadata{}, fmt.Errorf("failed to get IPFS stats: %w", err)
	}

	return PluginMetadata{
		Name:        p.Metadata.Name,
		Description: "", // Could be added to Plugin struct
		Source:      s.config.Name,
		Hash:        file.Hash,
		Created:     time.Unix(int64(stats.CumulativeSize), 0), // Using CumulativeSize as a proxy for creation time
		Updated:     time.Unix(int64(stats.CumulativeSize), 0),
	}, nil
}
