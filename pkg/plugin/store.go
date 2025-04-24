package plugin

import (
	"context"
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
