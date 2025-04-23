package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/compose-spec/compose-go/types"
	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/flags"
	"github.com/docker/compose/v2/pkg/api"
	"github.com/docker/compose/v2/pkg/compose"
	"gopkg.in/yaml.v3"
)

// Plugin represents a plugin definition
type Plugin struct {
	APIVersion string         `json:"apiVersion" yaml:"apiVersion"`
	Kind       string         `json:"kind" yaml:"kind"`
	Metadata   PluginMetadata `json:"metadata" yaml:"metadata"`
	Spec       PluginSpec     `json:"spec" yaml:"spec"`
}

// PluginMetadata contains metadata about the plugin
type PluginMetadata struct {
	Name string `json:"name" yaml:"name"`
}

// PluginSpec contains the plugin specification
type PluginSpec struct {
	Parameters json.RawMessage `json:"parameters" yaml:"parameters"`
}

// PluginManager handles plugin operations
type PluginManager struct {
	pluginsDir string
	compose    api.Service
}

// NewPluginManager creates a new plugin manager
func NewPluginManager(pluginsDir string) (*PluginManager, error) {
	// Create Docker client
	// dockerClient, err := client.NewClientWithOpts(client.FromEnv)
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to create Docker client: %w", err)
	// }
	// docker-compose up
	dockerCli, err := command.NewDockerCli()
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker CLI: %w", err)
	}

	err = dockerCli.Initialize(flags.NewClientOptions())
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Docker CLI: %w", err)
	}

	// check if docker engine is running
	_, err = dockerCli.Client().Info(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to check if Docker engine is running: %w", err)
	}

	// Create Compose service
	composeService := compose.NewComposeService(dockerCli)

	return &PluginManager{
		pluginsDir: pluginsDir,
		compose:    composeService,
	}, nil
}

// LoadPlugin loads a plugin from a file
func (pm *PluginManager) LoadPlugin(filePath string) (*Plugin, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read plugin file: %w", err)
	}

	var plugin Plugin
	if err := yaml.Unmarshal(data, &plugin); err != nil {
		return nil, fmt.Errorf("failed to unmarshal plugin: %w", err)
	}

	return &plugin, nil
}

// DeployPlugin deploys a plugin using docker-compose
func (pm *PluginManager) DeployPlugin(ctx context.Context, plugin *Plugin, parameters map[string]interface{}) error {
	// Create a temporary directory for the plugin
	tempDir, err := os.MkdirTemp("", plugin.Metadata.Name)
	if err != nil {
		return fmt.Errorf("failed to create temporary directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Create docker-compose.yml
	composeConfig := types.Project{
		Name: plugin.Metadata.Name,
		Services: types.Services{
			{
				Name:  plugin.Metadata.Name,
				Image: fmt.Sprintf("%s:latest", plugin.Metadata.Name),
				Environment: types.MappingWithEquals{
					"LOG_LEVEL":               stringPtr(parameters["LOG_LEVEL"].(string)),
					"CONNECTION_PROFILE_PATH": stringPtr(parameters["CONNECTION_PROFILE_PATH"].(string)),
					"API_CONFIG_PATH":         stringPtr(parameters["API_CONFIG_PATH"].(string)),
					"API_REFRESH_INTERVAL":    stringPtr(parameters["API_REFRESH_INTERVAL"].(string)),
					"MSP_ID":                  stringPtr(parameters["MSP_ID"].(string)),
				},
			},
		},
	}

	composeData, err := yaml.Marshal(composeConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal docker-compose config: %w", err)
	}

	composePath := filepath.Join(tempDir, "docker-compose.yml")
	if err := os.WriteFile(composePath, composeData, 0644); err != nil {
		return fmt.Errorf("failed to write docker-compose file: %w", err)
	}

	// Load the project
	project, err := pm.compose.Up(ctx, composePath)
	if err != nil {
		return fmt.Errorf("failed to load project: %w", err)
	}

	// Start the services
	options := api.UpOptions{
		Create: api.CreateOptions{
			RemoveOrphans: true,
			QuietPull:     true,
		},
		Start: api.StartOptions{
			Wait: true,
		},
	}

	if err := pm.compose.Up(ctx, project, options); err != nil {
		return fmt.Errorf("failed to start services: %w", err)
	}

	return nil
}

// stringPtr returns a pointer to a string
func stringPtr(s string) *string {
	return &s
}
