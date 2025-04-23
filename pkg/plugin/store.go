package plugin

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/davidviejo/projects/kfs/chainlaunch/pkg/db"
)

// Store defines the interface for plugin storage operations
type Store interface {
	CreatePlugin(ctx context.Context, plugin *Plugin) error
	GetPlugin(ctx context.Context, name string) (*Plugin, error)
	ListPlugins(ctx context.Context) ([]*Plugin, error)
	DeletePlugin(ctx context.Context, name string) error
}

// SQLStore implements the Store interface using SQL database
type SQLStore struct {
	queries *db.Queries
}

// NewSQLStore creates a new SQL store
func NewSQLStore(db *sql.DB) *SQLStore {
	return &SQLStore{
		queries: db.New(db),
	}
}

// CreatePlugin creates a new plugin in the database
func (s *SQLStore) CreatePlugin(ctx context.Context, plugin *Plugin) error {
	params, err := json.Marshal(plugin.Spec.Parameters)
	if err != nil {
		return fmt.Errorf("failed to marshal parameters: %w", err)
	}

	_, err = s.queries.CreatePlugin(ctx, db.CreatePluginParams{
		Name:             plugin.Metadata.Name,
		APIVersion:       plugin.APIVersion,
		Kind:             plugin.Kind,
		ParametersSchema: string(params),
	})
	return err
}

// GetPlugin retrieves a plugin by name
func (s *SQLStore) GetPlugin(ctx context.Context, name string) (*Plugin, error) {
	dbPlugin, err := s.queries.GetPlugin(ctx, name)
	if err != nil {
		return nil, err
	}

	var params json.RawMessage
	if err := json.Unmarshal([]byte(dbPlugin.ParametersSchema), &params); err != nil {
		return nil, fmt.Errorf("failed to unmarshal parameters: %w", err)
	}

	return &Plugin{
		APIVersion: dbPlugin.APIVersion,
		Kind:       dbPlugin.Kind,
		Metadata: PluginMetadata{
			Name: dbPlugin.Name,
		},
		Spec: PluginSpec{
			Parameters: params,
		},
	}, nil
}

// ListPlugins retrieves all plugins
func (s *SQLStore) ListPlugins(ctx context.Context) ([]*Plugin, error) {
	dbPlugins, err := s.queries.ListPlugins(ctx)
	if err != nil {
		return nil, err
	}

	plugins := make([]*Plugin, len(dbPlugins))
	for i, dbPlugin := range dbPlugins {
		var params json.RawMessage
		if err := json.Unmarshal([]byte(dbPlugin.ParametersSchema), &params); err != nil {
			return nil, fmt.Errorf("failed to unmarshal parameters: %w", err)
		}

		plugins[i] = &Plugin{
			APIVersion: dbPlugin.APIVersion,
			Kind:       dbPlugin.Kind,
			Metadata: PluginMetadata{
				Name: dbPlugin.Name,
			},
			Spec: PluginSpec{
				Parameters: params,
			},
		}
	}

	return plugins, nil
}

// DeletePlugin removes a plugin by name
func (s *SQLStore) DeletePlugin(ctx context.Context, name string) error {
	return s.queries.DeletePlugin(ctx, name)
}
