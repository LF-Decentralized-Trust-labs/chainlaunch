package plugin

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/chainlaunch/chainlaunch/pkg/db"
	"github.com/chainlaunch/chainlaunch/pkg/plugin/types"
)

// Store defines the interface for plugin storage operations
type Store interface {
	CreatePlugin(ctx context.Context, plugin *types.Plugin) error
	GetPlugin(ctx context.Context, name string) (*types.Plugin, error)
	ListPlugins(ctx context.Context) ([]*types.Plugin, error)
	DeletePlugin(ctx context.Context, name string) error
	UpdatePlugin(ctx context.Context, plugin *types.Plugin) error
	UpdateDeploymentMetadata(ctx context.Context, name string, metadata map[string]interface{}) error
	UpdateDeploymentStatus(ctx context.Context, name string, status string) error
	GetDeploymentMetadata(ctx context.Context, name string) (map[string]interface{}, error)
	GetDeploymentStatus(ctx context.Context, name string) (string, error)
}

// Service represents a docker-compose service
type Service struct {
	Name        string                 `json:"name"`
	Image       string                 `json:"image"`
	Ports       []string               `json:"ports,omitempty"`
	Environment map[string]string      `json:"environment,omitempty"`
	Volumes     []string               `json:"volumes,omitempty"`
	DependsOn   []string               `json:"depends_on,omitempty"`
	Config      map[string]interface{} `json:"config,omitempty"`
}

// ServiceStatus extends Service with runtime information
type ServiceStatus struct {
	Service
	State      string   `json:"state"`
	Running    bool     `json:"running"`
	Health     string   `json:"health,omitempty"`
	Containers []string `json:"containers,omitempty"`
	LastError  string   `json:"last_error,omitempty"`
	CreatedAt  string   `json:"created_at,omitempty"`
	StartedAt  string   `json:"started_at,omitempty"`
}

// SQLStore implements the Store interface using SQL database
type SQLStore struct {
	queries *db.Queries
}

// NewSQLStore creates a new SQL store
func NewSQLStore(db *db.Queries) *SQLStore {
	return &SQLStore{
		queries: db,
	}
}

// CreatePlugin creates a new plugin in the database
func (s *SQLStore) CreatePlugin(ctx context.Context, plugin *types.Plugin) error {
	metadataJSON, err := json.Marshal(plugin.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	specJSON, err := json.Marshal(plugin.Spec)
	if err != nil {
		return fmt.Errorf("failed to marshal spec: %w", err)
	}

	_, err = s.queries.CreatePlugin(ctx, &db.CreatePluginParams{
		Name:       plugin.Metadata.Name,
		ApiVersion: plugin.APIVersion,
		Kind:       plugin.Kind,
		Metadata:   metadataJSON,
		Spec:       specJSON,
	})
	if err != nil {
		return fmt.Errorf("failed to create plugin: %w", err)
	}
	return nil
}

// GetPlugin retrieves a plugin by name
func (s *SQLStore) GetPlugin(ctx context.Context, name string) (*types.Plugin, error) {
	dbPlugin, err := s.queries.GetPlugin(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("failed to get plugin: %w", err)
	}

	metadataBytes, ok := dbPlugin.Metadata.([]byte)
	if !ok {
		return nil, fmt.Errorf("invalid metadata type: expected []byte")
	}

	var metadata types.Metadata
	if err := json.Unmarshal(metadataBytes, &metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	specBytes, ok := dbPlugin.Spec.([]byte)
	if !ok {
		return nil, fmt.Errorf("invalid spec type: expected []byte")
	}

	var spec types.Spec
	if err := json.Unmarshal(specBytes, &spec); err != nil {
		return nil, fmt.Errorf("failed to unmarshal spec: %w", err)
	}

	return &types.Plugin{
		APIVersion: dbPlugin.ApiVersion,
		Kind:       dbPlugin.Kind,
		Metadata:   metadata,
		Spec:       spec,
	}, nil
}

// ListPlugins retrieves all plugins
func (s *SQLStore) ListPlugins(ctx context.Context) ([]*types.Plugin, error) {
	dbPlugins, err := s.queries.ListPlugins(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list plugins: %w", err)
	}

	plugins := make([]*types.Plugin, len(dbPlugins))
	for i, dbPlugin := range dbPlugins {
		metadataBytes, ok := dbPlugin.Metadata.([]byte)
		if !ok {
			return nil, fmt.Errorf("invalid metadata type: expected []byte")
		}

		var metadata types.Metadata
		if err := json.Unmarshal(metadataBytes, &metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}

		specBytes, ok := dbPlugin.Spec.([]byte)
		if !ok {
			return nil, fmt.Errorf("invalid spec type: expected []byte")
		}

		var spec types.Spec
		if err := json.Unmarshal(specBytes, &spec); err != nil {
			return nil, fmt.Errorf("failed to unmarshal spec: %w", err)
		}

		plugins[i] = &types.Plugin{
			APIVersion: dbPlugin.ApiVersion,
			Kind:       dbPlugin.Kind,
			Metadata:   metadata,
			Spec:       spec,
		}
	}

	return plugins, nil
}

// DeletePlugin removes a plugin by name
func (s *SQLStore) DeletePlugin(ctx context.Context, name string) error {
	err := s.queries.DeletePlugin(ctx, name)
	if err != nil {
		return fmt.Errorf("failed to delete plugin: %w", err)
	}
	return nil
}

// UpdatePlugin updates an existing plugin
func (s *SQLStore) UpdatePlugin(ctx context.Context, plugin *types.Plugin) error {
	metadataJSON, err := json.Marshal(plugin.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	specJSON, err := json.Marshal(plugin.Spec)
	if err != nil {
		return fmt.Errorf("failed to marshal spec: %w", err)
	}

	_, err = s.queries.UpdatePlugin(ctx, &db.UpdatePluginParams{
		ApiVersion: plugin.APIVersion,
		Kind:       plugin.Kind,
		Metadata:   metadataJSON,
		Spec:       specJSON,
		Name:       plugin.Metadata.Name,
	})
	if err != nil {
		return fmt.Errorf("failed to update plugin: %w", err)
	}
	return nil
}

// UpdateDeploymentMetadata updates the deployment metadata for a plugin
func (s *SQLStore) UpdateDeploymentMetadata(ctx context.Context, name string, metadata map[string]interface{}) error {
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal deployment metadata: %w", err)
	}

	err = s.queries.UpdateDeploymentMetadata(ctx, &db.UpdateDeploymentMetadataParams{
		Name:               name,
		DeploymentMetadata: metadataJSON,
	})
	if err != nil {
		return fmt.Errorf("failed to update deployment metadata: %w", err)
	}
	return nil
}

// UpdateDeploymentStatus updates the deployment status for a plugin
func (s *SQLStore) UpdateDeploymentStatus(ctx context.Context, name string, status string) error {
	err := s.queries.UpdateDeploymentStatus(ctx, &db.UpdateDeploymentStatusParams{
		Name:             name,
		DeploymentStatus: sql.NullString{String: status, Valid: true},
	})
	if err != nil {
		return fmt.Errorf("failed to update deployment status: %w", err)
	}
	return nil
}

// GetDeploymentMetadata retrieves the deployment metadata for a plugin
func (s *SQLStore) GetDeploymentMetadata(ctx context.Context, name string) (map[string]interface{}, error) {
	metadata, err := s.queries.GetDeploymentMetadata(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("failed to get deployment metadata: %w", err)
	}

	metadataBytes, ok := metadata.([]byte)
	if !ok {
		return nil, fmt.Errorf("invalid deployment metadata type: expected []byte")
	}

	var result map[string]interface{}
	if err := json.Unmarshal(metadataBytes, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal deployment metadata: %w", err)
	}

	return result, nil
}

// GetDeploymentStatus retrieves the deployment status for a plugin
func (s *SQLStore) GetDeploymentStatus(ctx context.Context, name string) (string, error) {
	status, err := s.queries.GetDeploymentStatus(ctx, name)
	if err != nil {
		return "", fmt.Errorf("failed to get deployment status: %w", err)
	}
	return status.String, nil
}
