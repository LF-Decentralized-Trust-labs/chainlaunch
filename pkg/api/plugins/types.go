package plugins

import (
	"encoding/json"
	"time"
)

// CreatePluginRequest represents the HTTP request body for creating a plugin
type CreatePluginRequest struct {
	Name       string          `json:"name"`
	APIVersion string          `json:"apiVersion"`
	Kind       string          `json:"kind"`
	Metadata   json.RawMessage `json:"metadata"`
	Spec       json.RawMessage `json:"spec"`
}

// UpdatePluginRequest represents the HTTP request body for updating a plugin
type UpdatePluginRequest struct {
	Name       string          `json:"name,omitempty"`
	APIVersion string          `json:"apiVersion,omitempty"`
	Kind       string          `json:"kind,omitempty"`
	Metadata   json.RawMessage `json:"metadata,omitempty"`
	Spec       json.RawMessage `json:"spec,omitempty"`
}

// PluginResponse represents the HTTP response body for a plugin
type PluginResponse struct {
	ID         string          `json:"id"`
	Name       string          `json:"name"`
	APIVersion string          `json:"apiVersion"`
	Kind       string          `json:"kind"`
	Metadata   json.RawMessage `json:"metadata"`
	Spec       json.RawMessage `json:"spec"`
	CreatedAt  time.Time       `json:"createdAt"`
	UpdatedAt  time.Time       `json:"updatedAt"`
}

// ListPluginsResponse represents the HTTP response body for listing plugins
type ListPluginsResponse struct {
	Plugins []PluginResponse `json:"plugins"`
	Total   int              `json:"total"`
}
