package registry

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/chainlaunch/chainlaunch/pkg/plugin/types"
)

// MarketplaceSource implements PluginSource for a centralized marketplace
type MarketplaceSource struct {
	config RegistrySource
	client *http.Client
}

// MarketplacePlugin represents a plugin in the marketplace
type MarketplacePlugin struct {
	Name        string            `json:"name"`
	Version     string            `json:"version"`
	Description string            `json:"description"`
	Tags        []string          `json:"tags"`
	Author      string            `json:"author"`
	License     string            `json:"license"`
	Downloads   int               `json:"downloads"`
	Rating      float64           `json:"rating"`
	Created     time.Time         `json:"created"`
	Updated     time.Time         `json:"updated"`
	Labels      map[string]string `json:"labels"`
	Plugin      types.Plugin      `json:"plugin"`
}

// NewMarketplaceSource creates a new marketplace source
func NewMarketplaceSource(config RegistrySource) (*MarketplaceSource, error) {
	if config.URL == "" {
		return nil, fmt.Errorf("marketplace source requires a URL")
	}

	// Validate URL
	_, err := url.Parse(config.URL)
	if err != nil {
		return nil, fmt.Errorf("invalid marketplace URL: %w", err)
	}

	return &MarketplaceSource{
		config: config,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// List returns all plugins from the marketplace
func (s *MarketplaceSource) List() ([]PluginMetadata, error) {
	resp, err := s.client.Get(fmt.Sprintf("%s/plugins", s.config.URL))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch plugins: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("marketplace returned status %d", resp.StatusCode)
	}

	var marketplacePlugins []MarketplacePlugin
	if err := json.NewDecoder(resp.Body).Decode(&marketplacePlugins); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	plugins := make([]PluginMetadata, len(marketplacePlugins))
	for i, p := range marketplacePlugins {
		plugins[i] = PluginMetadata{
			Name:        p.Name,
			Version:     p.Version,
			Description: p.Description,
			Tags:        p.Tags,
			Author:      p.Author,
			License:     p.License,
			Downloads:   p.Downloads,
			Rating:      p.Rating,
			Created:     p.Created,
			Updated:     p.Updated,
			Labels:      p.Labels,
			Source:      s.config.Name,
		}
	}

	return plugins, nil
}

// Get returns a specific plugin by name
func (s *MarketplaceSource) Get(name string) (*types.Plugin, error) {
	resp, err := s.client.Get(fmt.Sprintf("%s/plugins/%s", s.config.URL, url.PathEscape(name)))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch plugin: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("plugin %s not found", name)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("marketplace returned status %d: %s", resp.StatusCode, body)
	}

	var marketplacePlugin MarketplacePlugin
	if err := json.NewDecoder(resp.Body).Decode(&marketplacePlugin); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &marketplacePlugin.Plugin, nil
}

// Search finds plugins matching the query
func (s *MarketplaceSource) Search(query string) ([]PluginMetadata, error) {
	resp, err := s.client.Get(fmt.Sprintf("%s/plugins/search?q=%s", s.config.URL, url.QueryEscape(query)))
	if err != nil {
		return nil, fmt.Errorf("failed to search plugins: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("marketplace returned status %d", resp.StatusCode)
	}

	var marketplacePlugins []MarketplacePlugin
	if err := json.NewDecoder(resp.Body).Decode(&marketplacePlugins); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	plugins := make([]PluginMetadata, len(marketplacePlugins))
	for i, p := range marketplacePlugins {
		plugins[i] = PluginMetadata{
			Name:        p.Name,
			Version:     p.Version,
			Description: p.Description,
			Tags:        p.Tags,
			Author:      p.Author,
			License:     p.License,
			Downloads:   p.Downloads,
			Rating:      p.Rating,
			Created:     p.Created,
			Updated:     p.Updated,
			Labels:      p.Labels,
			Source:      s.config.Name,
		}
	}

	return plugins, nil
}

// Verify checks if a plugin is valid and verifies its signature if required
func (s *MarketplaceSource) Verify(p *types.Plugin) error {
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

	// Verify plugin signature with marketplace
	if s.config.Trust {
		resp, err := s.client.Get(fmt.Sprintf("%s/plugins/%s/verify", s.config.URL, url.PathEscape(p.Metadata.Name)))
		if err != nil {
			return fmt.Errorf("failed to verify plugin: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("marketplace verification failed: %s", body)
		}
	}

	return nil
}
