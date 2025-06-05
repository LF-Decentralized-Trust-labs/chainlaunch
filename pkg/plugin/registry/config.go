package registry

import (
	"fmt"
	"sync"
	"time"

	"github.com/chainlaunch/chainlaunch/pkg/plugin/types"
)

// RegistrySource represents a source for plugins
type RegistrySource struct {
	Name        string            `json:"name" yaml:"name"`
	Type        string            `json:"type" yaml:"type"` // local, git, ipfs, marketplace
	URL         string            `json:"url" yaml:"url"`
	Enabled     bool              `json:"enabled" yaml:"enabled"`
	Trust       bool              `json:"trust" yaml:"trust"` // Whether to trust plugins from this source
	Credentials map[string]string `json:"credentials,omitempty" yaml:"credentials,omitempty"`
}

// RegistryConfig represents the configuration for the plugin registry
type RegistryConfig struct {
	Sources     []RegistrySource `json:"sources" yaml:"sources"`
	CacheTTL    time.Duration    `json:"cacheTTL" yaml:"cacheTTL"`
	AutoUpdate  bool             `json:"autoUpdate" yaml:"autoUpdate"`
	VerifyHash  bool             `json:"verifyHash" yaml:"verifyHash"`
	AllowedTags []string         `json:"allowedTags" yaml:"allowedTags"`
}

// Registry manages plugin sources and discovery
type Registry struct {
	config  *RegistryConfig
	sources map[string]PluginSource
}

// PluginSource defines the interface for plugin sources
type PluginSource interface {
	List() ([]PluginMetadata, error)
	Get(name string) (*types.Plugin, error)
	Search(query string) ([]PluginMetadata, error)
	Verify(plugin *types.Plugin) error
}

// PluginMetadata contains searchable metadata about a plugin
type PluginMetadata struct {
	Name        string            `json:"name" yaml:"name"`
	Version     string            `json:"version" yaml:"version"`
	Description string            `json:"description" yaml:"description"`
	Tags        []string          `json:"tags" yaml:"tags"`
	Author      string            `json:"author" yaml:"author"`
	License     string            `json:"license" yaml:"license"`
	Source      string            `json:"source" yaml:"source"`
	Hash        string            `json:"hash" yaml:"hash"`
	Rating      float64           `json:"rating" yaml:"rating"`
	Downloads   int               `json:"downloads" yaml:"downloads"`
	Created     time.Time         `json:"created" yaml:"created"`
	Updated     time.Time         `json:"updated" yaml:"updated"`
	Labels      map[string]string `json:"labels" yaml:"labels"`
	RawYAML     string            `json:"raw_yaml" yaml:"raw_yaml"`
}

// AvailablePluginsCache caches available plugins from GitHub sources
// Thread-safe for concurrent access
type AvailablePluginsCache struct {
	mu          sync.RWMutex
	plugins     []PluginMetadata
	lastUpdated time.Time
}

func NewAvailablePluginsCache() *AvailablePluginsCache {
	return &AvailablePluginsCache{}
}

func (c *AvailablePluginsCache) Get() ([]PluginMetadata, time.Time) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.plugins, c.lastUpdated
}

func (c *AvailablePluginsCache) Set(plugins []PluginMetadata) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.plugins = plugins
	c.lastUpdated = time.Now()
}

// NewRegistry creates a new plugin registry
func NewRegistry(config *RegistryConfig) (*Registry, error) {
	registry := &Registry{
		config:  config,
		sources: make(map[string]PluginSource),
	}

	// Initialize sources
	for _, source := range config.Sources {
		if !source.Enabled {
			continue
		}

		src, err := newSource(source)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize source %s: %w", source.Name, err)
		}
		registry.sources[source.Name] = src
	}

	return registry, nil
}

// newSource creates a new plugin source based on type
func newSource(config RegistrySource) (PluginSource, error) {
	switch config.Type {
	case "local":
		return NewLocalSource(config)
	case "git":
		return NewGitSource(config)
	case "ipfs":
		return NewIPFSSource(config)
	case "marketplace":
		return NewMarketplaceSource(config)
	case "github":
		return NewGitHubSource(config)
	default:
		return nil, fmt.Errorf("unsupported source type: %s", config.Type)
	}
}

// Sources returns all plugin sources in the registry
func (r *Registry) Sources() []PluginSource {
	s := make([]PluginSource, 0, len(r.sources))
	for _, src := range r.sources {
		s = append(s, src)
	}
	return s
}

// ListAvailablePluginsFromGitHub returns all available plugins from GitHub sources
func (r *Registry) ListAvailablePluginsFromGitHub() ([]PluginMetadata, error) {
	result := []PluginMetadata{}
	for _, src := range r.sources {
		// Type assertion to *GitHubSource
		if githubSrc, ok := src.(*GitHubSource); ok {
			plugins, err := githubSrc.List()
			if err != nil {
				continue // skip sources that fail
			}
			result = append(result, plugins...)
		}
	}
	return result, nil
}
